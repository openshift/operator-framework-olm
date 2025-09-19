// Code generated for package testdata by go-bindata DO NOT EDIT. (@generated)
// sources:
// test/qe/testdata/olm/catalogsource-address.yaml
// test/qe/testdata/olm/catalogsource-configmap.yaml
// test/qe/testdata/olm/catalogsource-image-cacheless.yaml
// test/qe/testdata/olm/catalogsource-image-extract.yaml
// test/qe/testdata/olm/catalogsource-image-incorrect-updatestrategy.yaml
// test/qe/testdata/olm/catalogsource-image.yaml
// test/qe/testdata/olm/catalogsource-legacy.yaml
// test/qe/testdata/olm/catalogsource-namespace.yaml
// test/qe/testdata/olm/catalogsource-opm.yaml
// test/qe/testdata/olm/cm-21824-correct.yaml
// test/qe/testdata/olm/cm-21824-wrong.yaml
// test/qe/testdata/olm/cm-25644-etcd-csv.yaml
// test/qe/testdata/olm/cm-csv-etcd.yaml
// test/qe/testdata/olm/cm-namespaceconfig.yaml
// test/qe/testdata/olm/cm-template.yaml
// test/qe/testdata/olm/configmap-ectd-alpha-beta.yaml
// test/qe/testdata/olm/configmap-etcd.yaml
// test/qe/testdata/olm/configmap-test.yaml
// test/qe/testdata/olm/configmap-with-defaultchannel.yaml
// test/qe/testdata/olm/configmap-without-defaultchannel.yaml
// test/qe/testdata/olm/cr-webhookTest.yaml
// test/qe/testdata/olm/cr_devworkspace.yaml
// test/qe/testdata/olm/cr_pgadmin.yaml
// test/qe/testdata/olm/cs-image-template.yaml
// test/qe/testdata/olm/cs-without-image.yaml
// test/qe/testdata/olm/cs-without-interval.yaml
// test/qe/testdata/olm/cs-without-scc.yaml
// test/qe/testdata/olm/csc.yaml
// test/qe/testdata/olm/env-subscription.yaml
// test/qe/testdata/olm/envfrom-subscription.yaml
// test/qe/testdata/olm/etcd-cluster.yaml
// test/qe/testdata/olm/etcd-subscription-manual.yaml
// test/qe/testdata/olm/etcd-subscription.yaml
// test/qe/testdata/olm/mc-workload-partition.yaml
// test/qe/testdata/olm/og-allns.yaml
// test/qe/testdata/olm/og-multins.yaml
// test/qe/testdata/olm/olm-proxy-subscription.yaml
// test/qe/testdata/olm/olm-subscription.yaml
// test/qe/testdata/olm/operator.yaml
// test/qe/testdata/olm/operatorgroup-serviceaccount.yaml
// test/qe/testdata/olm/operatorgroup-upgradestrategy.yaml
// test/qe/testdata/olm/operatorgroup.yaml
// test/qe/testdata/olm/opsrc.yaml
// test/qe/testdata/olm/packageserver.yaml
// test/qe/testdata/olm/platform_operator.yaml
// test/qe/testdata/olm/prometheus-antiaffinity.yaml
// test/qe/testdata/olm/prometheus-nodeaffinity.yaml
// test/qe/testdata/olm/role-binding.yaml
// test/qe/testdata/olm/role.yaml
// test/qe/testdata/olm/scc.yaml
// test/qe/testdata/olm/scoped-sa-etcd.yaml
// test/qe/testdata/olm/scoped-sa-fine-grained-roles.yaml
// test/qe/testdata/olm/scoped-sa-roles.yaml
// test/qe/testdata/olm/secret.yaml
// test/qe/testdata/olm/secret_opaque.yaml
// test/qe/testdata/olm/vpa-crd.yaml
package testdata

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _testQeTestdataOlmCatalogsourceAddressYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    address: "${ADDRESS}"
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
`)

func testQeTestdataOlmCatalogsourceAddressYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceAddressYaml, nil
}

func testQeTestdataOlmCatalogsourceAddressYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceAddressYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-address.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceConfigmapYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    configMap: "${ADDRESS}"
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
`)

func testQeTestdataOlmCatalogsourceConfigmapYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceConfigmapYaml, nil
}

func testQeTestdataOlmCatalogsourceConfigmapYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceImageCachelessYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    grpcPodConfig:
      extractContent:
        catalogDir: /configs
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"
`)

func testQeTestdataOlmCatalogsourceImageCachelessYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceImageCachelessYaml, nil
}

func testQeTestdataOlmCatalogsourceImageCachelessYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceImageCachelessYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-image-cacheless.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceImageExtractYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    grpcPodConfig:
      extractContent:
        cacheDir: /tmp/cache
        catalogDir: /configs
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"
`)

func testQeTestdataOlmCatalogsourceImageExtractYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceImageExtractYaml, nil
}

func testQeTestdataOlmCatalogsourceImageExtractYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceImageExtractYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-image-extract.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy: 
      registryPoll: {}
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET`)

func testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYaml, nil
}

func testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-image-incorrect-updatestrategy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceImageYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"
`)

func testQeTestdataOlmCatalogsourceImageYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceImageYaml, nil
}

func testQeTestdataOlmCatalogsourceImageYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceImageYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-image.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceLegacyYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    grpcPodConfig: 
          securityContextConfig: legacy
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"
`)

func testQeTestdataOlmCatalogsourceLegacyYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceLegacyYaml, nil
}

func testQeTestdataOlmCatalogsourceLegacyYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceLegacyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-legacy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceNamespaceYaml = []byte(`---
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: scenario3
  namespace: scenario3
spec:
  sourceType: internal
  configMap: scenario3
  displayName: Scenario 3 Operators
  publisher: Red Hat
`)

func testQeTestdataOlmCatalogsourceNamespaceYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceNamespaceYaml, nil
}

func testQeTestdataOlmCatalogsourceNamespaceYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-namespace.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCatalogsourceOpmYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    grpcPodConfig:
      extractContent:
        cacheDir: /tmp/cache
        catalogDir: /configs
      memoryTarget: 30Mi
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    icon:
      base64data: ""
      mediatype: ""
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"
`)

func testQeTestdataOlmCatalogsourceOpmYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCatalogsourceOpmYaml, nil
}

