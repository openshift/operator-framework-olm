package util

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/tidwall/gjson"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

const (
	AKSNodeLabel = "kubernetes.azure.com/cluster"
)

// GetPullSec extracts the pull secret from the cluster's openshift-config namespace
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - dirname: directory path where the extracted secret will be saved
//
// Returns:
//   - error: error if extraction fails, nil on success
func GetPullSec(oc *CLI, dirname string) error {
	if oc == nil {
		return fmt.Errorf("CLI client cannot be nil")
	}
	if dirname == "" {
		return fmt.Errorf("directory name cannot be empty")
	}

	if err := oc.AsAdmin().WithoutNamespace().Run("extract").Args("secret/pull-secret", "-n", "openshift-config", "--to="+dirname, "--confirm").Execute(); err != nil {
		return fmt.Errorf("extract pull-secret failed: %w", err)
	}
	return nil
}

// GetMirrorRegistry retrieves the mirror registry URL from the first ImageContentSourcePolicy
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - registry: mirror registry URL without path
//   - err: error if ICSP retrieval fails, nil on success
func GetMirrorRegistry(oc *CLI) (string, error) {
	if oc == nil {
		return "", fmt.Errorf("CLI client cannot be nil")
	}

	registry, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ImageContentSourcePolicy",
		"-o", "jsonpath={.items[0].spec.repositoryDigestMirrors[0].mirrors[0]}").Output()
	if err != nil {
		return "", fmt.Errorf("failed to acquire mirror registry from ICSP: %w", err)
	}

	if registry == "" {
		return "", fmt.Errorf("mirror registry not found in ICSP")
	}

	registry, _, _ = strings.Cut(registry, "/")
	return registry, nil
}

// GetUserCAToFile dumps the user certificate from user-ca-bundle configmap to a file
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - filename: path where the certificate will be written
//
// Returns:
//   - error: error if configmap retrieval or file writing fails, nil on success
func GetUserCAToFile(oc *CLI, filename string) error {
	if oc == nil {
		return fmt.Errorf("CLI client cannot be nil")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	cert, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("configmap", "-n", "openshift-config",
		"user-ca-bundle", "-o", "jsonpath={.data.ca-bundle\\.crt}").Output()
	if err != nil {
		return fmt.Errorf("failed to acquire user ca bundle from configmap: %w", err)
	}

	if cert == "" {
		return fmt.Errorf("certificate data is empty")
	}

	// Use more restrictive file permissions for security
	if err = os.WriteFile(filename, []byte(cert), 0600); err != nil {
		return fmt.Errorf("failed to dump cert to file: %w", err)
	}
	return nil
}

// GetClusterVersion returns the cluster version and build information
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: cluster version in format "X.Y" (e.g., "4.8")
//   - string: full cluster build string (e.g., "4.8.0-0.nightly-2021-09-28-165247")
//   - error: error if cluster version retrieval fails, nil on success
func GetClusterVersion(oc *CLI) (string, string, error) {
	if oc == nil {
		return "", "", fmt.Errorf("CLI client cannot be nil")
	}

	clusterBuild, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "-o", "jsonpath={..desired.version}").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get cluster version: %w", err)
	}

	if clusterBuild == "" {
		return "", "", fmt.Errorf("cluster version data is empty")
	}

	splitValues := strings.Split(clusterBuild, ".")
	if len(splitValues) < 2 {
		return "", "", fmt.Errorf("invalid cluster version format: %s", clusterBuild)
	}

	clusterVersion := splitValues[0] + "." + splitValues[1]
	return clusterVersion, clusterBuild, nil
}

// GetReleaseImage returns the release image URL used by the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: release image URL with registry and digest
//   - error: error if release image retrieval fails, nil on success
func GetReleaseImage(oc *CLI) (string, error) {
	releaseImage, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "-o", "jsonpath={..desired.image}").Output()
	if err != nil {
		return "", err
	}
	return releaseImage, nil
}

