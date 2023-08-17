#!/usr/bin/env bash

if [ $# -eq 0 ]; then
    cat <<EOF
USAGE
    scripts/sync_pop_candidate.sh [-a] <cherry-pick>

OPTIONS
    -a            Apply all changes from the <cherry-pick>
    <cherry-pick> Filename (without .cherrypick extension) from which
                  commits are selected to be synced

DESCRIPTION
    Use this script to sync a selection of commits from the upstream
    repositories. This script is called by the sync.sh script to perform
    the actual downstream sync.

    The <cherry-pick> file is the basename of a file with a .cherrypick
    extension. Usually, it is the name of a remote repository, but can
    be any file that follows the .cherrypick format.

    Refer to the README.md file for additional information.
EOF
    exit 1
fi

set -o errexit
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[@]}")/..
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/common.sh"
RED=$(tput setaf 1)
RESET=$(tput sgr0)

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
    readarray -t rcs < <(echo "$rc" | tr " " "\n")
    remote="${rcs[1]}"
    subtree_dir="staging/${remote}"
    rc="${rcs[2]}"
    printf 'popping: %s\n' "${rc}"

    # Cherrypick the commit
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
        echo ""
        echo -e "Pausing script: ${RED}go.mod has changed, check for regressions!${RESET}"
        echo "Use another terminal window"
        echo -n '<ENTER> to continue, ^C to quit: '
        read
        popd
    fi

    # Update commit with make vendor
    if ! make vendor; then
        echo ""
        echo -e "Pausing script: ${RED}fix (or ignore) make vendor{$RESET}"
        echo "Use another terminal window"
        echo -n '<ENTER> to update commit and continue, ^C to quit: '
        read
    fi
    git add "${subtree_dir}" "${ROOT_GENERATED_PATHS[@]}"
    git commit --amend --allow-empty --no-edit

    # Update commit with make verify-manifests, this uses the proper OLM_VERSION
    if ! make verify-manifests; then
        echo ""
        echo -e "Pausing script: ${RED}fix (or ignore) make manifests${RESET}"
        echo "Use another terminal window"
        echo -n '<ENTER> to update commit and continue, ^C to quit: '
        read
    fi
    git add "${subtree_dir}" "${ROOT_GENERATED_PATHS[@]}"
    git commit --amend --allow-empty --no-edit

    # Update commit with make verify-nested-vendor
    if ! make verify-nested-vendor; then
        echo ""
        echo -e "Pausing script: ${RED}fix make (or ignore) verify-nested-vendors${RESET}"
        echo "Use another terminal window"
        echo -n '<ENTER> to update commit and continue, ^C to quit: '
        read
    fi
    git add "${subtree_dir}" "${ROOT_GENERATED_PATHS[@]}"
    git commit --amend --allow-empty --no-edit --trailer "Upstream-repository: ${remote}" --trailer "Upstream-commit: ${rc}"
    # need to add these trailers for make verify-commits

    # Remove from cherrypick set, now that the trailers are added, it's effectively complete
    tmp_set=$(mktemp)
    tail -n +2 "${cherrypick_set}" > "${tmp_set}"; cat "${tmp_set}" > "${cherrypick_set}"

    # Verify commit with make verify-commits - this should not error out
    if ! make verify-commits; then
        echo ""
        echo -e "Pausing script: ${RED}fix make verify-commits${RESET}"
        echo "Use another terminal window"
        echo -n '<ENTER> to continue, ^C to quit: '
        read
    fi
    # At this point "make verify" would pass

    # Note: handle edge case where there's zero remaining to avoid
    # returning a non-zero exit code.
    (( --remaining )) || true
    printf '%d picks remaining (pop_all=%s)\n' "${remaining}" "${pop_all}"

    if [[ $pop_all == 'true' ]] && ((  remaining >  0 )); then
        pop
    fi
}

pop
