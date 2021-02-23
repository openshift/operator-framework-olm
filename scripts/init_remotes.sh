#!/bin/bash -x

case "$*" in
	*" -h"*|"-h"|*" --help"*|"--help")
	echo "Add the remotes in $repo_list to the local git repository"
	echo "usage: $0"
	exit 0
	;;
esac

repo_list="scripts/tracked"
repo_root=$(git rev-parse --show-toplevel)

cd $repo_root

while read -r line; do
	remote_name=$(echo "$line" | awk '{print $1}')
	remote_url=$(echo "$line" | awk '{print $2}')
	if git remote get-url $remote_name &>/dev/null; then
		if [ $(git remote get-url $remote_name) != "$remote_url" ]; then
			echo -e "\e[91mremote $remote_name present but does not track $remote_url\e[0m"
		fi
	else
		git remote add $remote_name $remote_url
	fi
done <$repo_list