// GetInfraID returns the infrastructure ID of the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: infrastructure name/ID used for cluster resources
//   - error: error if infrastructure ID retrieval fails, nil on success
func GetInfraID(oc *CLI) (string, error) {
	infraID, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o", "jsonpath='{.status.infrastructureName}'").Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(infraID, "'"), err
}

// GetGcpProjectID returns the Google Cloud Platform project ID for GCP clusters
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: GCP project ID
//   - error: error if project ID retrieval fails, nil on success
func GetGcpProjectID(oc *CLI) (string, error) {
	projectID, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o", "jsonpath='{.status.platformStatus.gcp.projectID}'").Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(projectID, "'"), err
}

// GetClusterPrefixName returns the cluster prefix name derived from the console route
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: cluster prefix name, empty string if retrieval fails
func GetClusterPrefixName(oc *CLI) string {
	if oc == nil {
		e2e.Logf("CLI client is nil")
		return ""
	}

	output, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("route", "console", "-n", "openshift-console", "-o=jsonpath={.spec.host}").Output()
	if err != nil {
		e2e.Logf("Get cluster console route failed with err %v", err)
		return ""
	}

	if output == "" {
		e2e.Logf("Console route host is empty")
		return ""
	}

	parts := strings.Split(output, ".")
	if len(parts) < 3 {
		e2e.Logf("Invalid console route format: %s", output)
		return ""
	}

	return parts[2]
}

// SkipBaselineCaps skips the test if cluster's baselineCapabilitySet matches any of the specified sets
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - sets: comma-separated list of baselineCapabilitySets to skip (e.g., "None, v4.11")
func SkipBaselineCaps(oc *CLI, sets string) {
	if oc == nil {
		e2e.Failf("CLI client cannot be nil")
	}
	if sets == "" {
		e2e.Failf("capability sets cannot be empty")
	}

	baselineCapabilitySet, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion", "version", "-o=jsonpath={.spec.capabilities.baselineCapabilitySet}").Output()
	if err != nil {
		e2e.Failf("get baselineCapabilitySet failed err %v", err)
	}

	normalizedSets := strings.ReplaceAll(sets, " ", "")
	for _, s := range strings.Split(normalizedSets, ",") {
		if s != "" && strings.Contains(baselineCapabilitySet, s) {
			g.Skip("Skip for cluster with baselineCapabilitySet = '" + baselineCapabilitySet + "' matching filter: " + s)
		}
	}
}

// SkipNoCapabilities skips the test if the cluster does not have the specified capability enabled
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - capability: name of the capability to check for (e.g., "OperatorLifecycleManager")
func SkipNoCapabilities(oc *CLI, capability string) {
	o.Expect(oc).NotTo(o.BeNil(), "CLI client cannot be nil")
	o.Expect(capability).NotTo(o.BeEmpty(), "capability name cannot be empty")

	clusterVersion, err := oc.AdminConfigClient().ConfigV1().ClusterVersions().Get(context.Background(), "version", metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())

	hasCapability := func(capabilities []configv1.ClusterVersionCapability, checked string) bool {
		cap := configv1.ClusterVersionCapability(checked)
		for _, capability := range capabilities {
			if capability == cap {
				return true
			}
		}
		return false
	}
	if clusterVersion.Status.Capabilities.KnownCapabilities != nil &&
		hasCapability(clusterVersion.Status.Capabilities.KnownCapabilities, capability) &&
		(clusterVersion.Status.Capabilities.EnabledCapabilities == nil ||
			!hasCapability(clusterVersion.Status.Capabilities.EnabledCapabilities, capability)) {
		g.Skip(fmt.Sprintf("the cluster has no %v and skip it", capability))
	}
}

