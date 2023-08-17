# Syncing downstream OLM

All of the staged repositories live in the top level `staging` directory.

The downstreaming process is complex and helper scripts have been written
to facilitate downstreaming.

## Assumptions

The helper scripts assume that the upstream remote repos are configured
as follows:
```
git remote add api https://github.com/operator-framework/api
git remote add operator-registry https://github.com/operator-framework/operator-registry
git remote add operator-lifecycle-manager https://github.com/operator-framework/operator-lifecycle-manager
```
The [sync.sh](sync.sh) script will automatically create these
remote repositories.

## Bulk Sync

To sync all current changes from upstream, simply run the sync script:
```sh
scripts/sync.sh
```

This script may pause at certain points to ask the user to examine
command failures or possible regressions. Please open another terminal
and review the state of the workspace before continuing the script.

When the script completes, it will have created a branch whose name is
the current date (formatted: `sync-YYYY-MM-DD`).

If the `sync.sh` script fails, please refer to
[[1](https://spaces.redhat.com/display/OOLM/Downstream+to+operator-framework-olm)]
for continuation proceedures.

Before this branch can be used to create a PR, run the following:
```sh
make -k verify
```
If there are any diffs or modified files, these need to be added to 
your branch as either a separate commit (e.g. headline: `Run make verify`),
or amended to the last commit of the branch.

Once `make -k verify` is resolved, create a PR from this sync branch.


## Targeted Sync

To sync a subset of commits from the upstream repositories (i.e. critical
bugfix), create a new working sync branch. Then create a `sync.cherrypick`
file in the repositry root directory with the repos and commit SHAs.

The format of the cherrypick file is:
```
<order> <repo> <commit-SHA>
```

For example:
```
1 api 0123456789abcdef0123456789abcdef01234567
2 operator-lifecycle-manager 123456789abcdef0123456789abcdef012345678
3 operator-lifecycle-manager 23456789abcdef0123456789abcdef0123456789
```
Do _not_ commit the cherrypick file, it is a temporary working file that
is ignored by `git`.

Then run the following:
```sh
scripts/sync_pop_candidate.sh -a sync
```
The commits in the `sync.cherrypick` file will be applied in the specified
order.

Even if you only have a single commit, this procedure will follow the same
process that `sync.sh` does, to ensure no steps are missed.

Before this branch can be used to create a PR, run the following:
```sh
make -k verify
```
If there are any diffs or modified files, these need to be added to 
your branch as either a separate commit (e.g. headline: `Run make verify`),
or amended to the last commit of the branch.

Once `make -k verify` is resolved, create a PR from this working sync branch.

# Long-lived Carry Commits

It is required at times to write commits that will live in the `vendor/` directory
on top of upstream code and for those commits to be carried on top for the forseeable
future. In these cases, prefix your commit message with `[CARRY]` to pass the commit
verification routines.

## References
1. [Downstream to operator-framework-olm](https://spaces.redhat.com/display/OOLM/Downstream+to+operator-framework-olm)
2. [OLM downstreaming guide](https://docs.google.com/document/d/139yXeOqAJbV1ndC7Q4NbaOtzbSdNpcuJan0iemORd3g/edit#) (old)
