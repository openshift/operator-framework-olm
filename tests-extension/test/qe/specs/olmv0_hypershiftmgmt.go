package specs

import (
	"context"
	"path/filepath"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// it is mapping to the Describe "OLM on hypershift" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 on hypershift mgmt", g.Label("NonHyperShiftHOST"), func() {
	defer g.GinkgoRecover()

	var (
		oc                                                  = exutil.NewCLIWithoutNamespace("default")
		guestClusterName, guestClusterKube, hostedClusterNS string
		isAKS                                               bool
		errIsAKS                                            error
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		oc.SetupProject()
		if !exutil.IsHypershiftMgmtCluster(oc) {
			g.Skip("this is not a hypershift management cluster, skip test run")
		}

		isAKS, errIsAKS = exutil.IsAKSCluster(context.TODO(), oc)
		if errIsAKS != nil {
			g.Skip("can not determine if it is openshift cluster or aks cluster")
		}
		if !isAKS {
			exutil.SkipNoOLMCore(oc)
		}

		err := exutil.EnsureHypershiftBinary(oc)
		if err != nil {
			g.Skip("Failed to setup hypershift binary: " + err.Error())
		}

		guestClusterName, guestClusterKube, hostedClusterNS = exutil.ValidHypershiftAndGetGuestKubeConf(oc)
		e2e.Logf("%s, %s, %s", guestClusterName, guestClusterKube, hostedClusterNS)
		oc.SetGuestKubeconf(guestClusterKube)
	})

	g.It("PolarionID:45381-[OTP][Skipped:Disconnected]Support custom catalogs in hypershift", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 on hypershift mgmt PolarionID:45381-[Skipped:Disconnected]Support custom catalogs in hypershift"), func() {
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

	g.It("PolarionID:45408-[OTP][Skipped:Disconnected]Eliminate use of imagestreams in catalog management", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 on hypershift mgmt PolarionID:45408-[Skipped:Disconnected]Eliminate use of imagestreams in catalog management"), func() {
		controlProject := hostedClusterNS + "-" + guestClusterName
		if !isAKS {
			exutil.SkipBaselineCaps(oc, "None")
			g.By("1) check if uses the ImageStream resource")
			isOutput, err := oc.AsAdmin().Run("get").Args("is", "catalogs", "-n", controlProject, "-o", "yaml").Output()
			if err != nil {
				e2e.Failf("Fail to get cronjob in project: %s, error:%v", controlProject, err)
			}
			is := []string{"certified-operators", "community-operators", "redhat-marketplace", "redhat-operators"}
			for _, imageStream := range is {
				if !strings.Contains(isOutput, imageStream) {
					e2e.Failf("find ImageStream:%s in project:%v", imageStream, controlProject)
				}
			}
		}

		g.By("2) check if Deployment uses the ImageStream")
		deploys := []string{"certified-operators-catalog", "community-operators-catalog", "redhat-marketplace-catalog", "redhat-operators-catalog"}
		for _, deploy := range deploys {
			annotations, err := oc.AsAdmin().Run("get").Args("deployment", "-n", controlProject, deploy, "-o=jsonpath={.metadata.annotations}").Output()
			if err != nil {
				e2e.Failf("Fail to get deploy:%s in project: %s, error:%v", deploy, controlProject, err)
			}
			if !isAKS {
				if !strings.Contains(strings.ToLower(annotations), "imagestream") {
					e2e.Failf("The deploy does not use ImageStream: %v", annotations)
				}
			} else {
				if strings.Contains(strings.ToLower(annotations), "imagestream") {
					e2e.Failf("The deploy does not use ImageStream: %v", annotations)
				}
			}
		}
	})
	// Polarion ID: 45543
	g.It("PolarionID:45543-[OTP][Skipped:Disconnected]Enable hypershift to deploy OperatorLifecycleManager resources", func() {

		g.By("1, check if any resource running in the guest cluster")
		projects := []string{"openshift-operator-lifecycle-manager", "openshift-marketplace"}
		for _, project := range projects {
			resource, err := oc.AsGuestKubeconf().Run("get").Args("pods", "-n", project).Output()
			if err != nil {
				e2e.Failf("Fail to get resource in project: %s, error:%v", project, err)
			}
			// now, for guest cluster, there is may have a custom catalog resource for testing
			if project == "openshift-marketplace" && strings.Contains(resource, "marketplace-operator") {
				e2e.Failf("Found Marketplace related resources running on the guest cluster")
			}
			if project != "openshift-marketplace" && !strings.Contains(resource, "No resources found") {
				e2e.Failf("Found OLM related resources running on the guest cluster")
			}
		}

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("2, create an OperatorGroup")
		ns := "guest-cluster-45543"
		err := oc.AsGuestKubeconf().Run("create").Args("ns", ns).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		defer func() {
			_ = oc.AsGuestKubeconf().Run("delete").Args("ns", ns).Execute()
		}()
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-45543",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		defer og.Delete(itName, dr)
		og.CreateWithCheck(oc.AsGuestKubeconf(), itName, dr)

		g.By("3, create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-45543",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc.AsGuestKubeconf(), itName, dr)

		g.By("4, subscribe to learn-operator.v0.0.3")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-45543",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-45543",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		sub.Create(oc.AsGuestKubeconf(), itName, dr)
		defer sub.DeleteCSV(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc.AsGuestKubeconf())
	})
})