// SkipIfCapEnabled skips the test if the specified capability is enabled in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - capability: name of the capability to check for enablement
func SkipIfCapEnabled(oc *CLI, capability string) {
	clusterversion, err := oc.
		AdminConfigClient().
		ConfigV1().
		ClusterVersions().
		Get(context.Background(), "version", metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())
	var capKnown bool
	for _, knownCap := range clusterversion.Status.Capabilities.KnownCapabilities {
		if capability == string(knownCap) {
			capKnown = true
			break
		}
	}
	if !capKnown {
		g.Skip(fmt.Sprintf("Will skip as capability %s is unknown (i.e. cannot be disabled in the first place)", capability))
	}
	for _, enabledCap := range clusterversion.Status.Capabilities.EnabledCapabilities {
		if capability == string(enabledCap) {
			g.Skip(fmt.Sprintf("Will skip as capability %s is enabled", capability))
		}
	}
}

// SkipNoOLMCore skips the test if the cluster has no OLM (Operator Lifecycle Manager) component
// Note: From 4.15, OLM became an optional core component for some cluster profiles
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNoOLMCore(oc *CLI) {
	SkipNoCapabilities(oc, "OperatorLifecycleManager")
}

// SkipNoOLMv1Core skips the test if the cluster has no OLM v1 (Operator Lifecycle Manager v1) component
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNoOLMv1Core(oc *CLI) {
	SkipNoCapabilities(oc, "OperatorLifecycleManagerV1")
}

// SkipNoBuild skips the test if the cluster has no Build component
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNoBuild(oc *CLI) {
	SkipNoCapabilities(oc, "Build")
}

// SkipNoDeploymentConfig skips the test if the cluster has no DeploymentConfig component
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNoDeploymentConfig(oc *CLI) {
	SkipNoCapabilities(oc, "DeploymentConfig")
}

// SkipNoImageRegistry skips the test if the cluster has no ImageRegistry component
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNoImageRegistry(oc *CLI) {
	SkipNoCapabilities(oc, "ImageRegistry")
}

// SkipMicroshift skips the test if the cluster is microshift
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipMicroshift(oc *CLI) {
	if IsMicroshiftCluster(oc) {
		g.Skip("it does not support microshift, so skip it.")
	}
}

// IsTechPreviewNoUpgrade checks if a cluster is configured with TechPreviewNoUpgrade feature set
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster uses TechPreviewNoUpgrade feature set, false otherwise
func IsTechPreviewNoUpgrade(oc *CLI) bool {
	featureGate, err := oc.AdminConfigClient().ConfigV1().FeatureGates().Get(context.Background(), "cluster", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false
		}
		o.Expect(err).NotTo(o.HaveOccurred(), "could not retrieve feature-gate: %v", err)
	}

	return featureGate.Spec.FeatureSet == configv1.TechPreviewNoUpgrade
}

// GetAWSClusterRegion returns the AWS region where the cluster is deployed
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: AWS region name (e.g., "us-east-1")
//   - error: error if region retrieval fails, nil on success
func GetAWSClusterRegion(oc *CLI) (string, error) {
	region, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.aws.region}").Output()
	return region, err
}

// SkipNoDefaultSC skips the test if cluster has no default StorageClass or has multiple default StorageClasses
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNoDefaultSC(oc *CLI) {
	allSCRes, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sc", "-o", "json").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	defaultSCRes := gjson.Get(allSCRes, "items.#(metadata.annotations.storageclass\\.kubernetes\\.io\\/is-default-class=true)#.metadata.name")
	e2e.Logf("The default storageclass list: %s", defaultSCRes)
	defaultSCNub := len(defaultSCRes.Array())
	if defaultSCNub != 1 {
		e2e.Logf("oc get sc:\n%s", allSCRes)
		g.Skip("Skip for unexpected default storageclass!")
	}
}

// SkipIfPlatformTypeNot skips the test if cluster platform is not in the allowed list
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - platforms: comma-separated list of allowed platforms (e.g., "gcp, aws")
func SkipIfPlatformTypeNot(oc *CLI, platforms string) {
	platformType, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.type}").Output()
	if err != nil {
		e2e.Failf("get infrastructure platformStatus type failed err %v .", err)
	}
	if !strings.Contains(strings.ToLower(platforms), strings.ToLower(platformType)) {
		g.Skip("Skip for non-" + platforms + " cluster: " + platformType)
	}
}

