# Get the directory where this Makefile is, so we can use it below for including
# Include the same Bingo variables used by the project
DIR := $(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
# Use openshift's Bingo variables
include $(DIR)/.bingo/Variables.mk

# Definitions for the extended tests

GO_PKG_NAME := github.com/openshift-eng/openshift-tests-extension

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_TREE_STATE := $(shell if git diff --quiet; then echo clean; else echo dirty; fi)

LDFLAGS := -X '$(GO_PKG_NAME)/pkg/version.CommitFromGit=$(GIT_COMMIT)' \
           -X '$(GO_PKG_NAME)/pkg/version.BuildDate=$(BUILD_DATE)' \
           -X '$(GO_PKG_NAME)/pkg/version.GitTreeState=$(GIT_TREE_STATE)'


METADATA := $(shell pwd)/.openshift-tests-extension/openshift_payload_olmv0.json

#SECTION General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '#SECTION' and the
# target descriptions by '#HELP' or '#EXHELP'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: #HELP something, and then pretty-format the target and help. Then,
# if there's a line with #SECTION something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
# The extended-help target uses '#EXHELP' as the delineator.

.PHONY: help
help: #HELP Display essential help.
	@awk 'BEGIN {FS = ":[^#]*#HELP"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_0-9-]+:.*#HELP / { printf "  \033[36m%-17s\033[0m %s\n", $$1, $$2 } ' $(MAKEFILE_LIST)

#SECTION Tests
TOOLS_BIN_DIR := $(CURDIR)/bin

#SECTION Development
.PHONY: verify #HELP To verify the code
verify: tidy fmt vet lint

.PHONY: tidy #HELP Run go mod tidy.
tidy:
	go mod tidy

.PHONY: fmt
fmt: #HELP Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: #HELP Run go vet against code.
	go vet ./...

.PHONY: lint
lint: $(GOLANGCI_LINT) #HELP Run golangci linter.
	$(GOLANGCI_LINT) run

.PHONY: fix-lint
fix-lint: $(GOLANGCI_LINT) #HELP Fix lint issues
	$(GOLANGCI_LINT) run --fix

# Bindata generation
.PHONY: bindata
bindata: $(GO_BINDATA)

bindata: pkg/bindata/qe/bindata.go
pkg/bindata/qe/bindata.go: $(shell find test/qe/testdata -type f)
	mkdir -p $(@D)
	$(GO_BINDATA) -nocompress -nometadata \
		-pkg testdata -o $@ -prefix "testextended" test/qe/testdata/...
	gofmt -s -w $@

# GO_COMPLIANCE_POLICY="exempt_all" must only be used for test related binaries.
# It prevents various FIPS compliance policies from being applied to this compilation.
# Do not set globally.
.PHONY: build
build: #HELP Build the extended tests binary
	@mkdir -p $(TOOLS_BIN_DIR)
	GO_COMPLIANCE_POLICY="exempt_all" go build -ldflags "$(LDFLAGS)" -mod=vendor -o $(TOOLS_BIN_DIR)/olmv0-tests-ext ./cmd/...

.PHONY: update-metadata
update-metadata: #HELP Build and run 'update-metadata' to generate test metadata
	$(TOOLS_BIN_DIR)/olmv0-tests-ext update --component openshift:payload:olmv0
	$(MAKE) clean-metadata


# Ensure TestID is unique over time.
# TestID is built over Product:Type:ComponentName:TestDescription
# (i.e. openshift:payload:olmv0:TestName)
# Details:
# - https://github.com/openshift/enhancements/blob/master/enhancements/testing/openshift-tests-extension.md#test-id
# - https://github.com/openshift-eng/ci-test-mapping
#──────────────────────────────────────────────────────────────
# How to rename a test?
# 1. Run: make list-test-names
# 2. Find the current full test name (e.g. "[sig-abc] My test does XYZ")
# 3. Add a Ginkgo label: ginkgo.Label("original-name:[sig-abc] My test does XYZ")
# 4. Change the test name string and run: make build-update
# **Example**
# It("should pass a renamed sanity check",
# 	Label("original-name:[sig-operator] OLMv0 should pass a trivial sanity check"),
# 	func(ctx context.Context) {
# 		Expect(len("test")).To(BeNumerically(">", 0))
# 	})
# Note: You only add the label once. Do not update it after future renames.
#──────────────────────────────────────────────────────────────
# How to delete a test?
# 1. Run: make list-test-names
# 2. In main.go add:
#    ext.IgnoreObsoleteTests(
#        "[sig-operator] My removed test name",
#    )
# 3. Delete the test code in your suite file (e.g. olmv0.go)
# 4. Run: make build-update.
# This will regenerate the metadata without the test entry.
#────────────────────────────────────────────────────────────────────
.PHONY: build-update
build-update: bindata build update-metadata #HELP Build and update metadata and sanitize output

#SECTION Metadata

.PHONY: list-test-names
list-test-names: #HELP Show current full test names
	@$(TOOLS_BIN_DIR)/olmv0-tests-ext list -o names

# Remove 'codeLocations' to avoid absolute paths like:
# "/Users/$(USER)/go/src/.../olmv0.go:12"
# These are machine-specific and make the metadata non-idempotent.
# More info: https://issues.redhat.com/browse/TRT-2186
.PHONY: clean-metadata
clean-metadata: #HELP Remove 'codeLocations' from metadata JSON
	@echo "Cleaning metadata (removing codeLocations)..."
	@jq 'map(del(.codeLocations))' $(METADATA) > $(METADATA).tmp && mv $(METADATA).tmp $(METADATA)

.PHONY: verify-metadata #HELP To verify that the metadata was properly update
verify-metadata: update-metadata
	@if ! git diff --exit-code $(METADATA); then \
		echo "ERROR: Metadata is out of date. Please run 'make build-update' and commit the result."; \
		exit 1; \
	fi
