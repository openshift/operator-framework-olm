package e2e

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"

	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	packageserverclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	networkingv1ac "k8s.io/client-go/applyconfigurations/networking/v1"

	"github.com/operator-framework/api/pkg/lib/version"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/catalogtemplate"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/catalogsource"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/ownerutil"
	"github.com/operator-framework/operator-lifecycle-manager/test/e2e/ctx"
)

const (
	openshiftregistryFQDN = "image-registry.openshift-image-registry.svc:5000"
	catsrcImage           = "docker://quay.io/olmtest/catsrc-update-test:"
	badCSVDir             = "bad-csv"
)

var _ = Describe("Starting CatalogSource e2e tests", Label("CatalogSource"), func() {
	var (
		generatedNamespace  corev1.Namespace
		c                   operatorclient.ClientInterface
		crc                 versioned.Interface
		packageserverClient *packageserverclientset.Clientset
	)

	BeforeEach(func() {
		// In OCP, PSA labels for any namespace created that is not prefixed with "openshift-" is overridden to enforce
		// PSA restricted. This test namespace needs to prefixed with openshift- so that baseline/privileged enforcement
		// for the PSA specific tests are not overridden,
		// Change it only after https://github.com/operator-framework/operator-lifecycle-manager/issues/2859 is closed.
		namespaceName := genName("openshift-catsrc-e2e-")
		og := operatorsv1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-operatorgroup", namespaceName),
				Namespace: namespaceName,
			},
		}
		generatedNamespace = SetupGeneratedTestNamespaceWithOperatorGroup(namespaceName, og)
		c = ctx.Ctx().KubeClient()
		crc = ctx.Ctx().OperatorClient()
		packageserverClient = packageserverclientset.NewForConfigOrDie(ctx.Ctx().RESTConfig())
	})

	AfterEach(func() {
		TeardownNamespace(generatedNamespace.GetName())
	})

	It("loading between restarts", func() {
		By("create a simple catalogsource")
		packageName := genName("nginx")
		stableChannel := "stable"
		packageStable := packageName + "-stable"
		manifests := []registry.PackageManifest{
			{
				PackageName: packageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: packageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		crd := newCRD(genName("ins-"))
		csv := newCSV(packageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{crd}, nil, nil)

		defer func() {
			Eventually(func() error {
				return ctx.Ctx().KubeClient().ApiextensionsInterface().ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), crd.GetName(), metav1.DeleteOptions{})
			}).Should(Or(Succeed(), WithTransform(k8serror.IsNotFound, BeTrue())))
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &csv))
			}).Should(Succeed())
		}()

		catalogSourceName := genName("mock-ocs-")
		_, cleanupSource := createInternalCatalogSource(c, crc, catalogSourceName, generatedNamespace.GetName(), manifests, []apiextensionsv1.CustomResourceDefinition{crd}, []v1alpha1.ClusterServiceVersion{csv})
		defer cleanupSource()

		By("ensure the mock catalog exists and has been synced by the catalog operator")
		catalogSource, err := fetchCatalogSourceOnStatus(crc, catalogSourceName, generatedNamespace.GetName(), catalogSourceRegistryPodSynced())
		Expect(err).ShouldNot(HaveOccurred())

		By("get catalog operator deployment")
		deployment, err := getOperatorDeployment(c, operatorNamespace, labels.Set{"app": "catalog-operator"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(deployment).ToNot(BeNil(), "Could not find catalog operator deployment")

		By("rescale catalog operator")
		By("Rescaling catalog operator...")
		err = rescaleDeployment(c, deployment)
		Expect(err).ShouldNot(HaveOccurred(), "Could not rescale catalog operator")
		By("Catalog operator rescaled")

		By("check for last synced update to catalogsource")
		By("Checking for catalogsource lastSync updates")
		_, err = fetchCatalogSourceOnStatus(crc, catalogSourceName, generatedNamespace.GetName(), func(cs *v1alpha1.CatalogSource) bool {
			before := catalogSource.Status.GRPCConnectionState
			after := cs.Status.GRPCConnectionState
			if after != nil && after.LastConnectTime.After(before.LastConnectTime.Time) {
				ctx.Ctx().Logf("lastSync updated: %s -> %s", before.LastConnectTime, after.LastConnectTime)
				return true
			}
			return false
		})
		Expect(err).ShouldNot(HaveOccurred(), "Catalog source changed after rescale")
		By("Catalog source successfully loaded after rescale")
	})

	It("global update triggers subscription sync", func() {
		mainPackageName := genName("nginx-")

		mainPackageStable := fmt.Sprintf("%s-stable", mainPackageName)
		mainPackageReplacement := fmt.Sprintf("%s-replacement", mainPackageStable)

		stableChannel := "stable"

		mainCRD := newCRD(genName("ins-"))
		mainCSV := newCSV(mainPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{mainCRD}, nil, nil)
		replacementCSV := newCSV(mainPackageReplacement, generatedNamespace.GetName(), mainPackageStable, semver.MustParse("0.2.0"), []apiextensionsv1.CustomResourceDefinition{mainCRD}, nil, nil)

		defer func() {
			Eventually(func() error {
				return ctx.Ctx().KubeClient().ApiextensionsInterface().ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), mainCRD.GetName(), metav1.DeleteOptions{})
			}).Should(Or(Succeed(), WithTransform(k8serror.IsNotFound, BeTrue())))
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &mainCSV))
			}).Should(Succeed())
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &replacementCSV))
			}).Should(Succeed())
		}()

		mainCatalogName := genName("mock-ocs-main-")

		By("Create separate manifests for each CatalogSource")
		mainManifests := []registry.PackageManifest{
			{
				PackageName: mainPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: mainPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		By("Create the initial catalog source")
		cs, cleanup := createInternalCatalogSource(c, crc, mainCatalogName, globalCatalogNamespace, mainManifests, []apiextensionsv1.CustomResourceDefinition{mainCRD}, []v1alpha1.ClusterServiceVersion{mainCSV})
		defer cleanup()

		By("Attempt to get the catalog source before creating install plan")
		_, err := fetchCatalogSourceOnStatus(crc, cs.GetName(), cs.GetNamespace(), catalogSourceRegistryPodSynced())
		Expect(err).ToNot(HaveOccurred())

		subscriptionSpec := &v1alpha1.SubscriptionSpec{
			CatalogSource:          cs.GetName(),
			CatalogSourceNamespace: cs.GetNamespace(),
			Package:                mainPackageName,
			Channel:                stableChannel,
			StartingCSV:            mainCSV.GetName(),
			InstallPlanApproval:    v1alpha1.ApprovalManual,
		}

		By("Create Subscription")
		subscriptionName := genName("sub-")
		createSubscriptionForCatalogWithSpec(GinkgoT(), crc, generatedNamespace.GetName(), subscriptionName, subscriptionSpec)

		subscription, err := fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionHasInstallPlanChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ToNot(BeNil())

		installPlanName := subscription.Status.Install.Name
		requiresApprovalChecker := buildInstallPlanPhaseCheckFunc(v1alpha1.InstallPlanPhaseRequiresApproval)

		Eventually(func() error {
			fetchedInstallPlan, err := fetchInstallPlanWithNamespace(GinkgoT(), crc, installPlanName, generatedNamespace.GetName(), requiresApprovalChecker)
			if err != nil {
				return err
			}

			fetchedInstallPlan.Spec.Approved = true
			_, err = crc.OperatorsV1alpha1().InstallPlans(generatedNamespace.GetName()).Update(context.Background(), fetchedInstallPlan, metav1.UpdateOptions{})
			return err
		}).Should(Succeed())

		_, err = fetchCSV(crc, generatedNamespace.GetName(), mainCSV.GetName(), csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())

		By("Update manifest")
		mainManifests = []registry.PackageManifest{
			{
				PackageName: mainPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: replacementCSV.GetName()},
				},
				DefaultChannelName: stableChannel,
			},
		}

		By("Update catalog configmap")
		updateInternalCatalog(GinkgoT(), c, crc, cs.GetName(), cs.GetNamespace(), []apiextensionsv1.CustomResourceDefinition{mainCRD}, []v1alpha1.ClusterServiceVersion{mainCSV, replacementCSV}, mainManifests)

		By("Get updated catalogsource")
		fetchedUpdatedCatalog, err := fetchCatalogSourceOnStatus(crc, cs.GetName(), cs.GetNamespace(), catalogSourceRegistryPodSynced())
		Expect(err).ShouldNot(HaveOccurred())

		subscription, err = fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionStateUpgradePendingChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())

		By("Ensure the timing")
		catalogConnState := fetchedUpdatedCatalog.Status.GRPCConnectionState
		subUpdatedTime := subscription.Status.LastUpdated
		Expect(subUpdatedTime.Time).Should(BeTemporally("<", catalogConnState.LastConnectTime.Add(60*time.Second)))
	})

	It("config map update triggers registry pod rollout", func() {

		mainPackageName := genName("nginx-")
		dependentPackageName := genName("nginxdep-")

		mainPackageStable := fmt.Sprintf("%s-stable", mainPackageName)
		dependentPackageStable := fmt.Sprintf("%s-stable", dependentPackageName)

		stableChannel := "stable"

		dependentCRD := newCRD(genName("ins-"))
		mainCSV := newCSV(mainPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), nil, []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil)
		dependentCSV := newCSV(dependentPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil, nil)

		defer func() {
			Eventually(func() error {
				return ctx.Ctx().KubeClient().ApiextensionsInterface().ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), dependentCRD.GetName(), metav1.DeleteOptions{})
			}).Should(Or(Succeed(), WithTransform(k8serror.IsNotFound, BeTrue())))
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &mainCSV))
			}).Should(Succeed())
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &dependentCSV))
			}).Should(Succeed())
		}()

		mainCatalogName := genName("mock-ocs-main-")

		By("Create separate manifests for each CatalogSource")
		mainManifests := []registry.PackageManifest{
			{
				PackageName: mainPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: mainPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		dependentManifests := []registry.PackageManifest{
			{
				PackageName: dependentPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: dependentPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		By("Create the initial catalogsource")
		createInternalCatalogSource(c, crc, mainCatalogName, generatedNamespace.GetName(), mainManifests, nil, []v1alpha1.ClusterServiceVersion{mainCSV})

		By("Attempt to get the catalog source before creating install plan")
		fetchedInitialCatalog, err := fetchCatalogSourceOnStatus(crc, mainCatalogName, generatedNamespace.GetName(), catalogSourceRegistryPodSynced())
		Expect(err).ShouldNot(HaveOccurred())

		By("Get initial configmap")
		configMap, err := c.KubernetesInterface().CoreV1().ConfigMaps(generatedNamespace.GetName()).Get(context.Background(), fetchedInitialCatalog.Spec.ConfigMap, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		By("Check pod created")
		initialPods, err := c.KubernetesInterface().CoreV1().Pods(generatedNamespace.GetName()).List(context.Background(), metav1.ListOptions{LabelSelector: "olm.configMapResourceVersion=" + configMap.ResourceVersion})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(initialPods.Items).To(HaveLen(1))

		By("Update catalog configmap")
		updateInternalCatalog(GinkgoT(), c, crc, mainCatalogName, generatedNamespace.GetName(), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, []v1alpha1.ClusterServiceVersion{mainCSV, dependentCSV}, append(mainManifests, dependentManifests...))

		fetchedUpdatedCatalog, err := fetchCatalogSourceOnStatus(crc, mainCatalogName, generatedNamespace.GetName(), func(catalog *v1alpha1.CatalogSource) bool {
			before := fetchedInitialCatalog.Status.ConfigMapResource
			after := catalog.Status.ConfigMapResource
			if after != nil && before.LastUpdateTime.Before(&after.LastUpdateTime) &&
				after.ResourceVersion != before.ResourceVersion {
				ctx.Ctx().Logf("catalog updated")
				return true
			}
			ctx.Ctx().Logf("waiting for catalog pod to be available")
			return false
		})
		Expect(err).ShouldNot(HaveOccurred())

		var updatedConfigMap *corev1.ConfigMap
		Eventually(func() (types.UID, error) {
			var err error
			By("Get updated configmap")
			updatedConfigMap, err = c.KubernetesInterface().CoreV1().ConfigMaps(generatedNamespace.GetName()).Get(context.Background(), fetchedInitialCatalog.Spec.ConfigMap, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			if len(updatedConfigMap.ObjectMeta.OwnerReferences) == 0 {
				return "", nil
			}
			return updatedConfigMap.ObjectMeta.OwnerReferences[0].UID, nil
		}).Should(Equal(fetchedUpdatedCatalog.ObjectMeta.UID))

		Expect(configMap.ResourceVersion).ShouldNot(Equal(updatedConfigMap.ResourceVersion))
		Expect(fetchedInitialCatalog.Status.ConfigMapResource.ResourceVersion).ShouldNot(Equal(fetchedUpdatedCatalog.Status.ConfigMapResource.ResourceVersion))
		Expect(fetchedUpdatedCatalog.Status.ConfigMapResource.ResourceVersion).Should(Equal(updatedConfigMap.GetResourceVersion()))

		By("Await 1 CatalogSource registry pod matching the updated labels")
		singlePod := podCount(1)
		selector := labels.SelectorFromSet(map[string]string{"olm.catalogSource": mainCatalogName, "olm.configMapResourceVersion": updatedConfigMap.GetResourceVersion()})
		podList, err := awaitPods(GinkgoT(), c, generatedNamespace.GetName(), selector.String(), singlePod)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(podList.Items).To(HaveLen(1), "expected pod list not of length 1")

		By("Await 1 CatalogSource registry pod matching the updated labels")
		selector = labels.SelectorFromSet(map[string]string{"olm.catalogSource": mainCatalogName})
		podList, err = awaitPods(GinkgoT(), c, generatedNamespace.GetName(), selector.String(), singlePod)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(podList.Items).To(HaveLen(1), "expected pod list not of length 1")

		By("Create Subscription")
		subscriptionName := genName("sub-")
		createSubscriptionForCatalog(crc, generatedNamespace.GetName(), subscriptionName, fetchedUpdatedCatalog.GetName(), mainPackageName, stableChannel, "", v1alpha1.ApprovalAutomatic)

		subscription, err := fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionStateAtLatestChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())
		_, err = fetchCSV(crc, generatedNamespace.GetName(), subscription.Status.CurrentCSV, buildCSVConditionChecker(v1alpha1.CSVPhaseSucceeded))
		Expect(err).ShouldNot(HaveOccurred())

		ipList, err := crc.OperatorsV1alpha1().InstallPlans(generatedNamespace.GetName()).List(context.Background(), metav1.ListOptions{})
		ipCount := 0
		for _, ip := range ipList.Items {
			if ownerutil.IsOwnedBy(&ip, subscription) {
				ipCount += 1
			}
		}
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("config map replace triggers registry pod rollout", func() {

		mainPackageName := genName("nginx-")
		dependentPackageName := genName("nginxdep-")

		mainPackageStable := fmt.Sprintf("%s-stable", mainPackageName)

		dependentPackageStable := fmt.Sprintf("%s-stable", dependentPackageName)

		stableChannel := "stable"

		dependentCRD := newCRD(genName("ins-"))
		mainCSV := newCSV(mainPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), nil, []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil)
		dependentCSV := newCSV(dependentPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil, nil)

		defer func() {
			Eventually(func() error {
				return ctx.Ctx().KubeClient().ApiextensionsInterface().ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), dependentCRD.GetName(), metav1.DeleteOptions{})
			}).Should(Or(Succeed(), WithTransform(k8serror.IsNotFound, BeTrue())))
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &mainCSV))
			}).Should(Succeed())
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &dependentCSV))
			}).Should(Succeed())
		}()

		mainCatalogName := genName("mock-ocs-main-")

		By("Create separate manifests for each CatalogSource")
		mainManifests := []registry.PackageManifest{
			{
				PackageName: mainPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: mainPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		dependentManifests := []registry.PackageManifest{
			{
				PackageName: dependentPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: dependentPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		By("Create the initial catalogsource")
		_, cleanupSource := createInternalCatalogSource(c, crc, mainCatalogName, generatedNamespace.GetName(), mainManifests, nil, []v1alpha1.ClusterServiceVersion{mainCSV})

		By("Attempt to get the catalog source before creating install plan")
		fetchedInitialCatalog, err := fetchCatalogSourceOnStatus(crc, mainCatalogName, generatedNamespace.GetName(), catalogSourceRegistryPodSynced())
		Expect(err).ShouldNot(HaveOccurred())
		By("Get initial configmap")
		configMap, err := c.KubernetesInterface().CoreV1().ConfigMaps(generatedNamespace.GetName()).Get(context.Background(), fetchedInitialCatalog.Spec.ConfigMap, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		By("Check pod created")
		initialPods, err := c.KubernetesInterface().CoreV1().Pods(generatedNamespace.GetName()).List(context.Background(), metav1.ListOptions{LabelSelector: "olm.configMapResourceVersion=" + configMap.ResourceVersion})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(initialPods.Items).To(HaveLen(1))

		By("delete the first catalog")
		cleanupSource()

		By("create a catalog with the same name")
		createInternalCatalogSource(c, crc, mainCatalogName, generatedNamespace.GetName(), append(mainManifests, dependentManifests...), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, []v1alpha1.ClusterServiceVersion{mainCSV, dependentCSV})

		By("Create Subscription")
		subscriptionName := genName("sub-")
		createSubscriptionForCatalog(crc, generatedNamespace.GetName(), subscriptionName, mainCatalogName, mainPackageName, stableChannel, "", v1alpha1.ApprovalAutomatic)

		subscription, err := fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionStateAtLatestChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ToNot(BeNil())
		_, err = fetchCSV(crc, generatedNamespace.GetName(), subscription.Status.CurrentCSV, buildCSVConditionChecker(v1alpha1.CSVPhaseSucceeded))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("gRPC address catalog source", func() {

		By("Create an internal (configmap) CatalogSource with stable and dependency csv")
		By("Create an internal (configmap) replacement CatalogSource with a stable, stable-replacement, and dependency csv")
		By("Copy both configmap-server pods to the test namespace")
		By("Delete both CatalogSources")
		By("Create an \"address\" CatalogSource with a Spec.Address field set to the stable copied pod's PodIP")
		By("Create a Subscription to the stable package")
		By("Wait for the stable Subscription to be Successful")
		By("Wait for the stable CSV to be Successful")
		By("Update the \"address\" CatalogSources's Spec.Address field with the PodIP of the replacement copied pod's PodIP")
		By("Wait for the replacement CSV to be Successful")

		mainPackageName := genName("nginx-")
		dependentPackageName := genName("nginxdep-")

		mainPackageStable := fmt.Sprintf("%s-stable", mainPackageName)
		mainPackageReplacement := fmt.Sprintf("%s-replacement", mainPackageStable)
		dependentPackageStable := fmt.Sprintf("%s-stable", dependentPackageName)

		stableChannel := "stable"

		dependentCRD := newCRD(genName("ins-"))
		mainCSV := newCSV(mainPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), nil, []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil)
		replacementCSV := newCSV(mainPackageReplacement, generatedNamespace.GetName(), mainPackageStable, semver.MustParse("0.2.0"), nil, []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil)
		dependentCSV := newCSV(dependentPackageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, nil, nil)

		defer func() {
			Eventually(func() error {
				return ctx.Ctx().KubeClient().ApiextensionsInterface().ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), dependentCRD.GetName(), metav1.DeleteOptions{})
			}).Should(Or(Succeed(), WithTransform(k8serror.IsNotFound, BeTrue())))
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &mainCSV))
			}).Should(Succeed())
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &dependentCSV))
			}).Should(Succeed())
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &replacementCSV))
			}).Should(Succeed())
		}()

		mainSourceName := genName("mock-ocs-main-")
		replacementSourceName := genName("mock-ocs-main-with-replacement-")

		By("Create separate manifests for each CatalogSource")
		mainManifests := []registry.PackageManifest{
			{
				PackageName: mainPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: mainPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		replacementManifests := []registry.PackageManifest{
			{
				PackageName: mainPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: mainPackageReplacement},
				},
				DefaultChannelName: stableChannel,
			},
		}

		dependentManifests := []registry.PackageManifest{
			{
				PackageName: dependentPackageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: dependentPackageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		By("Create ConfigMap CatalogSources")
		createInternalCatalogSource(c, crc, mainSourceName, generatedNamespace.GetName(), append(mainManifests, dependentManifests...), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, []v1alpha1.ClusterServiceVersion{mainCSV, dependentCSV})
		createInternalCatalogSource(c, crc, replacementSourceName, generatedNamespace.GetName(), append(replacementManifests, dependentManifests...), []apiextensionsv1.CustomResourceDefinition{dependentCRD}, []v1alpha1.ClusterServiceVersion{replacementCSV, mainCSV, dependentCSV})

		By("Wait for ConfigMap CatalogSources to be ready")
		mainSource, err := fetchCatalogSourceOnStatus(crc, mainSourceName, generatedNamespace.GetName(), catalogSourceRegistryPodSynced())
		Expect(err).ShouldNot(HaveOccurred())
		replacementSource, err := fetchCatalogSourceOnStatus(crc, replacementSourceName, generatedNamespace.GetName(), catalogSourceRegistryPodSynced())
		Expect(err).ShouldNot(HaveOccurred())

		By("Replicate catalog pods with no OwnerReferences")
		mainCopy := replicateCatalogPod(c, mainSource)
		mainCopy = awaitPod(GinkgoT(), c, mainCopy.GetNamespace(), mainCopy.GetName(), hasPodIP)
		replacementCopy := replicateCatalogPod(c, replacementSource)
		replacementCopy = awaitPod(GinkgoT(), c, replacementCopy.GetNamespace(), replacementCopy.GetName(), hasPodIP)

		addressSourceName := genName("address-catalog-")

		By("Create a CatalogSource pointing to the grpc pod")
		addressSource := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      addressSourceName,
				Namespace: generatedNamespace.GetName(),
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Address:    net.JoinHostPort(mainCopy.Status.PodIP, "50051"),
				GrpcPodConfig: &v1alpha1.GrpcPodConfig{
					SecurityContextConfig: v1alpha1.Restricted,
				},
			},
		}

		addressSource, err = crc.OperatorsV1alpha1().CatalogSources(generatedNamespace.GetName()).Create(context.Background(), addressSource, metav1.CreateOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		defer func() {
			err := crc.OperatorsV1alpha1().CatalogSources(generatedNamespace.GetName()).Delete(context.Background(), addressSourceName, metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())
		}()

		By("Wait for the CatalogSource to be ready")
		_, err = fetchCatalogSourceOnStatus(crc, addressSource.GetName(), addressSource.GetNamespace(), catalogSourceRegistryPodSynced())
		Expect(err).ToNot(HaveOccurred(), "catalog source did not become ready")

		By("Delete CatalogSources")
		err = crc.OperatorsV1alpha1().CatalogSources(generatedNamespace.GetName()).Delete(context.Background(), mainSourceName, metav1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		err = crc.OperatorsV1alpha1().CatalogSources(generatedNamespace.GetName()).Delete(context.Background(), replacementSourceName, metav1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		By("Create Subscription")
		subscriptionName := genName("sub-")
		cleanupSubscription := createSubscriptionForCatalog(crc, generatedNamespace.GetName(), subscriptionName, addressSourceName, mainPackageName, stableChannel, "", v1alpha1.ApprovalAutomatic)
		defer cleanupSubscription()

		subscription, err := fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionStateAtLatestChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())
		_, err = fetchCSV(crc, generatedNamespace.GetName(), subscription.Status.CurrentCSV, csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())

		By("Update the catalog's address to point at the other registry pod's cluster ip")
		Eventually(func() error {
			addressSource, err = crc.OperatorsV1alpha1().CatalogSources(generatedNamespace.GetName()).Get(context.Background(), addressSourceName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			addressSource.Spec.Address = net.JoinHostPort(replacementCopy.Status.PodIP, "50051")
			_, err = crc.OperatorsV1alpha1().CatalogSources(generatedNamespace.GetName()).Update(context.Background(), addressSource, metav1.UpdateOptions{})
			return err
		}).Should(Succeed())

		By("Wait for the replacement CSV to be installed")
		_, err = fetchCSV(crc, generatedNamespace.GetName(), replacementCSV.GetName(), csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("delete internal registry pod triggers recreation", func() {

		By("Create internal CatalogSource containing csv in package")
		By("Wait for a registry pod to be created")
		By("Delete the registry pod")
		By("Wait for a new registry pod to be created")

		By("Create internal CatalogSource containing csv in package")
		packageName := genName("nginx-")
		packageStable := fmt.Sprintf("%s-stable", packageName)
		stableChannel := "stable"
		sourceName := genName("catalog-")
		crd := newCRD(genName("ins-"))
		csv := newCSV(packageStable, generatedNamespace.GetName(), "", semver.MustParse("0.1.0"), []apiextensionsv1.CustomResourceDefinition{crd}, nil, nil)
		manifests := []registry.PackageManifest{
			{
				PackageName: packageName,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: packageStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		defer func() {
			Eventually(func() error {
				return ctx.Ctx().KubeClient().ApiextensionsInterface().ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), crd.GetName(), metav1.DeleteOptions{})
			}).Should(Or(Succeed(), WithTransform(k8serror.IsNotFound, BeTrue())))
			Eventually(func() error {
				return client.IgnoreNotFound(ctx.Ctx().Client().Delete(context.Background(), &csv))
			}).Should(Succeed())
		}()

		_, cleanupSource := createInternalCatalogSource(c, crc, sourceName, generatedNamespace.GetName(), manifests, []apiextensionsv1.CustomResourceDefinition{crd}, []v1alpha1.ClusterServiceVersion{csv})
		defer cleanupSource()

		By("Wait for a new registry pod to be created")
		selector := labels.SelectorFromSet(map[string]string{"olm.catalogSource": sourceName})
		singlePod := podCount(1)
		registryPods, err := awaitPods(GinkgoT(), c, generatedNamespace.GetName(), selector.String(), singlePod)
		Expect(err).ShouldNot(HaveOccurred(), "error awaiting registry pod")
		Expect(registryPods).ToNot(BeNil(), "nil registry pods")
		Expect(registryPods.Items).To(HaveLen(1), "unexpected number of registry pods found")

		By("Store the UID for later comparison")
		uid := registryPods.Items[0].GetUID()

		By("Delete the registry pod")
		Eventually(func() error {
			backgroundDeletion := metav1.DeletePropagationBackground
			return c.KubernetesInterface().CoreV1().Pods(generatedNamespace.GetName()).DeleteCollection(context.Background(), metav1.DeleteOptions{PropagationPolicy: &backgroundDeletion}, metav1.ListOptions{LabelSelector: selector.String()})
		}).Should(Succeed())

		By("Wait for a new registry pod to be created")
		notUID := func(pods *corev1.PodList) bool {
			uids := make([]string, 0)
			for _, pod := range pods.Items {
				uids = append(uids, string(pod.GetUID()))
				if pod.GetUID() == uid {
					ctx.Ctx().Logf("waiting for %v not to contain %s", uids, uid)
					return false
				}
			}
			ctx.Ctx().Logf("waiting for %v to not be empty and not contain %s", uids, uid)
			return len(pods.Items) > 0
		}
		registryPods, err = awaitPods(GinkgoT(), c, generatedNamespace.GetName(), selector.String(), unionPodsCheck(singlePod, notUID))
		Expect(err).ShouldNot(HaveOccurred(), "error waiting for replacement registry pod")
		Expect(registryPods).ToNot(BeNil(), "nil replacement registry pods")
		Expect(registryPods.Items).To(HaveLen(1), "unexpected number of replacement registry pods found")
	})

	It("delete gRPC registry pod triggers recreation", func() {

		By("Create gRPC CatalogSource using an external registry image (community-operators)")
		By("Wait for a registry pod to be created")
		By("Delete the registry pod")
		By("Wait for a new registry pod to be created")

		By("Create gRPC CatalogSource using an external registry image (community-operators)")
		source := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      genName("catalog-"),
				Namespace: generatedNamespace.GetName(),
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Image:      communityOperatorsImage,
				GrpcPodConfig: &v1alpha1.GrpcPodConfig{
					SecurityContextConfig: v1alpha1.Restricted,
				},
			},
		}

		source, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		By("Wait for a new registry pod to be created")
		selector := labels.SelectorFromSet(map[string]string{"olm.catalogSource": source.GetName()})
		singlePod := podCount(1)
		registryPods, err := awaitPods(GinkgoT(), c, source.GetNamespace(), selector.String(), singlePod)
		Expect(err).ShouldNot(HaveOccurred(), "error awaiting registry pod")
		Expect(registryPods).ToNot(BeNil(), "nil registry pods")
		Expect(registryPods.Items).To(HaveLen(1), "unexpected number of registry pods found")

		By("Store the UID for later comparison")
		uid := registryPods.Items[0].GetUID()

		By("Delete the registry pod")
		Eventually(func() error {
			backgroundDeletion := metav1.DeletePropagationBackground
			return c.KubernetesInterface().CoreV1().Pods(generatedNamespace.GetName()).DeleteCollection(context.Background(), metav1.DeleteOptions{PropagationPolicy: &backgroundDeletion}, metav1.ListOptions{LabelSelector: selector.String()})
		}).Should(Succeed())

		By("Wait for a new registry pod to be created")
		notUID := func(pods *corev1.PodList) bool {
			uids := make([]string, 0)
			for _, pod := range pods.Items {
				uids = append(uids, string(pod.GetUID()))
				if pod.GetUID() == uid {
					ctx.Ctx().Logf("waiting for %v not to contain %s", uids, uid)
					return false
				}
			}
			ctx.Ctx().Logf("waiting for %v to not be empty and not contain %s", uids, uid)
			return len(pods.Items) > 0
		}
		registryPods, err = awaitPods(GinkgoT(), c, generatedNamespace.GetName(), selector.String(), unionPodsCheck(singlePod, notUID))
		Expect(err).ShouldNot(HaveOccurred(), "error waiting for replacement registry pod")
		Expect(registryPods).ShouldNot(BeNil(), "nil replacement registry pods")
		Expect(registryPods.Items).To(HaveLen(1), "unexpected number of replacement registry pods found")
	})

	for _, npType := range []string{"grpc-server", "unpack-bundles"} {
		It(fmt.Sprintf("delete registry %s network policy triggers recreation", npType), func() {
			By("Creating CatalogSource using an external registry image (community-operators)")
			source := &v1alpha1.CatalogSource{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.CatalogSourceKind,
					APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      genName("catalog-"),
					Namespace: generatedNamespace.GetName(),
				},
				Spec: v1alpha1.CatalogSourceSpec{
					SourceType: v1alpha1.SourceTypeGrpc,
					Image:      communityOperatorsImage,
					GrpcPodConfig: &v1alpha1.GrpcPodConfig{
						SecurityContextConfig: v1alpha1.Restricted,
					},
				},
			}

			source, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
			Expect(err).ShouldNot(HaveOccurred())

			npName := fmt.Sprintf("%s-%s", source.GetName(), npType)

			var networkPolicy *networkingv1.NetworkPolicy
			Eventually(func() error {
				networkPolicy, err = c.KubernetesInterface().NetworkingV1().NetworkPolicies(source.GetNamespace()).Get(context.Background(), npName, metav1.GetOptions{})
				return err
			}, pollDuration, pollInterval).Should(Succeed())
			Expect(networkPolicy).NotTo(BeNil())

			By("Storing the UID for later comparison")
			uid := networkPolicy.GetUID()

			By("Deleting the network policy")
			err = c.KubernetesInterface().NetworkingV1().NetworkPolicies(source.GetNamespace()).Delete(context.Background(), npName, metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for a new network policy be created")
			Eventually(func() error {
				networkPolicy, err = c.KubernetesInterface().NetworkingV1().NetworkPolicies(source.GetNamespace()).Get(context.Background(), npName, metav1.GetOptions{})
				if err != nil {
					if k8serror.IsNotFound(err) {
						ctx.Ctx().Logf("waiting for new network policy to be created")
					} else {
						ctx.Ctx().Logf("error getting network policy %q: %v", npName, err)
					}
					return err
				}
				if networkPolicy.GetUID() == uid {
					return fmt.Errorf("network policy with original uid still exists... (did the deletion somehow fail?)")
				}
				return nil
			}, pollDuration, pollInterval).Should(Succeed())
		})

		It(fmt.Sprintf("change registry %s network policy triggers revert to desired", npType), func() {
			By("Create CatalogSource using an external registry image (community-operators)")
			source := &v1alpha1.CatalogSource{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.CatalogSourceKind,
					APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      genName("catalog-"),
					Namespace: generatedNamespace.GetName(),
				},
				Spec: v1alpha1.CatalogSourceSpec{
					SourceType: v1alpha1.SourceTypeGrpc,
					Image:      communityOperatorsImage,
					GrpcPodConfig: &v1alpha1.GrpcPodConfig{
						SecurityContextConfig: v1alpha1.Restricted,
					},
				},
			}

			source, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
			Expect(err).ShouldNot(HaveOccurred())

			npName := fmt.Sprintf("%s-%s", source.GetName(), npType)

			var networkPolicy *networkingv1.NetworkPolicy
			Eventually(func() error {
				networkPolicy, err = c.KubernetesInterface().NetworkingV1().NetworkPolicies(source.GetNamespace()).Get(context.Background(), npName, metav1.GetOptions{})
				return err
			}, pollDuration, pollInterval).Should(Succeed())
			Expect(networkPolicy).NotTo(BeNil())

			By("Patching the network policy with an undesirable egress policy")
			npac := networkingv1ac.NetworkPolicy(npName, source.GetNamespace()).
				WithSpec(networkingv1ac.NetworkPolicySpec().
					WithEgress(networkingv1ac.NetworkPolicyEgressRule().
						WithPorts(networkingv1ac.NetworkPolicyPort().
							WithProtocol(corev1.ProtocolTCP).
							WithPort(intstr.FromString("foobar")),
						),
					),
				)
			np, err := c.KubernetesInterface().NetworkingV1().NetworkPolicies(source.GetNamespace()).Apply(context.Background(), npac, metav1.ApplyOptions{FieldManager: "olm-e2e-test", Force: true})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(np.Spec.Egress).To(HaveLen(1))

			By("Waiting for the network policy be reverted")
			Eventually(func() error {
				np, err := c.KubernetesInterface().NetworkingV1().NetworkPolicies(source.GetNamespace()).Get(context.Background(), npName, metav1.GetOptions{})
				if err != nil {
					ctx.Ctx().Logf("error getting network policy %q: %v", npName, err)
					return err
				}

				if needsRevert := func() bool {
					for _, rule := range np.Spec.Egress {
						for _, port := range rule.Ports {
							if port.Port.String() == "foobar" {
								return true
							}
						}
					}
					return false
				}(); needsRevert {
					ctx.Ctx().Logf("waiting for egress rule to be reverted")
					return fmt.Errorf("extra network policy egress rule has not been reverted")
				}
				return nil
			}, pollDuration, pollInterval).Should(Succeed())
		})
	}

	It("configure gRPC registry pod to extract content", func() {

		By("Create gRPC CatalogSource using an external registry image (community-operators)")
		source := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      genName("catalog-"),
				Namespace: generatedNamespace.GetName(),
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Image:      communityOperatorsImage,
				GrpcPodConfig: &v1alpha1.GrpcPodConfig{
					SecurityContextConfig: v1alpha1.Restricted,
					ExtractContent: &v1alpha1.ExtractContentConfig{
						CacheDir:   "/tmp/cache",
						CatalogDir: "/configs",
					},
				},
			},
		}

		source, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		By("Wait for the CatalogSource to be ready")
		source, err = fetchCatalogSourceOnStatus(crc, source.GetName(), source.GetNamespace(), catalogSourceRegistryPodSynced())
		Expect(err).ToNot(HaveOccurred(), "catalog source did not become ready")

		By("the gRPC endpoints are not exposed from the pod, and there's no simple way to get at them -")
		By("the index images don't contain `grpcurl`, port-forwarding is a mess, etc. let's use the")
		By("package-server as a proxy for a functional catalog")
		By("Waiting for packages from the catalog to show up in the Kubernetes API")
		Eventually(func() error {
			manifests, err := packageserverClient.PackagesV1().PackageManifests("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			if len(manifests.Items) == 0 {
				return errors.New("did not find any PackageManifests")
			}
			return nil
		}).Should(Succeed())
	})

	It("image update", func() {
		if ok, err := inKind(c); ok && err == nil {
			Skip("This spec fails when run using KIND cluster. See https://github.com/operator-framework/operator-lifecycle-manager/issues/2420 for more details")
		} else if err != nil {
			Skip("Could not determine whether running in a kind cluster. Skipping.")
		}
		By("Create an image based catalog source from public Quay image")
		By("Use a unique tag as identifier")
		By("See https://quay.io/repository/olmtest/catsrc-update-test?namespace=olmtest for registry")
		By("Push an updated version of the image with the same identifier")
		By("Confirm catalog source polling feature is working as expected: a newer version of the catalog source pod comes up")
		By("etcd operator updated from 0.9.0 to 0.9.2-clusterwide")
		By("Subscription should detect the latest version of the operator in the new catalog source and pull it")

		By("create internal registry for purposes of pushing/pulling IF running e2e test locally")
		By("registry is insecure and for purposes of this test only")

		local, err := Local(c)
		Expect(err).NotTo(HaveOccurred(), "cannot determine if test running locally or on CI: %s", err)

		By("Create an image based catalog source from public Quay image using a unique tag as identifier")
		var registryURL string
		var registryAuthSecretName string
		if local {
			By("Creating a local registry to use")
			registryURL, err = createDockerRegistry(c, generatedNamespace.GetName())
			Expect(err).NotTo(HaveOccurred(), "error creating container registry: %s", err)
			defer deleteDockerRegistry(c, generatedNamespace.GetName())

			By("ensure registry pod with local URL " + registryURL + " is ready before attempting port-forwarding")
			_ = awaitPod(GinkgoT(), c, generatedNamespace.GetName(), registryName, podReady)

			By("By port-forwarding to the registry")
			err = registryPortForward(generatedNamespace.GetName())
			Expect(err).NotTo(HaveOccurred(), "port-forwarding local registry: %s", err)
		} else {
			registryURL = fmt.Sprintf("%s/%s", openshiftregistryFQDN, generatedNamespace.GetName())
			By("Using the OpenShift registry at " + registryURL)
			registryAuthSecretName, err = getRegistryAuthSecretName(c, generatedNamespace.GetName())
			Expect(err).NotTo(HaveOccurred(), "error getting openshift registry authentication: %s", err)
		}

		By("testImage is the name of the image used throughout the test - the image overwritten by skopeo")
		By("the tag is generated randomly and appended to the end of the testImage")
		testImage := fmt.Sprint("docker://", registryURL, "/catsrc-update", ":")
		tag := genName("x")
		By("Generating a target test image name " + testImage + " with tag " + tag)

		By("copying old catalog image into test-specific tag in internal docker registry")
		if local {
			By("executing out to a local skopeo client")
			_, err := skopeoLocalCopy(testImage, tag, catsrcImage, "old")
			Expect(err).NotTo(HaveOccurred(), "error copying old registry file: %s", err)
		} else {
			By("creating a skopoeo Pod to do the copying")
			skopeoArgs := skopeoCopyCmd(testImage, tag, catsrcImage, "old", registryAuthSecretName)
			err = createSkopeoPod(c, skopeoArgs, generatedNamespace.GetName(), registryAuthSecretName)
			Expect(err).NotTo(HaveOccurred(), "error creating skopeo pod: %s", err)

			By("waiting for the skopeo pod to exit successfully")
			awaitPod(GinkgoT(), c, generatedNamespace.GetName(), skopeo, func(pod *corev1.Pod) bool {
				return pod.Status.Phase == corev1.PodSucceeded
			})

			By("removing the skopeo pod")
			err = deleteSkopeoPod(c, generatedNamespace.GetName())
			Expect(err).NotTo(HaveOccurred(), "error deleting skopeo pod: %s", err)
		}

		By("setting catalog source")
		sourceName := genName("catalog-")
		packageName := "busybox"
		channelName := "alpha"

		By("Create gRPC CatalogSource using an external registry image and poll interval")
		var image string
		image = testImage[9:] // strip off docker://
		image = fmt.Sprint(image, tag)

		source := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: generatedNamespace.GetName(),
				Labels:    map[string]string{"olm.catalogSource": sourceName},
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Image:      image,
				GrpcPodConfig: &v1alpha1.GrpcPodConfig{
					SecurityContextConfig: v1alpha1.Restricted,
				},
				UpdateStrategy: &v1alpha1.UpdateStrategy{
					RegistryPoll: &v1alpha1.RegistryPoll{
						// Using RawInterval rather than Interval due to this issue:
						// https://github.com/operator-framework/operator-lifecycle-manager/issues/2621
						RawInterval: "1m0s",
					},
				},
			},
		}

		source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		defer func() {
			err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Delete(context.Background(), source.GetName(), metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())
		}()

		By("wait for new catalog source pod to be created")
		By("Wait for a new registry pod to be created")
		selector := labels.SelectorFromSet(map[string]string{"olm.catalogSource": source.GetName()})
		singlePod := podCount(1)
		registryPods, err := awaitPods(GinkgoT(), c, source.GetNamespace(), selector.String(), singlePod)
		Expect(err).ToNot(HaveOccurred(), "error awaiting registry pod")
		Expect(registryPods).ShouldNot(BeNil(), "nil registry pods")
		Expect(registryPods.Items).To(HaveLen(1), "unexpected number of registry pods found")

		By("Create a Subscription for package")
		subscriptionName := genName("sub-")
		cleanupSubscription := createSubscriptionForCatalog(crc, source.GetNamespace(), subscriptionName, source.GetName(), packageName, channelName, "", v1alpha1.ApprovalAutomatic)
		defer cleanupSubscription()

		By("Wait for the Subscription to succeed")
		subscription, err := fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionStateAtLatestChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())

		By("Wait for csv to succeed")
		_, err = fetchCSV(crc, subscription.GetNamespace(), subscription.Status.CurrentCSV, csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())

		registryCheckFunc := func(podList *corev1.PodList) bool {
			if len(podList.Items) > 1 {
				return false
			}
			return podList.Items[0].Status.ContainerStatuses[0].ImageID != ""
		}
		By("get old catalog source pod")
		registryPod, err := awaitPods(GinkgoT(), c, source.GetNamespace(), selector.String(), registryCheckFunc)
		By("Updateing image on registry to trigger a newly updated version of the catalog source pod to be deployed after some time")
		if local {
			By("executing out to a local skopeo client")
			_, err := skopeoLocalCopy(testImage, tag, catsrcImage, "new")
			Expect(err).NotTo(HaveOccurred(), "error copying new registry file: %s", err)
		} else {
			By("creating a skopoeo Pod to do the copying")
			skopeoArgs := skopeoCopyCmd(testImage, tag, catsrcImage, "new", registryAuthSecretName)
			err = createSkopeoPod(c, skopeoArgs, generatedNamespace.GetName(), registryAuthSecretName)
			Expect(err).NotTo(HaveOccurred(), "error creating skopeo pod: %s", err)

			By("waiting for the skopeo pod to exit successfully")
			awaitPod(GinkgoT(), c, generatedNamespace.GetName(), skopeo, func(pod *corev1.Pod) bool {
				return pod.Status.Phase == corev1.PodSucceeded
			})

			By("removing the skopeo pod")
			err = deleteSkopeoPod(c, generatedNamespace.GetName())
			Expect(err).NotTo(HaveOccurred(), "error deleting skopeo pod: %s", err)
		}

		By("update catalog source with annotation (to kick resync)")
		Eventually(func() error {
			source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
			if err != nil {
				return nil
			}

			source.Annotations = make(map[string]string)
			source.Annotations["testKey"] = "testValue"
			_, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Update(context.Background(), source, metav1.UpdateOptions{})
			return err
		}).Should(Succeed())

		By("ensure new registry pod container image is as we expect")
		podCheckFunc := func(podList *corev1.PodList) bool {
			ctx.Ctx().Logf("pod list length %d\n", len(podList.Items))
			for _, pod := range podList.Items {
				ctx.Ctx().Logf("pod list name %v\n", pod.Name)
			}

			for _, pod := range podList.Items {
				ctx.Ctx().Logf("old image id %s\n new image id %s\n", registryPod.Items[0].Status.ContainerStatuses[0].ImageID,
					pod.Status.ContainerStatuses[0].ImageID)
				if pod.Status.ContainerStatuses[0].ImageID != registryPod.Items[0].Status.ContainerStatuses[0].ImageID {
					return true
				}
			}
			By("update catalog source with annotation (to kick resync)")
			Eventually(func() error {
				source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
				if err != nil {
					return err
				}

				if source.Annotations == nil {
					source.Annotations = make(map[string]string)
				}

				source.Annotations["testKey"] = genName("newValue")
				_, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Update(context.Background(), source, metav1.UpdateOptions{})
				return err
			}).Should(Succeed())

			return false
		}
		By("await new catalog source and ensure old one was deleted")
		registryPods, err = awaitPodsWithInterval(GinkgoT(), c, source.GetNamespace(), selector.String(), 30*time.Second, 10*time.Minute, podCheckFunc)
		Expect(err).ShouldNot(HaveOccurred(), "error awaiting registry pod")
		Expect(registryPods).ShouldNot(BeNil(), "nil registry pods")
		Expect(registryPods.Items).To(HaveLen(1), "unexpected number of registry pods found")

		By("update catalog source with annotation (to kick resync)")
		Eventually(func() error {
			source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
			if err != nil {
				return err
			}

			source.Annotations["testKey"] = "newValue"
			_, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Update(context.Background(), source, metav1.UpdateOptions{})
			return err
		}).Should(Succeed())

		subChecker := func(sub *v1alpha1.Subscription) bool {
			return sub.Status.InstalledCSV == "busybox.v2.0.0"
		}
		By("Wait for the Subscription to succeed")
		subscription, err = fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subChecker)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())

		By("Wait for csv to succeed")
		csv, err := fetchCSV(crc, subscription.GetNamespace(), subscription.Status.CurrentCSV, csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())

		By("check version of running csv to ensure the latest version (0.9.2) was installed onto the cluster")
		v := csv.Spec.Version
		busyboxVersion := semver.Version{
			Major: 2,
			Minor: 0,
			Patch: 0,
		}

		Expect(v).Should(Equal(version.OperatorVersion{Version: busyboxVersion}), "latest version of operator not installed: catalog source update failed")
	})

	It("Dependency has correct replaces field", func() {
		By("Create a CatalogSource that contains the busybox v1 and busybox-dependency v1 images")
		By("Create a Subscription for busybox v1, which has a dependency on busybox-dependency v1.")
		By("Wait for the busybox and busybox2 Subscriptions to succeed")
		By("Wait for the CSVs to succeed")
		By("Update the catalog to point to an image that contains the busybox v2 and busybox-dependency v2 images.")
		By("Wait for the new Subscriptions to succeed and check if they include the new CSVs")
		By("Wait for the CSVs to succeed and confirm that the have the correct Spec.Replaces fields.")

		sourceName := genName("catalog-")
		packageName := "busybox"
		channelName := "alpha"

		catSrcImage := "quay.io/olmtest/busybox-dependencies-index"

		By("creating gRPC CatalogSource")
		source := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: generatedNamespace.GetName(),
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Image:      catSrcImage + ":1.0.0-with-ListBundles-method",
				GrpcPodConfig: &v1alpha1.GrpcPodConfig{
					SecurityContextConfig: v1alpha1.Restricted,
				},
			},
		}
		source, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		defer func() {
			err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Delete(context.Background(), source.GetName(), metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())
		}()

		By("waiting for the CatalogSource to be ready")
		_, err = fetchCatalogSourceOnStatus(crc, source.GetName(), source.GetNamespace(), catalogSourceRegistryPodSynced())
		Expect(err).ToNot(HaveOccurred(), "catalog source did not become ready")

		By("creating a Subscription for busybox")
		subscriptionName := genName("sub-")
		cleanupSubscription := createSubscriptionForCatalog(crc, source.GetNamespace(), subscriptionName, source.GetName(), packageName, channelName, "", v1alpha1.ApprovalAutomatic)
		defer cleanupSubscription()

		By("waiting for the Subscription to succeed")
		subscription, err := fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subscriptionStateAtLatestChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())
		Expect(subscription.Status.InstalledCSV).To(Equal("busybox.v1.0.0"))

		By("confirming that a subscription was created for busybox-dependency")
		subscriptionList, err := crc.OperatorsV1alpha1().Subscriptions(source.GetNamespace()).List(context.Background(), metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		dependencySubscriptionName := ""
		for _, sub := range subscriptionList.Items {
			if strings.HasPrefix(sub.GetName(), "busybox-dependency") {
				dependencySubscriptionName = sub.GetName()
			}
		}
		Expect(dependencySubscriptionName).ToNot(BeEmpty())

		By("waiting for the Subscription to succeed")
		subscription, err = fetchSubscription(crc, generatedNamespace.GetName(), dependencySubscriptionName, subscriptionStateAtLatestChecker())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())
		Expect(subscription.Status.InstalledCSV).To(Equal("busybox-dependency.v1.0.0"))

		By("updating the catalog image")
		Eventually(func() error {
			existingSource, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), sourceName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			existingSource.Spec.Image = catSrcImage + ":2.0.0-with-ListBundles-method"

			source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Update(context.Background(), existingSource, metav1.UpdateOptions{})
			return err
		}).Should(Succeed())

		By("waiting for the CatalogSource to be ready")
		_, err = fetchCatalogSourceOnStatus(crc, source.GetName(), source.GetNamespace(), catalogSourceRegistryPodSynced())
		Expect(err).ToNot(HaveOccurred(), "catalog source did not become ready")

		By("waiting for the busybox v2 Subscription to succeed")
		subChecker := func(sub *v1alpha1.Subscription) bool {
			return sub.Status.InstalledCSV == "busybox.v2.0.0"
		}
		subscription, err = fetchSubscription(crc, generatedNamespace.GetName(), subscriptionName, subChecker)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())

		By("waiting for busybox v2 csv to succeed and check the replaces field")
		csv, err := fetchCSV(crc, subscription.GetNamespace(), subscription.Status.CurrentCSV, csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(csv.Spec.Replaces).To(Equal("busybox.v1.0.0"))

		By("waiting for the busybox-dependency v2 Subscription to succeed")
		subChecker = func(sub *v1alpha1.Subscription) bool {
			return sub.Status.InstalledCSV == "busybox-dependency.v2.0.0"
		}
		subscription, err = fetchSubscription(crc, generatedNamespace.GetName(), dependencySubscriptionName, subChecker)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(subscription).ShouldNot(BeNil())

		By("waiting for busybox-dependency v2 csv to succeed and check the replaces field")
		csv, err = fetchCSV(crc, subscription.GetNamespace(), subscription.Status.CurrentCSV, csvSucceededChecker)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(csv.Spec.Replaces).To(Equal("busybox-dependency.v1.0.0"))
	})
	When("A catalogSource is created with correct polling interval", func() {
		var source *v1alpha1.CatalogSource
		singlePod := podCount(1)
		sourceName := genName("catalog-")

		BeforeEach(func() {
			source = &v1alpha1.CatalogSource{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.CatalogSourceKind,
					APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sourceName,
					Namespace: generatedNamespace.GetName(),
					Labels:    map[string]string{"olm.catalogSource": sourceName},
				},
				Spec: v1alpha1.CatalogSourceSpec{
					SourceType: v1alpha1.SourceTypeGrpc,
					Image:      "quay.io/olmtest/catsrc-update-test:new",
					GrpcPodConfig: &v1alpha1.GrpcPodConfig{
						SecurityContextConfig: v1alpha1.Restricted,
					},
					UpdateStrategy: &v1alpha1.UpdateStrategy{
						RegistryPoll: &v1alpha1.RegistryPoll{
							RawInterval: "45s",
						},
					},
				},
			}

			source, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			By("wait for new catalog source pod to be created and report ready")
			selector := labels.SelectorFromSet(map[string]string{"olm.catalogSource": source.GetName()})

			catalogPods, err := awaitPods(GinkgoT(), c, source.GetNamespace(), selector.String(), singlePod)
			Expect(err).ToNot(HaveOccurred())
			Expect(catalogPods).ToNot(BeNil())

			Eventually(func() (bool, error) {
				podList, err := c.KubernetesInterface().CoreV1().Pods(source.GetNamespace()).List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
				if err != nil {
					return false, err
				}

				for _, p := range podList.Items {
					if podReady(&p) {
						return true, nil
					}
					return false, nil
				}

				return false, nil
			}).Should(BeTrue())
		})

		It("registry polls on the correct interval", func() {
			By("Wait roughly the polling interval for update pod to show up")
			updateSelector := labels.SelectorFromSet(map[string]string{"catalogsource.operators.coreos.com/update": source.GetName()})
			updatePods, err := awaitPodsWithInterval(GinkgoT(), c, source.GetNamespace(), updateSelector.String(), 5*time.Second, 2*time.Minute, singlePod)
			Expect(err).ToNot(HaveOccurred())
			Expect(updatePods).ToNot(BeNil())
			Expect(updatePods.Items).To(HaveLen(1))

			By("No update to image: update pod should be deleted quickly")
			noPod := podCount(0)
			updatePods, err = awaitPodsWithInterval(GinkgoT(), c, source.GetNamespace(), updateSelector.String(), 1*time.Second, 30*time.Second, noPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(updatePods.Items).To(HaveLen(0))
		})

	})

	When("A catalogSource is created with incorrect polling interval", func() {
		var (
			source     *v1alpha1.CatalogSource
			sourceName string
		)

		const (
			incorrectInterval = "45mError.code"
			correctInterval   = "45m"
		)

		BeforeEach(func() {
			sourceName = genName("catalog-")
			source = &v1alpha1.CatalogSource{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.CatalogSourceKind,
					APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sourceName,
					Namespace: generatedNamespace.GetName(),
					Labels:    map[string]string{"olm.catalogSource": sourceName},
				},
				Spec: v1alpha1.CatalogSourceSpec{
					SourceType: v1alpha1.SourceTypeGrpc,
					Image:      "quay.io/olmtest/catsrc-update-test:new",
					GrpcPodConfig: &v1alpha1.GrpcPodConfig{
						SecurityContextConfig: v1alpha1.Restricted,
					},
					UpdateStrategy: &v1alpha1.UpdateStrategy{
						RegistryPoll: &v1alpha1.RegistryPoll{
							RawInterval: incorrectInterval,
						},
					},
				},
			}

			_, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

		})
		AfterEach(func() {
			err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Delete(context.Background(), source.GetName(), metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
		})
		It("the catalogsource status communicates that a default interval time is being used instead", func() {
			Eventually(func() bool {
				catsrc, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				if catsrc.Status.Reason == v1alpha1.CatalogSourceIntervalInvalidError {
					if catsrc.Status.Message == "error parsing spec.updateStrategy.registryPoll.interval. Using the default value of 15m0s instead. Error: time: unknown unit \"mError\" in duration \"45mError.code\"" {
						return true
					}
				}
				return false
			}).Should(BeTrue())
		})
		When("the catalogsource is updated with a valid polling interval", func() {

			BeforeEach(func() {
				Eventually(func() error {
					catsrc, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
					if err != nil {
						return err
					}
					catsrc.Spec.UpdateStrategy.RegistryPoll.RawInterval = correctInterval
					_, err = crc.OperatorsV1alpha1().CatalogSources(catsrc.GetNamespace()).Update(context.Background(), catsrc, metav1.UpdateOptions{})
					return err
				}).Should(Succeed())
			})

			It("the catalogsource spec shows the updated polling interval, and the error message in the status is cleared", func() {
				Eventually(func() error {
					catsrc, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
					if err != nil {
						return err
					}
					expectedTime, err := time.ParseDuration(correctInterval)
					if err != nil {
						return err
					}
					if catsrc.Status.Reason != "" || (catsrc.Spec.UpdateStrategy.Interval != &metav1.Duration{Duration: expectedTime}) {
						return err
					}
					return nil
				}).Should(Succeed())
			})
		})
	})

	It("adding catalog template adjusts image used", func() {
		By("This test attempts to create a catalog source, and update it with a template annotation")
		By("and ensure that the image gets changed according to what's in the template as well as")
		By("check the status conditions are updated accordingly")
		sourceName := genName("catalog-")
		source := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: generatedNamespace.GetName(),
				Labels:    map[string]string{"olm.catalogSource": sourceName},
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Image:      "quay.io/olmtest/catsrc-update-test:old",
				GrpcPodConfig: &v1alpha1.GrpcPodConfig{
					SecurityContextConfig: v1alpha1.Restricted,
				},
			},
		}

		By("creating a catalog source")

		var err error
		Eventually(func() error {
			source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
			return err
		}).Should(Succeed())

		By("updating the catalog source with template annotation")

		Eventually(func() error {
			source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), source.GetName(), metav1.GetOptions{})
			if err != nil {
				return err
			}
			By("create an annotation using the kube templates")
			source.SetAnnotations(map[string]string{
				catalogsource.CatalogImageTemplateAnnotation: fmt.Sprintf("quay.io/olmtest/catsrc-update-test:%s.%s.%s", catalogsource.TemplKubeMajorV, catalogsource.TemplKubeMinorV, catalogsource.TemplKubePatchV),
			})

			By("Update the catalog image")
			_, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Update(context.Background(), source, metav1.UpdateOptions{})
			return err
		}).Should(Succeed())

		By("wait for status condition to show up")
		Eventually(func() (bool, error) {
			source, err = crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Get(context.Background(), sourceName, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			By("if the conditions array has the entry we know things got updated")
			condition := meta.FindStatusCondition(source.Status.Conditions, catalogtemplate.StatusTypeTemplatesHaveResolved)
			if condition != nil {
				return true, nil
			}

			return false, nil
		}).Should(BeTrue())

		By("source should be the latest we got from the eventually block")
		Expect(source.Status.Conditions).ToNot(BeNil())

		templatesResolvedCondition := meta.FindStatusCondition(source.Status.Conditions, catalogtemplate.StatusTypeTemplatesHaveResolved)
		if Expect(templatesResolvedCondition).ToNot(BeNil()) {
			Expect(templatesResolvedCondition.Reason).To(BeIdenticalTo(catalogtemplate.ReasonAllTemplatesResolved))
			Expect(templatesResolvedCondition.Status).To(BeIdenticalTo(metav1.ConditionTrue))
		}
		resolvedImageCondition := meta.FindStatusCondition(source.Status.Conditions, catalogtemplate.StatusTypeResolvedImage)
		if Expect(resolvedImageCondition).ToNot(BeNil()) {
			Expect(resolvedImageCondition.Reason).To(BeIdenticalTo(catalogtemplate.ReasonAllTemplatesResolved))
			Expect(resolvedImageCondition.Status).To(BeIdenticalTo(metav1.ConditionTrue))

			By("if we can, try to determine the server version so we can check the resulting image")
			if serverVersion, err := crc.Discovery().ServerVersion(); err != nil {
				if serverGitVersion, err := semver.Parse(serverVersion.GitVersion); err != nil {
					expectedImage := fmt.Sprintf("quay.io/olmtest/catsrc-update-test:%s.%s.%s", serverVersion.Major, serverVersion.Minor, strconv.FormatUint(serverGitVersion.Patch, 10))
					Expect(resolvedImageCondition.Message).To(BeIdenticalTo(expectedImage))
				}
			}
		}
	})

	When("A CatalogSource is created with an operator that has a CSV with missing metadata.ApiVersion", func() {
		var (
			magicCatalog      *MagicCatalog
			catalogSourceName string
			subscription      *v1alpha1.Subscription
			c                 client.Client
		)

		BeforeEach(func() {
			c = ctx.Ctx().Client()

			provider, err := NewFileBasedFiledBasedCatalogProvider(filepath.Join(testdataDir, badCSVDir, "bad-csv.yaml"))
			Expect(err).To(BeNil())

			catalogSourceName = genName("cat-bad-csv")
			magicCatalog = NewMagicCatalog(c, generatedNamespace.GetName(), catalogSourceName, provider)
			Expect(magicCatalog.DeployCatalog(context.Background())).To(BeNil())

		})

		AfterEach(func() {
			TeardownNamespace(generatedNamespace.GetName())
		})

		When("A Subscription is created catalogSource built with the malformed CSV", func() {

			BeforeEach(func() {
				subscription = &v1alpha1.Subscription{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-sub", catalogSourceName),
						Namespace: generatedNamespace.GetName(),
					},
					Spec: &v1alpha1.SubscriptionSpec{
						CatalogSource:          catalogSourceName,
						CatalogSourceNamespace: generatedNamespace.GetName(),
						Channel:                "stable",
						Package:                "test-package",
					},
				}
				Expect(c.Create(context.Background(), subscription)).To(BeNil())
			})

			It("fails with a BundleUnpackFailed error condition, and a message that highlights the missing field in the CSV", func() {
				Eventually(func(g Gomega) string {
					fetchedSubscription, err := crc.OperatorsV1alpha1().Subscriptions(generatedNamespace.GetName()).Get(context.Background(), subscription.GetName(), metav1.GetOptions{})
					g.Expect(err).NotTo(HaveOccurred())

					By("expect the message that API missing")
					failingCondition := fetchedSubscription.Status.GetCondition(v1alpha1.SubscriptionBundleUnpackFailed)
					return failingCondition.Message
				}).Should(ContainSubstring("missing APIVersion"))
			})
		})
	})
	When("The namespace is labled as Pod Security Admission policy enforce:restricted", func() {
		BeforeEach(func() {
			var err error
			testNS := &corev1.Namespace{}
			Eventually(func() error {
				testNS, err = c.KubernetesInterface().CoreV1().Namespaces().Get(context.TODO(), generatedNamespace.GetName(), metav1.GetOptions{})
				if err != nil {
					return err
				}
				return nil
			}).Should(BeNil())

			testNS.ObjectMeta.Labels = map[string]string{
				"pod-security.kubernetes.io/enforce":         "restricted",
				"pod-security.kubernetes.io/enforce-version": "latest",
			}

			Eventually(func() error {
				_, err := c.KubernetesInterface().CoreV1().Namespaces().Update(context.TODO(), testNS, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
				return nil
			}).Should(BeNil())
		})
		When("A CatalogSource built with opm v1.21.0 (<v1.23.2)is created with spec.GrpcPodConfig.SecurityContextConfig set to restricted", func() {
			var sourceName string
			BeforeEach(func() {
				sourceName = genName("catalog-")
				source := &v1alpha1.CatalogSource{
					TypeMeta: metav1.TypeMeta{
						Kind:       v1alpha1.CatalogSourceKind,
						APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      sourceName,
						Namespace: generatedNamespace.GetName(),
						Labels:    map[string]string{"olm.catalogSource": sourceName},
					},
					Spec: v1alpha1.CatalogSourceSpec{
						SourceType: v1alpha1.SourceTypeGrpc,
						Image:      "quay.io/olmtest/old-opm-catsrc:v1.21.0",
						GrpcPodConfig: &v1alpha1.GrpcPodConfig{
							SecurityContextConfig: v1alpha1.Restricted,
						},
					},
				}

				Eventually(func() error {
					_, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
					return err
				}).Should(Succeed())
			})
			It("The registry pod fails to become come up because of lack of permission", func() {
				Eventually(func() (bool, error) {
					podList, err := c.KubernetesInterface().CoreV1().Pods(generatedNamespace.GetName()).List(context.TODO(), metav1.ListOptions{})
					if err != nil {
						return false, err
					}
					for _, pod := range podList.Items {
						if pod.ObjectMeta.OwnerReferences != nil && pod.ObjectMeta.OwnerReferences[0].Name == sourceName {
							if pod.Status.ContainerStatuses != nil && pod.Status.ContainerStatuses[0].State.Terminated != nil {
								return true, nil
							}
						}
					}
					return false, nil
				}).Should(BeTrue())
			})
		})
	})
	When("The namespace is labled as Pod Security Admission policy enforce:baseline", func() {
		BeforeEach(func() {
			var err error
			testNS := &corev1.Namespace{}
			Eventually(func() error {
				testNS, err = c.KubernetesInterface().CoreV1().Namespaces().Get(context.TODO(), generatedNamespace.GetName(), metav1.GetOptions{})
				if err != nil {
					return err
				}
				return nil
			}).Should(BeNil())

			testNS.ObjectMeta.Labels = map[string]string{
				"pod-security.kubernetes.io/enforce":         "baseline",
				"pod-security.kubernetes.io/enforce-version": "latest",
			}

			Eventually(func() error {
				_, err := c.KubernetesInterface().CoreV1().Namespaces().Update(context.TODO(), testNS, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
				return nil
			}).Should(BeNil())
		})
		When("A CatalogSource built with opm v1.21.0 (<v1.23.2)is created with spec.GrpcPodConfig.SecurityContextConfig set to legacy", func() {
			var sourceName string
			BeforeEach(func() {
				sourceName = genName("catalog-")
				source := &v1alpha1.CatalogSource{
					TypeMeta: metav1.TypeMeta{
						Kind:       v1alpha1.CatalogSourceKind,
						APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      sourceName,
						Namespace: generatedNamespace.GetName(),
						Labels:    map[string]string{"olm.catalogSource": sourceName},
					},
					Spec: v1alpha1.CatalogSourceSpec{
						SourceType: v1alpha1.SourceTypeGrpc,
						Image:      "quay.io/olmtest/old-opm-catsrc:v1.21.0",
						GrpcPodConfig: &v1alpha1.GrpcPodConfig{
							SecurityContextConfig: v1alpha1.Legacy,
						},
					},
				}

				Eventually(func() error {
					_, err := crc.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(context.Background(), source, metav1.CreateOptions{})
					return err
				}).Should(Succeed())
			})
			It("The registry pod comes up successfully", func() {
				Eventually(func() (bool, error) {
					podList, err := c.KubernetesInterface().CoreV1().Pods(generatedNamespace.GetName()).List(context.TODO(), metav1.ListOptions{})
					if err != nil {
						return false, err
					}
					for _, pod := range podList.Items {
						if pod.ObjectMeta.OwnerReferences != nil && pod.ObjectMeta.OwnerReferences[0].Name == sourceName {
							if pod.Status.ContainerStatuses != nil {
								if *pod.Status.ContainerStatuses[0].Started == true {
									return true, nil
								}
							}
						}
					}
					return false, nil
				}).Should(BeTrue())
			})
		})
	})
})

