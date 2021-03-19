## Syncing downstream OLM

All of the staged repositories live in the top level `staging` directory. The versions of each staged dependency are tracked in the [`scripts/tracked`](./tracked) file.

## Setup

The sync process requires the git-subtree command. See [git subtree](https://github.com/git/git/blob/master/contrib/subtree/INSTALL) for more detailed instructions.

The local repo also needs to track the upstream remotes in [`scripts/tracked`](./tracked). To add these to your repo, run the `init_remotes` script from the root of your repo:
```
./scripts/init_remotes.sh
```

## Syncing with upstream

To sync a staged dependency with an upstream version, you can use the `pull_upstream.sh` helper script. This adds a staged repo if it is not present and updates it to the provided tag/branch otherwise. The script is run as follows:

```
./scripts/pull_upstream.sh <remote url or name> [<ref>]
```
The ref can be a valid tag or branch on the remote, and defaults to master. A successful run adds a new commit to the current branch, similar to 
```
1bec1e8bb Sync upstream api v0.6.1
```
Commit history for the staged repositories is not preserved. The latest synced upstream commit for each staged repo can be found in the `./scripts/tracked` file.

Once the sync is completed, verify it by running the unit tests for the dependencies.

## Pushing changes upstream

Changes made to the staged repositories may be pushed upstream by providing specific commit ranges. For this run:

```
./scripts/push_upstream.sh <remote name> <commit range or list>
```
This creates a local branch containing from the last synced version of the staged dependency with the specified commits cherry-picked onto it. You can then create a PR from this branch to the required upstream repository.
