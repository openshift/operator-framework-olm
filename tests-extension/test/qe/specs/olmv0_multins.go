package specs

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// it is mapping to the Describe "OLM for an end user handle to support" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 with multi ns", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLIWithoutNamespace("default")

		dr = make(olmv0util.DescriberResrouce)
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		oc.SetupProject()

		exutil.SkipNoOLMCore(oc)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
	})

	g.AfterEach(func() {})

	g.It("PolarionID:22226-[OTP][Skipped:Disconnected]the csv without support MultiNamespace fails for og with MultiNamespace", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 with multi ns PolarionID:22226-[Skipped:Disconnected]the csv without support MultiNamespace fails for og with MultiNamespace"), func() {
		var (
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			cmNcTemplate        = filepath.Join(buildPruningBaseDir, "cm-namespaceconfig.yaml")
			catsrcCmTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
			ogMultiTemplate     = filepath.Join(buildPruningBaseDir, "og-multins.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			itName              = g.CurrentSpecReport().FullText()
			og                  = olmv0util.OperatorGroupDescription{
				Name:         "og-multinamespace",
				Namespace:    "",
				Multinslabel: "olmtestmultins",
				Template:     ogMultiTemplate,
			}
			cm = olmv0util.ConfigMapDescription{
				Name:      "cm-community-namespaceconfig-operators",
				Namespace: "", //must be set in iT
				Template:  cmNcTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-community-namespaceconfig-operators",
				Namespace:   "", //must be set in iT
				DisplayName: "Community namespaceconfig Operators",
				Publisher:   "Community",
				SourceType:  "configmap",
				Address:     "cm-community-namespaceconfig-operators",
				Template:    catsrcCmTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "namespace-configuration-operator",
				Namespace:              "", //must be set in iT
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "namespace-configuration-operator",
				CatalogSourceName:      "catsrc-community-namespaceconfig-operators",
				CatalogSourceNamespace: "", //must be set in iT
				StartingCSV:            "",
				CurrentCSV:             "namespace-configuration-operator.v0.1.0", //it matches to that in cm, so set it.
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			p1 = olmv0util.ProjectDescription{
				Name:            "olm-enduser-multins-csv-1-fail",
				TargetNamespace: "",
			}
			p2 = olmv0util.ProjectDescription{
				Name:            "olm-enduser-multins-csv-2-fail",
				TargetNamespace: "",
			}
		)

		defer p1.Delete(oc)
		defer p2.Delete(oc)
		cm.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()
		p1.TargetNamespace = oc.Namespace()
		p2.TargetNamespace = oc.Namespace()
		g.By("Create new project")
		p1.Create(oc, itName, dr)
		p1.Label(oc, "olmtestmultins")
		p2.Create(oc, itName, dr)
		p2.Label(oc, "olmtestmultins")

		g.By("Create cm")
		cm.Create(oc, itName, dr)

		g.By("Create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create sub")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "MultiNamespace InstallModeType not supported", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.message}"}).Check(oc)
	})

	g.It("PolarionID:71119-[OTP]pod does not start for installing operator of multi-ns mode when og is in one of the ns[Serial][Slow]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 with multi ns PolarionID:71119-[Skipped:Disconnected]pod does not start for installing operator of multi-ns mode when og is in one of the ns[Serial]"), func() {
		exutil.SkipForSNOCluster(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipNoCapabilities(oc, "marketplace")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") ||
			strings.Contains(platform, "ibmcloud") || strings.Contains(platform, "nutanix") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		olmv0util.ValidateAccessEnvironment(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogMultiTemplate     = filepath.Join(buildPruningBaseDir, "og-multins.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			catsrc              = olmv0util.CatalogSourceDescription{
				Name:        "olm-71119-catalog",
				Namespace:   "",
				DisplayName: "OLM 71119 Catalog",
				Publisher:   "QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv71119",
				Template:    catsrcImageTemplate,
			}
			og = olmv0util.OperatorGroupDescription{
				Name:         "og-71119",
				Namespace:    "test-ns71119-1",
				Multinslabel: "label-71119",
				Template:     ogMultiTemplate,
			}
			p1 = olmv0util.ProjectDescription{
				Name:            "test-ns71119-1",
				TargetNamespace: "test-ns71119-1",
			}
			p2 = olmv0util.ProjectDescription{
				Name:            "test-ns71119-2",
				TargetNamespace: "test-ns71119-1",
			}
			subSample = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v71119",
				Namespace:              "test-ns71119-1",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: p1.Name,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v71119",
				Template:               subTemplate,
			}
		)

		g.By("create two ns and og")
		defer p1.Delete(oc)
		p1.Create(oc, itName, dr)
		p1.Label(oc, "label-71119")
		defer p2.Delete(oc)
		p2.Create(oc, itName, dr)
		p2.Label(oc, "label-71119")
		og.Create(oc, itName, dr)
		catsrc.Namespace = p1.Name
		catsrc.Create(oc, itName, dr)

		g.By("subscribe to operator with multinamespaces mode")
		defer subSample.Delete(itName, dr)
		subSample.Create(oc, itName, dr)
		defer subSample.GetCSV().Delete(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subSample.InstalledCSV, "-n", subSample.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		podName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", subSample.Namespace, "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(podName).NotTo(o.BeEmpty())

		o.Consistently(func() int {
			restartCount, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", podName, "-n", subSample.Namespace, "-o=jsonpath={.status..restartCount}").Output()
			if strings.Contains(restartCount, "NotFound") {
				return 0
			}
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(restartCount).NotTo(o.BeEmpty())
			count, err := strconv.Atoi(strings.Fields(restartCount)[0])
			o.Expect(err).NotTo(o.HaveOccurred())
			return count
		}, 150*time.Second, 10*time.Second).Should(o.Equal(0), "the pod restart")
	})

	g.It("PolarionID:29275-[OTP][Skipped:Disconnected]label to target namespace of operator group with multi namespace", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 with multi ns PolarionID:29275-[Skipped:Disconnected]label to target namespace of operator group with multi namespace"), func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogMultiTemplate     = filepath.Join(buildPruningBaseDir, "og-multins.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:         "og-1651-1",
				Namespace:    "",
				Multinslabel: "test-og-label-1651",
				Template:     ogMultiTemplate,
			}
			p1 = olmv0util.ProjectDescription{
				Name:            "test-ns1651-1",
				TargetNamespace: "",
			}
			p2 = olmv0util.ProjectDescription{
				Name:            "test-ns1651-2",
				TargetNamespace: "",
			}
		)

		p1.TargetNamespace = oc.Namespace()
		p2.TargetNamespace = oc.Namespace()
		og.Namespace = oc.Namespace()
		g.By("Create new projects and label them")
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "ns", p1.Name, "--ignore-not-found")
		}()
		err := oc.AsAdmin().WithoutNamespace().Run("create").Args("ns", p1.Name).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		p1.Label(oc, "test-og-label-1651")
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "ns", p2.Name, "--ignore-not-found")
		}()
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("ns", p2.Name).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		p2.Label(oc, "test-og-label-1651")

		g.By("Create og and check the label")
		og.Create(oc, itName, dr)
		ogUID := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithNamespace, "og", og.Name, "-o=jsonpath={.metadata.uid}")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+ogUID, exutil.Ok, []string{"ns", p1.Name, "-o=jsonpath={.metadata.labels}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+ogUID, exutil.Ok, []string{"ns", p2.Name, "-o=jsonpath={.metadata.labels}"}).Check(oc)

		g.By("delete og and check there is no label")
		og.Delete(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+ogUID, exutil.Nok, []string{"ns", p1.Name, "-o=jsonpath={.metadata.labels}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+ogUID, exutil.Nok, []string{"ns", p2.Name, "-o=jsonpath={.metadata.labels}"}).Check(oc)

		g.By("create another og to check the label")
		og.Name = "og-1651-2"
		og.Create(oc, itName, dr)
		ogUID = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithNamespace, "og", og.Name, "-o=jsonpath={.metadata.uid}")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+ogUID, exutil.Ok, []string{"ns", p1.Name, "-o=jsonpath={.metadata.labels}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+ogUID, exutil.Ok, []string{"ns", p2.Name, "-o=jsonpath={.metadata.labels}"}).Check(oc)
	})

})
