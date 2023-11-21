DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
THREESCALE_OPERATOR_TEST_HARNESS_IMAGE ?= $(REG)/$(ORG)/3scale-operator-test-harness:osde2e

.PHONY: image/osde2e/build
image/osde2e/build:
	go mod vendor
	$(CONTAINER_ENGINE) build --platform=$(CONTAINER_PLATFORM) . -f Dockerfile.osde2e -t $(THREESCALE_OPERATOR_TEST_HARNESS_IMAGE)

.PHONY: image/osde2e/push
image/osde2e/push:
	$(CONTAINER_ENGINE) push $(THREESCALE_OPERATOR_TEST_HARNESS_IMAGE)

.PHONY: image/osde2e/build/push
image/osde2e/build/push: image/osde2e/build image/osde2e/push

.PHONY: test/compile/osde2e
test/compile/osde2e:
	cd controllers && CGO_ENABLED=0 go test -mod=readonly -v -c -o ../3scale-operator-test-harness.test ./apps