// SkipIfPlatformType skips the test if cluster platform matches any in the specified list
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - platforms: comma-separated list of platforms to skip
func SkipIfPlatformType(oc *CLI, platforms string) {
	platformType, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.type}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	if strings.Contains(strings.ToLower(platforms), strings.ToLower(platformType)) {
		g.Skip("Skip for " + platforms + " cluster: " + platformType)
	}
}

// isCRDSpecificFieldExist checks whether the specified CRD field exists in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - crdFieldPath: path to the CRD field to check (e.g., "template.apiVersion")
//
// Returns:
//   - bool: true if the CRD field exists, false otherwise
func isCRDSpecificFieldExist(oc *CLI, crdFieldPath string) bool {
	var (
		crdFieldInfo string
		getInfoErr   error
	)
	err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		crdFieldInfo, getInfoErr = oc.AsAdmin().WithoutNamespace().Run("explain").Args(crdFieldPath).Output()
		if getInfoErr != nil && strings.Contains(crdFieldInfo, "the server doesn't have a resource type") {
			if strings.Contains(crdFieldInfo, "the server doesn't have a resource type") {
				e2e.Logf("The test cluster specified crd field: %s is not exist.", crdFieldPath)
				return true, nil
			}
			// TODO: The "couldn't find resource" error info sometimes(very low frequency) happens in few cases but I couldn't reproduce it, this retry solution should be an enhancement
			if strings.Contains(getInfoErr.Error(), "couldn't find resource") {
				e2e.Logf("Failed to check whether the specified crd field: %s exist, try again. Err:\n%v", crdFieldPath, getInfoErr)
				return false, nil
			}
			return false, getInfoErr
		}
		return true, nil
	})
	AssertWaitPollNoErr(err, fmt.Sprintf("Check whether the specified: %s crd field exist timeout.", crdFieldPath))
	return !strings.Contains(crdFieldInfo, "the server doesn't have a resource type")
}

// IsMicroshiftCluster determines whether the cluster is a MicroShift cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster is MicroShift, false otherwise
func IsMicroshiftCluster(oc *CLI) bool {
	return !isCRDSpecificFieldExist(oc, "template.apiVersion")
}

// IsHypershiftHostedCluster determines whether the cluster is a HyperShift hosted cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster has external control plane topology, false otherwise
func IsHypershiftHostedCluster(oc *CLI) bool {
	topology, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("infrastructures.config.openshift.io", "cluster", "-o=jsonpath={.status.controlPlaneTopology}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("topology is %s", topology)
	if topology == "" {
		status, _ := oc.WithoutNamespace().AsAdmin().Run("get").Args("infrastructures.config.openshift.io", "cluster", "-o=jsonpath={.status}").Output()
		e2e.Logf("cluster status %s", status)
		e2e.Failf("failure: controlPlaneTopology returned empty")
	}
	return strings.Compare(topology, "External") == 0
}

// IsRosaCluster determines whether the cluster is a Red Hat OpenShift Service on AWS (ROSA) cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster is ROSA, false otherwise
func IsRosaCluster(oc *CLI) bool {
	product, _ := oc.WithoutNamespace().AsAdmin().Run("get").Args("clusterclaims/product.open-cluster-management.io", "-o=jsonpath={.spec.value}").Output()
	return strings.Compare(product, "ROSA") == 0
}

// IsSTSCluster determines if an AWS cluster is using Security Token Service (STS) for authentication
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster uses STS, false otherwise
func IsSTSCluster(oc *CLI) bool {
	return IsWorkloadIdentityCluster(oc)
}

// IsWorkloadIdentityCluster determines whether the Azure/GCP cluster is using Workload Identity
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster uses Workload Identity (has serviceAccountIssuer configured), false otherwise
func IsWorkloadIdentityCluster(oc *CLI) bool {
	serviceAccountIssuer, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("authentication", "cluster", "-o=jsonpath={.spec.serviceAccountIssuer}").Output()
	o.Expect(err).ShouldNot(o.HaveOccurred(), "Failed to get serviceAccountIssuer")
	return len(serviceAccountIssuer) > 0
}

// GetOIDCProvider returns the OIDC provider URL for the current cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: OIDC provider URL without "https://" prefix
//   - error: error if OIDC provider retrieval fails, nil on success
func GetOIDCProvider(oc *CLI) (string, error) {
	oidc, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("authentication.config", "cluster", "-o=jsonpath={.spec.serviceAccountIssuer}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(oidc, "https://"), nil
}

// SkipMissingQECatalogsource skips the test if the qe-app-registry CatalogSource is not present in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipMissingQECatalogsource(oc *CLI) {
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", "openshift-marketplace", "catalogsource", "qe-app-registry").Output()
	if strings.Contains(output, "NotFound") || err != nil {
		g.Skip("Skip the test since no catalogsource/qe-app-registry in the cluster")
	}
}

// SkipIfDisableDefaultCatalogsource skips the test if default CatalogSources are disabled in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipIfDisableDefaultCatalogsource(oc *CLI) {
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("operatorhubs", "cluster", "-o=jsonpath={.spec.disableAllDefaultSources}").Output()
	if output == "true" || err != nil {
		g.Skip("Skip the test, the default catsrc is disable or don't have operatorhub resource")
	}
}