func testQeTestdataOlmCatalogsourceOpmYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCatalogsourceOpmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/catalogsource-opm.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCm21824CorrectYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-21824-template
objects:
- apiVersion: v1
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: '[{"apiVersion":"charts.helm.k8s.io/v1","kind":"Kubeturbo21824","metadata":{"name":"kubeturbo21824-release"},"spec":{"serverMeta":{"turboServer":"https://Turbo_server_URL"},"targetConfig":{"targetName":"Cluster_Name"}}}]'
            capabilities: Basic Install
            categories: Monitoring
            certified: "false"
            containerImage: quay.io/olmqe/kubeturbo-operator-base:8.5-multi-arch
            createdAt: 2019-05-01T00:00:00.000Z
            description: Turbonomic Workload Automation for Multicloud simultaneously optimizes performance, compliance, and cost in real-time. Workloads are precisely resourced, automatically, to perform while satisfying business constraints.
            repository: https://github.com/turbonomic/kubeturbo21824/tree/master/deploy/kubeturbo21824-operator
            support: Turbonomic, Inc.
          labels:
            operatorframework.io/arch.amd64: supported
            operatorframework.io/arch.arm64: supported
            operatorframework.io/arch.ppc64le: supported
            operatorframework.io/arch.s390x: supported
          name: kubeturbo21824-operator.v8.5.0
          namespace: placeholder
        spec:
          apiservicedefinitions: {}
          customresourcedefinitions:
            owned:
            - description: Turbonomic Workload Automation for Multicloud simultaneously optimizes performance, compliance, and cost in real-time. Workloads are precisely resourced, automatically, to perform while satisfying business constraints.
              displayName: Kubeturbo21824 Operator
              kind: Kubeturbo21824
              name: kubeturbo21824s.charts.helm.k8s.io
              version: v1
          description: |-
            ### Application Resource Management for Kubernetes
            Turbonomic AI-powered Application Resource Management simultaneously optimizes performance, compliance, and cost in real time.
            Software manages the complete application stack, automatically. Applications are continually resourced to perform while satisfying business constraints.

            Turbonomic makes workloads smart — enabling them to self-manage and determines the specific actions that will drive continuous health:

            * Continuous placement for Pods (rescheduling)
            * Continuous scaling for applications and  the underlying cluster.

            It assures application performance by giving workloads the resources they need when they need them.

            ### How does it work?
            Turbonomic uses a container — KubeTurbo — that runs in your Kubernetes or Red Hat OpenShift cluster to discover and monitor your environment.
            KubeTurbo runs together with the default scheduler and sends this data back to the Turbonomic analytics engine.
            Turbonomic determines the right actions that drive continuous health, including continuous placement for Pods and continuous scaling for applications and the underlying cluster.
          displayName: Kubeturbo21824 Operator
          icon:
            - base64data: iVBORw0KGgoAAAANSUhEUgAAAfQAAACzCAYAAAB//O7qAAAACXBIWXMAAC4jAAAuIwF4pT92AABD52lUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4KPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxMzIgNzkuMTU5Mjg0LCAyMDE2LzA0LzE5LTEzOjEzOjQwICAgICAgICAiPgogICA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPgogICAgICA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIgogICAgICAgICAgICB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iCiAgICAgICAgICAgIHhtbG5zOnBob3Rvc2hvcD0iaHR0cDovL25zLmFkb2JlLmNvbS9waG90b3Nob3AvMS4wLyIKICAgICAgICAgICAgeG1sbnM6ZGM9Imh0dHA6Ly9wdXJsLm9yZy9kYy9lbGVtZW50cy8xLjEvIgogICAgICAgICAgICB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIKICAgICAgICAgICAgeG1sbnM6c3RFdnQ9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZUV2ZW50IyIKICAgICAgICAgICAgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiCiAgICAgICAgICAgIHhtbG5zOnRpZmY9Imh0dHA6Ly9ucy5hZG9iZS5jb20vdGlmZi8xLjAvIgogICAgICAgICAgICB4bWxuczpleGlmPSJodHRwOi8vbnMuYWRvYmUuY29tL2V4aWYvMS4wLyI+CiAgICAgICAgIDx4bXA6Q3JlYXRvclRvb2w+QWRvYmUgUGhvdG9zaG9wIENDIDIwMTUuNSAoTWFjaW50b3NoKTwveG1wOkNyZWF0b3JUb29sPgogICAgICAgICA8eG1wOkNyZWF0ZURhdGU+MjAxMy0xMi0wNFQxNToxMTozNC0wNTowMDwveG1wOkNyZWF0ZURhdGU+CiAgICAgICAgIDx4bXA6TWV0YWRhdGFEYXRlPjIwMTYtMDgtMTFUMTQ6MzI6NTUrMDM6MDA8L3htcDpNZXRhZGF0YURhdGU+CiAgICAgICAgIDx4bXA6TW9kaWZ5RGF0ZT4yMDE2LTA4LTExVDE0OjMyOjU1KzAzOjAwPC94bXA6TW9kaWZ5RGF0ZT4KICAgICAgICAgPHBob3Rvc2hvcDpDb2xvck1vZGU+MzwvcGhvdG9zaG9wOkNvbG9yTW9kZT4KICAgICAgICAgPHBob3Rvc2hvcDpEb2N1bWVudEFuY2VzdG9ycz4KICAgICAgICAgICAgPHJkZjpCYWc+CiAgICAgICAgICAgICAgIDxyZGY6bGk+QTkxQjdEQ0MwNEMwQzdBOERGRDVDMTVGNDgwMzY3Njc8L3JkZjpsaT4KICAgICAgICAgICAgICAgPHJkZjpsaT51dWlkOjcxMzQxOTYxLTU5ODctZTE0Ny1iZjA3LTA2MmE5OTNiM2I3YTwvcmRmOmxpPgogICAgICAgICAgICAgICA8cmRmOmxpPnhtcC5kaWQ6MDQ4MDExNzQwNzIwNjgxMTgwODM5RjI5RUMyOTAwODg8L3JkZjpsaT4KICAgICAgICAgICAgPC9yZGY6QmFnPgogICAgICAgICA8L3Bob3Rvc2hvcDpEb2N1bWVudEFuY2VzdG9ycz4KICAgICAgICAgPGRjOmZvcm1hdD5pbWFnZS9wbmc8L2RjOmZvcm1hdD4KICAgICAgICAgPHhtcE1NOkluc3RhbmNlSUQ+eG1wLmlpZDo1MzI1MGY1Ni05MzRhLTQ1N2MtYTEwMS0zZjY0MmNiZmQxOTY8L3htcE1NOkluc3RhbmNlSUQ+CiAgICAgICAgIDx4bXBNTTpEb2N1bWVudElEPnhtcC5kaWQ6MDQ4MDExNzQwNzIwNjgxMTgwODM5RjI5RUMyOTAwODg8L3htcE1NOkRvY3VtZW50SUQ+CiAgICAgICAgIDx4bXBNTTpPcmlnaW5hbERvY3VtZW50SUQ+eG1wLmRpZDowNDgwMTE3NDA3MjA2ODExODA4MzlGMjlFQzI5MDA4ODwveG1wTU06T3JpZ2luYWxEb2N1bWVudElEPgogICAgICAgICA8eG1wTU06SGlzdG9yeT4KICAgICAgICAgICAgPHJkZjpTZXE+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPmNyZWF0ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0Omluc3RhbmNlSUQ+eG1wLmlpZDowNDgwMTE3NDA3MjA2ODExODA4MzlGMjlFQzI5MDA4ODwvc3RFdnQ6aW5zdGFuY2VJRD4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OndoZW4+MjAxMy0xMi0wNFQxNToxMTozNC0wNTowMDwvc3RFdnQ6d2hlbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnNvZnR3YXJlQWdlbnQ+QWRvYmUgUGhvdG9zaG9wIENTNiAoTWFjaW50b3NoKTwvc3RFdnQ6c29mdHdhcmVBZ2VudD4KICAgICAgICAgICAgICAgPC9yZGY6bGk+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPnNhdmVkPC9zdEV2dDphY3Rpb24+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDppbnN0YW5jZUlEPnhtcC5paWQ6OTZDNDQxRTcwQjZDRTMxMTg3Q0ZCQjM3Mzg4MzY1MTA8L3N0RXZ0Omluc3RhbmNlSUQ+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDp3aGVuPjIwMTMtMTItMjNUMTY6NTM6NTktMDU6MDA8L3N0RXZ0OndoZW4+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDpzb2Z0d2FyZUFnZW50PkFkb2JlIFBob3Rvc2hvcCBDUzYgKFdpbmRvd3MpPC9zdEV2dDpzb2Z0d2FyZUFnZW50PgogICAgICAgICAgICAgICAgICA8c3RFdnQ6Y2hhbmdlZD4vPC9zdEV2dDpjaGFuZ2VkPgogICAgICAgICAgICAgICA8L3JkZjpsaT4KICAgICAgICAgICAgICAgPHJkZjpsaSByZGY6cGFyc2VUeXBlPSJSZXNvdXJjZSI+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDphY3Rpb24+c2F2ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0Omluc3RhbmNlSUQ+eG1wLmlpZDpiYWFhNDExNC1jNjc5LTMzNDMtYjI5Ny1jZTc3Y2IwYTRlM2E8L3N0RXZ0Omluc3RhbmNlSUQ+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDp3aGVuPjIwMTQtMDQtMDJUMTQ6Mjk6MzEtMDQ6MDA8L3N0RXZ0OndoZW4+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDpzb2Z0d2FyZUFnZW50PkFkb2JlIFBob3Rvc2hvcCBDQyAoV2luZG93cyk8L3N0RXZ0OnNvZnR3YXJlQWdlbnQ+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDpjaGFuZ2VkPi88L3N0RXZ0OmNoYW5nZWQ+CiAgICAgICAgICAgICAgIDwvcmRmOmxpPgogICAgICAgICAgICAgICA8cmRmOmxpIHJkZjpwYXJzZVR5cGU9IlJlc291cmNlIj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OmFjdGlvbj5jb252ZXJ0ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnBhcmFtZXRlcnM+ZnJvbSBhcHBsaWNhdGlvbi92bmQuYWRvYmUucGhvdG9zaG9wIHRvIGltYWdlL3BuZzwvc3RFdnQ6cGFyYW1ldGVycz4KICAgICAgICAgICAgICAgPC9yZGY6bGk+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPmRlcml2ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnBhcmFtZXRlcnM+Y29udmVydGVkIGZyb20gYXBwbGljYXRpb24vdm5kLmFkb2JlLnBob3Rvc2hvcCB0byBpbWFnZS9wbmc8L3N0RXZ0OnBhcmFtZXRlcnM+CiAgICAgICAgICAgICAgIDwvcmRmOmxpPgogICAgICAgICAgICAgICA8cmRmOmxpIHJkZjpwYXJzZVR5cGU9IlJlc291cmNlIj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OmFjdGlvbj5zYXZlZDwvc3RFdnQ6YWN0aW9uPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6aW5zdGFuY2VJRD54bXAuaWlkOjM4ZjZlNDQ0LTFiZWMtYWQ0Zi1hZDUzLTQ3ODVjOTlhZjk4Mjwvc3RFdnQ6aW5zdGFuY2VJRD4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OndoZW4+MjAxNC0wNC0wMlQxNDoyOTozMS0wNDowMDwvc3RFdnQ6d2hlbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnNvZnR3YXJlQWdlbnQ+QWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKTwvc3RFdnQ6c29mdHdhcmVBZ2VudD4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OmNoYW5nZWQ+Lzwvc3RFdnQ6Y2hhbmdlZD4KICAgICAgICAgICAgICAgPC9yZGY6bGk+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPnNhdmVkPC9zdEV2dDphY3Rpb24+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDppbnN0YW5jZUlEPnhtcC5paWQ6NTMyNTBmNTYtOTM0YS00NTdjLWExMDEtM2Y2NDJjYmZkMTk2PC9zdEV2dDppbnN0YW5jZUlEPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6d2hlbj4yMDE2LTA4LTExVDE0OjMyOjU1KzAzOjAwPC9zdEV2dDp3aGVuPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6c29mdHdhcmVBZ2VudD5BZG9iZSBQaG90b3Nob3AgQ0MgMjAxNS41IChNYWNpbnRvc2gpPC9zdEV2dDpzb2Z0d2FyZUFnZW50PgogICAgICAgICAgICAgICAgICA8c3RFdnQ6Y2hhbmdlZD4vPC9zdEV2dDpjaGFuZ2VkPgogICAgICAgICAgICAgICA8L3JkZjpsaT4KICAgICAgICAgICAgPC9yZGY6U2VxPgogICAgICAgICA8L3htcE1NOkhpc3Rvcnk+CiAgICAgICAgIDx4bXBNTTpEZXJpdmVkRnJvbSByZGY6cGFyc2VUeXBlPSJSZXNvdXJjZSI+CiAgICAgICAgICAgIDxzdFJlZjppbnN0YW5jZUlEPnhtcC5paWQ6YmFhYTQxMTQtYzY3OS0zMzQzLWIyOTctY2U3N2NiMGE0ZTNhPC9zdFJlZjppbnN0YW5jZUlEPgogICAgICAgICAgICA8c3RSZWY6ZG9jdW1lbnRJRD54bXAuZGlkOjA0ODAxMTc0MDcyMDY4MTE4MDgzOUYyOUVDMjkwMDg4PC9zdFJlZjpkb2N1bWVudElEPgogICAgICAgICAgICA8c3RSZWY6b3JpZ2luYWxEb2N1bWVudElEPnhtcC5kaWQ6MDQ4MDExNzQwNzIwNjgxMTgwODM5RjI5RUMyOTAwODg8L3N0UmVmOm9yaWdpbmFsRG9jdW1lbnRJRD4KICAgICAgICAgPC94bXBNTTpEZXJpdmVkRnJvbT4KICAgICAgICAgPHRpZmY6T3JpZW50YXRpb24+MTwvdGlmZjpPcmllbnRhdGlvbj4KICAgICAgICAgPHRpZmY6WFJlc29sdXRpb24+MzAwMDAwMC8xMDAwMDwvdGlmZjpYUmVzb2x1dGlvbj4KICAgICAgICAgPHRpZmY6WVJlc29sdXRpb24+MzAwMDAwMC8xMDAwMDwvdGlmZjpZUmVzb2x1dGlvbj4KICAgICAgICAgPHRpZmY6UmVzb2x1dGlvblVuaXQ+MjwvdGlmZjpSZXNvbHV0aW9uVW5pdD4KICAgICAgICAgPGV4aWY6Q29sb3JTcGFjZT42NTUzNTwvZXhpZjpDb2xvclNwYWNlPgogICAgICAgICA8ZXhpZjpQaXhlbFhEaW1lbnNpb24+NTAwPC9leGlmOlBpeGVsWERpbWVuc2lvbj4KICAgICAgICAgPGV4aWY6UGl4ZWxZRGltZW5zaW9uPjE3OTwvZXhpZjpQaXhlbFlEaW1lbnNpb24+CiAgICAgIDwvcmRmOkRlc2NyaXB0aW9uPgogICA8L3JkZjpSREY+CjwveDp4bXBtZXRhPgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgIAo8P3hwYWNrZXQgZW5kPSJ3Ij8+aaFqSwAAACBjSFJNAAB6JQAAgIMAAPn/AACA6QAAdTAAAOpgAAA6mAAAF2+SX8VGAAAkpElEQVR42uydd5hVxfnHP1vYBXZZehdQEK/CTQADKhhQAbFGsBFjosYWo97EHn9GY2JBE2ONN8YaW7CiiF1jwS4oinoELk16kc6ysOwuu78/ZjYusOXec87t38/z3Idl9572zsz5zjsz7zs5NTU1CCGEECK9yZUJhBBCCAm6EEIIISToQgghhJCgCyGEEEKCLoQQQkjQhRBCCCFBF0IIIYQEXQghhBASdCGEEEKCLoQQQggJuhBCCCEk6EIIIYSQoAshhBASdCGEEEJI0IUQQgghQRdCCCGEBF0IIYTIXPIBcnJyZAkhPBAMB/KBbkBXoAvQFmgDtAZygAKgGVBmDykDNtnPamAVsNwJRUplTSGSQ01NTVrff05NTY0EXYjohbsZMMB+fgT0A/YGegJ5PlxiLbAAmAt8bT9fOKHIWllfCAm6BF0I9wJeDPwUOBQ4BBgEFCbhVuYB04B3gXecUGSRSkcICboEXYjGRbwncAJwLDAcM1yeaswFXgZeBD50QpEdKjkhJOgSdCERDwc6A6cCPwcOTLPbXw08C0x0QpFPVZpCSNCFyDYRzwWOAM4HjsIuEk1zZgEPAY84och6lbIQEnQhMlnIi4GzgIuA3hn6mNuAx4C7nFBktkpdCAm6EJkk5G2Ai4HfAe2y5R0FTAZudEKRL1ULhJCgC5HuHvnlVsxbZ7EpJgPXOKHILNUKISToQqSTkOcBZwPXA51lEQCqMXPs1zihyPcyhxASdCFSXcyHAvcAA2WNetkM/AW42wlFqmQOISToQqSakJcAtwLnyhpR8SVwrhOKzJAphARdgi5Eqoj50cADmJzqInp2AH8FrnNCkUqZQ0jQJehCJEvIWwC3YeLJhXu+Ak5xQpE5MoWQoEvQhUi0mO8HTMJskiK8UwaEnFDkEZlCSNAl6EIkSsxPAR4EimQN33nACnuFTCEk6BJ0IeIl5LnATcCVskZcmQ6Mc0KRlTKFkKBL0IXwW8xbAk8AY2WNhLAEOMYJRRyZQkjQU5tcFaFIIzHvCLwtMU8oPYGPguHASJlCCAm6EH6I+R7AR8BBskbCKQFeDYYD42QKISToQngR8z7AVKCvrJE0CoHnguHAL2UKIVITzaGLdPDMPwW6yxopQTVwuhOKTJQpRKahOXQh4ivmUyXmKffOeCwYDhwnUwgUMipBFyIKMW8NvAb0kTVS8r3xTDAcOEymyDp6YkJGvwS2AVuACmA+8C9gsEyUPDTkLlJRzAuB/wLDZY2UZjMwzAlFvpUpMp48zDbEVwDNmvjus8BvgfXp9pCKQxfCXzHPAR4HtPgqPVgCHOiEIqtkiowlF5gInBLDMQ5wGLBWgp7YghIilbhKYp5W9AQmB8OBApkiY7k4RjEHCAKPynQSdJEcmgE9gKHAycAlQJsEe+djgBtUFGnHQUBYZshIWgPXujz2aGC0TJg4NOSeHbSwYt3dfvawn9r/9wA6A7tWhL2ARQkS857AF0B7FVfacpoTivxHZsgofg087OH4J4FT0+VhNYcuUp0BwEyXxyZE0IPhQB4mPO2nKq60pgwYrP3UM4pHgdM9HL8G6CRBTwwacs980qG3dpXEPCMoAiYGw4FmMkXG4DUHREdMlkEhQReZTjAcGAj8WZbIGPYHrpEZMgY/OmfSGQm6yAIxbwY8AuTLGhnF1cFwYJDMkBEs93j8ZkwCGiFBFxnOFZg5fpFZ5AH327URIr35yOPx78mEEnSR+d55LzQ0m8kMBi6UGdKeyUC5h+O1iY8EXWQBd2HC6UTmcl0wHOgoM6Q1K3CfY+Bz4BmZMHG4nbvMwcQt98TEMPfExDXX/r8XJlH/9TKxqMc7PwQYK0tkPG0wiYJ+K1OkNX8CDsYknYqWtcDPgRqZL3FEE4c+HjPPWSvWPax4N5Xq8TrgLzJx0hmI2RnJDb7Hodtc7Z8BP1HRZAU7gP5OKBKRKdKaYuDfmCySTfGl1Y356faQ6R6HHo2H/ke0cEn4x1iJeVaRB9wYpRCI1GWLFelDgd8BhwOt6vy9EvgQuB8zzF4tkyUezaGLhBEMB3IxeymL7OIkm29ApD9TgROBdpjR2kHA3kAJMBJ4SmIuQRfZwfHAfjJDVvJ/MkFGUQUsw6SVXoC3lfBCgi7SkD/IBFnL+GA4EJAZhJCgizQnGA4MBw6QJbKWHOD3MoMQEnSR/uhlLs4IhgNtZQYhJOgifb3zbsA4WSLrKQJOkxmEkKCL9GUrcCkwQ6bIes6VCYSID9EklpmJuzh0JZaJr6fTE+iKCRHZBKzGpGnclYHEN7FMN1s/umOST9Tez2Jgtr2vut76aODv9r6SSl5uHl2Lu9GxqCOF+c1p2awlO6qr2Fq5lS0VW1hRupxN5ZvSu8eek8ceJd1p37IDRc2KICeHsootrNqyklWlq6hJfCKvL4BRTiiysYG/t8Zkmmxn63kzoAyzinolsASoSNPiKAH6YDLolQAbbfuYh0nA45eT1huTybMVJg/ABsxK9NUZ9h7sbt9R7YCWQHNr0wpgqX0HbY7lhNmQWEakBgWY4cozgGG2oe7KSuAOK5jxZB/raY3DxKA2xlzgVeBB4FsnFHkrGA4MAS7HJBxJ2I5c7Vq0Z3ivEQzpfgADugykR0lP8nIbv3xZZRmRNbOZuWom05d/yrRln1JVXZXSFaV32z6M2ftIhvcawb4d9qMgr/6kjqXbN/P5is95fd4rvLngjXg/lwP8wQlFXqvrUAAHAaOA4ZiEQ+2beudaUZ8GfAy8bMUqJR0mTAKWE4AxVnzqYysmKcvjmKQssXZYemMS9xyHiQtvaI+ExcAU4D5gVro5n5jUs2Mw8e4/sSLeFEuA6cC79tmXk8HUeuh3Ahcl8T4uBO6p5/ePWAGLlSnENmfbxvZi3XAYJtlCXXoCX8d4nmtoeBOEY+zf9oziPH9m5xz6fnrovYFbrW1zXJzvDczQ+yzrrR9qy6okfh5qLoftNYqT+/+coT2GkZvjbZZp8/bNvLngdSZ+9Tjz18/zfH8PH/94TN9/JfISk2bVv9/F/t0G89vBFzC0x7CY72PVlpXc9vHfeX3eq34XQQUmF/jtTihS22PoDISAX0VZp6Px+h/GpCbd6uE8d8Y4cjQLuKAB8fkFcC0Qa6jeQuBi4KUovhvA5Mo/IcaOcQ3wmG2L66P4/ln2WWJlFnC0x7JtY+vKmfb945X3gH9idpGrkoceH+ZkWEcpFzN0GKvXW9+L4Qbg6hjO822cescXAX8FCj2c5wjgK8xUzE1OKDI1GA6MAf7LzmkkfWFk79FcdNAl9G7bx7dzlhSWcFK/8ZzUbzzvLZrKbR/fwncbFro+3+BuQ2L6/qbyjbsJetvmbblqxDUc1fcY1/fRpbgrfx9zO8N7jeAv7/6Jyh2VfphrMTDWCUW+qr1VKwy/xQyP+sX+9vMX4G+YnfzcDMsPBA6J8fu7Cnov4NEYz7Ort/0icLNt9/UpTJ7929U0vadGQ+35DEwa12Pt6Emj1d4+V6xs9FCmzYErgUtcvEsb4xD7mWftN4kM2kAmVRbFzUXU56ncHaOYx6NzVIDZ0/gOj2JetxN5o21IBU4oMg2zK5Nv6SI7FXXmn8fey11HhX0V893eDHseyvOnvEjowIuaHLr3i77td3b4BnYZxKRTXvAk5nU5LjCO2464k9wcz88zHRhcR8zH27p5sc9iXpf2wC2Y0bGDElAcrTHpT2sZbUcLDvHh3FcBt9Xz+7bAK5g1SgUer9ELMxS9T4q9C4fbMvyLz2K+U1PCTG/8l4anQiToLtiCSSEoQd+ZSzFTEbFQ5XPnKM8K7y/i8Lwn2HPn2XlVX3K8D+42hGfGP8+IXocmpNDyc/M5b/D5PDT2ETq2jP/W33uU7EFhvtHDw/uM4d/jHqNTUWdfr3HYXqP43YGeZuCmA2OcUGQtZlHbP4CngU4JaksB4H0SM434o9q+EGatSDsfz30JO08ddrDPdYSP1+iAGX5uTmpwCWYKs2+CrjcK+MZ2OCXo8s7jIuiDMEOHsTIfs+uRX9wJ/CyOz/wzzNAimHn/r7ycbHTvw3lg7MO0b9k+4YX3k25DePSEiXQu7hLfBpuTS++2vTm67zHcesSdNMtrFpfrnLX/2QQ67Ovm0HnAMU4osgmzaOkFzO5ciaaZrb9/jfN1gpiFb5PsNf3mn5jV/sXWMw/G4Rr9MLtqJpMczDqq25OgS0W2w3miBN07sxBg5pA72J/DuFvfMNvnezo2Ac99BXCEE4pU2t65K47Y+yhuO/Iu8nOTtyykR+uePH7Ck3H31M8YeCYTRv/N8wK/xjsOeVx4QMw6vBUzZ74WMxz8Et4XRXnlyjiL+tg4ijmYsNCzMPPy8UydfClNRxjEk3uA85N4/a8xERMSdHnovnrp4zBhaW5w0vS57wGaO6HIu8BbMbsXHfszYdTNcRW4aOnaqit3Hh2mMK8wbtc4Zp+fJaTjcuheI+nWqlssh1zshCKz63RKR6ZI/boSd9Ey0TCMOEZpWG7DTFHF20s9PUnlcxlmoWSy2AqcAmyXoHtnNqKuoF/p4fh0jRboXadB3xLLgcUFrbjr6PD/5pVTgR93HsDFQy9L+8qYQw6j+0Q9Xfs+JtcAwNmkXka4+4B907QomiXoOsmYRz401jYfBy7OFB2Sh55aHIO31bnpPH1xJWYV/VuYtQBRcfnBf6BLcdeUe5hfDjiNgV0GpX2FjDKmvQb4vROK1GCyd92Wgo9SCDyE0l03xgEJGG2oS0tM/oBklskk4IFMKcBkx6HXSNB3wsvGFdUJ8NDLMSFsL9nOw1agi30RXAj093DuLsDxTijyVDAceBQTf9+kJ3xiv5M9PdC8dXN5dd7LzFkzm3Xb1tE8vzk9W/dkaI+DGdPnSNeLznLI4epDrmX80yckLL3qxvKNTJ79HJ8v/4z129bRtkU7hvcawfj+p7gOq+vfMaoifaZOeNqteAs1qrQv2Tdsx24bZnX8IEx44wAP5x6GSWbzWAKK4zXgScwiwbaYIV2/h7RX25GHjzDplgdgFrf1cnm+XExs/fsJet9dhz+JhTYDn2OipbbZTsle9lkaG7pbAvwmkwQkv86Qw8UNfGcmyuWeKLwE/y62ghsv3sAMoy7d5ffLMZuuPGBf5l5ChX4NPIUJo2lS0C8aeqnrC22vKueG967jxTkv7Ca4X678gilzXuCuT2/nptG3xJz8pZZ9O+zHYb1H8c7Ct+JecZ78ZiJ3fHIr2yq37fT7Dxa/x/Rln3LHUXe7Om/bFu0oalZEWWVZY1+rjcgYZIXLLR9Z0asvU8/rmIiIX1kRa+nyGtcCT1BPljCfWG875q/WI/CLcJdxrT6esmJUWud302zbmcHO8fGx0DdBgt4dkwHOC19be77SQHkWYlLFnocZ/azLDuBU3GcITUk0/JQ5xHNB3L22QSxt5DtVtlP4jIfrHA60dUKRb20HpUEGdd2fA7of6M4F3FHJeS+dy5Q5kxv1nleWruQ3L57FtGWfuH6gswadE9dCr66p5tp3ruam92/YTcxreWvhf5m66F3X1+hY1Gj4+CdOKFKbWthL6NN7mJjgptLu/ce+pN0uYOoDHB+n4tiEmRNuKH/ujfiTc+MhK0al9fxtjUdHKlHzV9fgLfb9AWAwJnV0Q52z7ZjRxGOBIbbDU8v1tgOZUUjQ04t1tqf/MGYB0nOYpAjVxG9Rx0RMestod4O6ELM7ltv6OMb+/E5jXzypv/v1O+HpdzFjxWdRi/8Vb17mete1AV0G0qfd3nGrEBPev57Js59r8ntvLXjD9TWaN77gsHb4ujPu97zfQGyrjD/y2HmI14K90217bLA6YeLyvTDNtsfG5nGew30605bEn1Z4m36YaL3uWHJufA4cbDs7U4EJmSgQEvT04CtMEpZOmLjes+xL6STgx5g5o3jssBbBDOvF8nJYazscbhlV6/k19IXigmKO6HOkq5Ov2rKSx796NDa12baeh7643/UDHRcYF5dK8bTzJM84T0VXkGsj8biFGisetWLmdk3OLcCqWPtlTY3iNMLoOHiiD2FysDfFlx6usR0zLdVUnvpNNL3tcTIF/RcerrMCE6/upsOyAzMVPBr/tquVoIuYeBCz6OxlGs53XmaF1G8uxN3uVf/xcM3a5BkNutDDehzsOkztGecpVxuPTJr1LNt3uBvlHd5rhO8FM3/9fG758Oaov79+2/p41M0ZTiiyxv481oNI3eviuArMXLobcoCjfLTDJqIPN/WyJ/ntRL/wNZW3CT3Vw7ETqH+qIVZhz0gk6KnNJOshVyTp+m73mf4M9zstBTFza3Ma6sAM7+V+74vX57/m6rjS7Zv5dKm7ufS+7ffxPd/6Pz69nYod0VeLbVXb4lE//mv/bYfZq9oNb3ioK17Wa/gp6A9ipsOioczDdWJJB11GatIaM/TthjISE6EgQRe+swI4h/Tc2q+anRegxEIe0NcJRcppYEh1SHd3GTCXb17G0k1LXD+Ul8VxA7v6G5O+pWJLKpTz5/bfn3p4l3gJAViA+2H3YX4WR4LsHctCjvIUfTeMwv3UzOsJtLUEXfjKn2NswKmGl/wCtXue7qa+JYUldC/Zw9VJne+/8fRAXo7fr0O/TKyjtYI+2MM5pvt0D7HSDeiISDRe8tG/JfNJ0NORNXibh04FFnk4tjYxxm4hPvt22M/1SRdvXOzpgbx49/t02CfT6mg5P4Qxehl+WODxPuZ7OPbHetUknJ94OHaGzCdBT0ceJHWHzKJloYdjayecv9/1D269c4Dlpd5CgNduXcv2KnfF0q1V90yro9/ZVK/gPttXKd4Xc37n4dieetUkHC/59LXvhwQ9LXkuA55ho4dja7dx3C2Lk5e87StLV3h+qBUuz+H3orgUoG6Ymdte1lIf7mORh2P3QCSSfMxUhxvWoflzCXoasozMGFryIuit63hwO9GhZQfXJy2r8L7wt9ylh15SWJIS27v6SG24WnOgjctz+LFGxMtIVgdEIuniQXNKZT4JejrydoY8h5dc2bVB5rvFZXnZJjWWMK8GOwWV7jsFRQXFmVRPt+1SVq6KxIf78NIpaI1IJO2SVM4SdJE0PsyQ5/BjeGy3pDYtmrVwf0M+hHp56RQ0y22WSfW0tmxaJfklXenh2HxEIinw810gJOjpwJcyQcO7zuXnuN+Qrrqm2rt67HCvH/m5+VlVVlHgR46F7R6OLVBTSyhehqgqZD4JejqSKWkJi3ywwW4vgFIPXnZRQZH3h/Jwjjhla0sWtbm4NybpBe/HOeT1JRYtapOgizSluQ8Nv3A3pa92PzXvh4dckOfeqXMb8paiFPrQAfVjyMLLZiIS9MTixcsukvkk6CJ5+DG3utsimo3lG9y7cj4sSmuR724Ov7yq3JdFeSlEbWhhqYcXdZskdxw3qJkmlHUejm0n80nQRfLwklZz+S6i8T9Wl7nfrKpDS++ZPjsWdXJ13KotKzOtfLvU+XmZD+dwi5eMPUvUTBPKatyP6HTH7JInJOgiCfTwcGyt+u2WyWv1FveCvkeJtzwixQXFtGnuzqlc4UNSmxSjlw+C3tmjhw2wl4djF6uZJpQqzKZTbij0+E6RoAvhgf08HBux/+656x/mr5/n+qTdSrylX/WSdnbu2kimlW9JMByoTcziJSVnb4/34eX4b9RME84sD8cOlPkk6CI5eNmEYVYwHGgG9N31D0s3LWFrpbu1TIH2+3p6IC8bw8xZm5FpqAfYf2f4cI5Ev+RX88PUjkgcXurKITKfd0GvlJlEjLTx8KJegAmF6kc9q6Cra6r59nvH3ZBBx34UNXO/WNbtPuwAM1dlZHqB2k7b5x7OMcLDse2AH7k89hM106Tgxe7j0Dy6Z0F3m+uyfRKfqzBW50tVwVeOw/3oT22mvAMb+sLHS90l08vPzWf4noe4PnZEL3fHfrdhIcs3L8vEcq4V45nW43XDzzzUFS/17HU106TwDu7z7/cGRsuE3gTdbTaMPkl8rliWMxcA/1JVqJe2Lo/7jccG36jn9uHiD1yf/KR+J7s6bmTv0bRt4S5y5r1FUzO1fhxqp0ZqPAhkd+Bol8ee6/K4GuAVNe+ksBV418PxN+MtO2ExGRzTHo2gu12VOAJo4fH+3OZ63heINnH2XWixRUOEiT095uHAwS6vVwW8HAwHchvric9ZO5uFGxa4usCBewxlcLchMXvnFx7wO9dGfDHyQqbWjyJglP35KQ/nudbFS/pYYJiHTuMyNe+k8YiHY38CXOfy2KGY0aS7slnQl3po7Bc38Z0B1LOSeZfenNtrj4niexOA36p9Ncgw4FGiz+jVBrjPw/WmAuuBgzAhTQ3y/KxJri9y/cgJlBSWRP393x90Cb3buhtw+mrVTOatm5vJdeRX9t83cR/XPQT4Ywzf74y3UbX71bSTygv8sP2uG64GborhvdQauAX4ADNyfDZwUrYK+rcezn+jbah1tylsacX2WcxGJI31trxk4/gbDW+P2AF4OsaXSLZyCmY4tWsT3+sEvIG3uOD76lyzUabMmex6K9MerXvywNiH6dREkpjcnFxCB/6eMwed7fqB/v3lg5leP04MhgPtgGqPIns9cDlNL3raC7PFsNsYwgXA82rWSaUCuN3jOa7CrJg/g/qnBvOAA4BbMfkGrmDnUaD78JaUKG0F/ROP559gva5lmDCRzfbFf5JtvKdhVjTXhxfXpj8wHTgL+DGwD2Y4+HZgPjBe7SpqRgHzMPNXwV1eul2BS2zH7wAP11gBTAmGA4XAL5v68sbyjTw282HXF+vXsT9TTn2FCw4I0bttH3LqPFLr5q05ep9jeerkSZw3+ALX15izdjbvLnw70+tGc+A8+/NdwCoP5/o78BFw8i6d8Vxb727GxI7393CNGzBTOyK53O2xrmDf649gUsouA77ALKqdhdkPYhpwWQOOXTvgMTIsdDuaIYtl9mXe16OwN9QbyrGif3w9f5vp8fn2AR5S2/GFIuD/7GcrZsisJd5SvNZlAiZE8hyizNv82MxHObHf+CY97YYoLmjF+UNCnD8kRMWOCjaWb6B5fouYhuMbooYabnr/Bmp82SE05bksGA7c7YQiW4A/423aZaj9gMm1vhUzolbow31+AjyuppwSlNkRmf/4cK4cqy+xetwj7T3ckk0eOsCkON/HOMw82q6swlsWKhEfWmJSf/ol5t8BDwbDgXzbwKJiS0UpE96/3pcbKMgroFNRZ1/EHODJryfy5covsqU+tAeusT8/gFkL4Qdt7UvaDzGvwMydVqv5pgwTgSlJvocbgEHZJugPJaAhTGjg98+p3mc859oX7plAIJYD31n4Fs9++3RKPcy33zvc/smt2VaGlwbDgf6YkLBf423RUzy4QM5BSnKO7dAniwLgCbxtw5t2gr4As4gsnhwOHNpAZ2JHAmzxGT8kNRGJ45/A28FwoD1mEWXM3PzBjXy+4rOUeJjvy1Zz0WuhTNv7PBqaAU8Ew4HmmEVIJ5I6WSbvRFNvqcpa4BjM2qpksS/eF+mllaCDWSVYGuf7uame3y3CW9xitJXq5CT3FFONRNjiI+BS+/PdmJXyMVO5o5LfvXIBX6/+Kulifubk01i9ZVW21pkf80NI2AeYha/J3gT+3jp1TKQmszF5S9Ym8R7OA47KJkFfjpmDiidDMQkjduWPxG8IbyNmccRitFlDXUZiRmbihYNZCFkRDAfOAX7h5WRbKko5+4UzeH/x1KQYa966uZz+/C9Zsinrt9g+LRgO/NX+/CJwJGZxWzK4ATPUXqPmnPJ8ZUV9TpKu/yzwcTYJeu1DXxzne6ov0cv3mDAzv4fwlmBCsmq3UVT2qB/ssAgzBRKPecf3bYdhTTAcGInJSOeZ8qpyQi+fzz8+vYMd1TsSZqwpcyZz6qTxmZqv3Q1XBsOBvwXDgRxMms+BmBCiRLEGGIvJQCcxTy9PfQgmnCxRrMPEso/HfWbStBV0MLGmp+B+05aGqE020JCnNhWzGn6LT9d7HRM3XXcp8kq1KeCH/ciXYdK4vuDTeauB2zBpXdcEw4Hh9tyFft14DTU8MOM+Tnx6LNOXx1dDFm9cxDlTzuSat6+iPPvmzJviD8DEYDjQ0nachwHnx9lbr8bMle9rRwdE+rHFCuwovCU1a4pK4B+2rjyWKcZzG1T/NLAfJn+z1x7wOvuS74tJAtDYPP2rwGDr4bllLvBzzHzJ6no89k0xfjIxSUVdr3wDcILtaHkZT/4As4Pa5UBlMBz4OSZdaKt4PMCC9fM5+4UzOOuF0/lg8Xu+xoPPXjOLP7x5Gcc9cTTTlmkXzkb4BfBFMBwYYsX2Xkyq5ytwn1K63sEZK+QBzKrp9TJ92vMOZk3GOPxdrLwBkz1ub+Aikjtv7zs5NTU15OR42mK2NyZM5TjM3sTRdBIimKG4Kfbf7S6ue4RtvMfQ9CYw621n4CngNeIfgpcLuA1oLsXfVf15LkSznPq3OCzALHQ6EziEpjfAWQ28hIlNng4QDAdKMIkczktkRe9U1InRfcZwyJ6HMaDLwJj2Ra+qrmLWmm/5ZOlHvDbvVRasn+/bfbVyGfe+tbIspmmFHHIoLnTXdyqrKKO6xlOVrMEslrvOCUVW1mkjB2MWo47EZIuM5UX0ve0kTgEm48/IXTHR5wePpr3UR769jhs2xvDdIqLfoKou29l9h81C3G20tQP/FlL3tI7YKDvaE0tl/g54D5Py983G9KamJr1naPwQ9J3eT1bU98SsWC62jbnUNsDvrPfnZ4hCPiYxQF97zdaY7FKlmHSiDmY+WAkl/KUVZuej/TCbZdTafbMt529sx63aCnkBZijteqBLUoelcnLp3bY3PdvsSfdW3Wnfsj2Fec0pLiimfEc55ZXlbN6+ieWly1m6aQnz1s3VkLp/nvRRTigytZ6/tbYe2V5AD0yymloh2WwFezWw0Nar+TJn1pKH2WRlb6s1HTApiFvaelJu3/0Lrd6sjrrnKUEXommC4cC9ifbKRcpRCnR3QpFSmUKkIuku6LkqQpEg/i0TqA5IzIWQoIs0xwlFpmMSyYgsdX4wyYOEEBJ0kQH8XSbIWp5xQpEFMoMQEnSRGbyI9y1xRXp659fLDEJI0EWG4IQiNZj9skX2eeezZAYhJOgis0T9RRKbBlQkl0rgTzKDEBJ0kZlcLhNkDf9yQpF5MoMQEnSRmV76h5isfSKz2YTmzoWQoIuM50p2TzEpMos/OqHIOplBCAm6yGwvfQlaIJfJTMNsxiKEkKCLLOAOYIbMkHFUAb9xQhHtnyCEBF1kiZdehdkxr0rWyChucEKRr2UGISToIrtEfSYaes8kpgM3yQxCSNBFdvJXYKrMkPaUAafZkRchhARdZKGXXg2cBqyXNdKa3zihyFyZQQgJushuUV8GjAe0kCo9uccJRZ6QGYSQoAuBE4q8jYlPF+nFx8AlMoMQEnQh6nIbIE8vfVgEHO+EIhUyhRDJJaempoacnBxZQqQMwXCgAHgDOFTWSGk2Awc5ochsmUJkAjU1NfLQhfAT6+2NA76VNVKWbcBYibkQEnQhmhL1TcCRwAJZI+WoBMY7ochUmUIICboQ0Yj6Msywu0Q9dagGznBCkZdlCiEk6ELEKupHAt/JGinhmZ/shCJPyhRCSNCFcCPq84ERgCNrJFXMT3JCkedlCiEk6EJ49dQPw2zLKRLLRmC0E4q8KFMIIUEXwg9RXwuMBCbJGgljMTDMCUXelymEkKAL4aeob8WkiL1Z1og7H6E4cyHSBiWWEWlLMBw4CXgYKJY1fCcMXOqEIpUyhcgW0j2xjARdpLuo9wWeB4Kyhi9sAS5wQpHHZQohQU8vNOQu0honFJkHHAD8Q9bwzDRgoMRcCHnoQiTbWx8FPALsIWvERCVwE3CjE4pUyRxCHroEXYhUEPUSYAJwARqBitYrP8cJRRTjLyToEnQhUlLYhwD3AvvLGvWyHvgTcK8TilTLHEJI0IVIZVHPBU6zHnt3WQSAHZgV7Nc5ocgGmUMICboQ6STsLYCLgMuB9tn6rgIeB26wqXSFEBJ0IdJW2IuBUJYJ+w7gGcyCt1mqBUJI0IXINI/9NOu198vQxywF7gfudkKRxSp1ISToQmSysOcAo4BzgOOBggx4rM+AB4CnnFCkVKUshARdiGwT97bAqcApwMFAOjWKhcCzwEQnFPlGpSmEBF0IYcS9G3Ac8DPMlq0tUvA2vwReB55zQpEZKjUhJOgSdCEaF/dC67GPBg7CpJktSvR7BpgFfAx8CLzhhCKrVTpCSNAl6EK4F/h8oD8wAPiR/XlvoBf+zMGvBhZYAXeAb4AZTiiySdYXQoIuQRci/kKfA3TD5JFvX+fT0gp9AWZefrs9pAzYiMnYtg5YASx2QpHtsqYQEnTXgi6EEEKI9EabVwghhBASdCGEEEJI0IUQQgghQRdCCCGEBF0IIYSQoAshhBBCgi6EEEIICboQQgghJOhCCCGEBF0IIYQQEnQhhBBCSNCFEEIIIUEXQgghJOhCCCGEkKALIYQQQoIuhBBCCAm6EEIIIUEXQgghRCrz/wMACn3Ca3R5CvgAAAAASUVORK5CYII=
              mediatype: image/png
          install:
            spec:
              permissions:
                - rules:
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                      resources:
                        - nodes
                        - pods
                        - configmaps
                        - endpoints
                        - events
                        - deployments
                        - persistentvolumeclaims
                        - replicasets
                        - replicationcontrollers
                        - services
                        - secrets
                        - serviceaccounts
                      verbs:
                        - '*'
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                        - policy
                      resources:
                        - daemonsets
                        - endpoints
                        - limitranges
                        - namespaces
                        - persistentvolumes
                        - persistentvolumeclaims
                        - poddisruptionbudget
                        - resourcequotas
                        - services
                        - statefulsets
                      verbs:
                        - get
                        - list
                        - watch
                    - apiGroups:
                        - ""
                      resources:
                        - nodes/spec
                        - nodes/stats
                      verbs:
                        - get
                    - apiGroups:
                        - charts.helm.k8s.io
                      resources:
                        - '*'
                      verbs:
                        - '*'
                  serviceAccountName: kubeturbo21824-operator
              clusterPermissions:
                - rules:
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                      resources:
                        - nodes
                        - pods
                        - configmaps
                        - deployments
                        - replicasets
                        - replicationcontrollers
                        - serviceaccounts
                      verbs:
                        - '*'
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                        - policy
                      resources:
                        - services
                        - endpoints
                        - namespaces
                        - limitranges
                        - resourcequotas
                        - daemonsets
                        - persistentvolumes
                        - persistentvolumeclaims
                        - poddisruptionbudget
                      verbs:
                        - get
                        - list
                        - watch
                    - apiGroups:
                        - ""
                      resources:
                        - nodes/spec
                        - nodes/stats
                      verbs:
                        - get
                    - apiGroups:
                        - charts.helm.k8s.io
                      resources:
                        - '*'
                      verbs:
                        - '*'
                    - apiGroups:
                        - rbac.authorization.k8s.io
                      resources:
                        - clusterroles
                        - clusterrolebindings
                      verbs:
                        - '*'
                  serviceAccountName: kubeturbo21824-operator
              deployments:
                - name: kubeturbo21824-operator
                  spec:
                    replicas: 1
                    selector:
                      matchLabels:
                        name: kubeturbo21824-operator
                    strategy: {}
                    template:
                      metadata:
                        labels:
                          name: kubeturbo21824-operator
                      spec:
                        containers:
                        - name: kubeturbo21824-operator
                          image: quay.io/olmqe/kubeturbo-operator-base:8.5-multi-arch
                          args:
                          - --leader-elect
                          - --leader-election-id=kubeturbo-operator
                          imagePullPolicy: Always
                          livenessProbe:
                            httpGet:
                              path: /healthz
                              port: 8081
                            initialDelaySeconds: 15
                            periodSeconds: 20
                          readinessProbe:
                            httpGet:
                              path: /readyz
                              port: 8081
                            initialDelaySeconds: 5
                            periodSeconds: 10
                            resources: {}
                          env:
                          - name: WATCH_NAMESPACE
                            valueFrom:
                              fieldRef:
                                fieldPath: metadata.namespace
                          - name: POD_NAME
                            valueFrom:
                              fieldRef:
                                fieldPath: metadata.name
                          - name: OPERATOR_NAME
                            value: "kubeturbo21824-operator"
                          securityContext:
                            readOnlyRootFilesystem: true
                            capabilities:
                              drop:
                                - ALL
                          volumeMounts:
                          - mountPath: /tmp
                            name: operator-tmpfs0
                        volumes:
                        - name: operator-tmpfs0
                          emptyDir: {}
                        serviceAccountName: kubeturbo21824-operator
            strategy: deployment
          installModes:
            - supported: true
              type: OwnNamespace
            - supported: true
              type: SingleNamespace
            - supported: false
              type: MultiNamespace
            - supported: false
              type: AllNamespaces
          links:
            - name: Turbonomic, Inc.
              url: https://www.turbonomic.com/
            - name: Kubeturbo21824 Operator
              url: https://github.com/turbonomic/kubeturbo21824/tree/master/deploy/kubeturbo21824-operator
          maintainers:
            - email: endre.sara@turbonomic.com
              name: Endre Sara
          maturity: alpha
          provider:
            name: Turbonomic, Inc.
          version: 8.5.0
    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: kubeturbo21824s.charts.helm.k8s.io
          annotations:
            "api-approved.kubernetes.io": "https://github.com/operator-framework/operator-sdk/pull/2703"
        spec:
          group: charts.helm.k8s.io
          names:
            kind: Kubeturbo21824
            listKind: Kubeturbo21824List
            plural: kubeturbo21824s
            singular: kubeturbo21824
          scope: Namespaced
          versions:
            # Each version can be enabled/disabled by Served flag.
            # One and only one version must be marked as the storage version.
            - name: v1alpha1
              served: true
              storage: false
              schema:
                openAPIV3Schema:
                  description: Kubeturbo21824 is the Schema for the kubeturbo21824s API
                  type: object
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this representation
                    of an object. Servers should convert recognized schemas to the latest
                    internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    kind:
                      description: 'Kind is a string value representing the REST resource this
                    object represents. Servers may infer this from the endpoint the client
                    submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    metadata:
                      type: object
                    spec:
                      x-kubernetes-preserve-unknown-fields: true
                      properties:
                      type: object
            - name: v1
              served: true
              storage: true
              schema:
                openAPIV3Schema:
                  description: Kubeturbo21824 is the Schema for the kubeturbo21824s API
                  type: object
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this representation
                      of an object. Servers should convert recognized schemas to the latest
                      internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    kind:
                      description: 'Kind is a string value representing the REST resource this
                      object represents. Servers may infer this from the endpoint the client
                      submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    metadata:
                      type: object
                    spec:
                      x-kubernetes-preserve-unknown-fields: true
                      description: Spec defines the desired state of Kubeturbo21824
                      type: object
                      properties:
                        roleBinding:
                          description: The name of cluster role binding. Default is turbo-all-binding
                          type: string
                        serviceAccountName:
                          description: The name of the service account name. Default is turbo-user
                          type: string
                        replicaCount:
                          description: Kubeturbo21824 replicaCount
                          type: integer
                        image:
                          description: Kubeturbo21824 image details for deployments outside of RH Operator Hub
                          type: object
                          properties:
                            repository:
                              description: Container repository. default is docker hub
                              type: string
                            tag:
                              description: Kubeturbo21824 container image tag
                              type: string
                            busyboxRepository:
                              description: Busybox repository. default is busybox
                              type: string
                            pullPolicy:
                              description: Define pull policy, Always is default
                              type: string
                            imagePullSecret:
                              description: Define the secret used to authenticate to the container image registry
                              type: string
                        serverMeta:
                          description: Configuration for Turbo Server
                          type: object
                          properties:
                            version:
                              description: Turbo Server major version
                              type: string
                            turboServer:
                              description: URL for Turbo Server endpoint
                              type: string
                        restAPIConfig:
                          description: Credentials to register probe with Turbo Server
                          type: object
                          properties:
                            turbonomicCredentialsSecretName:
                              description: Name of k8s secret that contains the turbo credentials
                              type: string
                            opsManagerUserName:
                              description: Turbo admin user id
                              type: string
                            opsManagerPassword:
                              description: Turbo admin user password
                              type: string
                        featureGates:
                          description: Disable features
                          type: object
                          properties:
                            disabledFeatures:
                              description: Feature names
                              type: string
                        HANodeConfig:
                          description: Create HA placement policy for Node to Hypervisor by node role. Master is default
                          type: object
                          properties:
                            nodeRoles:
                              description: Node role names
                              type: string
                        targetConfig:
                          description: Optional target configuration
                          type: object
                          properties:
                            targetName:
                              description: Optional target name for registration
                              type: string
                        args:
                          description: Kubeturbo21824 command line arguments
                          type: object
                          properties:
                            logginglevel:
                              description: Define logging level, default is info = 2
                              type: integer
                            kubelethttps:
                              description: Identify if kubelet requires https
                              type: boolean
                            kubeletport:
                              description: Identify kubelet port
                              type: integer
                            sccsupport:
                              description: Allow kubeturbo21824 to execute actions in OCP
                              type: string
                            failVolumePodMoves:
                              description: Allow kubeturbo21824 to reschedule pods with volumes attached
                              type: string
                            busyboxExcludeNodeLabels:
                              description: Do not run busybox on these nodes to discover the cpu frequency with k8s 1.18 and later, default is either of kubernetes.io/os=windows or beta.kubernetes.io/os=windows present as node label
                              type: string
                            stitchuuid:
                              description: Identify if using uuid or ip for stitching
                              type: boolean
                        resources:
                          description: Kubeturbo21824 resource configuration
                          type: object
                          properties:
                            limits:
                              description: Define limits
                              type: object
                              properties:
                                memory:
                                  description: Define memory limits in Gi or Mi, include units
                                  type: string
                                cpu:
                                  description: Define cpu limits in cores or millicores, include units
                                  type: string
                            requests:
                              description: Define requests
                              type: object
                              properties:
                                memory:
                                  description: Define memory requests in Gi or Mi, include units
                                  type: string
                                cpu:
                                  description: Define cpu requests in cores or millicores, include units
                                  type: string
    packages: |
      - channels:
        - currentCSV: kubeturbo21824-operator.v8.5.0
          name: alpha
        defaultChannel: alpha
        packageName: kubeturbo21824
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
parameters:
- name: NAME
- name: NAMESPACE

