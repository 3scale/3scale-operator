# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
# Current Operator version
VERSION ?= 0.0.1
# Current Threescale version
THREESCALE_VERSION ?= 2.16
# Address of the container registry
REGISTRY = quay.io
# Organization in container resgistry
ORG ?= 3scale
# Default bundle image tag
IMAGE_TAG_BASE ?= $(REGISTRY)/$(ORG)/3scale-operator
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/3scale/3scale-operator:master

GO ?= go
KUBECTL ?= kubectl
DOCKER ?= docker
NAMESPACE ?= 3scale-test
# Defaults to false, if false, mysql will be used, if true, postgresql will be used
DEV_SYSTEM_DB_POSTGRES ?= false

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif

# Set sed -i as it's different for mac vs gnu
ifeq ($(shell uname -s | tr A-Z a-z), darwin)
	SED_INLINE ?= sed -i ''
else
 	SED_INLINE ?= sed -i
endif

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml
CURRENT_DATE=$(shell date +%s)
LOCAL_RUN_NAMESPACE ?= $(shell oc project -q 2>/dev/null || echo operator-test)
PROMETHEUS_RULES = backend-worker.yaml backend-listener.yaml system-app.yaml system-sidekiq.yaml zync.yaml zync-que.yaml threescale-kube-state-metrics.yaml apicast.yaml
PROMETHEUS_RULES_TARGETS = $(foreach pr,$(PROMETHEUS_RULES),$(PROJECT_PATH)/doc/prometheusrules/$(pr))
PROMETHEUS_RULES_DEPS = $(shell find $(PROJECT_PATH)/pkg/3scale/amp/component -name '*.go')
PROMETHEUS_RULES_NAMESPACE ?= "__NAMESPACE__"

.PHONY: all
all: manager

# Run all tests
.PHONY: test
test: test-unit test-e2e test-crds test-manifests-version

# Run unit tests
TEST_UNIT_PKGS = $(shell $(GO) list ./... | grep -E 'github.com/3scale/3scale-operator/pkg|github.com/3scale/3scale-operator/apis|github.com/3scale/3scale-operator/test/unitcontrollers|github.com/3scale/3scale-operator/controllers/capabilities')
TEST_UNIT_COVERPKGS = $(shell $(GO) list ./... | grep -v github.com/3scale/3scale-operator/test | tr "\n" ",") # Exclude test directories as coverpkg does not accept only-tests packages
.PHONY: test-unit
test-unit: clean-cov generate fmt vet manifests
	mkdir -p "$(PROJECT_PATH)/_output"
	$(GO) test  -v $(TEST_UNIT_PKGS) -covermode=count -coverprofile $(PROJECT_PATH)/_output/unit.cov -coverpkg=$(TEST_UNIT_COVERPKGS)

$(PROJECT_PATH)/_output/unit.cov: test-unit

# Run CRD tests
TEST_CRD_PKGS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/test/crds')
.PHONY: test-crds
test-crds: generate fmt vet manifests
	$(GO) test -v $(TEST_CRD_PKGS)

TEST_MANIFESTS_VERSION_PKGS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/test/manifests-version')
## test-manifests-version: Run manifest version checks
.PHONY: test-manifests-version
test-manifests-version:
	$(GO) test -v $(TEST_MANIFESTS_VERSION_PKGS)

# Run e2e tests
TEST_E2E_PKGS_APPS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/controllers/apps')
TEST_E2E_PKGS_CAPABILITIES = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/controllers/capabilities')
ENVTEST_ASSETS_DIR=$(PROJECT_PATH)/testbin
.PHONY: test-e2e
test-e2e: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh || curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); USE_EXISTING_CLUSTER=true $(GO) test $(TEST_E2E_PKGS_APPS) -coverprofile cover.out -ginkgo.v -v -timeout 0
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); USE_EXISTING_CLUSTER=true $(GO) test $(TEST_E2E_PKGS_CAPABILITIES) -coverprofile cover.out -v -timeout 0


# Build manager binary
.PHONY: manager
manager: manifests generate fmt vet
	$(GO) build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: export WATCH_NAMESPACE=$(LOCAL_RUN_NAMESPACE)
run: export THREESCALE_DEBUG=1
run: export PREFLIGHT_CHECKS_BYPASS=true
run: generate fmt vet manifests
	@-oc process THREESCALE_VERSION=$(THREESCALE_VERSION) -f config/requirements/operator-requirements.yaml | oc apply -f - -n $(WATCH_NAMESPACE)
	$(GO) run ./main.go --zap-devel 

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.7
CONTROLLER_TOOLS_VERSION ?= v0.14.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: cluster/prepare/local
cluster/prepare/local: kustomize cluster/prepare/project install cluster/create/system-redis cluster/create/backend-redis cluster/create/provision-database

