package controllers

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(operatorsv1alpha1.AddToScheme(scheme))
	return scheme
}

func testReconciler(cl client.Client) *LifecycleServerReconciler {
	return &LifecycleServerReconciler{
		Client:                     cl,
		Log:                        logr.Discard(),
		Scheme:                     testScheme(),
		ServerImage:                "quay.io/test/lifecycle-server:latest",
		CatalogSourceLabelSelector: labels.Everything(),
		CatalogSourceFieldSelector: fields.Everything(),
	}
}

func newCatalogSource(name, namespace string, labelMap map[string]string) *operatorsv1alpha1.CatalogSource {
	return &operatorsv1alpha1.CatalogSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labelMap,
		},
		Spec: operatorsv1alpha1.CatalogSourceSpec{
			SourceType: operatorsv1alpha1.SourceTypeGrpc,
			Image:      "quay.io/test/catalog:latest",
		},
	}
}

func catalogPod(csName, namespace, nodeName, imageID string, phase corev1.PodPhase) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csName + "-pod",
			Namespace: namespace,
			Labels: map[string]string{
				catalogLabelKey: csName,
			},
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			Containers: []corev1.Container{
				{Name: "registry"},
			},
		},
		Status: corev1.PodStatus{
			Phase: phase,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:    "registry",
					ImageID: imageID,
				},
			},
		},
	}
}

// --- Pure function tests ---

func TestResourceName(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "my-catalog",
			expected: "my-catalog-lifecycle-server",
		},
		{
			name:     "dots replaced with hyphens",
			input:    "my.catalog",
			expected: "my-catalog-lifecycle-server",
		},
		{
			name:     "underscores replaced with hyphens",
			input:    "my_catalog",
			expected: "my-catalog-lifecycle-server",
		},
		{
			name:     "mixed case and special characters",
			input:    "My.Catalog_Name",
			expected: "my-catalog-name-lifecycle-server",
		},
		{
			name:     "truncation at 63 chars",
			input:    "this-is-a-very-long-catalog-source-name-that-exceeds-the-dns-limit",
			expected: "this-is-a-very-long-catalog-source-name-that-exceeds-the-dns-li",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "-lifecycle-server",
		},
		{
			name:     "already lowercase with hyphens",
			input:    "redhat-operators",
			expected: "redhat-operators-lifecycle-server",
		},
		{
			name:     "truncation should not produce trailing hyphen",
			input:    "this-is-a-very-long-catalog-source-name-that-exceeds-the-dns--",
			expected: "this-is-a-very-long-catalog-source-name-that-exceeds-the-dns",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := resourceName(tc.input)
			require.Equal(t, tc.expected, result)
			require.LessOrEqual(t, len(result), 63, "resource name must not exceed 63 characters")
		})
	}
}

