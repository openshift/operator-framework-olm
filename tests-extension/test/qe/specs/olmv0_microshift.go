package specs

import (
	"context"
	"fmt"
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// it is mapping to olm_microshift.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 on microshift", g.Label("NonHyperShiftHOST"), func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLIWithoutNamespace("default")

		dr = make(olmv0util.DescriberResrouce)
	)

	g.BeforeEach(func() {
		if !exutil.IsMicroshiftCluster(oc) {
			g.Skip("it is not microshift, so skip it.")
		}

		_, errCheckOlm := oc.AdminKubeClient().CoreV1().Namespaces().Get(context.Background(),
			"openshift-operator-lifecycle-manager", metav1.GetOptions{})
		if errCheckOlm != nil {
			if apierrors.IsNotFound(errCheckOlm) {
				g.Skip("there is no olm installed on microshift, so skip it")
			} else {
				o.Expect(errCheckOlm).NotTo(o.HaveOccurred())
			}
		}
		dr.AddIr(g.CurrentSpecReport().FullText())

	})

	g.It("PolarionID:69867-[OTP][Skipped:Disconnected]deployed in microshift and install one operator with single mode.", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 on microshift PolarionID:69867-[Skipped:Disconnected]deployed in microshift and install one operator with single mode."), func() {

		var (
			itName              = g.CurrentSpecReport().FullText()
			namespace           = "olm-mcroshift-69867"
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm", "microshift")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "og-single.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-restricted.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

			og = olmv0util.OperatorGroupDescription{
				Name:        "og-singlenamespace",
				Namespace:   namespace,
				Template:    ogSingleTemplate,
				ClusterType: "microshift",
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catalog",
				Namespace:   namespace,
				DisplayName: "Test Catsrc Operators",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:v1399-fbc-multi",
				Template:    catsrcImageTemplate,
				ClusterType: "microshift",
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok1-1399",
				Namespace:              namespace,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok1-1399",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: catsrc.Namespace,
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
				ClusterType:            "microshift",
			}
		)

		g.By("check olm related CRD")
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("crd").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.And(
			o.ContainSubstring("catalogsources.operators.coreos.com"),
			o.ContainSubstring("clusterserviceversions.operators.coreos.com"),
			o.ContainSubstring("installplans.operators.coreos.com"),
			o.ContainSubstring("olmconfigs.operators.coreos.com"),
			o.ContainSubstring("operatorconditions.operators.coreos.com"),
			o.ContainSubstring("operatorgroups.operators.coreos.com"),
			o.ContainSubstring("operators.operators.coreos.com"),
			o.ContainSubstring("subscriptions.operators.coreos.com"),
		), "some CRDs do not exist")

		g.By("the olm and catalog pod is running")
		output, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", "openshift-operator-lifecycle-manager",
			"-l", "app=olm-operator").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("Running"), "olm pod is not running")
		output, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", "openshift-operator-lifecycle-manager",
			"-l", "app=catalog-operator").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.ContainSubstring("Running"), "catalog pod is not running")

		g.By("create namespace")
		defer func() {
			_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("namespace", namespace, "--ignore-not-found").Execute()
		}()
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("namespace", namespace).Execute()
		o.Expect(err).NotTo(o.HaveOccurred(), fmt.Sprintf("Failed to create namespace/%s", namespace))

		g.By("Create opertor group")
		defer og.Delete(itName, dr)
		og.Create(oc, itName, dr)

		g.By("Create catalog")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("install operator")
		defer sub.DeleteCSV(itName, dr)
		defer sub.Delete(itName, dr)
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Succeeded+2+Installing-TIME-WAIT-300s", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

		g.By("check operator")
		output, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("operators.operators.coreos.com",
			sub.OperatorPackage+"."+namespace, "-o=jsonpath={.status}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(output).To(o.And(
			o.ContainSubstring("ClusterRole"),
			o.ContainSubstring("ClusterRoleBinding"),
			o.ContainSubstring("ClusterServiceVersion"),
			o.ContainSubstring("CustomResourceDefinition"),
			o.ContainSubstring("Deployment"),
			o.ContainSubstring("OperatorCondition"),
			o.ContainSubstring("Subscription"),
		), "some resources do not exist")

	})

	g.It("PolarionID:69868-[OTP][Skipped:Disconnected]olm microshift install operator with all mode muilt og error and delete one og to get it installed.", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 on microshift PolarionID:69868-[Skipped:Disconnected]olm microshift install operator with all mode muilt og error and delete one og to get it installed."), func() {

		var (
			itName              = g.CurrentSpecReport().FullText()
			namespace           = "openshift-operators"
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm", "microshift")
			ogAllTemplate       = filepath.Join(buildPruningBaseDir, "og-all.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image-restricted.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")

			og = olmv0util.OperatorGroupDescription{
				Name:        "og-all",
				Namespace:   namespace,
				Template:    ogAllTemplate,
				ClusterType: "microshift",
			}
			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catalog-all",
				Namespace:   "openshift-marketplace",
				DisplayName: "Test Catsrc Operators",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/olmqe/nginx-ok-index:v1399-fbc-multi",
				Template:    catsrcImageTemplate,
				ClusterType: "microshift",
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "nginx-ok2-1399",
				Namespace:              namespace,
				Channel:                "alpha",
				IpApproval:             "Automatic",
				OperatorPackage:        "nginx-ok2-1399",
				CatalogSourceName:      catsrc.Name,
				CatalogSourceNamespace: catsrc.Namespace,
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        false,
				ClusterType:            "microshift",
			}
		)

		g.By("check og in openshift-operators already")
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("og", "global-operators", "-n", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("it is \n: %v", output)

		g.By("Create opertor group")
		defer og.Delete(itName, dr)
		og.Create(oc, itName, dr)

		g.By("Create catalog")
		defer catsrc.Delete(itName, dr)
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("install operator with multi og")
		defer sub.DeleteCSV(itName, dr)
		defer sub.Delete(itName, dr)
		sub.CreateWithoutCheck(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "", exutil.Ok, []string{"sub", sub.SubName, "-n", namespace, "-o=jsonpath={.status.installedCSV}"}).Check(oc)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "MultipleOperatorGroupsFound", exutil.Ok, []string{"og", og.Name, "-n", namespace, "-o=jsonpath={.status}"}).Check(oc)

		g.By("delete more og")
		og.Delete(itName, dr)

		g.By("operator is installed")
		sub.FindInstalledCSV(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, sub.InstalledCSV, exutil.Ok, []string{"csv", "-n", "default"}).Check(oc)

	})

	g.It("PolarionID:83581-[OTP][Skipped:Disconnected]olmv0 networkpolicy on microshift.", g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 on microshift PolarionID:83581-[Skipped:Disconnected]olmv0 networkpolicy on microshift."), func() {

		policies := []olmv0util.NpExpecter{
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
				Name:      "default-allow-all",
				Namespace: "openshift-operators",
				ExpectIngress: []olmv0util.IngressRule{
					{
						Ports:     []olmv0util.Port{{Port: nil, Protocol: ""}},
						Selectors: nil,
					},
				},
				ExpectEgress: []olmv0util.EgressRule{
					{
						Ports:     []olmv0util.Port{{Port: nil, Protocol: ""}},
						Selectors: nil,
					},
				},
				ExpectSelector:    map[string]string{},
				ExpectPolicyTypes: []string{"Ingress", "Egress"},
			},
			// after https://issues.redhat.com/browse/OCPBUGS-59566 is fixed, more checkers are added here
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
