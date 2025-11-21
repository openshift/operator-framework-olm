package specs

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	olmv0util "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/olmv0util"
)

var _ = g.Describe("[sig-operator][Jira:OLM] OLM v0 for stress", func() {
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

	g.It("PolarionID:80299-[OTP][Skipped:Disconnected][OlmStress]create mass operator to see if they all are installed successfully with different ns [Slow][Timeout:180m]", g.Label("StressTest"), g.Label("NonHyperShiftHOST"), func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			caseID              = "80299"
			ns                  = "openshift-operator-lifecycle-manager"
			catalogLabel        = "app=catalog-operator"
			olmLabel            = "app=olm-operator"

			og = olmv0util.OperatorGroupDescription{
				Name:      "og-singlenamespace",
				Namespace: "",
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "local-storage-operator",
				Namespace:              "",
				Channel:                "stable",
				IpApproval:             "Automatic",
				OperatorPackage:        "local-storage-operator",
				CatalogSourceName:      "redhat-operators",
				CatalogSourceNamespace: "openshift-marketplace",
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)
		catsrcName := olmv0util.GetCatsrc(oc, "redhat-operators", "local-storage-operator")
		if len(catsrcName) == 0 {
			g.Skip("there is no package local-storage-operator")
		}
		e2e.Logf("the catsrc is %v", catsrcName)
		sub.CatalogSourceName = catsrcName
		startTime := time.Now().UTC()
		e2e.Logf("Start time: %s", startTime.Format(time.RFC3339))

		for i := 0; i < 45; i++ {
			e2e.Logf("=================it is round %v=================", i)
			func() {
				seed := time.Now().UnixNano()
				r := rand.New(rand.NewSource(seed))
				randomNum := r.Intn(5) + 5
				e2e.Logf("=================round %v has %v namespaces =================", i, randomNum)
				namespaces := []string{}
				for j := 0; j < randomNum; j++ {
					namespaces = append(namespaces, "olm-stress-"+exutil.GetRandomString())
				}

				for _, nsName := range namespaces {
					g.By(fmt.Sprintf("create ns %s, and then install og and sub", nsName))
					err := oc.AsAdmin().WithoutNamespace().Run("create").Args("ns", nsName).Execute()
					o.Expect(err).NotTo(o.HaveOccurred())
					defer func(ns string) {
						_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("ns", ns, "--force", "--grace-period=0", "--wait=false").Execute()
					}(nsName)
					og.Namespace = nsName
					og.Create(oc, itName, dr)
					sub.Namespace = nsName
					sub.CreateWithoutCheckNoPrint(oc, itName, dr)
				}
				for _, nsName := range namespaces {
					g.By(fmt.Sprintf("find the installed csv ns %s", nsName))
					sub.Namespace = nsName
					sub.FindInstalledCSV(oc, itName, dr)
				}
				for _, nsName := range namespaces {
					g.By(fmt.Sprintf("check the installed csv is ok in %s", nsName))
					sub.Namespace = nsName

					errWait := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 150*time.Second, false, func(ctx context.Context) (bool, error) {
						phase, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}").Output()
						if err != nil {
							if strings.Contains(phase, "NotFound") || strings.Contains(phase, "No resources found") {
								e2e.Logf("the existing csv does not exist, and try to get new csv")
								sub.FindInstalledCSV(oc, itName, dr)
							} else {
								e2e.Logf("the error: %v, and try next", err)
							}
							return false, nil
						}

						e2e.Logf("---> we expect value: %s, in returned value: %s", "Succeeded+2+InstallSucceeded", phase)
						if strings.Compare(phase, "Succeeded") == 0 || strings.Compare(phase, "InstallSucceeded") == 0 {
							e2e.Logf("the output %s matches one of the content %s, expected", phase, "Succeeded+2+InstallSucceeded")
							return true, nil
						}
						e2e.Logf("the output %s does not match one of the content %s, unexpected", phase, "Succeeded+2+InstallSucceeded")
						return false, nil
					})
					if errWait != nil {
						olmv0util.GetResource(oc, true, true, "pod", "-n", "openshift-marketplace")
						olmv0util.GetResource(oc, true, true, "operatorgroup", "-n", sub.Namespace, "-o", "yaml")
						olmv0util.GetResource(oc, true, true, "subscription", "-n", sub.Namespace, "-o", "yaml")
						olmv0util.GetResource(oc, true, true, "installplan", "-n", sub.Namespace)
						olmv0util.GetResource(oc, true, true, "csv", "-n", sub.Namespace)
						olmv0util.GetResource(oc, true, true, "pods", "-n", sub.Namespace)
					}
					exutil.AssertWaitPollNoErr(errWait, fmt.Sprintf("expected content %s not found by %v", "Succeeded+2+InstallSucceeded", strings.Join([]string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}, " ")))

				}

			}()
		}
		endTime := time.Now().UTC()
		e2e.Logf("End time:  %v", endTime.Format(time.RFC3339))

		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		if minutes < 1 {
			minutes = 1
		}

		podName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", catalogLabel, "-o=jsonpath={.items[0].metadata.name}", "-n", ns).Output()
		if err == nil {
			if !exutil.WriteErrToArtifactDir(oc, ns, podName, "error", "Unhandled|Reconciler error|try again|level=info|warning", caseID, minutes) {
				e2e.Logf("no error log into artifact for pod %s in %s", podName, ns)
			}
		}
		podName, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", olmLabel, "-o=jsonpath={.items[0].metadata.name}", "-n", ns).Output()
		if err == nil {
			if !exutil.WriteErrToArtifactDir(oc, ns, podName, "error", "Unhandled|Reconciler error|level=info|warning|update Operator status|no errors|warning", caseID, minutes) {
				e2e.Logf("no error log into artifact for pod %s in %s", podName, ns)
			}
		}

		if !exutil.IsPodReady(oc, ns, catalogLabel) {
			olmv0util.GetResource(oc, true, true, "pod", "-n", ns, "-l", catalogLabel, "-o", "yaml")
			exutil.AssertWaitPollNoErr(fmt.Errorf("the pod with %s is not correct", catalogLabel), "the pod with app=catalog-operator is not correct")
		}
		if !exutil.IsPodReady(oc, ns, olmLabel) {
			olmv0util.GetResource(oc, true, true, "pod", "-n", ns, "-l", olmLabel, "-o", "yaml")
			exutil.AssertWaitPollNoErr(fmt.Errorf("the pod with %s is not correct", olmLabel), "the pod with app=olm-operator is not correct")
		}
	})

	g.It("PolarionID:80413-[OTP][Skipped:Disconnected][OlmStress]install operator repeatedly serially with same ns [Slow][Timeout:180m]", g.Label("StressTest"), g.Label("NonHyperShiftHOST"), func() {
		var (
			itName              = g.CurrentSpecReport().FullText()
			buildPruningBaseDir = exutil.FixturePath("testdata", "olm")
			ogSingleTemplate    = filepath.Join(buildPruningBaseDir, "operatorgroup.yaml")
			catsrcImageTemplate = filepath.Join(buildPruningBaseDir, "catalogsource-image.yaml")
			subTemplate         = filepath.Join(buildPruningBaseDir, "olm-subscription.yaml")
			caseID              = "80413"
			ns                  = "openshift-must-gather-operator"
			nsOlm               = "openshift-operator-lifecycle-manager"
			catalogLabel        = "app=catalog-operator"
			olmLabel            = "app=olm-operator"

			catsrc = olmv0util.CatalogSourceDescription{
				Name:        "catsrc-80413",
				Namespace:   ns,
				DisplayName: "Test 80413",
				Publisher:   "OLM QE",
				SourceType:  "grpc",
				Address:     "quay.io/app-sre/must-gather-operator-registry@sha256:0a0610e37a016fb4eed1b000308d840795838c2306f305a151c64cf3b4fd6bb4",
				Template:    catsrcImageTemplate,
			}
			og = olmv0util.OperatorGroupDescription{
				Name:      "og",
				Namespace: ns,
				Template:  ogSingleTemplate,
			}
			sub = olmv0util.SubscriptionDescription{
				SubName:                "must-gather-operator",
				Namespace:              ns,
				Channel:                "stable",
				IpApproval:             "Automatic",
				OperatorPackage:        "must-gather-operator",
				CatalogSourceName:      "catsrc-80413",
				CatalogSourceNamespace: ns,
				StartingCSV:            "",
				CurrentCSV:             "",
				InstalledCSV:           "",
				Template:               subTemplate,
				SingleNamespace:        true,
			}
		)

		startTime := time.Now().UTC()
		e2e.Logf("Start time: %s", startTime.Format(time.RFC3339))

		for i := 0; i < 150; i++ {
			e2e.Logf("=================it is round %v=================", i)
			func() {
				g.By(fmt.Sprintf("create ns %s", ns))
				err := oc.AsAdmin().WithoutNamespace().Run("create").Args("ns", ns).Execute()
				o.Expect(err).NotTo(o.HaveOccurred())
				defer func() {
					_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("ns", ns).Execute()
				}()

				g.By(fmt.Sprintf("install catsrc in %s", ns))
				defer catsrc.Delete(itName, dr)
				catsrc.Create(oc, itName, dr)

				g.By(fmt.Sprintf("install og in %s", ns))
				og.Create(oc, itName, dr)

				g.By(fmt.Sprintf("install sub in %s", ns))
				sub.CreateWithoutCheckNoPrint(oc, itName, dr)

				g.By(fmt.Sprintf("find the installed csv ns %s", ns))
				sub.FindInstalledCSV(oc, itName, dr)

				g.By(fmt.Sprintf("check the installed csv is ok in %s", ns))
				olmv0util.NewCheck("expect", true, true, true, "Succeeded+2+InstallSucceeded", true, []string{"csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).Check(oc)

			}()
		}

		endTime := time.Now().UTC()
		e2e.Logf("End time:  %v", endTime.Format(time.RFC3339))

		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		if minutes < 1 {
			minutes = 1
		}

		podName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", catalogLabel, "-o=jsonpath={.items[0].metadata.name}", "-n", nsOlm).Output()
		if err == nil {
			if !exutil.WriteErrToArtifactDir(oc, nsOlm, podName, "error", "Unhandled|Reconciler error|try again|level=info|warning", caseID, minutes) {
				e2e.Logf("no error log into artifact for pod %s in %s", podName, nsOlm)
			}
		}
		podName, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-l", olmLabel, "-o=jsonpath={.items[0].metadata.name}", "-n", nsOlm).Output()
		if err == nil {
			if !exutil.WriteErrToArtifactDir(oc, nsOlm, podName, "error", "Unhandled|Reconciler error|level=info|warning|update Operator status|no errors|warning", caseID, minutes) {
				e2e.Logf("no error log into artifact for pod %s in %s", podName, nsOlm)
			}
		}

		if !exutil.IsPodReady(oc, nsOlm, catalogLabel) {
			olmv0util.GetResource(oc, true, true, "pod", "-n", nsOlm, "-l", catalogLabel, "-o", "yaml")
			exutil.AssertWaitPollNoErr(fmt.Errorf("the pod with %s is not correct", catalogLabel), "the pod with app=catalog-operator is not correct")
		}
		if !exutil.IsPodReady(oc, nsOlm, olmLabel) {
			olmv0util.GetResource(oc, true, true, "pod", "-n", nsOlm, "-l", olmLabel, "-o", "yaml")
			exutil.AssertWaitPollNoErr(fmt.Errorf("the pod with %s is not correct", olmLabel), "the pod with app=olm-operator is not correct")
		}
	})

})
