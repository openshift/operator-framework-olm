/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/openshift/library-go/pkg/crypto"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	catalogLabelKey        = "olm.catalogSource"
	catalogNameLabelKey    = "olm.lifecycle-server/catalog-name"
	fieldManager           = "lifecycle-controller"
	clusterRoleName        = "operator-lifecycle-manager-lifecycle-server"
	clusterRoleBindingName = "operator-lifecycle-manager-lifecycle-server"
	appLabelKey            = "app"
	appLabelVal            = "olm-lifecycle-server"
	resourceBaseName       = "lifecycle-server"
)

// LifecycleControllerReconciler reconciles CatalogSources and manages lifecycle-server resources
type LifecycleControllerReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	ServerImage                string
	CatalogSourceLabelSelector labels.Selector
	CatalogSourceFieldSelector fields.Selector
	TLSConfigProvider          *TLSConfigProvider
}

// matchesCatalogSource checks if a CatalogSource matches both label and field selectors
func (r *LifecycleControllerReconciler) matchesCatalogSource(cs *operatorsv1alpha1.CatalogSource) bool {
	if !r.CatalogSourceLabelSelector.Matches(labels.Set(cs.Labels)) {
		return false
	}
	fieldSet := fields.Set{
		"metadata.name":      cs.Name,
		"metadata.namespace": cs.Namespace,
	}
	return r.CatalogSourceFieldSelector.Matches(fieldSet)
}

