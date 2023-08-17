#!/usr/bin/env bash

if [ $# -eq 0 ]; then
    cat <<EOF
USAGE
    scripts/sync_get_candidates.sh <remote>

OPTIONS
    <remote>   Remote repository to search for commits

DESCRIPTION
    This script is used to automatically create cherrypick files for use
    by the sync_pop_candidate.sh script. This script is called by the
    sync.sh script to gather the commits to be part of the downsteam
    sync.

    Refer to the README.md file for additional information.
EOF
    exit 1
fi



set -o errexit
set -o pipefail

num_commits=256

set +u
while getopts 'n:' flag; do
  case "${flag}" in
    n) num_commits=${optarg}; shift ;;
    \?) exit 1 ;;
    *) echo "unexpected option ${flag}"; exit 1 ;;
  esac

  shift
done
set -u

remote="${1:-api}"
branch="${2:-master}"

# slurp command output by newline, creating an array
mapfile -t remote_commits < <(git rev-list --topo-order --no-merges -n "${num_commits}" "${remote}/${branch}" | tac)

picked=0
cherrypick_set="${remote}.cherrypick"
: > "${cherrypick_set}" # clear existing file
for rc in "${remote_commits[@]}"; do
    if [[ -z $(git log -n 1 --no-merges --grep "${rc}" HEAD) && -z $(grep "${rc}" "${remote}.blacklist") ]]; then
        git show -s --format="%cI ${remote} %H" "${rc}" >> "${cherrypick_set}"
        (( ++picked ))
    fi
done

printf '%d cherry-pick candidate(s) written to %s\n' "${picked}" "${cherrypick_set}"
printf 'run "pop_candidate.sh -a %s" to cherry-pick all\n' "${remote}"
