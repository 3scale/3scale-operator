.PHONY: build

Gopkg.lock: Gopkg.toml
	dep check

vendor: Gopkg.lock
	dep ensure -v

NAMESPACE ?= 3scale/3scale-operator
VERSION ?= latest

build:
	operator-sdk build $(NAMESPACE):$(VERSION)

test:
	operator-sdk test local $@

