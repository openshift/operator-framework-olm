#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

ROOT_DIR=$(dirname "${BASH_SOURCE[@]}")/..
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/common.sh"

function err() {
    echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
}

function info() {
    echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*"
}

function verify_staging_sync() {
    local remote="${1}"
    local downstream_commit="${2}"
    local staging_dir="staging/${remote}"

    local outside_staging
    outside_staging="$(git show --name-only "${downstream_commit}" -- ":!${staging_dir}" "${KNOWN_GENERATED_PATHS[@]}")"
    if [[ -n "${outside_staging}" ]]; then
        err "downstream staging commit ${downstream_commit} changes files outside of ${staging_dir}, vendor, and manifests directories"
        err "${outside_staging}"
        err "hint: factor out changes to these directories into a standalone commit"
        return 1
    fi
}

function verify_downstream_only() {
    local downstream_commit="${1}"

    local inside_staging
    inside_staging="$(git show --name-only "${downstream_commit}" -- staging)"
    if [[ -n "${inside_staging}" ]]; then
        err "downstream non-staging commit ${downstream_commit} changes staging"
        err "${inside_staging}"
        err "only staging commits (i.e. from an upstream cherry-pick) may change staging"
        return 1
    fi
}

function upstream_ref() {
    local downstream_commit="${1}"

    local log
    log="$(git log -n 1 "${downstream_commit}")"

    local -a upstream_repos
    mapfile -t upstream_repos < <(echo "${log}" | grep 'Upstream-repository:' | awk '{print $2}')

    local -a upstream_commits
    mapfile -t upstream_commits < <(echo "${log}" | grep 'Upstream-commit:' | awk '{print $2}')

    if (( ${#upstream_repos[@]} < 1 && ${#upstream_commits[@]} < 1 )); then
        # no upstream commit referenced
        return 0
    fi


    local invalid
    invalid=false
    if (( ${#upstream_repos[@]} != 1 )); then
        err "downstream staging commit ${downstream_commit} references an invalid number of repos: ${#upstream_repos[@]}"
        err "staging commits must reference a single upstream repo"
        invalid=true
    fi

    if (( ${#upstream_commits[@]} != 1 )); then
        err "downstream staging commit ${downstream_commit} references an invalid number of upstream commits: ${#upstream_commits[@]}"
        err "staging commits must reference a single upstream commit"
        invalid=true
    fi

    if [[ "${invalid}" == true ]]; then
        return 1
    fi

    if git branch -r --contains "${upstream_commits[0]}" | grep -vq "${upstream_repos[0]}/.*"; then
        err "downstream staging commit ${downstream_commit} references upstream commit ${upstream_commits[0]} that doesn't exist in ${upstream_repos[0]}"
        err "staging commits must reference a repository containing the given upstream commit"
        return 1
    fi

    echo "${upstream_repos[0]}"
    echo "${upstream_commits[0]}"
}

function fetch_remotes() {
    local -a remotes
    remotes=(
        'api'
        'operator-registry'
        'operator-lifecycle-manager'
    )
    for r in "${remotes[@]}"; do
        local url
        url="https://github.com/operator-framework/${r}.git"

        git remote add "${r}" "${url}" 2>/dev/null || git remote set-url "${r}" "${url}"
        git fetch -q "${r}" master
    done
}

function main() {
    fetch_remotes || { err "failed to fetch remotes" &&  exit 1; }

    local target_branch="${1:-master}"

    # get all commits we're introducing into the target branch
    local -a new_commits
    mapfile -t new_commits < <(git rev-list --no-merges HEAD "^${target_branch}")
    info "detected ${#new_commits[@]} commit(s) to verify"

    local -a sr
    local short
    for commit in "${new_commits[@]}"; do
        short="${commit:0:7}"
        info "verifying ${short}..."
        info "$(git log -n 1 "${commit}")"

        sr=( $(upstream_ref "${commit}")  ) || exit 1
        if (( ${#sr[@]} < 2 )); then # the ref contains a tuple if the values were properly parsed from the commit message
            # couldn't find upstream cherry-pick reference in the commit message
            # assume it's downstream-only commit
            info "${short} does not reference upstream commit, verifying as downstream-only"
            verify_downstream_only "${commit}" || exit 1
        else
            # found upstream cherry-pick reference in commit message
            # verify as an upstream sync
            info "${short} references upstream commit (${sr[*]}), verifying as upstream staging sync"
            verify_staging_sync "${sr[0]}" "${commit}" || exit 1
        fi

    done
}

main "$@"
test $? -gt 0 || echo "Successfully validated all commits"
