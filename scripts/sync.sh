#! /bin/bash

set -o nounset
set -o errexit
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/common.sh"

SYNC_BRANCH_NAME="sync-$(printf '%(%Y-%m-%d)T\n' -1)"

add_remote() {
    echo "Adding upstream remotes if they don't already exist"
    git config remote.api.url >&- || git remote add api https://github.com/operator-framework/api
    git config remote.operator-registry.url >&- || git remote add operator-registry https://github.com/operator-framework/operator-registry
    git config remote.operator-lifecycle-manager.url >&- || git remote add operator-lifecycle-manager https://github.com/operator-framework/operator-lifecycle-manager
    git config remote.upstream.url >&- || git remote add upstream https://github.com/openshift/operator-framework-olm
}

fetch_remote() {
    git fetch upstream

    echo "Fetching upstream remotes"
    for remote in "${UPSTREAM_REMOTES[@]}"; do
        git fetch "$remote"
    done
}

new_candidate_branch() {
    echo "Creating a sync branch if it doesn't already exist"
    git checkout -b "$SYNC_BRANCH_NAME" master 2>/dev/null || git checkout "$SYNC_BRANCH_NAME"
}

candidates() {
    # TODO: add support for only collecting a single remote.
    echo "Collecting all upstream commits since last sync"
    for remote in "${UPSTREAM_REMOTES[@]}"; do
        "${ROOT_DIR}"/scripts/sync_get_candidates.sh "$remote"
    done

    # Create uber cherry-pick list
    cat *.cherrypick | sort > all.cherrypick
    echo "Number of commits to cherrypick: $(cat all.cherrypick | wc -l)"
}

pop() {
    echo "Applying all upstream commit candidates"
    "${ROOT_DIR}"/scripts/sync_pop_candidate.sh -a "all"
}

check_local_branch_commit_diff() {
    commits_ahead=$(git rev-list master..HEAD | wc -l)

    if [[ "$commits_ahead" -gt 1 ]]; then
        # TODO: automatically open a new pull request here.
        echo "The local sync branch is $commits_ahead commits ahead of the master branch"
    else
        echo "No sync PR is needed as the master branch is up-to-date"
    fi
}

main() {
    add_remote
    fetch_remote
    new_candidate_branch
    candidates
    pop
    check_local_branch_commit_diff
}

script_help() {
    cat <<EOF
USAGE
    scripts/sync.sh

DESCRIPTION
    Use this script to bulk sync from the upstream repositories.

    There are no arguments to this script.

    Refer to the README.md file for additional information.
EOF
    exit 1
}

# no arguments are required, look for any help-type arguments and print out help
for var in "$@"; do
    if [ "${var}" == "-h" -o "${var}" == "--help" ]; then
        script_help
    fi
done

main
