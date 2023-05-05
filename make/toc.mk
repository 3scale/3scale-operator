
##@ TOC generation

.PHONY: toc-user-guide
toc-user-guide: gh-md-toc ## Generate TOC for the main user guide
	$(GH-MD-TOC) --insert --no-backup --hide-footer  $(PROJECT_PATH)/doc/operator-user-guide.md

.PHONY: toc-app-capabilities
toc-app-capabilities: gh-md-toc ## Generate TOC for the app capabilities guide
	$(GH-MD-TOC) --insert --no-backup --hide-footer  $(PROJECT_PATH)/doc/operator-application-capabilities.md

.PHONY: toc-apimanager-reference
toc-apimanager-reference: gh-md-toc ## Generate TOC for the apimanager reference
	$(GH-MD-TOC) --insert --no-backup --hide-footer  $(PROJECT_PATH)/doc/apimanager-reference.md

.PHONY: doc-update-test
doc-update-test:
	git diff --exit-code ./doc
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./doc)" ]
