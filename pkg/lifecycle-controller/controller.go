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
	"crypto/tls"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/crypto"
	"sigs.k8s.io/controller-runtime/pkg/event"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	networkingv1ac "k8s.io/client-go/applyconfigurations/networking/v1"
	rbacv1ac "k8s.io/client-go/applyconfigurations/rbac/v1"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

// LifecycleServerReconciler reconciles CatalogSources and manages lifecycle-server resources
type LifecycleServerReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	ServerImage                string
	CatalogSourceLabelSelector labels.Selector
	CatalogSourceFieldSelector fields.Selector
	TLSConfigProvider          *TLSConfigProvider
}

// matchesCatalogSource checks if a CatalogSource matches both label and field selectors
func (r *LifecycleServerReconciler) matchesCatalogSource(cs *operatorsv1alpha1.CatalogSource) bool {
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
func (r *LifecycleServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
func (r *LifecycleServerReconciler) getCatalogPodInfo(ctx context.Context, cs *operatorsv1alpha1.CatalogSource) (string, string, error) {
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
func (r *LifecycleServerReconciler) ensureResources(ctx context.Context, log logr.Logger, cs *operatorsv1alpha1.CatalogSource, imageRef, nodeName string) error {
	name := resourceName(cs.Name)
	applyOpts := []client.ApplyOption{client.FieldOwner(fieldManager), client.ForceOwnership}

	// Apply ServiceAccount (in catalog's namespace)
	sa := r.buildServiceAccount(name, cs)
	if err := r.Apply(ctx, sa, applyOpts...); err != nil {
		log.Error(err, "failed to apply serviceaccount")
		return err
	}

	// Apply Service (in catalog's namespace)
	svc := r.buildService(name, cs)
	if err := r.Apply(ctx, svc, applyOpts...); err != nil {
		log.Error(err, "failed to apply service")
		return err
	}

	// Apply Deployment (in catalog's namespace)
	deploy := r.buildDeployment(name, cs, imageRef, nodeName)
	if err := r.Apply(ctx, deploy, applyOpts...); err != nil {
		log.Error(err, "failed to apply deployment")
		return err
	}

	// Apply NetworkPolicy (in catalog's namespace)
	np := r.buildNetworkPolicy(name, cs)
	if err := r.Apply(ctx, np, applyOpts...); err != nil {
		log.Error(err, "failed to apply networkpolicy")
		return err
	}

	log.Info("applied resources", "name", name, "namespace", cs.Namespace, "imageRef", imageRef, "nodeName", nodeName)
	return nil
}

// reconcileClusterRoleBinding maintains a single CRB with all lifecycle-server ServiceAccounts
func (r *LifecycleServerReconciler) reconcileClusterRoleBinding(ctx context.Context, log logr.Logger) error {
	// List all matching CatalogSources
	var allCatalogSources operatorsv1alpha1.CatalogSourceList
	if err := r.List(ctx, &allCatalogSources); err != nil {
		log.Error(err, "failed to list catalog sources for CRB reconciliation")
		return err
	}

	// Build subjects list from matching CatalogSources
	var subjects []*rbacv1ac.SubjectApplyConfiguration
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
		subjects = append(subjects, rbacv1ac.Subject().
			WithKind("ServiceAccount").
			WithName(saName).
			WithNamespace(cs.Namespace))
	}

	// Sort subjects for deterministic ordering
	sort.Slice(subjects, func(i, j int) bool {
		if *subjects[i].Namespace != *subjects[j].Namespace {
			return *subjects[i].Namespace < *subjects[j].Namespace
		}
		return *subjects[i].Name < *subjects[j].Name
	})

	// Apply the CRB
	crb := rbacv1ac.ClusterRoleBinding(clusterRoleBindingName).
		WithLabels(map[string]string{
			appLabelKey: appLabelVal,
		}).
		WithRoleRef(rbacv1ac.RoleRef().
			WithAPIGroup("rbac.authorization.k8s.io").
			WithKind("ClusterRole").
			WithName(clusterRoleName)).
		WithSubjects(subjects...)

	if err := r.Apply(ctx, crb, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply clusterrolebinding")
		return err
	}

	log.Info("reconciled clusterrolebinding", "subjectCount", len(subjects))
	return nil
}

// cleanupResources deletes namespace-scoped resources for a CatalogSource
func (r *LifecycleServerReconciler) cleanupResources(ctx context.Context, log logr.Logger, csNamespace, csName string) error {
	name := resourceName(csName)
	log = log.WithValues("resourceName", name, "namespace", csNamespace)

	var deleted bool

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
	} else if err == nil {
		deleted = true
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
	} else if err == nil {
		deleted = true
	}

	// Delete NetworkPolicy
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: csNamespace,
		},
	}
	if err := r.Delete(ctx, np); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "failed to delete networkpolicy")
		return err
	} else if err == nil {
		deleted = true
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
	} else if err == nil {
		deleted = true
	}

	if deleted {
		log.Info("cleaned up resources")
	}
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
	name = strings.TrimRight(name, "-")
	return strings.ToLower(name)
}

