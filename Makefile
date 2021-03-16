SHELL := /bin/bash
ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

GO_BUILD_OPTS := -mod=vendor
GO_BUILD_TAGS := -tags "json1"

GIT_COMMIT := $(or $(SOURCE_GIT_COMMIT),$(shell git rev-parse --short HEAD))
OPM_VERSION := $(or $(SOURCE_GIT_TAG),$(shell git describe --always --tags HEAD))
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

GO_PKG := github.com/operator-framework
REGISTRY_PKG := $(GO_PKG)/operator-registry
OLM_PKG := $(GO_PKG)/operator-lifecycle-manager
API_PKG := $(GO_PKG)/api

arch_flags=GOOS=linux GOARCH=386

OLM_CMDS  := $(shell go list -mod=vendor $(OLM_PKG)/cmd/...)
REGISTRY_CMDS  := $(shell go list -mod=vendor $(REGISTRY_PKG)/cmd/...)

KUBEBUILDER_ASSETS := $(or $(or $(KUBEBUILDER_ASSETS),$(dir $(shell command -v kubebuilder))),/usr/local/kubebuilder/bin)
export KUBEBUILDER_ASSETS
# Ensure kubebuilder is installed before continuing
KUBEBUILDER_ASSETS_ERR := not detected in $(KUBEBUILDER_ASSETS), to override the assets path set the KUBEBUILDER_ASSETS environment variable, for install instructions see https://book.kubebuilder.io/quick-start.html
kubebuilder:
ifeq (, $(wildcard $(KUBEBUILDER_ASSETS)/kubebuilder))
	$(error kubebuilder $(KUBEBUILDER_ASSETS_ERR))
endif
ifeq (, $(wildcard $(KUBEBUILDER_ASSETS)/etcd))
	$(error etcd $(KUBEBUILDER_ASSETS_ERR))
endif
ifeq (, $(wildcard $(KUBEBUILDER_ASSETS)/kube-apiserver))
	$(error kube-apiserver $(KUBEBUILDER_ASSETS_ERR))
endif

# Phony prerequisite for targets that rely on the go build cache to determine staleness.
.PHONY: FORCE
FORCE:

build-util: bin/cpb

bin/cpb: FORCE
	CGO_ENABLED=0 $(arch_flags) go build $(GO_BUILD_OPTS) -ldflags '-extldflags "-static"' -o $@ ./util/cpb

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

build: $(OLM_CMDS) $(REGISTRY_CMDS)

$(REGISTRY_CMDS): version_flags=-ldflags "-X '$(REGISTRY_PKG)/cmd/opm/version.gitCommit=$(GIT_COMMIT)' -X '$(REGISTRY_PKG)/cmd/opm/version.opmVersion=$(OPM_VERSION)' -X '$(REGISTRY_PKG)/cmd/opm/version.buildDate=$(BUILD_DATE)'"
$(REGISTRY_CMDS):
	go build $(version_flags) $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o bin/$(shell basename $@) $@

$(OLM_CMDS): version_flags=-ldflags "-X $(OLM_PKG)/pkg/version.GitCommit=$(GIT_COMMIT) -X $(OLM_PKG)/pkg/version.OLMVersion=`cat staging/operator-lifecycle-manager/OLM_VERSION`"
$(OLM_CMDS):
	go build $(version_flags) $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o bin/$(shell basename $@) $@

build/olm: version_flags=-ldflags "-X $(OLM_PKG)/pkg/version.GitCommit=$(GIT_COMMIT) -X $(OLM_PKG)/pkg/version.OLMVersion=`cat staging/operator-lifecycle-manager/OLM_VERSION`"
build/olm:
	go build $(version_flags) $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o bin/olm $(OLM_PKG)/cmd/olm

bin/kubebuilder:
	$(ROOT_DIR)/scripts/install_kubebuilder.sh

unit/operator-lifecycle-manager: bin/kubebuilder
	$(MAKE) unit WHAT=operator-lifecycle-manager

unit/operator-registry:
	$(MAKE) unit WHAT=operator-registry

unit/api:
	$(MAKE) unit WHAT=api TARGET_NAME=test

unit:
	$(ROOT_DIR)/scripts/unit.sh

e2e/operator-registry:
	go run -mod=vendor github.com/onsi/ginkgo/ginkgo --v --randomizeAllSpecs --randomizeSuites --race $(TAGS) $(REGISTRY_PKG)/test/e2e

olm:
	podman build -f operator-lifecycle-manager.Dockerfile -t test:test --build-arg STAGING_DIR=./staging/operator-lifecycle-manager .

# TODO(tflannag): Do we care about non-opm binary packages in operator-registry?
# DO we still build the things like the initializer, app-registry, etc. binaries?
registry:
	podman build -f operator-registry.Dockerfile -t test:test --build-arg STAGING_DIR=./staging/operator-registry .
