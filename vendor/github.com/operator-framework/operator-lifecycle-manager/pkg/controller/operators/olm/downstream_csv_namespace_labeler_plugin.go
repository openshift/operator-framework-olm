package olm

import (
	"context"
	"fmt"
	"strings"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/internal/pruning"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/kubestate"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/queueinformer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const labelSyncerLabelKey = "security.openshift.io/scc.podSecurityLabelSync"

// systemNSSyncExemptions is the list of namespaces deployed by an OpenShift install
// payload, as retrieved by listing the namespaces after a successful installation
// IMPORTANT: The Namespace openshift-operators must be an exception to this rule
// since it is used by OCP/OLM users to install their Operator bundle solutions.
var systemNSSyncExemptions = sets.NewString(
	// kube-specific system namespaces
	"default",
	"kube-node-lease",
	"kube-public",
	"kube-system",

	// openshift payload namespaces
	"openshift",
	"openshift-apiserver",
	"openshift-apiserver-operator",
	"openshift-authentication",
	"openshift-authentication-operator",
	"openshift-cloud-controller-manager",
	"openshift-cloud-controller-manager-operator",
	"openshift-cloud-credential-operator",
	"openshift-cloud-network-config-controller",
	"openshift-cluster-csi-drivers",
	"openshift-cluster-machine-approver",
	"openshift-cluster-node-tuning-operator",
	"openshift-cluster-samples-operator",
	"openshift-cluster-storage-operator",
	"openshift-cluster-version",
	"openshift-config",
	"openshift-config-managed",
	"openshift-config-operator",
	"openshift-console",
	"openshift-console-operator",
	"openshift-console-user-settings",
	"openshift-controller-manager",
	"openshift-controller-manager-operator",
	"openshift-dns",
	"openshift-dns-operator",
	"openshift-etcd",
	"openshift-etcd-operator",
	"openshift-host-network",
	"openshift-image-registry",
	"openshift-infra",
	"openshift-ingress",
	"openshift-ingress-canary",
	"openshift-ingress-operator",
	"openshift-insights",
	"openshift-kni-infra",
	"openshift-kube-apiserver",
	"openshift-kube-apiserver-operator",
	"openshift-kube-controller-manager",
	"openshift-kube-controller-manager-operator",
	"openshift-kube-scheduler",
	"openshift-kube-scheduler-operator",
	"openshift-kube-storage-version-migrator",
	"openshift-kube-storage-version-migrator-operator",
	"openshift-machine-api",
	"openshift-machine-config-operator",
	"openshift-marketplace",
	"openshift-monitoring",
	"openshift-multus",
	"openshift-network-diagnostics",
	"openshift-network-operator",
	"openshift-node",
	"openshift-nutanix-infra",
	"openshift-oauth-apiserver",
	"openshift-openstack-infra",
	"openshift-operator-lifecycle-manager",
	"openshift-ovirt-infra",
	"openshift-sdn",
	"openshift-service-ca",
	"openshift-service-ca-operator",
	"openshift-user-workload-monitoring",
	"openshift-vsphere-infra",
)

type CSVNamespaceLabelerPlugin struct {
	operator        *Operator
	namespaceLister v12.NamespaceLister
}

func (p *CSVNamespaceLabelerPlugin) Init(ctx context.Context, config *operatorConfig, op *Operator) error {
	if op == nil {
		return fmt.Errorf("cannot initialize plugin: operator undefined")
	}

	p.operator = op

	for _, namespace := range config.watchedNamespaces {

		// create a namespace informer for namespaces that do not include
		// the label syncer label
		namespaceInformer := informers.NewSharedInformerFactoryWithOptions(
			op.opClient.KubernetesInterface(),
			config.resyncPeriod(),
			informers.WithNamespace(namespace),
		).Core().V1().Namespaces()

		if err := op.RegisterInformer(namespaceInformer.Informer()); err != nil {
			return err
		}
		p.namespaceLister = namespaceInformer.Lister()

		// create a new csv informer and prune status to reduce memory footprint
		csvNamespaceLabelerInformer := cache.NewSharedIndexInformer(
			pruning.NewListerWatcher(
				op.client,
				namespace,
				func(opts *metav1.ListOptions) {
					// opts.LabelSelector = fmt.Sprintf("!%s", v1alpha1.CopiedLabelKey)
				},
				pruning.PrunerFunc(func(csv *v1alpha1.ClusterServiceVersion) {
					*csv = v1alpha1.ClusterServiceVersion{
						TypeMeta:   csv.TypeMeta,
						ObjectMeta: csv.ObjectMeta,
					}
				}),
			),
			&v1alpha1.ClusterServiceVersion{},
			config.resyncPeriod(),
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)

		csvNamespaceLabelerPluginQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), fmt.Sprintf("%s/csv-ns-labeler-plugin", namespace))
		csvNamespaceLabelerPluginQueueInformer, err := queueinformer.NewQueueInformer(
			ctx,
			queueinformer.WithInformer(csvNamespaceLabelerInformer),
			queueinformer.WithLogger(op.logger),
			queueinformer.WithQueue(csvNamespaceLabelerPluginQueue),
			queueinformer.WithIndexer(op.csvIndexers[namespace]),
			queueinformer.WithSyncer(p),
		)
		if err != nil {
			return err
		}
		if err := op.RegisterQueueInformer(csvNamespaceLabelerPluginQueueInformer); err != nil {
			return err
		}
	}

	return nil
}

func (p *CSVNamespaceLabelerPlugin) Sync(ctx context.Context, event kubestate.ResourceEvent) error {
	csv, ok := event.Resource().(*v1alpha1.ClusterServiceVersion)
	if !ok {
		return fmt.Errorf("event resource is not a ClusterServiceVersion")
	}

	if p.operator == nil {
		return fmt.Errorf("plugin has not been correctly initialized: operator undefined")
	}

	// only act on csv added/updated events
	if event.Type() == kubestate.ResourceDeleted {
		return nil
	}

	// ignore copied csvs
	// informer should already be filtering these out - but just in case
	if csv.IsCopied() {
		return nil
	}

	// ignore non openshift-* namespaces
	if !strings.HasPrefix(csv.GetNamespace(), "openshift-") || systemNSSyncExemptions.Has(csv.GetNamespace()) {
		return nil
	}

	namespace, err := p.namespaceLister.Get(csv.GetNamespace())
	if err != nil {
		return fmt.Errorf("error getting csv namespace (%s) for label sync'er labeling", csv.GetNamespace())
	}

	// add label sync'er label if it does not exist
	if _, ok := namespace.GetLabels()[labelSyncerLabelKey]; !ok {
		nsCopy := namespace.DeepCopy()
		if nsCopy.GetLabels() == nil {
			nsCopy.SetLabels(map[string]string{})
		}
		nsCopy.GetLabels()[labelSyncerLabelKey] = "true"
		if _, err := p.operator.opClient.KubernetesInterface().CoreV1().Namespaces().Update(ctx, nsCopy, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating csv namespace (%s) with label sync'er label", nsCopy.GetNamespace())
		}

		if p.operator.logger != nil {
			p.operator.logger.Printf("[CSV LABEL] applied %s=true label to namespace %s", labelSyncerLabelKey, nsCopy.GetNamespace())
		}
	}

	return nil
}