`)

func testQeTestdataOlmCm21824CorrectYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCm21824CorrectYaml, nil
}

func testQeTestdataOlmCm21824CorrectYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCm21824CorrectYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cm-21824-correct.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCm21824WrongYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-21824-template
objects:
- apiVersion: v1
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: '[{"apiVersion":"charts.helm.k8s.io/v1","kind":"Kubeturbo21824","metadata":{"name":"kubeturbo21824-release"},"spec":{"serverMeta":{"turboServer":"https://Turbo_server_URL"},"targetConfig":{"targetName":"Cluster_Name"}}}]'
            capabilities: Basic Install
            categories: Monitoring
            certified: "false"
            containerImage: quay.io/olmqe/kubeturbo-operator-base:8.5-multi-arch
            createdAt: 2019-05-01T00:00:00.000Z
            description: Turbonomic Workload Automation for Multicloud simultaneously optimizes performance, compliance, and cost in real-time. Workloads are precisely resourced, automatically, to perform while satisfying business constraints.
            repository: https://github.com/turbonomic/kubeturbo21824/tree/master/deploy/kubeturbo21824-operator
            support: Turbonomic, Inc.
          labels:
            operatorframework.io/arch.amd64: supported
            operatorframework.io/arch.arm64: supported
            operatorframework.io/arch.ppc64le: supported
            operatorframework.io/arch.s390x: supported
          name: kubeturbo21824-operator.v8.5.0
          namespace: placeholder
        spec:
          apiservicedefinitions: {}
          customresourcedefinitions:
            owned:
            - description: Turbonomic Workload Automation for Multicloud simultaneously optimizes performance, compliance, and cost in real-time. Workloads are precisely resourced, automatically, to perform while satisfying business constraints.
              displayName: Kubeturbo21824 Operator
              kind: Kubeturbo21824
              name: kubeturbo21824s.charts.helm.k8s.io
              version: v1
          description: |-
            ### Application Resource Management for Kubernetes
            Turbonomic AI-powered Application Resource Management simultaneously optimizes performance, compliance, and cost in real time.
            Software manages the complete application stack, automatically. Applications are continually resourced to perform while satisfying business constraints.

            Turbonomic makes workloads smart — enabling them to self-manage and determines the specific actions that will drive continuous health:

            * Continuous placement for Pods (rescheduling)
            * Continuous scaling for applications and  the underlying cluster.

            It assures application performance by giving workloads the resources they need when they need them.

            ### How does it work?
            Turbonomic uses a container — KubeTurbo — that runs in your Kubernetes or Red Hat OpenShift cluster to discover and monitor your environment.
            KubeTurbo runs together with the default scheduler and sends this data back to the Turbonomic analytics engine.
            Turbonomic determines the right actions that drive continuous health, including continuous placement for Pods and continuous scaling for applications and the underlying cluster.
          displayName: Kubeturbo21824 Operator
          icon:
            - base64data: iVBORw0KGgoAAAANSUhEUgAAAfQAAACzCAYAAAB//O7qAAAACXBIWXMAAC4jAAAuIwF4pT92AABD52lUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4KPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxMzIgNzkuMTU5Mjg0LCAyMDE2LzA0LzE5LTEzOjEzOjQwICAgICAgICAiPgogICA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPgogICAgICA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIgogICAgICAgICAgICB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iCiAgICAgICAgICAgIHhtbG5zOnBob3Rvc2hvcD0iaHR0cDovL25zLmFkb2JlLmNvbS9waG90b3Nob3AvMS4wLyIKICAgICAgICAgICAgeG1sbnM6ZGM9Imh0dHA6Ly9wdXJsLm9yZy9kYy9lbGVtZW50cy8xLjEvIgogICAgICAgICAgICB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIKICAgICAgICAgICAgeG1sbnM6c3RFdnQ9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZUV2ZW50IyIKICAgICAgICAgICAgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiCiAgICAgICAgICAgIHhtbG5zOnRpZmY9Imh0dHA6Ly9ucy5hZG9iZS5jb20vdGlmZi8xLjAvIgogICAgICAgICAgICB4bWxuczpleGlmPSJodHRwOi8vbnMuYWRvYmUuY29tL2V4aWYvMS4wLyI+CiAgICAgICAgIDx4bXA6Q3JlYXRvclRvb2w+QWRvYmUgUGhvdG9zaG9wIENDIDIwMTUuNSAoTWFjaW50b3NoKTwveG1wOkNyZWF0b3JUb29sPgogICAgICAgICA8eG1wOkNyZWF0ZURhdGU+MjAxMy0xMi0wNFQxNToxMTozNC0wNTowMDwveG1wOkNyZWF0ZURhdGU+CiAgICAgICAgIDx4bXA6TWV0YWRhdGFEYXRlPjIwMTYtMDgtMTFUMTQ6MzI6NTUrMDM6MDA8L3htcDpNZXRhZGF0YURhdGU+CiAgICAgICAgIDx4bXA6TW9kaWZ5RGF0ZT4yMDE2LTA4LTExVDE0OjMyOjU1KzAzOjAwPC94bXA6TW9kaWZ5RGF0ZT4KICAgICAgICAgPHBob3Rvc2hvcDpDb2xvck1vZGU+MzwvcGhvdG9zaG9wOkNvbG9yTW9kZT4KICAgICAgICAgPHBob3Rvc2hvcDpEb2N1bWVudEFuY2VzdG9ycz4KICAgICAgICAgICAgPHJkZjpCYWc+CiAgICAgICAgICAgICAgIDxyZGY6bGk+QTkxQjdEQ0MwNEMwQzdBOERGRDVDMTVGNDgwMzY3Njc8L3JkZjpsaT4KICAgICAgICAgICAgICAgPHJkZjpsaT51dWlkOjcxMzQxOTYxLTU5ODctZTE0Ny1iZjA3LTA2MmE5OTNiM2I3YTwvcmRmOmxpPgogICAgICAgICAgICAgICA8cmRmOmxpPnhtcC5kaWQ6MDQ4MDExNzQwNzIwNjgxMTgwODM5RjI5RUMyOTAwODg8L3JkZjpsaT4KICAgICAgICAgICAgPC9yZGY6QmFnPgogICAgICAgICA8L3Bob3Rvc2hvcDpEb2N1bWVudEFuY2VzdG9ycz4KICAgICAgICAgPGRjOmZvcm1hdD5pbWFnZS9wbmc8L2RjOmZvcm1hdD4KICAgICAgICAgPHhtcE1NOkluc3RhbmNlSUQ+eG1wLmlpZDo1MzI1MGY1Ni05MzRhLTQ1N2MtYTEwMS0zZjY0MmNiZmQxOTY8L3htcE1NOkluc3RhbmNlSUQ+CiAgICAgICAgIDx4bXBNTTpEb2N1bWVudElEPnhtcC5kaWQ6MDQ4MDExNzQwNzIwNjgxMTgwODM5RjI5RUMyOTAwODg8L3htcE1NOkRvY3VtZW50SUQ+CiAgICAgICAgIDx4bXBNTTpPcmlnaW5hbERvY3VtZW50SUQ+eG1wLmRpZDowNDgwMTE3NDA3MjA2ODExODA4MzlGMjlFQzI5MDA4ODwveG1wTU06T3JpZ2luYWxEb2N1bWVudElEPgogICAgICAgICA8eG1wTU06SGlzdG9yeT4KICAgICAgICAgICAgPHJkZjpTZXE+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPmNyZWF0ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0Omluc3RhbmNlSUQ+eG1wLmlpZDowNDgwMTE3NDA3MjA2ODExODA4MzlGMjlFQzI5MDA4ODwvc3RFdnQ6aW5zdGFuY2VJRD4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OndoZW4+MjAxMy0xMi0wNFQxNToxMTozNC0wNTowMDwvc3RFdnQ6d2hlbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnNvZnR3YXJlQWdlbnQ+QWRvYmUgUGhvdG9zaG9wIENTNiAoTWFjaW50b3NoKTwvc3RFdnQ6c29mdHdhcmVBZ2VudD4KICAgICAgICAgICAgICAgPC9yZGY6bGk+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPnNhdmVkPC9zdEV2dDphY3Rpb24+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDppbnN0YW5jZUlEPnhtcC5paWQ6OTZDNDQxRTcwQjZDRTMxMTg3Q0ZCQjM3Mzg4MzY1MTA8L3N0RXZ0Omluc3RhbmNlSUQ+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDp3aGVuPjIwMTMtMTItMjNUMTY6NTM6NTktMDU6MDA8L3N0RXZ0OndoZW4+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDpzb2Z0d2FyZUFnZW50PkFkb2JlIFBob3Rvc2hvcCBDUzYgKFdpbmRvd3MpPC9zdEV2dDpzb2Z0d2FyZUFnZW50PgogICAgICAgICAgICAgICAgICA8c3RFdnQ6Y2hhbmdlZD4vPC9zdEV2dDpjaGFuZ2VkPgogICAgICAgICAgICAgICA8L3JkZjpsaT4KICAgICAgICAgICAgICAgPHJkZjpsaSByZGY6cGFyc2VUeXBlPSJSZXNvdXJjZSI+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDphY3Rpb24+c2F2ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0Omluc3RhbmNlSUQ+eG1wLmlpZDpiYWFhNDExNC1jNjc5LTMzNDMtYjI5Ny1jZTc3Y2IwYTRlM2E8L3N0RXZ0Omluc3RhbmNlSUQ+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDp3aGVuPjIwMTQtMDQtMDJUMTQ6Mjk6MzEtMDQ6MDA8L3N0RXZ0OndoZW4+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDpzb2Z0d2FyZUFnZW50PkFkb2JlIFBob3Rvc2hvcCBDQyAoV2luZG93cyk8L3N0RXZ0OnNvZnR3YXJlQWdlbnQ+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDpjaGFuZ2VkPi88L3N0RXZ0OmNoYW5nZWQ+CiAgICAgICAgICAgICAgIDwvcmRmOmxpPgogICAgICAgICAgICAgICA8cmRmOmxpIHJkZjpwYXJzZVR5cGU9IlJlc291cmNlIj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OmFjdGlvbj5jb252ZXJ0ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnBhcmFtZXRlcnM+ZnJvbSBhcHBsaWNhdGlvbi92bmQuYWRvYmUucGhvdG9zaG9wIHRvIGltYWdlL3BuZzwvc3RFdnQ6cGFyYW1ldGVycz4KICAgICAgICAgICAgICAgPC9yZGY6bGk+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPmRlcml2ZWQ8L3N0RXZ0OmFjdGlvbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnBhcmFtZXRlcnM+Y29udmVydGVkIGZyb20gYXBwbGljYXRpb24vdm5kLmFkb2JlLnBob3Rvc2hvcCB0byBpbWFnZS9wbmc8L3N0RXZ0OnBhcmFtZXRlcnM+CiAgICAgICAgICAgICAgIDwvcmRmOmxpPgogICAgICAgICAgICAgICA8cmRmOmxpIHJkZjpwYXJzZVR5cGU9IlJlc291cmNlIj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OmFjdGlvbj5zYXZlZDwvc3RFdnQ6YWN0aW9uPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6aW5zdGFuY2VJRD54bXAuaWlkOjM4ZjZlNDQ0LTFiZWMtYWQ0Zi1hZDUzLTQ3ODVjOTlhZjk4Mjwvc3RFdnQ6aW5zdGFuY2VJRD4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OndoZW4+MjAxNC0wNC0wMlQxNDoyOTozMS0wNDowMDwvc3RFdnQ6d2hlbj4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OnNvZnR3YXJlQWdlbnQ+QWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKTwvc3RFdnQ6c29mdHdhcmVBZ2VudD4KICAgICAgICAgICAgICAgICAgPHN0RXZ0OmNoYW5nZWQ+Lzwvc3RFdnQ6Y2hhbmdlZD4KICAgICAgICAgICAgICAgPC9yZGY6bGk+CiAgICAgICAgICAgICAgIDxyZGY6bGkgcmRmOnBhcnNlVHlwZT0iUmVzb3VyY2UiPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6YWN0aW9uPnNhdmVkPC9zdEV2dDphY3Rpb24+CiAgICAgICAgICAgICAgICAgIDxzdEV2dDppbnN0YW5jZUlEPnhtcC5paWQ6NTMyNTBmNTYtOTM0YS00NTdjLWExMDEtM2Y2NDJjYmZkMTk2PC9zdEV2dDppbnN0YW5jZUlEPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6d2hlbj4yMDE2LTA4LTExVDE0OjMyOjU1KzAzOjAwPC9zdEV2dDp3aGVuPgogICAgICAgICAgICAgICAgICA8c3RFdnQ6c29mdHdhcmVBZ2VudD5BZG9iZSBQaG90b3Nob3AgQ0MgMjAxNS41IChNYWNpbnRvc2gpPC9zdEV2dDpzb2Z0d2FyZUFnZW50PgogICAgICAgICAgICAgICAgICA8c3RFdnQ6Y2hhbmdlZD4vPC9zdEV2dDpjaGFuZ2VkPgogICAgICAgICAgICAgICA8L3JkZjpsaT4KICAgICAgICAgICAgPC9yZGY6U2VxPgogICAgICAgICA8L3htcE1NOkhpc3Rvcnk+CiAgICAgICAgIDx4bXBNTTpEZXJpdmVkRnJvbSByZGY6cGFyc2VUeXBlPSJSZXNvdXJjZSI+CiAgICAgICAgICAgIDxzdFJlZjppbnN0YW5jZUlEPnhtcC5paWQ6YmFhYTQxMTQtYzY3OS0zMzQzLWIyOTctY2U3N2NiMGE0ZTNhPC9zdFJlZjppbnN0YW5jZUlEPgogICAgICAgICAgICA8c3RSZWY6ZG9jdW1lbnRJRD54bXAuZGlkOjA0ODAxMTc0MDcyMDY4MTE4MDgzOUYyOUVDMjkwMDg4PC9zdFJlZjpkb2N1bWVudElEPgogICAgICAgICAgICA8c3RSZWY6b3JpZ2luYWxEb2N1bWVudElEPnhtcC5kaWQ6MDQ4MDExNzQwNzIwNjgxMTgwODM5RjI5RUMyOTAwODg8L3N0UmVmOm9yaWdpbmFsRG9jdW1lbnRJRD4KICAgICAgICAgPC94bXBNTTpEZXJpdmVkRnJvbT4KICAgICAgICAgPHRpZmY6T3JpZW50YXRpb24+MTwvdGlmZjpPcmllbnRhdGlvbj4KICAgICAgICAgPHRpZmY6WFJlc29sdXRpb24+MzAwMDAwMC8xMDAwMDwvdGlmZjpYUmVzb2x1dGlvbj4KICAgICAgICAgPHRpZmY6WVJlc29sdXRpb24+MzAwMDAwMC8xMDAwMDwvdGlmZjpZUmVzb2x1dGlvbj4KICAgICAgICAgPHRpZmY6UmVzb2x1dGlvblVuaXQ+MjwvdGlmZjpSZXNvbHV0aW9uVW5pdD4KICAgICAgICAgPGV4aWY6Q29sb3JTcGFjZT42NTUzNTwvZXhpZjpDb2xvclNwYWNlPgogICAgICAgICA8ZXhpZjpQaXhlbFhEaW1lbnNpb24+NTAwPC9leGlmOlBpeGVsWERpbWVuc2lvbj4KICAgICAgICAgPGV4aWY6UGl4ZWxZRGltZW5zaW9uPjE3OTwvZXhpZjpQaXhlbFlEaW1lbnNpb24+CiAgICAgIDwvcmRmOkRlc2NyaXB0aW9uPgogICA8L3JkZjpSREY+CjwveDp4bXBtZXRhPgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgIAo8P3hwYWNrZXQgZW5kPSJ3Ij8+aaFqSwAAACBjSFJNAAB6JQAAgIMAAPn/AACA6QAAdTAAAOpgAAA6mAAAF2+SX8VGAAAkpElEQVR42uydd5hVxfnHP1vYBXZZehdQEK/CTQADKhhQAbFGsBFjosYWo97EHn9GY2JBE2ONN8YaW7CiiF1jwS4oinoELk16kc6ysOwuu78/ZjYusOXec87t38/z3Idl9572zsz5zjsz7zs5NTU1CCGEECK9yZUJhBBCCAm6EEIIISToQgghhJCgCyGEEEKCLoQQQkjQhRBCCCFBF0IIIYQEXQghhBASdCGEEEKCLoQQQggJuhBCCCEk6EIIIYSQoAshhBASdCGEEEJI0IUQQgghQRdCCCGEBF0IIYTIXPIBcnJyZAkhPBAMB/KBbkBXoAvQFmgDtAZygAKgGVBmDykDNtnPamAVsNwJRUplTSGSQ01NTVrff05NTY0EXYjohbsZMMB+fgT0A/YGegJ5PlxiLbAAmAt8bT9fOKHIWllfCAm6BF0I9wJeDPwUOBQ4BBgEFCbhVuYB04B3gXecUGSRSkcICboEXYjGRbwncAJwLDAcM1yeaswFXgZeBD50QpEdKjkhJOgSdCERDwc6A6cCPwcOTLPbXw08C0x0QpFPVZpCSNCFyDYRzwWOAM4HjsIuEk1zZgEPAY84och6lbIQEnQhMlnIi4GzgIuA3hn6mNuAx4C7nFBktkpdCAm6EJkk5G2Ai4HfAe2y5R0FTAZudEKRL1UL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
              mediatype: image/png
          install:
            spec:
              permissions:
                - rules:
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                      resources:
                        - nodes
                        - pods
                        - configmaps
                        - endpoints
                        - events
                        - deployments
                        - persistentvolumeclaims
                        - replicasets
                        - replicationcontrollers
                        - services
                        - secrets
                        - serviceaccounts
                      verbs:
                        - '*'
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                        - policy
                      resources:
                        - daemonsets
                        - endpoints
                        - limitranges
                        - namespaces
                        - persistentvolumes
                        - persistentvolumeclaims
                        - poddisruptionbudget
                        - resourcequotas
                        - services
                        - statefulsets
                      verbs:
                        - get
                        - list
                        - watch
                    - apiGroups:
                        - ""
                      resources:
                        - nodes/spec
                        - nodes/stats
                      verbs:
                        - get
                    - apiGroups:
                        - charts.helm.k8s.io
                      resources:
                        - '*'
                      verbs:
                        - '*'
                  serviceAccountName: kubeturbo21824-operator
              clusterPermissions:
                - rules:
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                      resources:
                        - nodes
                        - pods
                        - configmaps
                        - deployments
                        - replicasets
                        - replicationcontrollers
                        - serviceaccounts
                      verbs:
                        - '*'
                    - apiGroups:
                        - ""
                        - apps
                        - extensions
                        - policy
                      resources:
                        - services
                        - endpoints
                        - namespaces
                        - limitranges
                        - resourcequotas
                        - daemonsets
                        - persistentvolumes
                        - persistentvolumeclaims
                        - poddisruptionbudget
                      verbs:
                        - get
                        - list
                        - watch
                    - apiGroups:
                        - ""
                      resources:
                        - nodes/spec
                        - nodes/stats
                      verbs:
                        - get
                    - apiGroups:
                        - charts.helm.k8s.io
                      resources:
                        - '*'
                      verbs:
                        - '*'
                    - apiGroups:
                        - rbac.authorization.k8s.io
                      resources:
                        - clusterroles
                        - clusterrolebindings
                      verbs:
                        - '*'
                  serviceAccountName: kubeturbo21824-operator
              deployments:
                - name: kubeturbo21824-operator
                  spec:
                    replicas: 1
                    selector:
                      matchLabels:
                        name: kubeturbo21824-operator
                    strategy: {}
                    template:
                      metadata:
                        labels:
                          name: kubeturbo21824-operator
                      spec:
                        containers:
                        - name: kubeturbo21824-operator
                          image: quay.io/olmqe/kubeturbo-operator-base:8.5-multi-arch
                          args:
                          - --leader-elect
                          - --leader-election-id=kubeturbo-operator
                          imagePullPolicy: Always
                          livenessProbe:
                            httpGet:
                              path: /healthz
                              port: 8081
                            initialDelaySeconds: 15
                            periodSeconds: 20
                          readinessProbe:
                            httpGet:
                              path: /readyz
                              port: 8081
                            initialDelaySeconds: 5
                            periodSeconds: 10
                            resources: {}
                          env:
                          - name: WATCH_NAMESPACE
                            valueFrom:
                              fieldRef:
                                fieldPath: metadata.namespace
                          - name: POD_NAME
                            valueFrom:
                              fieldRef:
                                fieldPath: metadata.name
                          - name: OPERATOR_NAME
                            value: "kubeturbo21824-operator"
                          securityContext:
                            readOnlyRootFilesystem: true
                            capabilities:
                              drop:
                                - ALL
                          volumeMounts:
                          - mountPath: /tmp
                            name: operator-tmpfs0
                        volumes:
                        - name: operator-tmpfs0
                          emptyDir: {}
                        serviceAccountName: kubeturbo21824-operator
            strategy: deployment
          installModes:
            - supported: true
              type: OwnNamespace
            - supported: true
              type: SingleNamespace
            - supported: false
              type: MultiNamespace
            - supported: false
              type: AllNamespaces
          links:
            - name: Turbonomic, Inc.
              url: https://www.turbonomic.com/
            - name: Kubeturbo21824 Operator
              url: https://github.com/turbonomic/kubeturbo21824/tree/master/deploy/kubeturbo21824-operator
          maintainers:
            - email: endre.sara@turbonomic.com
              name: Endre Sara
          maturity: alpha
          provider:
            name: Turbonomic, Inc.
          version: 8.5.0
    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: kubeturbo21824.charts.helm.k8s.io
          annotations:
            "api-approved.kubernetes.io": "https://github.com/operator-framework/operator-sdk/pull/2703"
        spec:
          group: charts.helm.k8s.io
          names:
            kind: Kubeturbo21824
            listKind: Kubeturbo21824List
            plural: kubeturbo21824s
            singular: kubeturbo21824
          scope: Namespaced
          versions:
            # Each version can be enabled/disabled by Served flag.
            # One and only one version must be marked as the storage version.
            - name: v1alpha1
              served: true
              storage: false
              schema:
                openAPIV3Schema:
                  description: Kubeturbo21824 is the Schema for the kubeturbo21824s API
                  type: object
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this representation
                    of an object. Servers should convert recognized schemas to the latest
                    internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    kind:
                      description: 'Kind is a string value representing the REST resource this
                    object represents. Servers may infer this from the endpoint the client
                    submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    metadata:
                      type: object
                    spec:
                      x-kubernetes-preserve-unknown-fields: true
                      properties:
                      type: object
            - name: v1
              served: true
              storage: true
              schema:
                openAPIV3Schema:
                  description: Kubeturbo21824 is the Schema for the kubeturbo21824s API
                  type: object
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this representation
                      of an object. Servers should convert recognized schemas to the latest
                      internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    kind:
                      description: 'Kind is a string value representing the REST resource this
                      object represents. Servers may infer this from the endpoint the client
                      submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    metadata:
                      type: object
                    spec:
                      x-kubernetes-preserve-unknown-fields: true
                      description: Spec defines the desired state of Kubeturbo21824
                      type: object
                      properties:
                        roleBinding:
                          description: The name of cluster role binding. Default is turbo-all-binding
                          type: string
                        serviceAccountName:
                          description: The name of the service account name. Default is turbo-user
                          type: string
                        replicaCount:
                          description: Kubeturbo21824 replicaCount
                          type: integer
                        image:
                          description: Kubeturbo21824 image details for deployments outside of RH Operator Hub
                          type: object
                          properties:
                            repository:
                              description: Container repository. default is docker hub
                              type: string
                            tag:
                              description: Kubeturbo21824 container image tag
                              type: string
                            busyboxRepository:
                              description: Busybox repository. default is busybox
                              type: string
                            pullPolicy:
                              description: Define pull policy, Always is default
                              type: string
                            imagePullSecret:
                              description: Define the secret used to authenticate to the container image registry
                              type: string
                        serverMeta:
                          description: Configuration for Turbo Server
                          type: object
                          properties:
                            version:
                              description: Turbo Server major version
                              type: string
                            turboServer:
                              description: URL for Turbo Server endpoint
                              type: string
                        restAPIConfig:
                          description: Credentials to register probe with Turbo Server
                          type: object
                          properties:
                            turbonomicCredentialsSecretName:
                              description: Name of k8s secret that contains the turbo credentials
                              type: string
                            opsManagerUserName:
                              description: Turbo admin user id
                              type: string
                            opsManagerPassword:
                              description: Turbo admin user password
                              type: string
                        featureGates:
                          description: Disable features
                          type: object
                          properties:
                            disabledFeatures:
                              description: Feature names
                              type: string
                        HANodeConfig:
                          description: Create HA placement policy for Node to Hypervisor by node role. Master is default
                          type: object
                          properties:
                            nodeRoles:
                              description: Node role names
                              type: string
                        targetConfig:
                          description: Optional target configuration
                          type: object
                          properties:
                            targetName:
                              description: Optional target name for registration
                              type: string
                        args:
                          description: Kubeturbo21824 command line arguments
                          type: object
                          properties:
                            logginglevel:
                              description: Define logging level, default is info = 2
                              type: integer
                            kubelethttps:
                              description: Identify if kubelet requires https
                              type: boolean
                            kubeletport:
                              description: Identify kubelet port
                              type: integer
                            sccsupport:
                              description: Allow kubeturbo21824 to execute actions in OCP
                              type: string
                            failVolumePodMoves:
                              description: Allow kubeturbo21824 to reschedule pods with volumes attached
                              type: string
                            busyboxExcludeNodeLabels:
                              description: Do not run busybox on these nodes to discover the cpu frequency with k8s 1.18 and later, default is either of kubernetes.io/os=windows or beta.kubernetes.io/os=windows present as node label
                              type: string
                            stitchuuid:
                              description: Identify if using uuid or ip for stitching
                              type: boolean
                        resources:
                          description: Kubeturbo21824 resource configuration
                          type: object
                          properties:
                            limits:
                              description: Define limits
                              type: object
                              properties:
                                memory:
                                  description: Define memory limits in Gi or Mi, include units
                                  type: string
                                cpu:
                                  description: Define cpu limits in cores or millicores, include units
                                  type: string
                            requests:
                              description: Define requests
                              type: object
                              properties:
                                memory:
                                  description: Define memory requests in Gi or Mi, include units
                                  type: string
                                cpu:
                                  description: Define cpu requests in cores or millicores, include units
                                  type: string
    packages: |
      - channels:
        - currentCSV: kubeturbo21824-operator.v8.5.0
          name: alpha
        defaultChannel: alpha
        packageName: kubeturbo21824
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
parameters:
- name: NAME
- name: NAMESPACE

`)

