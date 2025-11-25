package specs

import (
	"context"
	"crypto/sha256"
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	semver "github.com/blang/semver/v4"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	// TODO: Add github package to vendor to enable GitHub API tests
	// "github.com/google/go-github/v60/github"
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
		oc = exutil.NewCLIWithoutNamespace("default")
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		oc.SetupProject()
		exutil.SkipNoOLMCore(oc)
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

	// Polarion ID: 83027
	g.It("PolarionID:83027-[OTP]-Unnecessary churn with operatorgroup clusterrole management[Serial]", func() {
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) install an operator with AllNamespaces mode")
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		g.By("1-1), create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		// Across namespaces
		ns := "openshift-marketplace"
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-83027",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-83027",
			Namespace:              "openshift-operators",
			CatalogSourceName:      "catsrc-83027",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        false,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer func() {
			if sub.InstalledCSV == "" {
				sub.FindInstalledCSV(oc, itName, dr)
			}
			sub.DeleteCSV(itName, dr)
		}()
		sub.Create(oc, itName, dr)
		g.By("2) get related clusterrole")
		olmClusterRole, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterrole", "-l", "olm.owner=global-operators", "-o", "jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		clusterRoleSelectors, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterrole", olmClusterRole, "-o", "jsonpath={.aggregationRule.clusterRoleSelectors}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		g.By("3) create and delete some projects")
		err = oc.AsAdmin().WithoutNamespace().Run("adm", "new-project").Args("test0-83027").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = oc.AsAdmin().WithoutNamespace().Run("adm", "new-project").Args("test1-83027").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = oc.AsAdmin().WithoutNamespace().Run("adm", "new-project").Args("test2-83027").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("ns", "test0-83027", "test1-83027", "test2-83027", "--force", "--grace-period=0", "--wait=false").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		// waiting for the project deleting
		items := []string{"test0-83027", "test1-83027", "test2-83027"}
		for _, v := range items {
			err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
				err = oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", v).Execute()
				if err != nil {
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("The project %s still exist after 120s", v))
		}
		g.By("4) check if the clusterRoleSelectors order changes")
		newClusterRoleSelectors, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterrole", olmClusterRole, "-o", "jsonpath={.aggregationRule.clusterRoleSelectors}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !reflect.DeepEqual(newClusterRoleSelectors, clusterRoleSelectors) {
			e2e.Failf("Fail: the clusterRoleSelectors order changes, new:%s; old:%s", newClusterRoleSelectors, clusterRoleSelectors)
		}
	})

	// Polarion ID: 70162
	g.It("PolarionID:70162-[OTP]-Leverage Composable OpenShift feature to make OperatorLifecycleManager optional", func() {
		capability := "OperatorLifecycleManager"
		knownCapabilities, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.status.capabilities.knownCapabilities}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("knownCapabilities: %s", knownCapabilities)
		enabledCapabilities, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.status.capabilities.enabledCapabilities}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("enabledCapabilities: %s", enabledCapabilities)
		if strings.Contains(knownCapabilities, capability) && !strings.Contains(enabledCapabilities, capability) {
			// marketplace depnens on olm, so once marketplace enabled, olm enabled
			if strings.Contains(knownCapabilities, "marketplace") && strings.Contains(enabledCapabilities, "marketplace") {
				g.Skip("the cluster enabled marketplace cap and skip it")
			}
			cos := []string{"operator-lifecycle-manager", "operator-lifecycle-manager-catalog", "operator-lifecycle-manager-packageserver"}
			resources := []string{"sub", "csv", "installplan", "operatorgroup", "operatorhub", "catalogsource", "packagemanifest", "olmconfig", "operatorcondition", "operator.operators.coreos.com"}
			clusterroles := []string{"aggregate-olm-edit", "aggregate-olm-view", "cluster-olm-operator"}
			ns := "openshift-operator-lifecycle-manager"
			for _, co := range cos {
				_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("co", co).Output()
				if err == nil {
					e2e.Failf("should not get %v cluster operator", co)
				}
			}
			for _, resource := range resources {
				_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(resource).Output()
				if err == nil {
					e2e.Failf("should not get %v resource", resource)
				}
			}
			for _, clusterrole := range clusterroles {
				// when TP enable, the "cluster-olm-operator" exist
				if exutil.IsTechPreviewNoUpgrade(oc) && clusterrole == "cluster-olm-operator" {
					continue
				}
				_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterrole", clusterrole).Output()
				if err == nil {
					e2e.Failf("should not get %v cluster role", clusterrole)
				}
			}
			_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", ns).Output()
			if err == nil {
				e2e.Failf("should not get %v project", ns)
			}
		} else {
			g.Skip(fmt.Sprintf("the cluster has capability %v and skip it", capability))
		}
	})

	// Polarion ID: 73201
	g.It("PolarionID:73201-[OTP][Skipped:Disconnected]catalog pods do not recover from node failure [Disruptive][Serial]", func() {
		if exutil.IsSNOCluster(oc) {
			g.Skip("This is a SNO cluster, skip.")
		}
		// The cluster node doesn't recover in OSP, GCP, BM... due to the platform issue frequently. So, use the AWS only.
		exutil.SkipIfPlatformTypeNot(oc, "AWS")
		g.By("1, create a custom catalogsource in a random project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-opm.yaml")

		cs := olmv0util.CatalogSourceDescription{
			Name:        "cs-73201",
			Namespace:   oc.Namespace(),
			DisplayName: "QE Operators",
			Publisher:   "QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Interval:    "4m",
			Template:    csImageTemplate,
		}
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)

		g.By("2, get the pod's node and name")
		nodeName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=cs-73201", "-o=jsonpath={.items[0].spec.nodeName}", "-n", oc.Namespace()).Output()
		if err != nil {
			e2e.Failf("Fail to get pod's node:%v", err)
		}

		podName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=cs-73201", "-o=jsonpath={.items[0].metadata.name}", "-n", oc.Namespace()).Output()
		if err != nil {
			e2e.Failf("Fail to get pod's Name:%v", err)
		}

		g.By("3, make the node to NotReady and recover after 600s")
		timeSleep := "600"
		channel := make(chan string)
		go func() {
			cmdStr := fmt.Sprintf(`systemctl stop kubelet; sleep %s; systemctl start kubelet`, timeSleep)
			output, _ := oc.AsAdmin().WithoutNamespace().Run("debug").Args("-n", "default", fmt.Sprintf("nodes/%s", nodeName), "--", "chroot", "/host", "/bin/bash", "-c", cmdStr).Output()
			// if err != nil {
			// 	e2e.Failf("fail to stop node:%v", err)
			// }
			e2e.Logf("!!!!output:%s", output)
			channel <- output
		}()
		defer func() {
			receivedMsg := <-channel
			e2e.Logf("!!!!receivedMsg:%s", receivedMsg)
		}()

		// defer cmd.Process.Kill()
		defer func() {
			var nodeStatus string
			err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 900*time.Second, false, func(ctx context.Context) (bool, error) {
				nodeStatus, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("node", nodeName, "--no-headers").Output()
				if !strings.Contains(nodeStatus, "NotReady") {
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("The node(%s) doesn't recover to Ready status(%s) after 15 mins", nodeName, nodeStatus))
		}()

		g.By("4, check if the node is NotReady")
		var nodeStatus string
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
			nodeStatus, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("node", nodeName, "--no-headers").Output()
			if strings.Contains(nodeStatus, "NotReady") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("The node(%s) still in Ready status(%s) after 300s", nodeName, nodeStatus))

		g.By("5, check if new catalogsource pod generated")
		var podStatus string
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
			podStatus, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=cs-73201", "-n", oc.Namespace(), "--no-headers").Output()
			podNewName, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=cs-73201", "-o=jsonpath={.items[0].metadata.name}", "-n", oc.Namespace()).Output()
			if strings.Contains(podStatus, "Running") && podName != podNewName {
				e2e.Logf("new pod(%s) generated, old pod(%s)", podNewName, podName)
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("No new pod generated after 600s, old pod(%s) status(%s)", podName, podStatus))
	})

	// Polarion ID: 72192
	g.It("PolarionID:72192-[Level0][OTP]-is not correctly refreshing operator catalogs due to IfNotPresent imagePullPolicy", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) get marketplace and OLM pods' image/imagePullPolicy")
		allImageMap := make(map[string]string)
		podMap := make(map[string]string)
		podSlice := olmv0util.GetProjectPods(oc, "openshift-marketplace")
		for _, pod := range podSlice {
			// remove duplicates
			if _, ok := podMap[pod]; !ok {
				podMap[pod] = "openshift-marketplace"
			}
		}
		podSlice1 := olmv0util.GetProjectPods(oc, "openshift-operator-lifecycle-manager")
		for _, pod := range podSlice1 {
			// skip those cronjob pod since they will be deleted every 15 mins that leads error
			if strings.Contains(pod, "collect-profiles") {
				continue
			}
			if _, ok := podMap[pod]; !ok {
				podMap[pod] = "openshift-operator-lifecycle-manager"
			}
		}
		for pod, project := range podMap {
			podImageMap := olmv0util.GetPodImageAndPolicy(oc, pod, project)
			for image, policy := range podImageMap {
				if _, ok := allImageMap[image]; !ok {
					allImageMap[image] = policy
				}
			}
		}
		g.By("2) check the imagePullPolicy of the container that uses the tag image.")
		// remove the cronjob pod imagePullPolicy checking since it will create every 15 mins
		// image, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("cronjob", "collect-profiles", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.jobTemplate.spec.template.spec.containers[0].image}").Output()
		// imagePullPolicy, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("cronjob", "collect-profiles", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.jobTemplate.spec.template.spec.containers[0].imagePullPolicy}").Output()
		// allImageMap[image] = imagePullPolicy
		for image, policy := range allImageMap {
			// check the tag kind image, not the digest image
			if !strings.Contains(image, "@sha256") && strings.Contains(image, ":") {
				if !strings.Contains(policy, "Always") {
					e2e.Failf("%s doesn't use the Always imagePullPolicy! %v", image, allImageMap)
				}
			}
		}
	})

	// Polarion ID: 72017
	g.It("PolarionID:72017-[OTP]-pod panics when EnsureSecretOwnershipAnnotations runs", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) create a secret in the openshift-operator-lifecycle-manager project")
		_, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", "secret-72017", "-n", "openshift-operator-lifecycle-manager").Output()
		if err != nil {
			e2e.Failf("Fail to create secret-72017, error:%v", err)
		}
		defer func() {
			_, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("secret", "secret-72017", "-n", "openshift-operator-lifecycle-manager").Output()
			if err != nil {
				e2e.Failf("Fail to delete secret-72017, error:%v", err)
			}
		}()
		g.By("2) add the olm.managed to it")
		_, err = oc.AsAdmin().WithoutNamespace().Run("label").Args("secret", "secret-72017", "olm.managed=true", "-n", "openshift-operator-lifecycle-manager").Output()
		if err != nil {
			e2e.Failf("Fail to add label olm.managed for secret-72017, error:%v", err)
		}
		g.By("3) restart the olm-operator pod and check if it works well")
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("pods", "-l", "app=olm-operator", "-n", "openshift-operator-lifecycle-manager").Output()
		if err != nil {
			e2e.Failf("Fail to delete olm-operator pod, error:%v", err)
		}
		var status string
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			status, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "app=olm-operator", "-n", "openshift-operator-lifecycle-manager").Output()
			if strings.Contains(status, "Running") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("The olm-operator pod didn't recover after 180s: %s", status))
	})

	// Polarion ID: 72013
	g.It("PolarionID:72013-[Level0][OTP]-Creating an OperatorGroup with Name cluster breaks the whole cluster", func() {
		g.By("1) install a custom OG with the name cluster in the default project")
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogTemplate := filepath.Join(buildPruningBaseDir, "og-allns.yaml")
		err := olmv0util.ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", ogTemplate, "-p", "NAME=cluster", "NAMESPACE=default")
		o.Expect(err).NotTo(o.HaveOccurred())
		defer func() {
			_, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("og", "cluster", "-n", "default").Output()
			if err != nil {
				e2e.Failf("Fail to delete the cluster OG, error:%v", err)
			}
		}()
		g.By("2) the rules of the cluster-admin clusterrole should not null")
		rules, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterrole", "cluster-admin", "-o=jsonpath={.rules}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster-admin clusterrole, error:%v", err)
		}
		if strings.Contains(rules, "null") {
			e2e.Failf("The clusterrole cluster-admin has been changed: %s", rules)
		}
		g.By("3) check if the monitoring CO works well")
		status, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("co", "monitoring").Output()
		if err != nil {
			e2e.Failf("Fail to get monitoring CO, error:%v", err)
		}
		if strings.Contains(status, "subjectaccessreviews.authorization.k8s.io is forbidden") {
			e2e.Failf("The monitoring CO doesn't work well: %s", status)
		}
	})

	// Polarion ID: 71996
	g.It("PolarionID:71996-[OTP]-package-server-manager forbidden securityContext.seLinuxOptions [Serial]", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) Install a custom SCC which the priority is high")
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		sccYAML := filepath.Join(buildPruningBaseDir, "scc.yaml")
		_, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", sccYAML).Output()
		if err != nil {
			e2e.Failf("Fail to create the custom SCC, error:%v", err)
		}
		defer func() {
			_, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("scc", "datadog").Output()
			if err != nil {
				e2e.Failf("Fail to put OLM into a managed state, error:%v", err)
			}
		}()
		g.By("2) delete the PSM pod")
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("pod", "-l", "app=package-server-manager", "-n", "openshift-operator-lifecycle-manager").Output()
		if err != nil {
			e2e.Failf("Fail to delete the PSM pod, error:%v", err)
		}
		g.By("3) check if the PSM pod is recreated well")
		var status string
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			status, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=package-server-manager").Output()
			if strings.Contains(status, "Running") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("PSM pod didn't recover after 180s: %s", status))
	})

	// Polarion ID: 53771
	g.It("PolarionID:53771-[OTP][Skipped:Disconnected]The certificate relating to operator-lifecycle-manager-packageserver isn't rotated after expired [Slow][Disruptive]", g.Label("NonHyperShiftHOST"), func() {
		g.By("enhance steps to cover bug https://issues.redhat.com/browse/OCPBUGS-36138")
		crtTime := strings.Fields(olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.certsRotateAt}\" \"{.status.certsLastUpdated}"))
		o.Expect(crtTime).NotTo(o.BeEmpty())
		certsRotateAt := crtTime[0]
		certsLastUpdated := crtTime[1]

		g.By("1) update the packageserver-service-cert secret to change the crt")
		_, err := oc.AsAdmin().WithoutNamespace().Run("patch").Args("secret", "packageserver-service-cert", "-n", "openshift-operator-lifecycle-manager", "-p", "{\"data\": {\"olmCAKey\" : \"\"}}", "--type=merge").Output()
		if err != nil {
			e2e.Failf("Fail to update packageserver-service-cert secret, error:%v", err)
		}
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			updatedCrtTime := strings.Fields(olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.certsRotateAt}\" \"{.status.certsLastUpdated}"))
			if len(updatedCrtTime) == 0 {
				return false, fmt.Errorf("updatedCrtTime is empty")
			}
			updatedCertsRotateAt := updatedCrtTime[0]
			updatedCertsLastUpdated := updatedCrtTime[1]

			if (updatedCertsRotateAt == certsRotateAt) || (updatedCertsLastUpdated == certsLastUpdated) {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("the packageserver CSV's certsRotateAt(%s) or certsLastUpdated(%s) not updated after 180s", certsRotateAt, certsLastUpdated))

		var image, phase, olmPhase, packagePhase string
		customOLMImage := "quay.io/openshifttest/operator-framework-olm:cert5-rotation-rhel9"
		defer func() {
			_, err := oc.AsAdmin().WithoutNamespace().Run("patch").Args("clusterversion", "version", "-p", "{\"spec\": {\"overrides\":[{\"kind\": \"Deployment\", \"name\": \"olm-operator\", \"namespace\": \"openshift-operator-lifecycle-manager\", \"unmanaged\": false, \"group\": \"apps\"}]}}", "--type=merge").Output()
			if err != nil {
				e2e.Failf("Fail to put OLM into a managed state, error:%v", err)
			}
			err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
				image, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-operator", "-o=jsonpath={.items[0].spec.containers[0].image}").Output()
				olmPhase, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-operator").Output()
				packagePhase, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=packageserver").Output()
				if image != customOLMImage && strings.Contains(olmPhase, "Running") && strings.Contains(packagePhase, "Running") {
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("OLM pod image(%s),olmPhase(%s),packagePhase(%s) didn't recover after 180s", image, olmPhase, packagePhase))
		}()

		g.By("1, put OLM into an unmanaged state")
		_, err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("clusterversion", "version", "-p", "{\"spec\": {\"overrides\":[{\"kind\": \"Deployment\", \"name\": \"olm-operator\", \"namespace\": \"openshift-operator-lifecycle-manager\", \"unmanaged\": true, \"group\": \"apps\"}]}}", "--type=merge").Output()
		if err != nil {
			e2e.Failf("Fail to put OLM into an unmanaged state, error:%v", err)
		}
		g.By("2, patch the OLM operator deployment to utilize a custom version which issues certificates that expire faster")
		_, err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("deployment", "olm-operator", "-p", fmt.Sprintf("{\"spec\": {\"template\": {\"spec\": {\"containers\": [{\"name\": \"olm-operator\", \"image\": \"%s\"}]}}}}", customOLMImage), "-n", "openshift-operator-lifecycle-manager").Output()
		if err != nil {
			e2e.Failf("Fail to patch the OLM operator deployment, error:%v", err)
		}
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
			image, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-operator", "-o=jsonpath={.items[0].spec.containers[0].image}").Output()
			phase, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-operator").Output()
			if image == customOLMImage && strings.Contains(phase, "Running") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("the olm-operator pod image(%s) and phase(%s) not updated after 600s", image, phase))
		g.By("3, delete the existing packageserver cert to initiate the creation of a new one")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			info, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("secret", "packageserver-service-cert", "--wait=true", "-n", "openshift-operator-lifecycle-manager").Output()
			if !strings.Contains(info, "deleted") || err != nil {
				e2e.Logf("Warning! Fail to delete the old packageserver cert, error:%v, retrying...", err)
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "fail to delete the old packageserver cert after 180s")
		g.By("4, check that the cert has the faster expiration date as expected")
		certsLastUpdad0, certsRotateAt0 := olmv0util.GetCertRotation(oc, "packageserver-service-cert", "openshift-operator-lifecycle-manager")
		g.By("4-1, waiting 5 mins here until the expiration time, and check again if there is a new certificate that has been created.")
		time.Sleep(5 * time.Minute)
		var certsLastUpdad1, certsRotateAt1 time.Time
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			certsLastUpdad1, certsRotateAt1 = olmv0util.GetCertRotation(oc, "packageserver-service-cert", "openshift-operator-lifecycle-manager")
			if certsLastUpdad0.Equal(certsLastUpdad1) && certsRotateAt0.Equal(certsRotateAt1) {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("The rotation time Not changed! Before: certsLastUpdad:%v, certsRotateAt:%v\n After: certsLastUpdad:%v, certsRotateAt:%v\n", certsLastUpdad0, certsRotateAt0, certsLastUpdad1, certsRotateAt1))
		g.By("5, recreate the packageserver pods, and check if the cert is rotated")
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=packageserver").Output()
		if err != nil {
			e2e.Failf("Fail to delete packageserver pods, error:%v", err)
		}
		certsLastUpdad2, certsRotateAt2 := olmv0util.GetCertRotation(oc, "packageserver-service-cert", "openshift-operator-lifecycle-manager")
		if !certsLastUpdad1.Equal(certsLastUpdad2) || !certsRotateAt1.Equal(certsRotateAt2) {
			e2e.Failf("The rotation time changed! Before: certsLastUpdad:%v, certsRotateAt:%v\n After: certsLastUpdad:%v, certsRotateAt:%v\n", certsLastUpdad1, certsRotateAt1, certsLastUpdad2, certsRotateAt2)
		}
	})

	// Polarion ID: 68681
	g.It("PolarionID:68681-[OTP][Skipped:Disconnected]pods with no controller true ownerReferences", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		defaultCatalogSources := []string{"certified-operators", "community-operators", "redhat-marketplace", "redhat-operators"}
		g.By("1) check default catalog sources' pods if labeled with controller: true")
		for _, cs := range defaultCatalogSources {
			controller, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", fmt.Sprintf("olm.catalogSource=%s", cs), "-o=jsonpath={.items[0].metadata.ownerReferences[0].controller}", "-n", "openshift-marketplace").Output()
			if err != nil {
				e2e.Failf("fail to get %s's pod's controller label, error:%v", cs, err)
			}
			if controller != "true" {
				e2e.Failf("%s's pod's controller is not true!", cs)
			}
		}
	})

	// Polarion ID: 59413
	g.It("PolarionID:59413-[OTP][Skipped:Disconnected]Default CatalogSource aren't created in restricted mode [Serial]", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		defaultCatalogSources := []string{"certified-operators", "community-operators", "redhat-marketplace", "redhat-operators"}
		g.By("step 1 -> check if the SCC is restricted")
		for _, cs := range defaultCatalogSources {
			SCC, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", cs, "-o=jsonpath={.spec.grpcPodConfig.securityContextConfig}", "-n", "openshift-marketplace").Output()
			if err != nil {
				e2e.Failf("fail to get %s's SCC, error:%v", cs, err)
			}
			if SCC != "restricted" {
				e2e.Failf("%s's SCC is not restricted!", cs)
			}
		}
		g.By("step 2 -> change the default SCC to legacy")
		for _, cs := range defaultCatalogSources {
			olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", "openshift-marketplace", "catalogsource", cs, "-p", "{\"spec\":{\"grpcPodConfig\": {\"securityContextConfig\": \"legacy\"}}}", "--type=merge")
		}
		g.By("step 3 -> check if SCC reset the restricted")
		for _, cs := range defaultCatalogSources {
			SCC, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", cs, "-o=jsonpath={.spec.grpcPodConfig.securityContextConfig}", "-n", "openshift-marketplace").Output()
			if err != nil {
				e2e.Failf("fail to get %s's SCC, error:%v", cs, err)
			}
			if SCC != "restricted" {
				e2e.Failf("%s's SCC(%s) is not restricted!", cs, SCC)
			}
		}
	})

	// Polarion ID: 59422
	g.It("PolarionID:59422-[OTP]-package-server-manager does not stomp on changes made to packgeserver CSV", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) change the packageser CSV's securityContext")
		packageserverCSVYaml, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o", "yaml").OutputToFile("ocp59422-csv.yaml")
		o.Expect(err).NotTo(o.HaveOccurred())
		exutil.ModifyYamlFileContent(packageserverCSVYaml, []exutil.YamlReplace{
			{
				Path:  "spec.install.spec.deployments.0.spec.template.spec.containers.0.securityContext.allowPrivilegeEscalation",
				Value: "true",
			},
			{
				Path:  "spec.install.spec.deployments.0.spec.template.spec.securityContext.runAsNonRoot",
				Value: "false",
			},
		})
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", packageserverCSVYaml, "-n", "openshift-operator-lifecycle-manager").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("2) check if the packageserver CSV's securityContext config reback")
		err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			allowPrivilegeEscalation, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.install.spec.deployments[0].spec.template.spec.containers[0].securityContext.allowPrivilegeEscalation}").Output()
			if err != nil {
				return false, nil
			}
			if allowPrivilegeEscalation != "false" {
				// save output, so comment it
				// e2e.Logf("The packageserver CSV was not reset, allowPrivilegeEscalation is %s", allowPrivilegeEscalation)
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "The packageserver CSV was not reset, allowPrivilegeEscalation still is true after 60s!")
		runAsNonRoot, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.install.spec.deployments[0].spec.template.spec.securityContext.runAsNonRoot}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if runAsNonRoot != "true" {
			e2e.Failf("The packageserver CSV was not reset, runAsNonRoot is %s", runAsNonRoot)
		}
	})

	// Polarion ID: 59416
	g.It("PolarionID:59416-[OTP][Skipped:Disconnected]Revert Catalog PSA decisions for 4.12 [Serial]", g.Label("NonHyperShiftHOST"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		node, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		err = exutil.SetNamespacePrivileged(oc, oc.Namespace())
		o.Expect(err).NotTo(o.HaveOccurred())
		efips, err := oc.AsAdmin().WithoutNamespace().Run("debug").Args("node/"+node, "--to-namespace="+oc.Namespace(), "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
		if err != nil || strings.Contains(efips, "FIPS mode is enabled") {
			g.Skip("skip it with cmd failure or FIPS enabled")
		}
		g.By("step 1 -> check openshift-marketplace project labels")
		labels, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", "openshift-marketplace", "--show-labels").Output()
		if err != nil {
			e2e.Failf("fail to get openshift-marketplace project labels, error:%v", err)
		}
		if !strings.Contains(labels, "pod-security.kubernetes.io/enforce=baseline") {
			e2e.Failf("openshift-marketplace project PSA is not baseline: %s", labels)
		}
		g.By("step 2 -> deploy two catalog sources with old index images, both of them should work well without the restricted SCC")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "cs-without-scc.yaml")

		indexImages := []string{"quay.io/olmqe/ditto-index:test-xzha-1", "quay.io/olmqe/etcd-index:v1new"}
		for i, indexImage := range indexImages {
			cs := olmv0util.CatalogSourceDescription{
				Name:        fmt.Sprintf("cs-59416-%d", i),
				Namespace:   "openshift-marketplace",
				DisplayName: "QE Operators",
				Publisher:   "QE",
				SourceType:  "grpc",
				Address:     indexImage,
				Template:    csImageTemplate,
			}
			defer cs.Delete(itName, dr)
			cs.CreateWithCheck(oc, itName, dr)
		}
	})

	// Polarion ID: 53914
	g.It("PolarionID:53914-[OTP]-OLM controller plug-in for openshift-* namespace labelling [Serial]", func() {
		// openshifttest-53914 without openshift- prefix
		// openshift-test-53914 without the `security.openshift.io/scc.podSecurityLabelSync=true` label
		// openshift-test-53914 with the `security.openshift.io/scc.podSecurityLabelSync=true` label
		g.By("Starting ../ prepare projects")
		projects := []olmv0util.ProjectDescription{
			{Name: "openshifttest-53914", TargetNamespace: ""},
			{Name: "openshift-test1-53914", TargetNamespace: ""},
			{Name: "openshift-test2-53914", TargetNamespace: ""},
			{Name: "default", TargetNamespace: ""},
			{Name: "openshift-test3-53914", TargetNamespace: ""},
			{Name: "openshift-operators", TargetNamespace: ""},
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")

		g.By("create the learn-operator CatalogSource")
		ns := "openshift-marketplace"
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-53914",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		for i, project := range projects {
			g.By(fmt.Sprintf("step-%d, subscribe to learn perator v0.0.3 in project %s", i, project.Name))
			if project.Name != "default" && project.Name != "openshift-operators" {
				project.CreateWithCheck(oc, itName, dr)
				defer func(p olmv0util.ProjectDescription) {
					p.DeleteWithForce(oc)
				}(project)
			}
			// this project just for verifying the Copied CSV
			if project.Name == "openshift-test3-53914" {
				continue
			}
			if project.Name == "openshift-test2-53914" {
				_, err := oc.AsAdmin().WithoutNamespace().Run("label").Args("ns", project.Name, "security.openshift.io/scc.podSecurityLabelSync=false").Output()
				if err != nil {
					e2e.Failf("Fail to label project %s with security.openshift.io/scc.podSecurityLabelSync=false, error:%v", project.Name, err)
				}
			}
			var og olmv0util.OperatorGroupDescription
			if project.Name != "openshift-operators" {
				og = olmv0util.OperatorGroupDescription{
					Name:      fmt.Sprintf("og%d-53914", i),
					Namespace: project.Name,
					Template:  ogSingleTemplate,
				}
				defer og.Delete(itName, dr)
				og.CreateWithCheck(oc, itName, dr)
			}

			var single bool
			if project.Name == "openshift-operators" {
				single = false
			} else {
				single = true
			}

			sub := olmv0util.SubscriptionDescription{
				SubName:                fmt.Sprintf("sub%d-53914", i),
				Namespace:              project.Name,
				CatalogSourceName:      "catsrc-53914",
				CatalogSourceNamespace: ns,
				Channel:                "beta",
				IpApproval:             "Automatic",
				OperatorPackage:        "learn",
				StartingCSV:            "learn-operator.v0.0.3",
				SingleNamespace:        single,
				Template:               subTemplate,
			}
			defer sub.Delete(itName, dr)
			defer func() {
				if sub.InstalledCSV == "" {
					sub.FindInstalledCSV(oc, itName, dr)
				}
				sub.DeleteCSV(itName, dr)
			}()
			sub.Create(oc, itName, dr)
			// skip default namespace's csv status checking since it will fail due to PSA issue
			if project.Name == "default" {
				// it takes a long time to update to the Failed status
				olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, true, "", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", project.Name}).Check(oc)
			} else {
				olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded-TIME-WAIT-120s", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", project.Name, "-o=jsonpath={.status.phase}"}).Check(oc)
			}
			labels, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", project.Name, "-o=jsonpath={.metadata.labels}").Output()
			if err != nil {
				e2e.Failf("Fail to get project %s labels, error:%v", project, err)
			}
			switch project.Name {
			case "openshifttest-53914":
				if strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
					e2e.Failf("project %s should NOT be labeled with security.openshift.io/scc.podSecurityLabelSync=true, labels:%s", project.Name, labels)
				}
			case "openshift-test-53914":
				if !strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
					e2e.Failf("project %s should be labeled with security.openshift.io/scc.podSecurityLabelSync=true, labels:%s", project.Name, labels)
				}
			case "openshift-test2-53914":
				if strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
					e2e.Failf("project %s should NOT be updated with security.openshift.io/scc.podSecurityLabelSync=true, labels:%s", project.Name, labels)
				}
				// project should be re-labeled  with `security.openshift.io/scc.podSecurityLabelSync=true` after `security.openshift.io/scc.podSecurityLabelSync=false` removed
				_, err := oc.AsAdmin().WithoutNamespace().Run("label").Args("ns", project.Name, "security.openshift.io/scc.podSecurityLabelSync-").Output()
				if err != nil {
					e2e.Failf("Fail to unlabel project %s with security.openshift.io/scc.podSecurityLabelSync-, error:%v", project.Name, err)
				}
				err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
					labels, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", project.Name, "-o=jsonpath={.metadata.labels}").Output()
					if err != nil || !strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
						e2e.Logf("The label not updated, re-try: %s", err)
						return false, nil
					}
					return true, nil
				})
				exutil.AssertWaitPollNoErr(err, "Fail to re-label project openshift-test2-53914 after 60s!")
			case "default":
				if strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
					e2e.Failf("project %s should NOT be labeled with security.openshift.io/scc.podSecurityLabelSync=true, labels:%s", project.Name, labels)
				}
			case "openshift-operators":
				if !strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
					e2e.Failf("project %s should be labeled with security.openshift.io/scc.podSecurityLabelSync=true, labels:%s", project.Name, labels)
				}
				// The project with a copied CSV in should NOT be labeled with security.openshift.io/scc.podSecurityLabelSync=true
				labels, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", "openshift-test3-53914", "-o=jsonpath={.metadata.labels}").Output()
				if err != nil {
					e2e.Failf("Fail to get project openshift-test3-53914 labels:%s, error:%v", err, labels)
				}
				if strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
					e2e.Failf("project openshift-test-53914 should NOT be labeled with security.openshift.io/scc.podSecurityLabelSync=true since copied CSV, labels:%s", labels)
				}
			}
			sub.Delete(itName, dr)
			sub.DeleteCSV(itName, dr)

			if project.Name != "openshifttest-53914" && project.Name != "default" {
				//  The `security.openshift.io/scc.podSecurityLabelSync=true` won't be removed.
				err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
					labels, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", project.Name, "-o=jsonpath={.metadata.labels}").Output()
					if err != nil || !strings.Contains(labels, "\"security.openshift.io/scc.podSecurityLabelSync\":\"true\"") {
						e2e.Logf("security.openshift.io/scc.podSecurityLabelSync=true should NOT be removed from project %s after CSV removed, labels:%s", project.Name, labels)
						return false, nil
					}
					return true, nil
				})
				exutil.AssertWaitPollNoErr(err, fmt.Sprintf("The security.openshift.io/scc.podSecurityLabelSync=true label of project:%s should NOT be removed!", project.Name))
			}
		}
	})

	// Polarion ID: 53759
	g.It("PolarionID:53759-[OTP]-Opeatorhub status shows errors after disabling default catalogSources [Disruptive]", g.Label("NonHyperShiftHOST"), func() {
		g.By("1, check if the marketplace enabled")
		cap, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.status.capabilities.enabledCapabilities}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster capabilities: %s, error:%v", cap, err)
		}
		if !strings.Contains(cap, "marketplace") {
			g.Skip("marketplace is disabled, skip...")
		}
		g.By("2, check if the default catalogsource disabled")
		disable, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("operatorhub", "cluster", "-o=jsonpath={.spec.disableAllDefaultSources}").Output()
		if err != nil {
			e2e.Failf("Fail to get operatorhub spec, error:%v", err)
		}
		if disable != "true" {
			g.By("2-1, Disable the default catalogsource")
			// make sure the operatorhub enabled after this test
			defer func() {
				_, err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("operatorhub", "cluster", "-p", "{\"spec\": {\"disableAllDefaultSources\": false}}", "--type=merge").Output()
				if err != nil {
					e2e.Failf("Fail to re-enable operatorhub, error:%v", err)
				}
			}()
			_, err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("operatorhub", "cluster", "-p", "{\"spec\": {\"disableAllDefaultSources\": true}}", "--type=merge").Output()
			if err != nil {
				e2e.Failf("Fail to disable operatorhub, error:%v", err)
			}
		}
		g.By("3, Check the OperatorHub status")
		status, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("operatorhub", "cluster", "-o=jsonpath={.status.sources}").Output()
		if err != nil {
			e2e.Failf("Fail to get operatorhub status, error:%v", err)
		}
		if strings.Contains(status, "Error") {
			e2e.Failf("the operatorhub status(%s) is incorrect!", status)
		}
		log, _ := oc.AsAdmin().WithoutNamespace().Run("logs").Args("deploy/marketplace-operator", "--tail", "3").Output()
		if strings.Contains(log, "Error processing CatalogSource") {
			e2e.Failf("marketplace-operator is handling operatorhub wrongly: %s", log)
		}
	})

	// Polarion ID: 53758
	g.It("PolarionID:53758-[OTP][Skipped:Disconnected]failed to recreate SA for the CatalogSource that without poll Interval", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1, Create a CatalogSource that in a random project")
		oc.SetupProject()
		csImageTemplate := filepath.Join(buildPruningBaseDir, "cs-without-interval.yaml")
		indexImage := "quay.io/openshifttest/nginxolm-operator-index:nginxolm99999"
		cs := olmv0util.CatalogSourceDescription{
			Name:        "cs-53758",
			Namespace:   oc.Namespace(),
			DisplayName: "QE Operators",
			Publisher:   "QE",
			SourceType:  "grpc",
			Address:     indexImage,
			Template:    csImageTemplate,
		}
		defer cs.Delete(itName, dr)
		cs.CreateWithCheck(oc, itName, dr)

		g.By("2, delete this CatalogSource's SA")
		_, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("serviceaccount", cs.Name, "-n", cs.Namespace).Output()
		if err != nil {
			e2e.Failf("fail to delete the catalogsource SA:%s", cs.Name)
		}
		g.By("3, check if SA is recreated")
		err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			_, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("serviceaccount", cs.Name, "-n", cs.Namespace).Output()
			if err != nil {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("fail to recreate the catalogsource SA %s after 60s!", cs.Name))
	})

	// Polarion ID: 53740
	g.It("PolarionID:53740-[OTP]-CatalogSource incorrect parsing validation", g.Label("NonHyperShiftHOST"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1, Create a CatalogSource that in a random project")
		oc.SetupProject()
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-53740",
			Namespace: oc.Namespace(),
			Template:  ogSingleTemplate,
		}
		defer og.Delete(itName, dr)
		og.CreateWithCheck(oc, itName, dr)
		csImageTemplate := filepath.Join(buildPruningBaseDir, "cs-image-template.yaml")
		ocpVersionByte, err := exec.Command("bash", "-c", "oc version -o json | jq -r '.openshiftVersion' | cut -d '.' -f1,2").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		ocpVersion := strings.ReplaceAll(string(ocpVersionByte), "\n", "")
		indexImage := fmt.Sprintf("quay.io/openshift-qe-optional-operators/aosqe-index:v%s", ocpVersion)
		cs := olmv0util.CatalogSourceDescription{
			Name:        "cs-53740",
			Namespace:   oc.Namespace(),
			DisplayName: "QE Operators",
			Publisher:   "QE",
			SourceType:  "grpc",
			Address:     indexImage,
			Interval:    "15mError code",
			Template:    csImageTemplate,
		}
		defer cs.Delete(itName, dr)
		cs.Create(oc, itName, dr)
		var msg string
		err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status.message}").Output()
			if !strings.Contains(msg, "error parsing") {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			e2e.Failf("cannot find the parsing error from CatalogSource message: %s", msg)
		}

		// No error logs print as default after refactor, details: https://github.com/operator-framework/api/blob/master/pkg/operators/v1alpha1/catalogsource_types.go#L157-L177
		// log, _ := oc.AsAdmin().WithoutNamespace().Run("logs").Args("-n", "openshift-marketplace", "deploy/marketplace-operator", "--tail", "3").Output()
		// if !strings.Contains(log, "time: unknown unit") {
		// 	e2e.Failf("cannot find the parsing error logs from marketplace-operator: %s", log)
		// }
	})

	// Polarion ID: 49687
	g.It("PolarionID:49687-[OTP]-Make the marketplace operator optional", func() {
		exutil.SkipBaselineCaps(oc, "None")
		g.By("1, check if the marketplace disabled")
		cap, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.status.capabilities.enabledCapabilities}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster capabilities: %s, error:%v", cap, err)
		}
		if strings.Contains(cap, "marketplace") {
			g.Skip("marketplace is enabled, skip...")
		} else {
			e2e.Logf("marketplace is disabled")
			g.By("2, check marketplace namespace")
			_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ns", "openshift-marketplace").Output()
			if err == nil {
				e2e.Failf("error! openshift-marketplace project still exist")
			}
			g.By("3, check operatorhub namespace")
			_, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("operatorhub").Output()
			if err == nil {
				e2e.Failf("error! operatorhub resource still exist")
			}

			buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
			dr := make(olmv0util.DescriberResrouce)
			itName := g.CurrentSpecReport().FullText()
			dr.AddIr(itName)

			g.By("4, Create a CatalogSource that in a random project")
			oc.SetupProject()
			ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			og := olmv0util.OperatorGroupDescription{
				Name:      "og-49687",
				Namespace: oc.Namespace(),
				Template:  ogSingleTemplate,
			}
			defer og.Delete(itName, dr)
			og.CreateWithCheck(oc, itName, dr)
			csImageTemplate := filepath.Join(buildPruningBaseDir, "cs-image-template.yaml")
			ocpVersionByte, err := exec.Command("bash", "-c", "oc version -o json | jq -r '.openshiftVersion' | cut -d '.' -f1,2").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			ocpVersion := strings.ReplaceAll(string(ocpVersionByte), "\n", "")
			indexImage := fmt.Sprintf("quay.io/openshift-qe-optional-operators/aosqe-index:v%s", ocpVersion)
			cs := olmv0util.CatalogSourceDescription{
				Name:        "cs-49687",
				Namespace:   oc.Namespace(),
				DisplayName: "QE Operators",
				Publisher:   "QE",
				SourceType:  "grpc",
				Address:     indexImage,
				Template:    csImageTemplate,
			}
			defer cs.Delete(itName, dr)
			cs.CreateWithCheck(oc, itName, dr)

			g.By("5, Subscribe to learn perator v0.0.3 in this random project")
			subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			sub := olmv0util.SubscriptionDescription{
				SubName:                "sub-49687",
				Namespace:              oc.Namespace(),
				CatalogSourceName:      "cs-49687",
				CatalogSourceNamespace: oc.Namespace(),
				Channel:                "beta",
				IpApproval:             "Automatic",
				OperatorPackage:        "learn",
				StartingCSV:            "learn-operator.v0.0.3",
				SingleNamespace:        true,
				Template:               subTemplate,
			}
			defer sub.Delete(itName, dr)
			sub.Create(oc, itName, dr)
			defer sub.DeleteCSV(itName, dr)
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)
		}
	})

	// Polarion ID: 49352
	g.It("PolarionID:49352-[OTP]-SNO Leader election conventions for cluster topology", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) get the cluster topology")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.controlPlaneTopology}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster infra: %s, error:%v", infra, err)
		}
		g.By("2) get the leaseDurationSeconds of the packageserver-controller-lock")
		leaseDurationSeconds, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("lease", "packageserver-controller-lock", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.leaseDurationSeconds}").Output()
		if err != nil {
			e2e.Failf("Fail to get the leaseDurationSeconds: %s, error:%v", leaseDurationSeconds, err)
		}
		if infra == "SingleReplica" {
			e2e.Logf("This is a SNO cluster")
			if !strings.Contains(leaseDurationSeconds, "270") {
				e2e.Failf("The lease duration is not as expected: %s", leaseDurationSeconds)
			}
		} else {
			g.Skip("This is a HA cluster, skip.")
		}
	})

	// Polarion ID: 49167
	g.It("PolarionID:49167-[OTP]-fatal error", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) Check OLM related resources' logs")
		deps := []string{"catalog-operator", "olm-operator", "package-server-manager", "packageserver"}
		// since https://issues.redhat.com/browse/OCPBUGS-13369 closed as Wont'do. I remove the certification checking
		// since https://issues.redhat.com/browse/OCPBUGS-43581 fixed, I add the certification checking back, but for OCP4.18+
		re1, _ := regexp.Compile(".*x509.*")
		// since https://issues.redhat.com/browse/OCPBUGS-11370, add "bad certificate" checking for prometheus pods
		re2, _ := regexp.Compile("bad certificate")
		// remove the promtheus checking since many failure not caused by OLM
		// prometheusLogs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args("deployment/prometheus-operator-admission-webhook", "-n", "openshift-monitoring").Output()
		// if err != nil {
		// 	e2e.Failf("!!! Fail to get prometheus logs:%s", err)
		// }
		// prometheusTLS := re2.FindString(prometheusLogs)
		// if re2.FindString(prometheusLogs) != "" {
		// 	e2e.Failf("!!! prometheus occurs TLS error: %s", prometheusTLS)
		// }

		clusterReadyTime, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.status.history[0].completionTime}").Output()
		if err != nil {
			e2e.Failf("!!! Fail to get cluster completion time: %s.", clusterReadyTime)
		}
		re3, _ := regexp.Compile("fatal error.*")
		for _, dep := range deps {
			// got the logs since the cluster ready
			logs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args(fmt.Sprintf("deployment/%s", dep), fmt.Sprintf("--since-time=%s", clusterReadyTime), "-n", "openshift-operator-lifecycle-manager").Output()
			if err != nil {
				e2e.Failf("!!! Fail to get %s logs. Cluster ready time:%v, Error:%v", dep, clusterReadyTime, err)
			}
			str1 := re1.FindString(logs)
			str2 := re2.FindString(logs)
			str3 := re3.FindString(logs)
			// Example
			// str1 = `E0328 19:45:14.489789 1 authentication.go:74] "Unable to authenticate the request" err="[x509: certificate signed by unknown authority, verifying certificate SN=3501846746045108574, SKID=50:8E:D6:EA:A2:11:4B:88:5C:E7:37:47:9C:09:0F:C6:41:7F:80:FD, AKID=82:C7:6C:A5:D0:58:46:19:88:A8:19:31:2C:18:08:37:44:A6:31:7E failed: x509: certificate signed by unknown authority]"`
			// str1 = `2025/04/12 19:20:12 http: TLS handshake error from 10.131.0.11:52966: tls: failed to verify certificate: x509: certificate signed by unknown authority`
			// str1 = `2025-04-01T03:03:56.545422887Z 2025/03/30 03:03:56 http: TLS handshake error from 10.131.0.8:39932: tls: failed to verify certificate: x509: certificate signed by unknown authority`

			if str1 != "" && !strings.Contains(str1, "could not convert APIService CA bundle to x509 cert") {
				failure := true
				if dep == "packageserver" {
					output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("deploy", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.replicas} {.status.readyReplicas}").Output()
					if err != nil {
						e2e.Failf("Fail to get the packageserver replicas/readyReplicas: %s, error:%v", output, err)
					}
					fields := strings.Fields(output)
					if len(fields) != 2 {
						e2e.Failf("Unexpected output format: %q", output)
					}
					if fields[0] == fields[1] {
						failure = false
					} else {
						e2e.Logf("!!! packageserver not ready: replicas=%s, ready=%s", fields[0], fields[1])
					}
				}
				if failure {
					e2e.Failf("!!! %s occurs x509 error: %s", dep, str1)
				}
			}
			if str2 != "" {
				e2e.Failf("!!! %s occurs TLS error: %s", dep, str2)
			}
			if str3 != "" {
				e2e.Failf("!!! %s occurs fatal error: %s", dep, str3)
			}
		}
	})

	// Polarion ID: 46964
	g.It("PolarionID:46964-[OTP][Skipped:Disconnected]Disable Copied CSVs Toggle [Serial]", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Subscribe to learn operator v0.0.3 with AllNamespaces mode")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		ns := "openshift-marketplace"
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-46964",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-learn-46964",
			Namespace:              "openshift-operators",
			CatalogSourceName:      "catsrc-46964",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer func() {
			if sub.InstalledCSV == "" {
				sub.FindInstalledCSV(oc, itName, dr)
			}
			sub.DeleteCSV(itName, dr)
		}()
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", "openshift-operators", "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("2) Create testing projects and Multi OperatorGroup")
		ogMultiTemplate := filepath.Join(buildPruningBaseDir, "og-multins.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:         "og-46964",
			Namespace:    "",
			Multinslabel: "label-46964",
			Template:     ogMultiTemplate,
		}
		p1 := olmv0util.ProjectDescription{
			Name:            "test-46964",
			TargetNamespace: "",
		}
		p2 := olmv0util.ProjectDescription{
			Name:            "test1-46964",
			TargetNamespace: "",
		}

		defer p1.DeleteWithForce(oc)
		defer p2.DeleteWithForce(oc)
		oc.SetupProject()
		p1.TargetNamespace = oc.Namespace()
		p2.TargetNamespace = oc.Namespace()
		og.Namespace = oc.Namespace()
		g.By("2-1) create new projects and label them")
		p1.Create(oc, itName, dr)
		p1.Label(oc, "label-46964")
		p2.Create(oc, itName, dr)
		p2.Label(oc, "label-46964")
		og.Create(oc, itName, dr)

		g.By("3) Subscribe to Sample operator with MultiNamespaces mode")
		subSample := olmv0util.SubscriptionDescription{
			SubName:                "sub-sample-46964",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "catsrc-46964",
			CatalogSourceNamespace: ns,
			Channel:                "alpha",
			IpApproval:             "Automatic",
			OperatorPackage:        "sample-operator",
			Template:               subTemplate,
		}
		defer subSample.Delete(itName, dr)
		subSample.Create(oc, itName, dr)
		defer subSample.DeleteCSV(itName, dr)
		subSample.FindInstalledCSV(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", subSample.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("4) Enable this `disableCopiedCSVs` feature")
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "olmconfig", "cluster", "-p", "{\"spec\":{\"features\":{\"disableCopiedCSVs\": true}}}", "--type=merge")

		g.By("5) Check if the AllNamespaces Copied CSV are removed")
		err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			copiedCSV, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", oc.Namespace(), "--no-headers").Output()
			if err != nil {
				e2e.Failf("Error: %v, fail to get CSVs in project: %s", err, ns)
			}
			if strings.Contains(copiedCSV, "learn-operator.v0.0.3") || !strings.Contains(copiedCSV, subSample.InstalledCSV) {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "AllNamespace Copied CSV should be remove")

		g.By("6) Disable this `disableCopiedCSVs` feature")
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "olmconfig", "cluster", "-p", "{\"spec\":{\"features\":{\"disableCopiedCSVs\": false}}}", "--type=merge")

		g.By("7) Check if the AllNamespaces Copied CSV are back")
		err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
			copiedCSV, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", oc.Namespace(), "--no-headers").Output()
			if err != nil {
				e2e.Failf("Error: %v, fail to get CSVs in project: %s", err, ns)
			}
			if !strings.Contains(copiedCSV, "learn-operator.v0.0.3") || !strings.Contains(copiedCSV, subSample.InstalledCSV) {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "AllNamespaces CopiedCSV should be back")
	})

	// Polarion ID: 43487
	g.It("PolarionID:43487-[OTP]-3rd party Operator Catalog references change during an OCP Upgrade", func() {
		g.By("1) get the Kubernetes version")
		version, err := exec.Command("bash", "-c", "oc version | grep Kubernetes |awk '{print $3}'").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		v, _ := semver.ParseTolerant(string(version))
		majorVersion := strconv.FormatUint(v.Major, 10)
		minorVersion := strconv.FormatUint(v.Minor, 10)
		patchVersion := strconv.FormatUint(v.Patch, 10)

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		imageTemplates := map[string]string{
			"quay.io/kube-release-v{kube_major_version}/catalog:v{kube_major_version}":                                       majorVersion,
			"quay.io/kube-release-v{kube_major_version}/catalog:v{kube_major_version}.{kube_minor_version}":                  fmt.Sprintf("%s.%s", majorVersion, minorVersion),
			"quay.io/olmqe-v{kube_major_version}/etcd-index:v{kube_major_version}.{kube_minor_version}.{kube_patch_version}": fmt.Sprintf("%s.%s.%s", majorVersion, minorVersion, patchVersion),
		}

		oc.SetupProject()
		for k, fullV := range imageTemplates {
			g.By(fmt.Sprintf("create a CatalogSource with imageTemplate:%s", k))
			buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
			csImageTemplate := filepath.Join(buildPruningBaseDir, "cs-image-template.yaml")
			cs := olmv0util.CatalogSourceDescription{
				Name:          fmt.Sprintf("cs-43487-%s", fullV),
				Namespace:     oc.Namespace(),
				DisplayName:   "OLM QE Operators",
				Publisher:     "Jian",
				SourceType:    "grpc",
				Address:       "quay.io/olmqe-v1/etcd-index:v1.21",
				ImageTemplate: k,
				Template:      csImageTemplate,
			}

			defer cs.Delete(itName, dr)
			cs.Create(oc, itName, dr)
			// It will fail due to "ImagePullBackOff" since no this CatalogSource image in fact, so remove the status checking
			// olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", cs.Name, "-n", oc.Namespace(), "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

			g.By("3) get the real CatalogSource image version")
			err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
				// oc get catalogsource cs-43487 -o=jsonpath={.spec.image}
				image, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", cs.Name, "-n", oc.Namespace(), "-o=jsonpath={.spec.image}").Output()
				if err != nil {
					e2e.Failf("Fail to get the CatalogSource(%s)'s image, error: %v", cs.Name, err)
				}
				if image == "" {
					return false, nil
				}

				reg1 := regexp.MustCompile(`.*-v(\d+).*:v(\d+(.\d+)?(.\d+)?)`)
				if reg1 == nil {
					e2e.Failf("image regexp err!")
				}
				result1 := reg1.FindAllStringSubmatch(image, -1)
				imageMajorVersion := result1[0][1]
				imageFullVersion := result1[0][2]
				e2e.Logf("fullVersion:%s, majorVersion:%s, imageFullVersion:%s, imageMajorVersion:%s", fullV, majorVersion, imageFullVersion, imageMajorVersion)
				if imageMajorVersion != majorVersion || imageFullVersion != fullV {
					e2e.Failf("This CatalogSource(%s) image version(%s) doesn't follow the image template(%s)!", cs.Name, image, k)
				}
				return true, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("catsrc %s image version not expected", cs.Name))
		}
	})

	// Polarion ID: 43271
	g.It("PolarionID:43271-[OTP]-Bundle Content Compression", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Subscribe to the Learn operator in a random project")
		oc.SetupProject()
		ns := oc.Namespace()
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-43271",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		defer og.Delete(itName, dr)
		og.CreateWithCheck(oc, itName, dr)

		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-43271",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-43271",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-43271",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)
		defer sub.DeleteCSV(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.StartingCSV, "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("2) get its unpacking job name")
		installPlanName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", "-n", ns, "sub-43271", "-o=jsonpath={.status.installplan.name}").Output()
		if err != nil || installPlanName == "" {
			e2e.Failf("Fail to get its InstallPlan Name: %v", err)
		}

		installPlanPath, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("InstallPlan", "-n", ns, installPlanName, fmt.Sprintf("-o=jsonpath={.status.bundleLookups[?(@.identifier==\"%s\")].path}", sub.StartingCSV)).Output()
		if err != nil || installPlanPath == "" {
			e2e.Failf("Fail to get its InstallPlan path(%s): %v", installPlanPath, err)
		}
		e2e.Logf(">>> InstallPlan path:%v", installPlanPath)
		jobName := fmt.Sprintf("%x", sha256.Sum256([]byte(installPlanPath)))[:63]

		g.By("3) check if the extract job uses the zip flag")
		// ["opm","alpha","bundle","extract","-m","/bundle/","-n", ns,"-c","9b59f03f8e8ea2f818061847881908aae51cf41836e4a3b822dcc6d3a01481c","-z"]
		extractCommand, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("job", "-n", ns, jobName, "-o=jsonpath={.spec.template.spec.containers[?(@.name==\"extract\")].command}").Output()
		if err != nil {
			e2e.Failf("Fail to get the jobs in the %s project: %v", ns, err)
		}
		if !strings.Contains(extractCommand, "-z") {
			e2e.Failf("This bundle extract job doesn't use the opm compression feature!")
		}

		g.By("4) check if the compression content is empty")
		// jobName == Configmap name
		bData, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("configmap", "-n", ns, jobName, "-o=jsonpath={.binaryData}").Output()
		if err != nil {
			e2e.Failf("Fail to get ConfigMap's binaryData: %v", err)
		}
		if bData == "" {
			e2e.Failf("The compression content is empty!")
		}
	})

	// Polarion ID: 43101
	g.It("PolarionID:43101-[OTP][Skipped:Disconnected]OLM blocks minor OpenShift upgrades when incompatible optional operators are installed", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		// consumes this index imaage: quay.io/olmqe/etcd-index:upgrade-auto, it contains the etcdoperator v0.9.2, v0.9.4, v0.9.5
		g.By("1, create a random project")
		oc.SetupProject()
		g.By("1-1, create a CatalogSource in this random project")
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-opm.yaml")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "cs-43101",
			Namespace:   oc.Namespace(),
			DisplayName: "OLM QE Operators",
			Publisher:   "Jian",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-index:upgrade-fips",
			Template:    csImageTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		defer cs.Delete(itName, dr)
		cs.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", cs.Name, "-n", oc.Namespace(), "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

		g.By("2, install the OperatorGroup in that random project")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-43101",
			Namespace: oc.Namespace(),
			Template:  ogSingleTemplate,
		}
		defer og.Delete(itName, dr)
		og.CreateWithCheck(oc, itName, dr)

		g.By("3, install the etcdoperator v0.9.2 with Manual approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-43101",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "cs-43101",
			CatalogSourceNamespace: oc.Namespace(),
			Channel:                "singlenamespace-alpha",
			IpApproval:             "Manual",
			OperatorPackage:        "etcd",
			StartingCSV:            "etcdoperator.v0.9.2",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		defer sub.Update(oc, itName, dr)
		sub.Create(oc, itName, dr)

		g.By("4, apprrove this etcdoperator.v0.9.2, it should be in Complete state")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.2", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.2", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		// olm.properties: '[{"type": "olm.maxOpenShiftVersion", "value": " "}]'
		g.By("5, this operator's olm.maxOpenShiftVersion is empty, so it should block the upgrade")
		olmv0util.CheckUpgradeStatus(oc, "False")

		g.By("6, apprrove this etcdoperator.v0.9.4, it should be in Complete state")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.4", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.4", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)
		// olm.properties: '[{"type": "olm.maxOpenShiftVersion", "value": "4.9"}]'
		g.By("7, 4.9.0-xxx upgraded to 4.10.0-xxx < 4.10.0, or 4.9.1 upgraded to 4.9.x < 4.10.0, so it should NOT block 4.9 upgrade, but block 4.10+ upgrade")
		currentVersion, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.status.desired.version}").Output()
		if err != nil {
			e2e.Failf("Fail to get the OCP version")
		}
		v, _ := semver.ParseTolerant(currentVersion)
		maxVersion, _ := semver.ParseTolerant("4.9")
		// current version > the operator's max version: 4.9
		if v.Compare(maxVersion) > 0 {
			olmv0util.CheckUpgradeStatus(oc, "False")
		} else {
			olmv0util.CheckUpgradeStatus(oc, "True")
		}

		g.By("8, apprrove this etcdoperator.v0.9.5, it should be in Complete state")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.5", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.5", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)
		// olm.properties: '[{"type": "olm.maxOpenShiftVersion", "value": "4.10.0"}]'
		g.By("9, 4.9.0-xxx upgraded to 4.10.0-xxx < 4.10.0, or 4.9.1 upgraded to 4.9.x < 4.11.0, so it should NOT block 4.10 upgrade, but blocks 4.11+ upgrade")
		maxVersion2, _ := semver.ParseTolerant("4.10.0")
		// current version > the operator's max version: 4.10.0
		if v.Compare(maxVersion2) > 0 {
			olmv0util.CheckUpgradeStatus(oc, "False")
		} else {
			olmv0util.CheckUpgradeStatus(oc, "True")
		}
	})

	// Polarion ID: 43977
	g.It("PolarionID:43977-[OTP]-OPENSHIFT_VERSIONS in assisted operator subscription does not propagate [Serial]", func() {
		// From 4.12, improve the ns permissions so that pod can be run successfully.
		// it is already privileged for default, so no need to set it.

		// this operator must be installed in the default project since the env variable: MY_POD_NAMESPACE = default
		g.By("1) create the OperatorGroup in the default project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-43977",
			Namespace: "default",
			Template:  ogSingleTemplate,
		}
		defer og.Delete(itName, dr)
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) subscribe to the learn-operator.v0.0.3 with ENV variables")
		subTemplate := filepath.Join(buildPruningBaseDir, "env-subscription.yaml")

		g.By("create the learn-operator CatalogSource")
		ns := "openshift-marketplace"
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-43977",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-43977",
			Namespace:              "default",
			CatalogSourceName:      "catsrc-43977",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer func() {
			if sub.InstalledCSV == "" {
				sub.FindInstalledCSV(oc, itName, dr)
			}
			sub.DeleteCSV(itName, dr)
		}()
		// the create method fails due that timeout, but some times csv is created, so need to delete them with defer if you do not delete ns.
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded"+"InstallSucceeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", "default", "-o=jsonpath={.status.phase}{.status.reason}"}).Check(oc)

		g.By("3) check those env variables")
		envVars := map[string]string{
			"MY_POD_NAMESPACE":        "default",
			"OPERATOR_CONDITION_NAME": "learn-operator.v0.0.3",
			"OPENSHIFT_VERSIONS":      "4.8",
		}
		// oc get deployment etcd-operator -o=jsonpath={.spec.template.spec.containers[0].env[?(@.name==\"MY_POD_NAMESPACE\")].value}
		// oc get deployment etcd-operator -o=jsonpath={.spec.template.spec.containers[0].env[?(@.name==\"OPERATOR_CONDITION_NAME\")].value}
		// oc get deployment etcd-operator -o=jsonpath={.spec.template.spec.containers[0].env[?(@.name==\"OPENSHIFT_VERSIONS\")].value}
		for k, v := range envVars {
			jsonpath := fmt.Sprintf("-o=jsonpath={.spec.template.spec.containers[0].env[?(@.name==\"%s\")].value}", k)
			envVar, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "learn-operator", "-n", "default", jsonpath).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if !strings.Contains(envVar, v) {
				e2e.Failf("The value of the %s should be %s, but get %s!", k, v, envVar)
			}
		}
	})

	// Polarion ID: 43978
	g.It("PolarionID:43978-[OTP]-Catalog pods don't report termination logs to catalog-operator", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		catalogs, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", "-n", "openshift-marketplace").Output()
		if err != nil {
			e2e.Failf("Fail to get the CatalogSource in openshift-marketplace project")
		}
		defaultCatalogs := []string{"certified-operators", "community-operators", "redhat-marketplace", "redhat-operators"}
		for i, catalog := range defaultCatalogs {
			g.By(fmt.Sprintf("%d) check CatalogSource: %s", i+1, catalog))
			if strings.Contains(catalogs, catalog) {
				policy, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", fmt.Sprintf("olm.catalogSource=%s", catalog), "-n", "openshift-marketplace", "-o=jsonpath={.items[0].spec.containers[0].terminationMessagePolicy}").Output()
				if err != nil {
					e2e.Failf("Fail to get the policy of the CatalogSource's pod")
				}
				if policy != "FallbackToLogsOnError" {
					e2e.Failf("CatalogSource:%s uses the %s policy, not the FallbackToLogsOnError!", catalog, policy)
				}
			} else {
				e2e.Logf("CatalogSource:%s doesn't install on this cluster", catalog)
			}
		}
	})

	// Polarion ID: 43803
	g.It("PolarionID:43803-[OTP]-Only one of multiple subscriptions to the same package is honored", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) create the OperatorGroup in a random project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		oc.SetupProject()
		ns := oc.Namespace()
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-43803",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) subscribe to the learn-operator.v0.0.3 with Automatic approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-43803",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-43803",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-43803",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)
		defer sub.DeleteCSV(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) re-subscribe to this learn operator with another subscription name")
		sub2 := olmv0util.SubscriptionDescription{
			SubName:                "sub2-43803",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "catsrc-43803",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub2.Delete(itName, dr)
		sub2.CreateWithoutCheck(oc, itName, dr)

		g.By("4) Check OLM logs")
		err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			logs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args("deploy/catalog-operator", "-n", "openshift-operator-lifecycle-manager").Output()
			if err != nil {
				e2e.Failf("Fail to get the OLM logs")
			}
			res, _ := regexp.MatchString(".*constraints not satisfiable.*subscription sub2-43803", logs)
			if res {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "subscription sub2-43803 constraints satisfiable")
	})

	// Polarion ID: 45411
	g.It("PolarionID:45411-[OTP]-packageserver isn't following the OpenShift HA conventions", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) get the cluster infrastructure")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster infra")
		}
		num, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", "openshift-operator-lifecycle-manager", "deployment", "packageserver", "-o=jsonpath={.status.replicas}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		if infra == "HighlyAvailable" {
			e2e.Logf("This is a HA cluster!")
			g.By("2) check if there are two packageserver pods")
			if num != "2" {
				e2e.Failf("!!!Fail, should have 2 packageserver pod, but get %s!", num)
			}
			g.By("3) check if the two packageserver pods running on different nodes")
			names, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=packageserver", "-o", "name").Output()
			if err != nil {
				e2e.Failf("Fail to get the Packageserver pods' name")
			}
			podNames := strings.Split(names, "\n")
			name := ""
			for _, podName := range podNames {
				e2e.Logf("get the packageserver pod Name: %s", podName)
				nodeName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", "openshift-operator-lifecycle-manager", podName, "-o=jsonpath={.spec.nodeName}").Output()
				if err != nil {
					e2e.Failf("Fail to get the node name")
				}
				e2e.Logf("get the node Name: %s", nodeName)
				if name != "" && name == nodeName {
					e2e.Failf("!!!Fail, the two packageserver pods running on the same node: %s!", nodeName)
				}
				name = nodeName
			}
		} else {
			e2e.Logf("This is a SNO cluster, skip!")
		}
	})

	// Polarion ID: 24028
	g.It("PolarionID:24028-[OTP]-need to set priorityClassName as system-cluster-critical", g.Label("NonHyperShiftHOST"), func() {
		var deploymentResource = [3]string{"catalog-operator", "olm-operator", "packageserver"}
		for _, v := range deploymentResource {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", "openshift-operator-lifecycle-manager", "deployment", v, "-o=jsonpath={.spec.template.spec.priorityClassName}").Output()
			e2e.Logf("%s.priorityClassName:%s", v, msg)
			if err != nil {
				e2e.Failf("Unable to get %s, error:%v", msg, err)
			}
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(msg).To(o.Equal("system-cluster-critical"))
		}
	})

	// Polarion ID: 21548
	g.It("PolarionID:21548-[OTP]-aggregates CR roles to standard admin view edit", func() {
		isExternalOIDCCluster, odcErr := exutil.IsExternalOIDCCluster(oc)
		o.Expect(odcErr).NotTo(o.HaveOccurred())
		if isExternalOIDCCluster {
			// https://github.com/openshift/release/pull/42250/files#diff-8f1e971323cb1821595fd1633ab701de55de169795027930c53aa5e736d7301dR38-R52
			g.Skip("Skipping the test as we are running against an external OIDC cluster, which the user has the cluster-admin role")
		}

		oc.SetupProject()
		msg, err := oc.Run("whoami").Args("").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("oc whoami: %s", msg)
		o.Expect(msg).NotTo(o.Equal("system:admin"))

		authorizations := []struct {
			resource string
			action   []string
			result   bool
		}{
			{
				resource: "subscriptions",
				action:   []string{"create", "update", "patch", "delete", "get", "list", "watch"},
				result:   true,
			},
			{
				resource: "installplans",
				action:   []string{"create", "update", "patch"},
				result:   false,
			},
			{
				resource: "installplans",
				action:   []string{"get", "list", "watch", "delete"},
				result:   true,
			},
			{
				resource: "catalogsources",
				action:   []string{"get", "list", "watch", "delete"},
				result:   true,
			},
			{
				resource: "catalogsources",
				action:   []string{"create", "update", "patch"},
				result:   false,
			},
			{
				resource: "clusterserviceversions",
				action:   []string{"get", "list", "watch", "delete"},
				result:   true,
			},
			{
				resource: "clusterserviceversions",
				action:   []string{"create", "update", "patch"},
				result:   false,
			},
			{
				resource: "operatorgroups",
				action:   []string{"get", "list", "watch"},
				result:   true,
			},
			{
				resource: "operatorgroups",
				action:   []string{"create", "update", "patch", "delete"},
				result:   false,
			},
			{
				resource: "packagemanifests",
				action:   []string{"get", "list", "watch"},
				result:   true,
			},
			// Based on https://github.com/openshift/operator-framework-olm/blob/master/staging/operator-lifecycle-manager/deploy/chart/templates/0000_50_olm_09-aggregated.clusterrole.yaml#L30
			// But, it returns '*', I will reseach it later.
			// $ oc get clusterrole admin -o yaml |grep packagemanifests -A5
			// - packagemanifests
			// verbs:
			// - '*'
			// {
			// 	resource: "packagemanifests",
			// 	action:   []string{"create", "update", "patch", "delete"},
			// 	result:   false,
			// },
		}

		for _, v := range authorizations {
			for _, act := range v.action {
				res, err := oc.Run("auth").Args("can-i", act, v.resource).Output()
				e2e.Logf("oc auth can-i %s %s", act, v.resource)
				if res != "no" && err != nil {
					o.Expect(err).NotTo(o.HaveOccurred())
				}
				if v.result {
					o.Expect(res).To(o.Equal("yes"))
				} else {
					o.Expect(res).To(o.Equal("no"))
				}
			}
		}
	})

	// Polarion ID: 37442
	g.It("PolarionID:37442-[OTP]-create a Conditions CR for each Operator it installs", func() {
		g.By("1) Install the OperatorGroup in a random project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		oc.SetupProject()
		ns := oc.Namespace()
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-37442",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Install the learn-operator v0.9.4 with Automatic approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-37442",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-37442",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-37442",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) Check if OperatorCondition generated well")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "learn-operator", exutil.Ok, []string{"operatorcondition", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.spec.deployments[0]}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "learn-operator.v0.0.3", exutil.Ok, []string{"deployment", "learn-operator", "-n", ns, "-o=jsonpath={.spec.template.spec.containers[*].env[?(@.name==\"OPERATOR_CONDITION_NAME\")].value}"}).Check(oc)
		// this learn-operator.v0.0.3 role should be owned by OperatorCondition
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "OperatorCondition", exutil.Ok, []string{"role", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.metadata.ownerReferences[0].kind}"}).Check(oc)
		// this learn-operator.v0.0.3 role should be added to learn-operator SA
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "learn-operator", exutil.Ok, []string{"rolebinding", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.subjects[0].name}"}).Check(oc)

		g.By("4) delete the operator so that can check the related resource in next step")
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)

		g.By("5) Check if the related resources are removed successfully")
		olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"operatorcondition", "learn-operator.v0.0.3", "-n", ns}).Check(oc)
		olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"role", "learn-operator.v0.0.3", "-n", ns}).Check(oc)
		olmv0util.NewCheck("present", exutil.AsAdmin, exutil.WithoutNamespace, exutil.NotPresent, "", exutil.Ok, []string{"rolebinding", "learn-operator.v0.0.3", "-n", ns}).Check(oc)

	})

	// Polarion ID: 37710
	g.It("PolarionID:37710-[OTP][Skipped:Disconnected]supports the Upgradeable Supported Condition", func() {
		g.By("1) Install the OperatorGroup in a random project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		oc.SetupProject()
		ns := oc.Namespace()
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-37710",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)
		g.By("2) Install the learn-operator.v0.0.1 with Manual approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-37710",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-37710",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-37710",
			CatalogSourceNamespace: ns,
			Channel:                "alpha",
			IpApproval:             "Manual",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.1",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		defer sub.Update(oc, itName, dr)
		sub.Create(oc, itName, dr)

		g.By("3) Apprrove this learn-operator.v0.0.1, it should be in Complete state")
		sub.ApproveSpecificIP(oc, itName, dr, "learn-operator.v0.0.1", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.1", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		// The conditions array will be added to OperatorCondition’s spec and operator is now expected to only update the conditions in the spec to reflect its condition
		// and no longer push changes to OperatorCondition’s status.
		// $oc patch operatorcondition learn-operator.v0.0.1 -p '{"spec":{"conditions":[{"type":"Upgradeable", "observedCondition":1,"status":"False","reason":"bug","message":"not ready","lastUpdateTime":"2021-06-16T16:56:44Z","lastTransitionTime":"2021-06-16T16:56:44Z"}]}}' --type=merge
		g.By("4) Patch the spec.conditions[0].Upgradeable to False")
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", oc.Namespace(), "operatorcondition", "learn-operator.v0.0.1", "-p", "{\"spec\": {\"conditions\": [{\"type\": \"Upgradeable\", \"status\": \"False\", \"reason\": \"upgradeIsNotSafe\", \"message\": \"Disable the upgrade\", \"observedCondition\":1, \"lastUpdateTime\":\"2021-06-16T16:56:44Z\",\"lastTransitionTime\":\"2021-06-16T16:56:44Z\"}]}}", "--type=merge")

		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Upgradeable", exutil.Ok, []string{"operatorcondition", "learn-operator.v0.0.1", "-n", ns, "-o=jsonpath={.status.conditions[0].type}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "False", exutil.Ok, []string{"operatorcondition", "learn-operator.v0.0.1", "-n", ns, "-o=jsonpath={.status.conditions[0].status}"}).Check(oc)

		g.By("5) Apprrove this learn-operator.v0.0.2, the corresponding CSV should be in Pending state")
		sub.ApproveSpecificIP(oc, itName, dr, "learn-operator.v0.0.2", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Pending", exutil.Ok, []string{"csv", "learn-operator.v0.0.2", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("6) Check the CSV message, the operator is not upgradeable")
		err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", ns, "csv", "learn-operator.v0.0.2", "-o=jsonpath={.status.message}").Output()
			if !strings.Contains(msg, "operator is not upgradeable") {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "learn-operator.v0.0.2 operator is upgradeable")

		g.By("7) Patch the spec.conditions[0].Upgradeable to True")
		// $oc patch operatorcondition learn-operator.v0.0.1 -p '{"spec":{"conditions":[{"type":"Upgradeable", "observedCondition":1,"status":"True","reason":"bug","message":"ready","lastUpdateTime":"2021-06-16T16:56:44Z","lastTransitionTime":"2021-06-16T16:56:44Z"}]}}' --type=merge
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", ns, "operatorcondition", "learn-operator.v0.0.1", "-p", "{\"spec\": {\"conditions\": [{\"type\": \"Upgradeable\", \"status\": \"True\", \"reason\": \"ready\", \"message\": \"enable the upgrade\", \"observedCondition\":1, \"lastUpdateTime\":\"2021-06-16T17:56:44Z\",\"lastTransitionTime\":\"2021-06-16T17:56:44Z\"}]}}", "--type=merge")
		g.By("8) the learn-operator.v0.0.1 can be upgraded to etcdoperator.v0.9.4 successfully")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.2", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	// Polarion ID: 37631
	g.It("PolarionID:37631-[OTP]-Allow cluster admin to overwrite the OperatorCondition", func() {
		g.By("1) Install the OperatorGroup in a random project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		oc.SetupProject()
		ns := oc.Namespace()
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-37631",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Install the learn-operator.v0.0.1 with Manual approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-37631",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-37631",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-37631",
			CatalogSourceNamespace: ns,
			Channel:                "alpha",
			IpApproval:             "Manual",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.1",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		defer sub.Update(oc, itName, dr)
		sub.Create(oc, itName, dr)

		g.By("3) Apprrove this learn-operator.v0.0.1, it should be in Complete state")
		sub.ApproveSpecificIP(oc, itName, dr, "learn-operator.v0.0.1", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.1", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("4) Patch the OperatorCondition to set the Upgradeable to False")
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", ns, "operatorcondition", "learn-operator.v0.0.1", "-p", "{\"spec\": {\"overrides\": [{\"type\": \"Upgradeable\", \"status\": \"False\", \"reason\": \"upgradeIsNotSafe\", \"message\": \"Disable the upgrade\"}]}}", "--type=merge")

		g.By("5) Apprrove this learn-operator.v0.0.2, the corresponding CSV should be in Pending state")
		sub.ApproveSpecificIP(oc, itName, dr, "learn-operator.v0.0.2", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Pending", exutil.Ok, []string{"csv", "learn-operator.v0.0.2", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("6) Check the CSV message, the operator is not upgradeable")
		err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", ns, "csv", "learn-operator.v0.0.2", "-o=jsonpath={.status.message}").Output()
			if !strings.Contains(msg, "operator is not upgradeable") {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "learn-operator.v0.0.2 operator is upgradeable")

		g.By("7) Change the Upgradeable of the OperatorCondition to True")
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", ns, "operatorcondition", "learn-operator.v0.0.1", "-p", "{\"spec\": {\"overrides\": [{\"type\": \"Upgradeable\", \"status\": \"True\", \"reason\": \"upgradeIsNotSafe\", \"message\": \"Disable the upgrade\"}]}}", "--type=merge")

		g.By("8) the learn-operator.v0.0.1 should be upgraded to learn-operator.v0.0.2 successfully")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.2", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)
	})

	// Polarion ID: 33450
	g.It("PolarionID:33450-[OTP][Skipped:Disconnected]Operator upgrades can delete existing CSV before completion", g.Label("NonHyperShiftHOST"), func() {
		architecture.SkipNonAmd64SingleArch(oc)
		g.By("1) Install a customization CatalogSource CR")
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-opm.yaml")
		cs := olmv0util.CatalogSourceDescription{
			Name:        "cs-33450",
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE Operators",
			Publisher:   "Jian",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/etcd-index:33450-fips",
			Template:    csImageTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		cs.Create(oc, itName, dr)
		defer cs.Delete(itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", cs.Name, "-n", "openshift-marketplace", "-o=jsonpath={.status..lastObservedState}"}).Check(oc)

		g.By("2) Subscribe to the etcd operator with Manual approval")
		oc.SetupProject()
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")

		og := olmv0util.OperatorGroupDescription{
			Name:      "og-33450",
			Namespace: oc.Namespace(),
			Template:  ogSingleTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-33450",
			Namespace:              oc.Namespace(),
			CatalogSourceName:      "cs-33450",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "alpha",
			IpApproval:             "Manual",
			OperatorPackage:        "etcd",
			StartingCSV:            "etcdoperator.v0.9.2",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		sub.Create(oc, itName, dr)
		g.By("3) Apprrove the etcdoperator.v0.9.2, it should be in Complete state")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.2", "Complete")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "etcdoperator.v0.9.2", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("4) Apprrove the etcdoperator.v0.9.4, it should be in Failed state")
		sub.ApproveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.4", "Failed")

		g.By("5) The etcdoperator.v0.9.4 CSV should be in Pending status")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Pending", exutil.Ok, []string{"csv", "etcdoperator.v0.9.4", "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("6) The SA should be owned by the etcdoperator.v0.9.2")
		err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 10*time.Second, false, func(ctx context.Context) (bool, error) {
			saOwner := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sa", "etcd-operator", "-n", sub.Namespace, "-o=jsonpath={.metadata.ownerReferences[0].name}")
			if strings.Compare(saOwner, "etcdoperator.v0.9.2") != 0 {
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "sa etcd-operator owner is not etcdoperator.v0.9.2")

	})

	// Polarion ID: 37260
	g.It("PolarionID:37260-[OTP][Skipped:Disconnected]should allow to create the default CatalogSource [Disruptive]", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		g.By("1) Disable the default OperatorHub")
		olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operatorhub", "cluster", "-p", "{\"spec\": {\"disableAllDefaultSources\": true}}", "--type=merge")
		defer olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operatorhub", "cluster", "-p", "{\"spec\": {\"disableAllDefaultSources\": false}}", "--type=merge")
		g.By("1-1) Check if the default CatalogSource resource are removed")
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			res, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", "redhat-operators", "-n", "openshift-marketplace").Output()
			if strings.Contains(res, "not found") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "redhat-operators found")

		g.By("2) Create a CatalogSource with a default CatalogSource name")
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		// ocpVersionByte, err := exec.Command("bash", "-c", "oc version -o json | jq -r '.openshiftVersion' | cut -d '.' -f1,2").Output()
		// o.Expect(err).NotTo(o.HaveOccurred())
		// ocpVersion := strings.ReplaceAll(string(ocpVersionByte), "\n", "")
		// indexImage := fmt.Sprintf("quay.io/openshift-qe-optional-operators/aosqe-index:v%s", ocpVersion)
		indexImage := "quay.io/olmqe/learn-operator-index:v25"
		oc.SetupProject()
		cs := olmv0util.CatalogSourceDescription{
			Name:        "redhat-operators",
			Namespace:   "openshift-marketplace",
			DisplayName: "OLM QE",
			Publisher:   "OLM QE",
			SourceType:  "grpc",
			Address:     indexImage,
			Template:    csImageTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
		cs.CreateWithCheck(oc, itName, dr)
		g.By("2-1) Check if this custom CatalogSource resource works well")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			res, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest").Output()
			if strings.Contains(res, "OLM QE") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "packagemanifest does not exutil.Contain OLM QE")
		g.By("3) Delete the Marketplace pods and check if the custome CatalogSource still works well")
		_, err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("pods", "-l", "name=marketplace-operator", "-n", "openshift-marketplace").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		// waiting for the new marketplace pod ready
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			res, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "name=marketplace-operator", "-o=jsonpath={.items..status.phase}", "-n", "openshift-marketplace").Output()
			if strings.Contains(res, "Running") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "marketplace-operator pod is not running")
		g.By("3-3) check if the custom CatalogSource still there")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status..lastObservedState}"}).Check(oc)
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			res, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest").Output()
			if strings.Contains(res, "OLM QE") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "packagemanifest does not exutil.Contain OLM QE")

		g.By("4) Enable the default OperatorHub")
		olmv0util.PatchResource(oc, true, true, "operatorhub", "cluster", "-p", "{\"spec\": {\"disableAllDefaultSources\": false}}", "--type=merge")
		g.By("4-1) Check if the default CatalogSource resource are back")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", "redhat-operators", "-n", "openshift-marketplace", "-o=jsonpath={.status..lastObservedState}"}).Check(oc)
		g.By("4-2) Check if the default CatalogSource works and the custom one are removed")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			res, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest").Output()
			if strings.Contains(res, "Red Hat Operators") && !strings.Contains(res, "OLM QE") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "packagemanifest does exutil.Contain OLM QE or has no Red Hat Operators")
	})

	// Polarion ID: 25922
	g.It("PolarionID:25922-[OTP]-Support spec.config.volumes and volumemount in Subscription", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		oc.SetupProject()
		ns := oc.Namespace()
		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-25922",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-25922",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By(fmt.Sprintf("1) create the OperatorGroup in project: %s", ns))
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) install learn-operator.v0.0.3")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-25922",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-25922",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		defer sub.Update(oc, itName, dr)
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) create a ConfigMap")
		cmTemplate := filepath.Join(buildPruningBaseDir, "cm-template.yaml")

		cm := olmv0util.ConfigMapDescription{
			Name:      "special-config",
			Namespace: ns,
			Template:  cmTemplate,
		}
		cm.Create(oc, itName, dr)

		g.By("4) Patch this ConfigMap a volume")
		sub.Patch(oc, fmt.Sprintf("{\"spec\": {\"channel\":\"alpha\",\"config\":{\"volumeMounts\":[{\"mountPath\":\"/test\",\"name\":\"config-volume\"}],\"volumes\":[{\"configMap\":{\"name\":\"special-config\"},\"name\":\"config-volume\"}]},\"name\":\"learn\",\"source\":\"catsrc-25922\",\"sourceNamespace\":\"%s\"}}", ns))
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			podName, err := oc.AsAdmin().Run("get").Args("-n", ns, "pods", "-l", "name=learn-operator", "-o=jsonpath={.items[0].metadata.name}").Output()
			if err != nil {
				return false, err
			}
			e2e.Logf("4-1) Get learn operator pod Name:%s", podName)
			result, _ := oc.AsAdmin().Run("exec").Args("-n", ns, podName, "--", "cat", "/test/special.how").Output()
			e2e.Logf("4-2) Check if the ConfigMap mount well")
			if strings.Contains(result, "very") {
				e2e.Logf("4-3) The ConfigMap: special-config mount well")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "pod of learn-operator-alm-owned special-config not mount well")
		g.By("5) Patch a non-exist volume")
		sub.Patch(oc, fmt.Sprintf("{\"spec\":{\"channel\":\"alpha\",\"config\":{\"volumeMounts\":[{\"mountPath\":\"/test\",\"name\":\"volume1\"}],\"volumes\":[{\"persistentVolumeClaim\":{\"claimName\":\"claim1\"},\"name\":\"volume1\"}]},\"name\":\"learn\",\"source\":\"catsrc-25922\",\"sourceNamespace\":\"%s\"}}", ns))
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			for i := 0; i < 2; i++ {
				g.By("5-1) Check the pods status")
				podStatus, err := oc.AsAdmin().Run("get").Args("-n", ns, "pods", "-l", "name=learn-operator", fmt.Sprintf("-o=jsonpath={.items[%d].status.phase}", i)).Output()
				if err != nil {
					return false, err
				}
				if podStatus == "Pending" {
					g.By("5-2) The pod status is Pending as expected")
					return true, nil
				}
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "pod of learn-operator-alm-owned status is not Pending")
	})

	// Polarion ID: 35631
	g.It("PolarionID:35631-[OTP]-Remove OperatorSource API", func() {
		exutil.SkipBaselineCaps(oc, "None")
		g.By("1) Check the operatorsource resource")
		msg, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("operatorsource").Output()
		e2e.Logf("Get the expected error: %s", msg)
		o.Expect(msg).To(o.ContainSubstring("the server doesn't have a resource type"))

		// for current disconnected env, only have the default community CatalogSource CRs
		g.By("2) Check the default Community CatalogSource CRs")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", "-n", "openshift-marketplace").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("Get the installed CatalogSource CRs:\n %s", msg)
		o.Expect(msg).To(o.ContainSubstring("grpc"))
		// o.Expect(msg).To(o.ContainSubstring("certified-operators"))
		// o.Expect(msg).To(o.ContainSubstring("community-operators"))
		// o.Expect(msg).To(o.ContainSubstring("redhat-marketplace"))
		// o.Expect(msg).To(o.ContainSubstring("redhat-operators"))
		g.By("3) Check the Packagemanifest")
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "-n", "openshift-marketplace").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).NotTo(o.ContainSubstring("No resources found"))
	})

	// Polarion ID: 33902
	g.It("PolarionID:33902-[OTP][Skipped:Disconnected]Catalog Weighting", func() {
		architecture.SkipNonAmd64SingleArch(oc)
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

		oc.SetupProject()
		ns := oc.Namespace()

		// the priority ranking is bucket-test1 > bucket-test2 > community-operators(-400 default)
		csObjects := []struct {
			Name     string
			Address  string
			Priority int
		}{
			{"ocs-cs", "quay.io/olmqe/ocs-index:4.3.0", 0},
			{"bucket-test1", "quay.io/olmqe/bucket-index:1.0.0", 20},
			{"bucket-test2", "quay.io/olmqe/bucket-index:1.0.0", -1},
		}

		// create the OperatorGroup resource
		og := olmv0util.OperatorGroupDescription{
			Name:      "test-og-33902",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}

		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		defer func() {
			for _, v := range csObjects {
				g.By(fmt.Sprintf("9) Remove the %s CatalogSource", v.Name))
				cs := olmv0util.CatalogSourceDescription{
					Name:        v.Name,
					Namespace:   "openshift-marketplace",
					DisplayName: "Priority Test",
					Publisher:   "OLM QE",
					SourceType:  "grpc",
					Address:     v.Address,
					Template:    csImageTemplate,
					Priority:    v.Priority,
				}
				cs.Delete(itName, dr)
			}
		}()

		for i, v := range csObjects {
			g.By(fmt.Sprintf("%d) start to create the %s CatalogSource", i+1, v.Name))
			cs := olmv0util.CatalogSourceDescription{
				Name:        v.Name,
				Namespace:   "openshift-marketplace",
				DisplayName: "Priority Test",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     v.Address,
				Template:    csImageTemplate,
				Priority:    v.Priority,
			}
			cs.Create(oc, itName, dr)
			olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "READY", exutil.Ok, []string{"catsrc", cs.Name, "-n", cs.Namespace, "-o=jsonpath={.status.connectionState.lastObservedState}"}).Check(oc)
		}

		g.By("4) create the OperatorGroup")
		og.CreateWithCheck(oc, itName, dr)

		g.By("5) start to subscribe to the OCS operator")
		sub := olmv0util.SubscriptionDescription{
			SubName:                "ocs-sub",
			Namespace:              ns,
			CatalogSourceName:      "ocs-cs",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "4.3.0",
			IpApproval:             "Automatic",
			OperatorPackage:        "ocs-operator",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		sub.Create(oc, itName, dr)

		g.By("6) check the dependce operator's subscription")
		depSub := olmv0util.SubscriptionDescription{
			SubName:                "lib-bucket-provisioner-4.3.0-bucket-test1-openshift-marketplace",
			Namespace:              ns,
			CatalogSourceName:      "bucket-test1",
			CatalogSourceNamespace: "openshift-marketplace",
			Channel:                "4.3.0",
			IpApproval:             "Automatic",
			OperatorPackage:        "lib-bucket-provisioner",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		// The dependence is lib-bucket-provisioner-4.3.0, it should from the bucket-test1 CatalogSource since its priority is the highest.
		dr.GetIr(itName).Add(olmv0util.NewResource(oc, "sub", depSub.SubName, true, depSub.Namespace))
		depSub.FindInstalledCSV(oc, itName, dr)

		g.By(fmt.Sprintf("7) Remove subscription:%s, %s", sub.SubName, depSub.SubName))
		sub.Delete(itName, dr)
		sub.DeleteCSV(itName, dr)
		depSub.Delete(itName, dr)
		depSub.GetCSV().Delete(itName, dr)

	})

	// Polarion ID: 32559
	g.It("PolarionID:32559-[OTP]-catalog operator crashed", g.Label("NonHyperShiftHOST"), func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		csImageTemplate := filepath.Join(buildPruningBaseDir, "cs-without-image.yaml")
		oc.SetupProject()
		namespace := oc.Namespace()
		csTypes := []struct {
			name        string
			csType      string
			expectedMSG string
		}{
			{"cs-noimage", "grpc", "image and address unset"},
			{"cs-noimage-cm", "configmap", "configmap name unset"},
		}
		for _, t := range csTypes {
			g.By(fmt.Sprintf("test the %s type CatalogSource", t.csType))
			cs := olmv0util.CatalogSourceDescription{
				Name:        t.name,
				Namespace:   namespace,
				DisplayName: "OLM QE",
				Publisher:   "OLM QE",
				SourceType:  t.csType,
				Template:    csImageTemplate,
			}
			dr := make(olmv0util.DescriberResrouce)
			itName := g.CurrentSpecReport().FullText()
			dr.AddIr(itName)
			cs.Create(oc, itName, dr)

			err := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
				output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", namespace, "catalogsource", cs.Name, "-o=jsonpath={.status.message}").Output()
				if err != nil {
					e2e.Logf("Fail to get CatalogSource: %s, error: %s and try again", cs.Name, err)
					return false, nil
				}
				if strings.Contains(output, t.expectedMSG) {
					e2e.Logf("Get expected message: %s", t.expectedMSG)
					return true, nil
				}
				return false, nil
			})

			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("catsrc of %s does not exutil.Contain %v", namespace, t.expectedMSG))

			status, err := oc.AsAdmin().Run("get").Args("-n", "openshift-operator-lifecycle-manager", "pods", "-l", "app=catalog-operator", "-o=jsonpath={.items[0].status.phase}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if status != "Running" {
				e2e.Failf("The status of the CatalogSource: %s pod is: %s", cs.Name, status)
			}
		}

		// destroy the two CatalogSource CRs
		for _, t := range csTypes {
			_, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("-n", namespace, "catalogsource", t.name).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
		}
	})

	// Polarion ID: 22070
	g.It("PolarionID:22070-[Level0][OTP][Skipped:Disconnected]support grpc sourcetype [Serial]", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("1) Start to subscribe the learn operator")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

		g.By("create the learn-operator CatalogSource")
		ns := "openshift-marketplace"
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-22070",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-22070",
			Namespace:              "openshift-operators",
			CatalogSourceName:      "catsrc-22070",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Automatic",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        false,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		defer func() {
			if sub.InstalledCSV == "" {
				sub.FindInstalledCSV(oc, itName, dr)
			}
			sub.DeleteCSV(itName, dr)
		}()
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"-n", "openshift-operators", "csv", sub.InstalledCSV, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("3) Assert that learn operator dependency is resolved")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", ns).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).To(o.ContainSubstring("learn-operator.v0.0.3"))
	})

	// Polarion ID: 20981
	g.It("PolarionID:20981-[OTP]-contain the source commit id", g.Label("NonHyperShiftHOST"), func() {
		if os.Getenv("GITHUB_TOKEN") == "" {
			g.Skip("Skip it since no GITHUB_TOKEN configured")
		}

		sameCommit := ""
		subPods := []string{"catalog-operator", "olm-operator", "packageserver"}

		for _, v := range subPods {
			podName, err := oc.AsAdmin().Run("get").Args("-n", "openshift-operator-lifecycle-manager", "pods", "-l", fmt.Sprintf("app=%s", v), "-o=jsonpath={.items[0].metadata.name}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			e2e.Logf("get pod Name:%s", podName)

			g.By(fmt.Sprintf("get olm version from the %s pod", v))
			commands := []string{"-n", "openshift-operator-lifecycle-manager", "exec", podName, "--", "olm", "--version"}
			olmVersion, err := oc.AsAdmin().Run(commands...).Args().Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			idSlice := strings.Split(olmVersion, ":")
			gitCommitID := strings.TrimSpace(idSlice[len(idSlice)-1])
			e2e.Logf("olm source git commit ID:%s", gitCommitID)
			if len(gitCommitID) != 40 {
				e2e.Failf("the length of the git commit id is %d, != 40", len(gitCommitID))
			}

			if sameCommit == "" {
				sameCommit = gitCommitID
				g.By("checking this commitID in https://github.com/openshift/operator-framework-olm repo")
				// TODO: Uncomment after adding github package to vendor
				// ctx, tc := olmv0util.GithubClient()
				// client := github.NewClient(tc)
				// // OLM downstream repo has been changed to: https://github.com/openshift/operator-framework-olm
				// _, _, err := client.Git.GetCommit(ctx, "openshift", "operator-framework-olm", gitCommitID)
				// if err != nil {
				// 	e2e.Failf("Git.GetCommit returned error: %v", err)
				// }
				e2e.Logf("Skipping GitHub API check for commit %s (requires github package in vendor)", gitCommitID)

			} else if gitCommitID != sameCommit {
				e2e.Failf("These commitIDs inconformity!!!")
			}
		}
	})

	// Polarion ID: 21126
	g.It("PolarionID:21126-[OTP]-OLM Subscription status says CSV is installed when it is not", func() {
		g.By("1) Install the OperatorGroup in a random project")
		dr := make(olmv0util.DescriberResrouce)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		oc.SetupProject()
		ns := oc.Namespace()
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-21126",
			Namespace: ns,
			Template:  ogSingleTemplate,
		}
		og.CreateWithCheck(oc, itName, dr)

		g.By("2) Install learn-operator.v0.0.3 with Manual approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-21126",
			Namespace:   ns,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-21126",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-21126",
			CatalogSourceNamespace: ns,
			Channel:                "beta",
			IpApproval:             "Manual",
			OperatorPackage:        "learn",
			StartingCSV:            "learn-operator.v0.0.3",
			SingleNamespace:        true,
			Template:               subTemplate,
		}
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)
		g.By("3) Check the learn-operator.v0.0.3 related resources")
		// the installedCSV should be NULL
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "", exutil.Ok, []string{"sub", "sub-21126", "-n", ns, "-o=jsonpath={.status.InstalledCSV}"}).Check(oc)
		// the state should be UpgradePending
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "UpgradePending", exutil.Ok, []string{"sub", "sub-21126", "-n", ns, "-o=jsonpath={.status.state}"}).Check(oc)
		// the InstallPlan should not approved
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "false", exutil.Ok, []string{"installplan", sub.GetIP(oc), "-n", ns, "-o=jsonpath={.spec.approved}"}).Check(oc)
		// should no etcdoperator.v0.9.4 CSV found
		msg, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "learn-operator.v0.0.3", "-n", ns).Output()
		if !strings.Contains(msg, "not found") {
			e2e.Failf("still found the learn-operator.v0.0.3 in Namespace:%s, msg:%v", ns, msg)
		}
	})

})
