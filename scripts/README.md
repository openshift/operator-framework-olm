# Syncing downstream OLM

All of the staged repositories live in the top level `staging` directory.

The downstreaming process is complex and helper scripts have been written
to facilitate downstreaming.

## Automatic Downstreaming

There is now an automated downstreaming process for OLMv0 from the three
source repositories.

The "bumper" program is located in [openshift/operator-framework-tooling](https://github.com/openshift/operator-framework-tooling).
It is automatically run on a daily basis based on the following [openshift/release](https://github.com/openshift/release/blob/3bf0b3ae011debaefefb564ad6f233c380d033f7/ci-operator/jobs/infra-periodics.yaml#L926-L978) config.

If the bumper program fails to create a mergeable PR, manual intervention will be necessary.
This may require copying, modifying and resubmitting the PR.

# Long-lived Carry Commits

It is required at times to write commits that will live in the `vendor/` directory
on top of upstream code and for those commits to be carried on top for the forseeable
future. In these cases, prefix your commit message with `[CARRY]` to pass the commit
verification routines.

## References
1. [Downstream to operator-framework-olm](https://spaces.redhat.com/display/OOLM/Downstream+to+operator-framework-olm)
2. [OLM downstreaming guide](https://docs.google.com/document/d/139yXeOqAJbV1ndC7Q4NbaOtzbSdNpcuJan0iemORd3g/edit#) (old)
