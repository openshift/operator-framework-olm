#! /bin/bash

# TODO(tflannag): We'll likely need to transistion towards using
# a set of root e2e tests once upstream removes some of the
# downstream-specific e2e tests, but this should be sufficient
# until that happens.

set -o errexit
set -o nounset
set -o pipefail

: "${KUBECONFIG:?}"
: "${WHAT:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..

function run_test() {
    local staging_dir=$1

    pushd "${ROOT_DIR}/staging/${staging_dir}"
    echo "Running ${staging_dir} e2e tests"
    make e2e
    popd
}

run_test "${WHAT}"
