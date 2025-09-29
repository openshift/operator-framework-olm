package specs

import (
	"context"
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// it is mapping to the Describe "OLM on hypershift" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMvo on hypershift mgmt", g.Label("NonHyperShiftHOST"), func() {
	defer g.GinkgoRecover()

	var (
		oc                                                  = exutil.NewCLIForKubeOpenShift("hypershiftmgmt-" + exutil.GetRandomString())
		guestClusterName, guestClusterKube, hostedClusterNS string
		isAKS                                               bool
		errIsAKS                                            error
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		isAKS, errIsAKS = exutil.IsAKSCluster(context.TODO(), oc)
		if errIsAKS != nil {
			g.Skip("can not determine if it is openshift cluster or aks cluster")
		}
		if !isAKS {
			exutil.SkipNoOLMCore(oc)
		}
		guestClusterName, guestClusterKube, hostedClusterNS = exutil.ValidHypershiftAndGetGuestKubeConf(oc)
		e2e.Logf("%s, %s, %s", guestClusterName, guestClusterKube, hostedClusterNS)
		oc.SetGuestKubeconf(guestClusterKube)
	})

	g.It("ROSA-OSD_CCS-HyperShiftMGMT-ConnectedOnly-Author:kuiwang-Medium-45381-Support custom catalogs in hypershift", func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-2378-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 2378 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-index:OLM-2378-Oadp-GoodOne-multi",
				Template:    catsrcImageTemplate,
			}
			subOadp = olmv0util.SubscriptionDescription{
				SubName:                "oadp-operator",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "oadp-operator",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "oadp-operator.v0.5.3",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			dr = make(olmv0util.DescriberResrouce)
		)

		g.By("init resource")
		dr.AddIr(itName)
		ns := "guest-cluster-45381"
		err := oc.AsGuestKubeconf().Run("create").Args("ns", ns).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		defer func() {
			_ = oc.AsGuestKubeconf().Run("delete").Args("ns", ns).Execute()
		}()
		og.Namespace = ns
		catsrc.Namespace = ns
		subOadp.Namespace = ns
		subOadp.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc.AsGuestKubeconf(), itName, dr)

		g.By("Create og")
		defer og.Delete(itName, dr)
		og.Create(oc.AsGuestKubeconf(), itName, dr)

		g.By("install OADP")
		defer subOadp.Delete(itName, dr)
		subOadp.Create(oc.AsGuestKubeconf(), itName, dr)
		defer subOadp.DeleteCSV(itName, dr)

		g.By("Check the oadp-operator.v0.5.3 is installed successfully")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subOadp.InstalledCSV, "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc.AsGuestKubeconf())

	})

})
