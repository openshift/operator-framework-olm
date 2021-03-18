#! /bin/bash

# TODO(tflannag): We'll likely need to transistion towards using
# a set of root e2e tests once upstream removes some of the
# downstream-specific e2e tests, but this should be sufficient
# until that happens.

set -o errexit
set -o nounset
set -o pipefail

: "${KUBECONFIG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..

function run_olm_tests() {
    pushd "${ROOT_DIR}/staging/operator-lifecycle-manager"

    echo "Running OLM e2e tests"
    go test \
        -mod=vendor \
        -v \
        -failfast \
        -timeout 150m \
        ./test/e2e/... \
        -namespace=openshift-operators \
        -kubeconfig="${KUBECONFIG}" \
        -olmNamespace=openshift-operator-lifecycle-manager \
        -dummyImage=bitnami/nginx:latest \
        -ginkgo.flakeAttempts=3

    popd
}

function run_registry_tests() {
    pushd "${ROOT_DIR}/staging/operator-registry"

    echo "Running registry e2e tests"
    go run -mod=vendor github.com/onsi/ginkgo/ginkgo --v --randomizeAllSpecs --randomizeSuites --race -tags "json1" ./test/e2e

    popd
}

run_registry_tests
