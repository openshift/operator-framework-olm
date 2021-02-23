# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
  BUILD_VERBOSE = $(V)
endif

SHELL := /bin/bash
PKG = github.com/openshift/operator-framework-olm
PKGS = $(shell $(GO) list ./... | grep -v /vendor/)
export GO111MODULE=on
GO := GOFLAGS="-mod=mod" go
REGISTRY_CMDS := $(addprefix bin/, $(shell ls ./cmd | grep -v "^\(olm\|opm\|operator-verify\)"))
CMDS  := $(shell $(GO) list ./cmd/olm/... ./cmd/opm/...)
TCMDS := $(shell $(GO) list ./test/e2e/...)

SPECIFIC_UNIT_TEST := $(if $(TEST),-run $(TEST),)

# ART builds are performed in dist-git, with content (but not commits) copied 
# from the source repo. Thus at build time if your code is inspecting the local
# git repo it is getting unrelated commits and tags from the dist-git repo, 
# not the source repo.
# For ART image builds, SOURCE_GIT_COMMIT, SOURCE_GIT_TAG, SOURCE_DATE_EPOCH 
# variables are inserted in Dockerfile to enable recovering the original git 
# metadata at build time.
GIT_COMMIT := $(or $(SOURCE_GIT_COMMIT),$(shell git rev-parse --short HEAD))

OLM_VERSION := $(or $(SOURCE_GIT_TAG),$(shell git describe --always --tags HEAD))
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
TAGS := -tags "json1"
registry_api := "./staging/operator-registry/pkg/api"
# -race is only supported on linux/amd64, linux/ppc64le, linux/arm64, freebsd/amd64, netbsd/amd64, darwin/amd64 and windows/amd64
ifeq ($(shell $(GO) env GOARCH),s390x)
TEST_RACE :=
else
TEST_RACE := -race
endif
KUBEBUILDER_ASSETS := $(or $(or $(KUBEBUILDER_ASSETS),$(dir $(shell command -v kubebuilder))),/usr/local/kubebuilder/bin)
export KUBEBUILDER_ASSETS

# Undefine GOFLAGS environment variable.
ifdef GOFLAGS
$(warning Undefining GOFLAGS set in CI)
undefine GOFLAGS
endif

.PHONY: all
all: clean test build

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

bin/operator-verify:
	$(GO) build \
		-gcflags "all=-trimpath=${GOPATH}" \
		-asmflags "all=-trimpath=${GOPATH}" \
		-ldflags " \
			-X '${PKG}/version.GitVersion=${OLM_VERSION}' \
			-X '${PKG}/version.GitCommit=${GIT_COMMIT}' \
		" \
		-o "bin/operator-verify" \
		"$(PKG)/cmd/operator-verify"

$(REGISTRY_CMDS):
	$(arch_flags) $(GO) build $(extra_flags) $(TAGS) -o bin/$(shell basename $@) ./cmd/$(notdir $@)

$(CMDS): version_flags=-ldflags "-X '$(PKG)/pkg/version.GitCommit=$(GIT_COMMIT)' -X '$(PKG)/pkg/version.OLMVersion=$(OLM_VERSION)' -X '$(PKG)/pkg/version.buildDate=$(BUILD_DATE)'"
$(CMDS):
	$(arch_flags) $(GO) build $(version_flags) $(extra_flags) $(TAGS) -o bin/$(shell basename $@) $@

$(TCMDS):
	$(GO) test -c $(BUILD_TAGS) -o bin/$(shell basename $@) $@


.PHONY: build ## Build binaries
build: clean $(REGISTRY_CMDS) $(CMDS) bin/operator-verify

.PHONY: cross
cross: version_flags=-ldflags "-X '$(PKG)/pkg/version.GitCommit=$(GIT_COMMIT)' -X '$(PKG)/pkg/version.OLMVersion=$(OLM_VERSION)' -X '$(PKG)/pkg/version.buildDate=$(BUILD_DATE)'"
cross: ## Cross-compile opm binary
ifeq ($(shell $(GO) env GOARCH),amd64)
	GOOS=darwin CC=o64-clang CXX=o64-clang++ CGO_ENABLED=1 $(GO) build $(version_flags) $(TAGS) -o "bin/darwin-amd64-opm" --ldflags "-extld=o64-clang" ./cmd/opm
	GOOS=windows CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_ENABLED=1 $(GO) build $(version_flags) $(TAGS)  -o "bin/windows-amd64-opm" --ldflags "-extld=x86_64-w64-mingw32-gcc" ./cmd/opm
endif

.PHONY: static
static: extra_flags=-ldflags '-w -extldflags "-static"'
static: build

.PHONY: registry-image
registry-image:
	docker build -f operator-registry.Dockerfile .

# Code management.
.PHONY: format tidy clean vendor generate

format: ## Format the source code
	$(GO) fmt $(PKGS)

tidy: ## Update dependencies
	$(GO) mod tidy -v

vendor: tidy ## Update vendor directory
	$(GO) mod vendor

clean: ## Clean up the build artifacts
	@rm -rf build
	@rm -rf ./bin
	@rm -rf cover.out
	@rm -rf bin
	@rm -rf test/e2e/resources
	@rm -rf test/e2e/test-resources
	@rm -rf test/e2e/log
	@rm -rf e2e.namespace

manifests: ## Generate and copy CRD manifests
	./scripts/olm/generate_manifests.sh

