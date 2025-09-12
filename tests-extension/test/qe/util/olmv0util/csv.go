// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains ClusterServiceVersion (CSV) management utilities
package olmv0util

import (
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// CsvDescription represents a ClusterServiceVersion resource configuration
// CSV resources define operator metadata, installation requirements, and permissions
type CsvDescription struct {
	Name      string // Name of the ClusterServiceVersion
	Namespace string // Namespace where the CSV is installed
}

// Delete removes the ClusterServiceVersion resource from the cluster
// This method unregisters the CSV from the test cleanup framework, allowing
// the test infrastructure to properly clean up operator installations
//
// Parameters:
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (csv CsvDescription) Delete(itName string, dr DescriberResrouce) {
	e2e.Logf("remove %s, ns %s", csv.Name, csv.Namespace)
	dr.GetIr(itName).Remove(csv.Name, "csv", csv.Namespace)
}
