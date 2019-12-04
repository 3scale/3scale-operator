MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
.DEFAULT_GOAL := help
UNAME := $(shell uname)

ifeq (${UNAME}, Linux)
  SED=sed
else ifeq (${UNAME}, Darwin)
  SED=gsed
endif

OPERATORCOURIER := $(shell command -v operator-courier 2> /dev/null)
LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml
IMAGE ?= quay.io/3scale/3scale-operator
SOURCE_VERSION ?= master
VERSION ?= v0.0.1
NAMESPACE ?= operator-test
OPERATOR_NAME ?= threescale-operator
MANIFEST_RELEASE ?= 1.0.$(shell git rev-list --count master)
APPLICATION_REPOSITORY_NAMESPACE ?= 3scaleoperatormaster

help: Makefile
	@sed -n 's/^##//p' $<

## vendor: Populate vendor directory
vendor:
	@GO111MODULE=on go mod vendor

## build: Build operator
.PHONY: build
build:
	operator-sdk build $(IMAGE):$(VERSION)

## push: push operator docker image to remote repo
.PHONY: push
push:
	docker push $(IMAGE):$(VERSION)

## pull: pull operator docker image from remote repo
.PHONY: pull
pull:
	docker pull $(IMAGE):$(VERSION)

.PHONY: tag
tag:
	docker tag $(IMAGE):$(SOURCE_VERSION) $(IMAGE):$(VERSION)

## local: push operator docker image to remote repo
.PHONY: local
local:
	OPERATOR_NAME=$(OPERATOR_NAME) operator-sdk up local --namespace $(NAMESPACE)

#
# Tests
#
INTEGRATION_TEST_GO_TEST_FLAGS = -tags=integration -coverpkg ./... -coverprofile $(PROJECT_PATH)/integration.cov -covermode=count -v -timeout 0

## e2e-setup: create OCP project for the operator
.PHONY: e2e-setup
e2e-setup:
	oc new-project $(NAMESPACE)

## e2e: e2e-clean e2e-setup e2e-run
.PHONY: e2e
e2e: e2e-clean e2e-setup e2e-run

$(PROJECT_PATH)/unit.cov:
	go test ./pkg/... -v -tags=unit -covermode=count -coverprofile $(PROJECT_PATH)/unit.cov -coverpkg ./...

# generated using local run (--up-local) mode
$(PROJECT_PATH)/integration.cov:
	OPERATOR_NAME=$(OPERATOR_NAME) operator-sdk test local ./test/e2e --up-local --namespace $(NAMESPACE) --go-test-flags '$(INTEGRATION_TEST_GO_TEST_FLAGS)'

$(PROJECT_PATH)/coverage.txt: $(PROJECT_PATH)/unit.cov $(PROJECT_PATH)/integration.cov
	@echo "mode: count" > $(PROJECT_PATH)/coverage.txt
	@tail -q -n +2 $^ >> coverage.txt

## e2e-local-run: operator integration tests locally with go run instead of as an image in the cluster
.PHONY: e2e-local-run
e2e-local-run: $(PROJECT_PATH)/integration.cov

## e2e-run: operator integration tests
.PHONY: e2e-run
e2e-run:
	operator-sdk test local ./test/e2e --go-test-flags '$(INTEGRATION_TEST_GO_TEST_FLAGS)' --debug --image $(IMAGE) --namespace $(NAMESPACE)

## unit: Run unit tests in pkg directory
.PHONY: unit
unit: $(PROJECT_PATH)/unit.cov

## coverage_analysis: Analyze coverage via a browse
.PHONY: coverage_analysis
coverage_analysis: $(PROJECT_PATH)/coverage.txt
	go tool cover -html="$(PROJECT_PATH)/coverage.txt"

## coverage_total_report: Simple coverage report
.PHONY: coverage_total_report
coverage_total_report: $(PROJECT_PATH)/coverage.txt
	@go tool cover -func=$(PROJECT_PATH)/coverage.txt | grep total | awk '{print $$3}'

## test-crds: Run CRD unittests
.PHONY: test-crds
test-crds:
	cd $(PROJECT_PATH)/test/crds && go test -v

#
# Operator Bundle tasks
#

## verify-manifest: Test manifests have expected format
.PHONY: verify-manifest
verify-manifest:
ifndef OPERATORCOURIER
	$(error "operator-courier is not available please install pip3 install operator-courier")
endif
	cd $(PROJECT_PATH)/deploy/olm-catalog && operator-courier verify --ui_validate_io 3scale-operator-master/

## push-manifest: Push manifests to application repository
.PHONY: push-manifest
push-manifest:
ifndef OPERATORCOURIER
	$(error "operator-courier is not available please install pip3 install operator-courier")
endif
	cd $(PROJECT_PATH)/deploy/olm-catalog && operator-courier push 3scale-operator-master/ $(APPLICATION_REPOSITORY_NAMESPACE) 3scale-operator-master $(MANIFEST_RELEASE) "$(TOKEN)"

#
# Licensing tasks
#

## licenses.xml: Generate licenses.xml file
licenses.xml:
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	license_finder report --decisions-file=$(DEPENDENCY_DECISION_FILE) --quiet --format=xml > licenses.xml

## licenses-check: Check license compliance of dependencies
.PHONY: licenses-check
licenses-check: vendor
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	@echo "Checking license compliance"
	license_finder --decisions-file=$(DEPENDENCY_DECISION_FILE)

#
# Clean up
#

## clean: Clean build resources
clean:
	rm -rf $(PROJECT_PATH)/coverage.txt  $(PROJECT_PATH)/unit.cov $(PROJECT_PATH)/integration.cov

## e2e-clean: delete operator OCP project
e2e-clean:
	oc delete --force project $(NAMESPACE) || true
