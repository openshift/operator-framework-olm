#!/bin/bash

if [ $# -lt 2 ]; then
	echo "Push a commit range or list to a staged upstream repository"
	echo "If provided a ref, it will cherrypick only the commit pointed to"
	echo "usage: $0 <subtree to push> <ref> [<ref>...]"
	exit 0
fi

source "$(dirname $0)/utils.sh"

remote_name=$1
remote_dir="${staging_dir}/${remote_name}"

tracked_ref=$(grep "^${remote_name} " "${repo_list}" | awk '{print $3;}')
remote_ref=${tracked_ref:-master}

if [ "$remote_dir" = "${staging_dir}/" ] || [ ! -d "${remote_dir}" ]; then
	echo "Missing remote from ${repo_list}"
	exit 1
fi

rel_remote_dir="$(realpath --relative-to ${repo_root} ${remote_dir})"

git fetch -t "${remote_name}" "${remote_ref}"

localrev=$(git subtree split --prefix="${rel_remote_dir}") || exit_on_error "failed to create subtree branch"

refs=" ${@:2:$#} "
mapped_refs=""
cachedir="${repo_root}/.git/subtree-cache/$(ls -t ${repo_root}/.git/subtree-cache/ | head -n 1)"

st=0
ln=0
sst=0
sln=0
for i in $(seq 0 $(( ${#refs} - 1 )) ); do
	if [[ ${refs:$i:1} =~ [^\ ~:^\.\\] ]]; then
		if [ $sln -gt 0 ]; then
			mapped_refs="$mapped_refs""${refs:$sst:$sln}"
			sln=0
			st=$i
			ln=1
		else
			ln=$(( ln + 1 ))
		fi
	else
		if [ $ln -gt 0 ]; then
			ds_ref="${refs:$st:$ln}"
			ds_commit=$(git rev-parse "${ds_ref}")
			commit_count=$(ls "${cachedir}/${ds_commit}"* | wc -l)
			if [ "${commit_count}" -eq 0 ]; then
				exit_on_error "no commit ${ds_commit} found for subtree"
			elif [ "${commit_count}" -gt 1 ]; then
				exit_on_error "ambiguous ref ${ds_commit}: $(ls ${cachedir}/${ds_commit}*)"
			fi
			us_commit=$(cat "${cachedir}/${ds_commit}"*)
			mapped_refs="${mapped_refs}""${us_commit}"
			ln=0
			sst=$i
			sln=1
		else
			sln=$(( sln + 1 ))
		fi
	fi
done

newbranch="${remote_name}-downstream-cherry-pick-$(date "+%s")"

staged_mods=$(find "${staging_dir}" -mindepth 1 -maxdepth 1 ! -path "${remote_dir}" -exec sh -c "cd {} &&  go list -m -mod=mod" \;)

git checkout -b "${newbranch}" "${remote_name}/${remote_ref}"
git branch -D "${temp_branch}"
temp_branch="${newbranch}"
git cherry-pick ${mapped_refs} --strategy recursive -X theirs

# revert go build files
git checkout "${remote_name}/${remote_ref}" -- OWNERS vendor
go mod edit -dropreplace "${downstream_repo}"
git show "${remote_name}/${remote_ref}":go.mod > ".go.mod.bk"
for mod in ${staged_mods}; do
	go mod edit -dropreplace "${mod}"
	mod_version=$(grep "${mod}" ".go.mod.bk" | awk '{print $2;}' || true)
	if [ -n "${mod_version}" ]; then
		go mod edit -require "${mod}@${mod_version}"
	else
		go mod edit -droprequire "${mod}"
	fi
done
go mod tidy && go mod vendor || true # leave vendor errors to be corrected later
rm ".go.mod.bk"
git add go.mod go.sum
git commit --amend --no-edit

git diff --dirstat "${current_branch}".."${temp_branch}"
printf "\\n\\n!!! Upstream cherry-pick complete!\\n"
echo "!!! You can now inspect the branch."
echo ""
echo "!!! To switch back to your original branch, run:"
echo "  git checkout ${current_branch}"
echo ""
echo "!!! Once the changes look good, you can push the changes to the remote repository with:"
echo "  git push ${remote_name} ${temp_branch}:<target branch>"
#git push ${remote_name} ${temp_branch}:"refs/heads/${temp_branch}"

#cleanup_and_reset_branch

exit 0