// Reconcile watches CatalogSources and manages lifecycle-server resources per catalog
func (r *LifecycleControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("catalogSource", req.NamespacedName)

	log.Info("handling reconciliation request")
	defer log.Info("finished reconciliation")

	// Get the CatalogSource
	var cs operatorsv1alpha1.CatalogSource
	if err := r.Get(ctx, req.NamespacedName, &cs); err != nil {
		if errors.IsNotFound(err) {
			// CatalogSource was deleted, cleanup resources
			if err := r.cleanupResources(ctx, log, req.Namespace, req.Name); err != nil {
				return ctrl.Result{}, err
			}
			// Also reconcile the shared CRB to remove this SA
			return ctrl.Result{}, r.reconcileClusterRoleBinding(ctx, log)
		}
		log.Error(err, "failed to get catalog source")
		return ctrl.Result{}, err
	}

	// Check if CatalogSource matches our selectors
	if !r.matchesCatalogSource(&cs) {
		// CatalogSource doesn't match, cleanup any existing resources
		if err := r.cleanupResources(ctx, log, cs.Namespace, cs.Name); err != nil {
			return ctrl.Result{}, err
		}
		// Also reconcile the shared CRB to remove this SA
		return ctrl.Result{}, r.reconcileClusterRoleBinding(ctx, log)
	}

	// Get the catalog image ref from running pod
	imageRef, nodeName, err := r.getCatalogPodInfo(ctx, &cs)
	if err != nil {
		log.Error(err, "failed to get catalog pod info")
		return ctrl.Result{}, err
	}
	if imageRef == "" {
		log.Info("no valid image ref for catalog source, waiting for pod")
		return ctrl.Result{}, nil
	}

	// Ensure all resources exist for this CatalogSource
	if err := r.ensureResources(ctx, log, &cs, imageRef, nodeName); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile the shared ClusterRoleBinding
	if err := r.reconcileClusterRoleBinding(ctx, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// getCatalogPodInfo gets the image digest and node name from the catalog's running pod
func (r *LifecycleControllerReconciler) getCatalogPodInfo(ctx context.Context, cs *operatorsv1alpha1.CatalogSource) (string, string, error) {
	var pods corev1.PodList
	if err := r.List(ctx, &pods,
		client.InNamespace(cs.Namespace),
		client.MatchingLabels{catalogLabelKey: cs.Name},
	); err != nil {
		return "", "", err
	}

	// Find a running pod with a valid digest
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Status.Phase != corev1.PodRunning {
			continue
		}
		digest := imageID(p)
		if digest != "" {
			return digest, p.Spec.NodeName, nil
		}
	}

	return "", "", nil
}

// ensureResources creates or updates namespace-scoped resources for a CatalogSource
func (r *LifecycleControllerReconciler) ensureResources(ctx context.Context, log logr.Logger, cs *operatorsv1alpha1.CatalogSource, imageRef, nodeName string) error {
	name := resourceName(cs.Name)

	// Apply ServiceAccount (in catalog's namespace)
	sa := r.buildServiceAccount(name, cs)
	if err := r.Patch(ctx, sa, client.Apply, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply serviceaccount")
		return err
	}

	// Apply Service (in catalog's namespace)
	svc := r.buildService(name, cs)
	if err := r.Patch(ctx, svc, client.Apply, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply service")
		return err
	}

	// Apply Deployment (in catalog's namespace)
	deploy := r.buildDeployment(name, cs, imageRef, nodeName)
	if err := r.Patch(ctx, deploy, client.Apply, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply deployment")
		return err
	}

	// Apply NetworkPolicy (in catalog's namespace)
	np := r.buildNetworkPolicy(name, cs)
	if err := r.Patch(ctx, np, client.Apply, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply networkpolicy")
		return err
	}

	log.Info("applied resources", "name", name, "namespace", cs.Namespace, "imageRef", imageRef, "nodeName", nodeName)
	return nil
}

// reconcileClusterRoleBinding maintains a single CRB with all lifecycle-server ServiceAccounts
func (r *LifecycleControllerReconciler) reconcileClusterRoleBinding(ctx context.Context, log logr.Logger) error {
	// List all matching CatalogSources
	var allCatalogSources operatorsv1alpha1.CatalogSourceList
	if err := r.List(ctx, &allCatalogSources); err != nil {
		log.Error(err, "failed to list catalog sources for CRB reconciliation")
		return err
	}

	// Build subjects list from matching CatalogSources
	var subjects []rbacv1.Subject
	for i := range allCatalogSources.Items {
		cs := &allCatalogSources.Items[i]
		if !r.matchesCatalogSource(cs) {
			continue
		}
		// Check if SA exists (only add if we've created resources for this catalog)
		saName := resourceName(cs.Name)
		var sa corev1.ServiceAccount
		if err := r.Get(ctx, types.NamespacedName{Name: saName, Namespace: cs.Namespace}, &sa); err != nil {
			if errors.IsNotFound(err) {
				continue // SA doesn't exist yet, skip
			}
			return err
		}
		subjects = append(subjects, rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      saName,
			Namespace: cs.Namespace,
		})
	}

	// Sort subjects for deterministic ordering
	sort.Slice(subjects, func(i, j int) bool {
		if subjects[i].Namespace != subjects[j].Namespace {
			return subjects[i].Namespace < subjects[j].Namespace
		}
		return subjects[i].Name < subjects[j].Name
	})

	// Apply the CRB
	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
			Labels: map[string]string{
				appLabelKey: appLabelVal,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: subjects,
	}

	if err := r.Patch(ctx, crb, client.Apply, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply clusterrolebinding")
		return err
	}

	log.Info("reconciled clusterrolebinding", "subjectCount", len(subjects))
	return nil
}

// cleanupResources deletes namespace-scoped resources for a CatalogSource
func (r *LifecycleControllerReconciler) cleanupResources(ctx context.Context, log logr.Logger, csNamespace, csName string) error {
	name := resourceName(csName)
	log = log.WithValues("resourceName", name, "namespace", csNamespace)

	// Delete Deployment
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: csNamespace,
		},
	}
	if err := r.Delete(ctx, deploy); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "failed to delete deployment")
		return err
	}

	// Delete Service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: csNamespace,
		},
	}
	if err := r.Delete(ctx, svc); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "failed to delete service")
		return err
	}

	// Delete ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: csNamespace,
		},
	}
	if err := r.Delete(ctx, sa); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "failed to delete serviceaccount")
		return err
	}

	log.Info("cleaned up resources")
	return nil
}

