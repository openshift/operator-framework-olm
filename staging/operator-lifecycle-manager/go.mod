module github.com/operator-framework/operator-lifecycle-manager

go 1.16

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/go-logr/logr v0.4.0
	github.com/go-openapi/spec v0.19.5
	github.com/golang/mock v1.4.1
	github.com/google/go-cmp v0.5.6
	github.com/googleapis/gnostic v0.5.5
	github.com/irifrance/gini v1.0.1
	github.com/itchyny/gojq v0.11.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.2
	github.com/mikefarah/yq/v3 v3.0.0-20201202084205-8846255d1c37
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/openshift/api v0.0.0-20200331152225-585af27e34fd
	github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0
	github.com/operator-framework/api v0.7.1
	github.com/operator-framework/operator-registry v0.0.0-00010101000000-000000000000
	github.com/otiai10/copy v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.26.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.30.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.1.0-rc.1.0.20201215141456-e71d38b414eb
	// can't update to 0.21 until https://github.com/kubernetes/apiserver/issues/65 is resolved
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/apiserver v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/code-generator v0.21.1
	k8s.io/component-base v0.21.1
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/controller-tools v0.6.0
	sigs.k8s.io/kind v0.11.1
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1

	// controller runtime
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200331152225-585af27e34fd // release-4.5
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0 // release-4.5

	// pinned because latest etcd does not yet work with the latest grpc version (1.30.0)
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200520232829-54ba9589114f
	google.golang.org/grpc => google.golang.org/grpc v1.27.0
	google.golang.org/grpc/examples => google.golang.org/grpc/examples v0.0.0-20200709232328-d8193ee9cc3e

	// pinned for delegated authentication watch request bug fix.
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0-beta.1.0.20210308143346-a13af1068ef1
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20210409112051-49d90ce0ad13
	k8s.io/component-base => k8s.io/component-base v0.0.0-20210105235135-9c158118ed58

	// pinned because no tag supports 1.18 yet
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
)

replace github.com/operator-framework/operator-registry => ../operator-registry

replace github.com/operator-framework/api => ../api
