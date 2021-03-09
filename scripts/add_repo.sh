#!/bin/bash

set -eu

case "$*" in
	*" -h"*|"-h"|*" --help"*|"--help"|"")
	echo "Add a new upstream repository to track"
	echo "usage: $0 <git remote url> [<branch/tag>]"
	exit 0
	;;
esac

source "$(dirname $0)/utils.sh"

# check if remote is valid
git ls-remote $1 

remote_url=$1

# remote_name is the same as repository name
remote_name=$(echo "${remote_url}" | sed 's!.*/\([^\/]*\)!\1!' | sed 's/.git$//')
if [ -z "${remote_name}" ]; then
	exit_on_error "cannot get repository name from ${remote_url}"
fi

update_tracked=false
add_subtree=false

tracked_info=$(grep "^${remote_name} " "${repo_list}") || update_tracked=true
# If remote isn't tracked, update list of tracked repositories later.
# Branch needs to be clean to stage the subtree
if ! ${update_tracked} ; then
	# repo entry present
	read -r tracked_name tracked_url <<<$(echo "${tracked_info}")
	
	if [ ${tracked_url} != ${remote_url} ]; then
		echo "remote ${remote_name} tracked with url ${tracked_url}"
		cleanup_and_reset_branch
		exit 0
	fi
fi

# add remote to git
local_url=$(git remote get-url "${remote_name}" 2>/dev/null) || \
	git remote add "${remote_name}" "${remote_url}" 

if [ -n "${local_url}" ] && [ "${local_url}" != "${remote_url}" ]; then
	exit_on_error "remote url ${remote_url} differs from tracked url ${local_url} for ${remote_name}"
fi

remote_ref=${2:-${tracked_ref:-master}}
remote_dir="${staging_dir}/${remote_name}"

# Create subtree at subtree_dir if it doesn't already exist
if [ ! -d "${remote_dir}" ]; then
	git rev-parse --symbolic-full-name --abbrev-ref HEAD >/dev/null 2>&1 || exit_on_error "invalid ref HEAD, cannot add subtree" $?

	git diff-index --quiet HEAD || exit_on_error "Git status not clean, aborting !!\\n\\n$(git status)" $?

	# add the subtree
	git remote update "${remote_name}"

	rel_remote_dir="$(realpath --relative-to ${repo_root} ${remote_dir})"

	git subtree add --prefix="${rel_remote_dir}" "${remote_name}" --squash "${remote_ref}"
	new_mod=$(cd ${remote_dir} && go list -m)

	for staged_dep in $(ls "${staging_dir}" | grep -v "^${remote_name}$"); do
		staged_mod=$(cd "${staging_dir}/${staged_dep}" && go list -m)
		grep "${staged_mod}" "${remote_dir}/go.mod" && sh -c "cd ${remote_dir} && \
								go mod edit -require ${staged_mod}@v0.0.0-00010101000000-000000000000 && \
								go mod edit -replace ${staged_mod}=../${staged_dep}"
		grep "${new_mod}" "${staging_dir}/${staged_dep}/go.mod" && sh -c "cd ${staging_dir}/${staged_dep} && \
								go mod edit -require ${new_mod}@v0.0.0-00010101000000-000000000000 && \
								go mod edit -replace ${new_mod}=../${remote_name} && \
								go mod vendor && \
								git add go.mod go.sum vendor"
	done

	rel_repo_root="$(realpath --relative-to ${remote_dir} ${repo_root})"
	sh -c "cd ${remote_dir} && \
		go mod edit -replace ${downstream_repo}=${rel_repo_root} && \
		go mod vendor && \
		git add go.mod go.sum vendor"
	
	# remove nested OWNERS file for openshift CI
	git rm "${remote_dir}/OWNERS"

	# find commit for tracked target to write to repo_list
	git fetch --tags "${remote_name}" +"${remote_ref}":refs/remotetags/"${remote_name}"/"${remote_ref}"
	remote_hash=$(git show-ref remotetags/${remote_name}/${remote_ref} -s)
	${update_tracked} && \
		echo "${remote_name} ${remote_url} ${remote_ref} ${remote_hash}" >> "${repo_list}" || \
		sed 's!\('"${remote_name}"' '"${remote_url}"'\).*!\1 '"${remote_ref}"' '"${remote_hash}"'!' "${repo_list}"
		
	git add "${repo_list}"

	git commit --amend --no-edit
	echo "Added new subtree ${remote_dir}"
	add_subtree=true
elif ${update_tracked} ; then
	echo "${remote_dir} already exists, "
	echo "${remote_name} ${remote_url}" >> "${repo_list}"
	git add "${repo_list}"
	git commit -m "update tracked remotes for ${remote_name}"
	add_subtree=true
else
	echo "repository already present and tracked, nothing to do"
	cleanup_and_reset_branch
fi


if ${add_subtree} ; then
	# push to subtree dir

	FORK_REMOTE=${FORK_REMOTE:-origin}
	git diff --dirstat "${current_branch}".."${temp_branch}"

	echo ""
	echo "!!! Added a new subrepo, you can now make any needed updates to the build files and Makefile"
	echo ""
	echo "!!! To cherry-pick the changes to your original branch, run:"
	echo "  git checkout ${current_branch} && git cherry-pick -m 2 "'$('"git merge-base ${current_branch} ${temp_branch})..${temp_branch}"
	echo ""
	echo "!!! To merge the changes to your original branch, run:"
	echo "  git checkout ${current_branch} && git merge --squash -s recursive -X theirs -m 'Sync upstream ${remote_name} ${remote_ref}'" "${temp_branch}"
	echo "!!! Once the changes look good, you can push the changes to the remote repository with:"
	echo "  git push ${FORK_REMOTE} ${temp_branch}"
fi

#cleanup_and_reset_branch

exit 0

 
