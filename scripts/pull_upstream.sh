#!/bin/bash
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

remote_url=$(git remote get-url "$1")
remote_name="$(cat ${repo_list} | grep ${remote_url} | awk '{print $1;}')"
remote_dir="${staging_dir}/${remote_name}"
split_branch="${remote_name}-$(date +%s)"

if [ -z $2 ]; then
	tracked_ref=$(grep "^${remote_name} " "${repo_list}" | awk '{print $3;}')
fi
remote_ref=${2:-${tracked_ref:-master}}

if [ "$remote_dir" = "${staging_dir}/" ] || [ ! -d "${remote_dir}" ]; then
	echo "Missing remote from ${repo_list}"
	exit 1
fi

rel_remote_dir="$(realpath --relative-to ${repo_root} ${remote_dir})"

git fetch -t "${remote_name}" "${remote_ref}"
git subtree split --prefix="${rel_remote_dir}" --rejoin -b "${split_branch}" 

git subtree pull --squash -m "Sync upstream ${remote_name} ${remote_ref}" --prefix="${rel_remote_dir}" "${remote_name}" "${remote_ref}"
git branch -D "${split_branch}" || true

for staged_dep in $(ls "${staging_dir}" | grep -v "^${remote_name}$"); do
	staged_mod=$(cd ${staging_dir}/${staged_dep} && go list -m)
	grep "${staged_mod}" "${remote_dir}/go.mod" && sh -c "cd ${remote_dir} && \
							go mod edit -require ${staged_mod}@v0.0.0-00010101000000-000000000000 && \
							go mod edit -replace ${staged_mod}=../${staged_dep}"
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
grep "^${remote_name} " "${repo_list}" && \
		sed -i 's!\('"${remote_name}"' '"${remote_url}"'\).*!\1 '"${remote_ref}"' '"${remote_hash}"'!' "${repo_list}" || \
		echo "${remote_name} ${remote_url} ${remote_ref} ${remote_hash}" >> "${repo_list}"
git add "${repo_list}"

git commit --amend --no-edit

FORK_REMOTE=${FORK_REMOTE:-origin}
git diff --dirstat "${current_branch}".."${temp_branch}"


printf "\\n\\n!!! Upstream merge complete!\\n"
echo "!!! You can now inspect the branch."
echo ""
echo "!!! To cherry-pick the changes to your original branch, run:"
echo "  git checkout ${current_branch}"
echo "  git cherry-pick -m 2 "'$('"git merge-base ${current_branch} ${temp_branch})..${temp_branch}"
echo ""
echo "!!! To merge the changes to your original branch, run:"
echo "  git checkout ${current_branch}"
echo "  git merge --squash -s recursive -X theirs -m 'Sync upstream ${remote_name} ${remote_ref}'" "${temp_branch}"
echo ""
echo "!!! Once the changes look good, you can push the changes to the remote repository with:"
echo "  git push ${FORK_REMOTE} ${temp_branch}"
