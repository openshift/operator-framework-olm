package specs

import (
	"path/filepath"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/architecture"
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

	g.It("PolarionID:24870-[Skipped:Disconnected]can not create csv without operator group", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

			og = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-index:OLM-2378-Oadp-GoodOne-withCache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "oadp-operator",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "oadp-operator",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create csv with failure because of no operator group")
		sub.CurrentCSV = "oadp-operator.v0.5.3"
		sub.CreateWithoutCheck(oc, itName, dr)
		olmv0util.NewCheck("present", exutil.AsUser, exutil.WithNamespace, exutil.NotPresent, "", exutil.Ok, []string{"csv", sub.CurrentCSV}).Check(oc)
		sub.Delete(itName, dr)

		g.By("Create opertor group and then csv is created with success")
		og.Create(oc, itName, dr)
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded"+"InstallSucceeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}{.status.reason}"}).Check(oc)
	})

	g.It("PolarionID:37263-[Skipped:Disconnected][Skipped:Proxy]Subscription stays in UpgradePending but InstallPlan not installing [Slow]", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
		platform := exutil.CheckPlatform(oc)
		if strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || strings.Contains(platform, "none") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "olm-1860185-catalog",
				Namespace:   "",
				DisplayName: "OLM 1860185 Catalog",
				Publisher:   "QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv37263",
				Template:    catsrcImageTemplate,
			}
			catsrc1 = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-index:OLM-2378-Oadp-GoodOne-withCache",
				Template:    catsrcImageTemplate,
			}
			catsrc2 = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-nginx-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:v1399-fbc",
				Template:    catsrcImageTemplate,
			}
			subStrimzi = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v37263",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v37263",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			subBuildv2 = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok1-1399",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok1-1399",
				CatalogSourceName:      catsrc2.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok1-1399.v0.0.4",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			subMta = olmv0util.SubscriptionDescription{
				SubName:                "oadp-operator",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "oadp-operator",
				CatalogSourceName:      catsrc1.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		catsrc1.Namespace = oc.Namespace()
		catsrc2.Namespace = oc.Namespace()
		subStrimzi.Namespace = oc.Namespace()
		subStrimzi.CatalogSourceNamespace = catsrc.Namespace
		subBuildv2.Namespace = oc.Namespace()
		subBuildv2.CatalogSourceNamespace = catsrc2.Namespace
		subMta.Namespace = oc.Namespace()
		subMta.CatalogSourceNamespace = catsrc1.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)
		catsrc1.CreateWithCheck(oc, itName, dr)
		catsrc2.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install Strimzi")
		subStrimzi.Create(oc, itName, dr)

		g.By("check if Strimzi is installed")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subStrimzi.InstalledCSV, "-n", subStrimzi.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("install Portworx")
		subMta.Create(oc, itName, dr)

		g.By("check if Portworx is installed")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subMta.InstalledCSV, "-n", subMta.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("get IP of Portworx")
		mtaIP := subMta.GetIP(oc)

		g.By("Delete Portworx sub")
		subMta.Delete(itName, dr)

		g.By("check if Portworx sub is Deleted")
		olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"sub", subMta.SubName, "-n", subMta.Namespace}).Check(oc)

		g.By("Delete Portworx csv")
		csvPortworx := olmv0util.CsvDescription{
			Name:      subMta.InstalledCSV,
			Namespace: subMta.Namespace,
		}
		csvPortworx.Delete(itName, dr)

		g.By("check if Portworx csv is Deleted")
		olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"csv", subMta.InstalledCSV, "-n", subMta.Namespace}).Check(oc)

		g.By("install Couchbase")
		subBuildv2.Create(oc, itName, dr)

		g.By("get IP of Couchbase")
		couchbaseIP := subBuildv2.GetIP(oc)

		g.By("it takes different IP")
		o.Expect(couchbaseIP).NotTo(o.Equal(mtaIP))

	})

})
