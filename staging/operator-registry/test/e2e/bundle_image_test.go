package e2e_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/operator-framework/operator-registry/pkg/client"
	"github.com/operator-framework/operator-registry/pkg/configmap"
	"github.com/operator-framework/operator-registry/pkg/lib/bundle"
)

var builderCmd string

const (
	imageDirectory = "testdata/image-bundle/"
)

func Logf(format string, a ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "INFO: "+format+"\n", a...)
}

// checks command that it exists in $PATH, isn't a directory, and has executable permissions set
func checkCommand(filename string) string {
	path, err := exec.LookPath(filename)
	if err != nil {
		Logf("LookPath error: %v", err)
		return ""
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		Logf("IsNotExist error: %v", err)
		return ""
	}
	if info.IsDir() {
		return ""
	}
	perm := info.Mode()
	if perm&0111 != 0x49 { // 000100100100 - execute bits set for user/group/everybody
		Logf("permissions failure: %#v %#v", perm, perm&0111)
		return ""
	}

	Logf("Using builder at '%v'\n", path)
	return path
}

func init() {
	logrus.SetOutput(GinkgoWriter)

	if builderCmd = checkCommand("docker"); builderCmd != "" {
		return
	}
	if builderCmd = checkCommand("podman"); builderCmd != "" {
		return
	}
}

