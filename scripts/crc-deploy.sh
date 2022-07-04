#!/usr/bin/env bash

set -e

# This script manages the deployment of downstream olm on a local CRC cluster
# It will deploy the local images: olm:test and opm:test
# Both built with make build/olm-container and build/registry-container respectively

echo "Deploying OLM to CRC"

# push_images opens a kubectl port-forward session to the crc registry service
# appropriately tags the locally built olm images
# and pushes them to the global registry "openshift"
function push_images {
  # local images
  LOCAL_OLM_IMAGE=${1-olm:test}
  LOCAL_OPM_IMAGE=${2-opm:test}

  # push images to the global crc registry "openshift"
  # so that they can be pulled from any other namespace in the cluster
  CRC_GLOBAL_REGISTRY="openshift"
  CRC_REGISTRY="default-route-openshift-image-registry.apps-crc.testing"

  # CRC destined images
  CRC_OLM_IMAGE="${CRC_REGISTRY}/${CRC_GLOBAL_REGISTRY}/olm:test"
  CRC_OPM_IMAGE="${CRC_REGISTRY}/${CRC_GLOBAL_REGISTRY}/opm:test"

  # CRC registry coordinates
  OPENSHIFT_REGISTRY_NAMESPACE="openshift-image-registry"
  OLM_NAMESPACE="openshift-operator-lifecycle-manager"
  IMAGE_REGISTRY_SVC="image-registry"
  IMAGE_REGISTRY_PORT=5000

  # Login to the CRC registry
  oc whoami -t | docker login "${CRC_REGISTRY}" --username user --password-stdin

  # Tag and push olm image
  echo "Pushing olm image"
  docker tag "${LOCAL_OLM_IMAGE}" "${CRC_OLM_IMAGE}"
  docker push "${CRC_OLM_IMAGE}"

  # Tag and push registry image
  echo "Pushing registry image"
  docker tag "${LOCAL_OPM_IMAGE}" "${CRC_OPM_IMAGE}"
  docker push "${CRC_OPM_IMAGE}"

  # Create image streams
  echo "Creating image streams: ${CRC_OLM_IMAGE} ${CRC_OPM_IMAGE}"
  OLM_OPENSHIFT=image-registry.openshift-image-registry.svc:5000/openshift/olm:test
  OPM_OPENSHIFT=image-registry.openshift-image-registry.svc:5000/openshift/opm:test
  oc import-image olm --from="${OLM_OPENSHIFT}" --confirm
  oc import-image opm --from="${OPM_OPENSHIFT}" --confirm
}

# make_manifest_patches takes in two parameters
# OLM_IMG: the internal registry istag for the olm image
# OPM_IMG: the internal registry istag for the opm image
# it creates a helm values files and other manifest patch files applied by
# scripts/generate_crds_manifests.sh
function make_manifest_patches {
  OLM_IMG=${1?Error:\ olm image undefined}
  OPM_IMG=${2?Error:\ opm image undefined}
  SED_OPTS=(-e 's#OLM_IMAGE#'"${OLM_IMG}"'#g' -e 's#OPM_IMAGE#'"${OPM_IMG}"'#g')

  # helm values file
  VALUES_TEMPLATE=$(mktemp)
  cat << EOF > "${VALUES_TEMPLATE}"
olm:
  image:
    ref: OLM_IMAGE
catalog:
  commandArgs: --configmapServerImage=OPM_IMAGE
  opmImageArgs: --opmImage=OPM_IMAGE
  image:
    ref: OLM_IMAGE
package:
  image:
    ref: OLM_IMAGE
EOF

  sed "${SED_OPTS[@]}" "${VALUES_TEMPLATE}" > "${CRC_E2E_VALUES}"

  # psm operator patch
  PSM_OPERATOR_PATCH_TEMPLATE=$(mktemp)
  cat << EOF > "${PSM_OPERATOR_PATCH_TEMPLATE}"
- command: update
  path: spec.template.spec.containers[0].image
  value: OLM_IMAGE
- command: update
  path: spec.template.spec.containers[0].env[1].value
  value: OLM_IMAGE
EOF

  sed "${SED_OPTS[@]}" "${PSM_OPERATOR_PATCH_TEMPLATE}" > scripts/psm-operator-deployment.crc.e2e.patch.yaml

  # collect-profiles patch
  COLLECT_PROFILE_PATCH_TEMPLATE=$(mktemp)
  cat << EOF > "${COLLECT_PROFILE_PATCH_TEMPLATE}"
- command: update
  path: spec.jobTemplate.spec.template.spec.containers[0].image
  value: OLM_IMAGE
EOF

  sed "${SED_OPTS[@]}" "${COLLECT_PROFILE_PATCH_TEMPLATE}" > scripts/collect-profiles.crc.e2e.patch.yaml
}

