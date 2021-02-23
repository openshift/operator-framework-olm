#!/bin/bash

export GOFLAGS="-mod=vendor"
repo_root=$(git rev-parse --show-toplevel)
registry_repo=$repo_root/staging/operator-registry

cd $registry_repo
go generate ./...
cd $repo_root