// resourceName generates a DNS-compatible name for lifecycle-server resources
func resourceName(csName string) string {
	name := fmt.Sprintf("%s-%s", csName, resourceBaseName)
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.ToLower(name)
}

// buildServiceAccount creates a ServiceAccount for a lifecycle-server
func (r *LifecycleControllerReconciler) buildServiceAccount(name string, cs *operatorsv1alpha1.CatalogSource) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cs.Namespace,
			Labels: map[string]string{
				appLabelKey:         appLabelVal,
				catalogNameLabelKey: cs.Name,
			},
		},
	}
}

// buildService creates a Service for a lifecycle-server
func (r *LifecycleControllerReconciler) buildService(name string, cs *operatorsv1alpha1.CatalogSource) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cs.Namespace,
			Labels: map[string]string{
				appLabelKey:         appLabelVal,
				catalogNameLabelKey: cs.Name,
			},
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": fmt.Sprintf("%s-tls", name),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				appLabelKey:         appLabelVal,
				catalogNameLabelKey: cs.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "api",
					Port:       8443,
					TargetPort: intstr.FromString("api"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// buildDeployment creates a Deployment for a lifecycle-server
func (r *LifecycleControllerReconciler) buildDeployment(name string, cs *operatorsv1alpha1.CatalogSource, imageRef, nodeName string) *appsv1.Deployment {
	podLabels := map[string]string{
		appLabelKey:         appLabelVal,
		catalogNameLabelKey: cs.Name,
	}

	// Determine the catalog directory inside the image
	catalogDir := "/configs" // default for standard catalog images
	if cs.Spec.GrpcPodConfig != nil && cs.Spec.GrpcPodConfig.ExtractContent != nil && cs.Spec.GrpcPodConfig.ExtractContent.CatalogDir != "" {
		catalogDir = cs.Spec.GrpcPodConfig.ExtractContent.CatalogDir
	}

	const catalogMountPath = "/catalog"
	fbcPath := catalogMountPath + catalogDir

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cs.Namespace,
			Labels:    podLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: ptr.To(intstr.FromInt(1)),
					MaxSurge:       ptr.To(intstr.FromInt(1)),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
					Annotations: map[string]string{
						"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
						"openshift.io/required-scc":               "restricted-v2",
						"kubectl.kubernetes.io/default-container": "lifecycle-server",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					ServiceAccountName: name,
					PriorityClassName:  "system-cluster-critical",
					// Prefer scheduling on the same node as the catalog pod (only if nodeName is known)
					Affinity: nodeAffinityForNode(nodeName),
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: ptr.To(int64(120)),
						},
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: ptr.To(int64(120)),
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "lifecycle-server",
							Image:           r.ServerImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/lifecycle-server"},
							Args:            r.buildLifecycleServerArgs(fbcPath),
							Env: []corev1.EnvVar{
								{
									Name:  "GOMEMLIMIT",
									Value: "50MiB",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "api",
									ContainerPort: 8443,
								},
								{
									Name:          "health",
									ContainerPort: 8081,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "catalog",
									MountPath: catalogMountPath,
									ReadOnly:  true,
								},
								{
									Name:      "serving-cert",
									MountPath: "/var/run/secrets/serving-cert",
									ReadOnly:  true,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 30,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptr.To(false),
								ReadOnlyRootFilesystem:   ptr.To(true),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
							TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "catalog",
							VolumeSource: corev1.VolumeSource{
								Image: &corev1.ImageVolumeSource{
									Reference:  imageRef,
									PullPolicy: corev1.PullIfNotPresent,
								},
							},
						},
						{
							Name: "serving-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-tls", name),
								},
							},
						},
					},
				},
			},
		},
	}
}

