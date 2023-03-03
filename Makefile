SHELL := /bin/bash
ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
CONTAINER_ENGINE := docker

OPM_VERSION := $(or $(SOURCE_GIT_TAG),$(shell git describe --always --tags HEAD))
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
# ART builds are performed in dist-git, with content (but not commits) copied
# from the source repo. Thus at build time if your code is inspecting the local
# git repo it is getting unrelated commits and tags from the dist-git repo,
# not the source repo.
# For ART image builds, SOURCE_GIT_COMMIT, SOURCE_GIT_TAG, SOURCE_DATE_EPOCH
# variables are inserted in Dockerfile to enable recovering the original git
# metadata at build time.
GIT_COMMIT := $(if $(SOURCE_GIT_COMMIT),$(SOURCE_GIT_COMMIT),$(shell git rev-parse HEAD))
# ART also populates OS_GIT_MAJOR, OS_GIT_MINOR, OS_GIT_PATCH, OS_GIT_VERSION
# and BUILD_VERSION as part of the Docker ENV reflecting the OpenShift version
# for the image.
# OLM_VERSION needs to be semver2 compatible for helm v2.
OLM_VERSION := $(or $(OS_GIT_VERSION),0.0.0-$(GIT_COMMIT))

GO_BUILD_OPTS := -mod=vendor
GO_BUILD_TAGS := -tags "json1"

GO_PKG := github.com/operator-framework
REGISTRY_PKG := $(GO_PKG)/operator-registry
OLM_PKG := $(GO_PKG)/operator-lifecycle-manager
API_PKG := $(GO_PKG)/api
ROOT_PKG := github.com/openshift/operator-framework-olm

COLLECT_PROFILES_CMD := $(addprefix bin/, collect-profiles)
OPM := $(addprefix bin/, opm)
OLM_CMDS  := $(shell go list -mod=vendor $(OLM_PKG)/cmd/...)
PSM_CMD := $(addprefix bin/, psm)
REGISTRY_CMDS  := $(addprefix bin/, $(shell ls staging/operator-registry/cmd | grep -v opm))

# Default image tag for build/olm-container and build/registry-container
IMG ?= test:test

# Phony prerequisite for targets that rely on the go build cache to determine staleness.
.PHONY: FORCE
FORCE:

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

build: $(REGISTRY_CMDS) $(OLM_CMDS) $(OPM) ## build opm and olm binaries

build/opm:
	$(MAKE) $(OPM)

build/registry:
	$(MAKE) $(REGISTRY_CMDS) $(OPM)

build/olm:
	$(MAKE) $(PSM_CMD) $(OLM_CMDS) $(COLLECT_PROFILES_CMD)

$(OPM): version_flags=-ldflags "-X '$(REGISTRY_PKG)/cmd/opm/version.gitCommit=$(GIT_COMMIT)' -X '$(REGISTRY_PKG)/cmd/opm/version.opmVersion=$(OPM_VERSION)' -X '$(REGISTRY_PKG)/cmd/opm/version.buildDate=$(BUILD_DATE)'"
$(OPM):
	go build $(version_flags) $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o $@ $(REGISTRY_PKG)/cmd/$(notdir $@)

$(REGISTRY_CMDS): version_flags=-ldflags "-X '$(REGISTRY_PKG)/cmd/opm/version.gitCommit=$(GIT_COMMIT)' -X '$(REGISTRY_PKG)/cmd/opm/version.opmVersion=$(OPM_VERSION)' -X '$(REGISTRY_PKG)/cmd/opm/version.buildDate=$(BUILD_DATE)'"
$(REGISTRY_CMDS):
	go build $(version_flags) $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o $@ $(REGISTRY_PKG)/cmd/$(notdir $@)

$(OLM_CMDS): version_flags=-ldflags "-X $(OLM_PKG)/pkg/version.GitCommit=$(GIT_COMMIT) -X $(OLM_PKG)/pkg/version.OLMVersion=$(OLM_VERSION)"
$(OLM_CMDS):
	go build $(version_flags) $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o bin/$(shell basename $@) $@

$(PSM_CMD): FORCE
	go build $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o $(PSM_CMD) $(ROOT_PKG)/cmd/package-server-manager

$(COLLECT_PROFILES_CMD): FORCE
	go build $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o $(COLLECT_PROFILES_CMD) $(ROOT_PKG)/cmd/collect-profiles

