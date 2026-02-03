package specs

import (
	"fmt"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// Separate Describe block for networkpolicy tests that don't need project setup
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 networkpolicy", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLIWithoutNamespace("default")
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		exutil.SkipNoOLMCore(oc)
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

		// Dynamically add collect-profiles policy if the pods exist
		if _, err := oc.AsAdmin().WithoutNamespace().
			Run("get").
			Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-collect-profiles").
			Output(); err == nil {
			policies = append(policies, olmv0util.NpExpecter{
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
			})
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

	g.It("PolarionID:83583-[OTP][Skipped:Disconnected]olmv0 networkpolicy on hosted hypershift", g.Label("ReleaseGate"), g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 should PolarionID:83583-[Skipped:Disconnected]olmv0 networkpolicy on hosted hypershift"), func() {

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

		// Dynamically add collect-profiles policy if the pods exist
		if _, err := oc.AsAdmin().WithoutNamespace().
			Run("get").
			Args("pods", "-n", "openshift-operator-lifecycle-manager", "-l", "app=olm-collect-profiles").
			Output(); err == nil {
			policies = append(policies, olmv0util.NpExpecter{
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
			})
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