func testQeTestdataOlmCm21824WrongYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCm21824WrongYaml, nil
}

func testQeTestdataOlmCm21824WrongYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCm21824WrongYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cm-21824-wrong.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCm25644EtcdCsvYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-etd-csv-template
objects:
- apiVersion: v1
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: '[{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdCluster","metadata":{"name":"example","namespace":"default"},"spec":{"size":3,"version":"3.2.13"}},{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdRestore","metadata":{"name":"example-etcd-cluster"},"spec":{"etcdCluster":{"name":"example-etcd-cluster"},"backupStorageType":"S3","s3":{"path":"<full-s3-path>","awsSecret":"<aws-secret>"}}},{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdBackup","metadata":{"name":"example-etcd-cluster-backup"},"spec":{"etcdEndpoints":["<etcd-cluster-endpoints>"],"storageType":"S3","s3":{"path":"<full-s3-path>","awsSecret":"<aws-secret>"}}}]'
            tectonic-visibility: ocs
          creationTimestamp: null
          name: etcdoperator.v0.9.2
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              - description: Limits describes the minimum/maximum amount of compute resources
                  required/allowed
                displayName: Resource Requirements
                path: pod.resources
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:resourceRequirements
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: serviceName
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to backup an etcd cluster.
              displayName: etcd Backup
              kind: EtcdBackup
              name: etcdbackups.etcd.database.coreos.com
              specDescriptors:
              - description: Specifies the endpoints of an etcd cluster.
                displayName: etcd Endpoint(s)
                path: etcdEndpoints
                x-descriptors:
                - urn:alm:descriptor:etcd:endpoint
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the backup was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any backup related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to restore an etcd cluster from a backup.
              displayName: etcd Restore
              kind: EtcdRestore
              name: etcdrestores.etcd.database.coreos.com
              specDescriptors:
              - description: References the EtcdCluster which should be restored,
                displayName: etcd Cluster
                path: etcdCluster.name
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:EtcdCluster
                - urn:alm:descriptor:text
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the restore was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any restore related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          description: |
            etcd is a distributed key value store that provides a reliable way to store data across a cluster of machines. It’s open-source and available on GitHub. etcd gracefully handles leader elections during network partitions and will tolerate machine failure, including the leader. Your applications can read and write data into etcd.
            A simple use-case is to store database connection details or feature flags within etcd as key value pairs. These values can be watched, allowing your app to reconfigure itself when they change. Advanced uses take advantage of the consistency guarantees to implement database leader elections or do distributed locking across a cluster of workers.
            _The etcd Open Cloud Service is Public Alpha. The goal before Beta is to fully implement backup features._
            ### Reading and writing to etcd
            Communicate with etcd though its command line utility ` + "`" + `etcdctl` + "`" + ` or with the API using the automatically generated Kubernetes Service.
            [Read the complete guide to using the etcd Open Cloud Service](https://coreos.com/tectonic/docs/latest/alm/etcd-ocs.html)
            ### Supported Features
            **High availability**
            Multiple instances of etcd are networked together and secured. Individual failures or networking issues are transparently handled to keep your cluster up and running.
            **Automated updates**
            Rolling out a new etcd version works like all Kubernetes rolling updates. Simply declare the desired version, and the etcd service starts a safe rolling update to the new version automatically.
            **Backups included**
            Coming soon, the ability to schedule backups to happen on or off cluster.
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-operator
                      - command:
                        - etcd-backup-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-backup-operator
                      - command:
                        - etcd-restore-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-restore-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  - etcdbackups
                  - etcdrestores
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: support@coreos.com
            name: CoreOS, Inc
          maturity: alpha
          provider:
            name: CoreOS, Inc
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.9.2
    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          creationTimestamp: null
          name: etcdclusters.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdCluster
            listKind: EtcdClusterList
            plural: etcdclusters
            shortNames:
            - etcdclus
            - etcd
            singular: etcdcluster
          scope: Namespaced
          version: v1beta2
        status:
          acceptedNames:
            kind: ""
            plural: ""
          conditions: null
          storedVersions: null
      - apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          creationTimestamp: null
          name: etcdbackups.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdBackup
            listKind: EtcdBackupList
            plural: etcdbackups
            singular: etcdbackup
          scope: Namespaced
          version: v1beta2
        status:
          acceptedNames:
            kind: ""
            plural: ""
          conditions: null
          storedVersions: null
      - apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          creationTimestamp: null
          name: etcdrestores.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdRestore
            listKind: EtcdRestoreList
            plural: etcdrestores
            singular: etcdrestore
          scope: Namespaced
          version: v1beta2
        status:
          acceptedNames:
            kind: ""
            plural: ""
          conditions: null
          storedVersions: null
    packages: |
      - channels:
        - currentCSV: etcdoperator.v0.9.2
          name: alpha
        defaultChannel: ""
        packageName: etcd
  kind: ConfigMap
  metadata:
    name:        "${NAME}"
    namespace:   "${NAMESPACE}"
    displayName: QE Operators
    publisher:   QE
parameters:
- name: NAME
- name: NAMESPACE

`)

func testQeTestdataOlmCm25644EtcdCsvYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCm25644EtcdCsvYaml, nil
}

func testQeTestdataOlmCm25644EtcdCsvYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCm25644EtcdCsvYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cm-25644-etcd-csv.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCmCsvEtcdYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-csv-etcd-template
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: '[{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdCluster","metadata":{"name":"example","namespace":"default"},"spec":{"size":3,"version":"3.2.13"}},{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdRestore","metadata":{"name":"example-etcd-cluster"},"spec":{"etcdCluster":{"name":"example-etcd-cluster"},"backupStorageType":"S3","s3":{"path":"<full-s3-path>","awsSecret":"<aws-secret>"}}},{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdBackup","metadata":{"name":"example-etcd-cluster-backup"},"spec":{"etcdEndpoints":["<etcd-cluster-endpoints>"],"storageType":"S3","s3":{"path":"<full-s3-path>","awsSecret":"<aws-secret>"}}}]'
            tectonic-visibility: ocs
          creationTimestamp: null
          name: etcdoperator.v0.9.2
          namespace: "${NAMESPACE}"
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              - description: Limits describes the minimum/maximum amount of compute resources
                  required/allowed
                displayName: Resource Requirements
                path: pod.resources
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:resourceRequirements
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: serviceName
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to backup an etcd cluster.
              displayName: etcd Backup
              kind: EtcdBackup
              name: etcdbackups.etcd.database.coreos.com
              specDescriptors:
              - description: Specifies the endpoints of an etcd cluster.
                displayName: etcd Endpoint(s)
                path: etcdEndpoints
                x-descriptors:
                - urn:alm:descriptor:etcd:endpoint
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the backup was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any backup related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to restore an etcd cluster from a backup.
              displayName: etcd Restore
              kind: EtcdRestore
              name: etcdrestores.etcd.database.coreos.com
              specDescriptors:
              - description: References the EtcdCluster which should be restored,
                displayName: etcd Cluster
                path: etcdCluster.name
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:EtcdCluster
                - urn:alm:descriptor:text
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the restore was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any restore related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          minKubeVersion: a.b.1
          description: |
            etcd is a distributed key value store that provides a reliable way to store data across a cluster of machines. It’s open-source and available on GitHub. etcd gracefully handles leader elections during network partitions and will tolerate machine failure, including the leader. Your applications can read and write data into etcd.
            A simple use-case is to store database connection details or feature flags within etcd as key value pairs. These values can be watched, allowing your app to reconfigure itself when they change. Advanced uses take advantage of the consistency guarantees to implement database leader elections or do distributed locking across a cluster of workers.

            _The etcd Open Cloud Service is Public Alpha. The goal before Beta is to fully implement backup features._

            ### Reading and writing to etcd

            Communicate with etcd though its command line utility ` + "`" + `etcdctl` + "`" + ` or with the API using the automatically generated Kubernetes Service.

            [Read the complete guide to using the etcd Open Cloud Service](https://coreos.com/tectonic/docs/latest/alm/etcd-ocs.html)

            ### Supported Features


            **High availability**


            Multiple instances of etcd are networked together and secured. Individual failures or networking issues are transparently handled to keep your cluster up and running.


            **Automated updates**


            Rolling out a new etcd version works like all Kubernetes rolling updates. Simply declare the desired version, and the etcd service starts a safe rolling update to the new version automatically.


            **Backups included**


            Coming soon, the ability to schedule backups to happen on or off cluster.
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-operator
                      - command:
                        - etcd-backup-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-backup-operator
                      - command:
                        - etcd-restore-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-restore-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  - etcdbackups
                  - etcdrestores
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: true
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: support@coreos.com
            name: CoreOS, Inc
          maturity: alpha
          provider:
            name: CoreOS, Inc
          replaces: etcdoperator.v0.9.0
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.9.2
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: '[{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdCluster","metadata":{"name":"example","namespace":"default"},"spec":{"size":3,"version":"3.2.13"}},{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdRestore","metadata":{"name":"example-etcd-cluster"},"spec":{"etcdCluster":{"name":"example-etcd-cluster"},"backupStorageType":"S3","s3":{"path":"<full-s3-path>","awsSecret":"<aws-secret>"}}},{"apiVersion":"etcd.database.coreos.com/v1beta2","kind":"EtcdBackup","metadata":{"name":"example-etcd-cluster-backup"},"spec":{"etcdEndpoints":["<etcd-cluster-endpoints>"],"storageType":"S3","s3":{"path":"<full-s3-path>","awsSecret":"<aws-secret>"}}}]'
            tectonic-visibility: ocs
          creationTimestamp: null
          name: etcdoperator.v0.9.0
          namespace: "${NAMESPACE}"
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              - description: Limits describes the minimum/maximum amount of compute resources
                  required/allowed
                displayName: Resource Requirements
                path: pod.resources
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:resourceRequirements
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: serviceName
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to backup an etcd cluster.
              displayName: etcd Backup
              kind: EtcdBackup
              name: etcdbackups.etcd.database.coreos.com
              specDescriptors:
              - description: Specifies the endpoints of an etcd cluster.
                displayName: etcd Endpoint(s)
                path: etcdEndpoints
                x-descriptors:
                - urn:alm:descriptor:etcd:endpoint
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the backup was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any backup related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to restore an etcd cluster from a backup.
              displayName: etcd Restore
              kind: EtcdRestore
              name: etcdrestores.etcd.database.coreos.com
              specDescriptors:
              - description: References the EtcdCluster which should be restored,
                displayName: etcd Cluster
                path: etcdCluster.name
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:EtcdCluster
                - urn:alm:descriptor:text
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the restore was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any restore related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          description: |
            etcd is a distributed key value store that provides a reliable way to store data across a cluster of machines. It’s open-source and available on GitHub. etcd gracefully handles leader elections during network partitions and will tolerate machine failure, including the leader. Your applications can read and write data into etcd.
            A simple use-case is to store database connection details or feature flags within etcd as key value pairs. These values can be watched, allowing your app to reconfigure itself when they change. Advanced uses take advantage of the consistency guarantees to implement database leader elections or do distributed locking across a cluster of workers.

            _The etcd Open Cloud Service is Public Alpha. The goal before Beta is to fully implement backup features._

            ### Reading and writing to etcd

            Communicate with etcd though its command line utility ` + "`" + `etcdctl` + "`" + ` or with the API using the automatically generated Kubernetes Service.

            [Read the complete guide to using the etcd Open Cloud Service](https://coreos.com/tectonic/docs/latest/alm/etcd-ocs.html)

            ### Supported Features


            **High availability**


            Multiple instances of etcd are networked together and secured. Individual failures or networking issues are transparently handled to keep your cluster up and running.


            **Automated updates**


            Rolling out a new etcd version works like all Kubernetes rolling updates. Simply declare the desired version, and the etcd service starts a safe rolling update to the new version automatically.


            **Backups included**


            Coming soon, the ability to schedule backups to happen on or off cluster.
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:db563baa8194fcfe39d1df744ed70024b0f1f9e9b55b5923c2f3a413c44dc6b8
                        name: etcd-operator
                      - command:
                        - etcd-backup-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:db563baa8194fcfe39d1df744ed70024b0f1f9e9b55b5923c2f3a413c44dc6b8
                        name: etcd-backup-operator
                      - command:
                        - etcd-restore-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:db563baa8194fcfe39d1df744ed70024b0f1f9e9b55b5923c2f3a413c44dc6b8
                        name: etcd-restore-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  - etcdbackups
                  - etcdrestores
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: true
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: support@coreos.com
            name: CoreOS, Inc
          maturity: alpha
          provider:
            name: CoreOS, Inc
          replaces: etcdoperator.v0.6.1
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.9.0
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            tectonic-visibility: ocs
          creationTimestamp: null
          name: etcdoperator.v0.6.1
          namespace: "${NAMESPACE}"
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: service
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          description: |
            etcd is a distributed key value store that provides a reliable way to store data across a cluster of machines. It’s open-source and available on GitHub. etcd gracefully handles leader elections during network partitions and will tolerate machine failure, including the leader. Your applications can read and write data into etcd.
            A simple use-case is to store database connection details or feature flags within etcd as key value pairs. These values can be watched, allowing your app to reconfigure itself when they change. Advanced uses take advantage of the consistency guarantees to implement database leader elections or do distributed locking across a cluster of workers.

            _The etcd Open Cloud Service is Public Alpha. The goal before Beta is to fully implement backup features._

            ### Reading and writing to etcd

            Communicate with etcd though its command line utility ` + "`" + `etcdctl` + "`" + ` or with the API using the automatically generated Kubernetes Service.

            [Read the complete guide to using the etcd Open Cloud Service](https://coreos.com/tectonic/docs/latest/alm/etcd-ocs.html)

            ### Supported Features
            **High availability**
            Multiple instances of etcd are networked together and secured. Individual failures or networking issues are transparently handled to keep your cluster up and running.
            **Automated updates**
            Rolling out a new etcd version works like all Kubernetes rolling updates. Simply declare the desired version, and the etcd service starts a safe rolling update to the new version automatically.
            **Backups included**
            Coming soon, the ability to schedule backups to happen on or off cluster.
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:bd944a211eaf8f31da5e6d69e8541e7cada8f16a9f7a5a570b22478997819943
                        name: etcd-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  verbs:
                  - '*'
                - apiGroups:
                  - storage.k8s.io
                  resources:
                  - storageclasses
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: true
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            alm-status-descriptors: etcdoperator.v0.6.1
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: support@coreos.com
            name: CoreOS, Inc
          maturity: alpha
          provider:
            name: CoreOS, Inc
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.6.1
    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          creationTimestamp: null
          name: etcdclusters.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdCluster
            listKind: EtcdClusterList
            plural: etcdclusters
            shortNames:
            - etcdclus
            - etcd
            singular: etcdcluster
          scope: Namespaced
          version: v1beta2
        status:
          acceptedNames:
            kind: ""
            plural: ""
          conditions: null
          storedVersions: null
      - apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          creationTimestamp: null
          name: etcdbackups.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdBackup
            listKind: EtcdBackupList
            plural: etcdbackups
            singular: etcdbackup
          scope: Namespaced
          version: v1beta2
        status:
          acceptedNames:
            kind: ""
            plural: ""
          conditions: null
          storedVersions: null
      - apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          creationTimestamp: null
          name: etcdrestores.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdRestore
            listKind: EtcdRestoreList
            plural: etcdrestores
            singular: etcdrestore
          scope: Namespaced
          version: v1beta2
        status:
          acceptedNames:
            kind: ""
            plural: ""
          conditions: null
          storedVersions: null
    packages: |
      - channels:
        - currentCSV: etcoperator.2
          name: alpha
        defaultChannel: ""
        packageName: etcd
parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmCmCsvEtcdYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCmCsvEtcdYaml, nil
}

func testQeTestdataOlmCmCsvEtcdYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCmCsvEtcdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cm-csv-etcd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCmNamespaceconfigYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-namespaceconfig-template
objects:
- apiVersion: v1
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: |-
              [
                {
                  "apiVersion": "redhatcop.redhat.io/v1alpha1",
                  "kind": "NamespaceConfig",
                  "metadata": {
                    "name": "example-namespaceconfig"
                  },
                  "spec": {
                    "size": 3
                  }
                }
              ]
            capabilities: Full Lifecycle
            categories: Security
            certified: "false"
            containerImage: quay.io/redhat-cop/namespace-configuration-operator:latest
            createdAt: 5/28/2019
            description: This operator provides a facility to define and enforce namespace
              configurations
            repository: https://github.com/redhat-cop/namespace-configuration-operator
            support: Best Effort
          name: namespace-configuration-operator.v0.1.0
          namespace: namespace-configuration-operator
        spec:
          apiservicedefinitions: {}
          customresourcedefinitions:
            owned:
            - description: Represent the desired configuration for a set of namespaces selected
                via labels
              displayName: Namespace Configuration
              kind: NamespaceConfig
              name: namespaceconfigs.redhatcop.redhat.io
              version: v1alpha1
          description: "The namespace configuration operator helps keeping a namespace's configuration
            aligned with one of more policies specified as a CRs.\n\nThe ` + "`" + `NamespaceConfig` + "`" + `
            CR allows specifying one or more objects that will be created in the selected
            namespaces.\n\nFor example using this operator an administrator can enforce a
            specific ResourceQuota or LimitRange on a set of namespaces. For example with
            the following snippet:\n\n` + "`" + `` + "`" + `` + "`" + `\napiVersion: redhatcop.redhat.io/v1alpha1\nkind:
            NamespaceConfig\nmetadata:\n  name: small-size\nspec:\n  selector:\n    matchLabels:\n
            \     size: small  \n  resources:\n  - apiVersion: v1\n    kind: ResourceQuota\n
            \   metadata:\n      name: small-size  \n    spec:\n      hard:\n        requests.cpu:
            \"4\"\n        requests.memory: \"2Gi\"\n` + "`" + `` + "`" + `` + "`" + `\n\nwe are enforcing that all the
            namespaces with label: ` + "`" + `size=small` + "`" + ` receive the specified resource quota.  \n"
          displayName: Namespace Configuration Operator
          icon:
          - base64data: iVBORw0KGgoAAAANSUhEUgAAAOoAAADYCAMAAADS+I/aAAAAgVBMVEX///8AAAD29vb8/Pz5+fnz8/Pq6urf3994eHjIyMi8vLzU1NTQ0NB8fHzx8fGrq6szMzOQkJBtbW2enp5BQUGxsbGkpKRJSUlfX1/d3d1mZmaEhITAwMA5OTkVFRXl5eVTU1MmJiYfHx+WlpZaWloODg6Li4suLi4ZGRk9PT1HR0fjV/a/AAAPPUlEQVR4nM1daUPqOhBVQHZEQJBVWhHw+v9/4LWt0LQ9k8xkaT2fru9BkiHbmTUPD0HxvNlPZ5P559vh5QeH72g9709mq+lw1GuF7blGHPen7eFRi8P2NDw2PU43dIe7b72QKr4vw27TI7ZCdzl/54t5w0d/+Nz0yGUYzWK5mDccTpumx8/FeGIv5g27UdNSmLHZucuZ4dJrWhYdWtMXX4ImOEw7TUtEoNf3KWeGyV+c2oHgWpHgbdC0ZCUM4zCCJvgYNi2dguVXOEETvC+blvAX+zisoAn+xMyODfzWFw7jhgXtrusRNMFnowx5Vp+gCWaNCTo41yvpD5q5eVrb2gX9wbwBAjVsQtAE+5oF7dhM6dt2t1gth4PxDwbD5eo0Wb9aNNOvVdKxbHCH/mpAGVTavf1iHouae6+RGF/4wzrPpxxFuzNaSdbJKriIGZ7Yq249FRnHequI2/I2lHAFjJij6Q/a8sY7+zmv9Y8a+MSKNZK5wwW45y3l4FcsR/1+WVrMp4rONGZ0E3jDvplHMPdiARsxpnbioyMCT1dj9xdvZtyu2SD36auvCo7Gvk9eaVvLqE28+uxOQc/U8cw7P+2YbvBrEEpsumT6TyF6fTZcPh8B/HgGSb+DkbWNnrG8e/+FDZIGtXMta5V1o+1tG1iJ7GhX8YfX3vVnbw0a5EDX/9WRsah41nUU1eLgb+vsdd/+utH5hBfeujFAt2O9cYl/mk5q9IMeY3oYnjiiho3+qzc6RbOIvXD/E93+3Ef7nsbiQafb063Xtk1Zo3HW1bt02414jOgL/sO16ZhsuiF3UZd0KTjam2ia0ljAzdMHNSSno4m+yxqMUeiQK83h56c3aqMRgW1K1nf7NkkvccOxj+S8Wt9+5C3WeIRNi6KqlpoHaWD5A0Fx5NayU+io5Vu31w+Cul+tljBlxucf6a3jcTPaHI9dj/rkHRRvsiCIlI7K1SGK7oi3i3fGscDjO8tbIrSIiPftJxBXOfMc1EzQm4u0HcJZfGauRezwmPjV+Qhfg/QmjJ2aaeFve1YRiGM4krUyxa1wbaCkqH5dSkToiehkauM2+Cc5HdPu1d+Nt2ssaQLzpC9BC7QbbScURwus0QkM8B08SBFLGpNmRp/XDnY6CC4c7P+SnuLExff4Imzn5+raflGd49XDpjl4UuUGjRaxiqUcOjXtX4mLCi9hbtPYg2tD8rHLUBhOdjs4sJ0ME8Qps234ZUtVELoMJQ20P+9fw5c6jHVi7lbM862JDriiBfpuV7UkwXnFuiaPq8Cj88QfnXks/Mug6IHDxwXcJKyzDy9+Fz2sQmrYV2t5gcFz+AkOmHO0QKbuZMiv0MSI+cUK53qDH4MHPYOVYa8xVyqICi1nWvaqBw5e+Zhxm+MG4E/k5p2pbgnOdmhVVd4D8VF4O5rHDH8h7k5dHFAcZfXcYCjpyGFBqZBwWo2UBx5K3KSP9BR5KQsLBmJWe5EuSvMCyGRNBxO0szDv1JvqFxXvP6DQmRk/8NRrDhpoCDOoxnApcNXpXM1QlwEyY5ov+OqRpNVCYeyuvgfoj+Iaa1Sh7sbiNsqCNO+I6gLWuorhvaG3RiBCGfEELWlE0ZFukRPxWf7RDVYU1I2WtsP1yzfVFNnH5OdmaxOx0mYWXDJZmq4OGMOl+wKySwk8eeWZuFDa+WMklNTMfZDeqnO5oLgdCX2gJKvCwPhLkv4zd43MYbofCA1KYpMX5Bppl3BJ0neGfw0eTOn/aXeej73RaLRRJUFGqUggqcb+W4Xmsi67FlhxOkj3vF6L5//2HkKAFoEs0leQu0trlLmkryempO3uwJxCkeBmYEOlHYSBxXxRCY1MXVs/3GWz3hlW72Y6iQQ5w9nUIUshNR4KkhxIHO+ZS8qxfJBnPIkRNUxxLJAk/RydkPkgGF13LMoXpD4KtFXloUCSmgJVPpNLyrFvxnJJM0qEyJVYUlYyGSmrTFJmVmQJiWUL/Gcrvxk7C/WxbP7O9ylHUn3uBIlXrA3bhe1JykupsuZzyrri7CY1MYcjA4Rl1J42geIHqsk/14aFkkoutgLwqWQn6YM2HjtZKyovu8kqlVQTqqzFB+T6DJpNob2i6cuiaCXLZM1XApOgcTPBy9jCgBG3yIXufjXrr6PXKjsd/9DlOP8ztfbfNUAuFbUVdQmXPtd3Z5Z6syxY0xK9RrGUqbKySbc20YlGjLUgryFkLdUKmFB45QLOZRWoF2S4txZHTAs9x4/lrtLH81Px75usEkXKqgTS8AEbC/1KWrgJ0/xEhURmsop2DHbB6ZH68oCOQPlI7KHYa9P8ROXUtwjykR9MUfo9YIQPkKGvEKlUN1fmuQZZf02CgGZ5jab6hbLBUmtw36k7oay/9gygfDlEBdBQeFSq97vJaixroOJ2oQDdy9u1WoByKHiQlc6mquCuu4Bg5UBR+hpZxXHLP+dwzJQ0/x3BN0KlvSnRHFHyt9u8QutfFVH+BeAxC5ZlMi2NwFFWTtUi1WcJ1ny41CHPsjJsAaqRF4gasFKVImt5v9qoU0RUeo6CjQH8/5B14T3LarB7FA9Y8IEgBWhuUGRN6zooFMamUqH2gi2ZyMAnwtYHUGRNeZMiq02Y6jOt05W3f52z2km3hiLrvxL3t5G1TZXwqVh46zyWvrNyOoqsh0RWRX9d2zSLHQvVwAtB9Jcr3m6lgxRZ03pYCjm10qpQTE9cDaUDFCJQ6nzGbtI1rPCmtB6Wwk6lPsAUwDABdmFc/VQYYnibunR7KLKef4TvKGsrXg164tOiEqKONiGg+54KQI4K3d0XabZEVePHERnzDtG6f1kMN8yQzpKskPAB0uylDNjwvXCD331XtwWqyjrSFfr8XrAoTcFGhqktOL58FGnZFluqSFr0eSz1kSN9zk3fvV/QO2IlAIOLe0X01s1l8KuI3levWvhKPUu2Bu8sK9OitT9Ndgt6+wEzmnNWYis/ZVKyByUVhTt5sXcB35lrzbGOytXmiqTlW13ikLUiF0UgPcixyWJQxvxOhqr8RfKGk7usSA1y4/tUoAByZUqe/XEu0Y60ICeLCxVShH0GkqrlzjcDaNPpYhVJWuD6RrgavYC+53LcEe5P2hEvCIyxqHNRAPhZrSj3L4hQBc03BE8rOBYiQO5Kh+aIClyaQbZ5wZ8p3PRLFM1uq8Y9Lcgp0pyfgnBitxsHRQjYeW2etVeHRlZBfI7btIIGraih6RUCjWOGrjtXPjTdCKKncEpzPBzN2YkyI4+TzkOvdPPay/ngKUiWyqwpgA7cxq6m7LcZFYim0wpG1FBKTDqsU5QeJlRy7qtA5elOHlF0AAqL+NOlIwugG0DXgLItO/mB51bPCNU8FtmCCwyJllpzrqMY3+IHbreYmzkTkQjJdaPeFYdnMoQq0jQBUiXKeyh7WFHXCAMoIE3ADVVJU3MgPqE+tLa/6uerBtLhP3djBBoZewWrkqb3MSb8Z71tt7KJ/MeJpUBOD+4ZrIbBZd4l+JqwQdKq3hqoijYiK8ziQ2oZ0Gx00FkUmwwblQ0eIk7sgQhQZOnBaiHbjOSm5qJxyWT1arTRV37tULXR0UHC6ktRZDJJE+p1LUXCcgzL1dsmUJgCDJ9g+IgUeTJJV7cpvJ/qy2dWldfqbfMVxvcJV7A51FBR2jJJk2WYXlP5KcwdAlB0w2xXaMwyfUmxDWarPaF3UfKP9p0ysUMzoaExxENpsOCjoSMlUrUkqWL0ZntKCfXcf2gnVBgNBrr8/sxcpsP7v3INWGCmpizC3l9Cgi4F7U+aT0OU/p1Lqgxa4CcgrS6xZy8+TB/UTut9zWfZ48u7zIo5QqRc0pnGns8naOrTscP7BZXO3OouqX2YL23o98sncCVazZ14s5qlrOqEJBVbMunMZr9BN7ALzdK5UblkJyVrNjNcKJJaPEdIrmG/gee4G1qXu7vwVsNEA8ve91ITamwGcSQ8OH5jAbGBUqOiFz53TndsOU3KAhtkZPX9hg4OTaB7Kehd6ewrt8yXvTu6MyxJu/Me9Ei4Tkj9S/18upeUxef6DGF3MN31+/P+brEPwvpxFAZ90Cs65qBYl+IcMlA8w3G1282mtmkGRAohHQikLNjBSAkTNFlXPOB2/l0ttzERSEQfMLjAiP8HQ8t4jvPe7JIVibrlNBWAj+kY7UjOKNok7YoAEH5DmvarWfK3nkM8eFJAp5Swb0cbUfXBR12SUbtswI3shi9BZdtYRb4Qz8W8a2aqeAnW8AhilSnbbVeCmOkaUxWFIKmvRaAr0WrPUAqyTh/r3JzRhxqeBoTGTbvYLaoKi97u0R0sl4PwvIHKDLO8XGNC1oYecyyCyAuzzDYgKzc1/CBeAmp72Q6NDL4OmOzJA/kknnWLVAhHDdRWC/JRR4tU9V+Qz5mHJ7c60LGIDkyUjCv7aFBWWlInozhppqxBESUA2PYvIqd2qUC4x8bOYU1GrqMipclrbuTBTvL48DAeOpaziSewNdVbPbjqNOk+NXD6IjQ1EZwzUhJo0gc8ZDJJoEk8ivz0ENM9xDUSpydNGoPowT9dH3QXNW5YbZELb9e85tirxdyQQJs153FtaetynWu4dTbaOGOvnkh9iZRAsXE59Emfnn9qvaxnTwn4RN/kS5gpvJsKDKVv1sE4cdeQFxjgVzaVN54FsXB3TAnLQcw/xpL6/kPG2tqKygkCHYnGMuFnvyFjbfP7AcFKmplT1s4rb8u4ZZzRx/eAXA08U1fBxUv/R0ai/cEyyoIJTpp05Hwo7jnZZl7f0kbg1R/eORCYDa9KRA0aJOGjKyOeWUm7mcW89oOSlhuOzME8nid7kbnnaT/h1kq91qU8Cl5sOFz2rFF19xfBgy1eTA484JBLEuuTpjJUuzc8SR6GeazZqNUVFDa44W17WSz349Gm1zv2NqPxYLm4mGoOIXzXbYI2JZQHQ4iMBQOOvLK8nvHdjBfQWL7VP7ynKnDRsnw8xxbz4DFfGowYpNgXDo14ThRoHB1ecW7AbVKB/FU6CzRw7iJ0zJqlIxbBwxXZMBtGXPBHZvSOKRGA6YprY/eLBgPJU3hMrP9EPBjA0erdIBJfp8YDpHQYeGMV/b86oTnaQ0lJQgLzQC+R+MeAbU4AOO9qMab4Q28q1LczrKd/IFjTAr1pX1Cv79q3Tgz6G2iNpxOjZhvtpqOw9uv68DwaTmeT+fbz7fUlwb/vt8/5fHJZTIfjXt2s7z9MqsTdLqoFFgAAAABJRU5ErkJggg==
            mediatype: image/png
          install:
            spec:
              clusterPermissions:
              - rules:
                - apiGroups:
                  - '*'
                  resources:
                  - '*'
                  verbs:
                  - '*'
                serviceAccountName: namespace-configuration-operator
              deployments:
              - name: namespace-configuration-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: namespace-configuration-operator
                  strategy: {}
                  template:
                    metadata:
                      labels:
                        name: namespace-configuration-operator
                    spec:
                      containers:
                      - command:
                        - namespace-configuration-operator
                        env:
                        - name: WATCH_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.annotations['olm.targetNamespaces']
                        - name: POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        - name: OPERATOR_NAME
                          value: namespace-configuration-operator
                        image: quay.io/redhat-cop/namespace-configuration-operator:v0.1.0
                        imagePullPolicy: Always
                        name: namespace-configuration-operator
                        resources: {}
                      serviceAccountName: namespace-configuration-operator
              permissions:
              - rules:
                - apiGroups:
                  - ""
                  resources:
                  - configmaps
                  - pods
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - services
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - replicasets
                  - deployments
                  verbs:
                  - get
                  - list
                - apiGroups:
                  - monitoring.coreos.com
                  resources:
                  - servicemonitors
                  verbs:
                  - get
                  - create
                - apiGroups:
                  - apps
                  resourceNames:
                  - namespace-configuration-operator
                  resources:
                  - deployments/finalizers
                  verbs:
                  - update
                serviceAccountName: namespace-configuration-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          keywords:
          - namespace
          - configuration
          - policy
          - management
          links:
          - name: repository
            url: https://github.com/redhat-cop/namespace-configuration-operator
          - name: conatinerImage
            url: https://quay.io/redhat-cop/namespace-configuration-operator:latest
          - name: blog
            url: https://blog.openshift.com/controlling-namespace-configurations
          maintainers:
          - email: rspazzol@redhat.com
            name: Raffaele Spazzoli
          maturity: alpha
          minKubeVersion: 1.10.0
          provider:
            name: Containers & PaaS CoP
          replaces: namespace-configuration-operator.v0.0.1
          skips:
          - namespace-configuration-operator.v0.0.2
          version: 0.1.0
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: '[{"apiVersion":"redhatcop.redhat.io/v1alpha1","kind":"NamespaceConfig","metadata":{"name":"example-namespaceconfig"},"spec":{"size":3}}]'
            capabilities: Full Lifecycle
            categories: Security
            certified: "false"
            containerImage: quay.io/redhat-cop/namespace-configuration-operator:latest
            createdAt: 5/28/2019
            description: This operator provides a facility to define and enforce namespace
              configurations
            repository: https://github.com/redhat-cop/namespace-configuration-operator
            support: Best Effort
          name: namespace-configuration-operator.v0.0.2
          namespace: namespace-configuration-operator
        spec:
          apiservicedefinitions: {}
          customresourcedefinitions:
            owned:
            - description: Represent the desired configuration for a set of namespaces selected
                via labels
              displayName: Namespace Configuration
              kind: NamespaceConfig
              name: namespaceconfigs.redhatcop.redhat.io
              version: v1alpha1
          description: "The namespace configuration operator helps keeping a namespace's configuration
            aligned with one of more policies specified as a CRs.\n\nThe ` + "`" + `NamespaceConfig` + "`" + `
            CR allows specifying one or more objects that will be created in the selected
            namespaces.\n\nFor example using this operator an administrator can enforce a
            specific ResourceQuota or LimitRange on a set of namespaces. For example with
            the following snippet:\n\n` + "`" + `` + "`" + `` + "`" + `\napiVersion: redhatcop.redhat.io/v1alpha1\nkind:
            NamespaceConfig\nmetadata:\n  name: small-size\nspec:\n  selector:\n    matchLabels:\n
            \     size: small  \n  resources:\n  - apiVersion: v1\n    kind: ResourceQuota\n
            \   metadata:\n      name: small-size  \n    spec:\n      hard:\n        requests.cpu:
            \"4\"\n        requests.memory: \"2Gi\"\n` + "`" + `` + "`" + `` + "`" + `\n\nwe are enforcing that all the
            namespaces with label: ` + "`" + `size=small` + "`" + ` receive the specified resource quota.  \n"
          displayName: Namespace Configuration Operator
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              clusterPermissions:
              - rules:
                - apiGroups:
                  - '*'
                  resources:
                  - '*'
                  verbs:
                  - '*'
                serviceAccountName: namespace-configuration-operator
              deployments:
              - name: namespace-configuration-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: namespace-configuration-operator
                  strategy: {}
                  template:
                    metadata:
                      labels:
                        name: namespace-configuration-operator
                    spec:
                      containers:
                      - command:
                        - namespace-configuration-operator
                        env:
                        - name: WATCH_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.annotations['olm.targetNamespaces']
                        - name: POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        - name: OPERATOR_NAME
                          value: namespace-configuration-operator
                        image: quay.io/redhat-cop/namespace-configuration-operator:v0.0.2
                        imagePullPolicy: Always
                        name: namespace-configuration-operator
                        resources: {}
                      serviceAccountName: namespace-configuration-operator
              permissions:
              - rules:
                - apiGroups:
                  - ""
                  resources:
                  - configmaps
                  - pods
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - services
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - replicasets
                  - deployments
                  verbs:
                  - get
                  - list
                - apiGroups:
                  - monitoring.coreos.com
                  resources:
                  - servicemonitors
                  verbs:
                  - get
                  - create
                - apiGroups:
                  - apps
                  resourceNames:
                  - namespace-configuration-operator
                  resources:
                  - deployments/finalizers
                  verbs:
                  - update
                serviceAccountName: namespace-configuration-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          keywords:
          - namespace
          - configuration
          - policy
          - management
          links:
          - name: repository
            url: https://github.com/redhat-cop/namespace-configuration-operator
          - name: conatinerImage
            url: https://quay.io/redhat-cop/namespace-configuration-operator:latest
          - name: blog
            url: https://blog.openshift.com/controlling-namespace-configurations
          maintainers:
          - email: rspazzol@redhat.com
            name: Raffaele Spazzoli
          maturity: alpha
          provider:
            name: Containers & PaaS CoP
          replaces: namespace-configuration-operator.v0.0.1
          version: 0.0.2
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            capabilities: Full Lifecycle
            categories: Security
            certified: "false"
            description: This operator provides a facility to define and enforce namespace configurations
            containerImage: quay.io/redhat-cop/namespace-configuration-operator:latest
            createdAt: 5/28/2019
            support: Best Effort
            repository: https://github.com/redhat-cop/namespace-configuration-operator
            alm-examples: |
              [
                {
                  "apiVersion": "redhatcop.redhat.io/v1alpha1",
                  "kind": "NamespaceConfig",
                  "metadata": {
                      "name": "small-size"
                  },
                  "spec": {
                      "selector": {
                        "matchLabels": {
                            "size": "small"
                        }
                      },
                      "resources": [
                        {
                            "apiVersion": "v1",
                            "kind": "ResourceQuota",
                            "metadata": {
                              "name": "small-size"
                            },
                            "spec": {
                              "hard": {
                                  "requests.cpu": "4",
                                  "requests.memory": "2Gi"
                              }
                            }
                        }
                      ]
                  }
                }
              ]
          name: namespace-configuration-operator.v0.0.1
          namespace: namespace-configuration-operator
        spec:
          icon:
          - base64data: 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
            mediatype: image/png
          links:
          - name: repository
            url: https://github.com/redhat-cop/namespace-configuration-operator
          - name: conatinerImage
            url: https://quay.io/redhat-cop/namespace-configuration-operator:latest
          - name: blog
            url: https://blog.openshift.com/controlling-namespace-configurations
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          maturity: alpha
          version: 0.0.1
          keywords: ['namespace', 'configuration', 'policy', 'management']
          maintainers:
          - name: Raffaele Spazzoli
            email: rspazzol@redhat.com
          provider:
            name: Containers & PaaS CoP
          apiservicedefinitions: {}
          description: |
            The namespace configuration operator helps keeping a namespace's configuration aligned with one of more policies specified as a CRs.
            The ` + "`" + `NamespaceConfig` + "`" + ` CR allows specifying one or more objects that will be created in the selected namespaces.
            For example using this operator an administrator can enforce a specific ResourceQuota or LimitRange on a set of namespaces. For example with the following snippet:
            ` + "`" + `` + "`" + `` + "`" + `
            apiVersion: redhatcop.redhat.io/v1alpha1
            kind: NamespaceConfig
            metadata:
              name: small-size
            spec:
              selector:
                matchLabels:
                  size: small
              resources:
              - apiVersion: v1
                kind: ResourceQuota
                metadata:
                  name: small-size
                spec:
                  hard:
                    requests.cpu: "4"
                    requests.memory: "2Gi"
            ` + "`" + `` + "`" + `` + "`" + `
            we are enforcing that all the namespaces with label: ` + "`" + `size=small` + "`" + ` receive the specified resource quota.
          customresourcedefinitions:
            owned:
            - kind: NamespaceConfig
              name: namespaceconfigs.redhatcop.redhat.io
              version: v1alpha1
              displayName: Namespace Configuration
              description: Represent the desired configuration for a set of namespaces selected via labels
          displayName: Namespace Configuration Operator
          install:
            spec:
              clusterPermissions:
              - rules:
                - apiGroups:
                  - "*"
                  resources:
                  - "*"
                  verbs:
                  - '*'
                serviceAccountName: namespace-configuration-operator
              deployments:
              - name: namespace-configuration-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: namespace-configuration-operator
                  strategy: {}
                  template:
                    metadata:
                      labels:
                        name: namespace-configuration-operator
                    spec:
                      containers:
                      - command:
                        - namespace-configuration-operator
                        env:
                        - name: WATCH_NAMESPACE
                          value: ""
                        - name: POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        - name: OPERATOR_NAME
                          value: namespace-configuration-operator
                        image: quay.io/redhat-cop/namespace-configuration-operator:latest
                        imagePullPolicy: Always
                        name: namespace-configuration-operator
                        resources: {}
                      serviceAccountName: namespace-configuration-operator
              permissions:
              - rules:
                - apiGroups:
                  - ""
                  resources:
                  - configmaps
                  - pods
                  verbs:
                  - '*'
                - apiGroups:
                  - ""
                  resources:
                  - services
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - replicasets
                  - deployments
                  verbs:
                  - get
                  - list
                - apiGroups:
                  - monitoring.coreos.com
                  resources:
                  - servicemonitors
                  verbs:
                  - get
                  - create
                - apiGroups:
                  - apps
                  resourceNames:
                  - namespace-configuration-operator
                  resources:
                  - deployments/finalizers
                  verbs:
                  - update
                serviceAccountName: namespace-configuration-operator
            strategy: deployment
    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: namespaceconfigs.redhatcop.redhat.io
        spec:
          group: redhatcop.redhat.io
          names:
            kind: NamespaceConfig
            listKind: NamespaceConfigList
            plural: namespaceconfigs
            singular: namespaceconfig
          scope: Namespaced
          versions:
          - name: v1alpha1
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
    packages: |
      - channels:
        - currentCSV: namespace-configuration-operator.v0.1.0
          name: alpha
        defaultChannel: alpha
        packageName: namespace-configuration-operator
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
parameters:
- name: NAME
- name: NAMESPACE

`)

