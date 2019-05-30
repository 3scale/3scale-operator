.DEFAULT_GOAL := help
SHELL := bash -eo pipefail

templates = amp.yml amp-s3.yml amp-eval.yml amp-eval-s3.yml amp-ha.yml amp-postgresql.yml

targets = $(foreach template,$(templates),./auto-generated-templates/amp/$(template) ./auto-generated-templates/productized_templates/amp/$(template))

.DELETE_ON_ERROR: $(targets)

DEPS = ../../../vendor $(shell find . -name '*.go')

all: $(targets) ## Generate all templates

clean: # Remove all generated templates
	rm -f $(targets)

component-name = $(notdir $(basename $(1)))-template

$(foreach t,$(templates),./auto-generated-templates/amp/$(t)): $(DEPS)
	go run main.go template $(call component-name,$@) >$@

$(foreach t,$(templates),./auto-generated-templates/productized_templates/amp/$(t)): $(DEPS)
	go run main.go template $(call component-name,$@)+productized >$@

# Check http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test:
	$(MAKE) all
	git diff --exit-code ./auto-generated-templates
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./auto-generated-templates)" ]
