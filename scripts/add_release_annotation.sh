#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# TODO(tflannag): Add ability to override yq binary as either a parameter
# or an environment varaibles like upstream OLM does.
# For now, you can get the yq binary by running `go install ./vendor/github.com/mikefarah/yq/v3`
# from the root repository and moving that binary into $PATH.
chartdir=$1

for f in $chartdir/*.yaml; do
   yq w -d'*' --inplace --style=double $f 'metadata.annotations['include.release.openshift.io/self-managed-high-availability']' true
   yq w -d'*' --inplace --style=double $f 'metadata.annotations['include.release.openshift.io/single-node-developer']' true
done
