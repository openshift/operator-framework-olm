package specs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/opmcli"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// it is mapping to /Users/kuiwang/GoProject/go-origin/src/github.com/openshift/openshift-tests-private/test/extended/opm/opm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 opm should", g.Label("NonHyperShiftHOST", "opm"), func() {
	defer g.GinkgoRecover()

	var (
		oc     = exutil.NewCLIWithoutNamespace("default")
		opmCLI = opmcli.NewOpmCLI()
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)

		oc.SetupProject()
		err := opmcli.EnsureOPMBinary()
		if err != nil {
			g.Skip("Failed to setup opm binary: " + err.Error())
		}
	})

	g.It("PolarionID:43185-[OTP]DC based opm subcommands out of alpha", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:43185-DC based opm subcommands out of alpha"), func() {
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

	g.It("PolarionID:43180-[OTP]opm init dc configuration package", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:43180-opm init dc configuration package"), func() {
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

	g.It("PolarionID:43171-[OTP][Skipped:Disconnected] opm render blob from bundle db based index dc based index db file and directory", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:43171-[Skipped:Disconnected] opm render blob from bundle db based index dc based index db file and directory"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("render db-based index image")
		output, err := opmCLI.Run("render").Args("quay.io/olmqe/olm-index:OLM-2199").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("\"image\": \"quay.io/olmqe/cockroachdb-operator:5.0.3-2199\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.3"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"replaces\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.4\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.5\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.5"))

		g.By("render dc-based index image with one file")
		output, err = opmCLI.Run("render").Args("quay.io/olmqe/olm-index:OLM-2199-DC-example").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("\"image\": \"quay.io/olmqe/cockroachdb-operator:5.0.3-2199\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.3"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"replaces\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.4\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.5\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.5"))

		g.By("render dc-based index image with different files")
		output, err = opmCLI.Run("render").Args("quay.io/olmqe/olm-index:OLM-2199-DC-example-Df").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("\"image\": \"quay.io/olmqe/cockroachdb-operator:5.0.3-2199\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.3"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"replaces\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.4\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.5\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.5"))

		g.By("render dc-based index image with different directory")
		output, err = opmCLI.Run("render").Args("quay.io/olmqe/olm-index:OLM-2199-DC-example-Dd").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("\"image\": \"quay.io/olmqe/cockroachdb-operator:5.0.3-2199\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.3"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"replaces\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.4\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.5\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.5"))

		g.By("render bundle image")
		output, err = opmCLI.Run("render").Args("quay.io/olmqe/cockroachdb-operator:5.0.4-2199", "quay.io/olmqe/cockroachdb-operator:5.0.3-2199").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).NotTo(o.ContainSubstring("\"name\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("\"package\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.4"))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.3"))
		o.Expect(output).To(o.ContainSubstring("\"group\": \"charts.operatorhub.io\""))
		o.Expect(output).To(o.ContainSubstring("\"version\": \"5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"version\": \"5.0.3\""))

		g.By("render directory")
		opmBaseDir := exutil.FixturePath("testdata", "opm")
		configDir := filepath.Join(opmBaseDir, "render", "configs")
		output, err = opmCLI.Run("render").Args(configDir).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb\""))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("\"image\": \"quay.io/olmqe/cockroachdb-operator:5.0.3-2199\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.3"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"cockroachdb.v5.0.4\""))
		o.Expect(output).To(o.ContainSubstring("\"replaces\": \"cockroachdb.v5.0.3\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/helmoperators/cockroachdb:v5.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.4\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.4"))
		o.Expect(output).To(o.ContainSubstring("\"name\": \"windup-operator.0.0.5\""))
		o.Expect(output).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.5"))
	})

	g.It("PolarionID:43248-[OTP]Support ignoring files when loading declarative configs", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:43248-Support ignoring files when loading declarative configs"), func() {
		opmBaseDir := exutil.FixturePath("testdata", "opm")
		correctIndex := filepath.Join(opmBaseDir, "render", "validate", "configs")
		wrongIndex := filepath.Join(opmBaseDir, "render", "validate", "configs-wrong")
		wrongIgnoreIndex := filepath.Join(opmBaseDir, "render", "validate", "configs-wrong-ignore")

		g.By("validate correct index")
		output, err := opmCLI.Run("validate").Args(correctIndex).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("%s", output)

		g.By("validate wrong index")
		output, err = opmCLI.Run("validate").Args(wrongIndex).Output()
		o.Expect(err).To(o.HaveOccurred())
		e2e.Logf("%s", output)

		g.By("validate index with ignore wrong json")
		output, err = opmCLI.Run("validate").Args(wrongIgnoreIndex).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("%s", output)
	})

	g.It("PolarionID:43768-[OTP]Improve formatting of opm validate", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:43768-Improve formatting of opm validate"), func() {
		opmBase := exutil.FixturePath("testdata", "opm")
		catalogdir := filepath.Join(opmBase, "render", "validate", "catalog")
		catalogerrdir := filepath.Join(opmBase, "render", "validate", "catalog-error")

		g.By("step: opm validate -h")
		output1, err := opmCLI.Run("validate").Args("--help").Output()
		e2e.Logf("%s", output1)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output1).To(o.ContainSubstring("opm validate "))

		g.By("opm validate catalog")
		output, err := opmCLI.Run("validate").Args(catalogdir).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.BeEmpty())

		g.By("opm validate catalog-error")
		output, err = opmCLI.Run("validate").Args(catalogerrdir).Output()
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("invalid package \\\"operator-1\\\""))
		o.Expect(output).To(o.ContainSubstring("invalid channel \\\"alpha\\\""))
		o.Expect(output).To(o.ContainSubstring("invalid bundle \\\"operator-1.v0.3.0\\\""))
		e2e.Logf("%s", output)
	})

	g.It("PolarionID:45401-[OTP]opm validate should detect cycles in channels", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:45401-opm validate should detect cycles in channels"), func() {
		opmBase := exutil.FixturePath("testdata", "opm")
		catalogerrdir := filepath.Join(opmBase, "render", "validate", "catalog-error", "operator-1")

		g.By("opm validate catalog-error/operator-1")
		output, err := opmCLI.Run("validate").Args(catalogerrdir).Output()
		if err != nil {
			e2e.Logf("%s", output)
		}
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("invalid channel \\\"45401-1\\\""))
		o.Expect(output).To(o.ContainSubstring("invalid channel \\\"45401-2\\\""))
		o.Expect(output).To(o.ContainSubstring("invalid channel \\\"45401-3\\\""))
		channelInfoList := strings.Split(output, "invalid channel")
		for _, channelInfo := range channelInfoList {
			if strings.Contains(channelInfo, "45401-1") {
				o.Expect(channelInfo).To(o.ContainSubstring("detected cycle in replaces chain of upgrade graph"))
			}
			if strings.Contains(channelInfo, "45401-2") {
				o.Expect(output).To(o.ContainSubstring("multiple channel heads found in graph"))
			}
			if strings.Contains(channelInfo, "45401-3") {
				o.Expect(output).To(o.ContainSubstring("no channel head found in graph"))
			}
		}
	})

	g.It("PolarionID:45402-[OTP][Skipped:Disconnected] opm render should automatically pulling in the images used in the deployments", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:45402-[Skipped:Disconnected] opm render should automatically pulling in the images used in the deployments"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("render bundle image")
		output, err := opmCLI.Run("render").Args("quay.io/olmqe/mta-operator:v0.0.4-45402", "quay.io/olmqe/eclipse-che:7.32.2-45402", "-oyaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("---"))
		bundleConfigBlobs := strings.Split(output, "---")
		for _, bundleConfigBlob := range bundleConfigBlobs {
			if strings.Contains(bundleConfigBlob, "packageName: mta-operator") {
				g.By("check putput of render bundle image which has no relatedimages defined in csv")
				o.Expect(bundleConfigBlob).To(o.ContainSubstring("relatedImages"))
				relatedImages := strings.Split(bundleConfigBlob, "relatedImages")[1]
				o.Expect(relatedImages).To(o.ContainSubstring("quay.io/olmqe/mta-operator:v0.0.4-45402"))
				o.Expect(relatedImages).To(o.ContainSubstring("quay.io/windupeng/windup-operator-native:0.0.4"))
				continue
			}
			if strings.Contains(bundleConfigBlob, "packageName: eclipse-che") {
				g.By("check putput of render bundle image which has relatedimages defined in csv")
				o.Expect(bundleConfigBlob).To(o.ContainSubstring("relatedImages"))
				relatedImages := strings.Split(bundleConfigBlob, "relatedImages")[1]
				o.Expect(relatedImages).To(o.ContainSubstring("index.docker.io/codercom/code-server"))
				o.Expect(relatedImages).To(o.ContainSubstring("quay.io/olmqe/eclipse-che:7.32.2-45402"))
			}
		}
	})

	g.It("PolarionID:48438-[OTP][Skipped:Disconnected] opm render should support olm.constraint which is defined in dependencies", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:48438-[Skipped:Disconnected] opm render should support olm.constraint which is defined in dependencies"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("render bundle image")
		output, err := opmCLI.Run("render").Args("quay.io/olmqe/etcd-bundle:v0.9.2-48438", "-oyaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("check output of render bundle image contain olm.constraint which is defined in dependencies.yaml")
		if !strings.Contains(output, "olm.constraint") {
			e2e.Failf("output doesn't contain olm.constraint")
		}
	})

	g.It("PolarionID:70013-[OTP][Skipped:Disconnected] opm support deprecated channel", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:70013-[Skipped:Disconnected] opm support deprecated channel"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		opmBaseDir := exutil.FixturePath("testdata", "opm", "70013")
		opmCLI.ExecCommandPath = opmBaseDir

		g.By("opm validate catalog")
		output, err := opmCLI.Run("validate").Args("catalog-valid").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.BeEmpty())

		g.By("opm validate catalog")
		output, err = opmCLI.Run("validate").Args("catalog-invalid").Output()
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("message must be set"))

		g.By("opm render")
		output, err = opmCLI.Run("render").Args("quay.io/olmqe/olmtest-operator-index:nginx70050", "-o", "yaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(strings.Contains(string(output), "schema: olm.deprecations")).To(o.BeTrue())
	})

	g.It("PolarionID:34016-[OTP]opm can prune operators from catalog", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:34016-opm can prune operators from catalog"), func() {
		opmBaseDir := exutil.FixturePath("testdata", "opm")
		indexDB := filepath.Join(opmBaseDir, "index_34016.db")
		output, err := opmCLI.Run("registry").Args("prune", "-d", indexDB, "-p", "lib-bucket-provisioner").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(output, "deleting packages") || !strings.Contains(output, "pkg=planetscale") {
			e2e.Failf("Failed to obtain the removed packages from prune : %s", output)
		}
	})

	g.It("PolarionID:54168-[OTP][Skipped:Disconnected] opm support '--use-http' global flag", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:54168-[Skipped:Disconnected] opm support '--use-http' global flag"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		if os.Getenv("HTTP_PROXY") != "" || os.Getenv("http_proxy") != "" {
			g.Skip("HTTP_PROXY is not empty - skipping test ...")
		}
		opmBaseDir := exutil.FixturePath("testdata", "opm", "53869")
		opmCLI.ExecCommandPath = opmBaseDir

		g.By("1) checking alpha list")
		output, err := opmCLI.Run("alpha").Args("list", "bundles", "quay.io/openshifttest/nginxolm-operator-index:v1", "--use-http").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(output, "nginx-operator") {
			e2e.Failf("Failed to obtain the packages from alpha list : %s", output)
		}

		g.By("2) checking render")
		output, err = opmCLI.Run("render").Args("quay.io/openshifttest/nginxolm-operator-index:v1", "--use-http").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(output, "nginx-operator") {
			e2e.Failf("Failed run render command : %s", output)
		}

		g.By("3) checking index add")
		output, err = opmCLI.Run("index").Args("add", "-b", "quay.io/openshifttest/nginxolm-operator-bundle:v0.0.1", "-t", "quay.io/olmqe/nginxolm-operator-index:v54168", "--use-http", "--generate").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(output, "writing dockerfile") {
			e2e.Failf("Failed run render command : %s", output)
		}

		g.By("4) checking render-veneer semver")
		output, err = opmCLI.Run("alpha").Args("render-template", "--use-http", "basic", filepath.Join(opmBaseDir, "catalog-basic-template.yaml"), "-o", "yaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(output, "nginx-operator") {
			e2e.Failf("Failed run render command : %s", output)
		}

		g.By("5) checking render-graph")
		output, err = opmCLI.Run("alpha").Args("render-graph", "quay.io/openshifttest/nginxolm-operator-index:v1", "--use-http").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(output, "nginx-operator") {
			e2e.Failf("Failed run render-graph command : %s", output)
		}
	})

	g.It("PolarionID:43409-[OTP][Skipped:Disconnected] opm can list catalog contents", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:43409-[Skipped:Disconnected] opm can list catalog contents"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		imagetag := "quay.io/olmqe/nginxolm-operator-index:v1"
		g.By("1, list packages")
		output, err := opmCLI.Run("alpha").Args("list", "packages", imagetag).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator"))

		g.By("2, list channels")
		output, err = opmCLI.Run("alpha").Args("list", "channels", imagetag).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator"))

		g.By("3, list channels in a package")
		output, err = opmCLI.Run("alpha").Args("list", "channels", imagetag, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("alpha"))

		g.By("4, list bundles")
		output, err = opmCLI.Run("alpha").Args("list", "bundles", imagetag).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator.v0.0.1"))

		g.By("5, list bundles in a package")
		output, err = opmCLI.Run("alpha").Args("list", "bundles", imagetag, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator.v0.0.1"))

		g.By("step: SUCCESS")
	})

	g.It("PolarionID:45407-[OTP]opm and oc should print sqlite deprecation warnings", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:45407-opm and oc should print sqlite deprecation warnings"), func() {
		g.By("opm render --help")
		output, err := opmCLI.Run("render").Args("--help").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("DEPRECATION NOTICE:"))

		g.By("opm index --help")
		output, err = opmCLI.Run("index").Args("--help").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("DEPRECATION NOTICE:"))

		g.By("opm registry --help")
		output, err = opmCLI.Run("registry").Args("--help").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("DEPRECATION NOTICE:"))

		g.By("oc adm catalog mirror --help")
		output, err = oc.AsAdmin().WithoutNamespace().Run("adm").Args("catalog", "mirror", "--help").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("DEPRECATION NOTICE:"))
	})

	g.It("PolarionID:53869-[OTP][Skipped:Disconnected] opm supports creating a catalog using basic veneer", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:53869-[Skipped:Disconnected] opm supports creating a catalog using basic veneer"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		if os.Getenv("HTTP_PROXY") != "" || os.Getenv("http_proxy") != "" {
			g.Skip("HTTP_PROXY is not empty - skipping test ...")
		}
		opmBaseDir := exutil.FixturePath("testdata", "opm", "53869")
		opmCLI.ExecCommandPath = opmBaseDir

		g.By("step: create dir catalog")
		catsrcPathYaml := filepath.Join(opmBaseDir, "catalog-yaml")
		err = os.MkdirAll(catsrcPathYaml, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("step: create a catalog using basic veneer with yaml format")
		output, err := opmCLI.Run("alpha").Args("render-template", "basic", filepath.Join(opmBaseDir, "catalog-basic-template.yaml"), "-o", "yaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator"))

		indexFilePath := filepath.Join(catsrcPathYaml, "index.yaml")
		if err = os.WriteFile(indexFilePath, []byte(output), 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("validate").Args(catsrcPathYaml).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		output, err = opmCLI.Run("alpha").Args("list", "bundles", catsrcPathYaml).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("quay.io/olmqe/nginxolm-operator-bundle:v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("quay.io/olmqe/nginxolm-operator-bundle:v1.0.1"))

		g.By("step: create dir catalog")
		catsrcPathJSON := filepath.Join(opmBaseDir, "catalog-json")
		err = os.MkdirAll(catsrcPathJSON, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("step: create a catalog using basic veneer with json format")
		output, err = opmCLI.Run("alpha").Args("render-template", "basic", filepath.Join(opmBaseDir, "catalog-basic-template.yaml"), "-o", "json").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		indexFilePath = filepath.Join(catsrcPathJSON, "index.json")
		if err = os.WriteFile(indexFilePath, []byte(output), 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("validate").Args(catsrcPathJSON).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		output, err = opmCLI.Run("alpha").Args("list", "bundles", catsrcPathJSON, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("quay.io/olmqe/nginxolm-operator-bundle:v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("quay.io/olmqe/nginxolm-operator-bundle:v1.0.1"))
	})

	g.It("PolarionID:53871-PolarionID:53915-PolarionID:53996-[OTP][Skipped:Disconnected][Slow] opm supports creating a catalog using semver veneer [Slow]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:53871-PolarionID:53915-PolarionID:53996-[Skipped:Disconnected][Slow] opm supports creating a catalog using semver veneer [Slow]"), func() {
		if os.Getenv("HTTP_PROXY") != "" || os.Getenv("http_proxy") != "" {
			g.Skip("HTTP_PROXY is not empty - skipping test ...")
		}
		errPolicy := opmcli.EnsureContainerPolicy()
		o.Expect(errPolicy).NotTo(o.HaveOccurred())
		opmBaseDir := exutil.FixturePath("testdata", "opm", "53871")
		opmCLI.ExecCommandPath = opmBaseDir

		g.By("step: create dir catalog-1")
		catsrcPath1 := filepath.Join(opmBaseDir, "catalog-1")
		err := os.MkdirAll(catsrcPath1, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("step: GenerateMajorChannels: true GenerateMinorChannels: false")
		output, err := opmCLI.Run("alpha").Args("render-template", "semver", filepath.Join(opmBaseDir, "catalog-semver-veneer-1.yaml"), "-o", "yaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		indexFilePath := filepath.Join(catsrcPath1, "index.yaml")
		if err = os.WriteFile(indexFilePath, []byte(output), 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("validate").Args(catsrcPath1).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		output, err = opmCLI.Run("alpha").Args("list", "channels", catsrcPath1, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v0  nginx-operator.v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v1  nginx-operator.v1.0.2"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v2  nginx-operator.v2.1.0"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v0       nginx-operator.v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v2       nginx-operator.v2.1.0"))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v1     nginx-operator.v1.0.2"))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v2     nginx-operator.v2.1.0"))

		g.By("step: create dir catalog-2")
		catsrcPath2 := filepath.Join(opmBaseDir, "catalog-2")
		err = os.MkdirAll(catsrcPath2, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("step: GenerateMajorChannels: true GenerateMinorChannels: true")
		output, err = opmCLI.Run("alpha").Args("render-template", "semver", filepath.Join(opmBaseDir, "catalog-semver-veneer-2.yaml"), "-o", "yaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		indexFilePath = filepath.Join(catsrcPath2, "index.yaml")
		if err = os.WriteFile(indexFilePath, []byte(output), 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("validate").Args(catsrcPath2).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		output, err = opmCLI.Run("alpha").Args("list", "channels", catsrcPath2, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v0    nginx-operator.v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v0.0  nginx-operator.v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v1    nginx-operator.v1.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v1.0  nginx-operator.v1.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v1         nginx-operator.v1.0.1-beta"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v1.0       nginx-operator.v1.0.1-beta"))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v1       nginx-operator.v1.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v1.0     nginx-operator.v1.0.1"))

		g.By("step: create dir catalog-3")
		catsrcPath3 := filepath.Join(opmBaseDir, "catalog-3")
		err = os.MkdirAll(catsrcPath3, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("step: not set GenerateMajorChannels and GenerateMinorChannels")
		output, err = opmCLI.Run("alpha").Args("render-template", "semver", filepath.Join(opmBaseDir, "catalog-semver-veneer-3.yaml")).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		indexFilePath = filepath.Join(catsrcPath3, "index.json")
		if err = os.WriteFile(indexFilePath, []byte(output), 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("validate").Args(catsrcPath3).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		output, err = opmCLI.Run("alpha").Args("list", "channels", catsrcPath3, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).NotTo(o.ContainSubstring("candidate-v0 "))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v0.0  nginx-operator.v0.0.1"))
		o.Expect(string(output)).NotTo(o.ContainSubstring("candidate-v1 "))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v1.0  nginx-operator.v1.0.2"))
		o.Expect(string(output)).NotTo(o.ContainSubstring("fast-v0 "))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v0.0       nginx-operator.v0.0.1"))
		o.Expect(string(output)).NotTo(o.ContainSubstring("fast-v2 "))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v2.0       nginx-operator.v2.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v2.1       nginx-operator.v2.1.0"))
		o.Expect(string(output)).NotTo(o.ContainSubstring("stable-v2 "))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v2.1     nginx-operator.v2.1.0"))

		g.By("step: generate mermaid graph data for generated-channels")
		output, err = opmCLI.Run("alpha").Args("render-template", "semver", filepath.Join(opmBaseDir, "catalog-semver-veneer-4.yaml"), "-o", "mermaid").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("channel \"fast-v2.0\""))
		o.Expect(string(output)).To(o.ContainSubstring("subgraph nginx-operator-fast-v2.0[\"fast-v2.0\"]"))
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator-fast-v2.0-nginx-operator.v2.0.1[\"nginx-operator.v2.0.1\"]"))

		g.By("step: semver veneer should validate bundle versions")
		output, err = opmCLI.Run("alpha").Args("render-template", "semver", filepath.Join(opmBaseDir, "catalog-semver-veneer-5.yaml")).Output()
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("encountered bundle versions which differ only by build metadata, which cannot be ordered"))
		o.Expect(string(output)).To(o.ContainSubstring("cannot be compared to \"1.0.1-alpha\""))

		g.By("OCP-53996")
		filePath := filepath.Join(opmBaseDir, "catalog-semver-veneer-1.yaml")
		g.By("step: create dir catalog")
		catsrcPath53996 := filepath.Join(opmBaseDir, "catalog-53996")
		err = os.MkdirAll(catsrcPath53996, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("step: generate index.yaml with yaml format")
		command := "cat " + filePath + "| opm alpha render-template semver -o yaml - "
		contentByte, err := exec.Command("bash", "-c", command).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		indexFilePath = filepath.Join(catsrcPath53996, "index.yaml")
		if err = os.WriteFile(indexFilePath, contentByte, 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("validate").Args(catsrcPath53996).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		output, err = opmCLI.Run("alpha").Args("list", "channels", catsrcPath53996, "nginx-operator").Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v0  nginx-operator.v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v1  nginx-operator.v1.0.2"))
		o.Expect(string(output)).To(o.ContainSubstring("candidate-v2  nginx-operator.v2.1.0"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v0       nginx-operator.v0.0.1"))
		o.Expect(string(output)).To(o.ContainSubstring("fast-v2       nginx-operator.v2.1.0"))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v1     nginx-operator.v1.0.2"))
		o.Expect(string(output)).To(o.ContainSubstring("stable-v2     nginx-operator.v2.1.0"))

		g.By("step: generate json format file")
		command = "cat " + filePath + `| opm alpha render-template semver  - | jq 'select(.schema=="olm.channel")'| jq '{name,entries}'`
		contentByte, err = exec.Command("bash", "-c", command).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(contentByte)).To(o.ContainSubstring("nginx-operator.v1.0.2"))
		o.Expect(string(contentByte)).To(o.ContainSubstring("candidate-v1"))

		g.By("step: generate mermaid graph data for generated-channels")
		command = "cat " + filePath + "| opm alpha render-template semver -o mermaid -"
		contentByte, err = exec.Command("bash", "-c", command).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(contentByte)).To(o.ContainSubstring("package \"nginx-operator\""))
		o.Expect(string(contentByte)).To(o.ContainSubstring("nginx-operator-candidate-v1-nginx-operator.v1.0.1"))
	})

	g.It("PolarionID:53917-[OTP][Skipped:Disconnected] opm can visualize the update graph for a given Operator from an arbitrary version", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:53917-[Skipped:Disconnected] opm can visualize the update graph for a given Operator from an arbitrary version"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		if os.Getenv("HTTP_PROXY") != "" || os.Getenv("http_proxy") != "" {
			g.Skip("HTTP_PROXY is not empty - skipping test ...")
		}

		g.By("step: check help message")
		output, err := opmCLI.Run("alpha").Args("render-graph", "-h").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("--minimum-edge"))
		o.Expect(string(output)).To(o.ContainSubstring("--use-http"))

		g.By("step: opm alpha render-graph index-image")
		output, err = opmCLI.Run("alpha").Args("render-graph", "quay.io/olmqe/nginxolm-operator-index:v1").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("package \"nginx-operator\""))
		o.Expect(string(output)).To(o.ContainSubstring("subgraph nginx-operator-alpha[\"alpha\"]"))

		g.By("step: opm alpha render-graph index-image with --minimum-edge")
		output, err = opmCLI.Run("alpha").Args("render-graph", "quay.io/olmqe/nginxolm-operator-index:v1", "--minimum-edge", "nginx-operator.v1.0.1").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("package \"nginx-operator\""))
		o.Expect(string(output)).To(o.ContainSubstring("nginx-operator.v1.0.1"))
		o.Expect(string(output)).NotTo(o.ContainSubstring("nginx-operator.v0.0.1"))

		g.By("step: create dir catalog")
		catsrcPath := filepath.Join("/tmp", "53917-catalog")
		defer func() {
			_ = os.RemoveAll(catsrcPath)
		}()
		errCreateDir := os.MkdirAll(catsrcPath, 0755)
		o.Expect(errCreateDir).NotTo(o.HaveOccurred())

		g.By("step: opm alpha render-graph fbc-dir")
		output, err = opmCLI.Run("render").Args("quay.io/olmqe/nginxolm-operator-index:v1", "-o", "json").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		indexFilePath := filepath.Join(catsrcPath, "index.json")
		if err = os.WriteFile(indexFilePath, []byte(output), 0644); err != nil {
			e2e.Failf("Writefile %s Error: %v", indexFilePath, err)
		}
		output, err = opmCLI.Run("alpha").Args("render-graph", catsrcPath).Output()
		if err != nil {
			e2e.Logf("%s", output)
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(string(output)).To(o.ContainSubstring("package \"nginx-operator\""))
		o.Expect(string(output)).To(o.ContainSubstring("subgraph nginx-operator-alpha[\"alpha\"]"))

		g.By("step: opm alpha render-graph sqlit-based catalog image")
		output, err = opmCLI.Run("alpha").Args("render-graph", "quay.io/olmqe/ditto-index:v1beta1").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("subgraph \"ditto-operator\""))
		o.Expect(string(output)).To(o.ContainSubstring("subgraph ditto-operator-alpha[\"alpha\"]"))
		o.Expect(string(output)).To(o.ContainSubstring("ditto-operator.v0.2.0"))
	})

	g.It("PolarionID:60573-[OTP][Skipped:Disconnected] opm exclude bundles with olm.deprecated property when rendering", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:60573-[Skipped:Disconnected] opm exclude bundles with olm.deprecated property when rendering"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("opm render the sqlite index image message")
		msg, err := opmCLI.Run("render").Args("quay.io/olmqe/catalogtest-index:v4.12depre", "-oyaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if strings.Contains(msg, "olm.deprecated") {
			e2e.Failf("opm render the sqlite index image message, doesn't show the bundle with olm.dreprecated label")
		}

		g.By("opm render the fbc index image message")
		msg, err = opmCLI.Run("render").Args("quay.io/olmqe/test-index:mix", "-oyaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if strings.Contains(msg, "olm.deprecated") {
			e2e.Logf("opm render the fbc index image message, still show the bundle with olm.dreprecated label")
		} else {
			e2e.Failf("opm render the fbc index image message, should show the bundle with olm.dreprecated label")
		}
	})

	g.It("PolarionID:73218-[OTP][Skipped:Disconnected] opm alpha render-graph indicate deprecated graph content", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 opm should PolarionID:73218-[Skipped:Disconnected] opm alpha render-graph indicate deprecated graph content"), func() {
		err := opmcli.EnsureContainerPolicy()
		o.Expect(err).NotTo(o.HaveOccurred())

		if os.Getenv("HTTP_PROXY") != "" || os.Getenv("http_proxy") != "" {
			g.Skip("HTTP_PROXY is not empty - skipping test ...")
		}

		g.By("step: opm alpha render-graph index-image with deprecated label")
		output, err := opmCLI.Run("alpha").Args("render-graph", "quay.io/olmqe/olmtest-operator-index:nginxolm73218").Output()
		if err != nil && strings.Contains(output, "failed to pull image") {
			g.Skip("Skipping test: failed to pull image")
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(output)).To(o.ContainSubstring("classDef deprecated"))
		o.Expect(string(output)).To(o.ContainSubstring("deprecated"))
		o.Expect(string(output)).To(o.ContainSubstring("nginx73218-candidate-v1.0-nginx73218.v1.0.1[\"nginx73218.v1.0.1\"]-- skip --> nginx73218-candidate-v1.0-nginx73218.v1.0.3[\"nginx73218.v1.0.3\"]"))
	})

})
