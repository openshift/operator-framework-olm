#! /bin/bash

set -o pipefail
set -o nounset
set -o errexit

: "${KUBECONFIG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
./${ROOT_DIR}/bin/olm \
    --namespace openshift-operator-lifecycle-manager \
    --writeStatusName operator-lifecycle-manager \
    --writePackageServerStatusName operator-lifecycle-manager-packageserver

# TODO: handle tls key/certs