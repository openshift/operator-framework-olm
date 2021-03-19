#!/bin/bash

set -E

case "$*" in
	*" -h"*|"-h"|*" --help"*|"--help"|"")
	echo "Pull from specified upstream staged repository. Syncs with upstream master if branch isn't specified."
	echo "usage: $0 <remote> [<branch/ref>]"
	exit 0
	;;
esac

if [ $# -lt 1 ]; then
	echo "Pull from specified upstream staged repository. Syncs with upstream master if branch isn't specified."
	echo "usage: $0 <remote> [<branch/ref>]"
	exit 0
fi

source "$(dirname $0)/utils.sh"

remote_url=$(git remote get-url "$1" 2>/dev/null) && remote_name=$1 || remote_url=$1
if $(git ls-remote "$1" --quiet 2>/dev/null) && [ -z "${remote_name}" ]; then # $1 is a url
	# Attempt to get staged remote name from tracked list, default to repository name.
	tracked_name=$(cat ${repo_list} | grep " ${remote_url}" | awk '{print $1;}')
	tracked_name=${tracked_name:-$(echo "${remote_url}" | sed 's!.*/\([^\/]*\)!\1!' | sed 's/.git$//')}
	if [ $(echo "${tracked_name}" | wc -w) -eq 1 ]; then
		tracked_url=$(git remote get-url "${tracked_name}" 2>/dev/null) || true
		if [ -n "${tracked_url}" ] && [ "${tracked_url}" != "${remote_url}" ]; then
			exit_on_error "cannot add ${remote_url}; default remote name ${tracked_name} already tracked by ${tracked_url}" 1
		fi 
		remote_name="${tracked_name}"
	else
		exit_on_error "ambiguous remote url ${remote_url} in ${repo_list}, tracked by: ${tracked_name}" 1
	fi
else
	if [ -z "${remote_name}" ]; then
		remote_name=$1
		remote_url=$(cat "${repo_list}" | grep "^${remote_name} " | awk '{print $2;}')
		if [ $(echo "${remote_name}" | wc -w) -ne 1 ]; then
			exit_on_error "invalid remote ${remote_name}: not tracked in git or ${repo_list}" 1
		fi
	fi

	tracked_url=$(cat "${repo_list}" | grep "^${remote_name} " | awk '{print $2;}') || true
	if [ -n "${tracked_url}" ] && [ "${tracked_url}" != "${remote_url}" ]; then
		exit_on_error "ambiguous remote ""${remote_name}"": expected: ""${remote_url}""; tracking: ""${tracked_url}" 1
	fi
fi

if [ -z "${remote_name}" ] || [ -z "${remote_url}" ]; then
	exit_on_error "empty remote: ${remote_name}; url: ${remote_url}" 1
fi

# track url if not already tracked.
git remote get-url "${remote_name}" 2>/dev/null || \
	git remote add "${remote_name}" "${remote_url}"

remote_dir="${staging_dir}/${remote_name}"

tracked_ref=$(cat "${repo_list}" | grep "^${remote_name} " | awk '{print $3;}')
remote_ref=${2:-${tracked_ref:-master}}

git rev-parse --symbolic-full-name --abbrev-ref HEAD >/dev/null 2>&1 || exit_on_error "invalid ref HEAD, cannot add subtree" $?
git diff-index --quiet HEAD || exit_on_error "Git status not clean, aborting !!\\n\\n$(git status)" $?

rel_remote_dir="$(realpath --relative-to ${repo_root} ${remote_dir})"
git fetch -t "${remote_name}" "${remote_ref}"
if [ ! -d "$remote_dir" ]; then
	#subtree add
	git subtree add -q --prefix="${rel_remote_dir}" "${remote_name}" --squash "${remote_ref}"

	new_mod=$(cd ${remote_dir} && go list -m)
	sh -c "go mod edit -require ${new_mod}@v0.0.0-00010101000000-000000000000 && \
			go mod edit -replace ${new_mod}=./$(realpath --relative-to ${repo_root} ${remote_dir}) && \
			git add go.mod go.sum"
	for staged_dep in $(find "${staging_dir}" -mindepth 1 -maxdepth 1 ! -path "${remote_dir}"); do
		staged_mod=$(cd "${staged_dep}" && go list -m -mod=mod)
		grep "${staged_mod}" "${remote_dir}/go.mod" | grep -v "^module ${staged_mod}" && sh -c "cd ${remote_dir} && \
								go mod edit -require ${staged_mod}@v0.0.0-00010101000000-000000000000 && \
								go mod edit -replace ${staged_mod}=$(realpath --relative-to ${remote_dir} ${staged_dep})"
		grep "${new_mod}" "${staged_dep}/go.mod" | grep -v "^module ${new_mod}" && sh -c "cd ${staged_dep} && \
								go mod edit -require ${new_mod}@v0.0.0-00010101000000-000000000000 && \
								go mod edit -replace ${new_mod}=$(realpath --relative-to ${staged_dep} ${remote_dir}) && \
								go mod tidy -e ; \
								git add go.mod go.sum"
	done
else
	split_branch="${remote_name}-$(date +%s)"
	git subtree split -q --prefix="${rel_remote_dir}" --rejoin -b "${split_branch}" 

	git subtree pull -q --squash -m "Sync upstream ${remote_name} ${remote_ref}" --prefix="${rel_remote_dir}" "${remote_name}" "${remote_ref}"
	git branch -D "${split_branch}" || true
	for staged_dep in $(find "${staging_dir}" -mindepth 1 -maxdepth 1 ! -path "${remote_dir}"); do
		staged_mod=$(cd "${staged_dep}" && go list -m -mod=mod)
		grep "${staged_mod}" "${remote_dir}/go.mod" | grep -v "^module ${staged_mod}" && sh -c "cd ${remote_dir} && \
								go mod edit -require ${staged_mod}@v0.0.0-00010101000000-000000000000 && \
								go mod edit -replace ${staged_mod}=$(realpath --relative-to ${remote_dir} ${staged_dep})"
	done
fi

sh -c "cd ${remote_dir} && \
	git rm -rq vendor && \
	go mod tidy -e && \
	git add go.mod go.sum"

# remove nested OWNERS file for openshift CI
git rm "${remote_dir}/OWNERS"

# find commit for tracked target to write to repo_list
git fetch --tags "${remote_name}" +"${remote_ref}":refs/remotetags/"${remote_name}"/"${remote_ref}"
remote_hash=$(git show-ref remotetags/${remote_name}/${remote_ref} -s)
grep "^${remote_name} " "${repo_list}" && \
		sed -i 's!\('"${remote_name}"' '"${remote_url}"'\).*!\1 '"${remote_ref}"' '"${remote_hash}"'!' "${repo_list}" || \
		echo "${remote_name} ${remote_url} ${remote_ref} ${remote_hash}" >> "${repo_list}"
git add "${repo_list}" "${staging_dir}"

git commit --amend --no-edit

FORK_REMOTE=${FORK_REMOTE:-origin}
git diff --dirstat "${current_branch}".."${temp_branch}"

git checkout "${current_branch}"
git merge --squash -s recursive -X theirs -m "Sync upstream ${remote_name} ${remote_ref}" "${temp_branch}"
git commit -m "Sync upstream ${remote_name} ${remote_ref}"
git branch -D "${temp_branch}"
echo ""
echo "!!! Upstream merge complete!"
echo ""
echo "!!! You can now inspect the branch."
echo ""
echo "!!! Once the changes look good, you can push the changes to the remote repository with:"
echo "  git push ${FORK_REMOTE} ${current_branch}"
