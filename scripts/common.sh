#! /bin/bash

export KNOWN_GENERATED_PATHS=(':!vendor' ':!pkg/manifests' ':!manifests' ':!go.sum' ':!go.mod')
export UPSTREAM_REMOTES=("api" "operator-registry" "operator-lifecycle-manager")