# YQ for applying yaml patches
YQ="go run ./vendor/github.com/mikefarah/yq/v3/"

# CRC_E2E_VALUES is the name of the help values file
# used to populate manifest templates with the locally built olm images
export CRC_E2E_VALUES="values-crc-e2e.yaml"

# Set kubeconfig to CRC if necessary
export KUBECONFIG=${KUBECONFIG:-${HOME}/.crc/machines/crc/kubeconfig}
KUBE_ADMIN_USER=$(crc console --credentials -o json | jq -r .clusterConfig.adminCredentials.username)
KUBE_ADMIN_PASSWORD=$(crc console --credentials -o json | jq -r .clusterConfig.adminCredentials.password)

# login to crc
echo "Logging in as kubeadmin"
oc login -u "${KUBE_ADMIN_USER}" -p "${KUBE_ADMIN_PASSWORD}" > /dev/null

# Scale down CVO to stop changes from returning to stock configuration
echo "Scaling down CVO"
oc scale --replicas 0 -n openshift-cluster-version deployments/cluster-version-operator

echo "Allow cluster wide access to openshift repository"
oc policy add-role-to-group system:image-puller system:serviceaccounts --namespace=openshift

# push olm images to crc global repository
push_images olm:test opm:test

SKIP_MANIFESTS=${SKIP_MANIFESTS:-0}
if [ "${SKIP_MANIFESTS}" = 0 ]; then
  # Create values and patches files
  # Get images with the specific shas
  OLM_IMG="$(oc get istag/olm:latest -o json | jq -r .image.dockerImageReference)"
  OPM_IMG="$(oc get istag/opm:latest -o json | jq -r .image.dockerImageReference)"

  make_manifest_patches "${OLM_IMG}" "${OPM_IMG}"

  # Build e2e manifests
  echo "Generating manifests"
  # CRC_E2E_VALUES is already set in this script and exported
  # If set, it will include the file referenced by it in the helm template command
  # and update the manifests to use the locally built images
  ./scripts/generate_crds_manifests.sh

  # Apply patches
  ${YQ} write --inplace -s scripts/psm-operator-deployment.crc.e2e.patch.yaml manifests/0000_50_olm_06-psm-operator.deployment.yaml
  ${YQ} write --inplace -s scripts/collect-profiles.crc.e2e.patch.yaml manifests/0000_50_olm_07-collect-profiles.cronjob.yaml

  echo "Deploying OLM"
  find_flags=(-regex ".*\.yaml" -not -regex ".*\.removed\.yaml" -not -regex ".*\.ibm-cloud-managed\.yaml")

  # Use the numbered ordering in the manifest file names to delete and deploy the manifests in order
  echo "Replacing manifests"
  find manifests "${find_flags[@]}" | sort | while read -r manifest; do
    echo "Deleting ${manifest}"
    set +e
    kubectl replace -f "${manifest}"
    set -e
  done
fi

# Force recreation of olm pods
echo "Restarting OLM pods"
kubectl delete pod --all -n "${OLM_NAMESPACE}"

# Wait for deployments to be available
SKIP_WAIT_READY=${SKIP_WAIT_READY:-0}
if [ "${SKIP_WAIT_READY}" = 0 ]; then
  echo "Waiting on deployments to be ready"
  for DEPLOYMENT in $(oc get deployments --no-headers=true -n "${OLM_NAMESPACE}" | awk '{ print $1 }'); do
    echo "Waiting for ${DEPLOYMENT}"
    kubectl wait --for=condition=available --timeout=120s "deployment/${DEPLOYMENT}" -n "${OLM_NAMESPACE}"
  done
fi

echo "Done"

exit 0
