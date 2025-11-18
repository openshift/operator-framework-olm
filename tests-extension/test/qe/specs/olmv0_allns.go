package specs

import (
	"context"
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

	g.It("PolarionID:21418-PolarionID:25679-[OTP]Cluster resource created and deleted correctly [Serial]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within all namespace PolarionID:21418-PolarionID:25679-[Skipped:Disconnected]Cluster resource created and deleted correctly [Serial]"), func() {
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
		olmv0util.ValidateAccessEnvironment(oc)
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

	g.It("PolarionID:25783-[OTP][Skipped:Disconnected]Subscriptions are not getting processed taking very long to get processed[Serial]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within all namespace PolarionID:25783-[Skipped:Disconnected]Subscriptions are not getting processed taking very long to get processed[Serial]"), func() {
		architecture.SkipArchitectures(oc, architecture.PPC64LE, architecture.S390X, architecture.MULTI)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipNoCapabilities(oc, "marketplace")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
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
			ogAllTemplate       = filepath.Join(buildPruningBaseDir, "og-allns.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			ogAll               = olmv0util.OperatorGroupDescription{
				Name:      "og-all",
				Namespace: "",
				Template:  ogAllTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-25783-operator",
				Namespace:   "",
				DisplayName: "Test Catsrc 25783 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv25783",
				Template:    catsrcImageTemplate,
			}
			subCockroachdb = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v25783",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v25783",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}

			csvCockroachdb = olmv0util.CsvDescription{
				Name:      "",
				Namespace: "",
			}
		)

		exutil.SkipForSNOCluster(oc)
		platform := exutil.CheckPlatform(oc)
		proxy, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}

		g.By("Create og")
		ns := oc.Namespace()
		ogAll.Namespace = ns
		ogAll.Create(oc, itName, dr)

		g.By("create catsrc")
		catsrc.Namespace = ns
		catsrc.Create(oc, itName, dr)
		defer catsrc.Delete(itName, dr)

		g.By("create operator nginx-ok")
		subCockroachdb.CatalogSourceNamespace = catsrc.Namespace
		subCockroachdb.Namespace = catsrc.Namespace
		defer subCockroachdb.Delete(itName, dr)
		subCockroachdb.Create(oc, itName, dr)
		csvCockroachdb.Name = subCockroachdb.InstalledCSV
		csvCockroachdb.Namespace = subCockroachdb.Namespace
		defer csvCockroachdb.Delete(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subCockroachdb.InstalledCSV, "-n", subCockroachdb.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:21484-PolarionID:21532-[OTP]watch special or all namespace by operator group", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within all namespace PolarionID:21484-PolarionID:21532-[Skipped:Disconnected]watch special or all namespace by operator group"), func() {
		architecture.SkipArchitectures(oc, architecture.PPC64LE, architecture.S390X, architecture.MULTI)
		exutil.SkipNoCapabilities(oc, "marketplace")
		olmv0util.ValidateAccessEnvironment(oc)
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogAllTemplate       = filepath.Join(buildPruningBaseDir, "og-allns.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			ogAll               = olmv0util.OperatorGroupDescription{
				Name:      "og-all",
				Namespace: "",
				Template:  ogAllTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "olm-21532-catalog",
				Namespace:   "",
				DisplayName: "OLM 21532 Catalog",
				Publisher:   "QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv21532",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v21532",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v21532",
				CatalogSourceName:      "olm-21532-catalog",
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}

			project = olmv0util.ProjectDescription{
				Name:            "olm-enduser-specific-21484",
				TargetNamespace: oc.Namespace(),
			}
			cl = olmv0util.CheckList{}
		)

		// OCP-21532
		g.By("Check the global operator global-operators support all namesapces")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "[]", exutil.Ok, []string{"og", "global-operators", "-n", "openshift-operators", "-o=jsonpath={.status.namespaces}"}))

		g.By("Create og")
		ns := oc.Namespace()
		ogAll.Namespace = ns
		ogAll.Create(oc, itName, dr)

		g.By("create catsrc")
		catsrc.Namespace = ns
		catsrc.CreateWithCheck(oc, itName, dr)
		defer catsrc.Delete(itName, dr)

		// OCP-21484, OCP-21532
		g.By("Create operator targeted at all namespace")
		sub.Namespace = ns
		sub.CatalogSourceNamespace = ns
		sub.Create(oc, itName, dr)

		g.By("Create new namespace")
		project.Create(oc, itName, dr)

		// OCP-21532
		g.By("New annotations is added to copied CSV in current namespace")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Contain, "alm-examples", exutil.Ok, []string{"csv", sub.InstalledCSV, "-o=jsonpath={.metadata.annotations}"}))

		// OCP-21484, OCP-21532
		g.By("Check the csv within new namespace is copied. note: the step is slow because it wait to copy csv to new namespace")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Copied", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", project.Name, "-o=jsonpath={.status.reason}"}))

		cl.Check(oc)

	})

	g.It("PolarionID:24906-[OTP]Operators requesting cluster-scoped permission can trigger kube GC bug[Serial]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within all namespace PolarionID:24906-[Skipped:Disconnected]Operators requesting cluster-scoped permission can trigger kube GC bug[Serial]"), func() {
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
		olmv0util.ValidateAccessEnvironment(oc)
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
				Address:     "quay.io/olmqe/nginx-ok-index:vokv24906",
				Template:    catsrcImageTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v24906",
				Namespace:              "",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v24906",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: "",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
			cl = olmv0util.CheckList{}
		)

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
		sub.Update(oc, itName, dr)

		g.By("Check clusterrolebinding has no OwnerReferences")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "", exutil.Ok, []string{"clusterrolebinding", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..OwnerReferences}"}))

		g.By("Check clusterrole has no OwnerReferences")
		cl.Add(olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "", exutil.Ok, []string{"clusterrole", fmt.Sprintf("--selector=olm.owner=%s", sub.InstalledCSV), "-n", sub.Namespace, "-o=jsonpath={..OwnerReferences}"}))
		//do check parallelly
		cl.Check(oc)
	})

	g.It("PolarionID:33241-[OTP][Skipped:Disconnected]Enable generated operator component adoption for operators with all ns mode[Serial]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within all namespace PolarionID:33241-[Skipped:Disconnected]Enable generated operator component adoption for operators with all ns mode[Serial]"), func() {
		if isAKS, _ := exutil.IsAKSCluster(context.TODO(), oc); isAKS {
			g.Skip("skip for aks cluster")
		}
		architecture.SkipArchitectures(oc, architecture.PPC64LE, architecture.S390X, architecture.MULTI)
		exutil.SkipNoCapabilities(oc, "marketplace")
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
		proxy, errProxy := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status.httpProxy}{.status.httpsProxy}").Output()
		o.Expect(errProxy).NotTo(o.HaveOccurred())
		if proxy != "" || strings.Contains(platform, "openstack") || strings.Contains(platform, "baremetal") || strings.Contains(platform, "vsphere") || exutil.Is3MasterNoDedicatedWorkerNode(oc) ||
			os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" {
			g.Skip("it is not supported")
		}
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			catsrc              = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-33241-operator",
				Namespace:   "openshift-marketplace",
				DisplayName: "Test Catsrc 33241 Operators",
				Publisher:   "Red Hat",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:vokv33241",
				Template:    catsrcImageTemplate,
			}
			subCockroachdb = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok-v33241",
				Namespace:              "openshift-operators",
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok-v33241",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: catsrc.Namespace,
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
			}
		)

		g.By("check if cockroachdb is already installed with all ns.")
		csvList := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", subCockroachdb.Namespace, "-o=jsonpath={.items[*].metadata.name}")
		if !strings.Contains(csvList, subCockroachdb.OperatorPackage) {
			g.By("create catsrc")
			catsrc.CreateWithCheck(oc, itName, dr)
			defer catsrc.Delete(itName, dr)

			g.By("Create operator targeted at all namespace")
			subCockroachdb.Create(oc, itName, dr)
			csvCockroachdb := olmv0util.CsvDescription{
				Name:      subCockroachdb.InstalledCSV,
				Namespace: subCockroachdb.Namespace,
			}
			defer subCockroachdb.Delete(itName, dr)
			defer csvCockroachdb.Delete(itName, dr)
			crdName := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subCockroachdb.OperatorPackage+"."+subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='CustomResourceDefinition')].name}")
			o.Expect(crdName).NotTo(o.BeEmpty())
			defer func() {
				_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "crd", crdName)
			}()
			defer func() {
				_, _ = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subCockroachdb.OperatorPackage+"."+subCockroachdb.Namespace)
			}()

			g.By("Check all resources via operators")
			resourceKind := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operator.operators.coreos.com", subCockroachdb.OperatorPackage+"."+subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[*].kind}")
			o.Expect(resourceKind).To(o.ContainSubstring("Deployment"))
			o.Expect(resourceKind).To(o.ContainSubstring("Role"))
			o.Expect(resourceKind).To(o.ContainSubstring("RoleBinding"))
			o.Expect(resourceKind).To(o.ContainSubstring("ClusterRole"))
			o.Expect(resourceKind).To(o.ContainSubstring("ClusterRoleBinding"))
			o.Expect(resourceKind).To(o.ContainSubstring("CustomResourceDefinition"))
			o.Expect(resourceKind).To(o.ContainSubstring("Subscription"))
			o.Expect(resourceKind).To(o.ContainSubstring("InstallPlan"))
			o.Expect(resourceKind).To(o.ContainSubstring("ClusterServiceVersion"))
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, subCockroachdb.Namespace, exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].namespace}"}).Check(oc)
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "InstallSucceeded", exutil.Ok, []string{"operator.operators.coreos.com", subCockroachdb.OperatorPackage + "." + subCockroachdb.Namespace, "-o=jsonpath={.status.components.refs[?(.kind=='ClusterServiceVersion')].conditions[*].reason}"}).Check(oc)
		}
	})

	g.It("PolarionID:22226-[OTP][Skipped:Disconnected]the csv without support AllNamespaces fails for og with allnamespace", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 within all namespace PolarionID:22226-[Skipped:Disconnected]the csv without support AllNamespaces fails for og with allnamespace"), func() {
		var (
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			cmNcTemplate        = filepath.Join(buildPruningBaseDir, "cm-namespaceconfig.yaml")
			catsrcCmTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
			ogAllTemplate       = filepath.Join(buildPruningBaseDir, "og-allns.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			itName              = g.CurrentSpecReport().FullText()
			og                  = olmv0util.OperatorGroupDescription{
				Name:      "og-allnamespace",
				Namespace: "",
				Template:  ogAllTemplate,
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
		)

		cm.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()
		g.By("Create cm")
		cm.Create(oc, itName, dr)

		g.By("Create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create sub")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "AllNamespaces InstallModeType not supported", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.message}"}).Check(oc)
	})

})
