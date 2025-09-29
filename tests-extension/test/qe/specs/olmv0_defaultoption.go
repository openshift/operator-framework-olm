package specs

import (
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// it is mapping to the Describe "OLM should" and "OLM optional" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 optional should", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLI("default-"+exutil.GetRandomString(), exutil.KubeConfigPath())
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		exutil.SkipNoOLMCore(oc)
	})

	g.It("PolarionID:68679-[Skipped:Disconnected]catalogsource with invalid name is created", g.Label("NonHyperShiftHOST"), func() {
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-opm.yaml")

		cs := olmv0util.CatalogSourceDescription{
			Name:        "bug-68679-4.14", // the name contains "."
			Namespace:   oc.Namespace(),
			DisplayName: "QE Operators",
			Publisher:   "QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
			Template:    csImageTemplate,
		}
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)
	})

})
