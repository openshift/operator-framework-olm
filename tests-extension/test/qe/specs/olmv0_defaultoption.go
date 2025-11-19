package specs

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/architecture"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// it is mapping to the Describe "OLM should" and "OLM optional" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 optional should", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLI("default-"+exutil.GetRandomString(), exutil.KubeConfigPath())
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		exutil.SkipNoOLMCore(oc)
	})

	g.It("PolarionID:54038-[OTP][Skipped:Disconnected]Comply with Operator Anti-Affinity definition", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-54038-Comply with Operator Anti-Affinity definition"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		exutil.SkipBaselineCaps(oc, "None")

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		prometheusCR := filepath.Join(buildPruningBaseDir, "prometheus-antiaffinity.yaml")

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-54038",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-54038",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "prometheus",
			SingleNamespace:        true,
			Template:               subTemplate,
		}

		workerNodes, _ := exutil.GetSchedulableLinuxWorkerNodes(oc)
		if len(workerNodes) == 0 {
			g.Skip("No schedulable Linux worker nodes found")
		}
		firstNode := workerNodes[0]

		exists, _ := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing prometheus does not exist in catalog community-operators")
		}

		if olmv0util.IsSNOCluster(oc) {
			g.Skip("SNO cluster - skipping test ...")
		}

		if len(strings.TrimSpace(firstNode.Name)) == 0 {
			g.Skip("Skipping because there's no cluster with READY state")
		}

		g.By("1) Install the OperatorGroup in a random project")
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Install the Prometheus with Automatic approval")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) Add app label")
		defer func() {
			if _, delErr := exutil.DeleteLabelFromNode(oc, firstNode.Name, "app_54038"); delErr != nil {
				e2e.Logf("failed to remove label from node %s: %v", firstNode.Name, delErr)
			}
		}()
		_, addErr := exutil.AddLabelToNode(oc, firstNode.Name, "app_54038", "dev")
		o.Expect(addErr).NotTo(o.HaveOccurred())

		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("nodes", "--show-labels", "--no-headers").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("Node labels %s", msg)

		g.By("4) Install the Prometheus CR")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", prometheusCR, "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Available", exutil.Ok,
			[]string{"Prometheus", "example", "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[0].type}"}).Check(oc)

		g.By("5) Ensure that pod is not scheduled in the node with the defined label")
		deployedNode := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace,
			"pods", "prometheus-example-0", "-n", oc.Namespace(), "-o=jsonpath={.spec.nodeName}")
		if firstNode.Name == deployedNode {
			e2e.Failf("Prometheus is deployed in the same node of app_54038 label. Node: %s . Node Labels: %s", deployedNode, msg)
		}
	})

	g.It("PolarionID:54036-[OTP][Skipped:Disconnected]Comply with Operator NodeAffinity definition", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-54036-Comply with Operator NodeAffinity definition"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		exutil.SkipBaselineCaps(oc, "None")

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		prometheusCRTemplate := filepath.Join(buildPruningBaseDir, "prometheus-nodeaffinity.yaml")

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-54036",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-54036",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "prometheus",
			SingleNamespace:        true,
			Template:               subTemplate,
		}

		workerNodes, _ := exutil.GetSchedulableLinuxWorkerNodes(oc)
		firstNode := ""
		for _, worker := range workerNodes {
			for _, condition := range worker.Status.Conditions {
				_, edge := worker.Labels["node-role.kubernetes.io/edge"]
				if condition.Type == "Ready" && condition.Status == "True" && !edge {
					firstNode = worker.Name
				}
			}
		}
		if olmv0util.IsSNOCluster(oc) || firstNode == "" {
			g.Skip("SNO cluster - skipping test ...")
		}

		if len(strings.TrimSpace(firstNode)) == 0 {
			g.Skip("Skipping because there's no cluster with READY state")
		}

		g.By("1) Install the OperatorGroup in a random project")
		og.CreateWithCheck(oc, itName, dr)

		exists, _ := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing prometheus does not exist in catalog community-operators")
		}

		g.By("2) Install the Prometheus with Automatic approval")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) Install the Prometheus CR")
		err := olmv0util.ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", prometheusCRTemplate,
			"-p", "NODE_NAME="+firstNode, "NAMESPACE="+oc.Namespace())
		o.Expect(err).NotTo(o.HaveOccurred())

		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Available", exutil.Ok,
			[]string{"Prometheus", "example", "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[0].type}"}).Check(oc)

		g.By("4) Ensure that pod is scaled in the specified node")
		deployedNode := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace,
			"pods", "prometheus-example-0", "-n", oc.Namespace(), "-o=jsonpath={.spec.nodeName}")
		o.Expect(firstNode).To(o.Equal(deployedNode))
	})

	g.It("PolarionID:24850-[OTP]Allow users to edit the deployment of an active CSV", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-24850-Allow users to edit the deployment of an active CSV"), func() {
		exutil.SkipMissingQECatalogsource(oc)
		g.By("1) Install the OperatorGroup in a random project")

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		oc.SetupProject()
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-24850",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Install the learn operator with Automatic approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-24850",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "qe-app-registry",
			CatalogSourceNamespace: "openshift-marketplace",
			IpApproval:             "Automatic",
			Channel:                "beta",
			OperatorPackage:        "learn",
			SingleNamespace:        true,
			Template:               subTemplate,
		}

		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) Get pod name")
		podName, err := oc.AsAdmin().Run("get").Args("pods", "-l", "name=learn-operator", "-n", oc.Namespace(), "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("4) Patch the deploy object by adding an environment variable")
		_, err = oc.AsAdmin().WithoutNamespace().Run("set").Args("env", "deploy/learn-operator", "A=B", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("5) Get restarted pod name")
		podNameAfterPatch, err := oc.AsAdmin().Run("get").Args("pods", "-l", "name=learn-operator", "-n", oc.Namespace(), "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(podName).NotTo(o.Equal(podNameAfterPatch))
	})

	g.It("PolarionID:24387-[OTP][Skipped:Disconnected][Disruptive]Any CRD upgrade is allowed if there is only one owner in a cluster", g.Label("original-name:[sig-operators] OLM should Author:bandrade-ConnectedOnly-High-24387-Any CRD upgrade is allowed if there is only one owner in a cluster [Disruptive]"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		catName := "cs-24387"
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")

		oc.SetupProject()
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		cs := olmv0util.CatalogSourceDescription{
			Name:        catName,
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE Operators",
			Publisher:   "bandrade",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-index-24387:5.0",
			Template:    csTemplate,
		}

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-24387",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}

		sub := olmv0util.SubscriptionDescription{
			SubName:                "etcd",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd",
			SingleNamespace:        true,
			Template:               subTemplate,
			StartingCSV:            "etcdoperator.v0.9.4",
		}

		subModified := olmv0util.SubscriptionDescription{
			SubName:                "etcd",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      catName,
			CatalogSourceNamespace: "openshift-marketplace",
			IpApproval:             "Automatic",
			Template:               subTemplate,
			Channel:                "singlenamespace-alpha",
			OperatorPackage:        "etcd",
			StartingCSV:            "etcdoperator.v0.9.4",
			SingleNamespace:        true,
		}

		g.By("1) Check if this operator ready for installing")
		e2e.Logf("Check if %v exists in the %v catalog", sub.OperatorPackage, sub.CatalogSourceName)
		exists, err := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing etcd does not exist in catalog community-operators")
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) Create catalog source")
		defer cs.Delete(itName, dr)
		cs.Create(oc, itName, dr)

		g.By("3) Create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)

		g.By("4) Start to subscribe to the Etcd operator")
		sub.Create(oc, itName, dr)

		g.By("5) Delete Etcd subscription and csv")
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)

		g.By("6) Start to subscribe to the Etcd operator with the modified CRD")
		subModified.Create(oc, itName, dr)

		g.By("7) Get property propertyIncludedTest in etcdclusters.etcd.database.coreos.com")
		crdYamlOutput, err := oc.AsAdmin().Run("get").Args("crd", "etcdclusters.etcd.database.coreos.com", "-o=yaml").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(crdYamlOutput).To(o.ContainSubstring("propertyIncludedTest"))
	})

	g.It("PolarionID:42970-[OTP]OperatorGroup status indicates cardinality conflicts - SingleNamespace", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-42970-OperatorGroup status indicates cardinality conflicts - SingleNamespace"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")

		oc.SetupProject()
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		ns := oc.Namespace()
		dr.AddIr(itName)

		og := olmv0util.OperatorGroupDescription{
			Name:      "og-42970",
			Namespace: ns,
			Template:  ogTemplate,
		}
		og1 := olmv0util.OperatorGroupDescription{
			Name:      "og-42970-1",
			Namespace: ns,
			Template:  ogTemplate,
		}

		g.By("1) Create first OperatorGroup")
		og.Create(oc, itName, dr)

		g.By("2) Create second OperatorGroup")
		og1.Create(oc, itName, dr)

		g.By("3) Check OperatorGroup Status")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "MultipleOperatorGroupsFound", exutil.Ok,
			[]string{"og", og.Name, "-n", ns, "-o=jsonpath={.status.conditions..reason}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "MultipleOperatorGroupsFound", exutil.Ok,
			[]string{"og", og1.Name, "-n", ns, "-o=jsonpath={.status.conditions..reason}"}).Check(oc)

		g.By("4) Delete second OperatorGroup")
		og1.Delete(itName, dr)

		g.By("5) Check OperatorGroup status")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("og", og.Name, "-n", ns, "-o=jsonpath={.status.conditions..reason}").Output()
			if err != nil {
				e2e.Logf("Fail to get og: %s, error: %s and try again", og.Name, err)
				return false, nil
			}
			if strings.TrimSpace(output) == "" {
				return true, nil
			}
			e2e.Logf("The error MultipleOperatorGroupsFound still be reported in status, try again")
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "The error MultipleOperatorGroupsFound still be reported in status")
		g.By("6) OCP-42970 SUCCESS")
	})

	g.It("PolarionID:42972-[OTP]OperatorGroup status should indicate if the SA named in spec not found", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-42972-OperatorGroup status should indicate if the SA named in spec not found"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
		serviceAccount := "scoped-42972"

		oc.SetupProject()
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		ns := oc.Namespace()
		dr.AddIr(itName)

		og := olmv0util.OperatorGroupDescription{
			Name:               "og-42972",
			Namespace:          ns,
			Template:           ogTemplate,
			ServiceAccountName: serviceAccount,
		}

		g.By("1) Create OperatorGroup")
		og.Create(oc, itName, dr)

		g.By("2) Check OperatorGroup Status")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "ServiceAccountNotFound", exutil.Ok,
			[]string{"og", og.Name, "-n", ns, "-o=jsonpath={.status.conditions..reason}"}).Check(oc)

		g.By("3) Create Service Account")
		_, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("sa", serviceAccount, "-n", ns).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("4) Check OperatorGroup status")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("og", og.Name, "-n", ns, "-o=jsonpath={.status.conditions..reason}").Output()
			if err != nil {
				e2e.Logf("Fail to get og: %s, error: %s and try again", og.Name, err)
				return false, nil
			}
			if strings.TrimSpace(output) == "" {
				return true, nil
			}
			e2e.Logf("The error ServiceAccountNotFound still be reported in status, try again")
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "The error ServiceAccountNotFound still be reported in status")
	})

	g.It("PolarionID:24771-[OTP]OLM should support for user defined ServiceAccount for OperatorGroup", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-24771-OLM should support for user defined ServiceAccount for OperatorGroup"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		saRoles := filepath.Join(buildPruningBaseDir, "scoped-sa-roles.yaml")
		oc.SetupProject()
		namespace := oc.Namespace()
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		targetCSV := "learn-operator.v0.0.3"
		serviceAccount := "scoped-24771"

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-24771",
			Namespace:              namespace,
			CatalogSourceName:      "qe-app-registry",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            targetCSV,
			SingleNamespace:        true,
			Template:               subTemplate,
		}

		og := olmv0util.OperatorGroupDescription{
			Name:               "test-og-24771",
			Namespace:          namespace,
			ServiceAccountName: serviceAccount,
			Template:           ogTemplate,
		}

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Check if this operator ready for installing")
		e2e.Logf("Check if %v exists in the %v catalog", sub.OperatorPackage, sub.CatalogSourceName)
		exists, err := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing learn does not exist in catalog qe-app-registry")
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) Create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) Create the service account")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("sa", serviceAccount, "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("4) Create a Subscription")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("5) The install plan is Failed")
		installPlan := sub.GetIP(oc)
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			message, _ := oc.AsAdmin().Run("get").Args("installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.message}").Output()
			if strings.Contains(message, "cannot create resource") {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			message, _ := oc.AsAdmin().Run("get").Args("installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.message}").Output()
			e2e.Logf(message)
			conditions, _ := oc.AsAdmin().Run("get").Args("installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath-as-json={.status.conditions}").Output()
			e2e.Logf(conditions)
		}
		exutil.AssertWaitPollNoErr(err, "cannot create resource not in install plan message")

		g.By("6) Grant the proper permissions to the service account")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", saRoles, "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("7) Recreate the Subscription")
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("8) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", targetCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:43073-[OTP][Skipped:Disconnected]Indicate dependency class in resolution constraint text", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-43073-Indicate dependency class in resolution constraint text"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		oc.SetupProject()

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		catName := "cs-43073"
		baseDir := exutil.FixturePath("testdata", "olm")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")
		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")

		cs := olmv0util.CatalogSourceDescription{
			Name:        catName,
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE Operators",
			Publisher:   "bandrade",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/bundle-with-dep-error-index:4.0",
			Template:    csTemplate,
		}

		og := olmv0util.OperatorGroupDescription{
			Name:      "og-43073",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}

		defer cs.Delete(itName, dr)
		g.By("1) Create the CatalogSource")
		cs.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok,
			[]string{"catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

		g.By("2) Install the OperatorGroup in a random project")
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) Install the lib-bucket-provisioner with Automatic approval")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "lib-bucket-provisioner-43073",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      catName,
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "lib-bucket-provisioner",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)

		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "ConstraintsNotSatisfiable", exutil.Ok,
			[]string{"subs", "lib-bucket-provisioner-43073", "-n", oc.Namespace(), "-o=jsonpath={.status.conditions[?(.type==\"ResolutionFailed\")].reason}"}).Check(oc)
	})

	g.It("PolarionID:24772-[OTP]OLM should support for user defined ServiceAccount for OperatorGroup with fine grained permission", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-24772-OLM should support for user defined ServiceAccount for OperatorGroup with fine grained permission"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		saRoles := filepath.Join(buildPruningBaseDir, "scoped-sa-fine-grained-roles.yaml")
		oc.SetupProject()
		namespace := oc.Namespace()
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		secTemplate := filepath.Join(buildPruningBaseDir, "secret.yaml")
		targetCSV := "learn-operator.v0.0.3"
		serviceAccount := "scoped-24772"

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-24772",
			Namespace:              namespace,
			CatalogSourceName:      "qe-app-registry",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            targetCSV,
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		secret := olmv0util.SecretDescription{
			Name:      serviceAccount,
			Namespace: namespace,
			Saname:    serviceAccount,
			Sectype:   "kubernetes.io/service-account-token",
			Template:  secTemplate,
		}
		project := olmv0util.ProjectDescription{
			Name: namespace,
		}
		og := olmv0util.OperatorGroupDescription{
			Name:               "test-og-24772",
			Namespace:          namespace,
			ServiceAccountName: serviceAccount,
			Template:           ogTemplate,
		}

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Check if this operator ready for installing")
		e2e.Logf("Check if %v exists in the %v catalog", sub.OperatorPackage, sub.CatalogSourceName)
		exists, err := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing learn does not exist in catalog qe-app-registry")
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) Create the namespace")
		project.CreateWithCheck(oc, itName, dr)

		g.By("3) Create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)

		g.By("4) Create the service account and secret")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("sa", serviceAccount, "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		secret.Create(oc)

		g.By("5) Create a Subscription")
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("6) The install plan is Failed")
		installPlan := sub.GetIP(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "forbidden", exutil.Ok,
			[]string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.message}"}).Check(oc)

		g.By("7) Grant the proper permissions to the service account")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", saRoles, "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("8) Recreate the Subscription")
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)

		g.By("9) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", targetCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:24886-[OTP]OLM should support for user defined ServiceAccount permission changes", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-24886-OLM should support for user defined ServiceAccount permission changes"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		saRoles := filepath.Join(buildPruningBaseDir, "scoped-sa-etcd.yaml")
		oc.SetupProject()
		namespace := oc.Namespace()
		ogTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		ogSATemplate := filepath.Join(buildPruningBaseDir, "operatorgroup-serviceaccount.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		secTemplate := filepath.Join(buildPruningBaseDir, "secret.yaml")
		targetCSV := "learn-operator.v0.0.3"
		serviceAccount := "scoped-24886"

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-24772",
			Namespace:              namespace,
			CatalogSourceName:      "qe-app-registry",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            targetCSV,
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		secret := olmv0util.SecretDescription{
			Name:      serviceAccount,
			Namespace: namespace,
			Saname:    serviceAccount,
			Sectype:   "kubernetes.io/service-account-token",
			Template:  secTemplate,
		}
		project := olmv0util.ProjectDescription{
			Name: namespace,
		}
		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-24886",
			Namespace: namespace,
			Template:  ogTemplate,
		}

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Check if this operator ready for installing")
		e2e.Logf("Check if %v exists in the %v catalog", sub.OperatorPackage, sub.CatalogSourceName)
		exists, err := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing learn does not exist in catalog qe-app-registry")
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) Create the namespace")
		project.CreateWithCheck(oc, itName, dr)

		g.By("3) Create the OperatorGroup without service account")
		og.CreateWithCheck(oc, itName, dr)

		g.By("4) Create a Subscription")
		sub.Create(oc, itName, dr)

		g.By("5) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", targetCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("6) Delete the Operator Group")
		og.Delete(itName, dr)

		ogSA := olmv0util.OperatorGroupDescription{
			Name:               "test-og-24886",
			Namespace:          namespace,
			ServiceAccountName: serviceAccount,
			Template:           ogSATemplate,
		}

		g.By("7) Create the OperatorGroup with service account")
		ogSA.CreateWithCheck(oc, itName, dr)

		g.By("8) Create the service account and secret")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("sa", serviceAccount, "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		secret.Create(oc)

		g.By("9) Grant the proper permissions to the service account")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", saRoles, "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("10) Recreate the Subscription")
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		sub.Create(oc, itName, dr)

		g.By("11) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", targetCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:30765-[OTP][Skipped:Disconnected]Operator-version based dependencies metadata", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-30765-Operator-version based dependencies metadata"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		baseDir := exutil.FixturePath("testdata", "olm")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")

		oc.SetupProject()
		g.By("1) Start to create the CatalogSource CR")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "prometheus-dependency-cs",
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-prometheus-dependency-index:11.0",
			Template:    csTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		defer cs.Delete(itName, dr)
		cs.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok,
			[]string{"catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

		g.By("2) Install the OperatorGroup in a random project")
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-30765",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) Install the etcdoperator v0.9.4 with Automatic approval")
		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-30765",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      cs.Name,
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd-service-monitor",
			StartingCSV:            "etcdoperator.v0.9.4",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", "etcdoperator.v0.9.4", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("4) Assert that prometheus dependency is resolved")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).To(o.ContainSubstring("prometheus"))
	})

	g.It("PolarionID:27680-[OTP][Skipped:Disconnected][Serial]OLM Bundle support for Prometheus Types", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-27680-OLM Bundle support for Prometheus Types [Serial]"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		baseDir := exutil.FixturePath("testdata", "olm")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")

		g.By("Start to create the CatalogSource CR")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "prometheus-dependency1-cs",
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-prometheus-dependency-index:11.0",
			Template:    csTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		defer cs.Delete(itName, dr)
		cs.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok,
			[]string{"catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

		g.By("Start to subscribe the Etcd operator")
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-27680",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-27680",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      cs.Name,
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd-service-monitor",
			StartingCSV:            "etcdoperator.v0.9.4",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", "etcdoperator.v0.9.4", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("Assert that prometheus dependency is resolved")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).To(o.ContainSubstring("prometheus"))

		g.By("Assert that ServiceMonitor is created")
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("ServiceMonitor", "my-servicemonitor", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).To(o.ContainSubstring("my-servicemonitor"))

		g.By("Assert that PrometheusRule is created")
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("PrometheusRule", "my-prometheusrule", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).To(o.ContainSubstring("my-prometheusrule"))
	})

	g.It("PolarionID:24916-[OTP][Skipped:Disconnected]Operators in AllNamespaces should be granted namespace list", g.Label("original-name:[sig-operators] OLM should Author:bandrade-ConnectedOnly-Medium-24916-Operators in AllNamespaces should be granted namespace list"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipMissingQECatalogsource(oc)

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("Start to subscribe the Learn operator")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "learn",
			Namespace:              "openshift-operators",
			CatalogSourceName:      "qe-app-registry",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "beta",
			IpApproval:             "Automatic",
			StartingCSV:            "learn-operator.v0.0.3",
			OperatorPackage:        "learn",
			SingleNamespace:        false,
			Template:               subTemplate,
		}
		exists, err := olmv0util.ClusterPackageExists(oc, sub)
		if !exists {
			g.Skip("SKIP:PackageMissing learn does not exist in catalog qe-app-registry")
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		defer sub.Delete(itName, dr)
		defer func() {
			if sub.InstalledCSV == "" {
				sub.FindInstalledCSV(oc, itName, dr)
			}
			sub.DeleteCSV(itName, dr)
		}()
		sub.Create(oc, itName, dr)

		g.By("check if learn is already installed")
		csvList := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithNamespace, "csv", "-o=jsonpath={.items[*].metadata.name}")
		e2e.Logf("CSV list %s", csvList)
		if !strings.Contains(csvList, "learn") {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("policy").Args("who-can", "list", "namespaces").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(msg).To(o.ContainSubstring("system:serviceaccount:openshift-operators:learn-operator"))
		} else {
			e2e.Failf("Not able to install Learn Operator")
		}
	})

	g.It("PolarionID:47149-[OTP][Skipped:Disconnected]Conjunctive constraint of one packages and one GVK", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-47149-Conjunctive constraint of one packages and one GVK"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		baseDir := exutil.FixturePath("testdata", "olm")
		oc.SetupProject()
		namespace := oc.Namespace()
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")

		g.By("Start to create the CatalogSource CR")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "ocp-47149",
			Namespace:   namespace,
			DisplayName: "ocp-47149",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-47149:1.0",
			Template:    csTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-47149",
			Namespace: namespace,
			Template:  ogTemplate,
		}

		g.By("1) Create the OperatorGroup without service account")
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Create a Subscription")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "etcd",
			Namespace:              namespace,
			CatalogSourceName:      cs.Name,
			CatalogSourceNamespace: namespace,
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)

		g.By("3) Checking the state of CSV")
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
			csvList, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", sub.Namespace).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			lines := strings.Split(csvList, "\n")
			for _, line := range lines {
				if strings.Contains(line, "prometheusoperator") {
					e2e.Logf("found csv prometheusoperator")
					if strings.Contains(line, "Succeeded") {
						e2e.Logf("the status csv prometheusoperator is Succeeded")
						return true, nil
					}
					e2e.Logf("the status csv prometheusoperator is not Succeeded")
					return false, nil
				}
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(waitErr, "csv prometheusoperator is not Succeeded")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", "etcdoperator.v0.9.4", "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:47181-[OTP][Skipped:Disconnected]Disjunctive constraint of one package and one GVK", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-47181-Disjunctive constraint of one package and one GVK"), func() {
		architecture.SkipNonAmd64SingleArch(oc)

		baseDir := exutil.FixturePath("testdata", "olm")
		oc.SetupProject()
		namespace := oc.Namespace()
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")

		g.By("Start to create the CatalogSource CR")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "ocp-47181",
			Namespace:   namespace,
			DisplayName: "ocp-47181",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-47181:1.0",
			Template:    csTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-47181",
			Namespace: namespace,
			Template:  ogTemplate,
		}

		g.By("1) Create the OperatorGroup without service account")
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Create a Subscription")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "etcd",
			Namespace:              namespace,
			CatalogSourceName:      cs.Name,
			CatalogSourceNamespace: namespace,
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)

		g.By("3) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	g.It("PolarionID:47179-[OTP][Skipped:Disconnected]Disjunctive constraint of one package and one GVK", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-47179-Disjunctive constraint of one package and one GVK"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		baseDir := exutil.FixturePath("testdata", "olm")
		oc.SetupProject()
		namespace := oc.Namespace()
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")

		g.By("Start to create the CatalogSource CR")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "ocp-47179",
			Namespace:   namespace,
			DisplayName: "ocp-47179",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-47179:1.0",
			Template:    csTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-47179",
			Namespace: namespace,
			Template:  ogTemplate,
		}

		g.By("1) Create the OperatorGroup without service account")
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Create a Subscription")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "etcd",
			Namespace:              namespace,
			CatalogSourceName:      cs.Name,
			CatalogSourceNamespace: namespace,
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)

		g.By("3) Checking the state of CSV")
		olmv0util.NewCheck("expect", exutil.AsUser, exutil.WithoutNamespace, exutil.Contain, "red-hat-camel-k-operator", exutil.Ok,
			[]string{"csv", "-n", sub.Namespace}).Check(oc)
	})

	g.It("PolarionID:49130-[OTP][Skipped:Disconnected]Default CatalogSources deployed by marketplace do not have toleration for tainted nodes", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operators] OLM should NonHyperShiftHOST-ConnectedOnly-Author:bandrade-Medium-49130-Default CatalogSources deployed by marketplace do not have toleration for tainted nodes"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		catalogSources := map[string]string{
			"certified-operators": "Certified Operators",
			"community-operators": "Community Operators",
			"redhat-operators":    "Red Hat Operators",
		}

		for catalog, label := range catalogSources {
			rawPods, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
				"pods", "-n", "openshift-marketplace",
				"-l", fmt.Sprintf("olm.catalogSource=%s", catalog),
				"-o", "name").Output()
			o.Expect(err).NotTo(o.HaveOccurred())

			pods := strings.Split(strings.TrimSpace(rawPods), "\n")
			if len(pods) == 0 || pods[0] == "" {
				e2e.Logf("No pods found for %s, skipping...", label)
				continue
			}

			found := false
			for _, name := range pods {
				output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(name, "-n", "openshift-marketplace", "-o=jsonpath={.spec.tolerations}").Output()
				if err != nil && apierrors.IsNotFound(err) {
					e2e.Logf("pod %v does not exist", name)
					continue
				}
				o.Expect(err).NotTo(o.HaveOccurred())
				if !strings.Contains(output, "node-role.kubernetes.io/master") || !strings.Contains(output, "tolerationSeconds\":120") {
					e2e.Logf("pod %v with incorrect tolerations found: %v", name, output)
					found = true
					break
				}
			}

			if found {
				e2e.Failf("Pod with incorrect tolerations found for %s", label)
			}
		}
	})

	g.It("PolarionID:21130-[OTP]Fetching non-existent PackageManifest should return 404", g.Label("original-name:[sig-operators] OLM should Author:bandrade-Medium-21130-Fetching non-existent `PackageManifest` should return 404"), func() {
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "--all-namespaces", "--no-headers").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		packageserverLines := strings.Split(msg, "\n")
		if len(packageserverLines) > 0 {
			raw, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "a_package_that_not_exists", "-o", "yaml", "--loglevel=8").Output()
			o.Expect(err).To(o.HaveOccurred())
			o.Expect(raw).To(o.ContainSubstring("\"code\": 404"))
		} else {
			e2e.Failf("No packages to evaluate if 404 works when a PackageManifest does not exist")
		}
	})

	g.It("PolarionID:24057-[OTP]Have terminationMessagePolicy defined as FallbackToLogsOnError", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operators] OLM should NonHyperShiftHOST-Author:bandrade-Low-24057-Have terminationMessagePolicy defined as FallbackToLogsOnError"), func() {
		labels := [...]string{"app=packageserver", "app=catalog-operator", "app=olm-operator"}
		for _, l := range labels {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-o=jsonpath={range .items[*].spec}{.containers[*].name}{\"\\t\"}", "-n", "openshift-operator-lifecycle-manager", "-l", l).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			amountOfContainers := len(strings.Split(msg, "\t"))
			msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-o=jsonpath={range .items[*].spec}{.containers[*].terminationMessagePolicy}{\"\\t\"}", "-n", "openshift-operator-lifecycle-manager", "-l", l).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			regex := regexp.MustCompile("FallbackToLogsOnError")
			amountWithFallback := len(regex.FindAllStringIndex(msg, -1))
			o.Expect(amountOfContainers).To(o.Equal(amountWithFallback))
			if amountOfContainers != amountWithFallback {
				e2e.Failf("OLM does not have all containers defined with FallbackToLogsOnError terminationMessagePolicy")
			}
		}
	})

	g.It("PolarionID:32613-[OTP][Skipped:Disconnected]Operators won't install if the CSV dependency is already installed", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-High-32613-Operators won't install if the CSV dependency is already installed"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		exutil.SkipBaselineCaps(oc, "None")

		baseDir := exutil.FixturePath("testdata", "olm")
		csTemplate := filepath.Join(baseDir, "catalogsource-image.yaml")

		oc.SetupProject()
		g.By("1) Start to create the CatalogSource CR")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "prometheus-dependency-cs",
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-prometheus-dependency-index:11.0",
			Template:    csTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		defer cs.Delete(itName, dr)
		cs.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok,
			[]string{"catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

		g.By("2) Install the OperatorGroup in a random project")
		ogTemplate := filepath.Join(baseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-32613",
			Namespace: oc.Namespace(),
			Template:  ogTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("3) Install the etcdoperator v0.9.4 with Automatic approval")
		subTemplate := filepath.Join(baseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-32613",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      cs.Name,
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "etcd-service-monitor",
			StartingCSV:            "etcdoperator.v0.9.4",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok,
			[]string{"csv", "etcdoperator.v0.9.4", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("4) Assert that prometheus dependency is resolved")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).To(o.ContainSubstring("prometheus"))

		sub = olmv0util.SubscriptionDescription{
			SubName:                "prometheus-32613",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			IpApproval:             "Automatic",
			Channel:                "beta",
			OperatorPackage:        "prometheus",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.CreateWithoutCheck(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "prometheus-beta-community-operators-openshift-marketplace exists", exutil.Ok,
			[]string{"subs", "prometheus-32613", "-n", oc.Namespace(), "-o=jsonpath={.status.conditions..message}"}).Check(oc)
	})

	g.It("PolarionID:24055-[OTP][Skipped:Disconnected]Check for defaultChannel mandatory field when having multiple channels", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Low-24055-Check for defaultChannel mandatory field when having multiple channels"), func() {
		baseDir := exutil.FixturePath("testdata", "olm")
		cmMapWithoutDefaultChannel := filepath.Join(baseDir, "configmap-without-defaultchannel.yaml")
		cmMapWithDefaultChannel := filepath.Join(baseDir, "configmap-with-defaultchannel.yaml")
		csNamespaced := filepath.Join(baseDir, "catalogsource-namespace.yaml")

		namespace := "scenario3"
		defer olmv0util.RemoveNamespace(namespace, oc)

		g.By("1) Creating a namespace")
		_, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("ns", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) Creating a ConfigMap without a default channel")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", cmMapWithoutDefaultChannel).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("3) Creating a CatalogSource")
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", csNamespaced).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("4) Checking CatalogSource error statement due to the absence of a default channel")
		err = wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			output, pollErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=scenario3", "-n", namespace).Output()
			if pollErr != nil {
				return false, nil
			}
			return strings.Contains(output, "CrashLoopBackOff"), nil
		})
		exutil.AssertWaitPollNoErr(err, "pod of olm.catalogSource=scenario3 is not CrashLoopBackOff")

		g.By("5) Changing the CatalogSource to include default channel for each package")
		_, err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", cmMapWithDefaultChannel).Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("6) Checking the state of CatalogSource(Running)")
		err = wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			output, pollErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=scenario3", "-n", namespace).Output()
			if pollErr != nil {
				return false, nil
			}
			return strings.Contains(output, "Running"), nil
		})
		exutil.AssertWaitPollNoErr(err, "pod of olm.catalogSource=scenario3 is not running")
	})

	g.It("PolarionID:68679-[OTP][Skipped:Disconnected]catalogsource with invalid name is created", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 optional should PolarionID:68679-[Skipped:Disconnected]catalogsource with invalid name is created"), func() {
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-opm.yaml")

		cs := olmv0util.CatalogSourceDescription{
			Name:        "bug-68679-4.14", // the name contains "."
			Namespace:   oc.Namespace(),
			DisplayName: "QE Operators",
			Publisher:   "QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/nginxolm-operator-index:v1",
			Template:    csImageTemplate,
		}
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)
	})

	g.It("PolarionID:31693-[OTP][Skipped:Disconnected]Check CSV information on the PackageManifest", g.Label("original-name:[sig-operators] OLM should ConnectedOnly-Author:bandrade-Medium-31693-Check CSV information on the PackageManifest"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		g.By("1) The relatedImages should exist")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
			"packagemanifest",
			"-n", "openshift-marketplace",
			"prometheus",
			"-o=jsonpath={.status.channels[?(.name=='beta')].currentCSVDesc.relatedImages}",
		).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).NotTo(o.BeEmpty())

		g.By("2) The minKubeVersion should exist")
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args(
			"packagemanifest",
			"-n", "openshift-marketplace",
			"prometheus",
			"-o=jsonpath={.status.channels[?(.name=='beta')].currentCSVDesc.minKubeVersion}",
		).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).NotTo(o.BeEmpty())

		g.By("3) nativeAPI should be optional and empty for prometheus")
		_, err = oc.AsAdmin().WithoutNamespace().Run("get").Args(
			"packagemanifest",
			"-n", "openshift-marketplace",
			"prometheus",
			"-o=jsonpath={.status.channels[?(.name=='beta')].currentCSVDesc.nativeAPIs}",
		).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
	})

})
