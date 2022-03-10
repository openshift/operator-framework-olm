package e2e

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/operator-lifecycle-manager/test/e2e/ctx"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MagicCatalog", func() {
	var (
		generatedNamespace corev1.Namespace
	)

	BeforeEach(func() {
		generatedNamespace = SetupGeneratedTestNamespace(genName("magic-catalog-e2e-"))
	})

	AfterEach(func() {
		TeardownNamespace(generatedNamespace.GetName())
	})

	It("Deploys and Undeploys a File-based Catalog", func() {
		// create dependencies
		const catalogName = "test"
		namespace := generatedNamespace.GetName()
		kubeClient := ctx.Ctx().Client()

		// The following path to fbc_catalog.json works on the downstream only
		// This has been changed as a downstream only patch for now, but an upstream tracking issue has also been
		// created: https://github.com/operator-framework/operator-lifecycle-manager/issues/2687
		provider, err := NewFileBasedFiledBasedCatalogProvider("testdata/fbc_catalog.json")
		Expect(err).To(BeNil())

		// create and deploy and undeploy the magic catalog
		magicCatalog := NewMagicCatalog(kubeClient, namespace, catalogName, provider)

		// deployment blocks until the catalog source has reached a READY status
		Expect(magicCatalog.DeployCatalog(context.TODO())).To(BeNil())
		Expect(magicCatalog.UndeployCatalog(context.TODO())).To(BeNil())
	})
})
