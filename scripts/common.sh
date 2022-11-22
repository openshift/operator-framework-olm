#! /bin/bash

export KNOWN_GENERATED_PATHS=(':!vendor' ':!pkg/manifests' ':!manifests' ':!go.sum' ':!go.mod' ':!.gitignore' ':!operator-registry.blacklist' ':!api.blacklist' ':!operator-lifecycle-manager.blacklist' ':!scripts/common.sh')
# TODO(tflannag): This is hacky but works in the current setup.
export ROOT_GENERATED_PATHS=( "vendor" "pkg/manifests" "manifests" "go.mod" "go.sum" )
export UPSTREAM_REMOTES=("api" "operator-registry" "operator-lifecycle-manager")
