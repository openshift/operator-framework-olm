package specs

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0", func() {
	g.It("should pass a trivial sanity check", g.Label("ReleaseGate"), func(ctx context.Context) {
		o.Expect(len("test")).To(o.BeNumerically(">", 0))
	})
})

// it is mapping to the Describe "OLM for an end user handle common object" and "OLM for an end user use" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 should", func() {
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

	g.It("PolarionID:22259-[OTP][Skipped:Disconnected]marketplace operator CR status on a running cluster[Serial]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:22259-[Skipped:Disconnected]marketplace operator CR status on a running cluster[Serial]"), func() {

		exutil.SkipForSNOCluster(oc)
		exutil.SkipNoCapabilities(oc, "marketplace")
		g.By("check marketplace status")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "TrueFalseFalse", exutil.Ok, []string{"clusteroperator", "marketplace",
			"-o=jsonpath={.status.conditions[?(@.type==\"Available\")].status}{.status.conditions[?(@.type==\"Progressing\")].status}{.status.conditions[?(@.type==\"Degraded\")].status}"}).Check(oc)

		g.By("delete pod of marketplace")
		_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "pod", "--selector=name=marketplace-operator",
			"-n", "openshift-marketplace", "--ignore-not-found")
		o.Expect(err).NotTo(o.HaveOccurred())

		_, _ = exec.Command("bash", "-c", "sleep 10").Output()

		g.By("pod of marketplace restart")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "TrueFalseFalse", exutil.Ok, []string{"clusteroperator", "marketplace",
			"-o=jsonpath={.status.conditions[?(@.type==\"Available\")].status}{.status.conditions[?(@.type==\"Progressing\")].status}{.status.conditions[?(@.type==\"Degraded\")].status}"}).Check(oc)
	})

	g.It("PolarionID:32862-[OTP][Skipped:Disconnected]Pods found with invalid container images not present in release payload", func() {

		exutil.SkipBaselineCaps(oc, "None")
		g.By("Verify the version of marketplace_operator")
		pods, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", "openshift-marketplace", "--no-headers").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		lines := strings.Split(pods, "\n")
		for _, line := range lines {
			e2e.Logf("line: %v", line)
			if strings.Contains(line, "certified-operators") || strings.Contains(line, "community-operators") || strings.Contains(line, "marketplace-operator") || strings.Contains(line, "redhat-marketplace") || strings.Contains(line, "redhat-operators") && strings.Contains(line, "1/1") {
				name := strings.Split(line, " ")
				checkRel, err := oc.AsAdmin().WithoutNamespace().Run("exec").Args(name[0], "-n", "openshift-marketplace", "--", "cat", "/etc/redhat-release").Output()
				if err != nil {
					e2e.Logf("can not get content with error %v, and try next", err)
					continue
				}
				o.Expect(checkRel).To(o.ContainSubstring("Red Hat"))
			}
		}

	})

	g.It("PolarionID:42041-[OTP]Available=False despite unavailableReplicas <= maxUnavailable", g.Label("NonHyperShiftHOST"), func() {
		g.By("get the cluster infrastructure")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster infra")
		}
		if infra == "SingleReplica" {
			e2e.Logf("This is a SNO cluster")
			maxUnavailable, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.strategy.rollingUpdate.maxUnavailable}").Output()
			e2e.Logf("%s", maxUnavailable)
			o.Expect(err1).NotTo(o.HaveOccurred())
			o.Expect(maxUnavailable).NotTo(o.BeEmpty())

			maxSurge, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.strategy.rollingUpdate.maxSurge}").Output()
			e2e.Logf("%s", maxSurge)
			o.Expect(err1).NotTo(o.HaveOccurred())
			o.Expect(maxSurge).NotTo(o.BeEmpty())

		} else {
			maxUnavailableInCsv, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={..install.spec.deployments[0].spec.strategy.rollingUpdate.maxUnavailable}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(maxUnavailableInCsv).NotTo(o.BeEmpty())
			maxSurgeInCsv, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={..install.spec.deployments[0].spec.strategy.rollingUpdate.maxSurge}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(maxSurgeInCsv).NotTo(o.BeEmpty())

			_, err1 := oc.AsAdmin().WithoutNamespace().Run("patch").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager",
				"--type=json", "--patch", "[{\"op\": \"add\",\"path\": \"/spec/install/spec/deployments/0/spec/template/metadata/annotations\", \"value\": { \"custom.csv\": \"custom csv value\"} }]").Output()
			o.Expect(err1).NotTo(o.HaveOccurred())

			maxUnavailable, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.strategy.rollingUpdate.maxUnavailable}").Output()
			e2e.Logf("%s", maxUnavailable)
			o.Expect(err1).NotTo(o.HaveOccurred())
			o.Expect(maxUnavailable).To(o.Equal(maxUnavailableInCsv))

			maxSurge, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec.strategy.rollingUpdate.maxSurge}").Output()
			e2e.Logf("%s", maxSurge)
			o.Expect(err1).NotTo(o.HaveOccurred())
			o.Expect(maxSurge).To(o.Equal(maxSurgeInCsv))
		}
	})

	g.It("PolarionID:42068-[OTP]Available condition set to false on any Deployment spec change", g.Label("NonHyperShiftHOST"), func() {
		available, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusteroperator", "operator-lifecycle-manager-packageserver", "-o=jsonpath={.status.conditions[1].type}").Output()
		e2e.Logf("%s", available)
		o.Expect(err1).NotTo(o.HaveOccurred())
		o.Expect(available).To(o.Equal("Available"))

		statusAvailable, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusteroperator", "operator-lifecycle-manager-packageserver", "-o=jsonpath={.status.conditions[1].status}").Output()
		e2e.Logf("%s", statusAvailable)
		o.Expect(err1).NotTo(o.HaveOccurred())
		o.Expect(statusAvailable).To(o.Equal("True"))
	})

	g.It("PolarionID:42073-[OTP]deployment sets neither CPU or memory request on the packageserver container", g.Label("NonHyperShiftHOST"), func() {
		cpu, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={..containers..resources.requests.cpu}").Output()
		e2e.Logf("%s", cpu)
		o.Expect(err1).NotTo(o.HaveOccurred())
		o.Expect(cpu).NotTo(o.Equal(""))

		memory, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={..containers..resources.requests.memory}").Output()
		e2e.Logf("%s", memory)
		o.Expect(err1).NotTo(o.HaveOccurred())
		o.Expect(memory).NotTo(o.Equal(""))

		catPodnames, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "--selector=app=packageserver", "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(catPodnames).NotTo(o.BeEmpty())

		lines := strings.Split(catPodnames, " ")
		for _, line := range lines {
			e2e.Logf("line: %v", line)

			pkg1Cpu, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", line, "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec..resources.requests.cpu}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(pkg1Cpu).To(o.Equal(cpu))

			pkg1Memory, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", line, "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.spec..resources.requests.memory}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(pkg1Memory).To(o.Equal(memory))
		}
	})

	g.It("PolarionID:41283-[OTP][Skipped:Disconnected]Marketplace extract container request CPU or memory", func() {
		var buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
		var subFile = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
		var ogFile = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		var operatorWait = 150 * time.Second

		namespace := oc.Namespace()
		itName := g.CurrentSpecReport().FullText()

		og := olmv0util.OperatorGroupDescription{
			Name:      "test-operators-og",
			Namespace: namespace,
			Template:  ogFile,
		}
		defer og.Delete(itName, dr)
		og.Create(oc, itName, dr)

		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-41283",
			Namespace:   namespace,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("Verify inside the jobs the value of spec.containers[].resources.requests field are setted")

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-41283",
			Namespace:              namespace,
			CatalogSourceName:      "catsrc-41283",
			CatalogSourceNamespace: namespace,
			IpApproval:             "Automatic",
			Channel:                "beta",
			OperatorPackage:        "learn",
			SingleNamespace:        true,
			Template:               subFile,
		}
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)
		sub.Create(oc, itName, dr)

		err := wait.PollUntilContextTimeout(context.TODO(), 60*time.Second, operatorWait, false, func(ctx context.Context) (bool, error) {
			checknameCsv, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("jobs", "-n", namespace, "-o", "jsonpath={.items[*].spec.template.spec.containers[*].resources.requests.cpu}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			e2e.Logf("%s", checknameCsv)
			if checknameCsv == "" {
				e2e.Logf("jobs KO Limit not setted ")
				return false, nil
			}
			e2e.Logf("jobs OK Limit setted ")
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "jobs KO Limit not setted")

	})

	g.It("PolarionID:23395-[OTP][Skipped:Disconnected]Deleted catalog registry pods and verify if them are recreated automatically[Disruptive]", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		g.By("delete pod of catsrc redhat-operators")
		_, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("pod", "--selector=olm.catalogSource=redhat-operators", "-n", "openshift-marketplace").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			res, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "--selector=olm.catalogSource=redhat-operators", "-o=jsonpath={.items..status.phase}", "-n", "openshift-marketplace").Output()
			if strings.Contains(res, "Running") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "fails to get pod of redhat-operators")
	})

	g.It("PolarionID:42069-[OTP][Skipped:Disconnected]component not found log should be debug level", g.Label("NonHyperShiftHOST"), func() {
		var since = "--since=60s"
		var snooze time.Duration = 90
		var tail = "--tail=10"

		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		itName := g.CurrentSpecReport().FullText()
		namespace := oc.Namespace()

		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-42069",
			Namespace:   namespace,
			DisplayName: "QE Operators",
			Publisher:   "OpenShift QE",
			SourceType:  "grpc",
			Address:     "quay.io/olmqe/learn-operator-index:v25",
			Template:    catsrcImageTemplate,
		}
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("1) Install the OperatorGroup in a random project")

		ogSingleTemplate := filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
		og := olmv0util.OperatorGroupDescription{
			Name:      "og-42069",
			Namespace: namespace,
			Template:  ogSingleTemplate,
		}
		defer og.Delete(itName, dr)
		og.Create(oc, itName, dr)

		g.By("2) Install the learn-operator with Automatic approval")
		subTemplate := filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

		sub := olmv0util.SubscriptionDescription{
			SubName:                "sub-42069",
			Namespace:              namespace,
			CatalogSourceName:      "catsrc-42069",
			CatalogSourceNamespace: namespace,
			IpApproval:             "Automatic",
			Channel:                "beta",
			OperatorPackage:        "learn",
			SingleNamespace:        true,
			Template:               subTemplate,
		}

		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", oc.Namespace(), "-o=jsonpath={.status.phase}"}).Check(oc)
		defer sub.Delete(itName, dr)
		defer sub.DeleteCSV(itName, dr)

		nameIP := sub.GetIP(oc)
		deteleIP, err1 := oc.AsAdmin().WithoutNamespace().Run("delete").Args("installplan", nameIP, "-n", oc.Namespace()).Output()
		e2e.Logf("%s", deteleIP)
		o.Expect(err1).NotTo(o.HaveOccurred())
		o.Expect(deteleIP).To(o.ContainSubstring("deleted"))

		catPodname, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "--selector=app=olm-operator", "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(catPodname).NotTo(o.BeEmpty())

		waitErr := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, snooze*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args(catPodname, "-n", "openshift-operator-lifecycle-manager", tail, since).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, "component not found") {
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollWithErr(waitErr, "log 'component not found' is not debug level")

	})

	g.It("PolarionID:21534-[OTP][Skipped:Disconnected]Check OperatorGroups on console", func() {
		ogNamespace, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("og", "global-operators", "-n", "openshift-operators", "-o", "jsonpath={.status.namespaces}").Output()
		e2e.Logf("%s", ogNamespace)
		o.Expect(err1).NotTo(o.HaveOccurred())
		o.Expect(ogNamespace).To(o.Equal("[\"\"]"))

	})

	g.It("PolarionID:24058-[OTP]components should have resource limits defined", func() {
		olmUnlimited := 0
		olmNames := []string{""}
		olmNamespace := "openshift-operator-lifecycle-manager"
		olmJpath := "-o=jsonpath={range .items[*]}{@.metadata.name}{','}{@.spec.containers[0].resources.requests.*}{'\\n'}"
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", olmNamespace, olmJpath).Output()
		if err != nil {
			e2e.Failf("Unable to get pod -n %v %v.", olmNamespace, olmJpath)
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).NotTo(o.ContainSubstring("No resources found"))
		lines := strings.Split(msg, "\n")
		for _, line := range lines {
			name := strings.Split(line, ",")
			// e2e.Logf("Line is %v, len %v, len name %v, name0 %v, name1 %v\n", line, len(line), len(name), name[0], name[1])
			if strings.Contains(line, "packageserver") {
				continue
			} else {
				if len(line) > 1 {
					if len(name) > 1 && len(name[1]) < 1 {
						olmUnlimited++
						olmNames = append(olmNames, name[0])
					}
				}
			}
		}
		if olmUnlimited > 0 && len(olmNames) > 0 {
			e2e.Failf("There are no limits set on %v of %v OLM components: %v", olmUnlimited, len(lines), olmNames)
		}
	})

	g.It("PolarionID:25782-[OTP][Skipped:Disconnected]CatalogSource Status should have information on last observed state", func() {
		var (
			catName             = "installed-community-25782-global-operators"
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			// the namespace and catName are hardcoded in the files
			cmTemplate       = filepath.Join(buildPruningBaseDir, "cm-csv-etcd.yaml")
			catsrcCmTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
		)

		itName := g.CurrentSpecReport().FullText()

		var (
			cm = olmv0util.ConfigMapDescription{
				Name:      catName,
				Namespace: oc.Namespace(),
				Template:  cmTemplate,
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        catName,
				Namespace:   oc.Namespace(),
				DisplayName: "Community bad Operators",
				Publisher:   "QE",
				SourceType:  "configmap",
				Address:     catName,
				Template:    catsrcCmTemplate,
			}
		)

		g.By("Create ConfigMap with bad operator manifest")
		defer cm.Delete(itName, dr)
		cm.Create(oc, itName, dr)

		// Make sure bad configmap was created
		g.By("Check configmap")
		msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("cm", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(strings.Contains(msg, catName)).To(o.BeTrue())

		g.By("Create catalog source")
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("Wait for pod to fail")
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", oc.Namespace()).Output()
			e2e.Logf("\n%v", msg)
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, "CrashLoopBackOff") {
				e2e.Logf("STEP pod is in  CrashLoopBackOff as expected")
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(waitErr, "the pod is not in CrashLoopBackOff")

		g.By("Check catsrc state for TRANSIENT_FAILURE in lastObservedState")
		waitErr = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
			msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", catName, "-n", oc.Namespace(), "-o=jsonpath={.status}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			if strings.Contains(msg, "TRANSIENT_FAILURE") && strings.Contains(msg, "lastObservedState") {
				msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("catalogsource", catName, "-n", oc.Namespace(), "-o=jsonpath={.status.connectionState.lastObservedState}").Output()
				o.Expect(err).NotTo(o.HaveOccurred())
				e2e.Logf("catalogsource had lastObservedState =  %v as expected ", msg)
				return true, nil
			}
			return false, nil
		})
		exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("catalogsource %s is not TRANSIENT_FAILURE", catName))
		e2e.Logf("cleaning up")
	})

	g.It("PolarionID:73695-[OTP][Skipped:Disconnected]PO is disable", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:73695-[Skipped:Disconnected]PO is disable"), func() {

		if !exutil.IsTechPreviewNoUpgrade(oc) {
			g.Skip("PO is supported in TP only currently, so skip it")
		}
		_, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("co", "platform-operators-aggregated").Output()
		o.Expect(err).To(o.HaveOccurred(), "PO is not disable")
	})

	g.It("PolarionID:24076-[OTP]check the version of olm operator is appropriate in ClusterOperator", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:24076-check the version of olm operator is appropriate in ClusterOperator"), func() {
		var (
			olmClusterOperatorName = "operator-lifecycle-manager"
		)

		g.By("get the version of olm operator")
		olmVersion := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "clusteroperator", olmClusterOperatorName, "-o=jsonpath={.status.versions[?(@.name==\"operator\")].version}")
		o.Expect(olmVersion).NotTo(o.BeEmpty())

		g.By("Check if it is appropriate in ClusterOperator")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, olmVersion, exutil.Ok, []string{"clusteroperator", "-o=jsonpath={.items[?(@.metadata.name==\"" + olmClusterOperatorName + "\")].status.versions[?(@.name==\"operator\")].version}"}).Check(oc)
	})

	g.It("PolarionID:29775-PolarionID:29786-[OTP][Skipped:Disconnected]as oc user on linux to mirror catalog image[Slow][Timeout:30m]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:29775-PolarionID:29786-[Skipped:Disconnected]as oc user on linux to mirror catalog image[Slow][Timeout:30m]"), g.Label("NonHyperShiftHOST"), func() {
		var (
			bundleIndex1         = "quay.io/kuiwang/operators-all:v1"
			bundleIndex2         = "quay.io/kuiwang/operators-dockerio:v1"
			operatorAllPath      = "/tmp/operators-all-manifests-" + exutil.GetRandomString()
			operatorDockerioPath = "/tmp/operators-dockerio-manifests-" + exutil.GetRandomString()
		)
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr "+operatorAllPath).Output() }()
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr "+operatorDockerioPath).Output() }()

		g.By("mirror to quay.io/kuiwang")
		var output string
		var err error
		// Add timeout and retry mechanism for network resilience
		// Timeout: 6 minutes (360s), Retry interval: 60s, Max retries: 2
		err = wait.PollUntilContextTimeout(context.TODO(), 60*time.Second, 6*time.Minute, false, func(ctx context.Context) (bool, error) {
			e2e.Logf("Executing 'oc adm catalog mirror' for %s (may take several minutes)...", bundleIndex1)
			output, err = oc.AsAdmin().WithoutNamespace().Run("adm", "catalog", "mirror").Args("--manifests-only", "--to-manifests="+operatorAllPath, bundleIndex1, "quay.io/kuiwang").Output()
			if err != nil {
				e2e.Logf("Warning: catalog mirror command failed (will retry): %v", err)
				return false, nil // Retry on failure
			}
			if !strings.Contains(output, "operators-all-manifests") {
				e2e.Logf("Warning: expected output not found (will retry)")
				return false, nil // Retry if output is unexpected
			}
			e2e.Logf("Successfully completed catalog mirror for %s", bundleIndex1)
			return true, nil // Success
		})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("operators-all-manifests"))

		g.By("check mapping.txt")
		result, err := exec.Command("bash", "-c", "cat "+operatorAllPath+"/mapping.txt|grep -E \"atlasmap-atlasmap-operator:0.1.0|quay.io/kuiwang/jmckind-argocd-operator:[a-z0-9][a-z0-9]|redhat-cop-cert-utils-operator:latest\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("atlasmap-atlasmap-operator:0.1.0"))
		o.Expect(result).To(o.ContainSubstring("redhat-cop-cert-utils-operator:latest"))
		o.Expect(result).To(o.ContainSubstring("quay.io/kuiwang/jmckind-argocd-operator"))

		g.By("check icsp yaml")
		result, err = exec.Command("bash", "-c", "cat "+operatorAllPath+"/imageContentSourcePolicy.yaml | grep -E \"quay.io/kuiwang/strimzi-operator|docker.io/strimzi/operator$\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("- quay.io/kuiwang/strimzi-operator"))
		o.Expect(result).To(o.ContainSubstring("source: docker.io/strimzi/operator"))

		g.By("mirror to localhost:5000")
		// Add timeout and retry mechanism for network resilience
		// Timeout: 6 minutes (360s), Retry interval: 60s, Max retries: 2
		err = wait.PollUntilContextTimeout(context.TODO(), 60*time.Second, 6*time.Minute, false, func(ctx context.Context) (bool, error) {
			e2e.Logf("Executing 'oc adm catalog mirror' for %s (may take several minutes)...", bundleIndex2)
			output, err = oc.AsAdmin().WithoutNamespace().Run("adm", "catalog", "mirror").Args("--manifests-only", "--to-manifests="+operatorDockerioPath, bundleIndex2, "localhost:5000").Output()
			if err != nil {
				e2e.Logf("Warning: catalog mirror command failed (will retry): %v", err)
				return false, nil // Retry on failure
			}
			if !strings.Contains(output, "operators-dockerio-manifests") {
				e2e.Logf("Warning: expected output not found (will retry)")
				return false, nil // Retry if output is unexpected
			}
			e2e.Logf("Successfully completed catalog mirror for %s", bundleIndex2)
			return true, nil // Success
		})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("operators-dockerio-manifests"))

		g.By("check mapping.txt to localhost:5000")
		result, err = exec.Command("bash", "-c", "cat "+operatorDockerioPath+"/mapping.txt|grep -E \"localhost:5000/atlasmap/atlasmap-operator:0.1.0|localhost:5000/strimzi/operator:[a-z0-9][a-z0-9]\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("localhost:5000/atlasmap/atlasmap-operator:0.1.0"))
		o.Expect(result).To(o.ContainSubstring("localhost:5000/strimzi/operator"))

		g.By("check icsp yaml to localhost:5000")
		result, err = exec.Command("bash", "-c", "cat "+operatorDockerioPath+"/imageContentSourcePolicy.yaml | grep -E \"localhost:5000/strimzi/operator|docker.io/strimzi/operator$\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("- localhost:5000/strimzi/operator"))
		o.Expect(result).To(o.ContainSubstring("source: docker.io/strimzi/operator"))
		o.Expect(result).NotTo(o.ContainSubstring("docker.io/atlasmap/atlasmap-operator"))
	})

	g.It("PolarionID:33452-[OTP][Skipped:Disconnected]oc adm catalog mirror does not mirror the index image itself", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:33452-[Skipped:Disconnected]oc adm catalog mirror does not mirror the index image itself"), g.Label("NonHyperShiftHOST"), func() {
		var (
			bundleIndex1 = "quay.io/olmqe/olm-api@sha256:71cfd4deaa493d31cd1d8255b1dce0fb670ae574f4839c778f2cfb1bf1f96995"
			manifestPath = "/tmp/manifests-olm-api-" + exutil.GetRandomString()
		)
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr "+manifestPath).Output() }()

		g.By("mirror to localhost:5000/test")
		output, err := oc.AsAdmin().WithoutNamespace().Run("adm", "catalog", "mirror").Args("--manifests-only", "--to-manifests="+manifestPath, bundleIndex1, "localhost:5000/test").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("manifests-olm-api"))

		g.By("check mapping.txt to localhost:5000")
		result, err := exec.Command("bash", "-c", "cat "+manifestPath+"/mapping.txt|grep -E \"quay.io/olmqe/olm-api\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("quay.io/olmqe/olm-api"))

		g.By("check icsp yaml to localhost:5000")
		result, err = exec.Command("bash", "-c", "cat "+manifestPath+"/imageContentSourcePolicy.yaml | grep -E \"quay.io/olmqe/olm-api\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("quay.io/olmqe/olm-api"))
	})

	g.It("PolarionID:21825-[OTP][Skipped:Disconnected]Certs for packageserver can be rotated successfully [Serial]", g.Label("NonHyperShiftHOST"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:21825-[Skipped:Disconnected]Certs for packageserver can be rotated successfully [Serial]"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)
		var (
			packageserverName = "packageserver"
		)

		g.By("Get certsRotateAt and APIService name")
		resources := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", packageserverName, "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.certsRotateAt}{\" \"}{.status.requirementStatus[?(@.kind==\"APIService\")].name}")
		o.Expect(resources).NotTo(o.BeEmpty())
		resourceFields := strings.Fields(resources)
		o.Expect(len(resourceFields)).To(o.BeNumerically(">=", 2))
		apiServiceName := resourceFields[1]
		certsRotateAt, err := time.Parse(time.RFC3339, resourceFields[0])
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Get caBundle")
		caBundle := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiservices", apiServiceName, "-o=jsonpath={.spec.caBundle}")
		o.Expect(caBundle).NotTo(o.BeEmpty())

		g.By("Change caBundle")
		exutil.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiservices", apiServiceName, "-p", "{\"spec\":{\"caBundle\":\"test"+caBundle+"\"}}")

		g.By("Check updated certsRotataAt")
		err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 150*time.Second, false, func(ctx context.Context) (bool, error) {
			updatedCertsRotateAtStr := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", packageserverName, "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.certsRotateAt}")
			updatedCertsRotateAt, err := time.Parse(time.RFC3339, updatedCertsRotateAtStr)
			if err != nil {
				e2e.Logf("the get error is %v, and try next", err)
				return false, nil
			}
			if !updatedCertsRotateAt.Equal(certsRotateAt) {
				e2e.Logf("wait update, and try next")
				return false, nil
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "csv "+packageserverName+" cert is not updated")

		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "redhat-operators", exutil.Ok, []string{"packagemanifest", "--selector=catalog=redhat-operators", "-o=jsonpath={.items[*].status.catalogSource}"}).Check(oc)
	})

	g.It("PolarionID:83105-[OTP][Skipped:Disconnected]olmv0 static networkpolicy on ocp", g.Label("NonHyperShiftHOST", "ReleaseGate"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:83105-[Skipped:Disconnected]olmv0 static networkpolicy on ocp"), func() {

		policies := []olmv0util.NpExpecter{
			{
				Name:      "default-allow-all",
				Namespace: "openshift-operators",
				ExpectIngress: []olmv0util.IngressRule{
					{Ports: []olmv0util.Port{{}}, Selectors: nil},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{Ports: []olmv0util.Port{{}}, Selectors: nil},
				},
				ExpectSelector:    map[string]string{},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "catalog-operator",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{
						Ports:     []olmv0util.Port{{Port: "metrics", Protocol: "TCP"}},
						Selectors: nil,
					},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports:     []olmv0util.Port{{Port: 6443, Protocol: "TCP"}},
						Selectors: nil,
					},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
					{
						Ports:     []olmv0util.Port{{Port: 50051, Protocol: "TCP"}},
						Selectors: nil,
					},
				},
				ExpectSelector:    map[string]string{"app": "catalog-operator"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:          "collect-profiles",
				Namespace:     "openshift-operator-lifecycle-manager",
				ExpectIngress: nil,
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports: []olmv0util.Port{{Port: 8443, Protocol: "TCP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"name": "openshift-operator-lifecycle-manager"}},
							{PodLabels: map[string]string{"app": "olm-operator"}},
							{PodLabels: map[string]string{"app": "catalog-operator"}},
						},
					},
					{
						Ports:     []olmv0util.Port{{Port: 6443, Protocol: "TCP"}},
						Selectors: nil,
					},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
				},
				ExpectSelector:    map[string]string{"app": "olm-collect-profiles"},
				ExpectPolicyTypes: []string{"Egress", "Ingress"},
			},
			{
				Name:              "default-deny-all-traffic",
				Namespace:         "openshift-operator-lifecycle-manager",
				ExpectIngress:     nil,
				ExpectEgress:      nil,
				ExpectSelector:    map[string]string{},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "olm-operator",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{
						Ports:     []olmv0util.Port{{Port: "metrics", Protocol: "TCP"}},
						Selectors: nil,
					},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports:     []olmv0util.Port{{Port: 6443, Protocol: "TCP"}},
						Selectors: nil,
					},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
				},
				ExpectSelector:    map[string]string{"app": "olm-operator"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "package-server-manager",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{
						Ports:     []olmv0util.Port{{Port: 8443, Protocol: "TCP"}},
						Selectors: nil,
					},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports:     []olmv0util.Port{{Port: 6443, Protocol: "TCP"}},
						Selectors: nil,
					},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
				},
				ExpectSelector:    map[string]string{"app": "package-server-manager"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "packageserver",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{
						Ports:     []olmv0util.Port{{Port: 5443, Protocol: "TCP"}},
						Selectors: nil,
					},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports:     []olmv0util.Port{{Port: 6443, Protocol: "TCP"}},
						Selectors: nil,
					},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
					{
						Ports:     []olmv0util.Port{{Port: 50051, Protocol: "TCP"}},
						Selectors: nil,
					},
				},
				ExpectSelector:    map[string]string{"app": "packageserver"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
		}
		if _, err := oc.AsAdmin().WithoutNamespace().
			Run("get").
			Args("catsrc", "redhat-operators", "-n", "openshift-marketplace").
			Output(); err == nil {

			if status, err := oc.AsAdmin().WithoutNamespace().
				Run("get").
				Args("catsrc", "redhat-operators", "-n", "openshift-marketplace",
					"-o=jsonpath={.status.connectionState.lastObservedState}").
				Output(); err == nil && status == "READY" {

				policies = append(policies,
					olmv0util.NpExpecter{
						Name:      "redhat-operators-grpc-server",
						Namespace: "openshift-marketplace",
						ExpectIngress: []olmv0util.IngressRule{
							{
								Ports:     []olmv0util.Port{{Port: 50051, Protocol: "TCP"}},
								Selectors: nil,
							},
						},
						ExpectEgress:      nil,
						ExpectSelector:    map[string]string{"olm.catalogSource": "redhat-operators", "olm.managed": "true"},
						ExpectPolicyTypes: []string{"Ingress", "Egress"},
					},
					olmv0util.NpExpecter{
						Name:          "redhat-operators-unpack-bundles",
						Namespace:     "openshift-marketplace",
						ExpectIngress: nil,
						ExpectEgress: []olmv0util.EgressRule{
							{
								Ports:     []olmv0util.Port{{Port: 6443, Protocol: "TCP"}},
								Selectors: nil,
							},
						},
						ExpectSelector:    map[string]string{},
						ExpectPolicyTypes: []string{"Ingress", "Egress"},
					},
				)
			}
		}

		for _, policy := range policies {

			g.By(fmt.Sprintf("Checking NP %s in %s", policy.Name, policy.Namespace))
			specs, err := oc.AsAdmin().WithoutNamespace().
				Run("get").Args("networkpolicy", policy.Name, "-n", policy.Namespace, "-o=jsonpath={.spec}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(specs).NotTo(o.BeEmpty())
			e2e.Logf("specs: %v", specs)

			olmv0util.VerifySelector(specs, policy.ExpectSelector, policy.Name)
			olmv0util.VerifyPolicyTypes(specs, policy.ExpectPolicyTypes, policy.Name)
			olmv0util.VerifyIngress(specs, policy.ExpectIngress, policy.Name)
			olmv0util.VerifyEgress(specs, policy.ExpectEgress, policy.Name)
			if strings.Contains(policy.Name, "redhat-operators-unpack-bundles") {
				exprs := gjson.Get(specs, "podSelector.matchExpressions").Array()
				o.Expect(len(exprs)).To(o.Equal(2), "expect two matchExpressions")
				o.Expect(exprs[0].Get("key").String()).To(o.ContainSubstring("operatorframework.io/bundle-unpack-ref"))
				o.Expect(exprs[0].Get("operator").String()).To(o.ContainSubstring("Exists"))
				o.Expect(exprs[1].Get("key").String()).To(o.ContainSubstring("olm.managed"))
				o.Expect(exprs[1].Get("operator").String()).To(o.ContainSubstring("In"))
			}
			if strings.Contains(policy.Name, "redhat-operators-grpc-server") {
				err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifests", "-n", "openshift-marketplace", "--selector=catalog=redhat-operators").Execute()
				o.Expect(err).NotTo(o.HaveOccurred())
			}
			if strings.Contains(policy.Name, "collect-profiles") {
				status, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-collect-profiles").Output()
				o.Expect(status).To(o.ContainSubstring("Completed"))
			}
		}

	})

	g.It("PolarionID:83583-[OTP][Skipped:Disconnected]olmv0 networkpolicy on hosted hypershift", g.Label("NonHyperShiftHOST", "ReleaseGate"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:83583-[Skipped:Disconnected]olmv0 networkpolicy on hosted hypershift"), func() {

		topology, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("infrastructures.config.openshift.io",
			"cluster", "-o=jsonpath={.status.controlPlaneTopology}").Output()
		if err != nil || strings.Compare(topology, "External") != 0 {
			g.Skip("the cluster is unhealthy or not hypershift hosted cluster")
		}

		policies := []olmv0util.NpExpecter{
			{
				Name:      "default-allow-all",
				Namespace: "openshift-operators",
				ExpectIngress: []olmv0util.IngressRule{
					{Ports: []olmv0util.Port{{}}, Selectors: nil},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{Ports: []olmv0util.Port{{}}, Selectors: nil},
				},
				ExpectSelector:    map[string]string{},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "catalog-operator",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{Ports: []olmv0util.Port{{Port: "metrics", Protocol: "TCP"}}, Selectors: nil},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{Ports: []olmv0util.Port{{Port: 6443, Protocol: "TCP"}}, Selectors: nil},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
					{Ports: []olmv0util.Port{{Port: 50051, Protocol: "TCP"}}, Selectors: nil},
				},
				ExpectSelector:    map[string]string{"app": "catalog-operator"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:          "collect-profiles",
				Namespace:     "openshift-operator-lifecycle-manager",
				ExpectIngress: nil,
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports: []olmv0util.Port{{Port: 8443, Protocol: "TCP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"name": "openshift-operator-lifecycle-manager"}},
							{PodLabels: map[string]string{"app": "olm-operator"}},
							{PodLabels: map[string]string{"app": "catalog-operator"}},
						},
					},
					{Ports: []olmv0util.Port{{Port: 6443, Protocol: "TCP"}}, Selectors: nil},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
				},
				ExpectSelector:    map[string]string{"app": "olm-collect-profiles"},
				ExpectPolicyTypes: []string{"Egress", "Ingress"},
			},
			{
				Name:              "default-deny-all-traffic",
				Namespace:         "openshift-operator-lifecycle-manager",
				ExpectIngress:     nil,
				ExpectEgress:      nil,
				ExpectSelector:    map[string]string{},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "olm-operator",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{Ports: []olmv0util.Port{{Port: "metrics", Protocol: "TCP"}}, Selectors: nil},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{Ports: []olmv0util.Port{{Port: 6443, Protocol: "TCP"}}, Selectors: nil},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
				},
				ExpectSelector:    map[string]string{"app": "olm-operator"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "package-server-manager",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{Ports: []olmv0util.Port{{Port: 8443, Protocol: "TCP"}}, Selectors: nil},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{Ports: []olmv0util.Port{{Port: 6443, Protocol: "TCP"}}, Selectors: nil},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
				},
				ExpectSelector:    map[string]string{"app": "package-server-manager"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			{
				Name:      "packageserver",
				Namespace: "openshift-operator-lifecycle-manager",
				ExpectIngress: []olmv0util.IngressRule{
					{Ports: []olmv0util.Port{{Port: 5443, Protocol: "TCP"}}, Selectors: nil},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{Ports: []olmv0util.Port{{Port: 6443, Protocol: "TCP"}}, Selectors: nil},
					{
						Ports: []olmv0util.Port{{Port: "dns-tcp", Protocol: "TCP"}, {Port: "dns", Protocol: "UDP"}},
						Selectors: []olmv0util.Selector{
							{NamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"}},
						},
					},
					{Ports: []olmv0util.Port{{Port: 50051, Protocol: "TCP"}}, Selectors: nil},
				},
				ExpectSelector:    map[string]string{"app": "packageserver"},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
		}

		for _, policy := range policies {

			g.By(fmt.Sprintf("Checking NP %s in %s", policy.Name, policy.Namespace))
			specs, err := oc.AsAdmin().WithoutNamespace().
				Run("get").Args("networkpolicy", policy.Name, "-n", policy.Namespace, "-o=jsonpath={.spec}").Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(specs).NotTo(o.BeEmpty())
			e2e.Logf("specs: %v", specs)

			olmv0util.VerifySelector(specs, policy.ExpectSelector, policy.Name)
			olmv0util.VerifyPolicyTypes(specs, policy.ExpectPolicyTypes, policy.Name)
			olmv0util.VerifyIngress(specs, policy.ExpectIngress, policy.Name)
			olmv0util.VerifyEgress(specs, policy.ExpectEgress, policy.Name)
		}

	})

	g.It("PolarionID:21080-[OTP][Skipped:Disconnected]Check metrics[Serial]", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipOnProxyCluster(oc)

		var (
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogTemplate          = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subFile             = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

			data          olmv0util.PrometheusQueryResult
			err           error
			exists        bool
			metricsBefore olmv0util.Metrics
			metricsAfter  olmv0util.Metrics
			olmToken      string
		)

		oc.SetupProject()
		ns := oc.Namespace()
		itName := g.CurrentSpecReport().FullText()
		og := olmv0util.OperatorGroupDescription{
			Name:      "test-21080-group",
			Namespace: ns,
			Template:  ogTemplate,
		}
		g.By("create the learn-operator CatalogSource")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "catsrc-21080",
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
			SubName:                "sub-21080",
			Namespace:              ns,
			CatalogSourceName:      "catsrc-21080",
			CatalogSourceNamespace: ns,
			IpApproval:             "Automatic",
			Channel:                "beta",
			OperatorPackage:        "learn",
			SingleNamespace:        true,
			Template:               subFile,
		}

		g.By("1, check if this operator ready for installing")
		e2e.Logf("Check if %v exists in the %v catalog", sub.OperatorPackage, sub.CatalogSourceName)
		exists, err = olmv0util.ClusterPackageExistsInNamespace(oc, sub, ns)
		if !exists {
			g.Skip(fmt.Sprintf("%s does not exist in the cluster", sub.OperatorPackage))
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(exists).To(o.BeTrue())

		g.By("2, Get token & pods so that access the Prometheus")
		olmToken, err = exutil.GetSAToken(oc, "prometheus-k8s", "openshift-monitoring")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(olmToken).NotTo(o.BeEmpty())

		// the reason why use it is to workaround the Network policy since OCP4.20
		g.By("2-1, get Prometheus Pod IP address")
		PrometheusPodIP, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", "openshift-monitoring", "prometheus-k8s-0", "-o=jsonpath={.status.podIP}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("3, Collect olm metrics before installing an operator")
		metricsBefore = olmv0util.GetMetrics(oc, olmToken, data, metricsBefore, sub.SubName, PrometheusPodIP)
		e2e.Logf("\nbefore {csv_count, csv_upgrade_count, catalog_source_count, install_plan_count, subscription_count, subscription_sync_total}\n%v", metricsBefore)

		g.By("4, Start to subscribe to etcdoperator")
		og.Create(oc, itName, dr)
		defer sub.Delete(itName, dr) // remove the subscription after test
		sub.Create(oc, itName, dr)

		g.By("4.5 Check for latest version")
		defer sub.DeleteCSV(itName, dr) // remove the csv after test
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded", exutil.Ok, []string{"csv", "learn-operator.v0.0.3", "-n", ns, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("5, learnoperator is at v0.0.3, start to collect olm metrics after")
		// The prometheus-k8s-0 IP might be changed, so rerun it here.
		PrometheusPodIP, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", "openshift-monitoring", "prometheus-k8s-0", "-o=jsonpath={.status.podIP}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		metricsAfter = olmv0util.GetMetrics(oc, olmToken, data, metricsAfter, sub.SubName, PrometheusPodIP)
		g.By("6, Check results")
		e2e.Logf("{csv_count csv_upgrade_count catalog_source_count install_plan_count subscription_count subscription_sync_total}")
		e2e.Logf("%v", metricsBefore)
		e2e.Logf("%v", metricsAfter)
		g.By("All PASS\n")
	})

	g.It("PolarionID:21953-[OTP]Ensure that operator deployment is in the master node", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		var (
			err            error
			msg            string
			olmErrs        = true
			olmJpath       = "-o=jsonpath={@.spec.template.spec.nodeSelector}"
			olmNamespace   = "openshift-marketplace"
			olmNodeName    string
			olmPodFullName string
			olmPodName     = "marketplace-operator"
			nodeRole       = "node-role.kubernetes.io/master"
			nodes          string
			nodeStatus     bool
			pod            string
			pods           string
			status         []string
			x              []string
		)

		g.By("Get deployment")
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "-n", olmNamespace, olmPodName, olmJpath).Output()
		if err != nil {
			e2e.Logf("Unable to get deployment -n %v %v %v.", olmNamespace, olmPodName, olmJpath)
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		if len(msg) < 1 || !strings.Contains(msg, nodeRole) {
			e2e.Failf("Could not find %v variable %v for %v: %v", olmJpath, nodeRole, olmPodName, msg)
		}

		g.By("Look at pods")
		// look for the marketplace-operator pod's full name
		pods, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", olmNamespace, "-o", "wide").Output()
		if err != nil {
			e2e.Logf("Unable to query pods -n %v %v %v.", olmNamespace, err, pods)
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(pods).NotTo(o.ContainSubstring("No resources found"))
		// e2e.Logf("Pods %v ", pods)

		for _, pod = range strings.Split(pods, "\n") {
			if len(pod) <= 0 {
				continue
			}
			// Find the node in the pod
			if strings.Contains(pod, olmPodName) {
				x = strings.Fields(pod)
				olmPodFullName = x[0]
				// olmNodeName = x[6]
				olmNodeName, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", olmNamespace, olmPodFullName, "-o=jsonpath={.spec.nodeName}").Output()
				o.Expect(err).NotTo(o.HaveOccurred())
				olmErrs = false
				// e2e.Logf("Found pod is %v", pod)
				break
			}
		}
		if olmErrs {
			e2e.Failf("Unable to find the full pod name for %v in %v: %v.", olmPodName, olmNamespace, pods)
		}

		g.By("Query node label value")
		// Look at the setting for the node to be on the master
		olmErrs = true
		nodes, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("nodes", "-n", olmNamespace, olmNodeName, "-o=jsonpath={.metadata.labels}").Output()
		if err != nil {
			e2e.Failf("Unable to query nodes -n %v %v %v.", olmNamespace, err, nodes)
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(nodes).To(o.ContainSubstring("node-role.kubernetes.io/master"))

		g.By("look at oc get nodes")
		// Found the setting, verify that it's really on the master node
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("nodes", "-n", olmNamespace, olmNodeName, "--show-labels", "--no-headers").Output()
		if err != nil {
			e2e.Failf("Unable to query the %v node of pod %v for %v's status", olmNodeName, olmPodFullName, msg)
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(msg).NotTo(o.ContainSubstring("No resources found"))
		status = strings.Fields(msg)
		if strings.Contains(status[2], "master") {
			olmErrs = false
			nodeStatus = true
			e2e.Logf("node %v is a %v", olmNodeName, status[2])
		}
		if olmErrs || !nodeStatus {
			e2e.Failf("The node %v of %v pod is not a master:%v", olmNodeName, olmPodFullName, msg)
		}
		g.By("Finish")
		e2e.Logf("The pod %v is on the master node %v", olmPodFullName, olmNodeName)
	})

	g.It("PolarionID:43135-[OTP]PackageServer respects single-node configuration[Slow][Disruptive]", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) get the cluster infrastructure")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster infra")
		}
		num, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", "openshift-operator-lifecycle-manager", "deployment", "packageserver", "-o=jsonpath={.status.replicas}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if infra == "SingleReplica" {
			e2e.Logf("This is a SNO cluster")
			g.By("2) check if only have one packageserver pod")
			if num != "1" {
				e2e.Failf("!!!Fail, should only have 1 packageserver pod, but get %s!", num)
			}
			// make sure the CVO recover if any error in the follow steps
			defer func() {
				_, err = oc.AsAdmin().WithoutNamespace().Run("scale").Args("--replicas", "1", "deployment/cluster-version-operator", "-n", "openshift-cluster-version").Output()
				if err != nil {
					e2e.Failf("Defer: fail to enable CVO")
				}
			}()
			g.By("3) stop CVO")
			_, err := oc.AsAdmin().WithoutNamespace().Run("scale").Args("--replicas", "0", "deployment/cluster-version-operator", "-n", "openshift-cluster-version").Output()
			if err != nil {
				e2e.Failf("Fail to stop CVO")
			}
			g.By("4) stop the PSM")
			_, err = oc.AsAdmin().WithoutNamespace().Run("scale").Args("--replicas", "0", "deployment/package-server-manager", "-n", "openshift-operator-lifecycle-manager").Output()
			if err != nil {
				e2e.Failf("Fail to stop the PSM")
			}
			g.By("5) patch the replica to 3")
			// oc get csv packageserver -o=jsonpath={.spec.install.spec.deployments[?(@.name==\"packageserver\")].spec.replicas}
			// oc patch csv/packageserver -p '{"spec":{"install":{"spec":{"deployments":[{"name":"packageserver", "spec":{"replicas":3, "template":{}, "selector":{"matchLabels":{"app":"packageserver"}}}}]}}}}' --type=merge
			// oc patch deploy/packageserver -p '{"spec":{"replicas":3}}' --type=merge
			// should update CSV
			olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", "openshift-operator-lifecycle-manager", "csv", "packageserver", "-p", "{\"spec\":{\"install\":{\"spec\":{\"deployments\":[{\"name\":\"packageserver\", \"spec\":{\"replicas\":3, \"template\":{}, \"selector\":{\"matchLabels\":{\"app\":\"packageserver\"}}}}]}}}}", "--type=merge")
			olmv0util.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "-n", "openshift-operator-lifecycle-manager", "deployment", "packageserver", "-p", "{\"spec\":{\"replicas\":3}}", "--type=merge")
			err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
				num, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.availableReplicas}").Output()
				e2e.Logf("packageserver replicas is %s", num)
				if num != "3" {
					return false, nil
				}
				return true, nil
			})
			exutil.AssertWaitPollNoErr(err, "packageserver replicas is not 3")
			g.By("6) enable CVO")
			_, err = oc.AsAdmin().WithoutNamespace().Run("scale").Args("--replicas", "1", "deployment/cluster-version-operator", "-n", "openshift-cluster-version").Output()
			if err != nil {
				e2e.Failf("Fail to enable CVO")
			}
			g.By("7) check if the PSM back")
			err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
				num, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "package-server-manager", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.replicas}").Output()
				if num != "1" {
					return false, nil
				}
				return true, nil
			})
			exutil.AssertWaitPollNoErr(err, "package-server-manager replicas is not reback to 1")
			g.By("8) check if the packageserver pods number back to 1")
			// for some SNO clusters, reback may take 10 mins around
			err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
				num, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.availableReplicas}").Output()
				if num != "1" {
					return false, nil
				}
				return true, nil
			})
			exutil.AssertWaitPollNoErr(err, "packageserver replicas is not reback to 1")
		} else {
			// HighlyAvailable
			e2e.Logf("This is HA cluster, not SNO")
			g.By("2) check if only have two packageserver pods")
			if num != "2" {
				e2e.Failf("!!!Fail, should only have 2 packageserver pods, but get %s!", num)
			}
		}
	})

	g.It("PolarionID:24075-[OTP][Skipped:Disconnected]The packagemanifest labels provider value should be correct", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		g.By("Get packagemanifest provider from status")
		provider, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "quay-operator", "-o", "jsonpath={.status.provider.name}", "-n", "openshift-marketplace").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Get packagemanifest provider from labels")
		providerInLabels, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "quay-operator", "-o", "jsonpath={.metadata.labels.provider}", "-n", "openshift-marketplace").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Verify provider value in status equals provider value in labels")
		o.Expect(provider).To(o.Equal(providerInLabels))
	})

	g.It("PolarionID:43276-[OTP][Skipped:Disconnected]oc adm catalog mirror can mirror declaritive index images", func() {
		indexImage := "quay.io/olmqe/etcd-index:dc-new"
		operatorAllPath := "/tmp/operators-all-manifests-" + exutil.GetRandomString()
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr "+operatorAllPath).Output() }()

		g.By("mirror to localhost:5000")
		output, err := oc.AsAdmin().WithoutNamespace().Run("adm", "catalog", "mirror").Args("--manifests-only", "--to-manifests="+operatorAllPath, indexImage, "localhost:5000").Output()

		e2e.Logf("mirror output: %s", output)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("no digest mapping available for quay.io/olmqe/etcd-bundle:dc, skip writing to ImageContentSourcePolicy"))
		o.Expect(output).To(o.ContainSubstring("no digest mapping available for quay.io/olmqe/etcd-index:dc-new, skip writing to ImageContentSourcePolicy"))
		o.Expect(output).To(o.ContainSubstring("wrote mirroring manifests"))

		g.By("check mapping.txt to localhost:5000")
		result, err := exec.Command("bash", "-c", "cat "+operatorAllPath+"/mapping.txt|grep -E \"localhost:5000/olmqe/etcd-bundle|localhost:5000/olmqe/etcd-index\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("mapping result:%s", result)

		o.Expect(result).To(o.ContainSubstring("quay.io/olmqe/etcd-bundle:dc=localhost:5000/olmqe/etcd-bundle:dc"))
		o.Expect(result).To(o.ContainSubstring("quay.io/olmqe/etcd-index:dc-new=localhost:5000/olmqe/etcd-index:dc-new"))

		g.By("check icsp yaml to localhost:5000")
		result, err = exec.Command("bash", "-c", "cat "+operatorAllPath+"/imageContentSourcePolicy.yaml | grep \"localhost:5000\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("icsp result:%s", result)
		o.Expect(result).To(o.ContainSubstring("- localhost:5000/coreos/etcd-operator"))
	})

	g.It("PolarionID:75328-[OTP][Skipped:Disconnected]CatalogSources that use binaryless images must set extractContent", func() {
		buildPruningBaseDir := exutil.FixturePath("testdata", "olm")
		catsrcImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
		catsrcExtractImageTemplate := filepath.Join(buildPruningBaseDir, "catalogsource-image-extract.yaml")
		namespace := oc.Namespace()
		catsrc := olmv0util.CatalogSourceDescription{
			Name:        "binless-catalog-75328",
			Namespace:   namespace,
			DisplayName: "Test Catsrc 75328 without bins",
			Publisher:   "Red Hat",
			SourceType:  "grpc",
			Address:     "quay.io/openshifttest/nginxolm-operator-index:nginxolm75148",
			Template:    catsrcImageTemplate,
		}
		catsrcExtract := olmv0util.CatalogSourceDescription{
			Name:        "binless-catalog-75328-extract",
			Namespace:   namespace,
			DisplayName: "Test Catsrc 75328 without bins",
			Publisher:   "Red Hat",
			SourceType:  "grpc",
			Address:     "quay.io/openshifttest/nginxolm-operator-index:nginxolm75148",
			Template:    catsrcExtractImageTemplate,
		}

		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)

		g.By("Create catalogsource that use binaryless images without extractContent")
		defer catsrc.Delete(itName, dr)
		catsrc.Create(oc, itName, dr)

		g.By("Check the catalogsource fail")
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			status, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status..lastObservedState}").Output()
			if strings.Compare(status, "TRANSIENT_FAILURE") != 0 {
				e2e.Logf("catsrc %s lastObservedState is %s, not TRANSIENT_FAILURE", catsrc.Name, status)
				return false, nil
			}
			return true, nil
		})
		if waitErr != nil {
			output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status}").Output()
			e2e.Logf("catsrc status: %s", output)
			olmv0util.LogDebugInfo(oc, catsrc.Namespace, "pod", "events")
		}
		exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("catsrc %s lastObservedState is not TRANSIENT_FAILURE", catsrc.Name))
		e2e.Logf("catsrc %s lastObservedState is TRANSIENT_FAILURE", catsrc.Name)

		podName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", "olm.catalogSource=binless-catalog-75328", "-o=jsonpath={.items[0].metadata.name}", "-n", oc.Namespace()).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(podName).NotTo(o.BeEmpty())

		log, _ := oc.AsAdmin().WithoutNamespace().Run("logs").Args(podName, "-n", catsrc.Namespace, "--tail", "3").Output()
		if !strings.Contains(log, "CreateContainerError") {
			e2e.Failf("need CreateContainerError: %s", log)
		}

		g.By("packagemanifest not be created")
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifests", "nginx75148", "-n", catsrc.Namespace).Output()
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("\"nginx75148\" not found"))

		catsrc.Delete(itName, dr)

		g.By("Create catalogsource that use binaryless images with extractContent")
		defer catsrcExtract.Delete(itName, dr)
		catsrcExtract.CreateWithCheck(oc, itName, dr)

		g.By("packagemanifest works well")
		entries := olmv0util.GetResourceNoEmpty(oc, exutil.AsAdmin, exutil.WithoutNamespace, "packagemanifest", "nginx75148", "-n", catsrcExtract.Namespace, "-o=jsonpath={.status.channels[?(@.name==\"candidate-v1.0\")].entries}")
		o.Expect(entries).To(o.ContainSubstring("nginx75148.v1.0.6"))

	})

	g.It("PolarionID:72018-[OTP][Skipped:Disconnected]Do not sync namespaces that have no subscriptions", g.Label("NonHyperShiftHOST"), func() {
		oc.SetupProject()
		namespaceName := oc.Namespace()
		catPodname, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", "openshift-operator-lifecycle-manager", "--selector=app=catalog-operator", "-o=jsonpath={.items..metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(catPodname).NotTo(o.BeEmpty())
		catalogs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args(catPodname, "-n", "openshift-operator-lifecycle-manager", "--since=60s").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if len(catalogs) != 0 {
			for _, line := range strings.Split(catalogs, "\n") {
				if strings.Contains(line, namespaceName) {
					e2e.Logf("catalog log line: %s", line)
					o.Expect(line).NotTo(o.ContainSubstring("found 0 operatorGroups"))
				}
			}
		} else {
			e2e.Logf("log is empty")
		}

	})

	g.It("PolarionID:25674-[OTP]restart the marketplace-operator [Disruptive]", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipNoCapabilities(oc, "marketplace")

		g.By("delete pod of marketplace")
		_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "pod", "--selector=name=marketplace-operator", "-n", "openshift-marketplace", "--ignore-not-found")
		o.Expect(err).NotTo(o.HaveOccurred())

		_, _ = exec.Command("bash", "-c", "sleep 10").Output()

		g.By("pod of marketplace restart")
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "TrueFalseFalse", exutil.Ok, []string{"clusteroperator", "marketplace",
			"-o=jsonpath={.status.conditions[?(@.type==\"Available\")].status}{.status.conditions[?(@.type==\"Progressing\")].status}{.status.conditions[?(@.type==\"Degraded\")].status}"}).Check(oc)

	})

	g.It("PolarionID:43642-[OTP][Skipped:Disconnected]Alert rule is configured to check catalogsource_ready in openshift-marketplace", g.Label("NonHyperShiftHOST"), func() {
		exutil.SkipBaselineCaps(oc, "None")
		exutil.SkipIfDisableDefaultCatalogsource(oc)

		g.By("Check PrometheusRule exists in openshift-marketplace namespace")
		prometheusRules, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("prometheusrule", "-n", "openshift-marketplace", "-o=jsonpath={.items[*].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(prometheusRules).NotTo(o.BeEmpty(), "PrometheusRule should exist in openshift-marketplace namespace")
		e2e.Logf("Found PrometheusRules in openshift-marketplace: %s", prometheusRules)

		g.By("Verify alert rule contains catalogsource_ready metric check")
		// Get all PrometheusRule resources and check if any contains catalogsource_ready
		rules := strings.Fields(prometheusRules)
		foundCatalogSourceReadyRule := false

		for _, rule := range rules {
			ruleYaml, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("prometheusrule", rule, "-n", "openshift-marketplace", "-o=yaml").Output()
			if err != nil {
				e2e.Logf("Failed to get PrometheusRule %s: %v", rule, err)
				continue
			}
			if strings.Contains(ruleYaml, "catalogsource_ready") {
				foundCatalogSourceReadyRule = true
				e2e.Logf("Found catalogsource_ready in PrometheusRule: %s", rule)
				break
			}
		}
		o.Expect(foundCatalogSourceReadyRule).To(o.BeTrue(), "PrometheusRule should contain catalogsource_ready metric check")
		e2e.Logf("PrometheusRule validation passed: Alert rule is properly configured to monitor catalogsource_ready metric")
	})

	g.It("PolarionID:43975-[OTP]olm operator serviceaccount should not rely on external networking for health check[Disruptive][Slow]", g.Label("NonHyperShiftHOST"), func() {
		g.By("1) get the cluster infrastructure")
		infra, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures", "cluster", "-o=jsonpath={.status.infrastructureTopology}").Output()
		if err != nil {
			e2e.Failf("Fail to get the cluster infra")
		}
		if infra != "SingleReplica" {
			g.Skip("Not SNO cluster - skipping test ...")
		}

		originProfile := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiserver", "cluster", "-o=jsonpath={.spec.audit.profile}")
		o.Expect(originProfile).NotTo(o.BeEmpty())
		if strings.Compare(originProfile, "Default") != 0 {
			g.Skip("audit profile is not Default - skipping test ...")
		}

		g.By("2) get revision number")
		revisionNumber1 := 0
		reg := regexp.MustCompile(`at revision (\d+)`)
		if reg == nil {
			e2e.Failf("get revision number regexp err!")
		}
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("kubeapiserver", "-o=jsonpath={..status.conditions[?(@.type==\"NodeInstallerProgressing\")]}").Output()
		if err != nil {
			e2e.Failf("Fail to get kubeapiserver")
		}
		result := reg.FindAllStringSubmatch(output, -1)
		if result != nil {
			revisionNumberStr1 := result[0][1]
			revisionNumber1, _ = strconv.Atoi(revisionNumberStr1)
			e2e.Logf("origin revision number is : %v", revisionNumber1)
		} else {
			e2e.Failf("Fail to get revision number")
		}

		g.By("3) Configuring the audit log policy to AllRequestBodies")
		defer func() {
			pathJSON := fmt.Sprintf("{\"spec\":{\"audit\":{\"profile\":\"%s\"}}}", originProfile)
			e2e.Logf("recover to be %v", pathJSON)
			exutil.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiserver", "cluster", "-p", pathJSON, "--type=merge")
			output = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiserver", "cluster", "-o=jsonpath={.spec.audit.profile}")
			o.Expect(output).To(o.Equal("Default"))
		}()
		exutil.PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiserver", "cluster", "-p", "{\"spec\":{\"audit\":{\"profile\":\"AllRequestBodies\"}}}", "--type=merge")
		output = olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "apiserver", "cluster", "-o=jsonpath={.spec.audit.profile}")
		o.Expect(output).To(o.Equal("AllRequestBodies"))

		g.By("4) Wait for api rollout")
		err = wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("kubeapiserver", "-o=jsonpath={..status.conditions[?(@.type==\"NodeInstallerProgressing\")]}").Output()
			e2e.Logf("kubeapiserver status output: %s", output)
			if err != nil {
				e2e.Logf("Fail to get kubeapiserver status, go next round")
				return false, nil
			}
			if !strings.Contains(output, "AllNodesAtLatestRevision") {
				e2e.Logf("the api is rolling, go next round")
				return false, nil
			}
			result := reg.FindAllStringSubmatch(output, -1)
			if result != nil {
				revisionNumberStr2 := result[0][1]
				revisionNumber2, _ := strconv.Atoi(revisionNumberStr2)
				e2e.Logf("revision number is : %v", revisionNumber2)
				if revisionNumber2 > revisionNumber1 {
					return true, nil
				}
				e2e.Logf("revision number is not changed, go next round")
				return false, nil

			}
			e2e.Logf("Fail to get revision number, go next round")
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "api not rollout")

		// According to the case steps, wait for 5 minutes, then check the audit log doesn't contain olm-operator-serviceaccount.
		g.By("Wait for 5 minutes, then check the audit log")
		time.Sleep(5 * time.Minute)

		g.By("check the audit log")
		nodeName, err := exutil.GetFirstMasterNode(oc)
		e2e.Logf("master node name: %s", nodeName)
		o.Expect(err).NotTo(o.HaveOccurred())
		auditlogPath := "43975.log"
		defer func() {
			_, _ = exec.Command("bash", "-c", "rm -fr "+auditlogPath).Output()
		}()
		outputPath, err := oc.AsAdmin().WithoutNamespace().Run("adm").Args("node-logs", nodeName, "--path=kube-apiserver/audit.log").OutputToFile(auditlogPath)
		o.Expect(err).NotTo(o.HaveOccurred())
		commandParserLog := "cat " + outputPath + " | grep -i health | grep -i subjectaccessreviews | grep -v Unhealth | jq -r '.user.username' | sort | uniq"
		resultParserLog, err := exec.Command("bash", "-c", commandParserLog).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("Found usernames in audit log: %s", string(resultParserLog))
		o.Expect(resultParserLog).NotTo(o.ContainSubstring("olm-operator-serviceaccount"))
	})

	g.It("PolarionID:43057-[OTP]Enable continuous heap profiling by default", g.Label("NonHyperShiftHOST"), func() {

		g.By("get pod of marketplace")
		configMaps := olmv0util.GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "configmaps", "-l", "olm.openshift.io/pprof", "-n", "openshift-operator-lifecycle-manager")
		o.Expect(configMaps).NotTo(o.BeEmpty())
		e2e.Logf("%s", configMaps)

		linesconfigMaps := strings.Split(configMaps, "\n")
		for i := 1; i < len(linesconfigMaps); i++ {
			e2e.Logf("i: %v", i)
			configMap := strings.Split(linesconfigMaps[i], " ")
			e2e.Logf("configMap: %v", configMap[0])

			binaryConfigMap, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("configmaps", configMap[0], "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.binaryData.*}").OutputToFile("config-43057.json")
			o.Expect(err).NotTo(o.HaveOccurred())
			e2e.Logf("binaryConfigMap: %v", binaryConfigMap)

			resultBase64, err := exec.Command("bash", "-c", fmt.Sprintf("cat %s | base64 -d", binaryConfigMap)).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(resultBase64).NotTo(o.BeEmpty())
		}

	})

})