func testQeTestdataOlmCmNamespaceconfigYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCmNamespaceconfigYaml, nil
}

func testQeTestdataOlmCmNamespaceconfigYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCmNamespaceconfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cm-namespaceconfig.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCmTemplateYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-sub-template
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  data:
    special.how: very
    special.type: charm
parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmCmTemplateYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCmTemplateYaml, nil
}

func testQeTestdataOlmCmTemplateYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCmTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cm-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmConfigmapEctdAlphaBetaYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-bad-operator-template
objects:
- apiVersion: v1
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: "[\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdCluster\",\n    \"metadata\": {\n      \"name\": \"example\"\
              \n    },\n    \"spec\": {\n      \"size\": 3,\n      \"version\": \"3.2.13\"\
              \n    }\n  },\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdRestore\",\n    \"metadata\": {\n      \"name\": \"example-etcd-cluster-restore\"\
              \n    },\n    \"spec\": {\n      \"etcdCluster\": {\n        \"name\": \"example-etcd-cluster\"\
              \n      },\n      \"backupStorageType\": \"S3\",\n      \"s3\": {\n        \"\
              path\": \"<full-s3-path>\",\n        \"awsSecret\": \"<aws-secret>\"\n     \
              \ }\n    }\n  },\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdBackup\",\n    \"metadata\": {\n      \"name\": \"example-etcd-cluster-backup\"\
              \n    },\n    \"spec\": {\n      \"etcdEndpoints\": [\"<etcd-cluster-endpoints>\"\
              ],\n      \"storageType\":\"S3\",\n      \"s3\": {\n        \"path\": \"<full-s3-path>\"\
              ,\n        \"awsSecret\": \"<aws-secret>\"\n      }\n    }\n  }\n]\n"
            capabilities: Full Lifecycle
            categories: Database
            description: Creates and maintain highly-available etcd clusters on Kubernetes
            tectonic-visibility: ocs
          name: etcdoperator.v0.9.2
          namespace: placeholder
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              - description: Limits describes the minimum/maximum amount of compute resources
                  required/allowed
                displayName: Resource Requirements
                path: pod.resources
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:resourceRequirements
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: serviceName
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to backup an etcd cluster.
              displayName: etcd Backup
              kind: EtcdBackup
              name: etcdbackups.etcd.database.coreos.com
              specDescriptors:
              - description: Specifies the endpoints of an etcd cluster.
                displayName: etcd Endpoint(s)
                path: etcdEndpoints
                x-descriptors:
                - urn:alm:descriptor:etcd:endpoint
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the backup was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any backup related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to restore an etcd cluster from a backup.
              displayName: etcd Restore
              kind: EtcdRestore
              name: etcdrestores.etcd.database.coreos.com
              specDescriptors:
              - description: References the EtcdCluster which should be restored,
                displayName: etcd Cluster
                path: etcdCluster.name
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:EtcdCluster
                - urn:alm:descriptor:text
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the restore was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any restore related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          description: "The etcd Operater creates and maintains highly-available etcd clusters\
            \ on Kubernetes, allowing engineers to easily deploy and manage etcd clusters\
            \ for their applications.\n\netcd is a distributed key value store that provides\
            \ a reliable way to store data across a cluster of machines. It\xE2\u20AC\u2122\
            s open-source and available on GitHub. etcd gracefully handles leader elections\
            \ during network partitions and will tolerate machine failure, including the leader.\n\
            \n\n### Reading and writing to etcd\n\nCommunicate with etcd though its command\
            \ line utility ` + "`" + `etcdctl` + "`" + ` via port forwarding:\n\n    $ kubectl --namespace default\
            \ port-forward service/example-client 2379:2379\n    $ etcdctl --endpoints http://127.0.0.1:2379\
            \ get /\n\nOr directly to the API using the automatically generated Kubernetes\
            \ Service:\n\n    $ etcdctl --endpoints http://example-client.default.svc:2379\
            \ get /\n\nBe sure to secure your etcd cluster (see Common Configurations) before\
            \ exposing it outside of the namespace or cluster.\n\n\n### Supported Features\n\
            \n* **High availability** - Multiple instances of etcd are networked together\
            \ and secured. Individual failures or networking issues are transparently handled\
            \ to keep your cluster up and running.\n\n* **Automated updates** - Rolling out\
            \ a new etcd version works like all Kubernetes rolling updates. Simply declare\
            \ the desired version, and the etcd service starts a safe rolling update to the\
            \ new version automatically.\n\n* **Backups included** - Create etcd backups and\
            \ restore them through the etcd Operator.\n\n### Common Configurations\n\n* **Configure\
            \ TLS** - Specify [static TLS certs](https://github.com/coreos/etcd-operator/blob/master/doc/user/cluster_tls.md)\
            \ as Kubernetes secrets.\n\n* **Set Node Selector and Affinity** - [Spread your\
            \ etcd Pods](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#three-member-cluster-with-node-selector-and-anti-affinity-across-nodes)\
            \ across Nodes and availability zones.\n\n* **Set Resource Limits** - [Set the\
            \ Kubernetes limit and request](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#three-member-cluster-with-resource-requirement)\
            \ values for your etcd Pods.\n\n* **Customize Storage** - [Set a custom StorageClass](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#custom-persistentvolumeclaim-definition)\
            \ that you would like to use.\n"
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-operator
                      - command:
                        - etcd-backup-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-backup-operator
                      - command:
                        - etcd-restore-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-restore-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  - etcdbackups
                  - etcdrestores
                  verbs:
                  - '*'
                - apiGroups:
                  - ''
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ''
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: etcd-dev@googlegroups.com
            name: etcd Community
          maturity: alpha
          provider:
            name: CNCF
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.9.2
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: "[\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdCluster\",\n    \"metadata\": {\n      \"name\": \"example\"\
              \n    },\n    \"spec\": {\n      \"size\": 3,\n      \"version\": \"3.2.13\"\
              \n    }\n  },\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdRestore\",\n    \"metadata\": {\n      \"name\": \"example-etcd-cluster-restore\"\
              \n    },\n    \"spec\": {\n      \"etcdCluster\": {\n        \"name\": \"example-etcd-cluster\"\
              \n      },\n      \"backupStorageType\": \"S3\",\n      \"s3\": {\n        \"\
              path\": \"<full-s3-path>\",\n        \"awsSecret\": \"<aws-secret>\"\n     \
              \ }\n    }\n  },\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdBackup\",\n    \"metadata\": {\n      \"name\": \"example-etcd-cluster-backup\"\
              \n    },\n    \"spec\": {\n      \"etcdEndpoints\": [\"<etcd-cluster-endpoints>\"\
              ],\n      \"storageType\":\"S3\",\n      \"s3\": {\n        \"path\": \"<full-s3-path>\"\
              ,\n        \"awsSecret\": \"<aws-secret>\"\n      }\n    }\n  }\n]\n"
            capabilities: Full Lifecycle
            categories: Database
            containerImage: quay.io/coreos/etcd-operator@sha256:66a37fd61a06a43969854ee6d3e21087a98b93838e284a6086b13917f96b0d9b
            createdAt: 2019-02-28 01:03:00
            description: Create and maintain highly-available etcd clusters on Kubernetes
            repository: https://github.com/coreos/etcd-operator
            tectonic-visibility: ocs
          name: etcdoperator.v0.9.4
          namespace: placeholder
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              - description: Limits describes the minimum/maximum amount of compute resources
                  required/allowed
                displayName: Resource Requirements
                path: pod.resources
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:resourceRequirements
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: serviceName
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta3
            - description: Represents the intent to backup an etcd cluster.
              displayName: etcd Backup
              kind: EtcdBackup
              name: etcdbackups.etcd.database.coreos.com
              specDescriptors:
              - description: Specifies the endpoints of an etcd cluster.
                displayName: etcd Endpoint(s)
                path: etcdEndpoints
                x-descriptors:
                - urn:alm:descriptor:etcd:endpoint
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the backup was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any backup related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to restore an etcd cluster from a backup.
              displayName: etcd Restore
              kind: EtcdRestore
              name: etcdrestores.etcd.database.coreos.com
              specDescriptors:
              - description: References the EtcdCluster which should be restored,
                displayName: etcd Cluster
                path: etcdCluster.name
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:EtcdCluster
                - urn:alm:descriptor:text
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the restore was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any restore related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          description: "The etcd Operater creates and maintains highly-available etcd clusters\
            \ on Kubernetes, allowing engineers to easily deploy and manage etcd clusters\
            \ for their applications.\n\netcd is a distributed key value store that provides\
            \ a reliable way to store data across a cluster of machines. It\xE2\u20AC\u2122\
            s open-source and available on GitHub. etcd gracefully handles leader elections\
            \ during network partitions and will tolerate machine failure, including the leader.\n\
            \n\n### Reading and writing to etcd\n\nCommunicate with etcd though its command\
            \ line utility ` + "`" + `etcdctl` + "`" + ` via port forwarding:\n\n    $ kubectl --namespace default\
            \ port-forward service/example-client 2379:2379\n    $ etcdctl --endpoints http://127.0.0.1:2379\
            \ get /\n\nOr directly to the API using the automatically generated Kubernetes\
            \ Service:\n\n    $ etcdctl --endpoints http://example-client.default.svc:2379\
            \ get /\n\nBe sure to secure your etcd cluster (see Common Configurations) before\
            \ exposing it outside of the namespace or cluster.\n\n\n### Supported Features\n\
            \n* **High availability** - Multiple instances of etcd are networked together\
            \ and secured. Individual failures or networking issues are transparently handled\
            \ to keep your cluster up and running.\n\n* **Automated updates** - Rolling out\
            \ a new etcd version works like all Kubernetes rolling updates. Simply declare\
            \ the desired version, and the etcd service starts a safe rolling update to the\
            \ new version automatically.\n\n* **Backups included** - Create etcd backups and\
            \ restore them through the etcd Operator.\n\n### Common Configurations\n\n* **Configure\
            \ TLS** - Specify [static TLS certs](https://github.com/coreos/etcd-operator/blob/master/doc/user/cluster_tls.md)\
            \ as Kubernetes secrets.\n\n* **Set Node Selector and Affinity** - [Spread your\
            \ etcd Pods](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#three-member-cluster-with-node-selector-and-anti-affinity-across-nodes)\
            \ across Nodes and availability zones.\n\n* **Set Resource Limits** - [Set the\
            \ Kubernetes limit and request](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#three-member-cluster-with-resource-requirement)\
            \ values for your etcd Pods.\n\n* **Customize Storage** - [Set a custom StorageClass](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#custom-persistentvolumeclaim-definition)\
            \ that you would like to use.\n"
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:66a37fd61a06a43969854ee6d3e21087a98b93838e284a6086b13917f96b0d9b
                        name: etcd-operator
                      - command:
                        - etcd-backup-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:66a37fd61a06a43969854ee6d3e21087a98b93838e284a6086b13917f96b0d9b
                        name: etcd-backup-operator
                      - command:
                        - etcd-restore-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:66a37fd61a06a43969854ee6d3e21087a98b93838e284a6086b13917f96b0d9b
                        name: etcd-restore-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  - etcdbackups
                  - etcdrestores
                  verbs:
                  - '*'
                - apiGroups:
                  - ''
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ''
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: etcd-dev@googlegroups.com
            name: etcd Community
          maturity: alpha
          provider:
            name: CNCF
          replaces: etcdoperator.v0.9.2
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.9.4

    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: etcdclusters.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdCluster
            listKind: EtcdClusterList
            plural: etcdclusters
            shortNames:
            - etcdclus
            - etcd
            singular: etcdcluster
          scope: Namespaced
          versions:
          - name: v1beta3
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
          - name: v1beta2
            served: true
            storage: false
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: etcdbackups.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdBackup
            listKind: EtcdBackupList
            plural: etcdbackups
            singular: etcdbackup
          scope: Namespaced
          versions:
          - name: v1beta2
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: etcdrestores.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdRestore
            listKind: EtcdRestoreList
            plural: etcdrestores
            singular: etcdrestore
          scope: Namespaced
          versions:
          - name: v1beta2
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
    packages: |
      - channels:
        - currentCSV: etcdoperator.v0.9.2
          name: alpha
        - currentCSV: etcdoperator.v0.9.4
          name: beta
        defaultChannel: alpha
        packageName: etcd-update
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmConfigmapEctdAlphaBetaYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmConfigmapEctdAlphaBetaYaml, nil
}

