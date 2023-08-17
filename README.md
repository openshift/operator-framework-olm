# operator-framework-olm

This repository is a monorepo superset of the projects that comprises the
Operator Lifecycle Manager runtime and tooling for use with Openshift.

The upstream projects are:
* [api](https://github.com/operator-framework/api)
* [operator-registry](https://github.com/operator-framework/operator-registry)
* [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)

These projects are located in the `staging` directory. Changes to the
contents of the `staging` directory need to be made upstream first (links
above), and then downstreamed to this repository.

Please refer to the [scripts README.md](scripts/README.md) to learn how to
downstream commits from those projects to this repo.