// buildServiceAccount creates a ServiceAccount for a lifecycle-server
func (r *LifecycleServerReconciler) buildServiceAccount(name string, cs *operatorsv1alpha1.CatalogSource) *corev1ac.ServiceAccountApplyConfiguration {
	return corev1ac.ServiceAccount(name, cs.Namespace).
		WithLabels(map[string]string{
			appLabelKey:         appLabelVal,
			catalogNameLabelKey: cs.Name,
		})
}

// buildService creates a Service for a lifecycle-server
func (r *LifecycleServerReconciler) buildService(name string, cs *operatorsv1alpha1.CatalogSource) *corev1ac.ServiceApplyConfiguration {
	return corev1ac.Service(name, cs.Namespace).
		WithLabels(map[string]string{
			appLabelKey:         appLabelVal,
			catalogNameLabelKey: cs.Name,
		}).
		WithAnnotations(map[string]string{
			"service.beta.openshift.io/serving-cert-secret-name": fmt.Sprintf("%s-tls", name),
		}).
		WithSpec(corev1ac.ServiceSpec().
			WithSelector(map[string]string{
				appLabelKey:         appLabelVal,
				catalogNameLabelKey: cs.Name,
			}).
			WithPorts(corev1ac.ServicePort().
				WithName("api").
				WithPort(8443).
				WithTargetPort(intstr.FromString("api")).
				WithProtocol(corev1.ProtocolTCP)).
			WithType(corev1.ServiceTypeClusterIP))
}

// buildDeployment creates a Deployment for a lifecycle-server
func (r *LifecycleServerReconciler) buildDeployment(name string, cs *operatorsv1alpha1.CatalogSource, imageRef, nodeName string) *appsv1ac.DeploymentApplyConfiguration {
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

	return appsv1ac.Deployment(name, cs.Namespace).
		WithLabels(podLabels).
		WithSpec(appsv1ac.DeploymentSpec().
			WithReplicas(1).
			WithStrategy(appsv1ac.DeploymentStrategy().
				WithType(appsv1.RollingUpdateDeploymentStrategyType).
				WithRollingUpdate(appsv1ac.RollingUpdateDeployment().
					WithMaxUnavailable(intstr.FromInt32(0)).
					WithMaxSurge(intstr.FromInt32(1)))).
			WithSelector(metav1ac.LabelSelector().
				WithMatchLabels(podLabels)).
			WithTemplate(corev1ac.PodTemplateSpec().
				WithLabels(podLabels).
				WithAnnotations(map[string]string{
					"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
					"openshift.io/required-scc":               "restricted-v2",
					"kubectl.kubernetes.io/default-container": "lifecycle-server",
				}).
				WithSpec(corev1ac.PodSpec().
					WithSecurityContext(corev1ac.PodSecurityContext().
						WithRunAsNonRoot(true).
						WithSeccompProfile(corev1ac.SeccompProfile().
							WithType(corev1.SeccompProfileTypeRuntimeDefault))).
					WithServiceAccountName(name).
					WithPriorityClassName("system-cluster-critical").
					WithAffinity(nodeAffinityForNode(nodeName)).
					WithNodeSelector(map[string]string{
						"kubernetes.io/os": "linux",
					}).
					WithTolerations(
						corev1ac.Toleration().
							WithKey("node-role.kubernetes.io/master").
							WithOperator(corev1.TolerationOpExists).
							WithEffect(corev1.TaintEffectNoSchedule),
						corev1ac.Toleration().
							WithKey("node.kubernetes.io/unreachable").
							WithOperator(corev1.TolerationOpExists).
							WithEffect(corev1.TaintEffectNoExecute).
							WithTolerationSeconds(120),
						corev1ac.Toleration().
							WithKey("node.kubernetes.io/not-ready").
							WithOperator(corev1.TolerationOpExists).
							WithEffect(corev1.TaintEffectNoExecute).
							WithTolerationSeconds(120),
					).
					WithContainers(corev1ac.Container().
						WithName("lifecycle-server").
						WithImage(r.ServerImage).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommand("/bin/lifecycle-server").
						WithArgs(r.buildLifecycleServerArgs(fbcPath)...).
						WithEnv(corev1ac.EnvVar().
							WithName("GOMEMLIMIT").
							WithValue("50MiB")).
						WithPorts(
							corev1ac.ContainerPort().
								WithName("api").
								WithContainerPort(8443),
							corev1ac.ContainerPort().
								WithName("health").
								WithContainerPort(8081),
						).
						WithVolumeMounts(
							corev1ac.VolumeMount().
								WithName("catalog").
								WithMountPath(catalogMountPath).
								WithReadOnly(true),
							corev1ac.VolumeMount().
								WithName("serving-cert").
								WithMountPath("/var/run/secrets/serving-cert").
								WithReadOnly(true),
						).
						WithLivenessProbe(corev1ac.Probe().
							WithHTTPGet(corev1ac.HTTPGetAction().
								WithPath("/healthz").
								WithPort(intstr.FromString("health")).
								WithScheme(corev1.URISchemeHTTP)).
							WithInitialDelaySeconds(30)).
						WithReadinessProbe(corev1ac.Probe().
							WithHTTPGet(corev1ac.HTTPGetAction().
								WithPath("/healthz").
								WithPort(intstr.FromString("health")).
								WithScheme(corev1.URISchemeHTTP)).
							WithInitialDelaySeconds(30)).
						WithResources(corev1ac.ResourceRequirements().
							WithRequests(corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							})).
						WithSecurityContext(corev1ac.SecurityContext().
							WithAllowPrivilegeEscalation(false).
							WithReadOnlyRootFilesystem(true).
							WithCapabilities(corev1ac.Capabilities().
								WithDrop(corev1.Capability("ALL")))).
						WithTerminationMessagePolicy(corev1.TerminationMessageFallbackToLogsOnError)).
					WithVolumes(
						corev1ac.Volume().
							WithName("catalog").
							WithImage(corev1ac.ImageVolumeSource().
								WithReference(imageRef).
								WithPullPolicy(corev1.PullIfNotPresent)),
						corev1ac.Volume().
							WithName("serving-cert").
							WithSecret(corev1ac.SecretVolumeSource().
								WithSecretName(fmt.Sprintf("%s-tls", name)))))))
}