.PHONY: cluster/create/system-redis
cluster/create/system-redis:
	sed "s/\$$(NAMESPACE)/$(NAMESPACE)/g" ./config/dev-databases/system-redis/secret.yaml > ./config/dev-databases/system-redis/secret-processed.yaml
	kustomize build ./config/dev-databases/system-redis | kubectl apply -f -
	rm ./config/dev-databases/system-redis/secret-processed.yaml

.PHONY: cluster/create/backend-redis
cluster/create/backend-redis:
	sed "s/\$$(NAMESPACE)/$(NAMESPACE)/g" ./config/dev-databases/backend-redis/secret.yaml > ./config/dev-databases/backend-redis/secret-processed.yaml
	kustomize build ./config/dev-databases/backend-redis | kubectl apply -f -
	rm ./config/dev-databases/backend-redis/secret-processed.yaml

.PHONY: cluster/create/system-postgres
cluster/create/system-postgres:
	sed "s/\$$(NAMESPACE)/$(NAMESPACE)/g" ./config/dev-databases/system-postgresql/secret.yaml > ./config/dev-databases/system-postgresql/secret-processed.yaml
	kustomize build ./config/dev-databases/system-postgresql | kubectl apply -f -
	rm ./config/dev-databases/system-postgresql/secret-processed.yaml

.PHONY: cluster/create/system-mysql
cluster/create/system-mysql:
	sed "s/\$$(NAMESPACE)/$(NAMESPACE)/g" ./config/dev-databases/system-mysql/secret.yaml > ./config/dev-databases/system-mysql/secret-processed.yaml
	kustomize build ./config/dev-databases/system-mysql | kubectl apply -f -
	rm ./config/dev-databases/system-mysql/secret-processed.yaml


.PHONY: cluster/create/provision-database
cluster/create/provision-database:
ifeq ($(DEV_SYSTEM_DB_POSTGRES),true)
	@echo "Using PostgreSQL for system database..."
	$(MAKE) cluster/create/system-postgres
else
	@echo "Using MySQL for system database..."
	$(MAKE) cluster/create/system-mysql
endif

.PHONY: cluster/prepare/project
cluster/prepare/project:
	@ - oc new-project $(NAMESPACE)

OPERATOR_SDK = $(PROJECT_PATH)/bin/operator-sdk
# Note: release file patterns changed after v1.2.0
# More info https://sdk.operatorframework.io/docs/installation/
OPERATOR_SDK_VERSION=v1.36.1
$(OPERATOR_SDK):
	curl -sSL https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$(OS)_$(ARCH) -o $(OPERATOR_SDK)
	chmod +x $(OPERATOR_SDK)

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)

GO_BINDATA=$(PROJECT_PATH)/bin/go-bindata
$(GO_BINDATA):
	$(call go-bin-install,$(GO_BINDATA),github.com/go-bindata/go-bindata/v3/...@v3.1.3)

.PHONY: go-bindata
go-bindata: $(GO_BINDATA)

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

# Install CRDs into a cluster
.PHONY: install
install: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) create -f - || $(KUSTOMIZE) build config/crd | $(KUBECTL) replace -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: manifests $(KUSTOMIZE)
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	$(GO) vet ./...

GOLANGCI-LINT_VERSION ?= v2.1.6
GOLANGCI-LINT = $(PROJECT_PATH)/bin/golangci-lint
$(GOLANGCI-LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_PATH)/bin $(GOLANGCI-LINT_VERSION)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI-LINT) ## Download golangci-lint locally if necessary.

.PHONY: run-lint
lint: $(GOLANGCI-LINT) ## Run lint tests
	$(GOLANGCI-LINT) run

# Generate code
generate: $(CONTROLLER_GEN) $(GO_BINDATA)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	@echo Generate Go embedded assets files by processing source
	export PATH=$(PROJECT_PATH)/bin:$$PATH;	$(GO) generate github.com/3scale/3scale-operator/pkg/assets

# Build the docker image
.PHONY: docker-build
docker-build: test docker-build-only

.PHONY: docker-build-only
docker-build-only:
	$(DOCKER) build . -t ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> than the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile .
	- docker buildx rm project-v3-builder

# Push the operator docker image
.PHONY: operator-image-push
operator-image-push:
	$(DOCKER) push ${IMG}

# Push the bundle docker image
.PHONY: bundle-image-push
bundle-image-push:
	$(DOCKER) push ${BUNDLE_IMG}

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests $(KUSTOMIZE) $(OPERATOR_SDK)
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build: bundle-validate
	$(DOCKER) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-validate-image
bundle-validate-image: $(OPERATOR_SDK)
	$(OPERATOR_SDK) bundle validate $(BUNDLE_IMG)

