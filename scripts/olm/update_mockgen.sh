#!/usr/bin/env bash

# install dependencies
go install -mod=vendor ./vendor/github.com/golang/mock/mockgen
go install -mod=vendor ./vendor/github.com/maxbrunsfeld/counterfeiter/v6

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
STAGING_DIR="${SCRIPT_ROOT}/staging"
OLM_STAGING_DIR="${STAGING_DIR}/operator-lifecycle-manager"

# generate fakes and mocks
cd ${OLM_STAGING_DIR}  && go generate -mod=vendor ./pkg/...