.PHONY: generate-fakes
generate-fakes: ## Generate deepcopy, conversion, clients, listers, and informers
	./scripts/olm/generate_fakes.sh


codegen: ## Generate clients, listers, and informers
	./scripts/olm/update_codegen.sh

mockgen: ## Generate mock types.
	./scripts/olm/update_mockgen.sh

gen-all: manifests generate-fakes gen-grpc codegen mockgen ## Generate everything.

.PHONY: gen-grpc
gen-grpc: ## Generate GRPC APIs for registry
	protoc -I $(registry_api) --go_out=$(registry_api) $(registry_api)/*.proto
	protoc -I $(registry_api) --go-grpc_out=$(registry_api) $(registry_api)/*.proto
	protoc -I $(registry_api)/grpc_health_v1 --go_out=$(registry_api)/grpc_health_v1 $(registry_api)/grpc_health_v1/*.proto
	protoc -I $(registry_api)/grpc_health_v1 --go-grpc_out=$(registry_api)/grpc_health_v1 $(registry_api)/grpc_health_v1/*.proto

.PHONY: container-gen-grpc
container-gen-grpc:
	docker build -t operator-registry:codegen -f codegen.Dockerfile .
	docker run --name temp-codegen operator-registry:codegen /bin/true
	docker cp temp-codegen:/codegen/pkg/api/. $(registry_api)
	docker rm temp-codegen

diff:
	git diff --exit-code

verify-codegen: codegen diff
verify-mockgen: mockgen diff
verify-manifests: manifests diff
verify: verify-codegen verify-mockgen verify-manifests

# Static tests.
.PHONY: test unit e2e

test: unit cover.out ## Run unit tests

unit: kubebuilder
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) $(GO) test $(SPECIFIC_UNIT_TEST) $(TAGS) $(TEST_RACE) -count=1 -v ./...

cover.out: 
	$(GO) test $(TAGS) -v $(TEST_RACE) -coverprofile=cover.out -covermode=atomic -coverpkg ./...

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

.PHONY: e2e

e2e:
	$(GO) test -v -failfast -timeout 150m ./test/e2e/... -namespace=openshift-operators -kubeconfig=${KUBECONFIG} -olmNamespace=openshift-operator-lifecycle-manager -dummyImage=bitnami/nginx:latest -ginkgo.flakeAttempts=3

e2e-operator-metrics:
	$(GO) test -v -timeout 70m ./test/rh-operators/...


##########################
#  OLM - Build and Test  #
##########################

IMAGE_REPO := quay.io/operator-framework/olm
IMAGE_TAG ?= "dev"
YQ_INTERNAL := $(GO) run ./vendor/github.com/mikefarah/yq/v3/

.PHONY: build test run clean vendor \
	vendor-update coverage coverage-html e2e \
	kubebuilder .FORCE

build-linux: arch_flags=GOOS=linux GOARCH=386
build-linux: clean $(CMDS)

build-wait: clean bin/wait

bin/wait:
	GOOS=linux GOARCH=386 $(GO) build -o $@ $(PKG)/test/e2e/wait

build-util-linux: arch_flags=GOOS=linux GOARCH=386
build-util-linux: build-util

build-util: bin/cpb

bin/cpb:
	CGO_ENABLED=0 $(arch_flags) $(GO) build -ldflags '-extldflags "-static"' -o $@ ./util/cpb


test-bare: BUILD_TAGS=-tags=bare
test-bare: clean $(TCMDS)

test-bin: clean $(TCMDS)

################################
#  OLM - Install/Uninstall/Run #
################################

LOCAL_NAMESPACE := "olm"

.PHONY: run-console-local
run-console-local: ## Run Openshift console locally
	@echo Running script to run the OLM console locally:
	. ./scripts/olm/run_console_local.sh

.PHONY: run-local ## Build and run OLM locally
run-local: build-linux build-wait build-util-linux build-local deploy-local

build-local: 
	rm -rf build
	. ./scripts/olm/build_local.sh
	
deploy-local:
	mkdir -p build/resources
	. ./scripts/olm/package_release.sh 1.0.0 build/resources doc/install/local-values.yaml
	. ./scripts/olm/install_local.sh $(LOCAL_NAMESPACE) build/resources
	rm -rf build

# e2e test exculding the rh-operators directory which tests rh-operators and their metric cardinality.
clean-e2e:
	kubectl delete crds --all
	kubectl delete apiservices.apiregistration.k8s.io v1.packages.operators.coreos.com || true
	kubectl delete -f test/e2e/resources/0000_50_olm_00-namespace.yaml

e2e-local: build-linux build-wait build-util-linux build-local ## Run e2e tests locally
	. ./scripts/olm/run_e2e_local.sh $(TEST)

e2e-local-docker:
	. ./scripts/olm/build_local.sh
	. ./scripts/olm/run_e2e_docker.sh $(TEST)

# useful if running e2e directly with `go test -tags=bare`
e2e.namespace:
	@printf "e2e-tests-$(shell date +%s)-$$RANDOM" > e2e.namespace

setup-bare: clean e2e.namespace
	. ./scripts/olm/build_bare.sh
	. ./scripts/olm/package_release.sh 1.0.0 test/e2e/resources test/e2e/e2e-bare-values.yaml
	. ./scripts/olm/install_bare.sh $(shell cat ./e2e.namespace) test/e2e/resources

e2e-bare: setup-bare
	. ./scripts/olm/run_e2e_bare.sh $(TEST)

