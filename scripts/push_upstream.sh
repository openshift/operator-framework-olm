#!/bin/bash -x

set -eu

if [ $# -lt 2 ]; then
	echo "Push a commit range or list to a staged upstream repository"
	echo "usage: $0 <subtree to push> <ref> [<ref>...]"
	exit 0
fi

source "$(dirname $0)/utils.sh"

remote=$1
repodir=staging/$1

git fetch -t $remote

localrev=$(git subtree split --prefix="$repodir") || exit_on_error "failed to create subtree branch"

refs=" ${@:2:$#} "
mapped_refs=""
cachedir=".git/subtree-cache/$(ls -t .git/subtree-cache/ | head -n 1)"

st=0
ln=0
sst=0
sln=0
for i in $(seq 0 $(( ${#refs} - 1 )) ); do
	if [[ ${refs:$i:1} =~ [a-f0-9] ]]; then
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
			ds_commit="${refs:$st:$ln}"
			commit_count=$(ls "$cachedir/$ds_commit"* | wc -l)
			if [ $commit_count -eq 0 ]; then
				exit_on_error "no commit $ds_commit found for subtree"
			elif [ $commit_count -gt 1 ]; then
				exit_on_error "ambiguous ref $ds_commit: "$(ls $cachedir | grep "^$ds_commit")
			fi
			ds_hash=$(ls "$cachedir" | grep "^$ds_commit")
			us_commit=$(cat "$cachedir/$ds_hash")
			mapped_refs="$mapped_refs""$us_commit"
			ln=0
			sst=$i
			sln=1
		else
			sln=$(( sln + 1 ))
		fi
	fi
done

newbranch="$remote-downstream-cherry-pick-$(date "+%s")"
git checkout -b $newbranch $remote/master
git branch -D $temp_branch
temp_branch=$newbranch

git cherry-pick $mapped_refs

git checkout $remote/master -- OWNERS

sh -c "go mod edit -dropreplace $downstream_repo"
git add go.mod
git commit --amend --no-edit

#git push $remote $newbranch:"refs/heads/$newbranch"

echo "** Ready to push changes!"
echo ""
git diff --dirstat $remote/master..$temp_branch
echo "** Once the updates have been verified, you can push using"
echo "$ git push $remote $newbranch"
echo ""
echo "The original downstream repo is present at the branch ${current_branch}"

#cleanup_and_reset_branch

exit 0
