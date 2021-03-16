module github.com/openshift/operator-framework-olm

go 1.15

require (
	github.com/Masterminds/sprig/v3 v3.2.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/gofrs/flock v0.8.0 // indirect
	github.com/golang/mock v1.4.4
	github.com/googleapis/gnostic v0.5.1
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/grpc-ecosystem/grpc-health-probe v0.3.6
	github.com/mattn/go-shellwords v1.0.10 // indirect
	github.com/mattn/go-sqlite3 v1.12.0 // indirect
	github.com/maxbrunsfeld/counterfeiter/v6 v6.3.0
	github.com/mikefarah/yq/v3 v3.0.0-20201202084205-8846255d1c37
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/openshift/api v0.0.0-20200331152225-585af27e34fd
	github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0
	github.com/operator-framework/api v0.0.0-00010101000000-000000000000
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-00010101000000-000000000000
	github.com/operator-framework/operator-registry v0.0.0-00010101000000-000000000000
	github.com/otiai10/copy v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.10.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	google.golang.org/grpc v1.34.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/apiserver v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.2
	k8s.io/kube-aggregator v0.20.2
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kubectl v0.20.0 // indirect
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/controller-tools v0.4.1
	sigs.k8s.io/kind v0.7.0
)

replace (
	// From staging/operator-registry
	// Currently on a fork for two issues:
	// 1. stage registry proxy didn't like requests with no scopes, see https://github.com/containerd/containerd/pull/4223
	// 2. prod registry proxy returns a 403 on post, see https://github.com/containerd/containerd/pull/3913
	// The fork can be removed when both issues are resolved in a release, which should be 1.4.0
	github.com/containerd/containerd => github.com/ecordell/containerd v1.3.1-0.20200629153125-0ff1a1be2fa5

	// latest tag resolves to a very old version. this is only used for spinning up local test registries
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

	// From staging/operator-lifecycle-manager
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1

	// controller runtime
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200331152225-585af27e34fd // release-4.5
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0 // release-4.5

	// use staged repositories
	github.com/operator-framework/api => ./staging/api
	github.com/operator-framework/operator-lifecycle-manager => ./staging/operator-lifecycle-manager
	github.com/operator-framework/operator-registry => ./staging/operator-registry

	// pinned because latest etcd does not yet work with the latest grpc version (1.30.0)
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200520232829-54ba9589114f
	google.golang.org/grpc => google.golang.org/grpc v1.27.0
	google.golang.org/grpc/examples => google.golang.org/grpc/examples v0.0.0-20200709232328-d8193ee9cc3e

	// pinned because no tag supports 1.18 yet
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
)
