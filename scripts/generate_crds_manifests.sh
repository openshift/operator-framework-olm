#!/bin/bash

set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd ${repo_root}
export GOFLAGS="-mod=vendor"
YQ="go run ./vendor/github.com/mikefarah/yq/v3/"
CONTROLLER_GEN="go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen"
ver=$(cat ./OLM_VERSION)

ln -snf  $(realpath --relative-to ./crds ./staging/api/pkg/operators/) ./crds/operators

${CONTROLLER_GEN} crd:crdVersions=v1 output:crd:dir=./crds paths=./crds/operators/...
${CONTROLLER_GEN} schemapatch:manifests=./crds output:dir=./crds paths=./crds/operators/...

${YQ} w --inplace ./crds/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.containers.items.properties.ports.items.properties.protocol.default TCP
${YQ} w --inplace ./crds/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.initContainers.items.properties.ports.items.properties.protocol.default TCP
${YQ} w --inplace ./crds/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.metadata.x-kubernetes-preserve-unknown-fields true
${YQ} d --inplace ./crds/operators.coreos.com_operatorconditions.yaml 'spec.versions[*].schema.openAPIV3Schema.properties.spec.properties.overrides.items.required(.==lastTransitionTime)'

rm ./deploy/chart/crds/*.yaml

for f in ./crds/*.yaml ; do
    ${YQ} d --inplace $f status
    cp "$f" "./deploy/chart/crds/0000_50_olm_00-$(basename $f | sed 's/^.*_\([^.]\+\)\.yaml/\1.crd.yaml/')"
done

charttmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t charttmpdir)/chart"

cp -R deploy/chart/ "${charttmpdir}"
sed -i "s/^[Vv]ersion:.*\$/version: ${ver}/" "${charttmpdir}/Chart.yaml"

go run helm.sh/helm/v3/cmd/helm template -n olm -f "deploy/ocp/values.yaml" --include-crds --output-dir "${charttmpdir}" "${charttmpdir}"

mkdir -p "deploy/ocp/manifests/${ver}"
cp -R "${charttmpdir}"/olm/{templates,crds}/. "deploy/ocp/manifests/${ver}"

for f in deploy/ocp/manifests/${ver}/*.yaml; do
   ${YQ} w -d'*' --inplace --style=double $f 'metadata.annotations['include.release.openshift.io/self-managed-high-availability']' true
done

ln -sfFn ./${ver} deploy/ocp/manifests/latest

rm -rf ./manifests/*

cp -R deploy/ocp/manifests/${ver}/. ./manifests
# requires gnu sed if on mac
find ./manifests -type f -exec sed -i "/^#/d" {} \;
find ./manifests -type f -exec sed -i "1{/---/d}" {} \;
