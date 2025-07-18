SHELL := /bin/bash
# Current Operator version
VERSION ?= 0.0.1
# Current Threescale version
THREESCALE_VERSION ?= 2.16
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

OS := $(shell uname | awk '{print tolower($$0)}' | sed -e s/linux/linux-gnu/ )
ARCH := $(shell uname -m)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/3scale/3scale-operator:master

CRD_OPTIONS ?= "crd:crdVersions=v1"

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

all: manager

# Run all tests
test: test-unit test-e2e test-crds test-manifests-version

# Run unit tests
TEST_UNIT_PKGS = $(shell $(GO) list ./... | grep -E 'github.com/3scale/3scale-operator/pkg|github.com/3scale/3scale-operator/apis|github.com/3scale/3scale-operator/test/unitcontrollers|github.com/3scale/3scale-operator/controllers/capabilities')
TEST_UNIT_COVERPKGS = $(shell $(GO) list ./... | grep -v github.com/3scale/3scale-operator/test | tr "\n" ",") # Exclude test directories as coverpkg does not accept only-tests packages
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
TEST_E2E_PKGS_APPS = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/controllers/apps')
TEST_E2E_PKGS_CAPABILITIES = $(shell $(GO) list ./... | grep 'github.com/3scale/3scale-operator/controllers/capabilities')
ENVTEST_ASSETS_DIR=$(PROJECT_PATH)/testbin
test-e2e: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh || curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); USE_EXISTING_CLUSTER=true $(GO) test $(TEST_E2E_PKGS_APPS) -coverprofile cover.out -ginkgo.v -v -timeout 0
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); USE_EXISTING_CLUSTER=true $(GO) test $(TEST_E2E_PKGS_CAPABILITIES) -coverprofile cover.out -v -timeout 0


# Build manager binary
manager: generate fmt vet
	$(GO) build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: export WATCH_NAMESPACE=$(LOCAL_RUN_NAMESPACE)
run: export THREESCALE_DEBUG=1
run: export PREFLIGHT_CHECKS_BYPASS=true
run: generate fmt vet manifests
	@-oc process THREESCALE_VERSION=$(THREESCALE_VERSION) -f config/requirements/operator-requirements.yaml | oc apply -f - -n $(WATCH_NAMESPACE)
	$(GO) run ./main.go --zap-devel 

# find or download controller-gen
# download controller-gen if necessary
CONTROLLER_GEN=$(PROJECT_PATH)/bin/controller-gen
$(CONTROLLER_GEN):
	$(call go-bin-install,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)

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

KUSTOMIZE=$(PROJECT_PATH)/bin/kustomize
$(KUSTOMIZE):
	$(call go-bin-install,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.7)

.PHONY: kustomize
kustomize: $(KUSTOMIZE)

OPERATOR_SDK = $(PROJECT_PATH)/bin/operator-sdk
# Note: release file patterns changed after v1.2.0
# More info https://sdk.operatorframework.io/docs/installation/
OPERATOR_SDK_VERSION=v1.2.0
$(OPERATOR_SDK):
	curl -sSL https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk-${OPERATOR_SDK_VERSION}-$(ARCH)-${OS} -o $(OPERATOR_SDK)
	chmod +x $(OPERATOR_SDK)

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)

GO_BINDATA=$(PROJECT_PATH)/bin/go-bindata
$(GO_BINDATA):
	$(call go-bin-install,$(GO_BINDATA),github.com/go-bindata/go-bindata/v3/...@v3.1.3)

.PHONY: go-bindata
go-bindata: $(GO_BINDATA)

# Install CRDs into a cluster
install: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) create -f - || $(KUSTOMIZE) build config/crd | $(KUBECTL) replace -f -

# Uninstall CRDs from a cluster
uninstall: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests $(KUSTOMIZE)
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
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

.PHONY: bundle-update-test
bundle-update-test:
	git diff --exit-code ./bundle
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
