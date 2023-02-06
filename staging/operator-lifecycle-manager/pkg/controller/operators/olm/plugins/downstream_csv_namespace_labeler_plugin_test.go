package plugins

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	listerv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/kubestate"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	v1fake "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	listerv1 "k8s.io/client-go/listers/core/v1"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
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

func NewFakeCSVNamespaceLabelerPlugin(t *testing.T, options ...fakeClientOption) (*csvNamespaceLabelerPlugin, context.CancelFunc) {

	resyncPeriod := 5 * time.Minute
	clientOptions := &fakeClientOptions{}
	for _, applyOption := range options {
		applyOption(clientOptions)
	}

	// create fake clients
	k8sClientFake := operatorclient.NewClient(k8sfake.NewSimpleClientset(clientOptions.k8sResources...), apiextensionsfake.NewSimpleClientset(), apiregistrationfake.NewSimpleClientset())
	extendedClient := fake.NewReactionForwardingClientsetDecorator(clientOptions.extendedResources)

	// create informers
	namespaceInformer := newNamespaceInformer(k8sClientFake, resyncPeriod)
	nonCopiedCsvInformer := newNonCopiedCsvInformerForNamespace(metav1.NamespaceAll, extendedClient, resyncPeriod)

	// sync caches
	ctx, cancel := context.WithCancel(context.TODO())
	go namespaceInformer.Run(ctx.Done())
	go nonCopiedCsvInformer.Run(ctx.Done())

	if ok := cache.WaitForCacheSync(ctx.Done(), namespaceInformer.HasSynced, nonCopiedCsvInformer.HasSynced); !ok {
		t.Fatalf("failed to wait for caches to sync")
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

func NewNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func NewLabeledNamespace(name string, labelValue string) *v1.Namespace {
	ns := NewNamespace(name)
	ns.SetLabels(map[string]string{
		NamespaceLabelSyncerLabelKey: labelValue,
	})
	return ns
}

func Test_SyncIgnoresDeletionEvent(t *testing.T) {
	// Sync ignores deletion events
	namespace := "test-namespace"
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceDeleted, NewCsvInNamespace(namespace))
	assert.Nil(t, plugin.Sync(context.Background(), event))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncIgnoresCopiedCsvs(t *testing.T) {
	// Sync ignores copied csvs
	namespace := "openshift-test"
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceAdded, NewCopiedCsvInNamespace(namespace))
	assert.Nil(t, plugin.Sync(context.Background(), event))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncIgnoresNonOpenshiftNamespaces(t *testing.T) {
	// Sync ignores non-openshift namespaces
	namespace := "test-namespace"
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceAdded, NewCopiedCsvInNamespace(namespace))
	assert.Nil(t, plugin.Sync(context.Background(), event))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncIgnoresPayloadOpenshiftNamespacesExceptOperators(t *testing.T) {
	// Sync ignores payload openshift namespaces, except openshift-operators
	// openshift-monitoring sync -> no label
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace("openshift-monitoring"), NewNamespace("openshift-operators")))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceAdded, NewCsvInNamespace("openshift-monitoring"))
	assert.Nil(t, plugin.Sync(context.Background(), event))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), "openshift-monitoring", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)

	// openshift-operators sync -> label added
	event = kubestate.NewResourceEvent(kubestate.ResourceAdded, NewCsvInNamespace("openshift-operators"))
	assert.Nil(t, plugin.Sync(context.Background(), event))
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
			plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewLabeledNamespace(namespace, labelValue)))
			defer shutdown()

			event := kubestate.NewResourceEvent(kubestate.ResourceUpdated, NewCsvInNamespace(namespace))
			assert.Nil(t, plugin.Sync(context.Background(), event))

			ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, labelValue, ns.GetLabels()[NamespaceLabelSyncerLabelKey])
		}()
	}
}