func TestImageID(t *testing.T) {
	tt := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "extract-content init container returns its ImageID",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "extract-content",
							ImageID: "sha256:abc123",
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "registry",
							ImageID: "sha256:def456",
						},
					},
				},
			},
			expected: "sha256:abc123",
		},
		{
			name: "no extract-content init container falls back to first container",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "other-init",
							ImageID: "sha256:other",
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "registry",
							ImageID: "sha256:def456",
						},
					},
				},
			},
			expected: "sha256:def456",
		},
		{
			name: "no init containers falls back to first container",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "registry",
							ImageID: "sha256:def456",
						},
					},
				},
			},
			expected: "sha256:def456",
		},
		{
			name: "no container statuses returns empty",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{},
			},
			expected: "",
		},
		{
			name: "extract-content with empty ImageID returns empty",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "extract-content",
							ImageID: "",
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := imageID(tc.pod)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNodeAffinityForNode(t *testing.T) {
	tt := []struct {
		name     string
		nodeName string
		isNil    bool
	}{
		{
			name:     "empty node name returns nil",
			nodeName: "",
			isNil:    true,
		},
		{
			name:     "non-empty node name returns affinity",
			nodeName: "worker-node-1",
			isNil:    false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := nodeAffinityForNode(tc.nodeName)
			if tc.isNil {
				require.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			require.NotNil(t, result.NodeAffinity)
			preferred := result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			require.Len(t, preferred, 1)
			require.Equal(t, int32(100), preferred[0].Weight)
			require.Len(t, preferred[0].Preference.MatchExpressions, 1)
			expr := preferred[0].Preference.MatchExpressions[0]
			require.Equal(t, "kubernetes.io/hostname", expr.Key)
			require.Equal(t, corev1.NodeSelectorOpIn, expr.Operator)
			require.Equal(t, []string{tc.nodeName}, expr.Values)
		})
	}
}

func TestLifecycleServerLabelSelector(t *testing.T) {
	sel := LifecycleServerLabelSelector()
	require.True(t, sel.Matches(labels.Set{appLabelKey: appLabelVal}))
	require.False(t, sel.Matches(labels.Set{"app": "other"}))
	require.False(t, sel.Matches(labels.Set{}))
}

// --- Builder method tests ---

func TestBuildServiceAccount(t *testing.T) {
	r := testReconciler(nil)
	cs := newCatalogSource("test-catalog", "test-ns", nil)
	name := resourceName(cs.Name)

	sa := r.buildServiceAccount(name, cs)

	require.Equal(t, name, sa.Name)
	require.Equal(t, "test-ns", sa.Namespace)
	require.Equal(t, appLabelVal, sa.Labels[appLabelKey])
	require.Equal(t, "test-catalog", sa.Labels[catalogNameLabelKey])
	require.Equal(t, "v1", sa.TypeMeta.APIVersion)
	require.Equal(t, "ServiceAccount", sa.TypeMeta.Kind)
}

func TestBuildService(t *testing.T) {
	r := testReconciler(nil)
	cs := newCatalogSource("test-catalog", "test-ns", nil)
	name := resourceName(cs.Name)

	svc := r.buildService(name, cs)

	require.Equal(t, name, svc.Name)
	require.Equal(t, "test-ns", svc.Namespace)
	require.Equal(t, appLabelVal, svc.Labels[appLabelKey])
	require.Equal(t, "test-catalog", svc.Labels[catalogNameLabelKey])

	// Serving cert annotation
	require.Equal(t, name+"-tls", svc.Annotations["service.beta.openshift.io/serving-cert-secret-name"])

	// Selector
	require.Equal(t, appLabelVal, svc.Spec.Selector[appLabelKey])
	require.Equal(t, "test-catalog", svc.Spec.Selector[catalogNameLabelKey])

	// Port
	require.Len(t, svc.Spec.Ports, 1)
	port := svc.Spec.Ports[0]
	require.Equal(t, "api", port.Name)
	require.Equal(t, int32(8443), port.Port)
	require.Equal(t, intstr.FromString("api"), port.TargetPort)
	require.Equal(t, corev1.ProtocolTCP, port.Protocol)
	require.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type)
}

