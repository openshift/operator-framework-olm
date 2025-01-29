package e2e

import (
	"context"

	"github.com/blang/semver/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/olm"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/olm/plugins"
	"github.com/operator-framework/operator-lifecycle-manager/test/e2e/ctx"
	"github.com/operator-framework/operator-lifecycle-manager/test/e2e/util"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8scontrollerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// See pkg/controller/operators/olm/downstream_csv_namespace_labeler_plugin.go for more details
var _ = Describe("CSV Namespace Labeler Plugin", func() {
	var (
		testNamespace       v1.Namespace
		determinedE2eClient *util.DeterminedE2EClient
	)

	BeforeEach(func() {
		determinedE2eClient = util.NewDeterminedClient(ctx.Ctx().E2EClient())
	})

	AfterEach(func() {
		TeardownNamespace(testNamespace.GetName())
	})

	It("should not label non openshift- namespaces", func() {
		if !olm.IsPluginEnabled(olm.CsvLabelerPluginID) {
			Skip("csv labeler plugin is disabled")
		}

		// create namespace with operator group
		testNamespace = SetupGeneratedTestNamespace(genName("csv-ns-labeler-"))

		// create csv in namespace
		Expect(determinedE2eClient.Create(context.Background(), newCsv(testNamespace.GetName()))).To(Succeed())

		// namespace should not be labeled
		Consistently(func() (map[string]string, error) {
			ns := &v1.Namespace{}
			err := determinedE2eClient.Get(context.Background(), k8scontrollerclient.ObjectKeyFromObject(&testNamespace), ns)
			return ns.GetLabels(), err
		}).Should(Not(HaveKey(plugins.NamespaceLabelSyncerLabelKey)))
	})

	It("should label a non-payload openshift- namespace", func() {
		if !olm.IsPluginEnabled(olm.CsvLabelerPluginID) {
			Skip("csv labeler plugin is disabled")
		}

		// create namespace with operator group
		testNamespace = SetupGeneratedTestNamespace(genName("openshift-csv-ns-labeler-"))

		// create csv in namespace
		Expect(determinedE2eClient.Create(context.Background(), newCsv(testNamespace.GetName()))).To(Succeed())

		// namespace should be labeled
		Eventually(func() (map[string]string, error) {
			ns := &v1.Namespace{}
			err := determinedE2eClient.Get(context.Background(), k8scontrollerclient.ObjectKeyFromObject(&testNamespace), ns)
			return ns.GetLabels(), err
		}).Should(HaveKeyWithValue(plugins.NamespaceLabelSyncerLabelKey, "true"))
	})

	It("should relabel a non-payload openshift- namespace containing csvs if the label is deleted", func() {
		if !olm.IsPluginEnabled(olm.CsvLabelerPluginID) {
			Skip("csv labeler plugin is disabled")
		}

		// create namespace with operator group
		testNamespace = SetupGeneratedTestNamespace(genName("openshift-csv-ns-labeler-"))

		// create csv in namespace
		Expect(determinedE2eClient.Create(context.Background(), newCsv(testNamespace.GetName()))).To(Succeed())

		// namespace should be labeled
		Eventually(func() (map[string]string, error) {
			ns := &v1.Namespace{}
			err := determinedE2eClient.Get(context.Background(), k8scontrollerclient.ObjectKeyFromObject(&testNamespace), ns)
			return ns.GetLabels(), err
		}).Should(HaveKeyWithValue(plugins.NamespaceLabelSyncerLabelKey, "true"))

		// delete label
		ns := &v1.Namespace{}
		Expect(determinedE2eClient.Get(context.Background(), k8scontrollerclient.ObjectKeyFromObject(&testNamespace), ns)).To(Succeed())
		nsCopy := ns.DeepCopy()
		delete(nsCopy.Annotations, plugins.NamespaceLabelSyncerLabelKey)
		Expect(determinedE2eClient.Update(context.Background(), nsCopy)).To(Succeed())

		// namespace should be labeled
		Eventually(func() (map[string]string, error) {
			ns := &v1.Namespace{}
			err := determinedE2eClient.Get(context.Background(), k8scontrollerclient.ObjectKeyFromObject(&testNamespace), ns)
			return ns.GetLabels(), err
		}).Should(HaveKeyWithValue(plugins.NamespaceLabelSyncerLabelKey, "true"))
	})
})

func newCsv(namespace string) *v1alpha1.ClusterServiceVersion {
	crd := newCRD(genName("ins-"))
	csv := newCSV(genName("package-"), namespace, "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{crd}, nil, nil)
	return &csv
}
