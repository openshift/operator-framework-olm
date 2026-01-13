package specs

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 webhook operator", func() {
	defer g.GinkgoRecover()

	var oc = exutil.NewCLIWithoutNamespace("default")

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)
		oc.SetupProject()
		exutil.SkipNoOLMCore(oc)
	})

	g.It("PolarionID:83979-[Skipped:Disconnected]ClusterExtension installs webhook-operator and webhooks work", func() {
		exutil.SkipNoOLMv1Core(oc)
		olmv0util.ValidateAccessEnvironment(oc)

		baseDir := exutil.FixturePath("testdata", "olm")
		clusterCatalogTemplate := filepath.Join(baseDir, "clustercatalog-webhook-operator.yaml")
		clusterExtensionTemplate := filepath.Join(baseDir, "clusterextension-webhook-operator.yaml")
		webhookCRTemplate := filepath.Join(baseDir, "cr-webhookTest.yaml")

		suffix := exutil.GetRandomString()
		namespace := "webhook-operator-" + suffix
		catalogName := "webhook-operator-catalog-" + suffix
		extensionName := "webhook-operator-" + suffix
		crbName := "webhook-operator-installer-" + suffix

		serviceAccountName := "webhook-operator-installer"
		secretName := "webhook-operator-webhook-service-cert"
		catalogImage := "quay.io/olmqe/webhook-operator-index:0.0.3-v1-cache"
		packageName := "webhook-operator"
		packageVersion := "0.0.1"

		defer func() {
			_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("namespace", namespace, "--ignore-not-found").Execute()
		}()
		defer func() {
			_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("clusterrolebinding", crbName, "--ignore-not-found").Execute()
		}()
		defer func() {
			_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("clustercatalog", catalogName, "--ignore-not-found").Execute()
		}()
		defer func() {
			_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("clusterextension", extensionName, "--ignore-not-found").Execute()
		}()

		g.By("Create a ClusterCatalog for the webhook-operator index")
		err := olmv0util.ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", clusterCatalogTemplate,
			"-p", "NAME="+catalogName, "IMAGE="+catalogImage)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Create a ClusterExtension for the webhook-operator package")
		err = olmv0util.ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", clusterExtensionTemplate,
			"-p", "EXTENSION_NAME="+extensionName, "NAMESPACE="+namespace, "SERVICE_ACCOUNT="+serviceAccountName,
			"CLUSTER_ROLE_BINDING="+crbName, "CATALOG_LABEL="+catalogName,
			"PACKAGE_NAME="+packageName, "VERSION="+packageVersion)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Wait for the webhooktest API to be available")
		err = wait.PollUntilContextTimeout(context.TODO(), 15*time.Second, 6*time.Minute, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("api-resources").Args("-o", "name").Output()
			if err != nil {
				e2e.Logf("failed to list api-resources: %v", err)
				return false, nil
			}
			return strings.Contains(output, "webhooktests.webhook.operators.coreos.io"), nil
		})
		exutil.AssertWaitPollNoErr(err, "webhooktests.webhook.operators.coreos.io does not exist")

		g.By("Wait for webhook-operator deployments to be ready")
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
			deployments, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployments", "-n", namespace, "-o=jsonpath={.items[*].metadata.name}").Output()
			if err != nil {
				return false, nil
			}
			deployments = strings.TrimSpace(deployments)
			if deployments == "" {
				return false, nil
			}
			for _, name := range strings.Fields(deployments) {
				available, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment", name, "-n", namespace, "-o=jsonpath={.status.availableReplicas}").Output()
				if err != nil || strings.TrimSpace(available) == "" || strings.TrimSpace(available) == "0" {
					return false, nil
				}
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, "webhook-operator deployments are not ready")

		processWebhook := func(name string, valid bool) string {
			validValue := "false"
			if valid {
				validValue = "true"
			}
			configFile, err := oc.AsAdmin().Run("process").Args("--ignore-unknown-parameters=true", "-f", webhookCRTemplate,
				"-p", "NAME="+name, "NAMESPACE="+namespace, "VALID="+validValue).OutputToFile("webhooktest-" + exutil.GetRandomString() + ".json")
			o.Expect(err).NotTo(o.HaveOccurred())
			return configFile
		}

		g.By("Verify the validating webhook rejects invalid resources")
		invalidName := "validating-webhook-test-" + suffix
		invalidConfig := processWebhook(invalidName, false)
		expectedError := "Spec.Valid must be true"
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
			output, err := oc.AsAdmin().WithoutNamespace().Run("apply").Args("--dry-run=server", "-f", invalidConfig).Output()
			if err == nil {
				e2e.Logf("expected validating webhook to deny invalid resource")
				return false, nil
			}
			if strings.Contains(output, expectedError) && strings.Contains(output, "denied the request") {
				return true, nil
			}
			e2e.Logf("unexpected validation error output: %s", output)
			return false, nil
		})
		exutil.AssertWaitPollNoErr(err, "validating webhook did not reject invalid webhooktest")

		verifyWebhookV2 := func(name string) {
			waitErr := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
				apiVersion, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("webhooktest", name, "-n", namespace, "-o=jsonpath={.apiVersion}").Output()
				if err != nil {
					return false, nil
				}
				if strings.TrimSpace(apiVersion) != "webhook.operators.coreos.io/v2" {
					return false, nil
				}
				mutate, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("webhooktest", name, "-n", namespace, "-o=jsonpath={.spec.conversion.mutate}").Output()
				if err != nil {
					return false, nil
				}
				valid, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("webhooktest", name, "-n", namespace, "-o=jsonpath={.spec.conversion.valid}").Output()
				if err != nil {
					return false, nil
				}
				return strings.TrimSpace(mutate) == "true" && strings.TrimSpace(valid) == "true", nil
			})
			exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("webhooktest %s was not converted to v2 with expected fields", name))
		}

		g.By("Verify the mutating webhook updates valid resources")
		mutatingName := "mutating-webhook-test-" + suffix
		mutatingConfig := processWebhook(mutatingName, true)
		output, err := oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", mutatingConfig).Output()
		o.Expect(err).NotTo(o.HaveOccurred(), output)
		verifyWebhookV2(mutatingName)

		g.By("Verify the conversion webhook serves v2 resources")
		conversionName := "conversion-webhook-test-" + suffix
		conversionConfig := processWebhook(conversionName, true)
		output, err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", conversionConfig).Output()
		o.Expect(err).NotTo(o.HaveOccurred(), output)
		verifyWebhookV2(conversionName)

		g.By("Verify webhook service cert secret is recreated after deletion")
		var originalUID string
		err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
			uid, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("secret", secretName, "-n", namespace, "-o=jsonpath={.metadata.uid}").Output()
			if err != nil || strings.TrimSpace(uid) == "" {
				return false, nil
			}
			originalUID = uid
			return true, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("secret %s not found before deletion", secretName))

		err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("secret", secretName, "-n", namespace).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
			uid, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("secret", secretName, "-n", namespace, "-o=jsonpath={.metadata.uid}").Output()
			if err != nil || strings.TrimSpace(uid) == "" {
				return false, nil
			}
			return uid != originalUID, nil
		})
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("secret %s was not recreated", secretName))
	})
})