func TestBuildDeployment(t *testing.T) {
	tt := []struct {
		name               string
		cs                 *operatorsv1alpha1.CatalogSource
		imageRef           string
		nodeName           string
		expectedCatalogDir string
		expectAffinity     bool
	}{
		{
			name:               "default catalog dir when GrpcPodConfig is nil",
			cs:                 newCatalogSource("test-catalog", "test-ns", nil),
			imageRef:           "sha256:abc123",
			nodeName:           "worker-1",
			expectedCatalogDir: "/catalog/configs",
			expectAffinity:     true,
		},
		{
			name: "custom catalog dir from ExtractContent",
			cs: &operatorsv1alpha1.CatalogSource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-catalog",
					Namespace: "test-ns",
				},
				Spec: operatorsv1alpha1.CatalogSourceSpec{
					GrpcPodConfig: &operatorsv1alpha1.GrpcPodConfig{
						ExtractContent: &operatorsv1alpha1.ExtractContentConfig{
							CatalogDir: "/custom/path",
						},
					},
				},
			},
			imageRef:           "sha256:def456",
			nodeName:           "",
			expectedCatalogDir: "/catalog/custom/path",
			expectAffinity:     false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := testReconciler(nil)
			name := resourceName(tc.cs.Name)
			deploy := r.buildDeployment(name, tc.cs, tc.imageRef, tc.nodeName)

			require.Equal(t, name, deploy.Name)
			require.Equal(t, tc.cs.Namespace, deploy.Namespace)

			// Replicas
			require.NotNil(t, deploy.Spec.Replicas)
			require.Equal(t, int32(1), *deploy.Spec.Replicas)

			// Strategy
			require.Equal(t, appsv1.RollingUpdateDeploymentStrategyType, deploy.Spec.Strategy.Type)

			// Pod template
			podSpec := deploy.Spec.Template.Spec

			// Security context
			require.NotNil(t, podSpec.SecurityContext)
			require.NotNil(t, podSpec.SecurityContext.RunAsNonRoot)
			require.True(t, *podSpec.SecurityContext.RunAsNonRoot)
			require.NotNil(t, podSpec.SecurityContext.SeccompProfile)
			require.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, podSpec.SecurityContext.SeccompProfile.Type)

			// Service account
			require.Equal(t, name, podSpec.ServiceAccountName)

			// Priority class
			require.Equal(t, "system-cluster-critical", podSpec.PriorityClassName)

			// Node affinity
			if tc.expectAffinity {
				require.NotNil(t, podSpec.Affinity)
			} else {
				require.Nil(t, podSpec.Affinity)
			}

			// Node selector
			require.Equal(t, "linux", podSpec.NodeSelector["kubernetes.io/os"])

			// Container
			require.Len(t, podSpec.Containers, 1)
			container := podSpec.Containers[0]
			require.Equal(t, "lifecycle-server", container.Name)
			require.Equal(t, r.ServerImage, container.Image)
			require.Equal(t, corev1.PullIfNotPresent, container.ImagePullPolicy)

			// GOMEMLIMIT env
			require.Len(t, container.Env, 1)
			require.Equal(t, "GOMEMLIMIT", container.Env[0].Name)
			require.Equal(t, "50MiB", container.Env[0].Value)

			// Ports
			require.Len(t, container.Ports, 2)
			require.Equal(t, "api", container.Ports[0].Name)
			require.Equal(t, int32(8443), container.Ports[0].ContainerPort)
			require.Equal(t, "health", container.Ports[1].Name)
			require.Equal(t, int32(8081), container.Ports[1].ContainerPort)

			// TerminationMessagePolicy
			require.Equal(t, corev1.TerminationMessageFallbackToLogsOnError, container.TerminationMessagePolicy)

			// Container security context
			require.NotNil(t, container.SecurityContext)
			require.NotNil(t, container.SecurityContext.AllowPrivilegeEscalation)
			require.False(t, *container.SecurityContext.AllowPrivilegeEscalation)
			require.NotNil(t, container.SecurityContext.ReadOnlyRootFilesystem)
			require.True(t, *container.SecurityContext.ReadOnlyRootFilesystem)
			require.Equal(t, []corev1.Capability{"ALL"}, container.SecurityContext.Capabilities.Drop)

			// Probes
			require.NotNil(t, container.LivenessProbe)
			require.NotNil(t, container.ReadinessProbe)
			require.Equal(t, "/healthz", container.LivenessProbe.HTTPGet.Path)
			require.Equal(t, "/healthz", container.ReadinessProbe.HTTPGet.Path)

			// Resource requests
			require.NotNil(t, container.Resources.Requests)
			require.Equal(t, "10m", container.Resources.Requests.Cpu().String())
			require.Equal(t, "50Mi", container.Resources.Requests.Memory().String())

			// Volume mounts (check FBC path in args)
			require.Len(t, container.VolumeMounts, 2)
			require.Equal(t, "catalog", container.VolumeMounts[0].Name)
			require.True(t, container.VolumeMounts[0].ReadOnly)
			require.Equal(t, "serving-cert", container.VolumeMounts[1].Name)
			require.Equal(t, "/var/run/secrets/serving-cert", container.VolumeMounts[1].MountPath)

			// Command includes fbc-path
			require.Contains(t, container.Args, "start")
			foundFBCArg := false
			for _, arg := range container.Args {
				if arg == "--fbc-path="+tc.expectedCatalogDir {
					foundFBCArg = true
				}
			}
			require.True(t, foundFBCArg, "expected --fbc-path=%s in args %v", tc.expectedCatalogDir, container.Args)

			// Volumes
			require.Len(t, podSpec.Volumes, 2)
			require.Equal(t, "catalog", podSpec.Volumes[0].Name)
			require.NotNil(t, podSpec.Volumes[0].Image)
			require.Equal(t, tc.imageRef, podSpec.Volumes[0].Image.Reference)
			require.Equal(t, "serving-cert", podSpec.Volumes[1].Name)
			require.NotNil(t, podSpec.Volumes[1].Secret)
			require.Equal(t, name+"-tls", podSpec.Volumes[1].Secret.SecretName)
		})
	}
}

