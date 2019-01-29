.DEFAULT_GOAL := help
.PHONY: build e2e
UNAME := $(shell uname)

ifeq (${UNAME}, Linux)
  SED=sed
else ifeq (${UNAME}, Darwin)
  SED=gsed
endif

help: Makefile
	@sed -n 's/^##//p' $<

## Gopkg.lock: Check toml is sync'ed with lock file
Gopkg.lock: Gopkg.toml
	dep check

## vendor: Populate vendor directory
vendor: Gopkg.lock
	dep ensure -v

IMAGE ?= quay.io/3scale/3scale-operator
VERSION ?= v0.0.1
NAMESPACE ?= operator-test
TEST_IMAGE ?= $(IMAGE):$(VERSION)-$(USER)-test

## build: Build operator
build:
	operator-sdk build $(IMAGE):$(VERSION)

## push: push operator docker image to remote repo
push:
	docker push $(IMAGE):$(VERSION)

## local: push operator docker image to remote repo
local:
	operator-sdk up local --namespace $(NAMESPACE)

## e2e-setup: create OCP project for the operator
e2e-setup:
	oc new-project $(NAMESPACE)

## e2e-run: operator local test
e2e-run:
	operator-sdk test local ./test/e2e --namespace $(NAMESPACE) --up-local --go-test-flags '-v -timeout 0'

## e2e-clean: delete operator OCP project
e2e-clean:
	oc delete --force project $(NAMESPACE) || true

## e2e: e2e-clean e2e-setup e2e-run
e2e: e2e-clean e2e-setup e2e-run