// newlines included in each section, uses spaces and no tabs
func getConfigMapDataSection() map[string]string {
	return map[string]string{
		"kiali.crd.yaml": `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: kialis.kiali.io
  labels:
    app: kiali-operator
spec:
  group: kiali.io
  names:
    kind: Kiali
    listKind: KialiList
    plural: kialis
    singular: kiali
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
`,
		"kiali.monitoringdashboards.crd.yaml": `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: monitoringdashboards.monitoring.kiali.io
  labels:
    app: kiali
spec:
  group: monitoring.kiali.io
  names:
    kind: MonitoringDashboard
    listKind: MonitoringDashboardList
    plural: monitoringdashboards
    singular: monitoringdashboard
  scope: Namespaced
  version: v1alpha1
`,
		"kiali.package.yaml": `packageName: kiali
channels:
- name: alpha
  currentCSV: kiali-operator.v1.4.2
- name: stable
  currentCSV: kiali-operator.v1.4.2
defaultChannel: stable
`,
		"kiali.v1.4.2.clusterserviceversion.yaml": `apiVersion: operators.coreos.com/v1alpha1
kind: ClusterServiceVersion
metadata:
  name: kiali-operator.v1.4.2
  namespace: placeholder
  annotations:
    categories: Monitoring,Logging & Tracing
    certified: "false"
    containerImage: quay.io/kiali/kiali-operator:v1.4.2
    capabilities: Basic Install
    support: Kiali
    description: "Kiali project provides answers to the questions: What microservices are part of my Istio service mesh and how are they connected?"
    repository: https://github.com/kiali/kiali
    createdAt: 2019-09-12T00:00:00Z
    alm-examples: |-
      [
        {
          "apiVersion": "kiali.io/v1alpha1",
          "kind": "Kiali",
          "metadata": {
            "name": "kiali"
          },
          "spec": {
            "installation_tag": "My Kiali",
            "istio_namespace": "istio-system",
            "deployment": {
              "namespace": "istio-system",
              "verbose_mode": "4",
              "view_only_mode": false
            },
            "external_services": {
              "grafana": {
                "url": ""
              },
              "prometheus": {
                "url": ""
              },
              "tracing": {
                "url": ""
              }
            },
            "server": {
              "web_root": "/mykiali"
            }
          }
        },
        {
          "apiVersion": "monitoring.kiali.io/v1alpha1",
          "kind": "MonitoringDashboard",
          "metadata": {
            "name": "myappdashboard"
          },
          "spec": {
            "title": "My App Dashboard",
            "items": [
              {
                "chart": {
                  "name": "My App Processing Duration",
                  "unit": "seconds",
                  "spans": 6,
                  "metricName": "my_app_duration_seconds",
                  "dataType": "histogram",
                  "aggregations": [
                    {
                      "label": "id",
                      "displayName": "ID"
                    }
                  ]
                }
              }
            ]
          }
        }
      ]
spec:
  version: 1.4.2
  maturity: stable
  replaces: kiali-operator.v1.3.1
  displayName: Kiali Operator
  description: |-
    A Microservice Architecture breaks up the monolith into many smaller pieces that are composed together. Patterns to secure the communication between services like fault tolerance (via timeout, retry, circuit breaking, etc.) have come up as well as distributed tracing to be able to see where calls are going.

    A service mesh can now provide these services on a platform level and frees the application writers from those tasks. Routing decisions are done at the mesh level.

    Kiali works with Istio, in OpenShift or Kubernetes, to visualize the service mesh topology, to provide visibility into features like circuit breakers, request rates and more. It offers insights about the mesh components at different levels, from abstract Applications to Services and Workloads.

    See [https://www.kiali.io](https://www.kiali.io) to read more.

    ### Prerequisites

    Today Kiali works with Istio. So before you install Kiali, you must have already installed Istio. Note that Istio can come pre-bundled with Kiali (specifically if you installed the Istio demo helm profile or if you installed Istio with the helm option '--set kiali.enabled=true'). If you already have the pre-bundled Kiali in your Istio environment and you want to install Kiali via the Kiali Operator, uninstall the pre-bundled Kiali first. You can do this via this command:

        kubectl delete --ignore-not-found=true all,secrets,sa,templates,configmaps,deployments,clusterroles,clusterrolebindings,ingresses,customresourcedefinitions --selector="app=kiali" -n istio-system

    When you install Kiali in a non-OpenShift Kubernetes environment, the authentication strategy will default to ` + "`" + "login" + "`" + ". When using the authentication strategy of " + "`" + "login" + "`" + ", you are required to create a Kubernetes Secret with a " + "`" + "username" + "`" + " and " + "`" + "passphrase" + "`" + " that you want users to provide in order to successfully log into Kiali. Here is an example command you can execute to create such a secret (with a username of " + "`" + "admin" + "`" + " and a passphrase of " + "`" + "admin" + "`" + `):

        kubectl create secret generic kiali -n istio-system --from-literal "username=admin" --from-literal "passphrase=admin"

    ### Kiali Custom Resource Configuration Settings

    For quick descriptions of all the settings you can configure in the Kiali Custom Resource (CR), see the file [kiali_cr.yaml](https://github.com/kiali/kiali/blob/v1.4.2/operator/deploy/kiali/kiali_cr.yaml)

    ### Accessing the UI

    By default, the Kiali operator exposes the Kiali UI as a Route on OpenShift or Ingress on Kubernetes.
    On OpenShift, the default root context path is '/' and on Kubernetes it is '/kiali' though you can change this by configuring the 'web_root' setting in the Kiali CR.
  icon:
  - base64data: PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDIyLjAuMSwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPgo8c3ZnIHZlcnNpb249IjEuMSIgaWQ9IkxheWVyXzEiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiIHg9IjBweCIgeT0iMHB4IgoJIHZpZXdCb3g9IjAgMCAxMjgwIDEyODAiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDEyODAgMTI4MDsiIHhtbDpzcGFjZT0icHJlc2VydmUiPgo8c3R5bGUgdHlwZT0idGV4dC9jc3MiPgoJLnN0MHtmaWxsOiMwMTMxNDQ7fQoJLnN0MXtmaWxsOiMwMDkzREQ7fQo8L3N0eWxlPgo8Zz4KCTxwYXRoIGNsYXNzPSJzdDAiIGQ9Ik04MTAuOSwxODAuOWMtMjUzLjYsMC00NTkuMSwyMDUuNS00NTkuMSw0NTkuMXMyMDUuNSw0NTkuMSw0NTkuMSw0NTkuMVMxMjcwLDg5My42LDEyNzAsNjQwCgkJUzEwNjQuNSwxODAuOSw4MTAuOSwxODAuOXogTTgxMC45LDEwMjkuMmMtMjE1LDAtMzg5LjItMTc0LjMtMzg5LjItMzg5LjJjMC0yMTUsMTc0LjMtMzg5LjIsMzg5LjItMzg5LjJTMTIwMC4xLDQyNSwxMjAwLjEsNjQwCgkJUzEwMjUuOSwxMDI5LjIsODEwLjksMTAyOS4yeiIvPgoJPHBhdGggY2xhc3M9InN0MSIgZD0iTTY1My4zLDI4NGMtMTM2LjQsNjAuNS0yMzEuNiwxOTcuMS0yMzEuNiwzNTZjMCwxNTguOCw5NS4yLDI5NS41LDIzMS42LDM1NmM5OC40LTg3LjEsMTYwLjQtMjE0LjMsMTYwLjQtMzU2CgkJQzgxMy43LDQ5OC4zLDc1MS42LDM3MS4xLDY1My4zLDI4NHoiLz4KCTxwYXRoIGNsYXNzPSJzdDEiIGQ9Ik0zNTEuOCw2NDBjMC0xMDkuOCwzOC42LTIxMC41LDEwMi44LTI4OS41Yy0zOS42LTE4LjItODMuNi0yOC4zLTEzMC0yOC4zQzE1MC45LDMyMi4yLDEwLDQ2NC41LDEwLDY0MAoJCXMxNDAuOSwzMTcuOCwzMTQuNiwzMTcuOGM0Ni4zLDAsOTAuNC0xMC4xLDEzMC0yOC4zQzM5MC4zLDg1MC41LDM1MS44LDc0OS44LDM1MS44LDY0MHoiLz4KPC9nPgo8L3N2Zz4K
    mediatype: image/svg+xml
  keywords: ['service-mesh', 'observability', 'monitoring', 'maistra', 'istio']
  maintainers:
  - name: Kiali Developers Google Group
    email: kiali-dev@googlegroups.com
  provider:
    name: Kiali
  labels:
    name: kiali-operator
  selector:
    matchLabels:
      name: kiali-operator
  links:
  - name: Getting Started Guide
    url: https://www.kiali.io/documentation/getting-started/
  - name: Features
    url: https://www.kiali.io/documentation/features
  - name: Documentation Home
    url: https://www.kiali.io/documentation
  - name: Blogs and Articles
    url: https://medium.com/kialiproject
  - name: Server Source Code
    url: https://github.com/kiali/kiali
  - name: UI Source Code
    url: https://github.com/kiali/kiali-ui
  installModes:
  - type: OwnNamespace
    supported: true
  - type: SingleNamespace
    supported: true
  - type: MultiNamespace
    supported: false
  - type: AllNamespaces
    supported: true
  customresourcedefinitions:
    owned:
    - name: kialis.kiali.io
      group: kiali.io
      description: A configuration file for a Kiali installation.
      displayName: Kiali
      kind: Kiali
      version: v1alpha1
      resources:
      - kind: Deployment
        version: apps/v1
      - kind: Pod
        version: v1
      - kind: Service
        version: v1
      - kind: ConfigMap
        version: v1
      - kind: OAuthClient
        version: oauth.openshift.io/v1
      - kind: Route
        version: route.openshift.io/v1
      - kind: Ingress
        version: extensions/v1beta1
      specDescriptors:
      - displayName: Authentication Strategy
        description: "Determines how a user is to log into Kiali. Choose 'login' to use a username and passphrase as defined in a Secret. Choose 'anonymous' to allow full access to Kiali without requiring credentials (use this at your own risk). Choose 'openshift' if on OpenShift to use the OpenShift OAuth login which controls access based on the individual's OpenShift RBAC roles. Default: openshift (when deployed in OpenShift); login (when deployed in Kubernetes)"
        path: auth.strategy
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:label'
      - displayName: Kiali Namespace
        description: "The namespace where Kiali and its associated resources will be created. Default: istio-system"
        path: deployment.namespace
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:label'
      - displayName: Secret Name
        description: "If Kiali is configured with auth.strategy 'login', an admin must create a Secret with credentials ('username' and 'passphrase') which will be used to authenticate users logging into Kiali. This setting defines the name of that secret. Default: kiali"
        path: deployment.secret_name
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Secret'
      - displayName: Verbose Mode
        description: "Determines the priority levels of log messages Kiali will output. Typical values are '3' for INFO and higher priority messages, '4' for DEBUG and higher priority messages (this makes the logs more noisy). Default: 3"
        path: deployment.verbose_mode
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:label'
      - displayName: View Only Mode
        description: "When true, Kiali will be in 'view only' mode, allowing the user to view and retrieve management and monitoring data for the service mesh, but not allow the user to modify the service mesh. Default: false"
        path: deployment.view_only_mode
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:booleanSwitch'
      - displayName: Web Root
        description: "Defines the root context path for the Kiali console, API endpoints and readiness/liveness probes. Default: / (when deployed on OpenShift; /kiali (when deployed on Kubernetes)"
        path: server.web_root
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:label'
    - name: monitoringdashboards.monitoring.kiali.io
      group: monitoring.kiali.io
      description: A configuration file for defining an individual metric dashboard.
      displayName: Monitoring Dashboard
      kind: MonitoringDashboard
      version: v1alpha1
      resources: []
      specDescriptors:
      - displayName: Title
        description: "The title of the dashboard."
        path: title
        x-descriptors:
        - 'urn:alm:descriptor:com.tectonic.ui:label'
  apiservicedefinitions: {}
  install:
    strategy: deployment
    spec:
      deployments:
      - name: kiali-operator
        spec:
          replicas: 1
          selector:
            matchLabels:
              app: kiali-operator
          template:
            metadata:
              name: kiali-operator
              labels:
                app: kiali-operator
                version: v1.4.2
            spec:
              serviceAccountName: kiali-operator
              containers:
              - name: ansible
                command:
                - /usr/local/bin/ao-logs
                - /tmp/ansible-operator/runner
                - stdout
                image: quay.io/kiali/kiali-operator:v1.4.2
                imagePullPolicy: "IfNotPresent"
                volumeMounts:
                - mountPath: /tmp/ansible-operator/runner
                  name: runner
                  readOnly: true
              - name: operator
                image: quay.io/kiali/kiali-operator:v1.4.2
                imagePullPolicy: "IfNotPresent"
                volumeMounts:
                - mountPath: /tmp/ansible-operator/runner
                  name: runner
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
                  value: "kiali-operator"
              volumes:
              - name: runner
                emptyDir: {}
      clusterPermissions:
      - rules:
        - apiGroups: [""]
          resources:
          - configmaps
          - endpoints
          - events
          - persistentvolumeclaims
          - pods
          - serviceaccounts
          - services
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: [""]
          resources:
          - namespaces
          verbs:
          - get
          - list
          - patch
        - apiGroups: ["apps"]
          resources:
          - deployments
          - replicasets
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: ["monitoring.coreos.com"]
          resources:
          - servicemonitors
          verbs:
          - create
          - get
        - apiGroups: ["apps"]
          resourceNames:
          - kiali-operator
          resources:
          - deployments/finalizers
          verbs:
          - update
        - apiGroups: ["kiali.io"]
          resources:
          - '*'
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: ["rbac.authorization.k8s.io"]
          resources:
          - clusterrolebindings
          - clusterroles
          - rolebindings
          - roles
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: ["apiextensions.k8s.io"]
          resources:
          - customresourcedefinitions
          verbs:
          - get
          - list
          - watch
        - apiGroups: ["extensions"]
          resources:
          - ingresses
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: ["route.openshift.io"]
          resources:
          - routes
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: ["oauth.openshift.io"]
          resources:
          - oauthclients
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        - apiGroups: ["monitoring.kiali.io"]
          resources:
          - monitoringdashboards
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - update
          - watch
        # The permissions below are for Kiali itself; operator needs these so it can escalate when creating Kiali's roles
        - apiGroups: [""]
          resources:
          - configmaps
          - endpoints
          - namespaces
          - nodes
          - pods
          - pods/log
          - replicationcontrollers
          - services
          verbs:
          - get
          - list
          - watch
        - apiGroups: ["extensions", "apps"]
          resources:
          - deployments
          - replicasets
          - statefulsets
          verbs:
          - get
          - list
          - watch
        - apiGroups: ["autoscaling"]
          resources:
          - horizontalpodautoscalers
          verbs:
          - get
          - list
          - watch
        - apiGroups: ["batch"]
          resources:
          - cronjobs
          - jobs
          verbs:
          - get
          - list
          - watch
        - apiGroups: ["config.istio.io"]
          resources:
          - adapters
          - apikeys
          - bypasses
          - authorizations
          - checknothings
          - circonuses
          - cloudwatches
          - deniers
          - dogstatsds
          - edges
          - fluentds
          - handlers
          - instances
          - kubernetesenvs
          - kuberneteses
          - listcheckers
          - listentries
          - logentries
          - memquotas
          - metrics
          - noops
          - opas
          - prometheuses
          - quotas
          - quotaspecbindings
          - quotaspecs
          - rbacs
          - redisquotas
          - reportnothings
          - rules
          - signalfxs
          - solarwindses
          - stackdrivers
          - statsds
          - stdios
          - templates
          - tracespans
          - zipkins
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - watch
        - apiGroups: ["networking.istio.io"]
          resources:
          - destinationrules
          - gateways
          - serviceentries
          - sidecars
          - virtualservices
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - watch
        - apiGroups: ["authentication.istio.io"]
          resources:
          - meshpolicies
          - policies
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - watch
        - apiGroups: ["rbac.istio.io"]
          resources:
          - clusterrbacconfigs
          - rbacconfigs
          - servicerolebindings
          - serviceroles
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - watch
        - apiGroups: ["authentication.maistra.io"]
          resources:
          - servicemeshpolicies
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - watch
        - apiGroups: ["rbac.maistra.io"]
          resources:
          - servicemeshrbacconfigs
          verbs:
          - create
          - delete
          - get
          - list
          - patch
          - watch
        - apiGroups: ["apps.openshift.io"]
          resources:
          - deploymentconfigs
          verbs:
          - get
          - list
          - watch
        - apiGroups: ["project.openshift.io"]
          resources:
          - projects
          verbs:
          - get
        - apiGroups: ["route.openshift.io"]
          resources:
          - routes
          verbs:
          - get
        - apiGroups: ["monitoring.kiali.io"]
          resources:
          - monitoringdashboards
          verbs:
          - get
          - list
        serviceAccountName: kiali-operator
`,
	}
}