func TestBuildNetworkPolicy(t *testing.T) {
	r := testReconciler(nil)
	cs := newCatalogSource("test-catalog", "test-ns", nil)
	name := resourceName(cs.Name)

	np := r.buildNetworkPolicy(name, cs)

	require.Equal(t, name, np.Name)
	require.Equal(t, "test-ns", np.Namespace)
	require.Equal(t, appLabelVal, np.Labels[appLabelKey])

	// Pod selector
	require.Equal(t, appLabelVal, np.Spec.PodSelector.MatchLabels[appLabelKey])
	require.Equal(t, "test-catalog", np.Spec.PodSelector.MatchLabels[catalogNameLabelKey])

	// Policy types
	require.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	require.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)

	// Ingress: 8443/TCP
	require.Len(t, np.Spec.Ingress, 1)
	require.Len(t, np.Spec.Ingress[0].Ports, 1)
	require.Equal(t, int32(8443), np.Spec.Ingress[0].Ports[0].Port.IntVal)
	tcp := corev1.ProtocolTCP
	require.Equal(t, &tcp, np.Spec.Ingress[0].Ports[0].Protocol)

	// Egress: API server (6443) and DNS (53, 5353)
	require.Len(t, np.Spec.Egress, 2)
	// API server
	require.Len(t, np.Spec.Egress[0].Ports, 1)
	require.Equal(t, int32(6443), np.Spec.Egress[0].Ports[0].Port.IntVal)
	// DNS
	require.Len(t, np.Spec.Egress[1].Ports, 4)
}

func TestBuildLifecycleServerArgs(t *testing.T) {
	tt := []struct {
		name              string
		tlsConfigProvider *TLSConfigProvider
		fbcPath           string
		expectMinVersion  bool
		expectCipherSuite bool
	}{
		{
			name:              "without TLS provider",
			tlsConfigProvider: nil,
			fbcPath:           "/catalog/configs",
			expectMinVersion:  false,
			expectCipherSuite: false,
		},
		{
			name: "with TLS 1.2 profile includes min version and cipher suites",
			tlsConfigProvider: NewTLSConfigProvider(dummyGetCertificate, configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers: []string{
					"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
					"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
				},
			}),
			fbcPath:           "/catalog/configs",
			expectMinVersion:  true,
			expectCipherSuite: true,
		},
		{
			name: "with TLS 1.3 profile includes min version but NOT cipher suites",
			tlsConfigProvider: NewTLSConfigProvider(dummyGetCertificate, configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS13,
			}),
			fbcPath:           "/catalog/custom",
			expectMinVersion:  true,
			expectCipherSuite: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := testReconciler(nil)
			r.TLSConfigProvider = tc.tlsConfigProvider
			args := r.buildLifecycleServerArgs(tc.fbcPath)

			require.Contains(t, args, "start")
			require.Contains(t, args, "--fbc-path="+tc.fbcPath)

			hasMinVersion := false
			hasCipherSuites := false
			for _, arg := range args {
				if strings.HasPrefix(arg, "--tls-min-version=") {
					hasMinVersion = true
				}
				if strings.HasPrefix(arg, "--tls-cipher-suites=") {
					hasCipherSuites = true
				}
			}
			require.Equal(t, tc.expectMinVersion, hasMinVersion, "tls-min-version flag presence mismatch")
			require.Equal(t, tc.expectCipherSuite, hasCipherSuites, "tls-cipher-suites flag presence mismatch")
		})
	}
}

