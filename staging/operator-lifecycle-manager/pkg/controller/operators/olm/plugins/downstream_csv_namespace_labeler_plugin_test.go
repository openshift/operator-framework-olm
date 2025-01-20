package plugins

import (
	"context"
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/informers/externalversions"
	listerv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	v1fake "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	listerv1 "k8s.io/client-go/listers/core/v1"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	apiregistrationfake "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/fake"
)

type fakeClientOptions struct {
	k8sResources      []runtime.Object
	extendedResources []runtime.Object
}

type fakeClientOption func(options *fakeClientOptions)

func withK8sResources(resource ...runtime.Object) fakeClientOption {
	return func(options *fakeClientOptions) {
		options.k8sResources = append(options.k8sResources, resource...)
	}
}

func withExtendedResources(resource ...runtime.Object) fakeClientOption {
	return func(options *fakeClientOptions) {
		options.extendedResources = append(options.extendedResources, resource...)
	}
}

func init() {
	klog.InitFlags(flag.CommandLine)
	if err := flag.Lookup("v").Value.Set("14"); err != nil {
		panic(err)
	}
}

func newFakeCSVNamespaceLabelerPlugin(t *testing.T, options ...fakeClientOption) (*csvNamespaceLabelerPlugin, context.CancelFunc) {
	resyncPeriod := 5 * time.Minute
	clientOptions := &fakeClientOptions{}
	for _, applyOption := range options {
		applyOption(clientOptions)
	}

	t.Log("creating fake clients")
	k8sClientFake := operatorclient.NewClient(k8sfake.NewSimpleClientset(clientOptions.k8sResources...), apiextensionsfake.NewSimpleClientset(), apiregistrationfake.NewSimpleClientset())
	extendedClient := fake.NewReactionForwardingClientsetDecorator(clientOptions.extendedResources)

	t.Log("creating informers")
	informerFactory := informers.NewSharedInformerFactory(k8sClientFake.KubernetesInterface(), resyncPeriod)
	namespaceInformer := informerFactory.Core().V1().Namespaces().Informer()

	operatorsInformerFactory := externalversions.NewSharedInformerFactory(&extendedClient.Clientset, resyncPeriod)
	nonCopiedCsvInformer := operatorsInformerFactory.Operators().V1alpha1().ClusterServiceVersions().Informer()

	t.Log("starting informers")
	ctx, cancel := context.WithCancel(context.TODO())
	stopCtx := make(chan struct{})
	go func() {
		<-ctx.Done()
		stopCtx <- struct{}{}
	}()
	informerFactory.Start(stopCtx)
	operatorsInformerFactory.Start(stopCtx)

	t.Log("waiting for informers to sync")
	syncCtx, syncCancel := context.WithTimeout(ctx, 10*time.Second)
	defer func() {
		syncCancel()
	}()
	if ok := cache.WaitForCacheSync(syncCtx.Done(), namespaceInformer.HasSynced); !ok {
		t.Fatalf("failed to wait for namespace caches to sync")
	}
	if ok := cache.WaitForCacheSync(syncCtx.Done(), nonCopiedCsvInformer.HasSynced); !ok {
		t.Fatalf("failed to wait for non-copied caches to sync")
	}

	return &csvNamespaceLabelerPlugin{
		kubeClient:      k8sClientFake,
		externalClient:  extendedClient,
		logger:          nil,
		namespaceLister: listerv1.NewNamespaceLister(namespaceInformer.GetIndexer()),
		nonCopiedCsvListerMap: map[string]listerv1alpha1.ClusterServiceVersionLister{
			metav1.NamespaceAll: listerv1alpha1.NewClusterServiceVersionLister(nonCopiedCsvInformer.GetIndexer()),
		},
	}, cancel
}

func NewCsvInNamespace(namespace string) *v1alpha1.ClusterServiceVersion {
	return &v1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-csv",
			Namespace: namespace,
		},
	}
}

func NewCopiedCsvInNamespace(namespace string) *v1alpha1.ClusterServiceVersion {
	csv := NewCsvInNamespace(namespace)
	csv.SetLabels(map[string]string{
		v1alpha1.CopiedLabelKey: "true",
	})
	return csv
}

func NewNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func NewLabeledNamespace(name string, labelValue string) *corev1.Namespace {
	ns := NewNamespace(name)
	ns.SetLabels(map[string]string{
		NamespaceLabelSyncerLabelKey: labelValue,
	})
	return ns
}

