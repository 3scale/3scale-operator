1. Enable monitoring in the APIManager CR
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata: ...
spec:
  monitoring:
    enabled: true # <------ here
  ...
```
2. Install Prometheus operator v4.10.0 from the Operator Hub.

3. Install Grafana operator v4.10.1 from the Operator Hub.

4. Create additional-scrape-configs secret with 3scale scrape config

```bash
# Get the SECRET name that contains the THANOS_QUERIER_BEARER_TOKEN
$ SECRET=`oc get secret -n openshift-user-workload-monitoring | grep  prometheus-user-workload-token | head -n 1 | awk '{print $1 }'`
# Get the THANOS_QUERIER_BEARER_TOKEN using the SECRET name
$ oc get secret $SECRET -n openshift-user-workload-monitoring -o jsonpath="{.data.token}" | base64 -d

```
Update the file `3scale-scrape-configs.yaml` bearer_token field with the THANOS_QUERIER_BEARER_TOKEN.

Then create secret:

```bash
kubectl create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

5. Deploy prometheus

In `prometheus.yaml` file provided, fill the `spec.externalUrl` field with the external URL. The URL template should be:

```yaml
spec:
  ...
  externalUrl: https://prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

Then deploy prometheus server:

```bash
oc apply -f prometheus.yaml
```

6. Create Prometheus route

```bash
oc expose service prometheus-operated --hostname prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

7. Deploy grafana datasource

```bash
oc apply -f datasource.yaml
```

8. Deploy grafana

```bash
oc apply -f grafana.yaml
```