// --- matchesCatalogSource tests ---

func TestMatchesCatalogSource(t *testing.T) {
	tt := []struct {
		name          string
		labelSelector string
		fieldSelector string
		cs            *operatorsv1alpha1.CatalogSource
		expected      bool
	}{
		{
			name:          "everything selectors match all",
			labelSelector: "",
			fieldSelector: "",
			cs:            newCatalogSource("test", "test-ns", nil),
			expected:      true,
		},
		{
			name:          "label selector matches",
			labelSelector: "env=prod",
			fieldSelector: "",
			cs:            newCatalogSource("test", "test-ns", map[string]string{"env": "prod"}),
			expected:      true,
		},
		{
			name:          "label selector does not match",
			labelSelector: "env=prod",
			fieldSelector: "",
			cs:            newCatalogSource("test", "test-ns", map[string]string{"env": "dev"}),
			expected:      false,
		},
		{
			name:          "field selector matches name",
			labelSelector: "",
			fieldSelector: "metadata.name=test",
			cs:            newCatalogSource("test", "test-ns", nil),
			expected:      true,
		},
		{
			name:          "field selector does not match name",
			labelSelector: "",
			fieldSelector: "metadata.name=other",
			cs:            newCatalogSource("test", "test-ns", nil),
			expected:      false,
		},
		{
			name:          "both selectors must match",
			labelSelector: "env=prod",
			fieldSelector: "metadata.name=test",
			cs:            newCatalogSource("test", "test-ns", map[string]string{"env": "prod"}),
			expected:      true,
		},
		{
			name:          "label matches but field does not",
			labelSelector: "env=prod",
			fieldSelector: "metadata.name=other",
			cs:            newCatalogSource("test", "test-ns", map[string]string{"env": "prod"}),
			expected:      false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			labelSel, err := labels.Parse(tc.labelSelector)
			require.NoError(t, err)
			fieldSel, err := fields.ParseSelector(tc.fieldSelector)
			require.NoError(t, err)

			r := testReconciler(nil)
			r.CatalogSourceLabelSelector = labelSel
			r.CatalogSourceFieldSelector = fieldSel

			result := r.matchesCatalogSource(tc.cs)
			require.Equal(t, tc.expected, result)
		})
	}
}

// --- Reconcile integration tests with fake client ---