func Test_SyncIgnoresCopiedCsvs(t *testing.T) {
	// Sync ignores copied csvs
	namespace := "openshift-test"
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	assert.Nil(t, plugin.Sync(context.Background(), NewCopiedCsvInNamespace(namespace)))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncIgnoresNonOpenshiftNamespaces(t *testing.T) {
	// Sync ignores non-openshift namespaces
	namespace := "test-namespace"
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	assert.Nil(t, plugin.Sync(context.Background(), NewCopiedCsvInNamespace(namespace)))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncIgnoresPayloadOpenshiftNamespacesExceptOperators(t *testing.T) {
	// Sync ignores payload openshift namespaces, except openshift-operators
	// openshift-monitoring sync -> no label
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace("openshift-monitoring"), NewNamespace("openshift-operators")))
	defer shutdown()

	assert.Nil(t, plugin.Sync(context.Background(), NewCsvInNamespace("openshift-monitoring")))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), "openshift-monitoring", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)

	// openshift-operators sync -> label added
	assert.Nil(t, plugin.Sync(context.Background(), NewCsvInNamespace("openshift-operators")))
	ns, err = plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), "openshift-operators", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "true", ns.GetLabels()[NamespaceLabelSyncerLabelKey])
}

func Test_SyncIgnoresAlreadyLabeledNonPayloadOpenshiftNamespaces(t *testing.T) {
	// Sync ignores non-payload openshift namespaces that are already labeled
	labelValues := []string{"true", "false", " ", "", "gibberish"}
	namespace := "openshift-test"

	for _, labelValue := range labelValues {
		func() {
			plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewLabeledNamespace(namespace, labelValue)))
			defer shutdown()

			assert.Nil(t, plugin.Sync(context.Background(), NewCsvInNamespace(namespace)))

			ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, labelValue, ns.GetLabels()[NamespaceLabelSyncerLabelKey])
		}()
	}
}

func Test_SyncLabelsNonPayloadUnlabeledOpenshiftNamespaces(t *testing.T) {
	// Sync will label non-labeled non-payload openshift- namespaces
	namespace := "openshift-test"

	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	assert.Nil(t, plugin.Sync(context.Background(), NewCsvInNamespace(namespace)))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Contains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncFailsIfEventResourceIsNotCSV(t *testing.T) {
	// Sync fails if resource is not a csv\
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t)
	defer shutdown()

	assert.Error(t, plugin.Sync(context.Background(), &corev1.ConfigMap{}))
}

func Test_SyncFailsIfNamespaceNotFound(t *testing.T) {
	// Sync fails if the namespace is not found
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t)
	defer shutdown()

	assert.Error(t, plugin.Sync(context.Background(), NewCsvInNamespace("openshift-test")))
}

func Test_SyncFailsIfCSVCannotBeUpdated(t *testing.T) {
	// Sync fails if the namespace cannot be updated
	namespace := "openshift-test"
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	updateNsError := func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Namespace{}, errors.New("error updating namespace")
	}
	plugin.kubeClient.KubernetesInterface().CoreV1().(*v1fake.FakeCoreV1).PrependReactor("update", "namespaces", updateNsError)
	assert.Error(t, plugin.Sync(context.Background(), NewCsvInNamespace(namespace)))
}

func Test_SyncLabelsNamespaceWithCSV(t *testing.T) {
	// Given a namespace event for an unlabelled and non-payload openshift-* namespace
	// that contains at least one non-copied csv
	// Sync should apply the label syncer label to the namespace
	namespace := NewNamespace("openshift-test")
	csv := NewCsvInNamespace(namespace.GetName())
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(namespace), withExtendedResources(csv))
	defer shutdown()

	assert.NoError(t, plugin.Sync(context.Background(), namespace))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Contains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncDoesNotLabelNamespaceWithoutCSVs(t *testing.T) {
	// Given a namespace event for an unlabelled and non-payload openshift-* namespace
	// that contains zero non-copied csvs
	// Sync should *NOT* apply the label syncer label to the namespace
	namespace := NewNamespace("openshift-test")
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(namespace))
	defer shutdown()

	assert.NoError(t, plugin.Sync(context.Background(), namespace))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncDoesNotLabelNamespacesWithCopiedCSVs(t *testing.T) {
	// Given a namespace event for an unlabelled and non-payload openshift-* namespace
	// that only contains copied csvs
	// Sync should *NOT* apply the label syncer label to the namespace
	namespace := NewNamespace("openshift-test")
	csv := NewCopiedCsvInNamespace(namespace.GetName())
	plugin, shutdown := newFakeCSVNamespaceLabelerPlugin(t, withK8sResources(namespace), withExtendedResources(csv))
	defer shutdown()

	assert.NoError(t, plugin.Sync(context.Background(), namespace))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}