func testQeTestdataOlmConfigmapEctdAlphaBetaYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmConfigmapEctdAlphaBetaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/configmap-ectd-alpha-beta.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmConfigmapEtcdYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cm-bad-operator-template
objects:
- apiVersion: v1
  data:
    clusterServiceVersions: |
      - apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            alm-examples: "[\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdCluster\",\n    \"metadata\": {\n      \"name\": \"example\"\
              \n    },\n    \"spec\": {\n      \"size\": 3,\n      \"version\": \"3.2.13\"\
              \n    }\n  },\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdRestore\",\n    \"metadata\": {\n      \"name\": \"example-etcd-cluster-restore\"\
              \n    },\n    \"spec\": {\n      \"etcdCluster\": {\n        \"name\": \"example-etcd-cluster\"\
              \n      },\n      \"backupStorageType\": \"S3\",\n      \"s3\": {\n        \"\
              path\": \"<full-s3-path>\",\n        \"awsSecret\": \"<aws-secret>\"\n     \
              \ }\n    }\n  },\n  {\n    \"apiVersion\": \"etcd.database.coreos.com/v1beta2\"\
              ,\n    \"kind\": \"EtcdBackup\",\n    \"metadata\": {\n      \"name\": \"example-etcd-cluster-backup\"\
              \n    },\n    \"spec\": {\n      \"etcdEndpoints\": [\"<etcd-cluster-endpoints>\"\
              ],\n      \"storageType\":\"S3\",\n      \"s3\": {\n        \"path\": \"<full-s3-path>\"\
              ,\n        \"awsSecret\": \"<aws-secret>\"\n      }\n    }\n  }\n]\n"
            capabilities: Full Lifecycle
            categories: Database
            description: Creates and maintain highly-available etcd clusters on Kubernetes
            tectonic-visibility: ocs
          name: etcdoperator.v0.9.2
          namespace: placeholder
        spec:
          customresourcedefinitions:
            owned:
            - description: Represents a cluster of etcd nodes.
              displayName: etcd Cluster
              kind: EtcdCluster
              name: etcdclusters.etcd.database.coreos.com
              resources:
              - kind: Service
                version: v1
              - kind: Pod
                version: v1
              specDescriptors:
              - description: The desired number of member Pods for the etcd cluster.
                displayName: Size
                path: size
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podCount
              - description: Limits describes the minimum/maximum amount of compute resources
                  required/allowed
                displayName: Resource Requirements
                path: pod.resources
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:resourceRequirements
              statusDescriptors:
              - description: The status of each of the member Pods for the etcd cluster.
                displayName: Member Status
                path: members
                x-descriptors:
                - urn:alm:descriptor:com.tectonic.ui:podStatuses
              - description: The service at which the running etcd cluster can be accessed.
                displayName: Service
                path: serviceName
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Service
              - description: The current size of the etcd cluster.
                displayName: Cluster Size
                path: size
              - description: The current version of the etcd cluster.
                displayName: Current Version
                path: currentVersion
              - description: The target version of the etcd cluster, after upgrading.
                displayName: Target Version
                path: targetVersion
              - description: The current status of the etcd cluster.
                displayName: Status
                path: phase
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase
              - description: Explanation for the current status of the cluster.
                displayName: Status Details
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to backup an etcd cluster.
              displayName: etcd Backup
              kind: EtcdBackup
              name: etcdbackups.etcd.database.coreos.com
              specDescriptors:
              - description: Specifies the endpoints of an etcd cluster.
                displayName: etcd Endpoint(s)
                path: etcdEndpoints
                x-descriptors:
                - urn:alm:descriptor:etcd:endpoint
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the backup was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any backup related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
            - description: Represents the intent to restore an etcd cluster from a backup.
              displayName: etcd Restore
              kind: EtcdRestore
              name: etcdrestores.etcd.database.coreos.com
              specDescriptors:
              - description: References the EtcdCluster which should be restored,
                displayName: etcd Cluster
                path: etcdCluster.name
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:EtcdCluster
                - urn:alm:descriptor:text
              - description: The full AWS S3 path where the backup is saved.
                displayName: S3 Path
                path: s3.path
                x-descriptors:
                - urn:alm:descriptor:aws:s3:path
              - description: The name of the secret object that stores the AWS credential
                  and config files.
                displayName: AWS Secret
                path: s3.awsSecret
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret
              statusDescriptors:
              - description: Indicates if the restore was successful.
                displayName: Succeeded
                path: succeeded
                x-descriptors:
                - urn:alm:descriptor:text
              - description: Indicates the reason for any restore related failures.
                displayName: Reason
                path: reason
                x-descriptors:
                - urn:alm:descriptor:io.kubernetes.phase:reason
              version: v1beta2
          description: "The etcd Operater creates and maintains highly-available etcd clusters\
            \ on Kubernetes, allowing engineers to easily deploy and manage etcd clusters\
            \ for their applications.\n\netcd is a distributed key value store that provides\
            \ a reliable way to store data across a cluster of machines. It\xE2\u20AC\u2122\
            s open-source and available on GitHub. etcd gracefully handles leader elections\
            \ during network partitions and will tolerate machine failure, including the leader.\n\
            \n\n### Reading and writing to etcd\n\nCommunicate with etcd though its command\
            \ line utility ` + "`" + `etcdctl` + "`" + ` via port forwarding:\n\n    $ kubectl --namespace default\
            \ port-forward service/example-client 2379:2379\n    $ etcdctl --endpoints http://127.0.0.1:2379\
            \ get /\n\nOr directly to the API using the automatically generated Kubernetes\
            \ Service:\n\n    $ etcdctl --endpoints http://example-client.default.svc:2379\
            \ get /\n\nBe sure to secure your etcd cluster (see Common Configurations) before\
            \ exposing it outside of the namespace or cluster.\n\n\n### Supported Features\n\
            \n* **High availability** - Multiple instances of etcd are networked together\
            \ and secured. Individual failures or networking issues are transparently handled\
            \ to keep your cluster up and running.\n\n* **Automated updates** - Rolling out\
            \ a new etcd version works like all Kubernetes rolling updates. Simply declare\
            \ the desired version, and the etcd service starts a safe rolling update to the\
            \ new version automatically.\n\n* **Backups included** - Create etcd backups and\
            \ restore them through the etcd Operator.\n\n### Common Configurations\n\n* **Configure\
            \ TLS** - Specify [static TLS certs](https://github.com/coreos/etcd-operator/blob/master/doc/user/cluster_tls.md)\
            \ as Kubernetes secrets.\n\n* **Set Node Selector and Affinity** - [Spread your\
            \ etcd Pods](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#three-member-cluster-with-node-selector-and-anti-affinity-across-nodes)\
            \ across Nodes and availability zones.\n\n* **Set Resource Limits** - [Set the\
            \ Kubernetes limit and request](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#three-member-cluster-with-resource-requirement)\
            \ values for your etcd Pods.\n\n* **Customize Storage** - [Set a custom StorageClass](https://github.com/coreos/etcd-operator/blob/master/doc/user/spec_examples.md#custom-persistentvolumeclaim-definition)\
            \ that you would like to use.\n"
          displayName: etcd
          icon:
          - base64data: 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
            mediatype: image/png
          install:
            spec:
              deployments:
              - name: etcd-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: etcd-operator-alm-owned
                  template:
                    metadata:
                      labels:
                        name: etcd-operator-alm-owned
                      name: etcd-operator-alm-owned
                    spec:
                      containers:
                      - command:
                        - etcd-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-operator
                      - command:
                        - etcd-backup-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-backup-operator
                      - command:
                        - etcd-restore-operator
                        - --create-crd=false
                        env:
                        - name: MY_POD_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.namespace
                        - name: MY_POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        image: quay.io/coreos/etcd-operator@sha256:c0301e4686c3ed4206e370b42de5a3bd2229b9fb4906cf85f3f30650424abec2
                        name: etcd-restore-operator
                      serviceAccountName: etcd-operator
              permissions:
              - rules:
                - apiGroups:
                  - etcd.database.coreos.com
                  resources:
                  - etcdclusters
                  - etcdbackups
                  - etcdrestores
                  verbs:
                  - '*'
                - apiGroups:
                  - ''
                  resources:
                  - pods
                  - services
                  - endpoints
                  - persistentvolumeclaims
                  - events
                  verbs:
                  - '*'
                - apiGroups:
                  - apps
                  resources:
                  - deployments
                  verbs:
                  - '*'
                - apiGroups:
                  - ''
                  resources:
                  - secrets
                  verbs:
                  - get
                serviceAccountName: etcd-operator
            strategy: deployment
          installModes:
          - supported: true
            type: OwnNamespace
          - supported: true
            type: SingleNamespace
          - supported: false
            type: MultiNamespace
          - supported: false
            type: AllNamespaces
          keywords:
          - etcd
          - key value
          - database
          - coreos
          - open source
          labels:
            alm-owner-etcd: etcdoperator
            operated-by: etcdoperator
          links:
          - name: Blog
            url: https://coreos.com/etcd
          - name: Documentation
            url: https://coreos.com/operators/etcd/docs/latest/
          - name: etcd Operator Source Code
            url: https://github.com/coreos/etcd-operator
          maintainers:
          - email: etcd-dev@googlegroups.com
            name: etcd Community
          maturity: alpha
          provider:
            name: CNCF
          selector:
            matchLabels:
              alm-owner-etcd: etcdoperator
              operated-by: etcdoperator
          version: 0.9.2
    customResourceDefinitions: |
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: etcdclusters.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdCluster
            listKind: EtcdClusterList
            plural: etcdclusters
            shortNames:
            - etcdclus
            - etcd
            singular: etcdcluster
          scope: Namespaced
          versions:
          - name: v1beta2
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: etcdbackups.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdBackup
            listKind: EtcdBackupList
            plural: etcdbackups
            singular: etcdbackup
          scope: Namespaced
          versions:
          - name: v1beta2
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
      - apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
          name: etcdrestores.etcd.database.coreos.com
        spec:
          group: etcd.database.coreos.com
          names:
            kind: EtcdRestore
            listKind: EtcdRestoreList
            plural: etcdrestores
            singular: etcdrestore
          scope: Namespaced
          versions:
          - name: v1beta2
            served: true
            storage: true
            schema:
              openAPIV3Schema:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  metadata:
                    type: object
                  spec:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  status:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
    packages: |
      - channels:
        - currentCSV: etcdoperator.v0.9.2
          name: alpha
        defaultChannel: "alpha"
        packageName: etcd-update
  kind: ConfigMap
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmConfigmapEtcdYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmConfigmapEtcdYaml, nil
}

func testQeTestdataOlmConfigmapEtcdYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmConfigmapEtcdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/configmap-etcd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmConfigmapTestYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: default
`)

func testQeTestdataOlmConfigmapTestYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmConfigmapTestYaml, nil
}

func testQeTestdataOlmConfigmapTestYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmConfigmapTestYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/configmap-test.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmConfigmapWithDefaultchannelYaml = []byte(`---
kind: ConfigMap
apiVersion: v1
metadata:
  name: scenario3
  namespace: scenario3
data:
  customResourceDefinitions: |-
    - apiVersion: apiextensions.k8s.io/v1beta1
      kind: CustomResourceDefinition
      metadata:
        name: exampleas.examples.io
      spec:
        group: examples.io
        names:
          kind: ExampleA
          listKind: ExampleAList
          plural: exampleas
          singular: examplea
        scope: Namespaced
        subresources:
          status: {}
        version: v1alpha1
        versions:
        - name: v1alpha1
          served: true
          storage: true
    - apiVersion: apiextensions.k8s.io/v1beta1
      kind: CustomResourceDefinition
      metadata:
        name: examplebs.examples.io
      spec:
        group: examples.io
        names:
          kind: ExampleB
          listKind: ExampleBList
          plural: examplebs
          singular: exampleb
        scope: Namespaced
        subresources:
          status: {}
        version: v1alpha1
        versions:
        - name: v1alpha1
          served: true
          storage: true
  clusterServiceVersions: |-
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-a.v0.0.1
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleA
            name: exampleas.examples.io
            version: v1alpha1
            displayName: Example A
            description: Example A Custom Resource Definition
          required:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator A
        description: An example operator (A)
        provider:
          name: Example Provider A
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-a
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-a
            deployments:
            - name: example-operator-a
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-a
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-a
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-a:v1
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-a
                      image: docker.io/djzager/example-operator-a:v1
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-a
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: alpha
        version: 0.0.1
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-a.v1.0.0
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleA
            name: exampleas.examples.io
            version: v1alpha1
            displayName: Example A
            description: Example A Custom Resource Definition
          required:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator A
        description: An example operator (A)
        provider:
          name: Example Provider A
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-a
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-a
            deployments:
            - name: example-operator-a
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-a
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-a
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-a:v1stable
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-a
                      image: docker.io/djzager/example-operator-a:v1stable
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-a
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: stable
        version: 1.0.0
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-b.v0.0.1
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator B
        description: An example operator (B)
        provider:
          name: Example Provider B
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-b
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-b
            deployments:
            - name: example-operator-b
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-b
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-b
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-b:v1
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-b
                      image: docker.io/djzager/example-operator-b:v1
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-b
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: alpha
        version: 0.0.1
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-b.v1.0.0
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator B
        description: An example operator (B)
        provider:
          name: Example Provider B
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-b
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-b
            deployments:
            - name: example-operator-b
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-b
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-b
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-b:v1stable
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-b
                      image: docker.io/djzager/example-operator-b:v1stable
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-b
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: stable
        version: 1.0.0
  packages: |-
    - packageName: example-operator-a
      defaultChannel: alpha
      channels:
      - name: alpha
        currentCSV: example-operator-a.v0.0.1
      - name: stable
        currentCSV: example-operator-a.v1.0.0
    - packageName: example-operator-b
      defaultChannel: alpha
      channels:
      - name: alpha
        currentCSV: example-operator-b.v0.0.1
      - name: stable
        currentCSV: example-operator-b.v1.0.0
`)

func testQeTestdataOlmConfigmapWithDefaultchannelYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmConfigmapWithDefaultchannelYaml, nil
}

func testQeTestdataOlmConfigmapWithDefaultchannelYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmConfigmapWithDefaultchannelYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/configmap-with-defaultchannel.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmConfigmapWithoutDefaultchannelYaml = []byte(`---
kind: ConfigMap
apiVersion: v1
metadata:
  name: scenario3
  namespace: scenario3
