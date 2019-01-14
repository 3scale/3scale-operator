.PHONY: build local-e2e
UNAME := $(shell uname)

ifeq (${UNAME}, Linux)
  SED=sed
else ifeq (${UNAME}, Darwin)
  SED=gsed
endif

operator-sdk = ./vendor/github.com/operator-framework/operator-sdk/commands/operator-sdk

Gopkg.lock: Gopkg.toml
	dep check

vendor: Gopkg.lock
	dep ensure -v

$(operator-sdk): vendor

IMAGE ?= quay.io/3scale/3scale-operator
VERSION ?= v0.0.1
NAMESPACE ?= operator-test
TEST_IMAGE ?= $(IMAGE):$(VERSION)-$(USER)-test

build: $(operator-sdk)
	go run $(operator-sdk) build $(IMAGE):$(VERSION)

push: 
	docker push $(IMAGE):$(VERSION)

build-test: $(operator-sdk)
	go run $(operator-sdk) build $(TEST_IMAGE)

push-test:
	docker push $(TEST_IMAGE)

local-e2e-setup: $(operator-sdk)
	oc new-project $(NAMESPACE)

local-e2e-run: $(operator-sdk)
	cat deploy/operator.orig | sed "s@REPLACE_IMAGE@$(TEST_IMAGE)@g" > deploy/operator.yaml
	go run $(operator-sdk) test local ./test/e2e --namespace $(NAMESPACE) --go-test-flags "-v"

local-e2e-clean: $(operator-sdk)
	oc delete --force project $(NAMESPACE) || true

local-e2e: build-test push-test local-e2e-clean local-e2e-setup local-e2e-run local-e2e-clean
