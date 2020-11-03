1. Install Prometheus operator v0.37.0 from the Operator Hub.

1. Install Grafana operator v3.5.0 from the Operator Hub.

1. Create additional-scrape-configs secret with 3scale scrape config

Get basic auth password `basicAuthPassword` from `ns/openshift-monitoring/secrets/grafana-datasources/prometheus.yaml` and update `3scale-scrape-configs.yaml` basic auth field.

Then create secret:

```
kubectl create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

1. Deploy prometheus

In `prometheus.yaml` file provided, fill the `spec.externalUrl` field with the external URL. The URL template should be:

```
spec:
  ...
  externalUrl: https://prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

Then deploy prometheus server:

```
oc apply -f prometheus.yaml
```

1. Create Prometheus route

```
oc expose service prometheus-operated --hostname prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

1. Deploy grafana datasource

```
oc apply -f datasource.yaml
```

1. Deploy grafana

```
oc apply -f grafana.yaml
```
