SHELL := /bin/bash
ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

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

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

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