// IsInfrastructuresHighlyAvailable checks if the cluster infrastructure is highly available
// Compatible with both classic OpenShift and hosted clusters
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if infrastructure topology is HighlyAvailable, false otherwise
func IsInfrastructuresHighlyAvailable(oc *CLI) bool {
	topology, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructures.config.openshift.io", "cluster", `-o=jsonpath={.status.infrastructureTopology}`).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("infrastructures topology is %s", topology)
	if topology == "" {
		status, _ := oc.WithoutNamespace().AsAdmin().Run("get").Args("infrastructures.config.openshift.io", "cluster", "-o=jsonpath={.status}").Output()
		e2e.Logf("cluster status %s", status)
		e2e.Failf("failure: controlPlaneTopology returned empty")
	}
	return strings.Compare(topology, "HighlyAvailable") == 0
}

// IsExternalOIDCCluster checks if the cluster is configured to use external OIDC authentication
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster uses external OIDC authentication, false otherwise
//   - error: error if authentication type retrieval fails, nil on success
func IsExternalOIDCCluster(oc *CLI) (bool, error) {
	authType, stdErr, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("authentication/cluster", "-o=jsonpath={.spec.type}").Outputs()
	if err != nil {
		return false, fmt.Errorf("error checking if the cluster is using external OIDC: %v", stdErr)
	}
	e2e.Logf("Found authentication type used: %v", authType)

	return authType == string(configv1.AuthenticationTypeOIDC), nil
}

// IsOpenShiftCluster checks if the active cluster is OpenShift or a derivative
// Parameters:
//   - ctx: context for the API call
//   - c: Kubernetes namespace interface for checking OpenShift-specific namespaces
//
// Returns:
//   - bool: true if cluster is OpenShift, false if vanilla Kubernetes
//   - error: error if determination fails, nil on success
func IsOpenShiftCluster(ctx context.Context, c corev1client.NamespaceInterface) (bool, error) {
	switch _, err := c.Get(ctx, "openshift-controller-manager", metav1.GetOptions{}); {
	case err == nil:
		return true, nil
	case apierrors.IsNotFound(err):
		return false, nil
	default:
		return false, fmt.Errorf("unable to determine if we are running against an OpenShift cluster: %v", err)
	}
}

// SkipOnOpenShiftNess skips the test if the cluster type doesn't match the expected type
// Parameters:
//   - expectOpenShift: true to expect OpenShift cluster, false to expect vanilla Kubernetes
func SkipOnOpenShiftNess(expectOpenShift bool) {
	switch IsKubernetesClusterFlag {
	case "yes":
		if expectOpenShift {
			g.Skip("Expecting OpenShift but the active cluster is not, skipping the test")
		}
	// Treat both "no" and "unknown" as OpenShift
	default:
		if !expectOpenShift {
			g.Skip("Expecting non-OpenShift but the active cluster is OpenShift, skipping the test")
		}
	}
}