data:
  customResourceDefinitions: |-
    - apiVersion: apiextensions.k8s.io/v1beta1
      kind: CustomResourceDefinition
      metadata:
        name: exampleas.examples.io
      spec:
        group: examples.io
        names:
          kind: ExampleA
          listKind: ExampleAList
          plural: exampleas
          singular: examplea
        scope: Namespaced
        subresources:
          status: {}
        version: v1alpha1
        versions:
        - name: v1alpha1
          served: true
          storage: true
    - apiVersion: apiextensions.k8s.io/v1beta1
      kind: CustomResourceDefinition
      metadata:
        name: examplebs.examples.io
      spec:
        group: examples.io
        names:
          kind: ExampleB
          listKind: ExampleBList
          plural: examplebs
          singular: exampleb
        scope: Namespaced
        subresources:
          status: {}
        version: v1alpha1
        versions:
        - name: v1alpha1
          served: true
          storage: true
  clusterServiceVersions: |-
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-a.v0.0.1
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleA
            name: exampleas.examples.io
            version: v1alpha1
            displayName: Example A
            description: Example A Custom Resource Definition
          required:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator A
        description: An example operator (A)
        provider:
          name: Example Provider A
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-a
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-a
            deployments:
            - name: example-operator-a
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-a
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-a
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-a:v1
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-a
                      image: docker.io/djzager/example-operator-a:v1
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-a
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: alpha
        version: 0.0.1
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-a.v1.0.0
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleA
            name: exampleas.examples.io
            version: v1alpha1
            displayName: Example A
            description: Example A Custom Resource Definition
          required:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator A
        description: An example operator (A)
        provider:
          name: Example Provider A
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-a
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-a
            deployments:
            - name: example-operator-a
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-a
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-a
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-a:v1stable
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-a
                      image: docker.io/djzager/example-operator-a:v1stable
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-a
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: stable
        version: 1.0.0
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-b.v0.0.1
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator B
        description: An example operator (B)
        provider:
          name: Example Provider B
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-b
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-b
            deployments:
            - name: example-operator-b
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-b
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-b
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-b:v1
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-b
                      image: docker.io/djzager/example-operator-b:v1
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-b
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: alpha
        version: 0.0.1
    - apiVersion: operators.coreos.com/v1alpha1
      kind: ClusterServiceVersion
      metadata:
        annotations:
          capabilities: Basic Install
        name: example-operator-b.v1.0.0
        namespace: placeholder
      spec:
        apiservicedefinitions: {}
        customresourcedefinitions:
          owned:
          - kind: ExampleB
            name: examplebs.examples.io
            version: v1alpha1
            displayName: Example B
            description: Example B Custom Resource Definition
        displayName: Example Operator B
        description: An example operator (B)
        provider:
          name: Example Provider B
        links:
          - name: Source Code
            url: https://github.com/djzager/olm-playground
        keywords:
          - foo
          - bar
          - baz
        install:
          spec:
            clusterPermissions:
            - rules:
              - apiGroups:
                - ""
                resources:
                - pods
                - services
                - endpoints
                - persistentvolumeclaims
                - events
                - configmaps
                - secrets
                verbs:
                - '*'
              - apiGroups:
                - ""
                resources:
                - namespaces
                verbs:
                - get
              - apiGroups:
                - apps
                resources:
                - deployments
                - daemonsets
                - replicasets
                - statefulsets
                verbs:
                - '*'
              - apiGroups:
                - monitoring.coreos.com
                resources:
                - servicemonitors
                verbs:
                - get
                - create
              - apiGroups:
                - apps
                resourceNames:
                - example-operator-b
                resources:
                - deployments/finalizers
                verbs:
                - update
              - apiGroups:
                - examples.io
                resources:
                - '*'
                verbs:
                - '*'
              serviceAccountName: example-operator-b
            deployments:
            - name: example-operator-b
              spec:
                replicas: 1
                selector:
                  matchLabels:
                    name: example-operator-b
                strategy: {}
                template:
                  metadata:
                    labels:
                      name: example-operator-b
                  spec:
                    containers:
                    - command:
                      - /usr/local/bin/ao-logs
                      - /tmp/ansible-operator/runner
                      - stdout
                      image: docker.io/djzager/example-operator-b:v1stable
                      imagePullPolicy: IfNotPresent
                      name: ansible
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                        readOnly: true
                    - env:
                      - name: WATCH_NAMESPACE
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: example-operator-b
                      image: docker.io/djzager/example-operator-b:v1stable
                      imagePullPolicy: IfNotPresent
                      name: operator
                      resources: {}
                      volumeMounts:
                      - mountPath: /tmp/ansible-operator/runner
                        name: runner
                    serviceAccountName: example-operator-b
                    volumes:
                    - emptyDir: {}
                      name: runner
          strategy: deployment
        installModes:
        - supported: true
          type: OwnNamespace
        - supported: true
          type: SingleNamespace
        - supported: false
          type: MultiNamespace
        - supported: true
          type: AllNamespaces
        maturity: stable
        version: 1.0.0
  packages: |-
    - packageName: example-operator-a
      channels:
      - name: alpha
        currentCSV: example-operator-a.v0.0.1
      - name: stable
        currentCSV: example-operator-a.v1.0.0
    - packageName: example-operator-b
      channels:
      - name: alpha
        currentCSV: example-operator-b.v0.0.1
      - name: stable
        currentCSV: example-operator-b.v1.0.0
`)

func testQeTestdataOlmConfigmapWithoutDefaultchannelYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmConfigmapWithoutDefaultchannelYaml, nil
}

func testQeTestdataOlmConfigmapWithoutDefaultchannelYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmConfigmapWithoutDefaultchannelYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/configmap-without-defaultchannel.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCrWebhooktestYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: WebhookTest-template
objects:
- apiVersion: webhook.operators.coreos.io/v1
  kind: WebhookTest
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    valid: ${{VALID}}
parameters:
- name: NAME
- name: NAMESPACE
- name: VALID
`)

func testQeTestdataOlmCrWebhooktestYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCrWebhooktestYaml, nil
}

func testQeTestdataOlmCrWebhooktestYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCrWebhooktestYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cr-webhookTest.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCr_devworkspaceYaml = []byte(`kind: DevWorkspace
apiVersion: workspace.devfile.io/v1alpha2
metadata:
  name: empty-devworkspace
spec:
  started: true
  template: {}

`)

func testQeTestdataOlmCr_devworkspaceYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCr_devworkspaceYaml, nil
}

func testQeTestdataOlmCr_devworkspaceYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCr_devworkspaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cr_devworkspace.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCr_pgadminYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: PGAdmin-template
objects:
- apiVersion: postgres-operator.crunchydata.com/v1beta1
  kind: PGAdmin
  metadata:
    name: pgadmin-example
    namespace: ${NAMESPACE}
  spec:
    dataVolumeClaimSpec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
    serverGroups:
    - name: Crunchy Postgres for Kubernetes
      postgresClusterSelector: {}
    tolerations:
    - tolerationSeconds: 1726856593000774400

parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmCr_pgadminYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCr_pgadminYaml, nil
}

func testQeTestdataOlmCr_pgadminYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCr_pgadminYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cr_pgadmin.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCsImageTemplateYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cs-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    annotations:
      olm.catalogImageTemplate: "${IMAGETEMPLATE}"
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: IMAGETEMPLATE
  value: "quay.io/kube-release-v{kube_major_version}/catalog:v{kube_major_version}"
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"

`)

func testQeTestdataOlmCsImageTemplateYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCsImageTemplateYaml, nil
}

func testQeTestdataOlmCsImageTemplateYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCsImageTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cs-image-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCsWithoutImageYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: catalogsource-image-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    displayName: "${DISPLAYNAME}"
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: 10m0s
parameters:
- name: NAME
- name: NAMESPACE
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
`)

func testQeTestdataOlmCsWithoutImageYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCsWithoutImageYaml, nil
}

func testQeTestdataOlmCsWithoutImageYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCsWithoutImageYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cs-without-image.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCsWithoutIntervalYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cs-wihtout-interval
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    annotations:
      olm.catalogImageTemplate: "${IMAGETEMPLATE}"
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
parameters:
- name: IMAGETEMPLATE
  value: "quay.io/kube-release-v{kube_major_version}/catalog:v{kube_major_version}"
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
`)

func testQeTestdataOlmCsWithoutIntervalYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCsWithoutIntervalYaml, nil
}

func testQeTestdataOlmCsWithoutIntervalYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCsWithoutIntervalYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cs-without-interval.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCsWithoutSccYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cs-without-scc
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    annotations:
      olm.catalogImageTemplate: "${IMAGETEMPLATE}"
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    image: "${ADDRESS}"
    secrets:
    - "${SECRET}"  
    displayName: "${DISPLAYNAME}"
    publisher: "${PUBLISHER}"
    sourceType: "${SOURCETYPE}"
    updateStrategy:
      registryPoll:
        interval: "${INTERVAL}"
parameters:
- name: IMAGETEMPLATE
  value: "quay.io/kube-release-v{kube_major_version}/catalog:v{kube_major_version}"
- name: NAME
- name: NAMESPACE
- name: ADDRESS
- name: DISPLAYNAME
- name: PUBLISHER
- name: SOURCETYPE
- name: SECRET
- name: INTERVAL
  value: "10m0s"

`)

func testQeTestdataOlmCsWithoutSccYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCsWithoutSccYaml, nil
}

func testQeTestdataOlmCsWithoutSccYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCsWithoutSccYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/cs-without-scc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmCscYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: opsrc-template
objects:
- kind: CatalogSourceConfig
  apiVersion: operators.coreos.com/v1
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    packages: "${PACKAGES}"
    targetNamespace: "${TARGETNAMESPACE}"
    source: "${SOURCE}"
parameters:
- name: NAME
- name: NAMESPACE
- name: PACKAGES
- name: TARGETNAMESPACE
- name: SOURCE
`)

func testQeTestdataOlmCscYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmCscYaml, nil
}

func testQeTestdataOlmCscYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmCscYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/csc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmEnvSubscriptionYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: sub-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: "${SUBNAME}"
    namespace: "${SUBNAMESPACE}"
  spec:
    channel: "${CHANNEL}"
    installPlanApproval: "${APPROVAL}"
    name: "${OPERATORNAME}"
    source: "${SOURCENAME}"
    sourceNamespace: "${SOURCENAMESPACE}"
    startingCSV: "${STARTINGCSV}"
    config:
      env:
      - name: ISO_IMAGE_TYPE
        value: "minimal-iso"
      - name: OPENSHIFT_VERSIONS
        value: '{"4.6":{"display_name":"4.6.16","release_image":"quay.io/openshift-release-dev/ocp-release:4.6.16-x86_64","rhcos_image":"https://mirror.openshift.com/pub/openshift-v4/dependencies/rhcos/4.6/4.6.8/rhcos-4.6.8-x86_64-live.x86_64.iso","rhcos_version":"46.82.202012051820-0","support_level":"production"},"4.7":{"display_name":"4.7.2","release_image":"quay.io/openshift-release-dev/ocp-release:4.7.2-x86_64","rhcos_image":"https://mirror.openshift.com/pub/openshift-v4/dependencies/rhcos/4.7/4.7.0/rhcos-4.7.0-x86_64-live.x86_64.iso","rhcos_version":"47.83.202102090044-0","support_level":"production"},"4.8":{"display_name":"4.8","release_image":"registry.ci.openshift.org/ocp/release:4.8.0-0.nightly-2021-04-09-140229","rhcos_image":"https://mirror.openshift.com/pub/openshift-v4/dependencies/rhcos/4.7/4.7.0/rhcos-4.7.0-x86_64-live.x86_64.iso","rhcos_version":"47.83.202102090044-0","support_level":"production"}}'
      - name: OPERATOR_CONDITION_NAME
        value: etcdoperator.v0.9.5
      - name: MY_POD_NAMESPACE
        value: default
parameters:
- name: SUBNAME
- name: SUBNAMESPACE
- name: CHANNEL
- name: APPROVAL
  value: "Automatic"
- name: OPERATORNAME
- name: SOURCENAME
- name: SOURCENAMESPACE
  value: "openshift-marketplace"
- name: STARTINGCSV
  value: ""
`)

func testQeTestdataOlmEnvSubscriptionYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmEnvSubscriptionYaml, nil
}

func testQeTestdataOlmEnvSubscriptionYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmEnvSubscriptionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/env-subscription.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmEnvfromSubscriptionYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: sub-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: "${SUBNAME}"
    namespace: "${SUBNAMESPACE}"
  spec:
    channel: "${CHANNEL}"
    installPlanApproval: "${APPROVAL}"
    name: "${OPERATORNAME}"
    source: "${SOURCENAME}"
    sourceNamespace: "${SOURCENAMESPACE}"
    startingCSV: "${STARTINGCSV}"
    config:
      envFrom:
      - configMapRef:
          name: "${CONFIGMAPREF}"
      - secretRef:
          name: "${SECRETREF}"
parameters:
- name: SUBNAME
- name: SUBNAMESPACE
- name: CHANNEL
- name: APPROVAL
  value: "Automatic"
- name: OPERATORNAME
- name: SOURCENAME
- name: SOURCENAMESPACE
  value: "openshift-marketplace"
- name: STARTINGCSV
  value: ""
- name: CONFIGMAPREF
- name: SECRETREF
`)

func testQeTestdataOlmEnvfromSubscriptionYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmEnvfromSubscriptionYaml, nil
}

func testQeTestdataOlmEnvfromSubscriptionYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmEnvfromSubscriptionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/envfrom-subscription.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmEtcdClusterYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: etcdCluster-template
objects:
- apiVersion: etcd.database.coreos.com/v1beta2
  kind: EtcdCluster
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    size: 3
    version: 3.2.13
parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmEtcdClusterYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmEtcdClusterYaml, nil
}

func testQeTestdataOlmEtcdClusterYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmEtcdClusterYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/etcd-cluster.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmEtcdSubscriptionManualYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: subscription-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    channel: "${CHANNEL}"
    installPlanApproval: "${APPROVAL}"
    name: etcd
    source: "${SOURCENAME}"
    sourceNamespace: "${SOURCENAMESPACE}"
    startingCSV: "${STARTINGCSV}"
parameters:
- name: NAME
- name: NAMESPACE
- name: SOURCENAME
- name: SOURCENAMESPACE
- name: CHANNEL
  value: "singlenamespace-alpha"
- name: STARTINGCSV
  value: "etcdoperator.v0.9.4"
- name: APPROVAL
  value: "Manual"
`)

func testQeTestdataOlmEtcdSubscriptionManualYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmEtcdSubscriptionManualYaml, nil
}

func testQeTestdataOlmEtcdSubscriptionManualYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmEtcdSubscriptionManualYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/etcd-subscription-manual.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmEtcdSubscriptionYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: subscription-template
objects:
  - apiVersion: operators.coreos.com/v1alpha1
    kind: Subscription
    metadata:
      name: "${NAME}"
      namespace: "${NAMESPACE}"
    spec:
      channel: "${CHANNEL}"
      installPlanApproval: "${APPROVAL}"
      name: etcd
      source: "${SOURCENAME}"
      sourceNamespace: "${SOURCENAMESPACE}"
      startingCSV: "${STARTINGCSV}"
parameters:
  - name: NAME
  - name: NAMESPACE
  - name: SOURCENAME
  - name: SOURCENAMESPACE
  - name: CHANNEL
    value: "singlenamespace-alpha"
  - name: STARTINGCSV
    value: "etcdoperator.v0.9.4"
  - name: APPROVAL
    value: "Automatic"
`)

func testQeTestdataOlmEtcdSubscriptionYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmEtcdSubscriptionYaml, nil
}

func testQeTestdataOlmEtcdSubscriptionYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmEtcdSubscriptionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/etcd-subscription.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmMcWorkloadPartitionYaml = []byte(`apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: master
  name: 02-master-workload-partitioning
spec:
  config:
    ignition:
      version: 3.2.0
    storage:
      files:
      - contents:
          source: data:text/plain;charset=utf-8;base64,W2NyaW8ucnVudGltZS53b3JrbG9hZHMubWFuYWdlbWVudF0KYWN0aXZhdGlvbl9hbm5vdGF0aW9uID0gInRhcmdldC53b3JrbG9hZC5vcGVuc2hpZnQuaW8vbWFuYWdlbWVudCIKYW5ub3RhdGlvbl9wcmVmaXggPSAicmVzb3VyY2VzLndvcmtsb2FkLm9wZW5zaGlmdC5pbyIKW2NyaW8ucnVudGltZS53b3JrbG9hZHMubWFuYWdlbWVudC5yZXNvdXJjZXNdCmNwdXNoYXJlcyA9IDAKQ1BVcyA9ICIwLTEsIDUyLTUzIgo=
        mode: 420
        overwrite: true
        path: /etc/crio/crio.conf.d/01-workload-partitioning
        user:
          name: root
      - contents:
          source: data:text/plain;charset=utf-8;base64,ewogICJtYW5hZ2VtZW50IjogewogICAgImNwdXNldCI6ICIwLTEsNTItNTMiCiAgfQp9Cg==
        mode: 420
        overwrite: true
        path: /etc/kubernetes/openshift-workload-pinning
        user:
          name: root
`)

func testQeTestdataOlmMcWorkloadPartitionYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmMcWorkloadPartitionYaml, nil
}

func testQeTestdataOlmMcWorkloadPartitionYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmMcWorkloadPartitionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/mc-workload-partition.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOgAllnsYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: operatorgroup-allns-template
objects:
- kind: OperatorGroup
  apiVersion: operators.coreos.com/v1
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
parameters:
- name: NAME
- name: NAMESPACE

`)

func testQeTestdataOlmOgAllnsYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOgAllnsYaml, nil
}

func testQeTestdataOlmOgAllnsYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOgAllnsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/og-allns.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOgMultinsYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: operatorgroup-multins-template
objects:
- kind: OperatorGroup
  apiVersion: operators.coreos.com/v1
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    selector:
      matchLabels:
        env: "${MULTINSLABEL}"
parameters:
- name: NAME
- name: NAMESPACE
- name: MULTINSLABEL
`)

func testQeTestdataOlmOgMultinsYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOgMultinsYaml, nil
}

func testQeTestdataOlmOgMultinsYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOgMultinsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/og-multins.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOlmProxySubscriptionYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: sub-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: "${SUBNAME}"
    namespace: "${SUBNAMESPACE}"
  spec:
    config:
      env:
      - name: HTTP_PROXY
        value: ${SUBHTTPPROXY}
      - name: HTTPS_PROXY
        value: ${SUBHTTPSPROXY}
      - name: NO_PROXY
        value: ${SUBNOPROXY}
    channel: "${CHANNEL}"
    installPlanApproval: "${APPROVAL}"
    name: "${OPERATORNAME}"
    source: "${SOURCENAME}"
    sourceNamespace: "${SOURCENAMESPACE}"
    startingCSV: "${STARTINGCSV}"
parameters:
- name: SUBNAME
- name: SUBNAMESPACE
- name: CHANNEL
- name: APPROVAL
  value: "Automatic"
- name: OPERATORNAME
- name: SOURCENAME
- name: SOURCENAMESPACE
  value: "openshift-marketplace"
- name: STARTINGCSV
  value: ""
- name: SUBHTTPPROXY
  value: ""
- name: SUBHTTPSPROXY
  value: ""
- name: SUBNOPROXY
  value: ""
`)

func testQeTestdataOlmOlmProxySubscriptionYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOlmProxySubscriptionYaml, nil
}

func testQeTestdataOlmOlmProxySubscriptionYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOlmProxySubscriptionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/olm-proxy-subscription.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOlmSubscriptionYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: sub-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: "${SUBNAME}"
    namespace: "${SUBNAMESPACE}"
  spec:
    channel: "${CHANNEL}"
    installPlanApproval: "${APPROVAL}"
    name: "${OPERATORNAME}"
    source: "${SOURCENAME}"
    sourceNamespace: "${SOURCENAMESPACE}"
    startingCSV: "${STARTINGCSV}"
parameters:
- name: SUBNAME
- name: SUBNAMESPACE
- name: CHANNEL
- name: APPROVAL
  value: "Automatic"
- name: OPERATORNAME
- name: SOURCENAME
- name: SOURCENAMESPACE
  value: "openshift-marketplace"
- name: STARTINGCSV
  value: ""
`)

func testQeTestdataOlmOlmSubscriptionYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOlmSubscriptionYaml, nil
}

func testQeTestdataOlmOlmSubscriptionYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOlmSubscriptionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/olm-subscription.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOperatorYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: operator-template
objects:
- apiVersion: operators.operatorframework.io/v1alpha1
  kind: Operator
  metadata:
    name: "${NAME}"
  spec:
    packageName: "${PACKAGE}"
    channel: "${CHANNEL}"
    version: "${VERSION}"
parameters:
- name: NAME
- name: PACKAGE
  value: "quay-operator"
- name: CHANNEL
  value: "stable-3.8"
- name: VERSION
  value: "3.8.12"
`)

func testQeTestdataOlmOperatorYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOperatorYaml, nil
}

func testQeTestdataOlmOperatorYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOperatorYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/operator.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOperatorgroupServiceaccountYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: operatorgroup-template
objects:
  - kind: OperatorGroup
    apiVersion: operators.coreos.com/v1
    metadata:
      name: "${NAME}"
      namespace: "${NAMESPACE}"
    spec:
      serviceAccountName: "${SERVICE_ACCOUNT_NAME}"
      targetNamespaces:
        - "${NAMESPACE}"
parameters:
  - name: NAME
  - name: NAMESPACE
  - name: SERVICE_ACCOUNT_NAME
`)

func testQeTestdataOlmOperatorgroupServiceaccountYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOperatorgroupServiceaccountYaml, nil
}

func testQeTestdataOlmOperatorgroupServiceaccountYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOperatorgroupServiceaccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/operatorgroup-serviceaccount.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOperatorgroupUpgradestrategyYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: operatorgroup-upgradestrategy-template
objects:
- kind: OperatorGroup
  apiVersion: operators.coreos.com/v1
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    upgradeStrategy: "${UPGRADESTRATEGY}"
    targetNamespaces:
    - "${NAMESPACE}"

parameters:
- name: NAME
- name: NAMESPACE
- name: UPGRADESTRATEGY

`)

func testQeTestdataOlmOperatorgroupUpgradestrategyYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOperatorgroupUpgradestrategyYaml, nil
}

func testQeTestdataOlmOperatorgroupUpgradestrategyYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOperatorgroupUpgradestrategyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/operatorgroup-upgradestrategy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOperatorgroupYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: operatorgroup-template
objects:
- kind: OperatorGroup
  apiVersion: operators.coreos.com/v1
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    targetNamespaces:
    - "${NAMESPACE}"

parameters:
- name: NAME
- name: NAMESPACE
  
`)

func testQeTestdataOlmOperatorgroupYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOperatorgroupYaml, nil
}

func testQeTestdataOlmOperatorgroupYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOperatorgroupYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/operatorgroup.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmOpsrcYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: opsrc-template
objects:
- kind: OperatorSource
  apiVersion: operators.coreos.com/v1
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
    labels:
      opsrc-provider: "${NAMELABEL}"
  spec:
    type: appregistry
    endpoint: "https://quay.io/cnr"
    registryNamespace: "${REGISTRYNAMESPACE}"
    displayName: "${DISPLAYNAME}"
    publisher: "${PUBLISHER}"
parameters:
- name: NAME
- name: NAMESPACE
- name: NAMELABEL
- name: REGISTRYNAMESPACE
- name: DISPLAYNAME
- name: PUBLISHER
`)

func testQeTestdataOlmOpsrcYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmOpsrcYaml, nil
}

func testQeTestdataOlmOpsrcYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmOpsrcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/opsrc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmPackageserverYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: packageserver-csv-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: ClusterServiceVersion
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  spec:
    apiservicedefinitions:
      owned:
      - containerPort: 5443
        deploymentName: packageserver
        description: A PackageManifest is a resource generated from existing CatalogSources
          and their ConfigMaps
        displayName: PackageManifest
        group: packages.operators.coreos.com
        kind: PackageManifest
        name: packagemanifests
        version: v1
    customresourcedefinitions: {}
    description: Represents an Operator package that is available from a given CatalogSource
      which will resolve to a ClusterServiceVersion.
    displayName: Package Server
    install:
      spec:
        clusterPermissions:
        - rules:
          - apiGroups:
            - authorization.k8s.io
            resources:
            - subjectaccessreviews
            verbs:
            - create
            - get
          - apiGroups:
            - ""
            resources:
            - configmaps
            verbs:
            - get
            - list
            - watch
          - apiGroups:
            - operators.coreos.com
            resources:
            - catalogsources
            verbs:
            - get
            - list
            - watch
          - apiGroups:
            - packages.operators.coreos.com
            resources:
            - packagemanifests
            verbs:
            - get
            - list
          serviceAccountName: olm-operator-serviceaccount
        deployments:
        - name: packageserver
          spec:
            replicas: 1
            selector:
              matchLabels:
                app: packageserver
            strategy:
              type: RollingUpdate
            template:
              metadata:
                labels:
                  app: packageserver
              spec:
                containers:
                - command:
                  - /bin/package-server
                  - -v=4
                  - --secure-port
                  - "5443"
                  - --global-namespace
                  - olm
                  - --debug
                  image: quay.io/operator-framework/olm:local
                  imagePullPolicy: IfNotPresent
                  livenessProbe:
                    httpGet:
                      path: /healthz
                      port: 5443
                      scheme: HTTPS
                  name: packageserver
                  ports:
                  - containerPort: 5443
                    protocol: TCP
                  readinessProbe:
                    httpGet:
                      path: /healthz
                      port: 5443
                      scheme: HTTPS
                  resources:
                    requests:
                      cpu: 10m
                      memory: 50Mi
                  terminationMessagePolicy: FallbackToLogsOnError
                  volumeMounts:
                  - mountPath: /tmp
                    name: tmpfs
                nodeSelector:
                  kubernetes.io/os: linux
                serviceAccountName: olm-operator-serviceaccount
                tolerations:
                - effect: NoSchedule
                  key: node-role.kubernetes.io/master
                  operator: Exists
                - effect: NoExecute
                  key: node.kubernetes.io/unreachable
                  operator: Exists
                  tolerationSeconds: 120
                - effect: NoExecute
                  key: node.kubernetes.io/not-ready
                  operator: Exists
                  tolerationSeconds: 120
                volumes:
                - emptyDir: {}
                  name: tmpfs
      strategy: deployment
    installModes:
    - supported: true
      type: OwnNamespace
    - supported: true
      type: SingleNamespace
    - supported: true
      type: MultiNamespace
    - supported: true
      type: AllNamespaces
    keywords:
    - packagemanifests
    - olm
    - packages
    links:
    - name: Package Server
      url: https://github.com/operator-framework/operator-lifecycle-manager/tree/master/pkg/package-server
    maintainers:
    - email: openshift-operators@redhat.com
      name: Red Hat
    maturity: alpha
    minKubeVersion: 1.11.0
    provider:
      name: Red Hat
    replaces: packageserver
    version: 1.0.0
parameters:
- name: NAMESPACE
- name: NAME
`)

func testQeTestdataOlmPackageserverYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmPackageserverYaml, nil
}

func testQeTestdataOlmPackageserverYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmPackageserverYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/packageserver.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmPlatform_operatorYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: platform-operator-template
objects:
- apiVersion: platform.openshift.io/v1alpha1
  kind: PlatformOperator
  metadata:
    name: "${NAME}"
  spec:
    package:
      name: "${PACKAGE}"
parameters:
- name: NAME
- name: PACKAGE
`)

func testQeTestdataOlmPlatform_operatorYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmPlatform_operatorYaml, nil
}

