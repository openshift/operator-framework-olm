package specs

import (
	"context"
	"fmt"
	"os/exec"
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

	g.It("PolarionID:29775-PolarionID:29786-[OTP][Skipped:Disconnected]as oc user on linux to mirror catalog image[Slow][Timeout:30m]", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:29775-PolarionID:29786-[Skipped:Disconnected]as oc user on linux to mirror catalog image[Slow][Timeout:30m]"), func() {
		var (
			bundleIndex1         = "quay.io/kuiwang/operators-all:v1"
			bundleIndex2         = "quay.io/kuiwang/operators-dockerio:v1"
			operatorAllPath      = "operators-all-manifests-" + exutil.GetRandomString()
			operatorDockerioPath = "operators-dockerio-manifests-" + exutil.GetRandomString()
		)
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr ./"+operatorAllPath).Output() }()
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr ./"+operatorDockerioPath).Output() }()

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
		result, err := exec.Command("bash", "-c", "cat ./"+operatorAllPath+"/mapping.txt|grep -E \"atlasmap-atlasmap-operator:0.1.0|quay.io/kuiwang/jmckind-argocd-operator:[a-z0-9][a-z0-9]|redhat-cop-cert-utils-operator:latest\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("atlasmap-atlasmap-operator:0.1.0"))
		o.Expect(result).To(o.ContainSubstring("redhat-cop-cert-utils-operator:latest"))
		o.Expect(result).To(o.ContainSubstring("quay.io/kuiwang/jmckind-argocd-operator"))

		g.By("check icsp yaml")
		result, err = exec.Command("bash", "-c", "cat ./"+operatorAllPath+"/imageContentSourcePolicy.yaml | grep -E \"quay.io/kuiwang/strimzi-operator|docker.io/strimzi/operator$\"").Output()
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
		result, err = exec.Command("bash", "-c", "cat ./"+operatorDockerioPath+"/mapping.txt|grep -E \"localhost:5000/atlasmap/atlasmap-operator:0.1.0|localhost:5000/strimzi/operator:[a-z0-9][a-z0-9]\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("localhost:5000/atlasmap/atlasmap-operator:0.1.0"))
		o.Expect(result).To(o.ContainSubstring("localhost:5000/strimzi/operator"))

		g.By("check icsp yaml to localhost:5000")
		result, err = exec.Command("bash", "-c", "cat ./"+operatorDockerioPath+"/imageContentSourcePolicy.yaml | grep -E \"localhost:5000/strimzi/operator|docker.io/strimzi/operator$\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("- localhost:5000/strimzi/operator"))
		o.Expect(result).To(o.ContainSubstring("source: docker.io/strimzi/operator"))
		o.Expect(result).NotTo(o.ContainSubstring("docker.io/atlasmap/atlasmap-operator"))
	})

	g.It("PolarionID:33452-[OTP][Skipped:Disconnected]oc adm catalog mirror does not mirror the index image itself", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:33452-[Skipped:Disconnected]oc adm catalog mirror does not mirror the index image itself"), func() {
		var (
			bundleIndex1 = "quay.io/olmqe/olm-api@sha256:71cfd4deaa493d31cd1d8255b1dce0fb670ae574f4839c778f2cfb1bf1f96995"
			manifestPath = "manifests-olm-api-" + exutil.GetRandomString()
		)
		defer func() { _, _ = exec.Command("bash", "-c", "rm -fr ./"+manifestPath).Output() }()

		g.By("mirror to localhost:5000/test")
		output, err := oc.AsAdmin().WithoutNamespace().Run("adm", "catalog", "mirror").Args("--manifests-only", "--to-manifests="+manifestPath, bundleIndex1, "localhost:5000/test").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("manifests-olm-api"))

		g.By("check mapping.txt to localhost:5000")
		result, err := exec.Command("bash", "-c", "cat ./"+manifestPath+"/mapping.txt|grep -E \"quay.io/olmqe/olm-api\"").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).To(o.ContainSubstring("quay.io/olmqe/olm-api"))

		g.By("check icsp yaml to localhost:5000")
		result, err = exec.Command("bash", "-c", "cat ./"+manifestPath+"/imageContentSourcePolicy.yaml | grep -E \"quay.io/olmqe/olm-api\"").Output()
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

})