func TestReconcile_CatalogSourceNotFound(t *testing.T) {
	scheme := testScheme()
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	r := testReconciler(cl)

	// Reconcile a CatalogSource that doesn't exist - should not error
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "test-ns"},
	})
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_CatalogSourceDoesNotMatchSelectors(t *testing.T) {
	scheme := testScheme()
	cs := newCatalogSource("test-catalog", "test-ns", map[string]string{"env": "dev"})
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cs).
		Build()

	labelSel, err := labels.Parse("env=prod")
	require.NoError(t, err)

	r := testReconciler(cl)
	r.CatalogSourceLabelSelector = labelSel

	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "test-catalog", Namespace: "test-ns"},
	})
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_NoPodRunning(t *testing.T) {
	scheme := testScheme()
	cs := newCatalogSource("test-catalog", "test-ns", nil)
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cs).
		Build()

	r := testReconciler(cl)

	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "test-catalog", Namespace: "test-ns"},
	})
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_MatchingCatalogSourceWithRunningPod(t *testing.T) {
	scheme := testScheme()
	cs := newCatalogSource("test-catalog", "test-ns", nil)
	pod := catalogPod("test-catalog", "test-ns", "worker-1", "sha256:abc123", corev1.PodRunning)

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cs, pod).
		Build()

	r := testReconciler(cl)

	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "test-catalog", Namespace: "test-ns"},
	})
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{}, result)

	ctx := context.Background()
	name := resourceName("test-catalog")

	// Verify ServiceAccount was created
	var sa corev1.ServiceAccount
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &sa)
	require.NoError(t, err)
	require.Equal(t, appLabelVal, sa.Labels[appLabelKey])

	// Verify Service was created
	var svc corev1.Service
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &svc)
	require.NoError(t, err)
	require.Equal(t, int32(8443), svc.Spec.Ports[0].Port)

	// Verify Deployment was created
	var deploy appsv1.Deployment
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &deploy)
	require.NoError(t, err)
	require.Equal(t, r.ServerImage, deploy.Spec.Template.Spec.Containers[0].Image)

	// Verify NetworkPolicy was created
	var np networkingv1.NetworkPolicy
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &np)
	require.NoError(t, err)

	// Verify ClusterRoleBinding was created
	var crb rbacv1.ClusterRoleBinding
	err = cl.Get(ctx, types.NamespacedName{Name: clusterRoleBindingName}, &crb)
	require.NoError(t, err)
	require.Equal(t, clusterRoleName, crb.RoleRef.Name)
	require.Len(t, crb.Subjects, 1)
	require.Equal(t, name, crb.Subjects[0].Name)
	require.Equal(t, "test-ns", crb.Subjects[0].Namespace)
}

// --- cleanupResources tests ---

func TestCleanupResources(t *testing.T) {
	scheme := testScheme()
	name := resourceName("test-catalog")
	ctx := context.Background()

	t.Run("deletes all resources including NetworkPolicy", func(t *testing.T) {
		// Pre-create all 4 resource types that cleanupResources should delete
		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns"},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns"},
		}
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns"},
		}
		np := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns"},
		}

		cl := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(deploy, svc, sa, np).
			Build()

		r := testReconciler(cl)

		err := r.cleanupResources(ctx, logr.Discard(), "test-ns", "test-catalog")
		require.NoError(t, err)

		// Verify all 4 resource types are deleted
		err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &appsv1.Deployment{})
		require.True(t, errors.IsNotFound(err), "Deployment should be deleted")

		err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &corev1.Service{})
		require.True(t, errors.IsNotFound(err), "Service should be deleted")

		err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &corev1.ServiceAccount{})
		require.True(t, errors.IsNotFound(err), "ServiceAccount should be deleted")

		err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: "test-ns"}, &networkingv1.NetworkPolicy{})
		require.True(t, errors.IsNotFound(err), "NetworkPolicy should be deleted")
	})

	t.Run("handles not-found resources gracefully", func(t *testing.T) {
		// No resources exist
		cl := fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		r := testReconciler(cl)

		err := r.cleanupResources(ctx, logr.Discard(), "test-ns", "test-catalog")
		require.NoError(t, err)
	})
}

// --- reconcileClusterRoleBinding tests ---

