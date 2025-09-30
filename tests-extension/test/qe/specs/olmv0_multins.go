package specs

import (
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

// it is mapping to the Describe "OLM for an end user handle to support" in olm.go
var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0 with multi ns", func() {
	defer g.GinkgoRecover()

	var (
		oc = exutil.NewCLI("olm-multi-"+exutil.GetRandomString(), exutil.KubeConfigPath())

		dr = make(olmv0util.DescriberResrouce)
	)

	g.BeforeEach(func() {
		exutil.SkipMicroshift(oc)

		exutil.SkipNoOLMCore(oc)
		itName := g.CurrentSpecReport().FullText()
		dr.AddIr(itName)
	})

	g.AfterEach(func() {})

	g.It("PolarionID:22226-[Skipped:Disconnected]the csv without support MultiNamespace fails for og with MultiNamespace", func() {
		var (
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			cmNcTemplate        = filepath.Join(buildPruningBaseDir, "cm-namespaceconfig.yaml")
			catsrcCmTemplate    = filepath.Join(buildPruningBaseDir, "catalogsource-configmap.yaml")
			ogMultiTemplate     = filepath.Join(buildPruningBaseDir, "og-multins.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			itName              = g.CurrentSpecReport().FullText()
			og                  = olmv0util.OperatorGroupDescription{
				Name:         "og-multinamespace",
				Namespace:    "",
				Multinslabel: "olmtestmultins",
				Template:     ogMultiTemplate,
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
			p1 = olmv0util.ProjectDescription{
				Name:            "olm-enduser-multins-csv-1-fail",
				TargetNamespace: "",
			}
			p2 = olmv0util.ProjectDescription{
				Name:            "olm-enduser-multins-csv-2-fail",
				TargetNamespace: "",
			}
		)

		defer p1.Delete(oc)
		defer p2.Delete(oc)
		cm.Namespace = oc.Namespace()
		catsrc.Namespace = oc.Namespace()
		sub.Namespace = oc.Namespace()
		sub.CatalogSourceNamespace = catsrc.Namespace
		og.Namespace = oc.Namespace()
		p1.TargetNamespace = oc.Namespace()
		p2.TargetNamespace = oc.Namespace()
		g.By("Create new project")
		p1.Create(oc, itName, dr)
		p1.Label(oc, "olmtestmultins")
		p2.Create(oc, itName, dr)
		p2.Label(oc, "olmtestmultins")

		g.By("Create cm")
		cm.Create(oc, itName, dr)

		g.By("Create catalog source")
		catsrc.Create(oc, itName, dr)

		g.By("Create og")
		og.Create(oc, itName, dr)

		g.By("Create sub")
		sub.Create(oc, itName, dr)
		olmv0util.NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Contain, "MultiNamespace InstallModeType not supported", exutil.Ok, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.message}"}).Check(oc)
	})

})
