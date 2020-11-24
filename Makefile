SHELL := /bin/bash
# Current Operator version
VERSION ?= 0.0.1
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/3scale/3scale-operator:master

CRD_OPTIONS ?= "crd:crdVersions=v1"

GO ?= go
KUBECTL ?= kubectl
OPERATOR_SDK ?= operator-sdk
DOCKER ?= docker

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml
CURRENT_DATE=$(shell date +%s)
LOCAL_RUN_NAMESPACE ?= $(shell oc project -q 2>/dev/null || echo operator-test)

all: manager

# Run all tests
test: test-unit test-e2e test-crds test-manifests-version

# Run unit tests
TEST_UNIT_PKGS = $(shell $(GO) list ./... | grep -E 'github.com/3scale/3scale-operator/pkg|github.com/3scale/3scale-operator/apis|github.com/3scale/3scale-operator/test/unitcontrollers')
TEST_UNIT_COVERPKGS = $(shell $(GO) list ./... | grep -v test/unitcontrollers | tr "\n" ",") # Exclude test/unitcontrollers directory as coverpkg does not accept only-tests packages
test-unit: clean-cov generate fmt vet manifests
	mkdir -p "$(PROJECT_PATH)/_output"
	$(GO) test  -v $(TEST_UNIT_PKGS) -covermode=count -coverprofile $(PROJECT_PATH)/_output/unit.cov -coverpkg=$(TEST_UNIT_COVERPKGS)

$(PROJECT_PATH)/_output/unit.cov: test-unit

# Run CRD tests
TEST_CRD_PKGS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/test/crds')
test-crds: generate fmt vet manifests
	$(GO) test -v $(TEST_CRD_PKGS)

TEST_MANIFESTS_VERSION_PKGS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/test/manifests-version')
## test-manifests-version: Run manifest version checks
test-manifests-version:
	$(GO) test -v $(TEST_MANIFESTS_VERSION_PKGS)

# Run e2e tests
TEST_E2E_PKGS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/controllers')
ENVTEST_ASSETS_DIR=$(PROJECT_PATH)/testbin
test-e2e: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh || curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.6.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); USE_EXISTING_CLUSTER=true $(GO) test $(TEST_E2E_PKGS) -coverprofile cover.out -ginkgo.v -ginkgo.progress -v -timeout 0

# Build manager binary
manager: generate fmt vet
	$(GO) build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: export WATCH_NAMESPACE=$(LOCAL_RUN_NAMESPACE)
run: export THREESCALE_DEBUG=1
run: generate fmt vet manifests
	$(GO) run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
.PHONY: docker-build
docker-build: test docker-build-only

.PHONY: docker-build-only
docker-build-only:
	$(DOCKER) build . -t ${IMG}

# Push the operator docker image
.PHONY: operator-image-push
operator-image-push:
	$(DOCKER) push ${IMG}

# Push the bundle docker image
.PHONY: bundle-image-push
bundle-image-push:
	$(DOCKER) push ${BUNDLE_IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	$(GO) mod init tmp ;\
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	$(GO) mod init tmp ;\
	$(GO) get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build: bundle-validate
	$(DOCKER) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-validate-image
bundle-validate-image:
	$(OPERATOR_SDK) bundle validate $(BUNDLE_IMG)

.PHONY: bundle-custom-updates
bundle-custom-updates: BUNDLE_PREFIX=dev$(CURRENT_DATE)
bundle-custom-updates: yq
	@echo "Update metadata to avoid collision with existing 3scale Operator official public operators catalog entries"
	@echo "using BUNDLE_PREFIX $(BUNDLE_PREFIX)"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml metadata.name $(BUNDLE_PREFIX)-3scale-operator.$(VERSION)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml spec.displayName "$(BUNDLE_PREFIX) 3scale operator"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml spec.provider.name $(BUNDLE_PREFIX)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/metadata/annotations.yaml 'annotations."operators.operatorframework.io.bundle.package.v1"' $(BUNDLE_PREFIX)-3scale-operator
	sed -E -i 's/(operators\.operatorframework\.io\.bundle\.package\.v1=).+/\1$(BUNDLE_PREFIX)-3scale-operator/' $(PROJECT_PATH)/bundle.Dockerfile
	@echo "Update operator image reference URL"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml metadata.annotations.containerImage $(IMG)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml spec.install.spec.deployments[0].spec.template.spec.containers[1].image $(IMG)

.PHONY: bundle-restore
bundle-restore:
	git checkout bundle/manifests/3scale-operator.clusterserviceversion.yaml bundle/metadata/annotations.yaml bundle.Dockerfile

.PHONY: bundle-custom-build
bundle-custom-build: | bundle-custom-updates bundle-build bundle-restore

.PHONY: bundle-run
bundle-run:
	$(OPERATOR_SDK) run bundle --namespace $(LOCAL_RUN_NAMESPACE) $(BUNDLE_IMG)

# 3scale-specific targets

# find or download yq
# download yq if necessary
yq:
ifeq (, $(shell command -v yq 2> /dev/null))
	@{ \
	set -e ;\
	YQ_TMP_DIR=$$(mktemp -d) ;\
	cd $$YQ_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/mikefarah/yq/v3 ;\
	rm -rf $$YQ_TMP_DIR ;\
	}
YQ=$(GOBIN)/yq
else
YQ=$(shell command -v yq 2> /dev/null)
endif

download:
	@echo Download go.mod dependencies
	@$(GO) mod download

## licenses.xml: Generate licenses.xml file
licenses.xml: $(DEPENDENCY_DECISION_FILE)
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	license_finder report --decisions-file=$(DEPENDENCY_DECISION_FILE) --quiet --format=xml > licenses.xml

## licenses-check: Check license compliance of dependencies
licenses-check:
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	@echo "Checking license compliance"
	license_finder --decisions-file=$(DEPENDENCY_DECISION_FILE)

go-bindata:
ifeq (, $(shell which go-bindata))
	@{ \
	set -e ;\
	GOBINDATA_TMP_DIR=$$(mktemp -d) ;\
	cd $$GOBINDATA_TMP_DIR ;\
	$(GO) mod init tmp ;\
	$(GO) get github.com/go-bindata/go-bindata/v3/...@v3.1.3 ;\
	rm -rf $$GOBINDATA_TMP_DIR ;\
	}
GOBINDATA=$(GOBIN)/go-bindata
else
GOBINDATA=$(shell which go-bindata)
endif

## assets: Generate embedded assets
assets: go-bindata
	@echo Generate Go embedded assets files by processing source
	$(GO) generate github.com/3scale/3scale-operator/pkg/assets

## templates: generate templates
TEMPLATES_MAKEFILE_PATH = $(PROJECT_PATH)/pkg/3scale/amp
templates:
	$(MAKE) -C $(TEMPLATES_MAKEFILE_PATH) clean all

## coverage_analysis: Analyze coverage via a browse
.PHONY: coverage_analysis
coverage_analysis: $(PROJECT_PATH)/_output/unit.cov
	$(GO) tool cover -html="$(PROJECT_PATH)/_output/unit.cov"

## coverage_total_report: Simple coverage report
.PHONY: coverage_total_report
coverage_total_report: $(PROJECT_PATH)/_output/unit.cov
	@$(GO) tool cover -func=$(PROJECT_PATH)/_output/unit.cov | grep total | awk '{print $$3}'

clean-cov:
	rm -rf $(PROJECT_PATH)/_output
	rm -rf $(PROJECT_PATH)/cover.out

.PHONY: bundle-validate
bundle-validate:
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-update-test
bundle-update-test:
	git diff --exit-code ./bundle
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./bundle)" ]
