// Package olmv0util provides utility functions for OLM v0 operator testing
// This file contains CatalogSource management utilities for creating, configuring,
// and managing operator catalog sources in OpenShift/Kubernetes environments
package olmv0util

import (
	"context"
	"fmt"
	"strings"
	"time"

	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// CatalogSourceDescription represents a CatalogSource resource configuration
// Used for defining and managing operator catalogs that provide operator bundles
type CatalogSourceDescription struct {
	Name          string // Name of the catalog source
	Namespace     string // Namespace where the catalog source will be created
	DisplayName   string // Human-readable name for the catalog
	Publisher     string // Publisher information for the catalog
	SourceType    string // Type of catalog source (e.g., "grpc")
	Address       string // Address or image reference for the catalog
	Template      string // Template file path for creating the catalog
	Priority      int    // Priority for catalog source selection
	Secret        string // Secret reference for authenticated catalogs
	Interval      string // Update interval for the catalog source
	ImageTemplate string // Image template for catalog updates
	ClusterType   string // Target cluster type (e.g., "microshift")
}

// Create creates a CatalogSource resource using the provided template and parameters
// The method applies resource templates, configures security contexts for restricted environments,
// and registers the resource for cleanup in the test framework
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (catsrc *CatalogSourceDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// Set default update interval if not specified
	if strings.Compare(catsrc.Interval, "") == 0 {
		catsrc.Interval = "10m0s"
		e2e.Logf("set interval to be 10m0s")
	}
	// Choose appropriate template application function based on cluster type
	applyFn := ApplyResourceFromTemplate
	if strings.Compare(catsrc.ClusterType, "microshift") == 0 {
		applyFn = ApplyResourceFromTemplateOnMicroshift
	}
	// Apply the catalog source template with all configured parameters
	err := applyFn(oc, "--ignore-unknown-parameters=true", "-f", catsrc.Template,
		"-p", "NAME="+catsrc.Name, "NAMESPACE="+catsrc.Namespace, "ADDRESS="+catsrc.Address, "SECRET="+catsrc.Secret,
		"DISPLAYNAME="+"\""+catsrc.DisplayName+"\"", "PUBLISHER="+"\""+catsrc.Publisher+"\"", "SOURCETYPE="+catsrc.SourceType,
		"INTERVAL="+catsrc.Interval, "IMAGETEMPLATE="+catsrc.ImageTemplate)
	o.Expect(err).NotTo(o.HaveOccurred())
	// Configure security context constraints for non-microshift clusters
	if strings.Compare(catsrc.ClusterType, "microshift") != 0 {
		catsrc.SetSCCRestricted(oc)
	}
	// Register the created resource for test cleanup
	dr.GetIr(itName).Add(NewResource(oc, "catsrc", catsrc.Name, exutil.RequireNS, catsrc.Namespace))
	e2e.Logf("create catsrc %s SUCCESS", catsrc.Name)
}

// SetSCCRestricted configures security context constraints for the catalog source
// This method sets the securityContextConfig to "restricted" for catalogs that need
// to run in security-restricted environments, particularly when Pod Security Standards are enforced
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (catsrc *CatalogSourceDescription) SetSCCRestricted(oc *exutil.CLI) {
	// Skip SCC configuration for system marketplace namespace
	if strings.Compare(catsrc.Namespace, "openshift-marketplace") == 0 {
		e2e.Logf("the namespace is openshift-marketplace, skip setting SCC")
		return
	}
	// Determine Pod Security Admission (PSA) enforcement level
	psa := "restricted"
	if exutil.IsHypershiftHostedCluster(oc) {
		// Hypershift clusters may not have accessible kube-apiserver config
		e2e.Logf("cluster is Hypershift Hosted Cluster, cannot get default PSA setting, use default value restricted")
	} else {
		// Retrieve PSA enforcement level from cluster configuration
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("configmaps", "-n", "openshift-kube-apiserver", "config", `-o=jsonpath={.data.config\.yaml}`).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		psa = gjson.Get(output, "admission.pluginConfig.PodSecurity.configuration.defaults.enforce").String()
		e2e.Logf("pod-security.kubernetes.io/enforce is %s", string(psa))
	}
	// Apply restricted security context if PSA enforcement requires it
	if strings.Contains(string(psa), "restricted") {
		originSCC := catsrc.getSCC(oc)
		e2e.Logf("spec.grpcPodConfig.securityContextConfig is %s", originSCC)
		if strings.Compare(originSCC, "") == 0 {
			// Set restricted security context for catalog source pods
			e2e.Logf("set spec.grpcPodConfig.securityContextConfig to be restricted")
			err := oc.AsAdmin().WithoutNamespace().Run("patch").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "--type=merge", "-p", `{"spec":{"grpcPodConfig":{"securityContextConfig":"restricted"}}}`).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
		} else {
			// Security context already configured, skip modification
			e2e.Logf("spec.grpcPodConfig.securityContextConfig is %s, skip setting", originSCC)
		}

	}
}

// getSCC retrieves the current security context configuration for the catalog source
// Returns the current securityContextConfig value, or empty string if not set
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//
// Returns:
//   - string: Current security context configuration value
func (catsrc *CatalogSourceDescription) getSCC(oc *exutil.CLI) string {
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.spec.grpcPodConfig.securityContextConfig}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return output
}

// CreateWithCheck creates a CatalogSource and waits for it to reach READY state
// This method combines resource creation with status verification, ensuring the catalog
// is fully operational before proceeding with tests
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (catsrc *CatalogSourceDescription) CreateWithCheck(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// Create the catalog source resource
	catsrc.Create(oc, itName, dr)
	// Wait for catalog source to reach READY state
	err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
		status, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status..lastObservedState}").Output()
		if strings.Compare(status, "READY") != 0 {
			e2e.Logf("catsrc %s lastObservedState is %s, not READY", catsrc.Name, status)
			return false, nil
		}
		return true, nil
	})
	// Collect debug information if catalog source fails to become ready
	if err != nil {
		output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", catsrc.Name, "-n", catsrc.Namespace, "-o=jsonpath={.status}").Output()
		e2e.Logf("CatalogSource status: %s", output)
		LogDebugInfo(oc, catsrc.Namespace, "pod", "events")
	}
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("catsrc %s lastObservedState is not READY", catsrc.Name))
	e2e.Logf("catsrc %s lastObservedState is READY", catsrc.Name)
}

// Delete removes the CatalogSource resource from the cluster
// This method unregisters the resource from the test cleanup framework
//
// Parameters:
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (catsrc *CatalogSourceDescription) Delete(itName string, dr DescriberResrouce) {
	e2e.Logf("delete carsrc %s, ns is %s", catsrc.Name, catsrc.Namespace)
	dr.GetIr(itName).Remove(catsrc.Name, "catsrc", catsrc.Namespace)
}
