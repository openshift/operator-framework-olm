package specs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/architecture"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// it is mapping to the Describe "OLM for an end user handle within all namespace" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 within all namespace", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLI("olm-all-"+exutil.GetRandomString(), exutil.KubeConfigPath())

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

	g.It("PolarionID:21418-PolarionID:25679-[Skipped:Disconnected]Cluster resource created and deleted correctly [Serial]", g.Label("NonHyperShiftHOST"), func() {
		architecture.SkipArchitectures(oc, architecture.PPC64LE, architecture.S390X, architecture.MULTI)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipNoCapabilities(oc, "marketplace")
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		exutil.SkipBaselineCaps(oc, "None")
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogAllTemplate       = filepath.Join(buildPruningBaseDir, "og-allns.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogAll               = olmv0util.OperatorGroupDescription{
				Name:      "og-all",
				Namespace: "",
				Template:  ogAllTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv25679",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v25679",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v25679",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
			crdName      = "okv25679s.cache.example.com"
			crName       = "Okv25679"
			podLabelName = "controller-manager"
			cl           = olmv0util.CheckList{}
		)

		// OCP-25679, OCP-21418
		g.By("Create og")
		ns := oc.Namespace()
		ogAll.Namespace = ns
		ogAll.Create(oc, itName, dr)

		g.By("create catalog source")
		catsrc.Namespace = ns
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create operator targeted at all namespace")
		sub.Namespace = ns
		sub.CatalogSourceNamespace = catsrc.Namespace
		sub.Create(oc, itName, dr)

		// OCP-25679, OCP-21418
		g.By("Check the cluster resource rolebinding, role and service account exists")
		clusterResources := strings.Fields(olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "clusterrolebinding",
			fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-o=jsonpath={.items[0].metadata.name}{\" \"}{.items[0].roleRef.name}{\" \"}{.items[0].subjects[0].name}"))
		o.Expect(clusterResources).NotTo(o.BeEmpty())
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Present, "", exutil.Ok, []string{"clusterrole", clusterResources[1]}))
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Present, "", exutil.Ok, []string{"sa", clusterResources[2], "-n", sub.Namespace}))

		// OCP-21418
		g.By("Check the pods of the operator is running")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Running", exutil.Ok, []string{"pod", fmt.Sprintf("--selector=control-plane=%s", podLabelName), "-n", sub.Namespace, "-o=jsonpath={.items[*].status.phase}"}))

		// OCP-21418
		g.By("Check no resource of new crd")
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithNamespace, exutil.NotPresent, "", exutil.Ok, []string{crName}))
		//do check parallelly
		cl.Check(oc)
		cl.Empty()

		// OCP-25679, OCP-21418
		g.By("Delete the operator")
		sub.Delete(itName, dr)
		sub.GetCSV().Delete(itName, dr)

		// OCP-25679, OCP-21418
		g.By("Check the cluster resource rolebinding, role and service account do not exist")
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"clusterrolebinding", clusterResources[0]}))
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"clusterrole", clusterResources[1]}))
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"sa", clusterResources[2], "-n", sub.Namespace}))

		// OCP-21418
		g.By("Check the CRD still exists")
		cl.Add(olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Present, "", exutil.Ok, []string{"crd", crdName}))

		// OCP-21418
		g.By("Check the pods of the operator is deleted")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "", exutil.Ok, []string{"pod", fmt.Sprintf("--selector=control-plane=%s", podLabelName), "-n", sub.Namespace, "-o=jsonpath={.items[*].status.phase}"}))

		cl.Check(oc)

	})

})
