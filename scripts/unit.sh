#! /bin/bash

set -o errexit
set -o nounset
set -o pipefail

: "${WHAT:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
TARGET_NAME=${TARGET_NAME:="unit"}

pushd "${ROOT_DIR}/staging/${WHAT}"
make ${TARGET_NAME}
popd