func testQeTestdataOlmPlatform_operatorYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmPlatform_operatorYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/platform_operator.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmPrometheusAntiaffinityYaml = []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: example
spec:
  evaluationInterval: 30s
  serviceMonitorSelector: {}
  alerting:
    alertmanagers:
      - namespace: monitoring
        name: alertmanager-main
        port: web
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
              - values:
                  - dev
                key: app_54038
                operator: NotIn
  probeSelector: {}
  podMonitorSelector: {}
  scrapeInterval: 30s
  ruleSelector: {}
  replicas: 2
  serviceAccountName: prometheus-k8s
`)

func testQeTestdataOlmPrometheusAntiaffinityYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmPrometheusAntiaffinityYaml, nil
}

func testQeTestdataOlmPrometheusAntiaffinityYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmPrometheusAntiaffinityYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/prometheus-antiaffinity.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmPrometheusNodeaffinityYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: prometheus-nodeaffinity-template
objects:
  - apiVersion: monitoring.coreos.com/v1
    kind: Prometheus
    metadata:
      name: example
      namespace: "${NAMESPACE}"
    spec:
      evaluationInterval: 30s
      serviceMonitorSelector: {}
      alerting:
        alertmanagers:
          - namespace: monitoring
            name: alertmanager-main
            port: web
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - values:
                      - "${NODE_NAME}"
                    key: kubernetes.io/hostname
                    operator: In
      probeSelector: {}
      podMonitorSelector: {}
      scrapeInterval: 30s
      ruleSelector: {}
      replicas: 2
      serviceAccountName: prometheus-k8s
parameters:
  - name: NODE_NAME
  - name: NAMESPACE
`)

func testQeTestdataOlmPrometheusNodeaffinityYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmPrometheusNodeaffinityYaml, nil
}

func testQeTestdataOlmPrometheusNodeaffinityYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmPrometheusNodeaffinityYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/prometheus-nodeaffinity.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmRoleBindingYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: rolebinding-template
objects:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: RoleBinding
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: Role
    name: "${ROLE_NAME}"
  subjects:
  - kind: ServiceAccount
    name: "${SA_NAME}"
    namespace: "${NAMESPACE}"

parameters:
- name: NAME
- name: NAMESPACE
- name: SA_NAME
- name: ROLE_NAME
`)

func testQeTestdataOlmRoleBindingYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmRoleBindingYaml, nil
}

func testQeTestdataOlmRoleBindingYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/role-binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmRoleYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: role-template
objects:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  rules:
    - apiGroups: ["*"]
      resources: ["*"]
      verbs: ["*"]
parameters:
- name: NAME
- name: NAMESPACE

`)

func testQeTestdataOlmRoleYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmRoleYaml, nil
}

func testQeTestdataOlmRoleYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmSccYaml = []byte(`allowHostDirVolumePlugin: true
allowHostIPC: false
allowHostNetwork: true
allowHostPID: true
allowHostPorts: true
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
- SYS_ADMIN
- SYS_RESOURCE
- SYS_PTRACE
- NET_ADMIN
- NET_BROADCAST
- NET_RAW
- IPC_LOCK
- CHOWN
- AUDIT_CONTROL
- AUDIT_READ
- DAC_READ_SEARCH
apiVersion: security.openshift.io/v1
defaultAddCapabilities: []
fsGroup:
  type: MustRunAs
groups: []
kind: SecurityContextConstraints
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"allowHostDirVolumePlugin":true,"allowHostIPC":false,"allowHostNetwork":true,"allowHostPID":true,"allowHostPorts":true,"allowPrivilegeEscalation":true,"allowPrivilegedContainer":false,"allowedCapabilities":["SYS_ADMIN","SYS_RESOURCE","SYS_PTRACE","NET_ADMIN","NET_BROADCAST","NET_RAW","IPC_LOCK","CHOWN","AUDIT_CONTROL","AUDIT_READ","DAC_READ_SEARCH"],"apiVersion":"security.openshift.io/v1","defaultAddCapabilities":[],"fsGroup":{"type":"MustRunAs"},"groups":[],"kind":"SecurityContextConstraints","metadata":{"annotations":{},"creationTimestamp":"2021-10-23T21:34:21Z","generation":4,"labels":{"app.kubernetes.io/instance":"datadog","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"datadog","app.kubernetes.io/version":"7","helm.sh/chart":"datadog-3.10.1"},"name":"datadog","resourceVersion":"7173625748","uid":"afc7e4af-cd2e-4a67-b78e-0312c8a2d2fb"},"priority":8,"readOnlyRootFilesystem":false,"requiredDropCapabilities":[],"runAsUser":{"type":"RunAsAny"},"seLinuxContext":{"seLinuxOptions":{"level":"s0","role":"system_r","type":"spc_t","user":"system_u"},"type":"MustRunAs"},"seccompProfiles":["runtime/default","localhost/system-probe"],"supplementalGroups":{"type":"RunAsAny"},"users":["system:serviceaccount:datadog:datadog"],"volumes":["configMap","downwardAPI","emptyDir","hostPath","secret"]}
  labels:
    app.kubernetes.io/instance: datadog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: datadog
    app.kubernetes.io/version: "7"
    helm.sh/chart: datadog-3.10.1
  name: datadog
priority: 8
readOnlyRootFilesystem: false
requiredDropCapabilities: []
runAsUser:
  type: RunAsAny
seLinuxContext:
  seLinuxOptions:
    level: s0
    role: system_r
    type: spc_t
    user: system_u
  type: MustRunAs
seccompProfiles:
- runtime/default
- localhost/system-probe
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:datadog:datadog
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- secret
`)

func testQeTestdataOlmSccYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmSccYaml, nil
}

func testQeTestdataOlmSccYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmSccYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/scc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmScopedSaEtcdYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: scoped-24886
rules:
  - apiGroups: [""]
    resources:
      [
        "pods",
        "services",
        "services/finalizers",
        "endpoints",
        "persistentvolumeclaims",
        "events",
        "configmaps",
        "secrets",
        "serviceaccounts",
      ]
    verbs: ["*"]
  - apiGroups: ["apps"]
    resources: ["deployments", "daemonsets", "replicasets", "statefulsets"]
    verbs: ["*"]
  - apiGroups: ["monitoring.coreos.com"]
    resources: ["servicemonitors"]
    verbs: ["get", "create"]
  - apiGroups: ["apps"]
    resources: ["deployments/finalizers"]
    resourceNames: ["learn-operator"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
  - apiGroups: ["apps"]
    resources: ["replicasets", "deployments"]
    verbs: ["get"]
  - apiGroups: ["app.learn.com"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["operators.coreos.com"]
    resources: ["subscriptions", "clusterserviceversions"]
    verbs: ["get", "create", "update", "patch"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "create", "update", "patch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: scoped-bindings-24886
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: scoped-24886
subjects:
  - kind: ServiceAccount
    name: scoped-24886
`)

func testQeTestdataOlmScopedSaEtcdYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmScopedSaEtcdYaml, nil
}

func testQeTestdataOlmScopedSaEtcdYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmScopedSaEtcdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/scoped-sa-etcd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmScopedSaFineGrainedRolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: scoped-24772
rules:
  - apiGroups: [""]
    resources:
      [
        "pods",
        "services",
        "services/finalizers",
        "endpoints",
        "persistentvolumeclaims",
        "events",
        "configmaps",
        "secrets",
        "serviceaccounts",
      ]
    verbs: ["*"]
  - apiGroups: ["apps"]
    resources: ["deployments", "daemonsets", "replicasets", "statefulsets"]
    verbs: ["*"]
  - apiGroups: ["monitoring.coreos.com"]
    resources: ["servicemonitors"]
    verbs: ["get", "create"]
  - apiGroups: ["apps"]
    resources: ["deployments/finalizers"]
    resourceNames: ["learn-operator"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
  - apiGroups: ["apps"]
    resources: ["replicasets", "deployments"]
    verbs: ["get"]
  - apiGroups: ["app.learn.com"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["operators.coreos.com"]
    resources: ["subscriptions", "clusterserviceversions"]
    verbs: ["get", "create", "update", "patch"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: scoped-bindings-24772
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: scoped-24772
subjects:
  - kind: ServiceAccount
    name: scoped-24772
`)

func testQeTestdataOlmScopedSaFineGrainedRolesYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmScopedSaFineGrainedRolesYaml, nil
}

func testQeTestdataOlmScopedSaFineGrainedRolesYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmScopedSaFineGrainedRolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/scoped-sa-fine-grained-roles.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmScopedSaRolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: scoped-24771
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: scoped-bindings-24771
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: scoped-24771
subjects:
  - kind: ServiceAccount
    name: scoped-24771
`)

func testQeTestdataOlmScopedSaRolesYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmScopedSaRolesYaml, nil
}

func testQeTestdataOlmScopedSaRolesYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmScopedSaRolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/scoped-sa-roles.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmSecretYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: secret-template
objects:
- kind: Secret
  apiVersion: v1
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    annotations:
      kubernetes.io/service-account.name: ${SANAME}
  type: ${TYPE}
parameters:
- name: NAME
- name: NAMESPACE
- name: SANAME
- name: TYPE
  value: "kubernetes.io/service-account-token"
`)

func testQeTestdataOlmSecretYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmSecretYaml, nil
}

func testQeTestdataOlmSecretYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmSecretYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/secret.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmSecret_opaqueYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: secret-template
objects:
- apiVersion: v1
  kind: Secret
  metadata:
    name: "${NAME}"
    namespace: "${NAMESPACE}"
  type: Opaque
  stringData:
    mykey: mypass

parameters:
- name: NAME
- name: NAMESPACE
`)

func testQeTestdataOlmSecret_opaqueYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmSecret_opaqueYaml, nil
}

func testQeTestdataOlmSecret_opaqueYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmSecret_opaqueYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/secret_opaque.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testQeTestdataOlmVpaCrdYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: vpa-template
objects:
- kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1
  metadata:
    name: "${NAME}"
    annotations:
      "api-approved.kubernetes.io": "https://github.com/kubernetes/kubernetes/pull/63797"
  spec:
    group: autoscaling.k8s.io
    scope: Namespaced
    names:
      plural: verticalpodautoscalers
      singular: verticalpodautoscaler
      kind: VerticalPodAutoscaler
      shortNames:
        - vpa
    version: v1beta1
    versions:
      - name: v1beta1
        served: false
        storage: false
        schema:
          openAPIV3Schema:
            type: object
            properties:
              apiVersion:
                type: string
              kind:
                type: string
              metadata:
                type: object
              spec:
                type: object
                x-kubernetes-preserve-unknown-fields: true
              status:
                type: object
                x-kubernetes-preserve-unknown-fields: true
      - name: v1beta2
        served: true
        storage: true
        schema:
          openAPIV3Schema:
            type: object
            properties:
              apiVersion:
                type: string
              kind:
                type: string
              metadata:
                type: object
              spec:
                type: object
                x-kubernetes-preserve-unknown-fields: true
              status:
                type: object
                x-kubernetes-preserve-unknown-fields: true
      - name: v1
        served: true
        storage: false
        schema:
          openAPIV3Schema:
            type: object
            properties:
              apiVersion:
                type: string
              kind:
                type: string
              metadata:
                type: object
              spec:
                type: object
                x-kubernetes-preserve-unknown-fields: true
              status:
                type: object
                x-kubernetes-preserve-unknown-fields: true
parameters:
- name: NAME
  value: "verticalpodautoscalers.autoscaling.k8s.io"
`)

func testQeTestdataOlmVpaCrdYamlBytes() ([]byte, error) {
	return _testQeTestdataOlmVpaCrdYaml, nil
}

func testQeTestdataOlmVpaCrdYaml() (*asset, error) {
	bytes, err := testQeTestdataOlmVpaCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test/qe/testdata/olm/vpa-crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"test/qe/testdata/olm/catalogsource-address.yaml":                        testQeTestdataOlmCatalogsourceAddressYaml,
	"test/qe/testdata/olm/catalogsource-configmap.yaml":                      testQeTestdataOlmCatalogsourceConfigmapYaml,
	"test/qe/testdata/olm/catalogsource-image-cacheless.yaml":                testQeTestdataOlmCatalogsourceImageCachelessYaml,
	"test/qe/testdata/olm/catalogsource-image-extract.yaml":                  testQeTestdataOlmCatalogsourceImageExtractYaml,
	"test/qe/testdata/olm/catalogsource-image-incorrect-updatestrategy.yaml": testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYaml,
	"test/qe/testdata/olm/catalogsource-image.yaml":                          testQeTestdataOlmCatalogsourceImageYaml,
	"test/qe/testdata/olm/catalogsource-legacy.yaml":                         testQeTestdataOlmCatalogsourceLegacyYaml,
	"test/qe/testdata/olm/catalogsource-namespace.yaml":                      testQeTestdataOlmCatalogsourceNamespaceYaml,
	"test/qe/testdata/olm/catalogsource-opm.yaml":                            testQeTestdataOlmCatalogsourceOpmYaml,
	"test/qe/testdata/olm/cm-21824-correct.yaml":                             testQeTestdataOlmCm21824CorrectYaml,
	"test/qe/testdata/olm/cm-21824-wrong.yaml":                               testQeTestdataOlmCm21824WrongYaml,
	"test/qe/testdata/olm/cm-25644-etcd-csv.yaml":                            testQeTestdataOlmCm25644EtcdCsvYaml,
	"test/qe/testdata/olm/cm-csv-etcd.yaml":                                  testQeTestdataOlmCmCsvEtcdYaml,
	"test/qe/testdata/olm/cm-namespaceconfig.yaml":                           testQeTestdataOlmCmNamespaceconfigYaml,
	"test/qe/testdata/olm/cm-template.yaml":                                  testQeTestdataOlmCmTemplateYaml,
	"test/qe/testdata/olm/configmap-ectd-alpha-beta.yaml":                    testQeTestdataOlmConfigmapEctdAlphaBetaYaml,
	"test/qe/testdata/olm/configmap-etcd.yaml":                               testQeTestdataOlmConfigmapEtcdYaml,
	"test/qe/testdata/olm/configmap-test.yaml":                               testQeTestdataOlmConfigmapTestYaml,
	"test/qe/testdata/olm/configmap-with-defaultchannel.yaml":                testQeTestdataOlmConfigmapWithDefaultchannelYaml,
	"test/qe/testdata/olm/configmap-without-defaultchannel.yaml":             testQeTestdataOlmConfigmapWithoutDefaultchannelYaml,
	"test/qe/testdata/olm/cr-webhookTest.yaml":                               testQeTestdataOlmCrWebhooktestYaml,
	"test/qe/testdata/olm/cr_devworkspace.yaml":                              testQeTestdataOlmCr_devworkspaceYaml,
	"test/qe/testdata/olm/cr_pgadmin.yaml":                                   testQeTestdataOlmCr_pgadminYaml,
	"test/qe/testdata/olm/cs-image-template.yaml":                            testQeTestdataOlmCsImageTemplateYaml,
	"test/qe/testdata/olm/cs-without-image.yaml":                             testQeTestdataOlmCsWithoutImageYaml,
	"test/qe/testdata/olm/cs-without-interval.yaml":                          testQeTestdataOlmCsWithoutIntervalYaml,
	"test/qe/testdata/olm/cs-without-scc.yaml":                               testQeTestdataOlmCsWithoutSccYaml,
	"test/qe/testdata/olm/csc.yaml":                                          testQeTestdataOlmCscYaml,
	"test/qe/testdata/olm/env-subscription.yaml":                             testQeTestdataOlmEnvSubscriptionYaml,
	"test/qe/testdata/olm/envfrom-subscription.yaml":                         testQeTestdataOlmEnvfromSubscriptionYaml,
	"test/qe/testdata/olm/etcd-cluster.yaml":                                 testQeTestdataOlmEtcdClusterYaml,
	"test/qe/testdata/olm/etcd-subscription-manual.yaml":                     testQeTestdataOlmEtcdSubscriptionManualYaml,
	"test/qe/testdata/olm/etcd-subscription.yaml":                            testQeTestdataOlmEtcdSubscriptionYaml,
	"test/qe/testdata/olm/mc-workload-partition.yaml":                        testQeTestdataOlmMcWorkloadPartitionYaml,
	"test/qe/testdata/olm/og-allns.yaml":                                     testQeTestdataOlmOgAllnsYaml,
	"test/qe/testdata/olm/og-multins.yaml":                                   testQeTestdataOlmOgMultinsYaml,
	"test/qe/testdata/olm/olm-proxy-subscription.yaml":                       testQeTestdataOlmOlmProxySubscriptionYaml,
	"test/qe/testdata/olm/olm-subscription.yaml":                             testQeTestdataOlmOlmSubscriptionYaml,
	"test/qe/testdata/olm/operator.yaml":                                     testQeTestdataOlmOperatorYaml,
	"test/qe/testdata/olm/operatorgroup-serviceaccount.yaml":                 testQeTestdataOlmOperatorgroupServiceaccountYaml,
	"test/qe/testdata/olm/operatorgroup-upgradestrategy.yaml":                testQeTestdataOlmOperatorgroupUpgradestrategyYaml,
	"test/qe/testdata/olm/operatorgroup.yaml":                                testQeTestdataOlmOperatorgroupYaml,
	"test/qe/testdata/olm/opsrc.yaml":                                        testQeTestdataOlmOpsrcYaml,
	"test/qe/testdata/olm/packageserver.yaml":                                testQeTestdataOlmPackageserverYaml,
	"test/qe/testdata/olm/platform_operator.yaml":                            testQeTestdataOlmPlatform_operatorYaml,
	"test/qe/testdata/olm/prometheus-antiaffinity.yaml":                      testQeTestdataOlmPrometheusAntiaffinityYaml,
	"test/qe/testdata/olm/prometheus-nodeaffinity.yaml":                      testQeTestdataOlmPrometheusNodeaffinityYaml,
	"test/qe/testdata/olm/role-binding.yaml":                                 testQeTestdataOlmRoleBindingYaml,
	"test/qe/testdata/olm/role.yaml":                                         testQeTestdataOlmRoleYaml,
	"test/qe/testdata/olm/scc.yaml":                                          testQeTestdataOlmSccYaml,
	"test/qe/testdata/olm/scoped-sa-etcd.yaml":                               testQeTestdataOlmScopedSaEtcdYaml,
	"test/qe/testdata/olm/scoped-sa-fine-grained-roles.yaml":                 testQeTestdataOlmScopedSaFineGrainedRolesYaml,
	"test/qe/testdata/olm/scoped-sa-roles.yaml":                              testQeTestdataOlmScopedSaRolesYaml,
	"test/qe/testdata/olm/secret.yaml":                                       testQeTestdataOlmSecretYaml,
	"test/qe/testdata/olm/secret_opaque.yaml":                                testQeTestdataOlmSecret_opaqueYaml,
	"test/qe/testdata/olm/vpa-crd.yaml":                                      testQeTestdataOlmVpaCrdYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"test": {nil, map[string]*bintree{
		"qe": {nil, map[string]*bintree{
			"testdata": {nil, map[string]*bintree{
				"olm": {nil, map[string]*bintree{
					"catalogsource-address.yaml":                        {testQeTestdataOlmCatalogsourceAddressYaml, map[string]*bintree{}},
					"catalogsource-configmap.yaml":                      {testQeTestdataOlmCatalogsourceConfigmapYaml, map[string]*bintree{}},
					"catalogsource-image-cacheless.yaml":                {testQeTestdataOlmCatalogsourceImageCachelessYaml, map[string]*bintree{}},
					"catalogsource-image-extract.yaml":                  {testQeTestdataOlmCatalogsourceImageExtractYaml, map[string]*bintree{}},
					"catalogsource-image-incorrect-updatestrategy.yaml": {testQeTestdataOlmCatalogsourceImageIncorrectUpdatestrategyYaml, map[string]*bintree{}},
					"catalogsource-image.yaml":                          {testQeTestdataOlmCatalogsourceImageYaml, map[string]*bintree{}},
					"catalogsource-legacy.yaml":                         {testQeTestdataOlmCatalogsourceLegacyYaml, map[string]*bintree{}},
					"catalogsource-namespace.yaml":                      {testQeTestdataOlmCatalogsourceNamespaceYaml, map[string]*bintree{}},
					"catalogsource-opm.yaml":                            {testQeTestdataOlmCatalogsourceOpmYaml, map[string]*bintree{}},
					"cm-21824-correct.yaml":                             {testQeTestdataOlmCm21824CorrectYaml, map[string]*bintree{}},
					"cm-21824-wrong.yaml":                               {testQeTestdataOlmCm21824WrongYaml, map[string]*bintree{}},
					"cm-25644-etcd-csv.yaml":                            {testQeTestdataOlmCm25644EtcdCsvYaml, map[string]*bintree{}},
					"cm-csv-etcd.yaml":                                  {testQeTestdataOlmCmCsvEtcdYaml, map[string]*bintree{}},
					"cm-namespaceconfig.yaml":                           {testQeTestdataOlmCmNamespaceconfigYaml, map[string]*bintree{}},
					"cm-template.yaml":                                  {testQeTestdataOlmCmTemplateYaml, map[string]*bintree{}},
					"configmap-ectd-alpha-beta.yaml":                    {testQeTestdataOlmConfigmapEctdAlphaBetaYaml, map[string]*bintree{}},
					"configmap-etcd.yaml":                               {testQeTestdataOlmConfigmapEtcdYaml, map[string]*bintree{}},
					"configmap-test.yaml":                               {testQeTestdataOlmConfigmapTestYaml, map[string]*bintree{}},
					"configmap-with-defaultchannel.yaml":                {testQeTestdataOlmConfigmapWithDefaultchannelYaml, map[string]*bintree{}},
					"configmap-without-defaultchannel.yaml":             {testQeTestdataOlmConfigmapWithoutDefaultchannelYaml, map[string]*bintree{}},
					"cr-webhookTest.yaml":                               {testQeTestdataOlmCrWebhooktestYaml, map[string]*bintree{}},
					"cr_devworkspace.yaml":                              {testQeTestdataOlmCr_devworkspaceYaml, map[string]*bintree{}},
					"cr_pgadmin.yaml":                                   {testQeTestdataOlmCr_pgadminYaml, map[string]*bintree{}},
					"cs-image-template.yaml":                            {testQeTestdataOlmCsImageTemplateYaml, map[string]*bintree{}},
					"cs-without-image.yaml":                             {testQeTestdataOlmCsWithoutImageYaml, map[string]*bintree{}},
					"cs-without-interval.yaml":                          {testQeTestdataOlmCsWithoutIntervalYaml, map[string]*bintree{}},
					"cs-without-scc.yaml":                               {testQeTestdataOlmCsWithoutSccYaml, map[string]*bintree{}},
					"csc.yaml":                                          {testQeTestdataOlmCscYaml, map[string]*bintree{}},
					"env-subscription.yaml":                             {testQeTestdataOlmEnvSubscriptionYaml, map[string]*bintree{}},
					"envfrom-subscription.yaml":                         {testQeTestdataOlmEnvfromSubscriptionYaml, map[string]*bintree{}},
					"etcd-cluster.yaml":                                 {testQeTestdataOlmEtcdClusterYaml, map[string]*bintree{}},
					"etcd-subscription-manual.yaml":                     {testQeTestdataOlmEtcdSubscriptionManualYaml, map[string]*bintree{}},
					"etcd-subscription.yaml":                            {testQeTestdataOlmEtcdSubscriptionYaml, map[string]*bintree{}},
					"mc-workload-partition.yaml":                        {testQeTestdataOlmMcWorkloadPartitionYaml, map[string]*bintree{}},
					"og-allns.yaml":                                     {testQeTestdataOlmOgAllnsYaml, map[string]*bintree{}},
					"og-multins.yaml":                                   {testQeTestdataOlmOgMultinsYaml, map[string]*bintree{}},
					"olm-proxy-subscription.yaml":                       {testQeTestdataOlmOlmProxySubscriptionYaml, map[string]*bintree{}},
					"olm-subscription.yaml":                             {testQeTestdataOlmOlmSubscriptionYaml, map[string]*bintree{}},
					"operator.yaml":                                     {testQeTestdataOlmOperatorYaml, map[string]*bintree{}},
					"operatorgroup-serviceaccount.yaml":                 {testQeTestdataOlmOperatorgroupServiceaccountYaml, map[string]*bintree{}},
					"operatorgroup-upgradestrategy.yaml":                {testQeTestdataOlmOperatorgroupUpgradestrategyYaml, map[string]*bintree{}},
					"operatorgroup.yaml":                                {testQeTestdataOlmOperatorgroupYaml, map[string]*bintree{}},
					"opsrc.yaml":                                        {testQeTestdataOlmOpsrcYaml, map[string]*bintree{}},
					"packageserver.yaml":                                {testQeTestdataOlmPackageserverYaml, map[string]*bintree{}},
					"platform_operator.yaml":                            {testQeTestdataOlmPlatform_operatorYaml, map[string]*bintree{}},
					"prometheus-antiaffinity.yaml":                      {testQeTestdataOlmPrometheusAntiaffinityYaml, map[string]*bintree{}},
					"prometheus-nodeaffinity.yaml":                      {testQeTestdataOlmPrometheusNodeaffinityYaml, map[string]*bintree{}},
					"role-binding.yaml":                                 {testQeTestdataOlmRoleBindingYaml, map[string]*bintree{}},
					"role.yaml":                                         {testQeTestdataOlmRoleYaml, map[string]*bintree{}},
					"scc.yaml":                                          {testQeTestdataOlmSccYaml, map[string]*bintree{}},
					"scoped-sa-etcd.yaml":                               {testQeTestdataOlmScopedSaEtcdYaml, map[string]*bintree{}},
					"scoped-sa-fine-grained-roles.yaml":                 {testQeTestdataOlmScopedSaFineGrainedRolesYaml, map[string]*bintree{}},
					"scoped-sa-roles.yaml":                              {testQeTestdataOlmScopedSaRolesYaml, map[string]*bintree{}},
					"secret.yaml":                                       {testQeTestdataOlmSecretYaml, map[string]*bintree{}},
					"secret_opaque.yaml":                                {testQeTestdataOlmSecret_opaqueYaml, map[string]*bintree{}},
					"vpa-crd.yaml":                                      {testQeTestdataOlmVpaCrdYaml, map[string]*bintree{}},
				}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
