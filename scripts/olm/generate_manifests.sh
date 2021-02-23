#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

export GOFLAGS="-mod=vendor"
repo_root=$(git rev-parse --show-toplevel)
api_repo="$repo_root/staging/api"
controller_gen="$repo_root/bin/controller-gen"
crd_dir="$repo_root/crds"
yq="go run ./vendor/github.com/mikefarah/yq/v3/"
cd $repo_root

# Download and build controller-gen 
go mod tidy -v && go mod vendor
go build -o $controller_gen $repo_root/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen

cd $api_repo

# Create CRDs for new APIs
$controller_gen crd:crdVersions=v1 output:crd:dir=$crd_dir paths=./...

# Update existing CRDs from type changes
$controller_gen schemapatch:manifests=./crds output:dir=$crd_dir paths=./...

cd $repo_root
# Add missing defaults in embedded core API schemas
$yq w --inplace "$crd_dir/operators.coreos.com_clusterserviceversions.yaml" 'spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.containers.items.properties.ports.items.properties.protocol.default' 'TCP'
$yq w --inplace "$crd_dir/operators.coreos.com_clusterserviceversions.yaml" 'spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.initContainers.items.properties.ports.items.properties.protocol.default' 'TCP'

# Preserve fields for embedded metadata fields
$yq w --inplace "$crd_dir/operators.coreos.com_clusterserviceversions.yaml" 'spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.metadata.x-kubernetes-preserve-unknown-fields' 'true'

# Remove OperatorCondition.spec.overrides[*].lastTransitionTime requirement
$yq d --inplace "$crd_dir/operators.coreos.com_operatorconditions.yaml" 'spec.versions[*].schema.openAPIV3Schema.properties.spec.properties.overrides.items.required(.==lastTransitionTime)'

# Remove status subresource from the CRD manifests to ensure server-side apply works
for f in $crd_dir/*.yaml; do
	$yq d --inplace "$f" 'status'
done

# Update embedded CRD files.
go generate $crd_dir/...


# Copy CRDS manifests
rm "${repo_root}"/deploy/chart/crds/*.yaml
for f in "${crd_dir}"/*.yaml ; do
    echo "copying ${f}"
    cp "${f}" "${repo_root}/deploy/chart/crds/0000_50_olm_00-$(basename "$f" | sed 's/^.*_\([^.]\+\)\.yaml/\1.crd.yaml/')"
done

