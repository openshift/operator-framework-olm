package specs

import (
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/opmcli"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// it is mapping to /Users/kuiwang/GoProject/go-origin/src/github.com/openshift/openshift-tests-private/test/extended/opm/opm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 opm should", g.Label("NonHyperShiftHOST"), func() {
	defer g.GinkgoRecover()

	var (
		oc     = exutil.NewCLIForKubeOpenShift("opm-" + exutil.GetRandomString())
		opmCLI = opmcli.NewOpmCLI()
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)

		err := opmcli.EnsureOPMBinary()
		if err != nil {
			g.Skip("Failed to setup opm binary: " + err.Error())
		}
	})

	g.It("PolarionID:43185-DC based opm subcommands out of alpha", func() {
		g.By("check init, serve, render and validate under opm")
		output, err := opmCLI.Run("").Args("--help").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("OPM help output: %s", output)
		o.Expect(output).To(o.ContainSubstring("init "))
		o.Expect(output).To(o.ContainSubstring("serve "))
		o.Expect(output).To(o.ContainSubstring("render "))
		o.Expect(output).To(o.ContainSubstring("validate "))

		g.By("check init, serve, render and validate not under opm alpha")
		output, err = opmCLI.Run("alpha").Args("--help").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("OPM help output: %s", output)
		o.Expect(output).NotTo(o.ContainSubstring("init "))
		o.Expect(output).NotTo(o.ContainSubstring("serve "))
		o.Expect(output).NotTo(o.ContainSubstring("render "))
		o.Expect(output).NotTo(o.ContainSubstring("validate "))
	})

	g.It("PolarionID:43180-opm init dc configuration package", func() {
		g.By("init package")
		opmBaseDir := exutil.FixturePath("testdata", "opm")
		readme := filepath.Join(opmBaseDir, "render", "init", "readme.md")
		testpng := filepath.Join(opmBaseDir, "render", "init", "test.png")

		output, err := opmCLI.Run("init").Args("--default-channel=alpha", "-d", readme, "-i", testpng, "mta-operator").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("OPM init output: %s", output)
		o.Expect(output).To(o.ContainSubstring("\"schema\": \"olm.package\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"mta-operator\""))
		o.Expect(output).To(o.ContainSubstring("\"defaultChannel\": \"alpha\""))
		o.Expect(output).To(o.ContainSubstring("zcfHkVw9GfpbJmeev9F08WW8uDkaslwX6avlWGU6N"))
		o.Expect(output).To(o.ContainSubstring("\"description\": \"it is testing\""))

	})

})
