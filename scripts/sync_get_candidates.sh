#!/bin/bash

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
    if [[ -z $(git log -n 1 --no-merges --grep "${rc}" HEAD) ]]; then
        printf '%s\n' "${rc}" >> "${cherrypick_set}"
        (( ++picked ))
    fi
done

printf '%d cherry-pick candidate(s) written to %s\n' "${picked}" "${cherrypick_set}"
printf 'run "pop_candidate.sh -a %s" to cherry-pick all\n' "${remote}"
