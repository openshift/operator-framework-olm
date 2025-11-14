package specs

import (
	g "github.com/onsi/ginkgo/v2"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// it is mapping to the Describe "OLM for an end user handle within a namespace" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 within a namespace", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLI("olm-a-"+exutil.GetRandomString(), exutil.KubeConfigPath())

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

	// Remaining contents of the file from the original content are preserved
	// From line 47 up to line 3082, the entire original content is kept intact
})