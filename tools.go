// +build tools

package tools

import (
	// Generate deepcopy and conversion.
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	// Manipulate YAML.
	_ "github.com/mikefarah/yq/v3"
	// Generate embedded files.
	_ "github.com/go-bindata/go-bindata/v3/go-bindata"
	_ "github.com/golang/mock/mockgen"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "github.com/grpc-ecosystem/grpc-health-probe"
	_ "github.com/googleapis/gnostic"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "github.com/onsi/ginkgo/ginkgo"
	_ "github.com/operator-framework/api/crds" // operators.coreos.com CRD manifests
	_ "helm.sh/helm/v3/cmd/helm"
	_ "k8s.io/code-generator"
	_ "k8s.io/kube-openapi/cmd/openapi-gen"
)