.PHONY: bundle-custom-updates
bundle-custom-updates: BUNDLE_PREFIX=dev$(CURRENT_DATE)
bundle-custom-updates: $(YQ)
	@echo "Update metadata to avoid collision with existing 3scale Operator official public operators catalog entries"
	@echo "using BUNDLE_PREFIX $(BUNDLE_PREFIX)"
	$(YQ) --inplace '.metadata.name = "$(BUNDLE_PREFIX)-3scale-operator.$(VERSION)"' $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.spec.displayName = "$(BUNDLE_PREFIX) 3scale operator"' $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.spec.provider.name = "$(BUNDLE_PREFIX)"' $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.annotations."operators.operatorframework.io.bundle.package.v1" = "$(BUNDLE_PREFIX)-3scale-operator"' $(PROJECT_PATH)/bundle/metadata/annotations.yaml
	$(SED_INLINE) -E 's/(operators\.operatorframework\.io\.bundle\.package\.v1=).+/\1$(BUNDLE_PREFIX)-3scale-operator/' $(PROJECT_PATH)/bundle.Dockerfile
	@echo "Update operator image reference URL"
	$(YQ) --inplace '.metadata.annotations.containerImage = "$(IMG)"' $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = "$(IMG)"' $(PROJECT_PATH)/bundle/manifests/3scale-operator.clusterserviceversion.yaml

.PHONY: bundle-restore
bundle-restore:
	git checkout bundle/manifests/3scale-operator.clusterserviceversion.yaml bundle/metadata/annotations.yaml bundle.Dockerfile

.PHONY: bundle-custom-build
bundle-custom-build: | bundle-custom-updates bundle-build bundle-restore

.PHONY: bundle-run
bundle-run: $(OPERATOR_SDK)
	$(OPERATOR_SDK) run bundle --namespace openshift-marketplace $(BUNDLE_IMG)

# 3scale-specific targets

# find or download yq
# download yq if necessary
YQ=$(PROJECT_PATH)/bin/yq
$(YQ):
	$(call go-bin-install,$(YQ),github.com/mikefarah/yq/v4@latest)

.PHONY: yq
yq: $(YQ)

OPM = ./bin/opm
.PHONY: opm
opm:
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else 
OPM = $(shell which opm)
endif
endif
BUNDLE_IMGS ?= $(BUNDLE_IMG) 
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION) ifneq ($(origin CATALOG_BASE_IMG), undefined) FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG) endif 
.PHONY: catalog-build
catalog-build: opm
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

.PHONY: catalog-push
catalog-push: ## Push the catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

.PHONY: download
download:
	@echo Download go.mod dependencies
	@$(GO) mod download

## licenses.xml: Generate licenses.xml file
.PHONY: licenses.xml
licenses.xml: $(DEPENDENCY_DECISION_FILE)
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	license_finder report --decisions-file=$(DEPENDENCY_DECISION_FILE) --quiet --format=xml > licenses.xml

## licenses-check: Check license compliance of dependencies
.PHONY: licenses-check
licenses-check:
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	@echo "Checking license compliance"
	license_finder --decisions-file=$(DEPENDENCY_DECISION_FILE)

.PHONY: assets-update-test
assets-update-test: generate fmt
	git diff --exit-code ./pkg/assets
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./pkg/assets)" ]

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
bundle-validate: $(OPERATOR_SDK)
	$(OPERATOR_SDK) bundle validate ./bundle

# Since operator-sdk 1.26.0, `make bundle` changes the `createdAt` field from the bundle
# even if it is patched:
#   https://github.com/operator-framework/operator-sdk/pull/6136
# This code checks if only the createdAt field. If is the only change, it is ignored.
# Else, it will do nothing.
# https://github.com/operator-framework/operator-sdk/issues/6285#issuecomment-1415350333
# https://github.com/operator-framework/operator-sdk/issues/6285#issuecomment-1532150678
.PHONY: bundle-update-test
bundle-update-test:
	git diff --quiet -I'^    createdAt: ' ./bundle && git checkout ./bundle || true
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./bundle)" ]

$(PROMETHEUS_RULES_TARGETS): $(PROMETHEUS_RULES_DEPS)
	go run $(PROJECT_PATH)/pkg/3scale/amp/main.go prometheusrules --namespace $(PROMETHEUS_RULES_NAMESPACE) $(notdir $(basename $@)) >$@

.PHONY: prometheus-rules
prometheus-rules: prometheus-rules-clean $(PROMETHEUS_RULES_TARGETS)

.PHONY: prometheus-rules-clean
prometheus-rules-clean:
	rm -f $(PROMETHEUS_RULES_TARGETS)

.PHONY: prometheusrules-update-test
prometheusrules-update-test: prometheus-rules
	git diff --exit-code ./doc/prometheusrules
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./doc/prometheusrules)" ]

# go-bin-install will 'go get' any package $2 and install it to $1.
define go-bin-install
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_PATH)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Include last to avoid changing MAKEFILE_LIST used above
include $(PROJECT_PATH)/make/*.mk

