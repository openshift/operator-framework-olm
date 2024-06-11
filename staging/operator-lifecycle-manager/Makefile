##########################
#  OLM - Build and Test  #
##########################

# Undefine GOFLAGS environment variable.
ifdef GOFLAGS
$(warning Undefining GOFLAGS set in CI)
undefine GOFLAGS
endif

SHELL := /bin/bash
ORG := github.com/operator-framework
PKG   := $(ORG)/operator-lifecycle-manager
MOD_FLAGS := -mod=vendor -buildvcs=false
BUILD_TAGS := "json1"
CMDS  := $(shell go list $(MOD_FLAGS) ./cmd/...)
TCMDS := $(shell go list $(MOD_FLAGS) ./test/e2e/...)
MOCKGEN := ./scripts/update_mockgen.sh
CODEGEN := ./scripts/update_codegen.sh
IMAGE_REPO := quay.io/operator-framework/olm
IMAGE_TAG ?= "dev"
SPECIFIC_UNIT_TEST := $(if $(TEST),-run $(TEST),)
LOCAL_NAMESPACE := "olm"
export GO111MODULE=on
YQ_INTERNAL := go run $(MOD_FLAGS) ./vendor/github.com/mikefarah/yq/v3/
KUBEBUILDER_ASSETS := $(or $(or $(KUBEBUILDER_ASSETS),$(dir $(shell command -v kubebuilder))),/usr/local/kubebuilder/bin)
export KUBEBUILDER_ASSETS
GO := GO111MODULE=on GOFLAGS="$(MOD_FLAGS)" go
GINKGO := $(GO) run github.com/onsi/ginkgo/v2/ginkgo
BINDATA := $(GO) run github.com/go-bindata/go-bindata/v3/go-bindata
GIT_COMMIT := $(shell git rev-parse HEAD)
ifeq ($(shell arch), arm64) 
ARCH := arm64
else
ARCH := 386
endif

# Take operator registry tag from operator registry version in go.mod
export OPERATOR_REGISTRY_TAG ?= $(shell go list -m github.com/operator-framework/operator-registry | cut -d" " -f2)

# Pin operator registry images to the OPERATOR_REGISTRY_TAG
export OPERATOR_REGISTRY_IMAGE ?= quay.io/operator-framework/opm:$(OPERATOR_REGISTRY_TAG)
export CONFIGMAP_SERVER_IMAGE ?= quay.io/operator-framework/configmap-operator-registry:$(OPERATOR_REGISTRY_TAG)

# Phony prerequisite for targets that rely on the go build cache to determine staleness.
.PHONY: build test clean vendor \
	coverage coverage-html e2e \
	kubebuilder

.PHONY: FORCE
FORCE:

all: test build

test: clean cover.out

unit: kubebuilder
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) go test $(MOD_FLAGS) $(SPECIFIC_UNIT_TEST) -tags "json1" -race -count=1 ./pkg/... ./test/e2e/split/...

# Ensure kubectl installed before continuing
KUBEBUILDER_ASSETS_ERR := not detected in $(KUBEBUILDER_ASSETS), to override the assets path set the KUBEBUILDER_ASSETS environment variable, for install instructions see https://pkg.go.dev/sigs.k8s.io/controller-runtime/tools/setup-envtest
KUBECTL_ASSETS_ERR := kubectl not detected.
kubebuilder:
ifeq (, $(shell which kubectl))
	$(error $(KUBECTL_ASSETS_ERR))
endif
ifeq (, $(wildcard $(KUBEBUILDER_ASSETS)/etcd))
	$(error etcd $(KUBEBUILDER_ASSETS_ERR))
endif
ifeq (, $(wildcard $(KUBEBUILDER_ASSETS)/kube-apiserver))
	$(error kube-apiserver $(KUBEBUILDER_ASSETS_ERR))
endif

cover.out:
	go test $(MOD_FLAGS) -tags "json1" -race -coverprofile=cover.out -covermode=atomic \
		-coverpkg ./pkg/controller/... ./pkg/...

coverage: cover.out
	go tool cover -func=cover.out

coverage-html: cover.out
	go tool cover -html=cover.out

build: build_cmd=build
build: clean $(CMDS)

test-bare: BUILD_TAGS=-tags=bare
test-bare: clean $(TCMDS)

test-bin: clean $(TCMDS)

# build versions of the binaries with coverage enabled
build-coverage: build_cmd=test -c -covermode=count -coverpkg ./pkg/controller/...
build-coverage: clean $(CMDS)

build-linux: build_cmd=build
build-linux: arch_flags=GOOS=linux GOARCH=$(ARCH)
build-linux: clean $(CMDS)

build-wait: clean bin/wait

bin/wait: FORCE
	GOOS=linux GOARCH=$(ARCH) go build $(MOD_FLAGS) -o $@ $(PKG)/test/e2e/wait

build-util-linux: arch_flags=GOOS=linux GOARCH=$(ARCH)
build-util-linux: build-util

