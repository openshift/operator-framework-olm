#!/usr/bin/env bash

cph=$(git rev-list -n 1 CHERRY_PICK_HEAD 2> /dev/null)

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
subtree_dir="staging/${remote}"
cherrypick_set="${remote}.cherrypick"
remaining=$(wc -l < "${cherrypick_set}")

function pop() {
    rc=$(head -n 1 "${cherrypick_set}")
    if [[ ! $rc ]]; then
        printf 'nothing to pick'
        exit
    fi
    printf 'popping: %s\n' "${rc}"

    if [[ ! $cph ]]; then
        git cherry-pick --allow-empty --keep-redundant-commits -Xsubtree="${subtree_dir}" "${rc}"
    else
        if [[ $cph != "${rc}" ]]; then
            printf 'unexpected CHERRY_PICK_HEAD:\ngot %s\nexpected: %s\n' "${cph}" "${rc}"
            exit
        fi
        printf 'cherry-pick in progress for %s\n' "${cph}"
        git add .

        if [[ -z $(git status --porcelain) ]]; then
            git commit --allow-empty
        else
            git cherry-pick --continue
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
