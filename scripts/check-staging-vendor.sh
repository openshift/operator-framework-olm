#! /bin/bash

set -o errexit
set -o pipefail
set -o nounset

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
STAGE_BASE_DIR=${ROOT_DIR}/staging
STAGE_DIRS=( "api" "operator-lifecycle-manager" "operator-registry" )

for repo in "${STAGE_DIRS[@]}"; do
    repo_path="${STAGE_BASE_DIR}/$repo"

    echo "Checking for unstaged go.mod,go.sum changes in ${repo_path}"
    pushd "${repo_path}"

    go mod tidy && go mod vendor && go mod verify

    if ! git diff --quiet go.mod go.sum; then
        echo "The staging/${repo} has unsynced [go.mod,go.sum] changes"
        exit 1
    fi

    popd
done