// buildNetworkPolicy creates a NetworkPolicy for a lifecycle-server
func (r *LifecycleServerReconciler) buildNetworkPolicy(name string, cs *operatorsv1alpha1.CatalogSource) *networkingv1ac.NetworkPolicyApplyConfiguration {
	return networkingv1ac.NetworkPolicy(name, cs.Namespace).
		WithLabels(map[string]string{
			appLabelKey:         appLabelVal,
			catalogNameLabelKey: cs.Name,
		}).
		WithSpec(networkingv1ac.NetworkPolicySpec().
			WithPodSelector(metav1ac.LabelSelector().
				WithMatchLabels(map[string]string{
					appLabelKey:         appLabelVal,
					catalogNameLabelKey: cs.Name,
				})).
			WithIngress(networkingv1ac.NetworkPolicyIngressRule().
				WithPorts(networkingv1ac.NetworkPolicyPort().
					WithPort(intstr.FromInt32(8443)).
					WithProtocol(corev1.ProtocolTCP))).
			WithEgress(
				// API server
				networkingv1ac.NetworkPolicyEgressRule().
					WithPorts(networkingv1ac.NetworkPolicyPort().
						WithPort(intstr.FromInt32(6443)).
						WithProtocol(corev1.ProtocolTCP)),
				// DNS
				networkingv1ac.NetworkPolicyEgressRule().
					WithPorts(
						networkingv1ac.NetworkPolicyPort().WithPort(intstr.FromInt32(53)).WithProtocol(corev1.ProtocolTCP),
						networkingv1ac.NetworkPolicyPort().WithPort(intstr.FromInt32(53)).WithProtocol(corev1.ProtocolUDP),
						networkingv1ac.NetworkPolicyPort().WithPort(intstr.FromInt32(5353)).WithProtocol(corev1.ProtocolTCP),
						networkingv1ac.NetworkPolicyPort().WithPort(intstr.FromInt32(5353)).WithProtocol(corev1.ProtocolUDP)),
			).
			WithPolicyTypes(networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress))
}

