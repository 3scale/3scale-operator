.PHONY: build

operator-sdk = ./vendor/github.com/operator-framework/operator-sdk/commands/operator-sdk

Gopkg.lock: Gopkg.toml
	dep check

vendor: Gopkg.lock
	dep ensure -v

$(operator-sdk): vendor

NAMESPACE ?= 3scale/3scale-operator
VERSION ?= latest

build: $(operator-sdk)
	go run $(operator-sdk) build $(NAMESPACE):$(VERSION)

test: $(operator-sdk)
	go run $(operator-sdk) test local $@