// IsAKSCluster checks if the active cluster is an Azure Kubernetes Service (AKS) cluster
// Parameters:
//   - ctx: context for the API call
//   - oc: CLI client for interacting with the cluster
//
// Returns:
//   - bool: true if cluster is AKS, false otherwise
//   - error: error if determination fails, nil on success
func IsAKSCluster(ctx context.Context, oc *CLI) (bool, error) {
	if oc == nil {
		return false, fmt.Errorf("CLI client cannot be nil")
	}

	nodeList, err := oc.AdminKubeClient().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodeList.Items) == 0 {
		return false, fmt.Errorf("no nodes found in cluster")
	}

	_, labelFound := nodeList.Items[0].Labels[AKSNodeLabel]
	return labelFound, nil
}

// CheckAKSCluster checks if the cluster is AKS with error handling fallback
// Parameters:
//   - ctx: context for the API call
//   - oc: CLI client for interacting with the cluster
//
// Returns:
//   - bool: true if cluster is AKS, false otherwise (defaults to false on error)
func CheckAKSCluster(ctx context.Context, oc *CLI) bool {
	isAKS, err := IsAKSCluster(ctx, oc)
	if err != nil {
		e2e.Logf("failed to determine if the active cluster is AKS or not: %v, defaulting to non-AKS", err)
		return false
	}
	return isAKS
}

// SkipOnAKSNess skips the test if the cluster AKS type doesn't match the expected type
// Parameters:
//   - ctx: context for the API call
//   - oc: CLI client for interacting with the cluster
//   - expectAKS: true to expect AKS cluster, false to expect non-AKS cluster
func SkipOnAKSNess(ctx context.Context, oc *CLI, expectAKS bool) {
	isAKS := CheckAKSCluster(ctx, oc)
	if isAKS && !expectAKS {
		g.Skip("Expecting non-AKS but the active cluster is AKS, skip the test")
	}
	if !isAKS && expectAKS {
		g.Skip("Expecting AKS but the active cluster is not, skip the test")
	}
}

// SkipOnProxyCluster skips the test if the cluster is configured with HTTP/HTTPS proxy
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipOnProxyCluster(oc *CLI) {
	g.By("Check if cluster is a proxy platform")
	httpProxy, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy/cluster", "-o=jsonpath={.spec.httpProxy}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	httpsProxy, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy/cluster", "-o=jsonpath={.spec.httpsProxy}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	if len(httpProxy) != 0 || len(httpsProxy) != 0 {
		g.Skip("Skip for proxy platform")
	}
}

// CheckPlatform returns the cluster's platform type in lowercase
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: platform type in lowercase (e.g., "aws", "gcp", "azure")
func CheckPlatform(oc *CLI) string {
	output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.type}").Output()
	return strings.ToLower(output)
}

// SkipForSNOCluster skips the test if the cluster is a Single Node OpenShift (SNO) cluster
// SNO is identified by having only 1 master and 1 worker node with the same hostname
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipForSNOCluster(oc *CLI) {
	if oc == nil {
		e2e.Logf("CLI client is nil, cannot determine SNO status")
		return
	}

	// Only 1 master, 1 worker node and with the same hostname.
	masterNodes, err := GetClusterNodesBy(oc, "master")
	if err != nil {
		e2e.Logf("Failed to get master nodes: %v", err)
		return
	}

	workerNodes, err := GetClusterNodesBy(oc, "worker")
	if err != nil {
		e2e.Logf("Failed to get worker nodes: %v", err)
		return
	}

	if len(masterNodes) == 1 && len(workerNodes) == 1 &&
		len(masterNodes[0]) > 0 && len(workerNodes[0]) > 0 &&
		masterNodes[0] == workerNodes[0] {
		g.Skip("Skip for SNO cluster.")
	}
}