func Test_SyncLabelsNonPayloadUnlabeledOpenshiftNamespaces(t *testing.T) {
	// Sync will label non-labeled non-payload openshift- namespaces independent of event type (except deletion, tested separately)
	eventTypes := []kubestate.ResourceEventType{kubestate.ResourceUpdated, kubestate.ResourceAdded}
	namespace := "openshift-test"

	for _, eventType := range eventTypes {
		func() {
			plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
			defer shutdown()

			event := kubestate.NewResourceEvent(eventType, NewCsvInNamespace(namespace))
			assert.Nil(t, plugin.Sync(context.Background(), event))

			ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Contains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
		}()
	}
}

func Test_SyncFailsIfEventResourceIsNotCSV(t *testing.T) {
	// Sync fails if resource is not a csv\
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t)
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceAdded, v1.ConfigMap{})
	assert.Error(t, plugin.Sync(context.Background(), event))
}

func Test_SyncFailsIfNamespaceNotFound(t *testing.T) {
	// Sync fails if the namespace is not found
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t)
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceAdded, NewCsvInNamespace("openshift-test"))
	assert.Error(t, plugin.Sync(context.Background(), event))
}

func Test_SyncFailsIfCSVCannotBeUpdated(t *testing.T) {
	// Sync fails if the namespace cannot be updated
	namespace := "openshift-test"
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(NewNamespace(namespace)))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceAdded, NewCsvInNamespace(namespace))
	updateNsError := func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1.Namespace{}, errors.New("error updating namespace")
	}
	plugin.kubeClient.KubernetesInterface().CoreV1().(*v1fake.FakeCoreV1).PrependReactor("update", "namespaces", updateNsError)
	assert.Error(t, plugin.Sync(context.Background(), event))
}

func Test_SyncLabelsNamespaceWithCSV(t *testing.T) {
	// Given a namespace event for an unlabelled and non-payload openshift-* namespace
	// that contains at least one non-copied csv
	// Sync should apply the label syncer label to the namespace
	namespace := NewNamespace("openshift-test")
	csv := NewCsvInNamespace(namespace.GetName())
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(namespace), withExtendedResources(csv))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceUpdated, namespace)
	assert.NoError(t, plugin.Sync(context.Background(), event))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Contains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_SyncDoesNotLabelNamespaceWithoutCSVs(t *testing.T) {
	// Given a namespace event for an unlabelled and non-payload openshift-* namespace
	// that contains zero non-copied csvs
	// Sync should *NOT* apply the label syncer label to the namespace
	namespace := NewNamespace("openshift-test")
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(namespace))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceUpdated, namespace)
	assert.NoError(t, plugin.Sync(context.Background(), event))

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
	plugin, shutdown := NewFakeCSVNamespaceLabelerPlugin(t, withK8sResources(namespace), withExtendedResources(csv))
	defer shutdown()

	event := kubestate.NewResourceEvent(kubestate.ResourceUpdated, namespace)
	assert.NoError(t, plugin.Sync(context.Background(), event))

	ns, err := plugin.kubeClient.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), namespace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotContains(t, ns.GetLabels(), NamespaceLabelSyncerLabelKey)
}

func Test_OCPVersion(t *testing.T) {
	// This test is a maintenance alert that means the next OCP version is now in active development.
	// This plugin relies on a list of payload namespaces that comes from the cluster-policy-controller project
	// https://github.com/openshift/cluster-policy-controller/tree/master/pkg/psalabelsyncer
	// This list is dependent on the OCP version. Please update the dependency version to correspond to the one
	// vendored for the new OCP version (or contact the responsible team if it hasn't been updated yet).
	// Then, bump the OCP version in the `nextOCPUncutBranchName` constant below
	const nextOCPUncutBranchName = "release-4.15"
	const errorMessage = "[maintenance alert] new ocp version branch has been cut: please check comments in test for instructions"

	// Get branches
	branches, err := exec.Command("git", "branch", "-a").Output()
	assert.NoError(t, err)

	// check if the next uncut branch has been cut and fail if so
	assert.False(t, strings.Contains(string(branches), nextOCPUncutBranchName), errorMessage)
}
