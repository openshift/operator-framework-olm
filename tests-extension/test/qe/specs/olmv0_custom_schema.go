package specs

import (
	"context"
	"os"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

const (
	grpcRequestTimeout = 60 * time.Second
)

var _ = g.Describe("[sig-operator][Jira:OLM][OCPFeatureGate:OLMLifecycleAndCompatibility] OLMv0 custom schema gRPC endpoint", g.Label("NonHyperShiftHOST"), func() {
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

	g.It("ExperimentalListPackageCustomSchemas returns custom schema FBC", g.Label("ReleaseGate"), func() {
		namespace := oc.Namespace()
		catalogName := "custom-schema-" + exutil.GetRandomString()
		itName := g.CurrentSpecReport().FullText()

		g.By("get opm base image from catalog-operator deployment")
		baseImage := olmv0util.GetOPMBaseImage(oc)
		e2e.Logf("opm base image: %s", baseImage)
		o.Expect(baseImage).NotTo(o.BeEmpty())

		g.By("build custom catalog image with custom schema FBC in-cluster")
		fbcContent, err := os.ReadFile(exutil.FixturePath("testdata", "custom-schema", "index.json"))
		o.Expect(err).NotTo(o.HaveOccurred())

		imageRef := olmv0util.BuildCustomCatalogImage(oc, namespace, catalogName, baseImage, fbcContent)
		e2e.Logf("built catalog image: %s", imageRef)

		// Register build resources for cleanup
		dr.GetIr(itName).Add(olmv0util.NewResource(oc, "buildconfig", catalogName, exutil.RequireNS, namespace))
		dr.GetIr(itName).Add(olmv0util.NewResource(oc, "imagestream", catalogName, exutil.RequireNS, namespace))

		g.By("create CatalogSource and wait for READY")
		catsrc := olmv0util.CatalogSourceDescription{
			Name:       catalogName,
			Namespace:  namespace,
			SourceType: "grpc",
			Address:    imageRef,
			Template:   exutil.FixturePath("testdata", "olm", "catalogsource-image.yaml"),
		}
		catsrc.CreateWithCheck(oc, itName, dr)

		g.By("port-forward to catalog pod gRPC endpoint")
		grpcAddr, cleanupPF := olmv0util.PortForwardToCatalogPod(oc, namespace, catalogName)
		defer cleanupPF()
		e2e.Logf("gRPC address: %s", grpcAddr)

		ctx, cancel := context.WithTimeout(context.Background(), grpcRequestTimeout)
		defer cancel()

		g.By("query with valid schema and package returns expected results")
		results, err := olmv0util.ListPackageCustomSchemas(ctx, grpcAddr, "custom.operator.io", "test-custom-pkg")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(results).To(o.HaveLen(2), "expected 2 custom schema blobs for custom.operator.io/test-custom-pkg")
		e2e.Logf("results for schema=custom.operator.io pkg=test-custom-pkg: %v", results)

		foundNames := map[string]bool{}
		for _, r := range results {
			name, ok := r["name"].(string)
			o.Expect(ok).To(o.BeTrue(), "expected 'name' field in result")
			foundNames[name] = true
			o.Expect(r["schema"]).To(o.Equal("custom.operator.io"))
			o.Expect(r["package"]).To(o.Equal("test-custom-pkg"))
		}
		o.Expect(foundNames).To(o.HaveKey("custom-metadata-1"))
		o.Expect(foundNames).To(o.HaveKey("custom-metadata-2"))

		g.By("query with valid schema and nonexistent package returns empty")
		results, err = olmv0util.ListPackageCustomSchemas(ctx, grpcAddr, "custom.operator.io", "no-such-pkg")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(results).To(o.BeEmpty(), "expected empty results for nonexistent package")

		g.By("query with nonexistent schema returns empty")
		results, err = olmv0util.ListPackageCustomSchemas(ctx, grpcAddr, "no.such.schema", "test-custom-pkg")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(results).To(o.BeEmpty(), "expected empty results for nonexistent schema")

		g.By("query with valid schema and empty package returns packageless results")
		results, err = olmv0util.ListPackageCustomSchemas(ctx, grpcAddr, "custom.operator.io", "")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(results).To(o.HaveLen(1), "expected 1 packageless custom schema blob for custom.operator.io")
		o.Expect(results[0]["name"]).To(o.Equal("packageless-metadata"))
		o.Expect(results[0]["schema"]).To(o.Equal("custom.operator.io"))

		data, ok := results[0]["data"].(map[string]interface{})
		o.Expect(ok).To(o.BeTrue(), "expected 'data' field to be a map")
		o.Expect(data["key"]).To(o.Equal("global"))

		g.By("query without x-acknowledge-experimental header returns empty")
		results, err = olmv0util.ListPackageCustomSchemasWithoutExperimentalHeader(ctx, grpcAddr, "custom.operator.io", "test-custom-pkg")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(results).To(o.BeEmpty(), "expected empty results without experimental header")
	})
})
