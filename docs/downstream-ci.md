# Downstream CI

The CI configuration for each release branch can be found [here](https://github.com/openshift/release/tree/master/ci-operator/config/openshift/operator-framework-olm).
From `4.11` (`master` as of this writing) we've updated the configuration to able to influence CI on a PR basis. An overview of the `ci-operator` (the system used for ci)
can be found [here](https://docs.ci.openshift.org/docs/architecture/ci-operator/).

### Structure

 * `.ci-operator.yaml` defines the `build_root_image`. To be ART compliant, the image should come from the [ocp-build-data](https://github.com/openshift/ocp-build-data/) repo
 * [openshift-operator-framework-olm-master.yaml](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/operator-framework-olm/openshift-operator-framework-olm-master.yaml) defines the images that are used by ci, produced by ci, and the ci jobs the get executed.
 * `base.Dockerfile` defines the image used by ci to execute the ci jobs

From [openshift-operator-framework-olm-master.yaml](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/operator-framework-olm/openshift-operator-framework-olm-master.yaml), we see under the `images` stanza the `ci-image` definition.
It goes from `src` (the `build_root_image`) to `ci-image` by building `base.Dockerfile` with `src` as the base image.

```
- dockerfile_path: base.Dockerfile
  from: src
  to: ci-image
```

The image is excluded from promotion, to never be posted up anywhere:

```
promotion:
  excluded_images:
  - ci-image
```

and each `test` references `ci-image` as the image to be used to the test, e.g.:

```
tests:
- as: verify
  commands: make verify
  container:
    from: ci-image
```

### Updating go versions

All we need to do is update the `build_root_image` referenced in `.ci-operator.yaml` and we may also need to update the `base_images` in [openshift-operator-framework-olm-master.yaml](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/operator-framework-olm/openshift-operator-framework-olm-master.yaml). 

**NOTE**: I believe there is some automation that updates the base images, though I don't know. I'll leave this as a questions to the reviewer, and if no one knows, I'll go after it.

### Downstream sync

The complete information about the downstreaming process can be found [here](https://docs.google.com/document/d/139yXeOqAJbV1ndC7Q4NbaOtzbSdNpcuJan0iemORd3g/edit).

TL;DR;

We sync three upstream repositories ([api](https://github.com/operator-framework/api), [registry](https://github.com/operator-framework/operator-registry), [olm](https://github.com/operator-framework/operator-lifecycle-manager)) to the downstream [olm mono-repo](https://github.com/openshift/operator-framework-olm). Commits from the upstream repositories are cherry-picked to the appropriate `staging` directory in the downstream repository. Because this is a monorepo in the `Openshift` GitHub organization, two things need to be remembered:
 - we don't pull in upstream `vendor` folder changes
 - we don't pull in changes to `OWNERS` files
 - after each cherry-pick we execute: `make vendor` and `make manifests` to ensure a) the downstream dependencies are updated b) to ensure any manifest changes are picked up downstream
 -- Note: `make manifests` requires [GNU sed](https://www.gnu.org/software/sed/)

 While manual changes to the `staging` directory should be avoided, there could be instances where there drift between the downstream `staging` directory and the corresponding upstream repository. This can happen due to applying commits out-of-order, e.g. due to feature freeze, etc.

 Therefore, after a sync, it is important to manually verify the diff of `staging` and the upstream. Please note, though, that some downstream changes are downstream only. These are, however, few and far between and there are comments to indicate that a block of code is downstream only.

 The downstream sync process is facilitated by two scripts: `scripts/sync_get_candidates.sh` and `scripts/sync_pop_candidate.sh`, which compare the upstream remote with the appropriate `staging` directory and gets a stack of commits to sync, and cherry-pick those commits in reverse order. What does this look like in practice:

 ```bash
# Clone downstream
git clone git@github.com:openshift/operator-framework-olm.git && cd operator-framework-olm

# Add and fetch upstream remotes
git remote add api git@github.com:operator-framework/api.git && git fetch api
git remote add operator-registry git@github.com:operator-framework/operator-registry.git && git fetch operator-registry
git remote add operator-lifecycle-manager git@github.com:operator-framework/operator-lifecycle-manager.git && git fetch operator-lifecycle-manager

# Get upstream commit candidates: ./scripts/sync_get_candidates.sh <api|operator-registry|operator-lifecycle-manager> <branch>
# The shas will be found in ./<api|operator-registry|operator-lifecycle-manager>.cherrypick
./scripts/sync_get_candidates.sh api master 
./scripts/sync_get_candidates.sh operator-registry master 
./scripts/sync_get_candidates.sh operator-lifecycle-manager master

# Sync upstream commits: ./scripts/sync_pop_candidate.sh <api|operator-registry|operator-lifecycle-manager> [-a]
# Without -a, you'll proceed one commit at a time. With -a the process will conclude once there are no more commits.
# When a cherry pick encounters a conflict the script will stop so you can manually fix it.
sync_pop_candidate.sh operator-lifecycle-manager -a

# When finished
sync_pop_candidate.sh api -a

# When finished
sync_pop_candidate.sh operator-registry -a

# Depending on the changes being pulled in, the order of repos you sync _could_ matter and _could_ leave a commit in an unbuildable state
 ```

 Example: 

 ```bash
$ sync_pop_candidate.sh operator-lifecycle-manager -a

.github/workflows: Enable workflow_dispatch event triggers (#2464)
 Author: Tim Flannagan <timflannagan@gmail.com>
 Date: Mon Dec 20 15:13:33 2021 -0500
 9 files changed, 9 insertions(+), 1 deletion(-)
66 picks remaining (pop_all=true)
popping: 4daeb114ccd56cee7132883325da68c80ba70bed
Auto-merging staging/operator-lifecycle-manager/go.mod
CONFLICT (content): Merge conflict in staging/operator-lifecycle-manager/go.mod
Auto-merging staging/operator-lifecycle-manager/go.sum
CONFLICT (content): Merge conflict in staging/operator-lifecycle-manager/go.sum
CONFLICT (modify/delete): staging/operator-lifecycle-manager/vendor/github.com/operator-framework/api/pkg/validation/doc.go deleted in HEAD and modified in 4daeb114c (chore(api): Vendor the new version of api repo (#2525)).  Version 4daeb114c (chore(api): Vendor the new version of api repo (#2525)) of staging/operator-lifecycle-manager/vendor/github.com/operator-framework/api/pkg/validation/doc.go left in tree.
CONFLICT (modify/delete): staging/operator-lifecycle-manager/vendor/modules.txt deleted in HEAD and modified in 4daeb114c (chore(api): Vendor the new version of api repo (#2525)).  Version 4daeb114c (chore(api): Vendor the new version of api repo (#2525)) of staging/operator-lifecycle-manager/vendor/modules.txt left in tree.
error: could not apply 4daeb114c... chore(api): Vendor the new version of api repo (#2525)
hint: After resolving the conflicts, mark them with
hint: "git add/rm <pathspec>", then run
hint: "git cherry-pick --continue".
hint: You can instead skip this commit with "git cherry-pick --skip".
hint: To abort and get back to the state before "git cherry-pick",
hint: run "git cherry-pick --abort".

$ rm -rf staging/operator-lifecycle-manager/vendor

# make sure there are no conflics in 
# staging/operator-lifecycle-manager/go.mod and go.sum
$ cd staging/operator-lifecycle-manager
$ go mod tidy
$ cd ../../

# now that the conflict is fixed, advance again
$ sync_pop_candidate.sh operator-lifecycle-manager -a
 ```

### Troubleshooting

#### Running console test locally

The [console](https://github.com/openshift/console) repository contains all instructions you need to execute the console tests locally. The olm console tests can be found [here](https://github.com/openshift/console/tree/master/frontend/packages/operator-lifecycle-manager)