// buildLifecycleServerArgs builds the command-line arguments for lifecycle-server.
// TLS settings are passed as CLI args rather than dynamically watched because
// cluster TLS profile changes are expected to be rare. When a change occurs,
// the controller rebuilds the Deployment with updated args, causing a rolling restart.
func (r *LifecycleServerReconciler) buildLifecycleServerArgs(fbcPath string) []string {
	args := []string{
		"start",
		fmt.Sprintf("--fbc-path=%s", fbcPath),
	}

	if r.TLSConfigProvider != nil {
		cfg, _ := r.TLSConfigProvider.Get()
		args = append(args, fmt.Sprintf("--tls-min-version=%s", crypto.TLSVersionToNameOrDie(cfg.MinVersion)))
		if cfg.MinVersion <= tls.VersionTLS12 {
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
func nodeAffinityForNode(nodeName string) *corev1ac.AffinityApplyConfiguration {
	if nodeName == "" {
		return nil
	}
	return corev1ac.Affinity().
		WithNodeAffinity(corev1ac.NodeAffinity().
			WithPreferredDuringSchedulingIgnoredDuringExecution(
				corev1ac.PreferredSchedulingTerm().
					WithWeight(100).
					WithPreference(corev1ac.NodeSelectorTerm().
						WithMatchExpressions(corev1ac.NodeSelectorRequirement().
							WithKey("kubernetes.io/hostname").
							WithOperator(corev1.NodeSelectorOpIn).
							WithValues(nodeName)))))
}

// LifecycleServerLabelSelector returns a label selector matching lifecycle-server deployments
func LifecycleServerLabelSelector() labels.Selector {
	return labels.SelectorFromSet(labels.Set{appLabelKey: appLabelVal})
}

// catalogPodPredicate filters pod events to only those where fields relevant
// to the reconciler have changed: Status.Phase, container ImageIDs, or Spec.NodeName.
func catalogPodPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool { return true },
		DeleteFunc: func(e event.DeleteEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, ok := e.ObjectOld.(*corev1.Pod)
			if !ok {
				return false
			}
			newPod, ok := e.ObjectNew.(*corev1.Pod)
			if !ok {
				return false
			}
			if oldPod.Status.Phase != newPod.Status.Phase {
				return true
			}
			if oldPod.Spec.NodeName != newPod.Spec.NodeName {
				return true
			}
			if imageID(oldPod) != imageID(newPod) {
				return true
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool { return true },
	}
}

// mapPodToCatalogSource maps a Pod event to a reconcile request for its owning CatalogSource.
// Pods without a catalog label or with an empty catalog label value are ignored.
func mapPodToCatalogSource(_ context.Context, obj client.Object) []reconcile.Request {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil
	}
	// Check if this is a catalog pod
	catalogName := pod.Labels[catalogLabelKey]
	if catalogName == "" {
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
}

// mapLifecycleResourceToCatalogSource maps a lifecycle-server resource event to a reconcile request for its CatalogSource.
func mapLifecycleResourceToCatalogSource(_ context.Context, obj client.Object) []reconcile.Request {
	// Only watch our resources
	if obj.GetLabels()[appLabelKey] != appLabelVal {
		return nil
	}
	csName := obj.GetLabels()[catalogNameLabelKey]
	if csName == "" {
		return nil
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      csName,
				Namespace: obj.GetNamespace(),
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
// tlsChangeSource is an optional channel source that triggers reconciliation when TLS profileSpec changes.
func (r *LifecycleServerReconciler) SetupWithManager(mgr ctrl.Manager, tlsProfileChan <-chan event.TypedGenericEvent[configv1.TLSProfileSpec]) error {
	bldr := ctrl.NewControllerManagedBy(mgr).
		// Watch CatalogSources, but only reconcile on spec or label changes (not status-only updates).
		For(&operatorsv1alpha1.CatalogSource{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{}),
		)).
		// Watch Pods to detect catalog pod changes, but only when phase, imageID, or nodeName change.
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(mapPodToCatalogSource), builder.WithPredicates(catalogPodPredicate())).
		// Watch lifecycle-server resources to detect spec drift or deletion.
		Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(mapLifecycleResourceToCatalogSource), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&corev1.ServiceAccount{}, handler.EnqueueRequestsFromMapFunc(mapLifecycleResourceToCatalogSource)).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(mapLifecycleResourceToCatalogSource), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&networkingv1.NetworkPolicy{}, handler.EnqueueRequestsFromMapFunc(mapLifecycleResourceToCatalogSource), builder.WithPredicates(predicate.GenerationChangedPredicate{}))

	// Add TLS change source if provided
	bldr = bldr.WatchesRawSource(source.Channel(tlsProfileChan, handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, _ configv1.TLSProfileSpec) []reconcile.Request {
		// Trigger reconciliation of all CatalogSources to update lifecycle-server deployments
		var catalogSources operatorsv1alpha1.CatalogSourceList
		if err := mgr.GetClient().List(ctx, &catalogSources); err != nil {
			r.Log.Error(err, "failed to list CatalogSources to requeue for TLS reconfiguration; CatalogSources will not receive new TLS configuration until their next reconciliation")
			return nil
		}

		// Send events to trigger reconciliation
		var requests []reconcile.Request
		for _, obj := range catalogSources.Items {
			requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&obj)})
		}
		return requests
	})))

	return bldr.Complete(r)
}