func TestReconcileClusterRoleBinding(t *testing.T) {
	scheme := testScheme()
	ctx := context.Background()

	t.Run("no matching CatalogSources produces CRB with no subjects", func(t *testing.T) {
		cl := fake.NewClientBuilder().WithScheme(scheme).Build()
		r := testReconciler(cl)

		err := r.reconcileClusterRoleBinding(ctx, logr.Discard())
		require.NoError(t, err)

		var crb rbacv1.ClusterRoleBinding
		err = cl.Get(ctx, types.NamespacedName{Name: clusterRoleBindingName}, &crb)
		require.NoError(t, err)
		require.Empty(t, crb.Subjects)
		require.Equal(t, clusterRoleName, crb.RoleRef.Name)
	})

	t.Run("multiple matching CatalogSources produce sorted subjects", func(t *testing.T) {
		cs1 := newCatalogSource("catalog-z", "ns-a", nil)
		cs2 := newCatalogSource("catalog-a", "ns-b", nil)
		sa1Name := resourceName("catalog-z")
		sa2Name := resourceName("catalog-a")
		sa1 := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: sa1Name, Namespace: "ns-a"}}
		sa2 := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: sa2Name, Namespace: "ns-b"}}

		cl := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(cs1, cs2, sa1, sa2).
			Build()

		r := testReconciler(cl)

		err := r.reconcileClusterRoleBinding(ctx, logr.Discard())
		require.NoError(t, err)

		var crb rbacv1.ClusterRoleBinding
		err = cl.Get(ctx, types.NamespacedName{Name: clusterRoleBindingName}, &crb)
		require.NoError(t, err)
		require.Len(t, crb.Subjects, 2)

		// Subjects should be sorted by namespace, then name
		require.Equal(t, "ns-a", crb.Subjects[0].Namespace)
		require.Equal(t, sa1Name, crb.Subjects[0].Name)
		require.Equal(t, "ns-b", crb.Subjects[1].Namespace)
		require.Equal(t, sa2Name, crb.Subjects[1].Name)
	})

	t.Run("CatalogSource without SA is skipped", func(t *testing.T) {
		cs := newCatalogSource("catalog-no-sa", "test-ns", nil)

		cl := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(cs).
			Build()

		r := testReconciler(cl)

		err := r.reconcileClusterRoleBinding(ctx, logr.Discard())
		require.NoError(t, err)

		var crb rbacv1.ClusterRoleBinding
		err = cl.Get(ctx, types.NamespacedName{Name: clusterRoleBindingName}, &crb)
		require.NoError(t, err)
		require.Empty(t, crb.Subjects)
	})
}

// --- getCatalogPodInfo tests ---

func TestGetCatalogPodInfo(t *testing.T) {
	scheme := testScheme()
	ctx := context.Background()

	tt := []struct {
		name            string
		pods            []*corev1.Pod
		expectedImage   string
		expectedNode    string
		expectErr       bool
	}{
		{
			name:          "no pods returns empty",
			pods:          nil,
			expectedImage: "",
			expectedNode:  "",
		},
		{
			name: "running pod with digest",
			pods: []*corev1.Pod{
				catalogPod("test-catalog", "test-ns", "worker-1", "sha256:abc123", corev1.PodRunning),
			},
			expectedImage: "sha256:abc123",
			expectedNode:  "worker-1",
		},
		{
			name: "pending pod is skipped",
			pods: []*corev1.Pod{
				catalogPod("test-catalog", "test-ns", "worker-1", "sha256:abc123", corev1.PodPending),
			},
			expectedImage: "",
			expectedNode:  "",
		},
		{
			name: "running pod with empty imageID is skipped",
			pods: []*corev1.Pod{
				catalogPod("test-catalog", "test-ns", "worker-1", "", corev1.PodRunning),
			},
			expectedImage: "",
			expectedNode:  "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := fake.NewClientBuilder().WithScheme(scheme)
			for _, p := range tc.pods {
				builder = builder.WithObjects(p)
			}
			cl := builder.Build()

			r := testReconciler(cl)
			cs := newCatalogSource("test-catalog", "test-ns", nil)

			image, node, err := r.getCatalogPodInfo(ctx, cs)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedImage, image)
			require.Equal(t, tc.expectedNode, node)
		})
	}
}