func getOperatorDeployment(c operatorclient.ClientInterface, namespace string, operatorLabels labels.Set) (*appsv1.Deployment, error) {
	deployments, err := c.ListDeploymentsWithLabels(namespace, operatorLabels)
	if err != nil || deployments == nil || len(deployments.Items) != 1 {
		return nil, fmt.Errorf("Error getting single operator deployment for label: %v", operatorLabels)
	}
	return &deployments.Items[0], nil
}

func rescaleDeployment(c operatorclient.ClientInterface, deployment *appsv1.Deployment) error {
	// scale down
	var replicas int32 = 0
	deployment.Spec.Replicas = &replicas
	deployment, updated, err := c.UpdateDeployment(deployment)

	if err != nil || updated == false || deployment == nil {
		return fmt.Errorf("Failed to scale down deployment")
	}

	waitForScaleup := func() (bool, error) {
		fetchedDeployment, err := c.GetDeployment(deployment.GetNamespace(), deployment.GetName())
		if err != nil {
			return true, err
		}
		if fetchedDeployment.Status.Replicas == replicas {
			return true, nil
		}

		return false, nil
	}

	// wait for deployment to scale down
	Eventually(waitForScaleup, 5*time.Minute, 1*time.Second).Should(BeTrue())

	// scale up
	replicas = 1
	deployment.Spec.Replicas = &replicas
	deployment, updated, err = c.UpdateDeployment(deployment)
	if err != nil || updated == false || deployment == nil {
		return fmt.Errorf("Failed to scale up deployment")
	}

	// wait for deployment to scale up
	Eventually(waitForScaleup, 5*time.Minute, 1*time.Second).Should(BeTrue())

	return err
}

func replicateCatalogPod(c operatorclient.ClientInterface, catalog *v1alpha1.CatalogSource) *corev1.Pod {
	initialPods, err := c.KubernetesInterface().CoreV1().Pods(catalog.GetNamespace()).List(context.Background(), metav1.ListOptions{LabelSelector: "olm.catalogSource=" + catalog.GetName()})
	Expect(err).ToNot(HaveOccurred())
	Expect(initialPods.Items).To(HaveLen(1))

	pod := initialPods.Items[0]
	copied := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: catalog.GetNamespace(),
			Name:      catalog.GetName() + "-copy",
		},
		Spec: pod.Spec,
	}

	copied, err = c.KubernetesInterface().CoreV1().Pods(catalog.GetNamespace()).Create(context.Background(), copied, metav1.CreateOptions{})
	Expect(err).ToNot(HaveOccurred())

	return copied
}