// buildNetworkPolicy creates a NetworkPolicy for a lifecycle-server
func (r *LifecycleControllerReconciler) buildNetworkPolicy(name string, cs *operatorsv1alpha1.CatalogSource) *networkingv1.NetworkPolicy {
	tcp := corev1.ProtocolTCP
	udp := corev1.ProtocolUDP
	return &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cs.Namespace,
			Labels: map[string]string{
				appLabelKey:         appLabelVal,
				catalogNameLabelKey: cs.Name,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					appLabelKey:         appLabelVal,
					catalogNameLabelKey: cs.Name,
				},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: ptr.To(intstr.FromInt32(8443)), Protocol: &tcp},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					// API server
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: ptr.To(intstr.FromInt32(6443)), Protocol: &tcp},
					},
				},
				{
					// DNS
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: ptr.To(intstr.FromInt32(53)), Protocol: &tcp},
						{Port: ptr.To(intstr.FromInt32(53)), Protocol: &udp},
						{Port: ptr.To(intstr.FromInt32(5353)), Protocol: &tcp},
						{Port: ptr.To(intstr.FromInt32(5353)), Protocol: &udp},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
		},
	}
}

// buildLifecycleServerArgs builds the command-line arguments for lifecycle-server
func (r *LifecycleControllerReconciler) buildLifecycleServerArgs(fbcPath string) []string {
	args := []string{
		"start",
		fmt.Sprintf("--fbc-path=%s", fbcPath),
	}

	if r.TLSConfigProvider != nil {
		if cfg := r.TLSConfigProvider.Get(); cfg != nil {
			args = append(args, fmt.Sprintf("--tls-min-version=%s", crypto.TLSVersionToNameOrDie(cfg.MinVersion)))
			args = append(args, fmt.Sprintf("--tls-cipher-suites=%s", strings.Join(crypto.CipherSuitesToNamesOrDie(cfg.CipherSuites), ",")))
		}
	}

	return args
}

// imageID extracts digest from pod status (handles extract-content mode)
func imageID(pod *corev1.Pod) string {
	// In extract-content mode, look for the "extract-content" init container
	for i := range pod.Status.InitContainerStatuses {
		if pod.Status.InitContainerStatuses[i].Name == "extract-content" {
			return pod.Status.InitContainerStatuses[i].ImageID
		}
	}
	// Fallback to the first container (standard grpc mode)
	if len(pod.Status.ContainerStatuses) > 0 {
		return pod.Status.ContainerStatuses[0].ImageID
	}
	return ""
}

// nodeAffinityForNode returns a node affinity preferring the given node, or nil if nodeName is empty
func nodeAffinityForNode(nodeName string) *corev1.Affinity {
	if nodeName == "" {
		return nil
	}
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{nodeName},
							},
						},
					},
				},
			},
		},
	}
}

// LifecycleServerLabelSelector returns a label selector matching lifecycle-server deployments
func LifecycleServerLabelSelector() labels.Selector {
	return labels.SelectorFromSet(labels.Set{appLabelKey: appLabelVal})
}

// SetupWithManager sets up the controller with the Manager.
// tlsChangeSource is an optional channel source that triggers reconciliation when TLS config changes.
func (r *LifecycleControllerReconciler) SetupWithManager(mgr ctrl.Manager, tlsChangeSource source.Source) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.CatalogSource{}).
		// Watch Pods to detect catalog pod changes
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return nil
			}
			// Check if this is a catalog pod
			catalogName, ok := pod.Labels[catalogLabelKey]
			if !ok {
				return nil
			}
			// Enqueue the CatalogSource for reconciliation
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      catalogName,
						Namespace: pod.Namespace,
					},
				},
			}
		})).
		// Watch lifecycle-server Deployments to detect changes/deletion
		Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			deploy, ok := obj.(*appsv1.Deployment)
			if !ok {
				return nil
			}
			// Only watch our deployments
			if deploy.Labels[appLabelKey] != appLabelVal {
				return nil
			}
			csName := deploy.Labels[catalogNameLabelKey]
			if csName == "" {
				return nil
			}
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      csName,
						Namespace: deploy.Namespace,
					},
				},
			}
		}))

	// Add TLS change source if provided
	if tlsChangeSource != nil {
		builder = builder.WatchesRawSource(tlsChangeSource)
	}

	return builder.Complete(r)
}
