module github.com/openshift/operator-framework-olm

go 1.16

require (
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/go-logr/logr v0.4.0
	github.com/golang/mock v1.6.0
	github.com/googleapis/gnostic v0.5.5
	github.com/grpc-ecosystem/grpc-health-probe v0.4.4
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/mikefarah/yq/v3 v3.0.0-20201202084205-8846255d1c37
	github.com/onsi/ginkgo v1.16.4
	github.com/openshift/api v0.0.0-20200331152225-585af27e34fd
	github.com/operator-framework/api v0.10.8-0.20211210002341-1eb6c0266cce
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-00010101000000-000000000000
	github.com/operator-framework/operator-registry v1.17.5
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.6.2
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/code-generator v0.22.1
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	sigs.k8s.io/controller-runtime v0.10.0
	sigs.k8s.io/controller-tools v0.6.2
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

	// controller runtime
	github.com/openshift/api => github.com/openshift/api v0.0.0-20210517065120-b325f58df679
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0 // release-4.5

	// use staged repositories
	github.com/operator-framework/api => ./staging/api
	github.com/operator-framework/operator-lifecycle-manager => ./staging/operator-lifecycle-manager
	github.com/operator-framework/operator-registry => ./staging/operator-registry

	// pinned because no tag supports 1.18 yet
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
)
