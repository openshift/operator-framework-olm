#!/usr/bin/env bash

# This script is used to keep certain dependencies regularly updated
# against the same OCP branch of the current build.

BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo "updating olm plugin dependencies"
if [[ "$BRANCH" =~ ^master$|^release-\d+\.\d+$ ]]; then
  echo "attempting to update cluster-policy-controller"
  # needed for staging/operator-lifecycle-manager/pkg/controller/operators/olm/plugins/downstream_csv_namespace_labeler_plugin.go
  go get "github.com/openshift/cluster-policy-controller@${BRANCH}"
else
  echo "skipping dependency update as branch '$BRANCH' is not recognized"
fi
echo "finished updating olm plugin dependencies"
