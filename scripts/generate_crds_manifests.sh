#! /bin/bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
export GOFLAGS="-mod=vendor"

YQ="go run ./vendor/github.com/mikefarah/yq/v3/"
CONTROLLER_GEN="go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen"

tmpdir="$(mktemp -p . -d 2>/dev/null || mktemp -p . -d -t tmpdir)"
chartdir="${tmpdir}/chart"
crddir="${chartdir}/crds"
crdsrcdir="${tmpdir}/operators"

# TODO(tflannag): We shouldn't be making any assumptions on the staging OLM helm
# chart filesystem structure.
cp -R "${ROOT_DIR}/staging/operator-lifecycle-manager/deploy/chart/" "${chartdir}"
cp "${ROOT_DIR}/staging/operator-lifecycle-manager/deploy/ocp/values.yaml" ${tmpdir}
ln -snf $(realpath --relative-to ${tmpdir} ${ROOT_DIR}/staging/api/pkg/operators/) ${crdsrcdir}
rm -rf ${crddir}/*

trap "rm -rf ${tmpdir}" EXIT

generate_crds() {
   echo "Generating OLM CRDs"
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
}

add_downstream_metadata_annotations() {
   local manifests_dir=$1

   echo "Adding release metadata annotations"
   for f in "${manifests_dir}"/*.yaml; do
      # check if the current file is the IBM cloud managed deployment copy
      if grep -q "ibm-cloud-managed" "${f}"; then
         # Note: OLM typically schedules pods on master nodes, which isn't
         # possible in the context of IBM cloud managed (i.e. ROKS) clusters
         # as there are no master nodes in the cluster. Ensure that the IBM
         # cloud managed deployment copy removes the master nodeSelector from the
         # deployment specification.
         ${YQ} w -d'*' --inplace --style=double "${f}" 'metadata.annotations['include.release.openshift.io/ibm-cloud-managed']' true
         ${YQ} d -d'*' --inplace "${f}" 'spec.template.spec.nodeSelector."node-role.kubernetes.io/master"'
         continue
      fi

      if [[ ! "$(basename "${f}")" =~ .*\.deployment\..* ]]; then
         ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/ibm-cloud-managed']' true
      fi

      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/self-managed-high-availability']' true
      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/single-node-developer']' true
   done
}

update_csv() {
   local csv=$1

   echo "Replacing the ${csv} CSV manifest"
   ${YQ} w --inplace "${csv}" --tag '!!bool' 'spec.cleanup.enabled' false
   ${YQ} w --inplace "${csv}" 'spec.customresourcedefinitions' {}
   ${YQ} w --inplace "${csv}" --style="" 'spec.install.spec.deployments[0].spec.template.spec.containers[0].ports[0].protocol' TCP
   ${YQ} w --inplace "${csv}" --style="" 'spec.install.spec.deployments[0].spec.template.metadata.creationTimestamp' null
   sed -i "s/'{}'/{}/g" "${csv}"
}

sanitize_manifests() {
   local manifest_dir=$1

   # requires gnu sed if on mac
   echo "Sanitizing YAML manifests in the ${manifest_dir} directory"
   find "${manifest_dir}" -type f -exec sed -i "/^#/d" {} \;
   find "${manifest_dir}" -type f -exec sed -i "1{/---/d}" {} \;
}

generate_crds
add_downstream_metadata_annotations "${ROOT_DIR}/manifests"
update_csv "${ROOT_DIR}/manifests/0000_50_olm_15-packageserver.clusterserviceversion.yaml"
sanitize_manifests "${ROOT_DIR}/manifests"
