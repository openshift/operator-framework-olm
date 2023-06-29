#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -x

ROOT_DIR=$(dirname "${BASH_SOURCE[@]}")/..
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/common.sh"

pop_all=false

while getopts ':a' flag; do
  case "${flag}" in
    a) pop_all=true ;;
    \?) exit 1 ;;
    *) echo "unexpected option ${flag}"; exit 1 ;;
  esac
done
# Shift to non-option arguments
shift $((OPTIND-1))

remote="${1:-api}"
subtree_dir="staging/${2:-${remote}}"
cherrypick_set="${remote}.cherrypick"
remaining=$(wc -l < "${cherrypick_set}")

function pop() {
    rc=$(head -n 1 "${cherrypick_set}")
    if [[ ! $rc ]]; then
        printf 'nothing to pick'
        exit
    fi
    printf 'popping: %s\n' "${rc}"

    if ! git cherry-pick --allow-empty --keep-redundant-commits -Xsubtree="${subtree_dir}" "${rc}"; then
        # Always blast away the vendor directory given OLM/registry still commit it into source control.
        git rm -rf "${subtree_dir}"/vendor 2>/dev/null || true

        # Look for any deleted by us
        readarray -t deletes < <(git status --porcelain| grep -oP "^DU \K.*")
        for d in "${deletes[@]}"; do
            git rm "${d}"
        done
        echo "Done with deletes"

        # Handle other conflicts
        num_conflicts=$(git diff --name-only --diff-filter=U --relative | wc -l)
        while [[ $num_conflicts != 0 ]] ; do
            readarray -t files < <(git diff --name-only --diff-filter=U --relative)

            for f in "${files[@]}"; do

                # Note that this can be a problem if there are regressions! (e.g. 1.2 -> 1.1)
                if [[ ${f} == *"go.mod"* ]]; then
                    git diff "${subtree_dir}"/go.mod

                    git checkout --theirs "${subtree_dir}"/go.mod
                    pushd "${subtree_dir}"
                    go mod tidy
                    git add go.mod go.sum
                    popd
                else
                    git checkout --theirs "${f}"
                    git diff "${f}"
                    git add "${f}"
                fi
            done

            num_conflicts=$(git diff --name-only --diff-filter=U --relative | wc -l)
            echo "Number of merge conflicts remaining: $num_conflicts"
        done
        echo "Done with conflicts"

        if [[ -z $(git status --porcelain) ]]; then
            git commit --allow-empty
        else
            echo "Current cherry pick status: $(git status --porcelain)"
            git -c core.editor=true cherry-pick --continue
        fi
    fi

    # Did go.mod change?
    if ! git diff --quiet HEAD^ "${subtree_dir}"/go.mod; then
        git diff HEAD^ "${subtree_dir}"/go.mod
        pushd "${subtree_dir}"
        echo "Running BASH subshell: go.mod has changed, check for regressions!"
        echo -n '<ENTER> to continue, ^C to quit: '
        read
        popd
    fi

    # 1. Pop next commit off cherrypick set
    # 2. Cherry-pick
    # 3. Ammend commit
    # 4. Remove from cherrypick set
    if ! make vendor; then
        echo "Running BASH subshell: fix make vendor"
        echo -n '<ENTER> to continue, ^C to quit: '
        read
    fi
    git add "${subtree_dir}" "${ROOT_GENERATED_PATHS[@]}"
    git status
    git commit --amend --allow-empty --no-edit --trailer "Upstream-repository: ${remote}" --trailer "Upstream-commit: ${rc}"
    if ! make manifests; then
        echo "Running BASH subshell: fix make manifests"
        echo -n '<ENTER> to continue, ^C to quit: '
        read
    fi
    git add "${subtree_dir}" "${ROOT_GENERATED_PATHS[@]}"
    git status
    git commit --amend --allow-empty --no-edit

    tmp_set=$(mktemp)
    tail -n +2 "${cherrypick_set}" > "${tmp_set}"; cat "${tmp_set}" > "${cherrypick_set}"

    # Note: handle edge case where there's zero remaining to avoid
    # returning a non-zero exit code.
    (( --remaining )) || true
    printf '%d picks remaining (pop_all=%s)\n' "${remaining}" "${pop_all}"

    if [[ $pop_all == 'true' ]] && ((  remaining >  0 )); then
        pop
    fi
}

pop
