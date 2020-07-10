1. Install Prometheus opeator v0.32.0 from catalog

1. Install Grafana operator (v3.4.0). At the time of writing, not available from catalog, so manual installation from git repo.

1. Create additional-scrape-configs secret with 3scale scrape config

Get basic auth password `basicAuthPassword` from `ns/openshift-monitoring/secrets/grafana-datasources/prometheus.yaml` and update `3scale-scrape-configs.yaml` basic auth field.

Then create secret:

```
kubectl create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

1. Deploy prometheus

```
k apply -f prometheus.yaml
```

1. Create Prometheus route

```
oc expose service prometheus-operated --hostname prometheus.namespace_name.apps.DOMAIN
```

1. Deploy grafana

```
k apply -f grafana.yaml
```
