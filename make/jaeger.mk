##@ Jaeger

jaeger-deploy:
	$(KUSTOMIZE) build config/jaeger | $(KUBECTL) apply -f -

apicast-opentelemetry-deploy:
	$(KUSTOMIZE) build examples/opentelemetry | $(KUBECTL) apply -f -