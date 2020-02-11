MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
.DEFAULT_GOAL := help
.PHONY: build unit e2e test-crds verify-manifest licenses-check push-manifest
UNAME := $(shell uname)

ifeq (${UNAME}, Linux)
  SED=sed
else ifeq (${UNAME}, Darwin)
  SED=gsed
endif

OPERATORCOURIER := $(shell command -v operator-courier 2> /dev/null)
LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml

help: Makefile
	@sed -n 's/^##//p' $<

## vendor: Populate vendor directory
vendor:
	@GO111MODULE=on go mod vendor

IMAGE ?= quay.io/3scale/3scale-operator
SOURCE_VERSION ?= master
VERSION ?= v0.0.1
NAMESPACE ?= $(shell oc project -q 2>/dev/null || echo operator-test)
OPERATOR_NAME ?= threescale-operator
MANIFEST_RELEASE ?= 1.0.$(shell git rev-list --count master)
APPLICATION_REPOSITORY_NAMESPACE ?= 3scaleoperatormaster
TEMPLATES_MAKEFILE_PATH = $(PROJECT_PATH)/pkg/3scale/amp

## build: Build operator
build:
	operator-sdk build $(IMAGE):$(VERSION)

## push: push operator docker image to remote repo
push:
	docker push $(IMAGE):$(VERSION)

## pull: pull operator docker image from remote repo
pull:
	docker pull $(IMAGE):$(VERSION)

tag:
	docker tag $(IMAGE):$(SOURCE_VERSION) $(IMAGE):$(VERSION)

## local: push operator docker image to remote repo
local:
	OPERATOR_NAME=$(OPERATOR_NAME) operator-sdk up local --namespace $(NAMESPACE)

## e2e-setup: create OCP project for the operator
e2e-setup:
	oc new-project $(NAMESPACE)

## e2e-local-run: running operator locally with go run instead of as an image in the cluster
e2e-local-run:
	OPERATOR_NAME=$(OPERATOR_NAME) operator-sdk test local ./test/e2e --up-local --namespace $(NAMESPACE) --go-test-flags '-v -timeout 0'

## e2e-run: operator local test
e2e-run:
	operator-sdk test local ./test/e2e --go-test-flags '-v -timeout 0' --debug --image $(IMAGE) --namespace $(NAMESPACE)

## e2e-clean: delete operator OCP project
e2e-clean:
	oc delete --force project $(NAMESPACE) || true

## e2e: e2e-clean e2e-setup e2e-run
e2e: e2e-clean e2e-setup e2e-run

$(PROJECT_PATH)/_output/unit.cov:
	mkdir -p "$(PROJECT_PATH)/_output"
	go test ./pkg/... -v -tags=unit -covermode=count -coverprofile $(PROJECT_PATH)/_output/unit.cov -coverpkg ./...

## unit: Run unit tests in pkg directory
.PHONY: unit
unit: $(PROJECT_PATH)/_output/unit.cov

## coverage_analysis: Analyze coverage via a browse
.PHONY: coverage_analysis
coverage_analysis: $(PROJECT_PATH)/_output/unit.cov
	go tool cover -html="$(PROJECT_PATH)/_output/unit.cov"

## coverage_total_report: Simple coverage report
.PHONY: coverage_total_report
coverage_total_report: $(PROJECT_PATH)/_output/unit.cov
	@go tool cover -func=$(PROJECT_PATH)/_output/unit.cov | grep total | awk '{print $$3}'

## test-crds: Run CRD unittests
test-crds:
	cd $(PROJECT_PATH)/test/crds && go test -v

## verify-manifest: Test manifests have expected format
verify-manifest:
ifndef OPERATORCOURIER
	$(error "operator-courier is not available please install pip3 install operator-courier")
endif
	cd $(PROJECT_PATH)/deploy/olm-catalog && operator-courier verify --ui_validate_io 3scale-operator-master/

## licenses.xml: Generate licenses.xml file
licenses.xml:
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	license_finder report --decisions-file=$(DEPENDENCY_DECISION_FILE) --quiet --format=xml > licenses.xml

## licenses-check: Check license compliance of dependencies
licenses-check: vendor
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	@echo "Checking license compliance"
	license_finder --decisions-file=$(DEPENDENCY_DECISION_FILE)

## push-manifest: Push manifests to application repository
push-manifest:
ifndef OPERATORCOURIER
	$(error "operator-courier is not available please install pip3 install operator-courier")
endif
	cd $(PROJECT_PATH)/deploy/olm-catalog && operator-courier push 3scale-operator-master/ $(APPLICATION_REPOSITORY_NAMESPACE) 3scale-operator-master $(MANIFEST_RELEASE) "$(TOKEN)"

## templates: generate templates
templates:
	$(MAKE) -C $(TEMPLATES_MAKEFILE_PATH) clean all

## upgrade-tool: Build upgrade tool
upgrade-tool: $(TEMPLATES_MAKEFILE_PATH)/main.go $(TEMPLATES_MAKEFILE_PATH)/cmd/upgrader.go $(shell find $(TEMPLATES_MAKEFILE_PATH)/cmd/upgrader -name '*.go')
	go build -o $@ $<

## clean: Clean build resources
clean:
	rm -rf $(PROJECT_PATH)/_output