func buildContainer(tag, dockerfilePath, dockerContext string, w io.Writer) {
	cmd := exec.Command(builderCmd, "build", "-t", tag, "-f", dockerfilePath, dockerContext)
	cmd.Stderr = w
	cmd.Stdout = w
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("Launch bundle", func() {
	namespace := "default"
	initImage := "init-operator-manifest:test"
	bundleImage := "bundle-image:test"

	correctConfigMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Annotations: map[string]string{
				bundle.MediatypeLabel:                 "registry+v1",
				bundle.ManifestsLabel:                 "/manifests/",
				bundle.MetadataLabel:                  "/metadata/",
				bundle.PackageLabel:                   "kiali-operator.v1.4.2",
				bundle.ChannelsLabel:                  "alpha,stable",
				bundle.ChannelDefaultLabel:            "stable",
				configmap.ConfigMapImageAnnotationKey: bundleImage,
			},
		},
		Data: getConfigMapDataSection(),
	}

	Context("Deploy bundle job", func() {
		It("should populate specified configmap", func() {
			// these permissions are only necessary for the e2e (and not OLM using the feature)
			By("configuring configmap service account")
			kubeclient, err := client.NewKubeClient("", logrus.StandardLogger())
			Expect(err).NotTo(HaveOccurred())

			_, err = kubeclient.RbacV1().Roles(namespace).Create(context.TODO(), &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "olm-dev-configmap-access",
					Namespace: namespace,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
						Verbs:     []string{"create", "get", "update"},
					},
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			_, err = kubeclient.RbacV1().RoleBindings(namespace).Create(context.TODO(), &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "olm-dev-configmap-access-binding",
					Namespace: namespace,
				},
				Subjects: []rbacv1.Subject{
					{
						APIGroup:  "",
						Kind:      "ServiceAccount",
						Name:      "default",
						Namespace: namespace,
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "olm-dev-configmap-access",
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("building required images")
			buildContainer(initImage, imageDirectory+"serve.Dockerfile", "../../bin", GinkgoWriter)
			buildContainer(bundleImage, imageDirectory+"bundle.Dockerfile", imageDirectory, GinkgoWriter)

			err = maybeLoadKind(kubeclient, GinkgoWriter, initImage, bundleImage)
			Expect(err).ToNot(HaveOccurred(), "error loading required images into cluster")

			By("creating a batch job")
			bundleDataConfigMap, job, err := configmap.LaunchBundleImage(kubeclient, bundleImage, initImage, namespace)
			Expect(err).NotTo(HaveOccurred())

			// wait for job to complete
			jobWatcher, err := kubeclient.BatchV1().Jobs(namespace).Watch(context.TODO(), metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())

			done := make(chan struct{})
			quit := make(chan struct{})
			defer close(quit)
			go func() {
				for {
					select {
					case <-quit:
						return
					case evt, ok := <-jobWatcher.ResultChan():
						if !ok {
							Logf("watch channel closed unexpectedly")
							return
						}
						if evt.Type == watch.Modified {
							job, ok := evt.Object.(*batchv1.Job)
							if !ok {
								continue
							}
							for _, condition := range job.Status.Conditions {
								if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
									done <- struct{}{}
								}
							}
						}
					case <-time.After(15 * time.Second):
						done <- struct{}{}
					}
				}
			}()

			Logf("Waiting on job to update status")
			<-done
			Logf("Job complete")

			bundleDataConfigMap, err = kubeclient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), bundleDataConfigMap.GetName(), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			assert.EqualValues(GinkgoT(), correctConfigMap.Annotations, bundleDataConfigMap.Annotations)
			assert.EqualValues(GinkgoT(), correctConfigMap.Data, bundleDataConfigMap.Data)

			// clean up, perhaps better handled elsewhere
			err = kubeclient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), bundleDataConfigMap.GetName(), metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			// job deletion does not clean up underlying pods (but using kubectl will do the clean up)
			pods, err := kubeclient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", job.GetName())})
			Expect(err).NotTo(HaveOccurred())
			err = kubeclient.CoreV1().Pods(namespace).Delete(context.TODO(), pods.Items[0].GetName(), metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			err = kubeclient.BatchV1().Jobs(namespace).Delete(context.TODO(), job.GetName(), metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

const kindControlPlaneNodeName = "kind-control-plane"

func isKindCluster(client *kubernetes.Clientset) (bool, error) {
	kindNode, err := client.CoreV1().Nodes().Get(context.TODO(), kindControlPlaneNodeName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// not a kind cluster
			return false, nil
		}
		// transient error accessing nodes in cluster
		// return an error, failing the test
		return false, fmt.Errorf("accessing nodes in cluster")
	}
	if kindNode == nil {
		return true, fmt.Errorf("finding kind node %s in cluster", kindControlPlaneNodeName)
	}
	return true, nil
}

// maybeLoadKind loads the built images via kind load docker-image if the test is running in a kind cluster
func maybeLoadKind(client *kubernetes.Clientset, w io.Writer, images ...string) error {
	kind, err := isKindCluster(client)
	if err != nil {
		return err
	}
	if !kind {
		return nil
	}

	for _, image := range images {
		cmd := exec.Command("kind", "load", "docker-image", image)
		cmd.Stderr = w
		cmd.Stdout = w
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
