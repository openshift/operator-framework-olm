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
    local path_to_staging_dir="${ROOT_DIR}/staging/${staging_dir}"

    pushd "${path_to_staging_dir}"

    if [[ ! -d ${path_to_staging_dir}/vendor ]]; then
        echo "Populating nested vendor directory"
        go mod tidy && go mod vendor && go mod verify
    fi

    echo "Running ${staging_dir} e2e tests"
    make e2e
    popd
}

run_test "${WHAT}"
