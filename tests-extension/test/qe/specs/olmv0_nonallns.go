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
	g.It("PolarionID:23170-[OTP]API labels should be hash", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:23170-[Skipped:Disconnected]API labels should be hash"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		olmv0util.ValidateAccessEnvironment(oc)
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
	g.It("PolarionID:20979-[OTP]only one IP is generated", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:20979-[Skipped:Disconnected]only one IP is generated"), func() {
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
		olmv0util.ValidateAccessEnvironment(oc)
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
		var output string
		var err error
		errCsv := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err = oc.WithoutNamespace().Run("get").Args("csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.reason}").Output()
			if err != nil {
				return false, err
			}
			if strings.Contains(output, "RequirementsNotMet") {
				return true, nil
			}
			return false, nil
		})
		if strings.Contains(output, "InstallWaiting") {
			g.Skip("skip because of slow installation")
		}
		exutil.AssertWaitPollNoErr(errCsv, fmt.Sprintf("csv status %v is not expected", output))

		g.By("Recovery sa of csv")
		sa.Reapply(oc)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded+2+Installing", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	// OCP-29723 - As cluster admin find abnormal status condition via components of operator resource
	g.It("PolarionID:29723-[OTP]As cluster admin find abnormal status condition via components of operator resource", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:29723-[Skipped:Disconnected]As cluster admin find abnormal status condition via components of operator resource"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		olmv0util.ValidateAccessEnvironment(oc)
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
	g.It("PolarionID:30762-[OTP]installs bundles with v1 CRDs", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:30762-[Skipped:Disconnected]installs bundles with v1 CRDs"), func() {
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
		olmv0util.ValidateAccessEnvironment(oc)
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
	g.It("PolarionID:27683-[OTP]InstallPlans can install from extracted bundles", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:27683-[Skipped:Disconnected]InstallPlans can install from extracted bundles"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		olmv0util.ValidateAccessEnvironment(oc)
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

	g.It("PolarionID:24513-[OTP]Operator config support env only", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:24513-[Skipped:Disconnected]Operator config support env only"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		platform := exutil.CheckPlatform(oc)
		if strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || strings.Contains(platform, "none") || exutil.Is3MasterNoDedicatedWorkerNode(oc) {
			g.Skip("it is not supported")
		}
		olmv0util.ValidateAccessEnvironment(oc)
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
	g.It("PolarionID:25760-[OTP]Operator upgrades does not fail after change the channel", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:25760-[Skipped:Disconnected]Operator upgrades does not fail after change the channel"), func() {
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
		olmv0util.ValidateAccessEnvironment(oc)
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
		defer sub.DeleteCSV(itName, dr)
		defer sub.Delete(itName, dr)

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
	g.It("PolarionID:39897-[OTP]operator objects should not be recreated after all other associated resources have been deleted[Serial]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within a namespace PolarionID:39897-[Skipped:Disconnected]operator objects should not be recreated after all other associated resources have been deleted[Serial]"), func() {
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
		olmv0util.ValidateAccessEnvironment(oc)
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

	g.It("PolarionID:23673-[OTP]Installplan can be created while Install and uninstall operators via Marketplace for 5 times[Slow]", func() {
		olmv0util.SkipIfPackagemanifestNotExist(oc, "learn")
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subFile             = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "23673",
				Namespace: "",
				Template:  ogTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-23673",
				Namespace:              "",
				CatalogSourceName:      "qe-app-registry",
				CatalogSourceNamespace: "openshift-marketplace",
				IpApproval:             "Automatic",
				Channel:                "beta",
				OperatorPackage:        "learn",
				SingleNamespace:        true,
				Template:               subFile,
			}
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()

		g.By("Create operatorgroup")
		og.Create(oc, itName, dr)

		g.By("Subscribe to operator prometheus")
		sub.Create(oc, itName, dr)
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "AtLatestKnown", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.state}"}).Check(oc)

		g.By("Grab the installedCSV and use as startingCSV")
		finalCSV := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", "-n", oc.Namespace(), sub.SubName, "-o=jsonpath={.status.installedCSV}")
		o.Expect(finalCSV).NotTo(o.BeEmpty())
		sub.StartingCSV = finalCSV

		g.By("Unsubscribe to operator learn")
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		msgSub := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", "-n", oc.Namespace())
		msgCsv := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", oc.Namespace())
		if !strings.Contains(msgSub, "No resources found") && (!strings.Contains(msgCsv, "No resources found") || strings.Contains(msgCsv, finalCSV)) {
			e2e.Failf("Cycle #1 subscribe/unsubscribe failed:\n%v \n%v \n", msgSub, msgCsv)
		}

		g.By("Subscribe/unsubscribe to operator learn 4 more times")
		for i := 2; i < 6; i++ {
			e2e.Logf("Cycle #%v starts", i)

			g.By("Subscribe")
			sub.Create(oc, itName, dr)
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", finalCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

			g.By("Unsubscribe")
			sub.Delete(itName, dr)
			_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", oc.Namespace(), sub.InstalledCSV)

			err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
				msgSub = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", "-n", oc.Namespace())
				e2e.Logf("STEP %v sub msg: %v", i, msgSub)
				msgCsv = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", oc.Namespace())
				e2e.Logf("STEP %v csv msg: %v", i, msgCsv)
				if strings.Contains(msgSub, "No resources found") && (strings.Contains(msgCsv, "No resources found") || !strings.Contains(msgCsv, finalCSV)) {
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("STEP error sub or csv not deleted on cycle #%v:\nsub %v\ncsv %v\n", i, msgSub, msgCsv))
		}
	})

	g.It("PolarionID:24566-[OTP][Skipped:Disconnected]OLM automatically configures operators with global proxy config", func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			subTemplateProxy    = filepath.Join(buildPruningBaseDir, "olm-proxy-subscription.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-nginx-operator",
				Namespace:   "",
				DisplayName: "Test 24566 Operators",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-24566",
				Namespace:              "",
				CatalogSourceName:      "catsrc-nginx-operator",
				CatalogSourceNamespace: "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
			subP = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-24566",
				Namespace:              "",
				CatalogSourceName:      "catsrc-nginx-operator",
				CatalogSourceNamespace: "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				SingleNamespace:        true,
				Template:               subTemplateProxy,
			}
			subProxyTest = olmv0util.SubscriptionDescriptionProxy{
				SubscriptionDescription: subP,
				HttpProxy:               "test_http_proxy",
				HttpsProxy:              "test_https_proxy",
				NoProxy:                 "test_no_proxy",
			}
			subProxyFake = olmv0util.SubscriptionDescriptionProxy{
				SubscriptionDescription: subP,
				HttpProxy:               "fake_http_proxy",
				HttpsProxy:              "fake_https_proxy",
				NoProxy:                 "fake_no_proxy",
			}
			subProxyEmpty = olmv0util.SubscriptionDescriptionProxy{
				SubscriptionDescription: subP,
				HttpProxy:               "",
				HttpsProxy:              "",
				NoProxy:                 "",
			}
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = oc.Namespace()
		subP.Namespace = oc.Namespace()
		subP.CatalogSourceNamespace = oc.Namespace()
		subProxyTest.Namespace = oc.Namespace()
		subProxyTest.CatalogSourceNamespace = oc.Namespace()
		subProxyFake.Namespace = oc.Namespace()
		subProxyFake.CatalogSourceNamespace = oc.Namespace()
		subProxyEmpty.Namespace = oc.Namespace()
		subProxyEmpty.CatalogSourceNamespace = oc.Namespace()

		g.By("0) get the cluster proxy configuration")
		httpProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "proxy", "cluster", "-o=jsonpath={.status.httpProxy}")
		httpsProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "proxy", "cluster", "-o=jsonpath={.status.httpsProxy}")
		noProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "proxy", "cluster", "-o=jsonpath={.status.noProxy}")

		g.By(fmt.Sprintf("1) create the catsrc and OperatorGroup in project: %s", oc.Namespace()))
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		og.Create(oc, itName, dr)

		g.By("2) install sub")
		sub.Create(oc, itName, dr)
		g.By("install operator SUCCESS")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator-controller-manager", exutil.Ok, []string{"deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..metadata.name}"}).Check(oc)

		if httpProxy == "" {
			nodeHTTPProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTP_PROXY\")].value}")
			o.Expect(nodeHTTPProxy).To(o.BeEmpty())
			nodeHTTPSProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTPS_PROXY\")].value}")
			o.Expect(nodeHTTPSProxy).To(o.BeEmpty())
			nodeNoProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"NO_PROXY\")].value}")
			o.Expect(nodeNoProxy).To(o.BeEmpty())
			g.By("CHECK proxy configure SUCCESS")
			sub.Delete(itName, dr)
			sub.DeleteCSV(itName, dr)

			g.By("3) create subscription and set variables ( HTTP_PROXY, HTTPS_PROXY and NO_PROXY ) with non-empty values. ")
			subProxyTest.Create(oc, itName, dr)
			err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
				status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", subProxyTest.InstalledCSV, "-n", subProxyTest.Namespace, "-o=jsonpath={.status.phase}")
				if (strings.Compare(status, "Succeeded") == 0) || (strings.Compare(status, "Installing") == 0) {
					e2e.Logf("csv status is Succeeded or Installing")
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("csv %s is not Succeeded or Installing", subProxyTest.InstalledCSV))
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator-controller-manager", exutil.Ok, []string{"deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..metadata.name}"}).Check(oc)
			nodeHTTPProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTP_PROXY\")].value}")
			o.Expect(nodeHTTPProxy).To(o.Equal("test_http_proxy"))
			nodeHTTPSProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTPS_PROXY\")].value}")
			o.Expect(nodeHTTPSProxy).To(o.Equal("test_https_proxy"))
			nodeNoProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"NO_PROXY\")].value}")
			o.Expect(nodeNoProxy).To(o.Equal("test_no_proxy"))
			subProxyTest.Delete(itName, dr)
			subProxyTest.GetCSV().Delete(itName, dr)
		} else {
			o.Expect(httpProxy).NotTo(o.BeEmpty())
			o.Expect(httpsProxy).NotTo(o.BeEmpty())
			o.Expect(noProxy).NotTo(o.BeEmpty())
			nodeHTTPProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTP_PROXY\")].value}")
			o.Expect(nodeHTTPProxy).To(o.Equal(httpProxy))
			nodeHTTPSProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTPS_PROXY\")].value}")
			o.Expect(nodeHTTPSProxy).To(o.Equal(httpsProxy))
			nodeNoProxy := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"NO_PROXY\")].value}")
			o.Expect(nodeNoProxy).To(o.Equal(noProxy))
			g.By("CHECK proxy configure SUCCESS")
			sub.Delete(itName, dr)
			sub.DeleteCSV(itName, dr)

			g.By("3) create subscription and set variables ( HTTP_PROXY, HTTPS_PROXY and NO_PROXY ) with non-empty values. ")
			subProxyTest.Create(oc, itName, dr)
			err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
				status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", subProxyTest.InstalledCSV, "-n", subProxyTest.Namespace, "-o=jsonpath={.status.phase}")
				if (strings.Compare(status, "Succeeded") == 0) || (strings.Compare(status, "Installing") == 0) {
					e2e.Logf("csv status is Succeeded or Installing")
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("csv %s is not Succeeded or Installing", subProxyTest.InstalledCSV))
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator-controller-manager", exutil.Ok, []string{"deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..metadata.name}"}).Check(oc)
			nodeHTTPProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTP_PROXY\")].value}")
			o.Expect(nodeHTTPProxy).To(o.Equal("test_http_proxy"))
			nodeHTTPSProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTPS_PROXY\")].value}")
			o.Expect(nodeHTTPSProxy).To(o.Equal("test_https_proxy"))
			nodeNoProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyTest.InstalledCSV), "-n", subProxyTest.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"NO_PROXY\")].value}")
			o.Expect(nodeNoProxy).To(o.Equal("test_no_proxy"))
			subProxyTest.Delete(itName, dr)
			subProxyTest.GetCSV().Delete(itName, dr)

			g.By("4) Create a new subscription and set variables ( HTTP_PROXY, HTTPS_PROXY and NO_PROXY ) with a fake value.")
			subProxyFake.Create(oc, itName, dr)
			err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
				status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", subProxyFake.InstalledCSV, "-n", subProxyFake.Namespace, "-o=jsonpath={.status.phase}")
				if (strings.Compare(status, "Succeeded") == 0) || (strings.Compare(status, "Installing") == 0) {
					e2e.Logf("csv status is Succeeded or Installing")
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("csv %s is not Succeeded or Installing", subProxyFake.InstalledCSV))
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator-controller-manager", exutil.Ok, []string{"deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyFake.InstalledCSV), "-n", subProxyFake.Namespace, "-o=jsonpath={..metadata.name}"}).Check(oc)
			nodeHTTPProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyFake.InstalledCSV), "-n", subProxyFake.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTP_PROXY\")].value}")
			o.Expect(nodeHTTPProxy).To(o.Equal("fake_http_proxy"))
			nodeHTTPSProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyFake.InstalledCSV), "-n", subProxyFake.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTPS_PROXY\")].value}")
			o.Expect(nodeHTTPSProxy).To(o.Equal("fake_https_proxy"))
			nodeNoProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyFake.InstalledCSV), "-n", subProxyFake.Namespace, "-o=jsonpath={..spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"NO_PROXY\")].value}")
			o.Expect(nodeNoProxy).To(o.Equal("fake_no_proxy"))
			subProxyFake.Delete(itName, dr)
			subProxyFake.GetCSV().Delete(itName, dr)

			g.By("5) Create a new subscription and set variables ( HTTP_PROXY, HTTPS_PROXY and NO_PROXY ) with an empty value.")
			subProxyEmpty.Create(oc, itName, dr)
			err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
				status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", subProxyEmpty.InstalledCSV, "-n", subProxyEmpty.Namespace, "-o=jsonpath={.status.phase}")
				if (strings.Compare(status, "Succeeded") == 0) || (strings.Compare(status, "Installing") == 0) {
					e2e.Logf("csv status is Succeeded or Installing")
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("csv %s is not Succeeded or Installing", subProxyEmpty.InstalledCSV))
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator-controller-manager", exutil.Ok, []string{"deployment", fmt.Sprintf("--selector=olm.owner=%s", subProxyEmpty.InstalledCSV), "-n", subProxyEmpty.Namespace, "-o=jsonpath={..metadata.name}"}).Check(oc)
			nodeHTTPProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=marketplace.operatorSource=%s", subProxyEmpty.InstalledCSV), "-n", subProxyEmpty.Namespace, "-o=jsonpath={.spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTP_PROXY\")].value}")
			o.Expect(nodeHTTPProxy).To(o.BeEmpty())
			nodeHTTPSProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=marketplace.operatorSource=%s", subProxyEmpty.InstalledCSV), "-n", subProxyEmpty.Namespace, "-o=jsonpath={.spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"HTTPS_PROXY\")].value}")
			o.Expect(nodeHTTPSProxy).To(o.BeEmpty())
			nodeNoProxy = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=marketplace.operatorSource=%s", subProxyEmpty.InstalledCSV), "-n", subProxyEmpty.Namespace, "-o=jsonpath={.spec.template.spec.containers[?(.name==\"manager\")].env[?(.name==\"NO_PROXY\")].value}")
			o.Expect(nodeNoProxy).To(o.BeEmpty())
			subProxyEmpty.Delete(itName, dr)
			subProxyEmpty.GetCSV().Delete(itName, dr)
		}
	})

	g.It("PolarionID:24664-[OTP][Skipped:Disconnected]CRD updates if new schemas are backwards compatible", func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "nginx-24664-index",
				Namespace:   "",
				DisplayName: "nginx-24664",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-operator-index-24664:multi-arch",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-24664",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator-24664",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
			crd = olmv0util.CrdDescription{
				Name: "nginx24664s.cache.example.com",
			}
		)

		oc.SetupProject()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()

		g.By("ensure no such crd")
		crd.Delete(oc)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create sub")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "v2", exutil.Nok, []string{"crd", crd.Name, "-A", "-o=jsonpath={.status.storedVersions}"}).Check(oc)

		g.By("update channel of Sub")
		sub.Patch(oc, "{\"spec\": {\"channel\": \"beta\"}}")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "nginx-operator-24664.v0.0.2", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") == 0 {
				e2e.Logf("csv nginx-operator-24664.v0.0.2 is Succeeded")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "csv nginx-operator-24664.v0.0.2 is not Succeeded")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "v2", exutil.Ok, []string{"crd", crd.Name, "-A", "-o=jsonpath={.status.storedVersions}"}).Check(oc)
	})

	g.It("PolarionID:29809-[OTP][Skipped:Disconnected]updatation based on replaces can be completed automatically", func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-29809",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-29809",
				Namespace:   "",
				DisplayName: "Test Catsrc 29809 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-29809",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Template:               subTemplate,
				SingleNamespace:        true,
				StartingCSV:            "nginx-operator.v0.0.1",
			}
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create og")
		og.Create(oc, itName, dr)

		g.By("create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("install operator")
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)

		g.By("check the operator upgrade to nginx-operator.v0.0.1")
		err := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 480*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", sub.Namespace, "csv", "nginx-operator.v1.0.1", "-o=jsonpath={.spec.replaces}").Output()
			e2e.Logf("output: %s", output)
			if err != nil {
				e2e.Logf("The csv is not created, error:%v", err)
				return false, nil
			}
			if strings.Contains(output, "nginx-operator.v0.0.1") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "nginx-operator.v1.0.1 does not replace nginx-operator.v0.0.1")
	})

	g.It("PolarionID:30206-PolarionID:30242-[OTP][Skipped:Disconnected]can include secrets and configmaps in the bundle", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-30206",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-30206",
				Namespace:   "",
				DisplayName: "Test Catsrc 30206 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/cockroachdb-index:5.0.4-30206-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "cockroachdb-operator-30206",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "cockroachdb",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Template:               subTemplate,
				SingleNamespace:        true,
				StartingCSV:            "cockroachdb.v5.0.4",
			}
		)

		oc.SetupProject()
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create og")
		og.Create(oc, itName, dr)

		g.By("create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("install operator")
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)

		g.By("check secrets")
		errWait := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 240*time.Second, false, func(ctx context.Context) (bool, error) {
			err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", sub.Namespace, "secrets", "mysecret").Execute()
			if err != nil {
				e2e.Logf("Failed to create secrets, error:%v", err)
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(errWait, "mysecret is not created")

		g.By("check configmaps")
		errWait = wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 240*time.Second, false, func(ctx context.Context) (bool, error) {
			err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", sub.Namespace, "configmaps", "my-config-map").Execute()
			if err != nil {
				e2e.Logf("Failed to create secrets, error:%v", err)
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(errWait, "my-config-map is not found")

		g.By("start to test OCP-30242")
		g.By("delete csv")
		sub.DeleteCSV(itName, dr)

		g.By("check secrets has been deleted")
		errWait = wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
			err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", sub.Namespace, "secrets", "mysecret").Execute()
			if err != nil {
				e2e.Logf("The secrets has been deleted")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(errWait, "mysecret is not found")

		g.By("check configmaps has been deleted")
		errWait = wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
			err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", sub.Namespace, "configmaps", "my-config-map").Execute()
			if err != nil {
				e2e.Logf("The configmaps has been deleted")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(errWait, "my-config-map still exists")
	})

	g.It("PolarionID:21824-[OTP][Skipped:Disconnected]verify CRD should be ready before installing the operator", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			cmWrong             = filepath.Join(buildPruningBaseDir, "cm-21824-wrong.yaml")
			cmCorrect           = filepath.Join(buildPruningBaseDir, "cm-21824-correct.yaml")
			catsrcCmTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogTemplate,
			}
			cm = olmv0util.ConfigMapDescription{
				Name:      "cm-21824",
				Namespace: "",
				Template:  cmWrong,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-21824",
				Namespace:   "",
				DisplayName: "21824 Operators",
				Publisher:   "olmqe",
				SourceType:  "configmap",
				Address:     "cm-21824",
				Template:    catsrcCmTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "kubeturbo21824-operator-21824",
				Namespace:              "",
				IpApproval:             "Automatic",
				OperatorPackage:        "kubeturbo21824",
				CatalogSourceName:      "catsrc-21824",
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)

		oc.SetupProject()
		cm.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create cm with wrong crd")
		cm.Create(oc, itName, dr)

		g.By("Create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create sub and cannot succeed")
		sub.CreateWithoutCheck(oc, itName, dr)
		err := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
			subStatus := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}")
			e2e.Logf("subscription status: %s", subStatus)
			if strings.Contains(subStatus, "invalid") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("status.conditions of sub %s doesn't have expect meesage", sub.SubName))

		sub.FindInstalledCSV(oc, itName, dr)
		err = wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
			csvPhase := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.requirementStatus}")
			e2e.Logf("csv phase: %s", csvPhase)
			if strings.Contains(csvPhase, "NotPresent") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("status.requirementStatus of csv %s is not correct", sub.InstalledCSV))
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		cm.Delete(itName, dr)
		catsrc.Delete(itName, dr)

		g.By("update cm to correct crd")
		cm.Name = "cm-21824-correct"
		cm.Template = cmCorrect
		cm.Create(oc, itName, dr)
		catsrc.Name = "catsrc-21824-correct"
		catsrc.Address = cm.Name
		catsrc.Create(oc, itName, dr)
		sub.CatalogSourceName = catsrc.Name
		sub.Create(oc, itName, dr)

		g.By("sub succeed and csv succeed")
		sub.FindInstalledCSV(oc, itName, dr)
		err = wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			csvStatus, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if csvStatus == "Succeeded" {
				e2e.Logf("CSV status is Succeeded")
				return true, nil
			}
			e2e.Logf("CSV status is %s, not Succeeded, go next round", csvStatus)
			return false, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("status.phase of csv %s is not Succeeded", sub.InstalledCSV))
	})

	g.It("PolarionID:30312-[OTP][Skipped:Disconnected]can allow admission webhook definitions in CSV", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-30312",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-30312",
				Namespace:   "",
				DisplayName: "Test Catsrc 30312 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-operator-index-30312:v2-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-30312",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator-30312",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Template:               subTemplate,
			}
		)

		oc.SetupProject()
		ns := oc.Namespace()
		og.Namespace = ns
		sub.Namespace = ns
		catsrc.Namespace = ns
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = ns

		g.By("create og")
		og.Create(oc, itName, dr)

		g.By("create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("install operator")
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		err := wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			err := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", "-l", "olm.owner.namespace="+ns).Execute()
			if err != nil {
				e2e.Logf("The validatingwebhookconfiguration is not created:%v", err)
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("validatingwebhookconfiguration which owner ns %s is not created", ns))

		g.By("update csv")
		_, err1 := oc.AsAdmin().WithoutNamespace().Run("patch").Args("csv", sub.InstalledCSV, "-n", ns,
			"--type=json", "--patch", `[{"op":"replace","path":"/spec/webhookdefinitions/0/rules/0/operations", "value":["CREATE","DELETE"]}]`).Output()
		o.Expect(err1).NotTo(o.HaveOccurred())

		validatingwebhookName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", "-l", "olm.owner.namespace="+ns, "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", validatingwebhookName, "-o=jsonpath={..operations}").Output()
			e2e.Logf("output: %s", output)
			if err != nil {
				e2e.Logf("DELETE operations cannot be found:%v", err)
				return false, nil
			}
			if strings.Contains(output, "DELETE") {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", validatingwebhookName, "-o=jsonpath={.webhooks.rules}").Output()
			e2e.Logf("output: %s", output)
			output, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", sub.InstalledCSV, "-n", ns, "-o=jsonpath={.spec.webhookdefinitions}").Output()
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("validatingwebhookconfiguration %s has no DELETE operation", validatingwebhookName))
	})

	g.It("PolarionID:30317-PolarionID:30374-[OTP][Skipped:Disconnected]can allow mutating admission webhook definitions in CSV", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-30317",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-30317",
				Namespace:   "",
				DisplayName: "Test Catsrc 30317 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-operator-index-30317:v2-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-30317",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator-30317",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Template:               subTemplate,
			}
		)

		oc.SetupProject()
		ns := oc.Namespace()
		og.Namespace = ns
		sub.Namespace = ns
		catsrc.Namespace = ns
		sub.CatalogSourceName = catsrc.Name
		sub.CatalogSourceNamespace = ns

		g.By("create og")
		og.Create(oc, itName, dr)

		g.By("create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("install operator")
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		err := wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			err := oc.AsAdmin().WithoutNamespace().Run("get").Args("mutatingwebhookconfiguration", "-l", "olm.owner.namespace="+ns).Execute()
			if err != nil {
				e2e.Logf("The mutatingwebhookconfiguration is not created:%v", err)
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("mutatingwebhookconfiguration which owner ns %s is not created", ns))

		g.By("update csv")
		_, err1 := oc.AsAdmin().WithoutNamespace().Run("patch").Args("csv", sub.InstalledCSV, "-n", ns,
			"--type=json", "--patch", `[{"op":"replace","path":"/spec/webhookdefinitions/0/rules/0/operations", "value":["CREATE","DELETE"]}]`).Output()
		o.Expect(err1).NotTo(o.HaveOccurred())

		validatingwebhookName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("mutatingwebhookconfiguration", "-l", "olm.owner.namespace="+ns, "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("mutatingwebhookconfiguration", validatingwebhookName, "-o=jsonpath={..operations}").Output()
			e2e.Logf("output: %s", output)
			if err != nil {
				e2e.Logf("DELETE operations cannot be found:%v", err)
				return false, nil
			}
			if strings.Contains(output, "DELETE") {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("mutatingwebhookconfiguration", validatingwebhookName, "-o=jsonpath={.webhooks.rules}").Output()
			e2e.Logf("output: %s", output)
			output, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", sub.InstalledCSV, "-n", ns, "-o=jsonpath={.spec.webhookdefinitions}").Output()
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("mutatingwebhookconfiguration %s has no DELETE operation", validatingwebhookName))
	})

	g.It("PolarionID:30319-[OTP][Skipped:Disconnected]Admission Webhook Configuration names should be unique", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName                 = g.CurrentSpecReport().FullText()
			buildPruningBaseDir    = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate       = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
			subTemplate            = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			validatingwebhookName1 = ""
			validatingwebhookName2 = ""
			og                     = olmv0util.OperatorGroupDescription{
				Name:      "og-30319",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-30319",
				Namespace:   "",
				DisplayName: "Test Catsrc 30319 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-operator-index-30312:v2-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-operator-30319",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator-30312",
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Template:               subTemplate,
			}
		)

		for i := 1; i < 3; i++ {
			oc.SetupProject()
			ns := oc.Namespace()
			og.Namespace = ns
			sub.Namespace = ns
			catsrc.Namespace = ns
			sub.CatalogSourceName = catsrc.Name
			sub.CatalogSourceNamespace = ns

			g.By("create og")
			og.Create(oc, itName, dr)

			g.By("create catalog source")
			defer catsrc.Delete(itName, dr)
			catsrc.Create(oc, itName, dr)

			g.By("install operator")
			defer sub.Delete(itName, dr)
			sub.Create(oc, itName, dr)

			err := wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
				output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", "-l", "olm.owner.namespace="+ns).Output()
				if err != nil {
					e2e.Logf("The validatingwebhookconfiguration is not created:%v", err)
					return false, nil
				}
				if strings.Contains(output, "No resources") {
					e2e.Logf("The validatingwebhookconfiguration is not created:%v", err)
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", "--show-labels").Output()
				e2e.Logf("output: %s", output)
			}
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("validatingwebhookconfiguration which owner ns %s is not created", ns))

			validatingwebhookName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("validatingwebhookconfiguration", "-l", fmt.Sprintf("olm.owner.namespace=%s", ns), "-o=jsonpath={.items[0].metadata.name}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if i == 1 {
				validatingwebhookName1 = validatingwebhookName
			}
			if i == 2 {
				validatingwebhookName2 = validatingwebhookName
			}
		}
		o.Expect(validatingwebhookName1).NotTo(o.BeEmpty())
		o.Expect(validatingwebhookName2).NotTo(o.BeEmpty())
		o.Expect(validatingwebhookName2).NotTo(o.Equal(validatingwebhookName1))
	})

	g.It("PolarionID:40529-[OTP][Skipped:Disconnected]OPERATOR_CONDITION_NAME should have correct value", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		exutil.SkipBaselineCaps(oc, "None")
		var (
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		)

		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "og-40529",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-40529",
				Namespace:              namespaceName,
				CatalogSourceName:      "community-operators",
				CatalogSourceNamespace: "openshift-marketplace",
				Channel:                "singlenamespace-alpha",
				IpApproval:             "Manual",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
				StartingCSV:            "etcdoperator.v0.9.2",
			}
		)

		exists, err := olmv0util.ClusterPackageExists(oc, sub)
		o.Expect(err).NotTo(o.HaveOccurred())
		if !exists {
			g.Skip("SKIP:PackageMissing etcd does not exist in catalog community-operators")
		}

		itName := g.CurrentSpecReport().FullText()
		g.By("1: create the OperatorGroup ")
		og.Create(oc, itName, dr)

		g.By("2: create sub")
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		defer sub.Update(oc, itName, dr)

		sub.Create(oc, itName, dr)
		e2e.Logf("approve the install plan")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.2", "Complete")
		err = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.2", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
		if err != nil {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.2", "-n", namespaceName, "-o=jsonpath={.status.conditions}")
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, "state of csv etcdoperator.v0.9.2 is not Succeeded")

		g.By("3: check OPERATOR_CONDITION_NAME")
		err = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "etcdoperator.v0.9.2 etcdoperator.v0.9.2 etcdoperator.v0.9.2", exutil.Ok, []string{"deployment", "etcd-operator", "-n", namespaceName, "-o=jsonpath={.spec.template.spec.containers[*].env[?(@.name==\"OPERATOR_CONDITION_NAME\")].value}"}).CheckWithoutAssert(oc)
		if err != nil {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", "etcd-operator", "-n", namespaceName, "-o=jsonpath={..spec.template.spec.containers}")
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, "OPERATOR_CONDITION_NAME of etcd-operator is not correct")

		g.By("4: approve the install plan")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.4", "Complete")
		err = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.4", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
		if err != nil {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.4", "-n", namespaceName, "-o=jsonpath={.status.conditions}")
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, "state of csv etcdoperator.v0.9.4 is not Succeeded")

		g.By("5: check OPERATOR_CONDITION_NAME")
		err = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "etcdoperator.v0.9.4 etcdoperator.v0.9.4 etcdoperator.v0.9.4", exutil.Ok, []string{"deployment", "etcd-operator", "-n", namespaceName, "-o=jsonpath={.spec.template.spec.containers[*].env[?(@.name==\"OPERATOR_CONDITION_NAME\")].value}"}).CheckWithoutAssert(oc)
		if err != nil {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", "etcd-operator", "-n", namespaceName, "-o=jsonpath={..spec.template.spec.containers}")
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, "OPERATOR_CONDITION_NAME of etcd-operator is not correct")
	})

	g.It("PolarionID:40534-PolarionID:40532-[OTP][Skipped:Disconnected]the deployment should not lost the resources section", g.Label("NonHyperShiftHOST"), func() {
		var (
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		)

		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-40534",
				Namespace:   namespaceName,
				DisplayName: "Test Catsrc 40534 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-40534-operator",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-40534",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)

		itName := g.CurrentSpecReport().FullText()
		g.By("STEP 1: create the OperatorGroup and catalog source")
		og.Create(oc, itName, dr)
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("STEP 2: create sub")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator", exutil.Ok, []string{"deployment", "-n", sub.Namespace}).Check(oc)

		g.By("STEP 3: check OPERATOR_CONDITION_NAME")
		cpuCSV := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={..containers[?(@.name==\"manager\")].resources.requests.cpu}")
		o.Expect(cpuCSV).NotTo(o.BeEmpty())
		memoryCSV := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={..containers[?(@.name==\"manager\")].resources.requests.memory}")
		o.Expect(memoryCSV).NotTo(o.BeEmpty())
		cpuDeployment := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..containers[?(@.name==\"manager\")].resources.requests.cpu}")
		o.Expect(cpuDeployment).To(o.Equal(cpuCSV))
		memoryDeployment := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..containers[?(@.name==\"manager\")].resources.requests.memory}")
		o.Expect(memoryDeployment).To(o.Equal(memoryCSV))

		g.By("OCP-40532: OLM should not print debug logs")
		olmPodname, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "--selector=app=olm-operator", "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(olmPodname).NotTo(o.BeEmpty())
		olmlogs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args(olmPodname, "-n", "openshift-operator-lifecycle-manager", "--limit-bytes", "50000").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(olmlogs).NotTo(o.BeEmpty())
		o.Expect(olmlogs).NotTo(o.ContainSubstring("level=debug"))

		catPodname, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "--selector=app=catalog-operator", "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(catPodname).NotTo(o.BeEmpty())
		catalogs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args(catPodname, "-n", "openshift-operator-lifecycle-manager", "--limit-bytes", "50000").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(catalogs).NotTo(o.BeEmpty())
		o.Expect(catalogs).NotTo(o.ContainSubstring("level=debug"))
	})

	g.It("PolarionID:40531-PolarionID:41051-PolarionID:23172-[OTP][Skipped:Disconnected]the value of lastUpdateTime of csv and Components of Operator should be correct[Serial]", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		olmv0util.SkipIfPackagemanifestNotExist(oc, "learn")
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipForSNOCluster(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			sub                 = olmv0util.SubscriptionDescription{
				SubName:                "sub-40531",
				Namespace:              "openshift-operators",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "learn",
				CatalogSourceName:      "qe-app-registry",
				CatalogSourceNamespace: "openshift-marketplace",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
		)

		g.By("1, Check if the global operator global-operators support all namesapces")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, `[""]`, exutil.Ok, []string{"og", "global-operators", "-n", "openshift-operators", "-o=jsonpath={.status.namespaces}"}).Check(oc)

		g.By("2, Create operator targeted at all namespace")
		defer sub.Delete(itName, dr)
		defer func() {
			if sub.InstalledCSV == "" {
				sub.FindInstalledCSV(oc, itName, dr)
			}
			sub.DeleteCSV(itName, dr)
		}()
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3, Create new namespace")
		oc.SetupProject()

		g.By("4, OCP-23172 Check the csv within new namespace is copied.")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Copied", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.reason}"}).Check(oc)

		g.By("5, OCP-40531-Check the lastUpdateTime of copied CSV is equal to the original CSV.")
		originCh := make(chan string)
		defer close(originCh)
		copyCh := make(chan string)
		defer close(copyCh)
		go func() {
			originCh <- olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", "openshift-operators", "-o=jsonpath={.status.lastUpdateTime}")
		}()
		go func() {
			copyCh <- olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.lastUpdateTime}")
		}()
		lastUpdateTimeOrigin := <-originCh
		lastUpdateTimeNew := <-copyCh
		e2e.Logf("OriginTimeStamp:%s, CopiedTimeStamp:%s", lastUpdateTimeOrigin, lastUpdateTimeNew)
		o.Expect(lastUpdateTimeNew).To(o.Equal(lastUpdateTimeOrigin))

		g.By("6, OCP-41051-Check Operator.Status.Components does not contain copied CSVs.")
		operatorname := sub.OperatorPackage + ".openshift-operators"
		operatorinfo := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", operatorname, "-n", oc.Namespace(), "-o=jsonpath={.status.components.refs}")
		o.Expect(operatorinfo).NotTo(o.BeEmpty())
		o.Expect(operatorinfo).NotTo(o.ContainSubstring("Copied"))
	})

	g.It("PolarionID:41035-[OTP][Skipped:Disconnected]Fail InstallPlan on bundle unpack timeout[Slow]", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-41035",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-41035",
				Namespace:   "",
				DisplayName: "Test Catsrc 41035 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/ditto-index:41035",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "ditto-operator-41035",
				Namespace:              "",
				Channel:                "4.8",
				IpApproval:             "Automatic",
				OperatorPackage:        "ditto-operator",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		oc.SetupProject() // project and its resource are deleted automatically when out of It, so no need derfer or AfterEach
		og.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace

		g.By("create og")
		og.Create(oc, itName, dr)

		g.By("create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("install operator")
		defer sub.Delete(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("The install plan is Failed")
		err := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 900*time.Second, false, func(ctx context.Context) (bool, error) {
			conditions, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}").Output()
			if strings.Contains(conditions, "BundleUnpackFailed") {
				return true, nil
			}
			return false, nil
		})
		olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}")
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("sub %v is not Failed", sub.SubName))
	})

	g.It("PolarionID:42829-[OTP][Skipped:Disconnected]Install plan should be blocked till a valid OperatorGroup is detected", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		oc.SetupProject()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: oc.Namespace(),
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-42829",
				Namespace:   oc.Namespace(),
				DisplayName: "Test Operators",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-42829",
				Namespace:              oc.Namespace(),
				CatalogSourceName:      "catsrc-42829",
				CatalogSourceNamespace: oc.Namespace(),
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By(fmt.Sprintf("1) create the catsrc in project: %s", oc.Namespace()))
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("2) install sub")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("sleep 10 sencond, then create og")
		time.Sleep(time.Second * 10)

		g.By("4) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("check ip and csv")
		installPlan := sub.GetIP(oc)
		o.Expect(installPlan).NotTo(o.BeEmpty())
		err := olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Complete", exutil.Ok, []string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
		if err != nil {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}")
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, "status.phase of installplan is not Complete")
		sub.FindInstalledCSV(oc, itName, dr)
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") == 0 {
				e2e.Logf("get installedCSV failed")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("csv %s is not Succeeded", sub.InstalledCSV))
	})

	g.It("PolarionID:43110-[OTP][Skipped:Disconnected]OLM provide a helpful error message when install removed api", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-ditto-43110",
				Namespace:   namespaceName,
				DisplayName: "Test Catsrc ditto Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/ditto-index:v1beta1-cache",
				Template:    catsrcImageTemplate,
			}
			og = olmv0util.OperatorGroupDescription{
				Name:      "og-43110",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-43110",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-ditto-43110",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "ditto-operator",
				SingleNamespace:        true,
				Template:               subTemplate,
				StartingCSV:            "",
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By("1) create the catalog source and OperatorGroup")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) install sub")
		defer sub.Delete(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("3) check ip/sub conditions")
		installPlan := sub.GetIP(oc)
		o.Expect(installPlan).NotTo(o.BeEmpty())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Failed", exutil.Ok, []string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		ipConditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}")
		o.Expect(ipConditions).To(o.ContainSubstring("api-server resource not found installing CustomResourceDefinition"))
		o.Expect(ipConditions).To(o.ContainSubstring("apiextensions.k8s.io/v1beta1"))
		o.Expect(ipConditions).To(o.ContainSubstring("Kind=CustomResourceDefinition not found on the cluster"))
		o.Expect(ipConditions).To(o.ContainSubstring("InstallComponentFailed"))

		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			subConditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}")
			if strings.Contains(subConditions, "InstallComponentFailed") {
				o.Expect(subConditions).To(o.ContainSubstring("Kind=CustomResourceDefinition not found on the cluster"))
				return true, nil
			}
			e2e.Logf("subscription conditions: %s", subConditions)
			e2e.Logf("the status message of sub is not correct, retry...")
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "sub status is not correct")
		g.By("4) SUCCESS")
	})

	g.It("PolarionID:43114-[OTP][Skipped:Disconnected]Subscription status should show the message for InstallPlan failure conditions", g.Label("ReleaseGate"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSAtemplate := filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		namespace := "ns-43114"
		defer func() {
			_ = oc.WithoutNamespace().AsAdmin().Run("delete").Args("ns", namespace, "--ignore-not-found").Execute()
		}()
		err := oc.WithoutNamespace().AsAdmin().Run("create").Args("ns", namespace).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		og := olmv0util.OperatorGroupDescription{
			Name:               "test-og-43114",
			Namespace:          namespace,
			ServiceAccountName: "scoped-43114",
			Template:           ogSAtemplate,
		}
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-43114",
			Namespace:   namespace,
			DisplayName: "Test Catsrc 43114 Operators",
			Publisher:   "Red Hat",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
			Template:    catsrcImageTemplate,
		}

		sub := olmv0util.SubscriptionDescription{
			SubName:                "nginx-operator-43114",
			Namespace:              namespace,
			Channel:                "alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "nginx-operator",
			CatalogSourceName:      catsrc.Name,
			CatalogSourceNamespace: namespace,
			Template:               subTemplate,
			SingleNamespace:        true,
		}

		itName := g.CurrentSpecReport().FullText()

		g.By("1) Create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("3) Create a Subscription")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("4) check install plan message")
		ip := sub.GetIP(oc)
		msg := ""
		errorText := "no operator group found"
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("installplan", ip, "-n", sub.Namespace, "-o=jsonpath={..status.conditions}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(strings.ToLower(msg), errorText) {
				e2e.Logf("InstallPlan has the expected error")
				return true, nil
			}
			e2e.Logf("message: %s", msg)
			return false, nil
		})
		exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("The installplan %s did not include expected message.  The message was instead %s", ip, msg))

		g.By("5) Check sub message")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			subConditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}")
			if strings.Contains(strings.ToLower(subConditions), errorText) {
				return true, nil
			}
			e2e.Logf("subscription conditions: %s", subConditions)
			e2e.Logf("the status message of sub is not correct, retry...")
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "subscription also has the expected error")
		g.By("Finished")

	})

	g.It("PolarionID:43639-[OTP][Skipped:Disconnected]OLM must explicitly alert on deprecated APIs in use", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-ditto-43639",
				Namespace:   namespaceName,
				DisplayName: "Test Catsrc ditto Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/ditto-index:v1beta1-cache",
				Template:    catsrcImageTemplate,
			}
			og = olmv0util.OperatorGroupDescription{
				Name:      "og-43639",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-43639",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-ditto-43639",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "ditto-operator",
				SingleNamespace:        true,
				Template:               subTemplate,
				StartingCSV:            "",
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By("1) create the catalog source and OperatorGroup")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) install sub")
		defer sub.Delete(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)
		installPlan := sub.GetIP(oc)
		o.Expect(installPlan).NotTo(o.BeEmpty())
		err := wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
			ipPhase := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Contains(ipPhase, "Complete") {
				e2e.Logf("sub is installed")
				return true, nil
			}
			return false, nil
		})
		if err == nil {
			g.By("3) check events")
			err2 := wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 240*time.Second, false, func(ctx context.Context) (bool, error) {
				eventOutput, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("event", "-n", namespaceName).Output()
				o.Expect(err1).NotTo(o.HaveOccurred())
				lines := strings.Split(eventOutput, "\n")
				for _, line := range lines {
					if strings.Contains(line, "CustomResourceDefinition is deprecated") && strings.Contains(line, "piextensions.k8s.io") && strings.Contains(line, "ditto-operator") {
						return true, nil
					}
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err2, "event CustomResourceDefinition is deprecated, piextensions.k8s.io and ditto-operator not found")

		} else {
			g.By("3) the opeartor cannot be installed, skip test case")
		}

		g.By("4) SUCCESS")
	})

	g.It("PolarionID:48439-[OTP][Skipped:Disconnected]OLM upgrades operators immediately", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-48439",
				Namespace:   namespaceName,
				DisplayName: "Test Catsrc",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:ocp-48439",
				Template:    catsrcImageTemplate,
			}
			og = olmv0util.OperatorGroupDescription{
				Name:      "og-48439",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-48439",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-48439",
				CatalogSourceNamespace: namespaceName,
				Channel:                "v0.0.1",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				Template:               subTemplate,
				StartingCSV:            "nginx-operator.v0.0.1",
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By("1) create the catalog source and OperatorGroup")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) install sub")
		sub.Create(oc, itName, dr)
		_ = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx-operator.v0.0.1", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)

		g.By("3) update sub channel")
		sub.Patch(oc, "{\"spec\": {\"channel\": \"v1.0.1\"}}")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			ips := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", "-n", sub.Namespace)
			if strings.Contains(ips, "nginx-operator.v1.0.1") {
				e2e.Logf("Install plan for nginx-operator.v1.0.1 is created")
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath-as-json={.spec}")
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath-as-json={.status}")
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", "-n", sub.Namespace, "-o=jsonpath-as-json={..spec}")
		}
		exutil.AssertWaitPollNoErr(err, "no install plan for nginx-operator.v1.0.1")
		_ = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx-operator.v1.0.1", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
		g.By("4) SUCCESS")
	})

	g.It("PolarionID:47322-[OTP][Skipped:Disconnected]Arbitrary Constraints can be defined as bundle properties", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-47322",
				Namespace:   namespaceName,
				DisplayName: "Test 47322",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/etcd-index:47322-single-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "etcd-47322",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-47322",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha-1",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("2) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) install sub with channel alpha-1")
		sub.Create(oc, itName, dr)

		g.By("4) check csv")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.2", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status1, "Succeeded") != 0 {
				e2e.Logf("csv etcdoperator.v0.9.2 status is not Succeeded, go next round")
				return false, nil
			}
			status2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "ditto-operator.v0.1.1", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if (strings.Compare(status2, "Succeeded") != 0) && (strings.Compare(status2, "Installing") != 0) {
				e2e.Logf("csv ditto-operator.v0.1.1 status is not Succeeded nor Installing, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv etcdoperator.v0.9.2 or ditto-operator.v0.1.1 is not Succeeded nor Installing")

		g.By("5) delete sub etcd-47322 and csv etcdoperator.v0.9.2")
		sub.FindInstalledCSV(oc, itName, dr)
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)

		g.By("6) install sub with channel alpha-2")
		sub.Channel = "alpha-2"
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("7) check sub")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "ConstraintsNotSatisfiable", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].reason}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "require to have the property olm.type3 with value value31", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}"}).Check(oc)

		g.By("8) delete sub and csv ditto-operator.v0.1.1")
		selectorStr := "--selector=operators.coreos.com/ditto-operator." + namespaceName
		subDepName := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", selectorStr, "-n", sub.Namespace, "-o=jsonpath={..metadata.name}")
		o.Expect(subDepName).To(o.ContainSubstring("ditto-operator"))
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("sub", subDepName, "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("csv", "ditto-operator.v0.1.1", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", sub.Namespace)
			if strings.Contains(output, "ditto-operator.v0.1.1") {
				e2e.Logf("csv ditto-operator.v0.1.1 still exist, go next round")
				return false, nil
			}
			output = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", "-n", sub.Namespace)
			if strings.Contains(output, subDepName) {
				e2e.Logf("sub still exist, go next round")
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "delete sub and csv failed")

		g.By("9) check status of csv etcdoperator.v0.9.4 and ditto-operator.v0.2.0")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.4", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status1, "Succeeded") == 0 {
				e2e.Logf("csv etcdoperator.v0.9.4 status is Succeeded")
				return true, nil
			}
			e2e.Logf("csv etcdoperator.v0.9.4 status is not Succeeded, go next round")
			return false, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.SubName, "-n", namespaceName)
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName)
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv etcdoperator.v0.9.4 is not Succeeded")

		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "ditto-operator.v0.2.0", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if (strings.Compare(status2, "Succeeded") == 0) || (strings.Compare(status2, "Installing") == 0) {
				e2e.Logf("csv ditto-operator.v0.2.0 status is Succeeded")
				return true, nil
			}
			e2e.Logf("csv ditto-operator.v0.2.0 status is not Succeeded nor Installing, go next round")
			return false, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", sub.SubName, "-n", namespaceName)
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName)
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv ditto-operator.v0.2.0 is not Succeeded nor Installing")

	})

	g.It("PolarionID:47319-[OTP][Skipped:Disconnected]olm raised error when Arbitrary Compound Constraints is defined wrongly", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}

			catsrcError = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-47319-error",
				Namespace:   namespaceName,
				DisplayName: "Test 47319",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/etcd-index:47319-error",
				Template:    catsrcImageTemplate,
			}
			subError = olmv0util.SubscriptionDescription{
				SubName:                "etcd-47319-error",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-47319-error",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha-1",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrcError.Delete(itName, dr)
		catsrcError.CreateWithCheck(oc, itName, dr)

		g.By("2) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) install subError with channel alpha-1")
		subError.CreateWithoutCheck(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "ErrorPreventedResolution", exutil.Ok, []string{"sub", subError.SubName, "-n", namespaceName, "-o=jsonpath={.status.conditions[*].reason}"}).Check(oc)
		conditionsMsg := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", subError.SubName, "-n", namespaceName, "-o=jsonpath={.status.conditions[*].message}")
		o.Expect(conditionsMsg).To(o.ContainSubstring("convert olm.constraint to resolver predicate: ERROR"))
		subError.Delete(itName, dr)

		g.By("4) install subError with channel alpha-2")
		subError.Channel = "alpha-2"
		subError.CreateWithoutCheck(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "ConstraintsNotSatisfiable", exutil.Ok, []string{"sub", subError.SubName, "-n", namespaceName, "-o=jsonpath={.status.conditions[*].reason}"}).Check(oc)
		conditionsMsg = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", subError.SubName, "-n", namespaceName, "-o=jsonpath={.status.conditions[*].message}")
		o.Expect(conditionsMsg).To(o.MatchRegexp("(?i)require to have .*olm.type3.* and olm.package ditto-operator with version >= 0.2.1(?i)"))
		subError.Delete(itName, dr)
	})

	g.It("PolarionID:47319-[OTP][Skipped:Disconnected]Arbitrary Compound Constraints with AND can be defined as bundle properties with less than", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-47319",
				Namespace:   namespaceName,
				DisplayName: "Test 47319",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/etcd-index:47319-and",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "etcd-47319",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-47319",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha-1",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("2) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) install sub with channel alpha-1")
		sub.Create(oc, itName, dr)

		g.By("4) check csv")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.2", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status1, "Succeeded") != 0 {
				e2e.Logf("csv etcdoperator.v0.9.2 status is not Succeeded, go next round")
				return false, nil
			}
			status2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "ditto-operator.v0.1.1", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if (strings.Compare(status2, "Succeeded") != 0) && (strings.Compare(status2, "Installing") != 0) {
				e2e.Logf("csv ditto-operator.v0.1.1 status is not Succeeded nor Installing, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv etcdoperator.v0.9.2 or ditto-operator.v0.1.1 is not Succeeded nor Installing")

	})

	g.It("PolarionID:47319-[OTP][Skipped:Disconnected]Arbitrary Compound Constraints with AND can be defined as bundle properties with more than", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-47319",
				Namespace:   namespaceName,
				DisplayName: "Test 47319",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/etcd-index:47319-and",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "etcd-47319",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-47319",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha-2",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("2) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("5) install sub with channel alpha-1")
		sub.Create(oc, itName, dr)

		g.By("6) check csv")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.4", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status1, "Succeeded") != 0 {
				e2e.Logf("csv etcdoperator.v0.9.4 status is not Succeeded, go next round")
				return false, nil
			}
			status2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "ditto-operator.v0.2.0", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if (strings.Compare(status2, "Succeeded") != 0) && (strings.Compare(status2, "Installing") != 0) {
				e2e.Logf("csv ditto-operator.v0.1.1 status is not Succeeded nor Installing, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv etcdoperator.v0.9.4 or ditto-operator.v0.2.0 is not Succeeded or Installing")
	})

	g.It("PolarionID:47323-[OTP][Skipped:Disconnected]Arbitrary Compound Constraints with OR can be defined as bundle properties", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrcOr = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-47323-or",
				Namespace:   namespaceName,
				DisplayName: "Test 47323 OR",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/etcd-index:47323-or-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "etcd-47323",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-47323-or",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha-1",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrcOr.Delete(itName, dr)
		catsrcOr.CreateWithCheck(oc, itName, dr)

		g.By("2) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) test arbitrary compound constraints with OR")
		g.By("3.1) install sub with channel alpha-1")
		sub.Create(oc, itName, dr)

		g.By("3.2) check csv")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.2", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status1, "Succeeded") != 0 {
				e2e.Logf("csv etcdoperator.v0.9.2 status is not Succeeded, go next round")
				return false, nil
			}
			status2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "ditto-operator.v0.1.0", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if (strings.Compare(status2, "Succeeded") != 0) && (strings.Compare(status2, "Installing") != 0) {
				e2e.Logf("csv ditto-operator.v0.1.0 status is not Succeeded nor Installing, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv etcdoperator.v0.9.2 or ditto-operator.v0.1.0 is not Succeeded")

		g.By("3.3) switch channel to be alpha-2")
		sub.Patch(oc, "{\"spec\": {\"channel\": \"alpha-2\"}}")

		g.By("3.4) check csv")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.4", "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3.4) delete all subs and csvs")
		sub.FindInstalledCSV(oc, itName, dr)
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		selectorStr := "--selector=operators.coreos.com/ditto-operator." + namespaceName
		subDepName := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", selectorStr, "-n", sub.Namespace, "-o=jsonpath={..metadata.name}")
		o.Expect(subDepName).To(o.ContainSubstring("ditto-operator"))
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("sub", subDepName, "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("csv", "ditto-operator.v0.1.0", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", sub.Namespace)
			if strings.Contains(output, "ditto-operator.v0.1.0") {
				e2e.Logf("csv ditto-operator.v0.1.0 still exist, go next round")
				return false, nil
			}
			output = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", "-n", sub.Namespace)
			if strings.Contains(output, subDepName) {
				e2e.Logf("sub still exist, go next round")
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "delete sub and csv failed")
	})

	g.It("PolarionID:47323-[OTP][Skipped:Disconnected]Arbitrary Compound Constraints with NOT can be defined as bundle properties", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrcNot = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-47323-not",
				Namespace:   namespaceName,
				DisplayName: "Test 47323 NOT",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/etcd-index:47323-not-cache",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "etcd-47323",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-47323-not",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha-1",
				IpApproval:             "Automatic",
				OperatorPackage:        "etcd",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrcNot.Delete(itName, dr)
		catsrcNot.CreateWithCheck(oc, itName, dr)

		g.By("2) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) test arbitrary compound constraints with Not")
		g.By("3.1) install sub with channel alpha-1")
		sub.Channel = "alpha-1"
		sub.Create(oc, itName, dr)

		g.By("3.2) check csv")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "etcdoperator.v0.9.2", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status1, "Succeeded") != 0 {
				e2e.Logf("csv etcdoperator.v0.9.2 status is not Succeeded, go next round")
				return false, nil
			}
			status2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "ditto-operator.v0.1.0", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if (strings.Compare(status2, "Succeeded") != 0) && (strings.Compare(status2, "Installing") != 0) {
				e2e.Logf("csv ditto-operator.v0.1.0 status is not Succeeded nor Installing, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv etcdoperator.v0.9.2 or ditto-operator.v0.1.0 is not Succeeded")

		g.By("3.3) delete sub etcd-47323 and csv etcdoperator.v0.9.2")
		sub.FindInstalledCSV(oc, itName, dr)
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)

		g.By("3.4) install sub with channel alpha-2")
		sub.Channel = "alpha-2"
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("3.5) check sub")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "ConstraintsNotSatisfiable", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].reason}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "require to not have ", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}"}).Check(oc)

	})

	g.It("PolarionID:56371-[OTP][Skipped:Disconnected]service account token secret reference", func() {
		exutil.SkipMissingQECatalogsource(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		roletemplate := filepath.Join(buildPruningBaseDir, "role.yaml")
		rolebindingtemplate := filepath.Join(buildPruningBaseDir, "role-binding.yaml")
		ogSAtemplate := filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		secretTemplate := filepath.Join(buildPruningBaseDir, "secret.yaml")
		secretopaqueTemplate := filepath.Join(buildPruningBaseDir, "secret_opaque.yaml")
		oc.SetupProject()
		namespace := oc.Namespace()
		itName := g.CurrentSpecReport().FullText()
		var (
			sa = "scoped-56371"
			og = olmv0util.OperatorGroupDescription{
				Name:               "test-og-56371",
				Namespace:          namespace,
				ServiceAccountName: sa,
				Template:           ogSAtemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-56371",
				Namespace:              namespace,
				CatalogSourceName:      "qe-app-registry",
				CatalogSourceNamespace: "openshift-marketplace",
				Channel:                "beta",
				IpApproval:             "Automatic",
				OperatorPackage:        "learn",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
			role = olmv0util.RoleDescription{
				Name:      "role-56371",
				Namespace: namespace,
				Template:  roletemplate,
			}
			rolebinding = olmv0util.RolebindingDescription{
				Name:      "scoped-bindings-56371",
				Namespace: namespace,
				Rolename:  "role-56371",
				Saname:    sa,
				Template:  rolebindingtemplate,
			}
			secretopaque = olmv0util.SecretDescription{
				Name:      "zsecret-56371",
				Namespace: namespace,
				Template:  secretopaqueTemplate,
			}
			secret = olmv0util.SecretDescription{
				Name:      sa,
				Namespace: namespace,
				Saname:    sa,
				Sectype:   "kubernetes.io/service-account-token",
				Template:  secretTemplate,
			}
		)

		g.By("1) Create the service account")
		_, err := oc.WithoutNamespace().AsAdmin().Run("create").Args("sa", sa, "-n", sub.Namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		secret.Create(oc)

		g.By("2) Create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)
		err = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, sa, exutil.Ok, []string{"og", og.Name, "-n", og.Namespace, "-o=jsonpath={.status.serviceAccountRef.name}"}).CheckWithoutAssert(oc)
		if err != nil {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "og", og.Name, "-n", og.Namespace, "-o=jsonpath={.status}")
			e2e.Logf("output: %s", output)
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("status.serviceAccountRef.name of og %s is not %s", og.Name, sa))

		g.By("3) Create the Secret")
		secretopaque.Create(oc)

		g.By("4) Grant the proper permissions to the service account")
		role.Create(oc)
		rolebinding.Create(oc)

		g.By("5) create sub")
		olmv0util.SkipIfPackagemanifestNotExist(oc, sub.OperatorPackage)
		sub.Create(oc, itName, dr)

		g.By("6) Checking the secret")
		secrets, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("secret", "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(secrets).To(o.ContainSubstring(secretopaque.Name))

		g.By("7) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

	})

	g.It("PolarionID:59380-PolarionID:68671-[OTP][Skipped:Disconnected]Upgrade should be success when there are multiple upgrade paths between channel entries", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")

		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}

			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-59380",
				Namespace:   namespaceName,
				DisplayName: "Test-Catsrc-59380-Operators",
				Publisher:   "Red-Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:59380",
				Template:    catsrcImageTemplate,
			}
			subManual = olmv0util.SubscriptionDescription{
				SubName:                "sub-59380",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-59380",
				CatalogSourceNamespace: namespaceName,
				Channel:                "v1.6",
				IpApproval:             "Manual",
				OperatorPackage:        "nginx-operator",
				StartingCSV:            "nginx-operator.v1.6.0",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)

		itName := g.CurrentSpecReport().FullText()
		g.By("STEP1: create the OperatorGroup ")
		og.CreateWithCheck(oc, itName, dr)

		g.By("STEP 2: Create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("STEP 3: install operator ")
		subManual.CreateWithoutCheck(oc, itName, dr)

		g.By("OCP-68671 Only one operator name is in 'Manual approval required' info section")
		nameIP := subManual.GetIP(oc)
		clusterServiceVersionNames, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("installplan", nameIP, "-o=jsonpath={.spec.clusterServiceVersionNames}", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(clusterServiceVersionNames)).NotTo(o.ContainSubstring(`",`))
		g.By("OCP-68671 SUCCESS")

		e2e.Logf("approve the install plan")
		subManual.ApproveSpecificIP(oc, itName, dr, "nginx-operator.v1.6.0", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx-operator.v1.6.0", "-n", subManual.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("STEP 4: approve the install plan")
		subManual.ApproveSpecificIP(oc, itName, dr, "nginx-operator.v1.6.2", "Complete")

		g.By("STEP 5: check the csv nginx-operator.v1.6.2")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx-operator.v1.6.2", "-n", subManual.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

	})

	g.It("PolarionID:73061-[OTP][Skipped:Disconnected]Support envfrom on Operator Lifecycle Manager", g.Label("NonHyperShiftHOST"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "envfrom-subscription.yaml")
		cmTemplate := filepath.Join(buildPruningBaseDir, "cm-template.yaml")
		secretTemplate := filepath.Join(buildPruningBaseDir, "secret_opaque.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}

			cm = olmv0util.ConfigMapDescription{
				Name:      "special-config-73061",
				Namespace: namespaceName,
				Template:  cmTemplate,
			}
			secret = olmv0util.SecretDescription{
				Name:      "special-secret-73061",
				Namespace: namespaceName,
				Template:  secretTemplate,
			}

			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-73061",
				Namespace:   namespaceName,
				DisplayName: "Test Catsrc 73061 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-73061-operator",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-73061",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-operator",
				ConfigMapRef:           "special-config-73061",
				SecretRef:              "special-secret-73061",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By("STEP 1: create the OperatorGroup, catalog source, secret, configmap")
		og.CreateWithCheck(oc, itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)
		cm.Create(oc, itName, dr)
		secret.Create(oc)

		g.By("STEP 2: create sub")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "nginx-operator", exutil.Ok, []string{"deployment", "-n", sub.Namespace}).Check(oc)

		g.By("STEP 3: check deployment")
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			envFromDeployment := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, `-o=jsonpath='{..spec.containers}'`)
			if !strings.Contains(envFromDeployment, cm.Name) || !strings.Contains(envFromDeployment, secret.Name) {
				e2e.Logf("envFrom deployment: %s", envFromDeployment)
				return false, nil
			}
			return true, nil
		})
		if waitErr != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "deployment", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-oyaml")
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-oyaml")
		}
		exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("deployment doesn't contain %s, %s", cm.Name, secret.Name))

		g.By("STEP 4: check pod")
		envFromPod := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "pod", "--selector=control-plane=controller-manager", "-n", sub.Namespace, `-o=jsonpath='{..spec.containers}'`)
		if !strings.Contains(envFromPod, cm.Name) || !strings.Contains(envFromPod, secret.Name) {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "pod", "--selector=control-plane=controller-manager", "-n", sub.Namespace, "-oyaml")
		}
		o.Expect(envFromPod).To(o.ContainSubstring(cm.Name))
		o.Expect(envFromPod).To(o.ContainSubstring(secret.Name))
	})

	g.It("PolarionID:81389-[OTP][Skipped:Disconnected]Validating existing CRs against new CRD schema should be success", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		packageName, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "-l", "catalog=community-operators", "--field-selector", "metadata.name=postgresql").Output()
		if !strings.Contains(packageName, "postgresql") {
			g.Skip("no reqruied package postgresql, so skip it")
		}

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		crTemplate := filepath.Join(buildPruningBaseDir, "cr_pgadmin.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "og-81389",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-81389",
				Namespace:              namespaceName,
				CatalogSourceName:      "community-operators",
				CatalogSourceNamespace: "openshift-marketplace",
				Channel:                "v5",
				IpApproval:             "Automatic",
				OperatorPackage:        "postgresql",
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By("1) create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) install sub")
		sub.Create(oc, itName, dr)
		_ = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)

		g.By("3) create cr")
		crFile, err := oc.AsAdmin().Run("process").Args("--ignore-unknown-parameters=true", "-f", crTemplate, "-p", "NAMESPACE="+namespaceName).OutputToFile("cr-81389.json")
		o.Expect(err).NotTo(o.HaveOccurred())
		defer func() { _ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", crFile).Execute() }()
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", crFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			observedGeneration := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "PGAdmin", "pgadmin-example", "-n", sub.Namespace, "-o=jsonpath-as-json={.status.observedGeneration}")
			if strings.Contains(observedGeneration, "1") {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "PGAdmin", "-n", sub.Namespace, "-o=jsonpath-as-json={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "create cr failed")

		g.By("4) delete sub/csv")
		sub.DeleteCSV(itName, dr)
		sub.Delete(itName, dr)

		g.By("5) reinstall sub")
		sub.Create(oc, itName, dr)
		_ = olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
	})

	g.It("PolarionID:82135-[OTP][Skipped:Disconnected]Verify NetworkPolicy resources is supported", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")

		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}

			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-82135",
				Namespace:   namespaceName,
				DisplayName: "Test-82135-Operators",
				Publisher:   "Red-Hat",
				SourceType:  "grpc",
				Address:     "quay.io/openshifttest/nginxolm-operator-index:nginxolm82135",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-82135",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-82135",
				CatalogSourceNamespace: namespaceName,
				Channel:                "alpha",
				IpApproval:             "Manual",
				OperatorPackage:        "nginx82135",
				StartingCSV:            "nginx82135.v1.0.1",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		itName := g.CurrentSpecReport().FullText()
		g.By("STEP 1: Create catalog source")
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("STEP 2: create the OperatorGroup ")
		og.CreateWithCheck(oc, itName, dr)

		g.By("STEP 3: install operator")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("STEP 4: approve install plan for 1.0.1, no networkpolicy in bundle")
		sub.ApproveSpecificIP(oc, itName, dr, "nginx82135.v1.0.1", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx82135.v1.0.1", "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		networkpolicies := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "networkpolicy", "-n", sub.Namespace)
		o.Expect(networkpolicies).To(o.ContainSubstring("grpc-server"))
		o.Expect(networkpolicies).To(o.ContainSubstring("unpack-bundles"))

		g.By("STEP 4: approve install plan for 1.1.0, 1 networkpolicy in bundle")
		sub.ApproveSpecificIP(oc, itName, dr, "nginx82135.v1.1.0", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx82135.v1.1.0", "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		networkpolicies = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "networkpolicy", "-n", sub.Namespace)
		o.Expect(networkpolicies).To(o.ContainSubstring("nginx82135-controller-acceptall"))

		g.By("STEP 5: approve install plan for 2.0.0, 2 networkpolicies in bundle")
		sub.ApproveSpecificIP(oc, itName, dr, "nginx82135.v2.0.0", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "nginx82135.v2.0.0", "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
		networkpolicies = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "networkpolicy", "-n", sub.Namespace)
		o.Expect(networkpolicies).To(o.ContainSubstring("nginx82135-controller"))
		o.Expect(networkpolicies).To(o.ContainSubstring("default-deny-all"))
		o.Expect(networkpolicies).NotTo(o.ContainSubstring("nginx82135-controller-acceptall"))

		g.By("STEP 6: approve install plan for 2.1.0, wrong networkpolicies in bundle")
		sub.ApproveSpecificIP(oc, itName, dr, "nginx82135.v2.1.0", "Failed")
		err := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 900*time.Second, false, func(ctx context.Context) (bool, error) {
			conditions, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}").Output()
			if strings.Contains(conditions, "Unsupported value") {
				return true, nil
			}
			return false, nil
		})
		olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}")
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("conditions of sub %v is not correct", sub.SubName))

	})

	g.It("PolarionID:69986-[OTP][Skipped:Disconnected]OLM emits alert events for operators installed from a deprecated channel", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-69986",
				Namespace:   namespaceName,
				DisplayName: "Test 69986",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olmtest-operator-index:nginx69986",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-69986",
				Namespace:              namespaceName,
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Channel:                "",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx69986",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		sub.CatalogSourceNamespace = catsrc.Namespace
		sub.CatalogSourceName = catsrc.Name
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", namespaceName))
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("2) packagemanifests")
		message := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifests", "nginx69986", "-n", catsrc.Namespace, `-o=jsonpath='{.status.channels[?(@.name=="candidate-v0.0")].deprecation}'`)
		o.Expect(string(message)).To(o.ContainSubstring(`has been deprecated`))
		message = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifests", "nginx69986", "-n", catsrc.Namespace, `-o=jsonpath={.status.channels[?(@.name=="candidate-v1.0")].entries[?(@.name=="nginx69986.v1.0.3")].deprecation}`)
		o.Expect(string(message)).To(o.ContainSubstring(`has been deprecated`))

		g.By("3) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("4) install sub with channel candidate-v0.0")
		sub.Channel = "candidate-v0.0"
		sub.Create(oc, itName, dr)

		g.By("4.1 check csv")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "nginx69986.v0.0.1", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") != 0 {
				e2e.Logf("csv nginx69986.v0.0.1 status is not Succeeded, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv nginx69986.v0.0.1 is not Succeeded")

		g.By("4.2 check sub status")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			conditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].type}")
			if !strings.Contains(conditions, "ChannelDeprecated") || !strings.Contains(conditions, "Deprecated") {
				return false, nil
			}
			messages := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}")
			if !strings.Contains(messages, "has been deprecated. Please switch to a different one") {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status.conditions}")
		}
		exutil.AssertWaitPollNoErr(err, "the conditions of sub is not correct")

		g.By("4.3) delete sub and csv")
		sub.FindInstalledCSV(oc, itName, dr)
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)

		g.By("5) install sub with channel candidate-v1.0")
		sub.Channel = "candidate-v1.0"
		sub.StartingCSV = "nginx69986.v1.0.2"
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("5.1 check csv is updated to nginx69986.v1.0.3")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "nginx69986.v1.0.3", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") != 0 {
				e2e.Logf("csv nginx69986.v1.0.3 status is not Succeeded, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv nginx69986.v1.0.3 is not Succeeded")

		g.By("5.2 check sub status")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			conditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].type}")
			if !strings.Contains(conditions, "BundleDeprecated") || !strings.Contains(conditions, "Deprecated") {
				return false, nil
			}
			messages := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}")
			if !strings.Contains(messages, "has been deprecated") {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status.conditions}")
		}
		exutil.AssertWaitPollNoErr(err, "the conditions of sub is not correct")

		g.By("6) update sub to channel candidate-v1.1")
		sub.Patch(oc, `{"spec": {"channel": "candidate-v1.1"}}`)
		g.By("6.1 check csv is updated to nginx69986.v1.1.1")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "nginx69986.v1.1.1", "-n", sub.Namespace, "-o=jsonpath={.status.phase}")
			if strings.Compare(status, "Succeeded") != 0 {
				e2e.Logf("csv nginx69986.v1.1.1 status is not Succeeded, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status}")
		}
		exutil.AssertWaitPollNoErr(err, "csv nginx69986.v1.1.1 is not Succeeded")

		g.By("6.2 check sub status")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			conditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].type}")
			if strings.Contains(conditions, "BundleDeprecated") || strings.Contains(conditions, "Deprecated") {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status.conditions}")
		}
		exutil.AssertWaitPollNoErr(err, "the conditions of sub is not correct, still has BundleDeprecated or Deprecated")
	})

	g.It("PolarionID:70050-[OTP][Skipped:Disconnected]OLM emits alert events for operators installed from a deprecated channel if catalog in different ns [Serial]", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-70050",
				Namespace:   "openshift-marketplace",
				DisplayName: "Test 70050",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/olmtest-operator-index:nginx70050",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-70050",
				Namespace:              namespaceName,
				CatalogSourceName:      "",
				CatalogSourceNamespace: "",
				Channel:                "candidate-v1.0",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx70050",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)
		sub.CatalogSourceNamespace = catsrc.Namespace
		sub.CatalogSourceName = catsrc.Name
		itName := g.CurrentSpecReport().FullText()

		g.By(fmt.Sprintf("1) create the catsrc in project: %s", catsrc.Namespace))
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("2) packagemanifests")
		message := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifests", "nginx70050", "-n", catsrc.Namespace, `-o=jsonpath='{.status.deprecation}'`)
		o.Expect(string(message)).To(o.ContainSubstring(`has been deprecated`))
		message = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifests", "nginx70050", "-n", catsrc.Namespace, `-o=jsonpath='{.status.channels[?(@.name=="candidate-v1.0")].deprecation}'`)
		o.Expect(string(message)).To(o.ContainSubstring(`has been deprecated`))
		message = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifests", "nginx70050", "-n", catsrc.Namespace, `-o=jsonpath={.status.channels[?(@.name=="candidate-v1.0")].entries[?(@.name=="nginx70050.v1.0.1")].deprecation}`)
		o.Expect(string(message)).To(o.ContainSubstring(`has been deprecated`))

		g.By("3) install og")
		og.CreateWithCheck(oc, itName, dr)

		g.By("4) install sub with channel candidate-v1.0")
		sub.Create(oc, itName, dr)

		g.By("4.1 check csv")
		var status string
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			status, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "nginx70050.v1.0.1", "-n", sub.Namespace, "-o=jsonpath={.status.phase}").Output()
			if strings.Compare(status, "Succeeded") != 0 {
				e2e.Logf("csv nginx70050.v1.0.1 status is not Succeeded, go next round")
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status}")
			if strings.Contains(status, "Unable to connect to the server: proxyconnect tcp") {
				exutil.AssertWaitPollNoErr(err, status)
			}
		}
		exutil.AssertWaitPollNoErr(err, "csv nginx70050.v1.0.1 is not Succeeded")

		g.By("4.2 check sub status")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			conditions := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].type}")
			if !strings.Contains(conditions, "Deprecated") || !strings.Contains(conditions, "ChannelDeprecated") || !strings.Contains(conditions, "PackageDeprecated") || !strings.Contains(conditions, "BundleDeprecated") {
				return false, nil
			}
			messages := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions[*].message}")
			if !strings.Contains(messages, "has been deprecated") {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", namespaceName, "-o=jsonpath-as-json={.status.conditions}")
		}
		exutil.AssertWaitPollNoErr(err, "the conditions of sub is not correct")
	})

	g.It("PolarionID:81995-[OTP][Skipped:Disconnected]Verify cacheless catalogsources is supported", g.Label("NonHyperShiftHOST"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-cacheless.yaml")

		oc.SetupProject()
		namespaceName := oc.Namespace()
		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-og",
				Namespace: namespaceName,
				Template:  ogSingleTemplate,
			}

			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-81995",
				Namespace:   namespaceName,
				DisplayName: "Test-Catsrc-81995-Operators",
				Publisher:   "Red-Hat",
				SourceType:  "grpc",
				Address:     "quay.io/openshifttest/nginxolm-operator-index:nginxolm81995-binless",
				Template:    catsrcImageTemplate,
			}

			catsrc2 = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-81995-2",
				Namespace:   namespaceName,
				DisplayName: "Test-Catsrc-81995-Operators-2",
				Publisher:   "Red-Hat",
				SourceType:  "grpc",
				Address:     "quay.io/openshifttest/nginxolm-operator-index:nginxolm81995-bin",
				Template:    catsrcImageTemplate,
			}

			sub = olmv0util.SubscriptionDescription{
				SubName:                "sub-81995",
				Namespace:              namespaceName,
				CatalogSourceName:      "catsrc-81995",
				CatalogSourceNamespace: namespaceName,
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx81995",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
		)

		itName := g.CurrentSpecReport().FullText()
		g.By("STEP1: create the OperatorGroup ")
		og.CreateWithCheck(oc, itName, dr)

		g.By("STEP 2: Create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("STEP 3: install operator")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("STEP 4: create catsrc2")
		defer catsrc2.Delete(itName, dr)
		catsrc2.CreateWithCheck(oc, itName, dr)
	})

	g.It("PolarionID:71779-[OTP][Skipped:Disconnected]Failing unpack jobs can be auto retried [Slow]", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogtemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		oc.SetupProject()
		namespace := oc.Namespace()
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-71779",
			Namespace: namespace,
			Template:  ogtemplate,
		}
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-71779",
			Namespace:   namespace,
			DisplayName: "Test Catsrc 71779 Operators",
			Publisher:   "Red Hat",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/bug29194-index:v1",
			Template:    catsrcImageTemplate,
		}

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-71779",
			Namespace:              namespace,
			IpApproval:             "Automatic",
			OperatorPackage:        "bug29194",
			CatalogSourceName:      catsrc.Name,
			CatalogSourceNamespace: namespace,
			Template:               subTemplate,
			SingleNamespace:        true,
		}

		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)
		err := oc.AsAdmin().WithoutNamespace().Run("patch").Args("og", og.Name, "-n", namespace, "--type=merge", "-p", `{"metadata":{"annotations":{"operatorframework.io/bundle-unpack-timeout":"10s"}}}`).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("3) Create a Subscription")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("5) Check sub message")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "BundleUnpackFailed", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.conditions}"}).Check(oc)
		jobs1 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "job", "-n", sub.Namespace, "--selector=operatorframework.io/bundle-unpack-ref", "-o=jsonpath={.items[*].metadata.name}")

		g.By("6) Patch OperatorGroup")
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("og", og.Name, "-n", namespace, "--type=merge", "-p", `{"metadata":{"annotations":{"operatorframework.io/bundle-unpack-min-retry-interval":"1s"}}}`).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("7) check unpack job is auto retried")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "job", "-n", sub.Namespace, "--selector=operatorframework.io/bundle-unpack-ref", "-o=jsonpath={.items[*].metadata.name}")
			jobs2 := strings.Split(output, "")
			for _, jobname := range jobs2 {
				if !strings.Contains(jobs1, jobname) {
					return true, nil
				}
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "unpack job is not auto retried")

		g.By("8) check unpack job is auto retried again")
		jobs2 := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "job", "-n", sub.Namespace, "--selector=operatorframework.io/bundle-unpack-ref", "-o=jsonpath={.items[*].metadata.name}")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
			output := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "job", "-n", sub.Namespace, "--selector=operatorframework.io/bundle-unpack-ref", "-o=jsonpath={.items[*].metadata.name}")
			jobs3 := strings.Split(output, "")
			for _, jobname := range jobs3 {
				if !strings.Contains(jobs2, jobname) {
					return true, nil
				}
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "unpack job is not auto retried")

		g.By("SUCCESS")

	})

	g.It("PolarionID:40972-[OTP][Skipped:Disconnected]Provide more specific text when no candidates for Subscription spec", func() {
		exutil.SkipMissingQECatalogsource(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subFile             = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			err                 error
			exists              bool
			failures            = 0
			failureNames        = ""
			msg                 string
			s                   string
			snooze              time.Duration = 300
			step                string
			waitErr             error
		)

		oc.SetupProject()

		var (
			og = olmv0util.OperatorGroupDescription{
				Name:      "test-40972-group",
				Namespace: oc.Namespace(),
				Template:  ogTemplate,
			}
			subOriginal = olmv0util.SubscriptionDescription{
				SubName:                "learn-40972",
				Namespace:              oc.Namespace(),
				CatalogSourceName:      "qe-app-registry",
				CatalogSourceNamespace: "openshift-marketplace",
				IpApproval:             "Automatic",
				Channel:                "beta",
				OperatorPackage:        "learn",
				SingleNamespace:        true,
				Template:               subFile,
			}
			sub = subOriginal
		)

		g.By("1, check if this operator exists")
		e2e.Logf("Check if %v exists in the %v catalog", sub.OperatorPackage, sub.CatalogSourceName)
		exists, err = olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			e2e.Failf("FAIL:PackageMissing %v does not exist in catalog %v", sub.OperatorPackage, sub.CatalogSourceName)
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(exists).To(o.BeTrue())

		g.By("2, Create og")
		og.Create(oc, itName, dr)

		g.By("1/3 bad package name")
		sub = subOriginal
		sub.OperatorPackage = "xyzzy"
		s = fmt.Sprintf("no operators found in package %v in the catalog referenced by subscription %v", sub.OperatorPackage, sub.SubName)
		step = "1/3"

		sub.CreateWithoutCheck(oc, itName, dr)
		waitErr = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, snooze*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err = oc.AsAdmin().Run("get").Args("sub", sub.SubName, "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[*].message}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, s) {
				return true, nil
			}
			return false, nil
		})
		if !strings.Contains(msg, s) {
			e2e.Logf("STEP after %v, %v FAIL log is missing %v\nSTEP in: %v\n", waitErr, step, s, msg)
			failures++
			failureNames = s + "\n"
		}
		sub.DeleteCSV(itName, dr)
		sub.Delete(itName, dr)

		g.By("2/3 bad catalog name")
		sub = subOriginal
		sub.CatalogSourceName = "xyzzy"
		s = fmt.Sprintf("no operators found from catalog %v in namespace openshift-marketplace referenced by subscription %v", sub.CatalogSourceName, sub.SubName)
		step = "2/3"

		sub.CreateWithoutCheck(oc, itName, dr)
		waitErr = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, snooze*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err = oc.AsAdmin().Run("get").Args("sub", sub.SubName, "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[*].message}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, s) {
				return true, nil
			}
			return false, nil
		})
		if !strings.Contains(msg, s) {
			e2e.Logf("STEP after %v, %v FAIL log is missing %v\nSTEP in: %v\n", waitErr, step, s, msg)
			failures++
			failureNames = failureNames + s + "\n"
		}
		sub.DeleteCSV(itName, dr)
		sub.Delete(itName, dr)

		g.By("3/3 bad channel")
		sub = subOriginal
		sub.Channel = "xyzzy"
		s = fmt.Sprintf("no operators found in channel %v of package %v in the catalog referenced by subscription %v", sub.Channel, sub.OperatorPackage, sub.SubName)
		step = "3/3"

		sub.CreateWithoutCheck(oc, itName, dr)
		waitErr = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, snooze*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err = oc.AsAdmin().Run("get").Args("sub", sub.SubName, "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[*].message}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, s) {
				return true, nil
			}
			return false, nil
		})
		if !strings.Contains(msg, s) {
			e2e.Logf("STEP after %v, %v FAIL log is missing %v\nSTEP in: %v\n", waitErr, step, s, msg)
			failures++
			failureNames = failureNames + s + "\n"
		}
		sub.DeleteCSV(itName, dr)
		sub.Delete(itName, dr)

		g.By("4/4 bad CSV")
		sub = subOriginal
		sub.StartingCSV = "xyzzy.v0.9.2"
		s = fmt.Sprintf("no operators found with name %v in channel beta of package %v in the catalog referenced by subscription %v", sub.StartingCSV, sub.OperatorPackage, sub.SubName)
		step = "4/4"

		sub.CreateWithoutCheck(oc, itName, dr)
		waitErr = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, snooze*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err = oc.AsAdmin().Run("get").Args("sub", sub.SubName, "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[*].message}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, s) {
				return true, nil
			}
			return false, nil
		})
		if !strings.Contains(msg, s) {
			e2e.Logf("STEP after %v, %v FAIL log is missing %v\nSTEP in: %v\n", waitErr, step, s, msg)
			failures++
			failureNames = failureNames + s + "\n"
		}
		sub.DeleteCSV(itName, dr)
		sub.Delete(itName, dr)

		g.By("FINISH\n")
		if failures != 0 {
			e2e.Failf("FAILED: %v times for %v", failures, failureNames)
		}
	})

})
