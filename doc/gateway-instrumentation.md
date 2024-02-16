## Gateway instrumentation

3scale Operator can enable insturmentation on it's managed APIcasts by using the [OpenTelemetry](https://opentelemetry.io/) SDK.
More specifically, enabling the [Nginx opentelemetry tracing library](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx).

It works with Jaeger since version **1.35**.  If the existing collector does not support
OpenTelemetry traces, an OpenTelemetry Collector is required as tracing proxy.

Supported propagation types: [W3C](https://w3c.github.io/trace-context/)

### Prerequisites

* Opentelemetry Collector supporting the APIcast exporter. Currently, the only implemeneted [exporter](https://opentelemetry.io/docs/reference/specification/protocol/exporter/)
in APIcast is OTLP over gRPC `OTLP/gRPC`. Even though OpenTelemetry specification supports also OTLP over HTTP (`OTLP/HTTP`),
this exporter is not included in APIcast. It works with Jaeger since version **1.35**.

For dev/testing purposes, you can deploy quick&easy Jaeger with few commands. **Not suitable** for production use, though.
Create and connect with the cluster.

```
❯ make jaeger-deploy
```

That should deploy Jaeger service listening at `jaeger:4317`.

### Create secret with the APIcast instrumentation configuration

The configuration file specification is defined in the [Nginx instrumentation library repo](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx).

`otlp` is the only supported exporter.

The host/port address is set according to the dev/testing deployment defined in this repo.
Change it to whatever you have your collector deployed.

The name of the secret. `otel-config` in the example, will be used in the APIcast CR.

```yaml
oc apply -f - <<EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: otel-config
type: Opaque
stringData:
  config.toml: |
    exporter = "otlp"
    processor = "simple"
    [exporters.otlp]
    # Alternatively the OTEL_EXPORTER_OTLP_ENDPOINT environment variable can also be used.
    host = "jaeger"
    port = 4317
    # Optional: enable SSL, for endpoints that support it
    # use_ssl = true
    # Optional: set a filesystem path to a pem file to be used for SSL encryption
    # (when use_ssl = true)
    # ssl_cert_path = "/path/to/cert.pem"
    [processors.batch]
    max_queue_size = 2048
    schedule_delay_millis = 5000
    max_export_batch_size = 512
    [service]
    name = "apicast-staging" # Opentelemetry resource name <in this example, apicast staging is used
EOF
```

### Deploy API Manager with opentelemetry instrumentation on APIcasts

Only relevant content shown. Check out the [APIManager CRD reference](apimanager-reference.md) for
a comprehensive list of options.

Install the operator.

Create S3 sample secret:
```yaml
oc apply -f - <<EOF   
---
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: s3-credentials
stringData:
  AWS_ACCESS_KEY_ID: something
  AWS_SECRET_ACCESS_KEY: something
  AWS_BUCKET: something
  AWS_REGION: us-east-1
type: Opaque
EOF
```
Create APIManager with Opentelemetry enabled:

```
oc apply -f - <<EOF
---
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: $DOMAIN
  system:
    fileStorage:
      simpleStorageService:
        configurationSecretRef:
          name: s3-credentials
  apicast:
    stagingSpec:
      openTelemetry:
        enabled: true
        tracingConfigSecretRef:
            name: otel-config
EOF
```
### Verification steps for the opentelemetry instrumentation

Testing 3scale Operator APIcast instrumentation in the dev/testing environment should be easy.

* Login to 3scale portal and create backend and product with basic configuration
* Promote product to stage environment
* Copy sample curl command from configuration section of 3scale porta
* Hit the endpoint

Note that upstream echo'ed request headers show `Traceparent` W3C standard tracing header.

Open in local browser jaeger dashboard

```
❯ oc port-forward service/jaeger 16686&
❯ open http://127.0.0.1:16686
```

Hit "Find Traces" with `Service` set to `apicast-staging` or `apicast-production`. There should be at lease one trace for stage environment.