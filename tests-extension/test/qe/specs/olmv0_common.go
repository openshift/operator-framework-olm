package specs

import (
	"context"
	"os/exec"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0", func() {
	g.It("should pass a trivial sanity check", g.Label("ReleaseGate"), func(ctx context.Context) {
		o.Expect(len("test")).To(o.BeNumerically(">", 0))
	})
})

// it is mapping to the Describe "OLM for an end user handle common object" and "OLM for an end user use" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 should", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLI("olm-common-"+exutil.GetRandomString(), exutil.KubeConfigPath())

		dr = make(olmv0util.DescriberResrouce)
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		exutil.SkipNoOLMCore(oc)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
	})

	g.AfterEach(func() {
		itName := g.CurrentSpecReport().FullText()
		dr.GetIr(itName).Cleanup()
		dr.RmIr(itName)
	})

	g.It("PolarionID:22259-[Skipped:Disconnected]marketplace operator CR status on a running cluster[Serial]", g.Label("NonHyperShiftHOST"), func() {

		exutil.SkipForSNOCluster(oc)
		exutil.SkipNoCapabilities(oc, "marketplace")
		g.By("check marketplace status")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "TrueFalseFalse", exutil.Ok, []string{"clusteroperator", "marketplace",
			"-o=jsonpath={.status.conditions[?(@.type==\"Available\")].status}{.status.conditions[?(@.type==\"Progressing\")].status}{.status.conditions[?(@.type==\"Degraded\")].status}"}).Check(oc)

		g.By("delete pod of marketplace")
		_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "pod", "--selector=name=marketplace-operator",
			"-n", "openshift-marketplace", "--ignore-not-found")
		o.Expect(err).NotTo(o.HaveOccurred())

		_, _ = exec.Command("bash", "-c", "sleep 10").Output()

		g.By("pod of marketplace restart")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "TrueFalseFalse", exutil.Ok, []string{"clusteroperator", "marketplace",
			"-o=jsonpath={.status.conditions[?(@.type==\"Available\")].status}{.status.conditions[?(@.type==\"Progressing\")].status}{.status.conditions[?(@.type==\"Degraded\")].status}"}).Check(oc)
	})

	g.It("PolarionID:73695-[Skipped:Disconnected]PO is disable", func() {

		if !exutil.IsTechPreviewNoUpgrade(oc) {
			g.Skip("PO is supported in TP only currently, so skip it")
		}
		_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("co", "platform-operators-aggregated").Output()
		o.Expect(err).To(o.HaveOccurred(), "PO is not disable")
	})

})