build-util: bin/cpb bin/copy-content

bin/cpb: FORCE
	CGO_ENABLED=0 $(arch_flags) go build -buildvcs=false $(MOD_FLAGS) -ldflags '-extldflags "-static"' -o $@ ./util/cpb

bin/copy-content: FORCE
	CGO_ENABLED=0 $(arch_flags) go build -buildvcs=false $(MOD_FLAGS) -ldflags '-extldflags "-static"' -o $@ ./cmd/copy-content

$(CMDS): version_flags=-ldflags "-X $(PKG)/pkg/version.GitCommit=$(GIT_COMMIT) -X $(PKG)/pkg/version.OLMVersion=`cat OLM_VERSION`"
$(CMDS):
	$(arch_flags) go $(build_cmd) $(MOD_FLAGS) $(version_flags) -tags $(BUILD_TAGS) -o bin/$(shell basename $@) $@

build: clean $(CMDS)

$(TCMDS):
	go test -c $(BUILD_TAGS) $(MOD_FLAGS) -o bin/$(shell basename $@) $@

deploy-local:
	mkdir -p build/resources
	. ./scripts/package_release.sh 1.0.0 build/resources doc/install/local-values.yaml
	. ./scripts/install_local.sh $(LOCAL_NAMESPACE) build/resources
	rm -rf build

e2e.namespace:
	@printf "e2e-tests-$(shell date +%s)-$$RANDOM" > e2e.namespace

.PHONY: e2e
E2E_TIMEOUT ?= 90m
E2E_TEST_NS ?= operators
E2E_INSTALL_NS ?= operator-lifecycle-manager
E2E_CATALOG_NS ?= $(E2E_INSTALL_NS)
E2E_FLAKE_ATTEMPTS ?= 1
GINKGO_OPTS += -v -randomize-suites -race -trace --show-node-events --flake-attempts=$(E2E_FLAKE_ATTEMPTS) $(if $(E2E_SEED),-seed '$(E2E_SEED)') $(if $(TEST),-focus '$(TEST)',) $(if $(SKIP), -skip '$(SKIP)')
e2e:
	$(GINKGO) -timeout $(E2E_TIMEOUT) $(GINKGO_OPTS) ./test/e2e -- -namespace=$(E2E_TEST_NS) -olmNamespace=$(E2E_INSTALL_NS) -catalogNamespace=$(E2E_CATALOG_NS) $(E2E_OPTS)

.PHONY: e2e-local
e2e-local: e2e-build kind-create deploy e2e

# See workflows/e2e-tests.yml See test/e2e/README.md for details.
.PHONY: e2e-local
e2e-local: BUILD_TAGS="json1 e2e experimental_metrics"
e2e-local: extra_args=-kind.images=../test/e2e-local.image.tar -test-data-dir=../test/e2e/testdata -gather-artifacts-script-path=../test/e2e/collect-ci-artifacts.sh
e2e-local: run=bin/e2e-local.test
e2e-local: bin/e2e-local.test test/e2e-local.image.tar
e2e-local: e2e

# this target updates the zz_chart.go file with files found in deploy/chart
# this will always fire since it has been marked as phony
.PHONY: test/e2e/assets/chart/zz_chart.go
test/e2e/assets/chart/zz_chart.go: $(shell find deploy/chart -type f)
	$(BINDATA) -o $@ -pkg chart -prefix deploy/chart/ $^

.PHONY: deploy
OLM_IMAGE := quay.io/operator-framework/olm:local
deploy:
	$(KIND) load docker-image $(OLM_IMAGE) --name $(KIND_CLUSTER_NAME); \
	$(HELM) upgrade --install olm deploy/chart \
		--set debug=true \
		--set olm.image.ref=$(OLM_IMAGE) \
		--set olm.image.pullPolicy=IfNotPresent \
		--set catalog.image.ref=$(OLM_IMAGE) \
		--set catalog.image.pullPolicy=IfNotPresent \
		--set catalog.commandArgs=--configmapServerImage=$(CONFIGMAP_SERVER_IMAGE) \
		--set catalog.opmImageArgs=--opmImage=$(OPERATOR_REGISTRY_IMAGE) \
		--set package.image.ref=$(OLM_IMAGE) \
		--set package.image.pullPolicy=IfNotPresent \
		$(HELM_INSTALL_OPTS) \
		--wait;

.PHONY: e2e-build
e2e-build: BUILD_TAGS="json1 e2e experimental_metrics"
e2e-build: export GOOS=linux
e2e-build: build_cmd=build
e2e-build: e2e.Dockerfile bin/wait bin/cpb $(CMDS)
	docker build -t quay.io/operator-framework/olm:local -f $< bin
	docker save -o $@ quay.io/operator-framework/olm:local

e2e-bare: setup-bare
	. ./scripts/run_e2e_bare.sh $(TEST)

e2e-local-docker:
	. ./scripts/build_local.sh
	. ./scripts/run_e2e_docker.sh $(TEST)

