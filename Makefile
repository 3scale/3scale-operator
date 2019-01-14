.PHONY: build e2e
UNAME := $(shell uname)

ifeq (${UNAME}, Linux)
  SED=sed
else ifeq (${UNAME}, Darwin)
  SED=gsed
endif

Gopkg.lock: Gopkg.toml
	dep check

vendor: Gopkg.lock
	dep ensure -v

IMAGE ?= quay.io/3scale/3scale-operator
VERSION ?= v0.0.1
NAMESPACE ?= operator-test
TEST_IMAGE ?= $(IMAGE):$(VERSION)-$(USER)-test

build:
	operator-sdk build $(IMAGE):$(VERSION)

push:
	docker push $(IMAGE):$(VERSION)

build-test:
	operator-sdk build $(TEST_IMAGE)

push-test:
	docker push $(TEST_IMAGE)

e2e-setup:
	oc new-project $(NAMESPACE)

e2e-run:
	cat deploy/operator.orig | sed "s@REPLACE_IMAGE@$(TEST_IMAGE)@g" > deploy/operator.yaml
	operator-sdk test local ./test/e2e --namespace $(NAMESPACE) --go-test-flags "-v"

e2e-clean:
	oc delete --force project $(NAMESPACE) || true

e2e: build-test push-test e2e-clean e2e-setup e2e-run e2e-clean
