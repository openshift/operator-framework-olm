#!/usr/bin/env bash

set -o errexit
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[@]}")/..
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/common.sh"

pop_all=true

set +u
while getopts 'a:' flag; do
  case "${flag}" in
    a) pop_all=true; shift ;;
    \?) exit 1 ;;
    *) echo "unexpected option ${flag}"; exit 1 ;;
  esac

  shift
done
set -u

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

        num_conflicts=$(git diff --name-only --diff-filter=U --relative | wc -l)
        while [[ $num_conflicts != 0 ]] ; do
            file=$(git diff --name-only --diff-filter=U --relative)

            if [[ $file == *"go.mod"* ]]; then
                git diff "${subtree_dir}"/go.mod

                git checkout --theirs "${subtree_dir}"/go.mod
                pushd "${subtree_dir}"
                go mod tidy
                git add go.mod go.sum
                popd
            else
                git checkout --theirs "$file"
                git diff "$file"
                git add "$file"
            fi

            num_conflicts=$(git diff --name-only --diff-filter=U --relative | wc -l)
            echo "Number of merge conflicts remaining: $num_conflicts"
        done

        if [[ -z $(git status --porcelain) ]]; then
            git commit --allow-empty
        else
            echo "Current cherry pick status: $(git status --porcelain)"
            git -c core.editor=true cherry-pick --continue
        fi
    fi


    # 1. Pop next commit off cherrypick set
    # 2. Cherry-pick
    # 3. Ammend commit
    # 4. Remove from cherrypick set
    make vendor
    make manifests
    git add "${subtree_dir}" "${ROOT_GENERATED_PATHS[@]}"
    git status
    git commit --amend --allow-empty --no-edit --trailer "Upstream-repository: ${remote}" --trailer "Upstream-commit: ${rc}"

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