container:
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .

clean-e2e:
	kubectl delete crds --all
	kubectl delete apiservices.apiregistration.k8s.io v1.packages.operators.coreos.com || true
	kubectl delete -f test/e2e/resources/0000_50_olm_00-namespace.yaml

clean:
	@rm -rf cover.out
	@rm -rf bin
	@rm -rf test/e2e/resources
	@rm -rf test/e2e/test-resources
	@rm -rf test/e2e/log
	@rm -rf e2e.namespace

# Copy CRD manifests
manifests: vendor
	./scripts/copy_crds.sh

# Generate deepcopy, conversion, clients, listers, and informers
codegen:
	# Clients, listers, and informers
	$(CODEGEN)

# Generate mock types.
mockgen:
	$(MOCKGEN)

# Generates everything.
gen-all: codegen mockgen manifests

diff:
	git diff --exit-code

verify-codegen: codegen
	$(MAKE) diff

verify-mockgen: mockgen
	$(MAKE) diff

verify-manifests: manifests
	$(MAKE) diff

verify: verify-codegen verify-mockgen verify-manifests

.PHONY: pull-opm
pull-opm:
	docker pull $(OPERATOR_REGISTRY_IMAGE)

# before running release, bump the version in OLM_VERSION and push to master,
# then tag those builds in quay with the version in OLM_VERSION
release: ver=v$(shell cat OLM_VERSION)
# pull the opm image to get the digest
release: pull-opm manifests
	@echo "Generating the $(ver) release"
	docker pull $(IMAGE_REPO):$(ver)
	$(MAKE) target=upstream ver=$(ver) quickstart=true package

package: olmref=$(shell docker inspect --format='{{index .RepoDigests 0}}' $(IMAGE_REPO):$(ver))
package: opmref=$(shell docker inspect --format='{{index .RepoDigests 0}}' $(OPERATOR_REGISTRY_IMAGE))
package:
ifndef target
	$(error target is undefined)
endif
ifndef ver
	$(error ver is undefined)
endif
	@echo "Getting operator registry image"
	docker pull $(OPERATOR_REGISTRY_IMAGE)
	$(YQ_INTERNAL) w -i deploy/$(target)/values.yaml olm.image.ref $(olmref)
	$(YQ_INTERNAL) w -i deploy/$(target)/values.yaml catalog.image.ref $(olmref)
	$(YQ_INTERNAL) w -i deploy/$(target)/values.yaml package.image.ref $(olmref)
	$(YQ_INTERNAL) w -i deploy/$(target)/values.yaml -- catalog.opmImageArgs "--opmImage=$(opmref)"
	./scripts/package_release.sh $(ver) deploy/$(target)/manifests/$(ver) deploy/$(target)/values.yaml
	ln -sfFn ./$(ver) deploy/$(target)/manifests/latest
ifeq ($(quickstart), true)
	./scripts/package_quickstart.sh deploy/$(target)/manifests/$(ver) deploy/$(target)/quickstart/olm.yaml deploy/$(target)/quickstart/crds.yaml deploy/$(target)/quickstart/install.sh
endif

################################
#  OLM - Install/Uninstall/Run #
################################

.PHONY: uninstall
uninstall:
	@echo Uninstalling OLM:
	- kubectl delete -f deploy/upstream/quickstart/crds.yaml
	- kubectl delete -f deploy/upstream/quickstart/olm.yaml
	- kubectl delete catalogsources.operators.coreos.com
	- kubectl delete clusterserviceversions.operators.coreos.com
	- kubectl delete installplans.operators.coreos.com
	- kubectl delete operatorgroups.operators.coreos.com subscriptions.operators.coreos.com
	- kubectl delete apiservices.apiregistration.k8s.io v1.packages.operators.coreos.com
	- kubectl delete ns olm
	- kubectl delete ns openshift-operator-lifecycle-manager
	- kubectl delete ns openshift-operators
	- kubectl delete ns operators
	- kubectl delete clusterrole.rbac.authorization.k8s.io/aggregate-olm-edit
	- kubectl delete clusterrole.rbac.authorization.k8s.io/aggregate-olm-view
	- kubectl delete clusterrole.rbac.authorization.k8s.io/system:controller:operator-lifecycle-manager
	- kubectl delete clusterroles.rbac.authorization.k8s.io "system:controller:operator-lifecycle-manager"
	- kubectl delete clusterrolebindings.rbac.authorization.k8s.io "olm-operator-binding-openshift-operator-lifecycle-manager"

.PHONY: build-local
build-local: build-linux build-wait build-util-linux
	rm -rf build
	. ./scripts/build_local.sh

.PHONY: run-local
run-local: build-local
	mkdir -p build/resources
	. ./scripts/package_release.sh 1.0.0 build/resources doc/install/local-values.yaml
	. ./scripts/install_local.sh $(LOCAL_NAMESPACE) build/resources
	rm -rf build
