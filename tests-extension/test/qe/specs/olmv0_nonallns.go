package specs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"

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

	g.It("PolarionID:24870-[OTP][Skipped:Disconnected]can not create csv without operator group", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:24870-[Skipped:Disconnected]can not create csv without operator group"), func() {
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

	g.It("PolarionID:22200-[OTP][Skipped:Disconnected]add minimum kube version to CSV [Slow]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:22200-[Skipped:Disconnected]add minimum kube version to CSV [Slow]"), func() {
		checkArch := architecture.ClusterArchitecture(oc)
		e2e.Logf("the curent arch is %v", checkArch.String())
		architecture.SkipNonAmd64SingleArch(oc)
		e2e.Logf("done for SkipNonAmd64SingleArch and try the following method which is same to SkipNonAmd64SingleArch")
		architecture.SkipArchitectures(oc, architecture.PPC64LE, architecture.S390X, architecture.MULTI, architecture.ARM64, architecture.UNKNOWN)
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for ask cluster")
		}
		exutil.SkipNoCapabilities(oc, "marketplace")
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		infra, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "none") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			cmNcTemplate        = filepath.Join(buildPruningBaseDir, "cm-namespaceconfig.yaml")
			catsrcCmTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogTemplate,
			}
			cmNc = olmv0util.ConfigMapDescription{
				Name:      "cm-community-namespaceconfig-operators",
				Namespace: "", //must be set in iT
				Template:  cmNcTemplate,
			}
			catsrcNc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-community-namespaceconfig-operators",
				Namespace:   "", //must be set in iT
				DisplayName: "Community namespaceconfig Operators",
				Publisher:   "Community",
				SourceType:  "configmap",
				Address:     "cm-community-namespaceconfig-operators",
				Template:    catsrcCmTemplate,
			}
			subNc = olmv0util.SubscriptionDescription{
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
			cm     = cmNc
			catsrc = catsrcNc
			sub    = subNc
		)

		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		cm.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create configmap of csv")
		cm.Create(oc, itName, dr)

		g.By("Get minKubeVersionRequired and kubeVersionUpdated")
		output := olmv0util.GetResource(oc, exutil.AsUser, exutil.WithoutNamespace, "cm", cm.Name, "-n", cm.Namespace, "-o=json")
		csvDesc := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(strings.SplitN(output, "\"clusterServiceVersions\": ", 2)[1], "\"customResourceDefinitions\":", 2)[0]), ",")
		o.Expect(strings.Contains(csvDesc, "minKubeVersion:")).To(o.BeTrue())
		minKubeVersionRequired := strings.TrimSpace(strings.SplitN(strings.SplitN(csvDesc, "minKubeVersion:", 2)[1], "\\n", 2)[0])
		kubeVersionUpdated := olmv0util.GenerateUpdatedKubernatesVersion(oc)
		e2e.Logf("the kubeVersionUpdated version is %s, and minKubeVersionRequired is %s", kubeVersionUpdated, minKubeVersionRequired)

		g.By("Create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Update the minKubeVersion greater than the cluster KubeVersion")
		cm.Patch(oc, fmt.Sprintf("{\"data\": {\"clusterServiceVersions\": %s}}", strings.ReplaceAll(csvDesc, "minKubeVersion: "+minKubeVersionRequired, "minKubeVersion: "+kubeVersionUpdated)))

		g.By("Create sub with greater KubeVersion")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "not met+2+less than", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.requirementStatus[?(@.kind==\"ClusterServiceVersion\")].message}"}).Check(oc)

		g.By("Remove sub and csv and update the minKubeVersion to orignl")
		sub.Delete(itName, dr)
		sub.GetCSV().Delete(itName, dr)
		cm.Patch(oc, fmt.Sprintf("{\"data\": {\"clusterServiceVersions\": %s}}", csvDesc))

		g.By("Create sub with orignal KubeVersion")
		sub.Create(oc, itName, dr)
		err := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
			csvPhase := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Contains(csvPhase, "Succeeded") {
				e2e.Logf("sub is installed")
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			msg := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.requirementStatus[?(@.kind==\"ClusterServiceVersion\")].message}")
			if strings.Contains(msg, "CSV version requirement not met") && !strings.Contains(msg, kubeVersionUpdated) {
				e2e.Failf("the csv can not be installed with correct kube version")
			}
		}
	})

	g.It("PolarionID:23473-[OTP][Skipped:Disconnected]permit z-stream releases skipping during operator updates", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:23473-[Skipped:Disconnected]permit z-stream releases skipping during operator updates"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			cmNcTemplate        = filepath.Join(buildPruningBaseDir, "cm-namespaceconfig.yaml")
			catsrcCmTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogTemplate,
			}
			skippedVersion = "namespace-configuration-operator.v0.0.2"
			cmNc           = olmv0util.ConfigMapDescription{
				Name:      "cm-community-namespaceconfig-operators",
				Namespace: "", //must be set in iT
				Template:  cmNcTemplate,
			}
			catsrcNc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-community-namespaceconfig-operators",
				Namespace:   "", //must be set in iT
				DisplayName: "Community namespaceconfig Operators",
				Publisher:   "Community",
				SourceType:  "configmap",
				Address:     "cm-community-namespaceconfig-operators",
				Template:    catsrcCmTemplate,
			}
			subNc = olmv0util.SubscriptionDescription{
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
			cm     = cmNc
			catsrc = catsrcNc
			sub    = subNc
		)

		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		cm.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create configmap of csv")
		cm.Create(oc, itName, dr)

		g.By("Create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create sub")
		sub.IpApproval = "Manual"
		sub.StartingCSV = "namespace-configuration-operator.v0.0.1"
		sub.Create(oc, itName, dr)

		g.By("manually approve sub")
		sub.Approve(oc, itName, dr)

		g.By(fmt.Sprintf("there is skipped csv version %s", skippedVersion))
		o.Expect(strings.Contains(sub.IpCsv, skippedVersion)).To(o.BeFalse())
	})

	g.It("PolarionID:37263-[OTP][Skipped:Disconnected][Skipped:Proxy]Subscription stays in UpgradePending but InstallPlan not installing [Slow]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:37263-[Skipped:Disconnected][Skipped:Proxy]Subscription stays in UpgradePending but InstallPlan not installing [Slow]"), func() {
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

	// It will cover test case: OCP-29231 and OCP-29277
	g.It("PolarionID:29231-PolarionID:29277-[OTP]label to target namespace of group", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:29231-PolarionID:29277-label to target namespace of group"), func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			og1                 = olmv0util.OperatorGroupDescription{
				Name:      "og1-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			og2 = olmv0util.OperatorGroupDescription{
				Name:      "og2-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		og1.Namespace = oc.Namespace()
		og2.Namespace = oc.Namespace()

		g.By("Create og1 and check the label of target namespace of og1 is created")
		og1.Create(oc, itName, dr)
		og1Uid := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithNamespace, "og", og1.Name, "-o=jsonpath={.metadata.uid}")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+og1Uid, exutil.Ok,
			[]string{"ns", og1.Namespace, "-o=jsonpath={.metadata.labels}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+og1Uid, exutil.Nok,
			[]string{"ns", "openshift-operators", "-o=jsonpath={.metadata.labels}"}).Check(oc)

		g.By("Delete og1 and check the label of target namespace of og1 is removed")
		og1.Delete(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+og1Uid, exutil.Nok,
			[]string{"ns", og1.Namespace, "-o=jsonpath={.metadata.labels}"}).Check(oc)

		g.By("Create og2 and recreate og1 and check the label")
		og2.Create(oc, itName, dr)
		og2Uid := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithNamespace, "og", og2.Name, "-o=jsonpath={.metadata.uid}")
		og1.Create(oc, itName, dr)
		og1Uid = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithNamespace, "og", og1.Name, "-o=jsonpath={.metadata.uid}")
		labelNs := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "ns", og1.Namespace, "-o=jsonpath={.metadata.labels}")
		o.Expect(labelNs).To(o.ContainSubstring(og2Uid))
		o.Expect(labelNs).To(o.ContainSubstring(og1Uid))

		// OCP-29277
		g.By("Check no label of global operator group ")
		globalOgUID := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "og", "global-operators", "-n", "openshift-operators", "-o=jsonpath={.metadata.uid}")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "olm.operatorgroup.uid/"+globalOgUID, exutil.Nok,
			[]string{"ns", "default", "-o=jsonpath={.metadata.labels}"}).Check(oc)

	})

	// Group 2: OCP-25855
	g.It("PolarionID:25855-[OTP][Skipped:Disconnected][Serial]Add the channel field to subscription_sync_count", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:25855-[Skipped:Disconnected][Serial]Add the channel field to subscription_sync_count"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
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

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create operator")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("get information of catalog operator pod")
		output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "pods", "-l", "app=catalog-operator", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.items[0].metadata.name}{\" \"}{.items[0].status.podIP}{\":\"}{.items[0].spec.containers[0].ports[?(@.name==\"metrics\")].containerPort}")
		o.Expect(output).NotTo(o.BeEmpty())
		infoCatalogOperator := strings.Fields(output)

		g.By("check the subscription_sync_total")
		var subscriptionSyncTotal []byte
		var errExec error
		err = wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
			// Get prometheus service account token
			token, errToken := exutil.GetPrometheusSAToken(oc)
			if errToken != nil {
				e2e.Logf("failed to get prometheus token: %v", errToken)
				return false, nil
			}
			token = strings.TrimSpace(token)

			// Use the token in curl command
			curlCmd := fmt.Sprintf("oc exec -c catalog-operator %s -n openshift-operator-lifecycle-manager -- curl -s -k -H 'Authorization: Bearer %s' https://%s/metrics", infoCatalogOperator[0], token, infoCatalogOperator[1])
			subscriptionSyncTotal, errExec = exec.Command("bash", "-c", curlCmd).Output()
			if !strings.Contains(string(subscriptionSyncTotal), sub.InstalledCSV) {
				e2e.Logf("the metric is not counted and try next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			e2e.Logf("the output: %v \n the err: %v", string(subscriptionSyncTotal), errExec)
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("csv %s is not included in metric", sub.InstalledCSV))
	})

	// Group 2: OCP-23170
	g.It("PolarionID:23170-[OTP][Skipped:Disconnected]API labels should be hash", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:23170-[Skipped:Disconnected]API labels should be hash"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogD                 = olmv0util.OperatorGroupDescription{
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
				Address:     "quay.io/olmqe/nginx-ok-index:vokv23170",
				Template:    catsrcImageTemplate,
			}
			subD = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v23170",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v23170",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}

			og  = ogD
			sub = subD
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need defer or AfterEach
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create operator")
		sub.Create(oc, itName, dr)

		g.By("Check the API labels should be hash")
		apiLabels := olmv0util.GetResource(oc, exutil.AsUser, exutil.WithNamespace, "csv", sub.InstalledCSV, "-o=jsonpath={.metadata.labels}")
		o.Expect(len(apiLabels)).NotTo(o.BeZero())
		pattern, err := regexp.Compile(`^[a-fA-F0-9]{16}$|^[a-fA-F0-9]{15}$`)
		o.Expect(err).NotTo(o.HaveOccurred())
		for _, v := range strings.Split(strings.Trim(apiLabels, "{}"), ",") {
			if strings.Contains(v, "olm.api") {
				hash := strings.Trim(strings.Split(strings.Split(v, ":")[0], ".")[2], "\"")
				// calling regexp.MatchString in a loop has poor performance, consider using regexp.Compile (SA6000)
				// match, err := regexp.MatchString(`^[a-fA-F0-9]{16}$|^[a-fA-F0-9]{15}$`, hash)
				// o.Expect(err).NotTo(o.HaveOccurred())
				// o.Expect(match).To(o.BeTrue())
				res := pattern.Find([]byte(hash))
				o.Expect(string(res)).NotTo(o.BeEmpty())
			}
		}
	})

	// Group 3: OCP-20979
	g.It("PolarionID:20979-[OTP][Skipped:Disconnected]only one IP is generated", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:20979-[Skipped:Disconnected]only one IP is generated"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for aks cluster")
		}
		exutil.SkipNoCapabilities(oc, "marketplace")
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "none") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogD                 = olmv0util.OperatorGroupDescription{
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
				Address:     "quay.io/olmqe/nginx-ok-index:vokv20979",
				Template:    catsrcImageTemplate,
			}
			subD = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v20979",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v20979",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			og  = ogD
			sub = subD
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need defer or AfterEach
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			status, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status..lastObservedState}").Output()
			if strings.Compare(status, "READY") != 0 {
				e2e.Logf("catsrc %s lastObservedState is %s, not READY", catsrc.Name, status)
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status}").Output()
			e2e.Logf("catsrc status: %s", output)
			pods, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", catsrc.Namespace).Output()
			e2e.Logf("Pods in namespace %s: %s", catsrc.Namespace, pods)
			events, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("events", "-n", catsrc.Namespace).Output()
			e2e.Logf("Events in namespace %s: %s", catsrc.Namespace, events)
			g.Skip("catsrc is not ready, so skip")
		}

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create operator")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("Check there is only one ip")
		ips := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", "-n", sub.Namespace, "--no-headers")
		ipList := strings.Split(ips, "\n")
		for _, ip := range ipList {
			name := strings.Fields(ip)[0]
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", name, "-n", sub.Namespace, "-o=json")
		}
		o.Expect(strings.Count(ips, sub.InstalledCSV)).To(o.Equal(1))
	})

	// Group 3: OCP-25757/22656
	g.It("PolarionID:25757-PolarionID:22656-[OTP][Skipped:Disconnected]manual approval strategy apply to subsequent releases", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:25757-PolarionID:22656-[Skipped:Disconnected]manual approval strategy apply to subsequent releases"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for aks cluster")
		}
		exutil.SkipNoCapabilities(oc, "marketplace")
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		infra, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "none") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogD                 = olmv0util.OperatorGroupDescription{
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
				Address:     "quay.io/olmqe/olm-index:OLM-2378-Oadp-Good",
				Template:    catsrcImageTemplate,
			}
			subD = olmv0util.SubscriptionDescription{
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

			og  = ogD
			sub = subD
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need defer or AfterEach
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("prepare for manual approval")
		sub.IpApproval = "Manual"
		sub.StartingCSV = "oadp-operator.v0.5.5"

		g.By("Create Sub which apply manual approve install plan")
		sub.Create(oc, itName, dr)

		g.By("the install plan is RequiresApproval")
		installPlan := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installplan.name}")
		o.Expect(installPlan).NotTo(o.BeEmpty())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "RequiresApproval", exutil.Ok, []string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("manually approve sub")
		sub.Approve(oc, itName, dr)

		g.By("the target CSV is created with upgrade")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 150*time.Second, false, func(ctx context.Context) (bool, error) {
			currentCSV := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.currentCSV}")
			if strings.Compare(currentCSV, sub.StartingCSV) != 0 {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("the installedCSV %v is not expected", sub.InstalledCSV))
	})

	// Group 4 - OCP-24438 + OCP-24027
	g.It("PolarionID:24438-[OTP][Skipped:Disconnected]check subscription CatalogSource Status", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:24438-[Skipped:Disconnected]check subscription CatalogSource Status"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		platform := exutil.CheckPlatform(oc)
		if strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || strings.Contains(platform, "none") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogD                 = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-test-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "",
				Template:    catsrcImageTemplate,
			}
			subD = olmv0util.SubscriptionDescription{
				SubName:                "oadp-operator",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "oadp-operator",
				CatalogSourceName:      "test",
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}

			og  = ogD
			sub = subD
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()

		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("create sub with the above catalogsource")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("check its condition is UnhealthyCatalogSourceFound")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "UnhealthyCatalogSourceFound", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].reason}"}).Check(oc)

		g.By("create catalogsource")
		catsrc.Address = "quay.io/olmqe/olm-index:OLM-2378-Oadp-GoodOne-withCache"
		catsrc.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)
	})

	g.It("PolarionID:24027-[OTP][Skipped:Disconnected]can create and delete catalogsource and sub repeatedly", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:24027-[Skipped:Disconnected]can create and delete catalogsource and sub repeatedly"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		platform := exutil.CheckPlatform(oc)
		if strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || strings.Contains(platform, "none") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogD                 = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			subD = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v24027",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v24027",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok-v24027.v0.0.1",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-test-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv24027",
				Template:    catsrcImageTemplate,
			}
			repeatedCount = 2
			og            = ogD
			sub           = subD
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()

		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("Create og")
		og.Create(oc, itName, dr)

		for i := 0; i < repeatedCount; i++ {
			g.By("Create Catalogsource")
			catsrc.Create(oc, itName, dr)
			olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

			g.By("Create sub with the above catalogsource")
			sub.Create(oc, itName, dr)
			olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)

			g.By("Remove catalog and sub")
			sub.Delete(itName, dr)
			sub.DeleteCSV(itName, dr)
			catsrc.Delete(itName, dr)
			if i < repeatedCount-1 {
				time.Sleep(20 * time.Second)
			}
		}
	})

	// OCP-21404 - CSV will be RequirementsNotMet after SA is deleted
	g.It("PolarionID:21404-[OTP][Skipped:Disconnected]csv will be RequirementsNotMet after sa is delete", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:21404-[Skipped:Disconnected]csv will be RequirementsNotMet after sa is delete"), func() {
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for ask cluster")
		}
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || strings.Contains(platform, "none") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogD                 = olmv0util.OperatorGroupDescription{
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

			subD = olmv0util.SubscriptionDescription{
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
			og  = ogD
			sub = subD
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create operator")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("Get SA of csv")
		olmv0util.GetResource(oc, exutil.AsUser, exutil.WithNamespace, "csv", sub.InstalledCSV, "-o=json")
		sa := olmv0util.NewSa(strings.Fields(olmv0util.GetResource(oc, exutil.AsUser, exutil.WithNamespace, "csv", sub.InstalledCSV, "-o=jsonpath={.status.requirementStatus[?(@.kind==\"ServiceAccount\")].name}"))[0], sub.Namespace)

		g.By("Delete sa of csv")
		sa.GetDefinition(oc)
		sa.Delete(oc)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "RequirementsNotMet", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.reason}"}).Check(oc)

		g.By("Recovery sa of csv")
		sa.Reapply(oc)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded+2+Installing", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	// OCP-29723 - As cluster admin find abnormal status condition via components of operator resource
	g.It("PolarionID:29723-[OTP][Skipped:Disconnected]As cluster admin find abnormal status condition via components of operator resource", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:29723-[Skipped:Disconnected]As cluster admin find abnormal status condition via components of operator resource"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
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
				Name:        "catsrc-29723-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 29723 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:v1399-fbc-multi",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok1-1399",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok1-1399",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok1-1399.v0.0.4",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install perator")
		sub.Create(oc, itName, dr)

		g.By("delete catalog source")
		catsrc.Delete(itName, dr)
		g.By("delete sa")
		_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "sa", "nginx-ok1-1399-controller-manager", "-n", sub.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("check abnormal status")
		output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", sub.OperatorPackage+"."+sub.Namespace, "-o=json")
		o.Expect(output).NotTo(o.BeEmpty())

		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "CatalogSourcesUnhealthy", exutil.Ok, []string{"operator.operators.coreos.com", sub.OperatorPackage + "." + sub.Namespace,
			fmt.Sprintf("-o=jsonpath={.status.components.refs[?(@.name==\"%s\")].conditions[*].type}", sub.SubName)}).Check(oc)

		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "RequirementsNotMet+2+InstallWaiting", exutil.Ok, []string{"operator.operators.coreos.com", sub.OperatorPackage + "." + sub.Namespace,
			fmt.Sprintf("-o=jsonpath={.status.components.refs[?(@.name==\"%s\")].conditions[*].reason}", sub.InstalledCSV)}).Check(oc)
	})

	// OCP-30762 - installs bundles with v1 CRDs
	g.It("PolarionID:30762-[OTP][Skipped:Disconnected]installs bundles with v1 CRDs", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:30762-[Skipped:Disconnected]installs bundles with v1 CRDs"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		e2e.Logf("platform: %v", platform)
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "none") ||
			strings.Contains(platform, "vsphere") || strings.Contains(platform, "osp") || strings.Contains(platform, "ibmcloud") || strings.Contains(platform, "nutanix") ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" ||
			exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-30762-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 30762 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv30762",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v30762",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v30762",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok-v30762.v0.0.1",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install perator")
		sub.Create(oc, itName, dr)

		g.By("check csv")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	// OCP-27683 - InstallPlans can install from extracted bundles
	g.It("PolarionID:27683-[OTP][Skipped:Disconnected]InstallPlans can install from extracted bundles", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:27683-[Skipped:Disconnected]InstallPlans can install from extracted bundles"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-27683-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 27683 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv27683",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v27683",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v27683",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok-v27683.v0.0.1",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install perator")
		sub.Create(oc, itName, dr)

		g.By("check csv")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("get bundle package from ip")
		installPlan := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installplan.name}")
		o.Expect(installPlan).NotTo(o.BeEmpty())
		ipBundle := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.bundleLookups[0].path}")
		o.Expect(ipBundle).NotTo(o.BeEmpty())

		g.By("get bundle package from job")
		jobName := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "job", "-n", catsrc.Namespace, "-o=jsonpath={.items[0].metadata.name}")
		o.Expect(jobName).NotTo(o.BeEmpty())
		jobBundle := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "pod", "-l", "job-name="+jobName, "-n", catsrc.Namespace, "-o=jsonpath={.items[0].status.initContainerStatuses[*].image}")
		o.Expect(jobName).NotTo(o.BeEmpty())
		o.Expect(jobBundle).To(o.ContainSubstring(ipBundle))
	})

	g.It("PolarionID:24513-[OTP][Skipped:Disconnected]Operator config support env only", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:24513-[Skipped:Disconnected]Operator config support env only"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
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
				Name:        "catsrc-24513-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 24513 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:v1399-1-arg",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok1-1399",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok1-1399",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok1-1399.v0.0.5",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install operator")
		sub.Create(oc, itName, dr)

		g.By("check csv")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("get parameter of deployment")
		olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", "-n", sub.Namespace, "-o=yaml")

		g.By("patch env for sub")
		sub.Patch(oc, "{\"spec\": {\"config\": {\"env\": [{\"name\": \"EMPTY_ENV\"},{\"name\": \"ARGS1\",\"value\": \"-v=4\"}]}}}")

		g.By("check the empty env")
	})

	g.It("PolarionID:24382-[OTP][Skipped:Disconnected]Should restrict CRD update if schema changes[Serial]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:24382-[Skipped:Disconnected]Should restrict CRD update if schema changes[Serial]"), func() {
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for aks cluster")
		}
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
		node, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(err).NotTo(o.HaveOccurred())
		efips, err := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if err != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		platform := exutil.CheckPlatform(oc)
		e2e.Logf("platform: %v", platform)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-legacy.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			etcdCluster         = filepath.Join(buildPruningBaseDir, "etcd-cluster.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-24382-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 24382 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-dep:vschema-crdv3",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "etcd",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "etcdoperator.v0.9.2",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			etcdCr = olmv0util.CustomResourceDescription{
				Name:      "example-24382",
				Namespace: "",
				TypeName:  "EtcdCluster",
				Template:  etcdCluster,
			}
		)
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		etcdCr.Namespace = oc.Namespace()
		defer func() { _ = exutil.RecoverNamespaceRestricted(oc, oc.Namespace()) }()
		err = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install operator")
		sub.Create(oc, itName, dr)

		g.By("check csv")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		errCRD := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("crd", "etcdclusters.etcd.database.coreos.com", "-o=jsonpath={.status.storedVersions}").Output()
			if err != nil {
				return false, err
			}
			if strings.Contains(output, "v1beta2") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(errCRD, "crd etcdcluster does not exist")

		g.By("create cr")
		etcdCr.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Running", exutil.Ok, []string{etcdCr.TypeName, etcdCr.Name, "-n", etcdCr.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("update operator")
		sub.Patch(oc, "{\"spec\": {\"channel\": \"beta\"}}")
		sub.FindInstalledCSV(oc, itName, dr)

		errIP := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.currentCSV}").Output()
			if err != nil {
				return false, err
			}
			if strings.Contains(output, "etcdoperator.v0.9.4") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(errIP, "operator does not change to etcdoperator.v0.9.4")

		g.By("check schema does not work")
		installPlan := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installplan.name}")
		o.Expect(installPlan).NotTo(o.BeEmpty())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "error validating existing CRs", exutil.Ok, []string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}"}).Check(oc)
	})

	// Group 8: OCP-25760 + OCP-35895
	g.It("PolarionID:25760-[OTP][Skipped:Disconnected]Operator upgrades does not fail after change the channel", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:25760-[Skipped:Disconnected]Operator upgrades does not fail after change the channel"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipForSNOCluster(oc)
		platform := exutil.CheckPlatform(oc)
		if strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || strings.Contains(platform, "none") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		node, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(err).NotTo(o.HaveOccurred())
		efips, errFips := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errFips != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
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
				Name:        "catsrc-25760-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 25760 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv25760",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v25760",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v25760",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok-v25760.v0.0.1",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install operator")
		sub.Create(oc, itName, dr)

		g.By("check csv")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("switch channel")
		sub.Patch(oc, "{\"spec\": {\"channel\": \"beta\"}}")
		sub.FindInstalledCSV(oc, itName, dr)

		g.By("check csv of new channel")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:35895-[OTP][Skipped:Disconnected]can't install a CSV with duplicate roles", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:35895-[Skipped:Disconnected]can't install a CSV with duplicate roles"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		exutil.SkipForSNOCluster(oc)
		platform := exutil.CheckPlatform(oc)
		e2e.Logf("platform: %v", platform)
		proxy, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "none") ||
			strings.Contains(platform, "vsphere") || strings.Contains(platform, "osp") || strings.Contains(platform, "ibmcloud") || strings.Contains(platform, "nutanix") ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" ||
			exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-35895-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 35895 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-dep:vmtaduprol2-withCache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "mta-operator",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "mta-operator",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "windup-operator.0.0.5",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install operator")
		sub.Create(oc, itName, dr)

		g.By("check csv")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("check sa")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "windup-operator-haproxy", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={..serviceAccountName}"}).Check(oc)
	})

	// Group 9: OCP-32863
	g.It("PolarionID:32863-[OTP][Skipped:Disconnected]Support resources required for SAP Gardener Control Plane Operator[Disruptive]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:32863-[Skipped:Disconnected]Support resources required for SAP Gardener Control Plane Operator[Serial][Disruptive]"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
		platform := exutil.CheckPlatform(oc)
		e2e.Logf("platform: %v", platform)
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		proxy, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "none") ||
			strings.Contains(platform, "vsphere") || strings.Contains(platform, "external") || strings.Contains(platform, "osp") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}

		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			vpaTemplate         = filepath.Join(buildPruningBaseDir, "vpa-crd.yaml")
			crdVpa              = olmv0util.CrdDescription{
				Name:     "verticalpodautoscalers.autoscaling.k8s.io",
				Template: vpaTemplate,
			}
			og = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-32863-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 32863 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/single-bundle-index:pdb3",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "busybox",
				Namespace:              "",
				Channel:                "candidate-v2",
				IpApproval:             "Automatic",
				OperatorPackage:        "busybox",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "busybox.v2.0.0",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)

		// defer crdVpa.Delete(oc) //it is not needed in case it already exist
		if olmv0util.IsPresentResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "crd", crdVpa.Name) {

			oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
			og.Namespace = oc.Namespace()
			catsrc.Namespace = oc.Namespace()
			sub.Namespace = oc.Namespace()
			sub.CatalogSourceNamespace = catsrc.Namespace

			g.By("create vpa crd")
			crdVpa.Create(oc, itName, dr)
			defer crdVpa.Delete(oc)

			g.By("create catalog source")
			catsrc.CreateWithCheck(oc, itName, dr)

			g.By("Create og")
			og.Create(oc, itName, dr)

			g.By("install operator")
			sub.Create(oc, itName, dr)

			g.By("check csv")
			err := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
				status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
				if strings.Compare(status, "Succeeded") == 0 {
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, "csv busybox.v2.0.0 is not installed as expected")

			g.By("check additional resources")
			olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Present, "", exutil.Ok, []string{"VerticalPodAutoscaler", "busybox-vpa", "-n", sub.Namespace}).Check(oc)
			olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Present, "", exutil.Ok, []string{"PriorityClass", "super-priority", "-n", sub.Namespace}).Check(oc)
			olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Present, "", exutil.Ok, []string{"PodDisruptionBudget", "busybox-pdb", "-n", sub.Namespace}).Check(oc)
		}
	})

	// Group 10 - OCP-34472
	g.It("PolarionID:34472-[OTP][Skipped:Disconnected]olm label dependency", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:34472-[Skipped:Disconnected]olm label dependency"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
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
				Name:        "olm-1933-v8-catalog",
				Namespace:   "",
				DisplayName: "OLM 1933 v8 Operator Catalog",
				Publisher:   "QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-dep:v12",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "mta-operator",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "mta-operator",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "windup-operator.0.0.5",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			dependentOperator = "nginx-ok1-1399.v0.0.5"
		)
		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install operator")
		sub.Create(oc, itName, dr)

		g.By("check if dependent operator is installed")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", dependentOperator, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	// Group 10 - OCP-33176
	g.It("PolarionID:33176-[OTP][Skipped:Disconnected]Enable generated operator component adoption for operators with single ns mode[Slow][Serial]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:33176-[Skipped:Disconnected]Enable generated operator component adoption for operators with single ns mode[Slow][Serial]"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for ask cluster")
		}
		exutil.SkipNoCapabilities(oc, "marketplace")
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		infra, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "none") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName                  = g.CurrentSpecReport().FullText()
			buildPruningBaseDir     = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate        = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate             = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			catsrcImageTemplate     = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			apiserviceImageTemplate = filepath.Join(buildPruningBaseDir, "apiservice.yaml")
			apiserviceVersion       = "v33176"
			apiserviceName          = apiserviceVersion + ".foos.bar.com"
			og                      = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-33176-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 33176 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-api:v5",
				Template:    catsrcImageTemplate,
			}
			subEtcd = olmv0util.SubscriptionDescription{
				SubName:                "etcd33176",
				Namespace:              "",
				Channel:                "singlenamespace-alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "etcdoperator.v0.9.4",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
			subCockroachdb = olmv0util.SubscriptionDescription{
				SubName:                "cockroachdb33176",
				Namespace:              "",
				Channel:                "stable-5.x",
				IpApproval:             "Automatic",
				OperatorPackage:        "cockroachdb",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "cockroachdb.v5.0.4",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		subEtcd.Namespace = oc.Namespace()
		subEtcd.CatalogSourceNamespace = catsrc.Namespace
		subCockroachdb.Namespace = oc.Namespace()
		subCockroachdb.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install Etcd")
		subEtcd.Create(oc, itName, dr)
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subEtcd.OperatorPackage+"."+subEtcd.Namespace)
		}()

		g.By("Check all resources via operators")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "ServiceAccount", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Role", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "RoleBinding", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "CustomResourceDefinition", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Subscription", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "InstallPlan", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "ClusterServiceVersion", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Deployment", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, subEtcd.Namespace, exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].namespace}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "InstallSucceeded", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].conditions[*].reason}"}).Check(oc)

		g.By("delete operator and Operator still exists because of crd")
		subEtcd.Delete(itName, dr)
		_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "csv", subEtcd.InstalledCSV, "-n", subEtcd.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "CustomResourceDefinition", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)

		g.By("reinstall etcd and check Operator")
		subEtcd.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "InstallSucceeded", exutil.Ok, []string{"operator.operators.coreos.com", subEtcd.OperatorPackage + "." + subEtcd.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].conditions[*].reason}"}).Check(oc)

		g.By("delete etcd and the Operator again and Operator should recreated because of crd")
		_, err = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "sub", subEtcd.SubName, "-n", subEtcd.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "csv", subEtcd.InstalledCSV, "-n", subEtcd.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subEtcd.OperatorPackage+"."+subEtcd.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		// here there is issue and take WA
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "crd", "etcdbackups.etcd.database.coreos.com", "operators.coreos.com/"+subEtcd.OperatorPackage+"."+subEtcd.Namespace+"-")
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "crd", "etcdbackups.etcd.database.coreos.com", "operators.coreos.com/"+subEtcd.OperatorPackage+"."+subEtcd.Namespace+"=")
		o.Expect(err).NotTo(o.HaveOccurred())
		//done for WA
		var componentKind string
		err = wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 240*time.Second, false, func(ctx context.Context) (bool, error) {
			componentKind = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subEtcd.OperatorPackage+"."+subEtcd.Namespace, "-o=jsonpath={.status.components.refs[*].kind}")
			if strings.Contains(componentKind, "CustomResourceDefinition") {
				return true, nil
			}
			e2e.Logf("the got kind is %v", componentKind)
			return false, nil
		})
		if err != nil && strings.Compare(componentKind, "") != 0 {
			e2e.Failf("the operator has wrong component")
			// after the official is supported, will change it again.
		}

		g.By("install Cockroachdb")
		subCockroachdb.Create(oc, itName, dr)
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subCockroachdb.OperatorPackage+"."+subCockroachdb.Namespace)
		}()

		g.By("Check all resources of Cockroachdb via operators")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Role", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "RoleBinding", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "CustomResourceDefinition", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Subscription", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "InstallPlan", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "ClusterServiceVersion", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Deployment", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, subCockroachdb.Namespace, exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].namespace}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "InstallSucceeded", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].conditions[*].reason}"}).Check(oc)

		g.By("create ns test-33176 and label it")
		_, err = exutil.OcAction(oc, "create", exutil.AsAdmin, exutil.WithoutNamespace, "ns", "test-33176")
		o.Expect(err).NotTo(o.HaveOccurred())
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "ns", "test-33176", "--force", "--grace-period=0", "--wait=false")
		}()
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "ns", "test-33176", "operators.coreos.com/"+subCockroachdb.OperatorPackage+"."+subCockroachdb.Namespace+"=")
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "Namespace", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)

		g.By("create apiservice and label it")
		err = olmv0util.ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", apiserviceImageTemplate, "-p", "NAME="+apiserviceName, "VERSION="+apiserviceVersion)
		o.Expect(err).NotTo(o.HaveOccurred())
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "apiservice", apiserviceName)
		}()
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "apiservice", apiserviceName,
			"operators.coreos.com/"+subCockroachdb.OperatorPackage+"."+subCockroachdb.Namespace+"=")
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "apiservice", apiserviceName,
			"olm.managed="+`true`)
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "apiservice", apiserviceName,
			"olm.owner"+"="+subCockroachdb.InstalledCSV)
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "apiservice", apiserviceName,
			"olm.owner.kind"+"="+"ClusterServiceVersion")
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "apiservice", apiserviceName,
			"olm.owner.namespace"+"="+subCockroachdb.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "APIService", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)
	})

	// Group 11 - OCP-39897
	g.It("PolarionID:39897-[OTP][Skipped:Disconnected]operator objects should not be recreated after all other associated resources have been deleted[Serial]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:39897-[Skipped:Disconnected]operator objects should not be recreated after all other associated resources have been deleted[Serial]"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for ask cluster")
		}
		exutil.SkipNoCapabilities(oc, "marketplace")
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		infra, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
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
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-39897-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 39897 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv39897",
				Template:    catsrcImageTemplate,
			}
			subMta = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v39897",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v39897",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "nginx-ok-v39897.v0.0.1",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
			crd = olmv0util.CrdDescription{
				Name: "okv39897s.cache.example.com",
			}
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		subMta.Namespace = oc.Namespace()
		subMta.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install Teiid")
		subMta.Create(oc, itName, dr)
		defer func() {
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subMta.OperatorPackage+"."+subMta.Namespace)
		}()

		g.By("Check the resources via operators")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "CustomResourceDefinition", exutil.Ok, []string{"operator.operators.coreos.com", subMta.OperatorPackage + "." + subMta.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)

		g.By("delete operator and Operator still exists because of crd")
		subMta.Delete(itName, dr)
		_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "csv", subMta.InstalledCSV, "-n", subMta.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "CustomResourceDefinition", exutil.Ok, []string{"operator.operators.coreos.com", subMta.OperatorPackage + "." + subMta.Namespace, "-o=jsonpath={.status.components.refs[*].kind}"}).Check(oc)

		g.By("delete crd")
		crd.Delete(oc)

		g.By("delete Operator resource to check if it is recreated")
		_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subMta.OperatorPackage+"."+subMta.Namespace)
		olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"operator.operators.coreos.com", subMta.OperatorPackage + "." + subMta.Namespace}).Check(oc)
	})

	// Group 12 - OCP-50135
	g.It("PolarionID:50135-[OTP][Skipped:Disconnected]automatic upgrade for failed operator installation og created correctly", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:50135-[Skipped:Disconnected]automatic upgrade for failed operator installation og created correctly"), func() {
		var (
			itName                    = g.CurrentSpecReport().FullText()
			buildPruningBaseDir       = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			ogAllTemplate             = filepath.Join(buildPruningBaseDir, "og-allns.yaml")
			ogUpgradeStrategyTemplate = filepath.Join(buildPruningBaseDir, "operatorgroup-upgradestrategy.yaml")

			og = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			ogAll = olmv0util.OperatorGroupDescription{
				Name:      "og-all",
				Namespace: "",
				Template:  ogAllTemplate,
			}
			ogDefault = olmv0util.OperatorGroupDescription{
				Name:            "og-default",
				Namespace:       "",
				UpgradeStrategy: "Default",
				Template:        ogUpgradeStrategyTemplate,
			}
			ogFailForward = olmv0util.OperatorGroupDescription{
				Name:            "og-failforwad",
				Namespace:       "",
				UpgradeStrategy: "TechPreviewUnsafeFailForward",
				Template:        ogUpgradeStrategyTemplate,
			}
			ogFoo = olmv0util.OperatorGroupDescription{
				Name:            "og-foo",
				Namespace:       "",
				UpgradeStrategy: "foo",
				Template:        ogUpgradeStrategyTemplate,
			}
		)

		oc.SetupProject()
		ns := oc.Namespace()
		og.Namespace = ns
		ogAll.Namespace = ns
		ogDefault.Namespace = ns
		ogFailForward.Namespace = ns
		ogFoo.Namespace = ns

		g.By("Create og")
		og.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Default", exutil.Ok, []string{"og", og.Name, "-n", og.Namespace, "-o=jsonpath={.spec.upgradeStrategy}"}).Check(oc)
		og.Delete(itName, dr)

		g.By("Create og all")
		ogAll.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Default", exutil.Ok, []string{"og", ogAll.Name, "-n", ogAll.Namespace, "-o=jsonpath={.spec.upgradeStrategy}"}).Check(oc)
		ogAll.Delete(itName, dr)

		g.By("Create og Default")
		ogDefault.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Default", exutil.Ok, []string{"og", ogDefault.Name, "-n", ogDefault.Namespace, "-o=jsonpath={.spec.upgradeStrategy}"}).Check(oc)
		ogDefault.Delete(itName, dr)

		g.By("Create og failforward")
		ogFailForward.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "TechPreviewUnsafeFailForward", exutil.Ok, []string{"og", ogFailForward.Name, "-n", ogFailForward.Namespace, "-o=jsonpath={.spec.upgradeStrategy}"}).Check(oc)
		ogFailForward.Delete(itName, dr)

		g.By("Create og with invalid upgradeStrategy")
		err := olmv0util.ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", ogFoo.Template, "-p", "NAME="+ogFoo.Name, "NAMESPACE="+ogFoo.Namespace, "UPGRADESTRATEGY="+ogFoo.UpgradeStrategy)
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(err.Error()).To(o.ContainSubstring("exit status 1"))
	})

	// Group 12 - OCP-50136
	g.It("PolarionID:50136-[OTP][Skipped:Disconnected]automatic upgrade for failed operator installation csv fails[Slow][Timeout:30m]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:50136-[Skipped:Disconnected]automatic upgrade for failed operator installation csv fails[Slow][Timeout:30m]"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
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
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		subOadp.Namespace = oc.Namespace()
		subOadp.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install OADP")
		subOadp.Create(oc, itName, dr)

		g.By("Check the oadp-operator.v0.5.3 is installed successfully")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subOadp.InstalledCSV, "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("patch to index image with wrong bundle csv fails")
		err := oc.AsAdmin().WithoutNamespace().Run("patch").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "--type=merge", "-p", "{\"spec\":{\"image\":\"quay.io/olmqe/olm-index:OLM-2378-Oadp-csvfail-multi\"}}").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "oadp-operator.v0.5.4", exutil.Ok, []string{"sub", subOadp.SubName, "-n", subOadp.Namespace, "-o=jsonpath={.status.currentCSV}"}).Check(oc)

		g.By("check the csv fails")
		var status string
		// it fails after 10m which we can not control it. so, have to check it in 11m
		err = wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 15*time.Minute, false, func(ctx context.Context) (bool, error) {
			status = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "oadp-operator.v0.5.4", "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Failed") == 0 {
				e2e.Logf("csv oadp-operator.v0.5.4 fails expected")
				return true, nil
			}
			return false, nil
		})
		if strings.Contains(status, "nstalling") {
			return
		}
		exutil.AssertWaitPollNoErr(err, "csv oadp-operator.v0.5.4 is not failing as expected")

		g.By("change upgrade strategy to TechPreviewUnsafeFailForward")
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("og", og.Name, "-n", og.Namespace, "--type=merge", "-p", "{\"spec\":{\"upgradeStrategy\":\"TechPreviewUnsafeFailForward\"}}").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("check if oadp-operator.v0.5.6 is created")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			csv := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", subOadp.SubName, "-n", subOadp.Namespace, "-o=jsonpath={.status.currentCSV}")
			if strings.Compare(csv, "oadp-operator.v0.5.6") == 0 {
				e2e.Logf("csv %v is created", csv)
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "csv oadp-operator.v0.5.6 is not created")

		g.By("check if upgrade is done")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "oadp-operator.v0.5.6", "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") == 0 {
				e2e.Logf("csv oadp-operator.v0.5.6 is successful")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "csv oadp-operator.v0.5.6 is not successful")

	})

	g.It("PolarionID:50138-[OTP][Skipped:Disconnected]automatic upgrade for failed operator installation ip fails", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:50138-[Skipped:Disconnected]automatic upgrade for failed operator installation ip fails"), func() {
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
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
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		subOadp.Namespace = oc.Namespace()
		subOadp.CatalogSourceNamespace = catsrc.Namespace

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("install OADP")
		subOadp.Create(oc, itName, dr)

		g.By("Check the oadp-operator.v0.5.3 is installed successfully")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subOadp.InstalledCSV, "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("patch to index image with wrong bundle ip fails")
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "--type=merge", "-p", "{\"spec\":{\"image\":\"quay.io/olmqe/olm-index:OLM-2378-Oadp-ipfailTwo-multi\"}}").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "oadp-operator.v0.5.5", exutil.Ok, []string{"sub", subOadp.SubName, "-n", subOadp.Namespace, "-o=jsonpath={.status.currentCSV}"}).Check(oc)

		g.By("check the ip fails")
		ips := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", subOadp.SubName, "-n", subOadp.Namespace, "-o=jsonpath={.status.installplan.name}")
		o.Expect(ips).NotTo(o.BeEmpty())
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", ips, "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Failed") == 0 {
				e2e.Logf("ip %v fails expected", ips)
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("ip %v not failing as expected", ips))

		g.By("change upgrade strategy to TechPreviewUnsafeFailForward")
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("og", og.Name, "-n", og.Namespace, "--type=merge", "-p", "{\"spec\":{\"upgradeStrategy\":\"TechPreviewUnsafeFailForward\"}}").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("patch to index image again with fixed bundle")
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "--type=merge", "-p", "{\"spec\":{\"image\":\"quay.io/olmqe/olm-index:OLM-2378-Oadp-ipfailskip-multi\"}}").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			csv := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", subOadp.SubName, "-n", subOadp.Namespace, "-o=jsonpath={.status.currentCSV}")
			if strings.Compare(csv, "oadp-operator.v0.5.6") == 0 {
				e2e.Logf("csv %v is created", csv)
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "csv oadp-operator.v0.5.6 is not created")

		g.By("check if upgrade is done")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "oadp-operator.v0.5.6", "-n", subOadp.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") == 0 {
				e2e.Logf("csv oadp-operator.v0.5.6 is successful")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "csv oadp-operator.v0.5.6 is not successful")

	})

	g.It("PolarionID:40958-[OTP][Skipped:Disconnected]Indicate invalid OperatorGroup on InstallPlan status", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:40958-[Skipped:Disconnected]Indicate invalid OperatorGroup on InstallPlan status"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		if isAks, _ := exutil.IsAKSCluster(context.TODO(), oc); isAks {
			g.Skip("skip for ask cluster")
		}
		exutil.SkipNoCapabilities(oc, "marketplace")
		node, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		errGet = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errGet).NotTo(o.HaveOccurred())
		efips, errGet := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if errGet != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it without impacting function")
		}
		infra, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(errGet).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			g.Skip("it is not supported")
		}
		platform := exutil.CheckPlatform(oc)
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "none") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			ogSAtemplate        = filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-legacy.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			saName              = "scopedv40958"
			og1                 = olmv0util.OperatorGroupDescription{
				Name:      "og1-40958",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			og2 = olmv0util.OperatorGroupDescription{
				Name:      "og2-40958",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			ogSa = olmv0util.OperatorGroupDescription{
				Name:               "ogsa-40958",
				Namespace:          "",
				ServiceAccountName: saName,
				Template:           ogSAtemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-40958-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 40958 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olm-dep:v40958",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "teiid",
				Namespace:              "",
				Channel:                "beta",
				IpApproval:             "Automatic",
				OperatorPackage:        "teiid",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "teiid.v0.4.0",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		og1.Namespace = oc.Namespace()
		og2.Namespace = oc.Namespace()
		ogSa.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		defer func() { _ = exutil.RecoverNamespaceRestricted(oc, oc.Namespace()) }()
		errSet := exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errSet).NotTo(o.HaveOccurred())

		g.By("create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("install operator without og")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("no Installplan is generated, without og")
		// by https://issues.redhat.com/browse/OCPBUGS-9259
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 10*time.Second, false, func(ctx context.Context) (bool, error) {
			var err error
			installPlan, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installPlanRef.name}").Output()
			if strings.Compare(installPlan, "") == 0 || err != nil {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollWithErr(waitErr, fmt.Sprintf("sub %s has installplan", sub.SubName))

		g.By("delete operator")
		sub.Delete(itName, dr)

		g.By("Create og1")
		og1.Create(oc, itName, dr)

		g.By("Create og2")
		og2.Create(oc, itName, dr)

		g.By("install operator with multiple og")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("no Installplan is generated, multiple og")
		waitErr = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 10*time.Second, false, func(ctx context.Context) (bool, error) {
			var err error
			installPlan, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installPlanRef.name}").Output()
			if strings.Compare(installPlan, "") == 0 || err != nil {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollWithErr(waitErr, fmt.Sprintf("sub %s has installplan", sub.SubName))

		g.By("delete resource for next step")
		sub.Delete(itName, dr)
		og1.Delete(itName, dr)
		og2.Delete(itName, dr)

		g.By("create sa")
		_, err := oc.WithoutNamespace().AsAdmin().Run("create").Args("sa", saName, "-n", sub.Namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Create ogSa")
		ogSa.CreateWithCheck(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, saName, exutil.Ok, []string{"og", ogSa.Name, "-n", ogSa.Namespace, "-o=jsonpath={.status.serviceAccountRef.name}"}).Check(oc)

		g.By("delete the service account")
		_, err = oc.WithoutNamespace().AsAdmin().Run("delete").Args("sa", saName, "-n", sub.Namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("install operator without sa for og")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("no Installplan is generated, without sa for og")
		installPlan, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installPlanRef.name}").Output()
		if strings.Compare(installPlan, "") != 0 && err == nil {
			subContent, _ := oc.WithoutNamespace().AsAdmin().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-oyaml").Output()
			e2e.Logf("subContent: %v", subContent)
			e2e.Failf("should no ip")
		}
	})

	g.It("PolarionID:60114-[OTP][Skipped:Disconnected]olm serves an api to discover all versions of an operator[Slow]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:60114-[Skipped:Disconnected]olm serves an api to discover all versions of an operator[Slow]"), func() {
		architecture.SkipArchitectures(oc, architecture.PPC64LE, architecture.S390X, architecture.MULTI)
		_ = oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata}").Execute()
		nodes, errGet := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-l", "node.kubernetes.io/instance-type=Standard_EC4es_v6").Output()
		if errGet != nil || len(nodes) == 0 {
			e2e.Logf("nodes: %v", nodes)
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
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			catsrc              = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-run1399-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc RUN1399 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "",
				Template:    catsrcImageTemplate,
			}
		)

		catsrc.Namespace = oc.Namespace()

		ok1AlphaAssertion := func(entries string) {
			o.Expect(entries).To(o.ContainSubstring("nginx-ok1-1399.v0.0.4"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok1-1399.v0.0.2"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok1-1399.v0.0.1"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok1-1399.v0.0.5"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok1-1399.v0.0.3"))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.4\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.2\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.1\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.5\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.3\""))
		}
		ok1BetaAssertion := func(entries string) {
			o.Expect(entries).To(o.ContainSubstring("nginx-ok1-1399.v0.0.5"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok1-1399.v0.0.3"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok1-1399.v0.0.1"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok1-1399.v0.0.4"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok1-1399.v0.0.2"))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.5\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.3\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.1\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.4\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.2\""))
		}
		ok2AlphaNoDepAssertion := func(entries string) {
			o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.4"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.2"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.1"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.5"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.3"))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.4\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.2\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.1\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.5\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.3\""))
		}
		ok2BetaAssertion := func(entries string) {
			o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.5"))
			o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.3"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.4"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.2"))
			o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.1"))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.5\""))
			o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.3\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.4\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.2\""))
			o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.1\""))
		}

		g.By("fbc based image without deprecated bundle")
		catsrc.Address = "quay.io/olmqe/nginx-ok-index:v1399-fbc-multi"
		catsrc.CreateWithCheck(oc, itName, dr)
		entries := olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		ok1AlphaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok1BetaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		ok2AlphaNoDepAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok2BetaAssertion(entries)

		catsrc.Delete(itName, dr)

		g.By("ffbc based image with deprecated bundle made by properties.yaml")
		catsrc.Address = "quay.io/olmqe/nginx-ok-index:v1399-fbc-deprecate-nomigrate-multi"
		catsrc.CreateWithCheck(oc, itName, dr)
		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		ok1AlphaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok1BetaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.4"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.2"))
		o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.1"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.5"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.3"))
		o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.4\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.2\""))
		o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.1\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.5\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.3\""))

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok2BetaAssertion(entries)

		catsrc.Delete(itName, dr)

		g.By("sqlite based image without deprecated bundle")
		catsrc.Address = "quay.io/olmqe/nginx-ok-index:v1399-sql"
		defer func() {
			errRecover := exutil.RecoverNamespaceRestricted(oc, oc.Namespace())
			if errRecover != nil {
				e2e.Logf("RecoverNamespaceRestricted error: %v", errRecover)
			}
		}()
		errPriv := exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(errPriv).NotTo(o.HaveOccurred())
		catsrc.CreateWithCheck(oc, itName, dr)
		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		ok1AlphaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok1BetaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		ok2AlphaNoDepAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok2BetaAssertion(entries)

		catsrc.Delete(itName, dr)

		g.By("sqlite based image with deprecated bundle made by deprecatetruncate")
		catsrc.Address = "quay.io/olmqe/nginx-ok-index:v1399-sql-deprecate"
		catsrc.CreateWithCheck(oc, itName, dr)
		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		ok1AlphaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok1-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok1BetaAssertion(entries)

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"alpha\")].entries}")
		o.Expect(entries).To(o.ContainSubstring("nginx-ok2-1399.v0.0.4"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.2"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.1"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.5"))
		o.Expect(entries).NotTo(o.ContainSubstring("nginx-ok2-1399.v0.0.3"))
		o.Expect(entries).To(o.ContainSubstring("\"version\":\"0.0.4\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.2\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.1\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.5\""))
		o.Expect(entries).NotTo(o.ContainSubstring("\"version\":\"0.0.3\""))

		entries = olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx-ok2-1399", "-n", catsrc.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"beta\")].entries}")
		ok2BetaAssertion(entries)

	})

	g.It("PolarionID:62974-[OTP]olm sets invalid scc label on its namespaces", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:62974-olm sets invalid scc label on its namespaces"), func() {
		labelKey := "openshift\\.io\\/scc"

		for _, ns := range []string{"openshift-operators", "openshift-operator-lifecycle-manager"} {
			g.By("check label openshift.io/scc is empty on " + ns)
			sccLabel, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("namespace", ns, "-o=jsonpath={.metadata.labels}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(sccLabel).NotTo(o.BeEmpty())
			e2e.Logf("the lables: %v", sccLabel)
			gResult := gjson.Get(sccLabel, labelKey)
			if gResult.Exists() && gResult.String() != "" {
				o.Expect("the value of label openshift.io/scc").To(o.BeEmpty(), fmt.Sprintf("there is label openshift.io/scc on %v and is not empty on", ns))
			}
		}
	})

	g.It("PolarionID:62973-[OTP][Skipped:Disconnected]dedicated way collect profiles cronjob pod missing target.workload.openshift.io management annotation[Disruptive][Slow]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:62973-[Skipped:Disconnected]dedicated way collect profiles cronjob pod missing target.workload.openshift.io management annotation[Serial][Disruptive][Slow]"), func() {
		if !exutil.IsSNOCluster(oc) {
			g.Skip("it is not sno cluster, so skip it")
		}
		g.By("check if the current mcp is ready, or else skip")
		olmv0util.AssertOrCheckMCP(oc, "master", 10, 1, true)

		g.By("check if it is aleady in workload partition")
		wordLoadPartition, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.items[*].metadata.annotations}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if strings.Contains(wordLoadPartition, "resources.workload.openshift.io/collect-profiles") {
			e2e.Logf("it already works")
			return
		}

		var (
			buildPruningBaseDir  = exutil.FixturePath("testdata", "olm")
			mcWordloadPartiation = filepath.Join(buildPruningBaseDir, "mc-workload-partition.yaml")
		)

		g.By("apply MchineConfig to set workload partition")
		defer func() {
			g.By("wait mcp recovered")
			olmv0util.AssertOrCheckMCP(oc, "master", 240, 30, false)
		}()
		defer func() {
			_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", mcWordloadPartiation).Execute()
		}()
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", mcWordloadPartiation).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("check mcp updated successfully")
		olmv0util.AssertOrCheckMCP(oc, "master", 180, 30, false)

		g.By("check resources.workload.openshift.io/collect-profiles")
		o.Eventually(func() string {
			annotation, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.items[*].metadata.annotations}").Output()
			return annotation
		}, 20*time.Minute, 1*time.Minute).Should(o.ContainSubstring("resources.workload.openshift.io/collect-profiles"))
	})

	g.It("PolarionID:62973-[OTP]general way collect profiles cronjob pod missing target.workload.openshift.io management annotation", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:62973-general way collect profiles cronjob pod missing target.workload.openshift.io management annotation"), func() {
		g.By("https://issues.redhat.com/browse/OCPBUGS-1088 automated")

		g.By("check target.workload.openshift.io/management")
		o.Eventually(func() string {
			annotation, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("CronJob", "collect-profiles", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.jobTemplate.spec.template.metadata.annotations}").Output()
			return annotation
		}, 20*time.Second, 2*time.Second).Should(o.ContainSubstring("target.workload.openshift.io/management"))
	})

})
