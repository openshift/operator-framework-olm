#! /bin/bash

set -o errexit
set -o nounset
set -o pipefail

: "${WHAT:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
TARGET_NAME=${TARGET_NAME:="unit"}

# TODO(tflannag): Do we need to trap anything here? Maybe remove staging/*/vendor during SIGINT?
pushd "${ROOT_DIR}/staging/${WHAT}"

if [[ ! -d ${ROOT_DIR}/staging/${WHAT}/vendor ]]; then
    # Note(tflannag): We don't introduce nested vendor packages into source control,
    # and this script will fail when attempting to populate a vendor directory using
    # `go test ...`, so vendor first, then run unit tests.
    #
    # This is likely a poor strategy to maintain going forward as there's a chance we're testing
    # against dependencies that don't match what we build with but that should be fine for now.
    # We'll likely want to migrate towards having a dedicated test/e2e package that
    # runs downstream-specific OLM and registry tests instead of relying on the test packages that
    # we pull in from the staging equivalent. In the case of OSBS, where we can't `go get ...` packages from
    # external sources, so this should be fine in the testing scenario, but we need the vendor/ when building
    # binaries due to the restricted environment. Fortunately, it looks we're able to build from the root directory,
    # so it's sufficient for now.
    echo "Populating nested staging vendor directory"
    go mod vendor && go mod tidy
fi

make ${TARGET_NAME}
popd
