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

## vendor: Populate vendor directory
vendor:
	go mod vendor

IMAGE ?= quay.io/3scale/3scale-operator
SOURCE_VERSION ?= master
VERSION ?= v0.0.1
NAMESPACE ?= operator-test

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
	operator-sdk up local --namespace $(NAMESPACE)

## e2e-setup: create OCP project for the operator
e2e-setup:
	oc new-project $(NAMESPACE)

## e2e-local-run: running operator locally with go run instead of as an image in the cluster
e2e-local-run:
	operator-sdk test local ./test/e2e --up-local --namespace $(NAMESPACE) --go-test-flags '-v -timeout 0'

## e2e-run: operator local test
e2e-run:
	operator-sdk test local ./test/e2e --go-test-flags '-v -timeout 0'

## e2e-clean: delete operator OCP project
e2e-clean:
	oc delete --force project $(NAMESPACE) || true

## e2e: e2e-clean e2e-setup e2e-run
e2e: e2e-clean e2e-setup e2e-run