.PHONY: cross
cross: version_flags=-X '$(REGISTRY_PKG)/cmd/opm/version.gitCommit=$(GIT_COMMIT)' -X '$(REGISTRY_PKG)/cmd/opm/version.opmVersion=$(OPM_VERSION)' -X '$(REGISTRY_PKG)/cmd/opm/version.buildDate=$(BUILD_DATE)'
cross:
ifeq ($(shell go env GOARCH),amd64)
	GOOS=darwin CC=o64-clang CXX=o64-clang++ CGO_ENABLED=1 go build $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o "bin/darwin-amd64-opm" --ldflags "-extld=o64-clang $(version_flags)" $(REGISTRY_PKG)/cmd/opm
	GOOS=windows CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_ENABLED=1 go build $(GO_BUILD_OPTS) $(GO_BUILD_TAGS) -o "bin/windows-amd64-opm" --ldflags "-extld=x86_64-w64-mingw32-gcc $(version_flags)" -buildmode=exe $(REGISTRY_PKG)/cmd/opm
endif

build/olm-container: clean
	$(CONTAINER_ENGINE) build -f operator-lifecycle-manager.Dockerfile -t ${IMG} .

build/registry-container: clean
	$(CONTAINER_ENGINE) build -f operator-registry.Dockerfile -t ${IMG} .

bin/kubebuilder:
	$(ROOT_DIR)/scripts/install_kubebuilder.sh

bin/cpb: FORCE
	CGO_ENABLED=0 go build $(GO_BUILD_OPTS) -ldflags '-extldflags "-static"' -o $@ github.com/operator-framework/operator-lifecycle-manager/util/cpb

unit/olm: bin/kubebuilder
	echo "Running the OLM unit tests"
	$(MAKE) unit WHAT=operator-lifecycle-manager

unit/registry:
	$(MAKE) unit WHAT=operator-registry

unit/api:
	$(MAKE) unit WHAT=api TARGET_NAME=test

unit/psm:
	go test $(ROOT_DIR)/pkg/package-server-manager/...

unit: ## Run unit tests
	$(ROOT_DIR)/scripts/unit.sh

e2e:
	scripts/e2e.sh

e2e/operator-registry: ## Run e2e registry tests
	$(MAKE) e2e WHAT=operator-registry

e2e/olm: ## Run e2e olm tests
	$(MAKE) e2e WHAT=operator-lifecycle-manager E2E_CATALOG_NS=openshift-marketplace E2E_INSTALL_NS=openshift-operator-lifecycle-manager E2E_TEST_NS=openshift-operators E2E_TIMEOUT=120m KUBECTL=oc

.PHONY: update-plugin-deps
update-plugin-deps:
	./scripts/update_plugin_deps.sh

.PHONY: vendor
vendor: update-plugin-deps
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: manifests
manifests: ## Generate manifests
	OLM_VERSION=$(OLM_VERSION) ./scripts/generate_crds_manifests.sh

.PHONY: diff
diff:
	git diff --stat HEAD --ignore-submodules --exit-code

verify-vendor: vendor
	$(MAKE) diff

verify-manifests: OLM_VERSION=0.19.0 # set static version to avoid failing for commit based versioning
verify-manifests: manifests
	$(MAKE) diff

verify-nested-vendor:
	./scripts/check-staging-vendor.sh

.PHONY: verify-commits
verify-commits:
	./scripts/verify_commits.sh $(PULL_BASE_SHA) # see https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#job-environment-variables

.PHONY: verify
verify:
	echo "Checking for unstaged root vendor changes"
	$(MAKE) verify-vendor
	echo "Checking whether the CVO manifests need to be generated"
	$(MAKE) verify-manifests
	echo "Checking for unsynced nested [go.mod,go.sum] files"
	$(MAKE) verify-nested-vendor
	echo "Checking commit integrity"
	$(MAKE) verify-commits

.PHONY: crc-start
crc-start:
	echo "Starting CRC"
	./scripts/crc-start.sh

.PHONY: crc-build
crc-build:
	echo "Building olm image"
	IMG="olm:test" $(MAKE) build/olm-container
	echo "Building opm image"
	IMG="opm:test" $(MAKE) build/registry-container

.PHONY: crc-deploy
crc-deploy:
	echo "Deploying OLM"
	./scripts/crc-deploy.sh

.PHONY: crc
crc: crc-start crc-build crc-deploy

.PHONY: clean
clean:
	rm -rf bin

.PHONY: sync
sync:
	./scripts/sync.sh

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
