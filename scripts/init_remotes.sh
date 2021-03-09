#!/bin/bash

case "$*" in
	*" -h"*|"-h"|*" --help"*|"--help")
	echo "Add the remotes in ${repo_list} to the local git repository"
	echo "usage: $0"
	exit 0
	;;
esac

repo_list="$(git rev-parse --show-toplevel)/scripts/tracked"

while read -r remote_name remote_url; do
	if git remote get-url "${remote_name}" &>/dev/null; then
		tracked_url=$(git remote get-url "${remote_name}")
		if [ "${tracked_url}" != "${remote_url}" ]; then
			echo -e "\e[91mremote ${remote_name} present but does not track ${remote_url}\e[0m"
		fi
	else
		git remote add "${remote_name}" "${remote_url}"
	fi
done<<<"$(awk '{print $1, $2;}' ${repo_list})"
