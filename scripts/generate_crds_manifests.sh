#!/bin/bash 

set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd ${repo_root}
export  GOFLAGS="-mod=vendor"
YQ="go run ./vendor/github.com/mikefarah/yq/v3/"
CONTROLLER_GEN="go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen"
ver=$(cat ./OLM_VERSION)

tmpdir="$(mktemp -p . -d 2>/dev/null || mktemp -p . -d -t tmpdir)"
chartdir="${tmpdir}/chart"
crddir="${chartdir}/crds"
crdsrcdir="${tmpdir}/operators"

cp -R "./staging/operator-lifecycle-manager/deploy/chart/" "${chartdir}"
cp "./staging/operator-lifecycle-manager/deploy/ocp/values.yaml" ${tmpdir}
ln -snf $(realpath --relative-to ${tmpdir} ./staging/api/pkg/operators/) ${crdsrcdir}
rm -rf ./manifests/* ${crddir}/*

trap "rm -rf ${tmpdir}" EXIT

${CONTROLLER_GEN} crd:crdVersions=v1 output:crd:dir=${crddir} paths=${crdsrcdir}/...
${CONTROLLER_GEN} schemapatch:manifests=${crddir} output:dir=${crddir} paths=${crdsrcdir}/...

${YQ} w --inplace ${crddir}/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.containers.items.properties.ports.items.properties.protocol.default TCP
${YQ} w --inplace ${crddir}/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.initContainers.items.properties.ports.items.properties.protocol.default TCP
${YQ} w --inplace ${crddir}/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.metadata.x-kubernetes-preserve-unknown-fields true
${YQ} d --inplace ${crddir}/operators.coreos.com_operatorconditions.yaml 'spec.versions[*].schema.openAPIV3Schema.properties.spec.properties.overrides.items.required(.==lastTransitionTime)'

for f in ${crddir}/*.yaml ; do
    ${YQ} d --inplace $f status
    mv -v "$f" "${crddir}/0000_50_olm_00-$(basename $f | sed 's/^.*_\([^.]\+\)\.yaml/\1.crd.yaml/')"
done

sed -i "s/^[Vv]ersion:.*\$/version: ${ver}/" "${chartdir}/Chart.yaml"

go run helm.sh/helm/v3/cmd/helm template -n olm -f "${tmpdir}/values.yaml" --include-crds --output-dir "${chartdir}" "${chartdir}"

cp -R "${chartdir}"/olm/{templates,crds}/. "./manifests"

for f in ./manifests/*.yaml; do
   if [[ ! "$(basename "${f}")" =~ .*\.deployment\..* ]]; then
      ${YQ} w -d'*' --inplace --style=double $f 'metadata.annotations['include.release.openshift.io/ibm-cloud-managed']' true
   else
      g="${f/%.yaml/.ibm-cloud-managed.yaml}"
      cp "${f}" "${g}"
      ${YQ} w -d'*' --inplace --style=double $g 'metadata.annotations['include.release.openshift.io/ibm-cloud-managed']' true
      ${YQ} d -d'*' --inplace $g 'spec.template.spec.nodeSelector."node-role.kubernetes.io/master"'
   fi
   ${YQ} w -d'*' --inplace --style=double $f 'metadata.annotations['include.release.openshift.io/self-managed-high-availability']' true
   ${YQ} w -d'*' --inplace --style=double $f 'metadata.annotations['include.release.openshift.io/single-node-developer']' true
done

# requires gnu sed if on mac
find ./manifests -type f -exec sed -i "/^#/d" {} \;
find ./manifests -type f -exec sed -i "1{/---/d}" {} \;

