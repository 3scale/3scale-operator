## Enabling 3scale monitoring stack

This guide goes through the process of enabling 3scale monitoring stack using Prometheus and Grafana Operators.
Due to Grafana v4 deprecation on OpenShift 4.16+ this guide is divided into two sections, one covering OpenShift versions up to (including) 4.15, second section is specific to OpenShift 4.14+ and 3scale Operator 2.15+.
It is highly recommended to use Grafana v5 when possible, the guide also goes through the migration steps required to migrate from Grafana v4 to Grafana v5.

#### Common requirement:

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

#### OpenShift versions up to and including 4.15

1. Install Grafana operator v4.10.1 from the Operator Hub.

2. Create additional-scrape-configs secret with 3scale scrape config

```bash
# Get the SECRET name that contains the THANOS_QUERIER_BEARER_TOKEN
SECRET=`oc get secret -n openshift-user-workload-monitoring | grep  prometheus-user-workload-token | head -n 1 | awk '{print $1 }'`
# Get the THANOS_QUERIER_BEARER_TOKEN using the SECRET name
oc get secret $SECRET -n openshift-user-workload-monitoring -o jsonpath="{.data.token}" | base64 -d
```
Update the file `3scale-scrape-configs.yaml` bearer_token field with the THANOS_QUERIER_BEARER_TOKEN.

Then create secret:

```bash
oc create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

3. Deploy prometheus

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

4. Create Prometheus route

```bash
oc expose service prometheus-operated
```

5. Deploy grafana datasource

```bash
oc apply -f datasource-v4.yaml
```

6. Deploy grafana

```bash
oc apply -f grafana-v4.yaml
```

#### OpenShift versions from 4.14 and 3scale Operator version 2.15 upwards

1. Install Grafana operator v5.12.0 from the Operator Hub.

2. Create new service account for Prometheus

```bash
oc create serviceaccount prometheus-monitoring
```

3. Create a ClusterRoleBinding to give Prometheus ServiceAccount the RBAC permissions required to scrape metrics

```bash
cat << EOF | oc create -f -
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: prometheus-monitoring
subjects:
  - kind: ServiceAccount
    name: prometheus-monitoring
    namespace: <3scale namespace>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-monitoring-view
EOF
```

4. Create a token for the prometheus-monitoring ServiceAccount

```bash
oc create token prometheus-monitoring
```

Bare in mind that this token will expire which means the Prometheus will lose access to required resource. You can add the `--duration 0s` to token creation to have token active forever. Refer to official OpenShift documentation on managing token lifetime.

5. Update the file `3scale-scrape-configs.yaml` bearer_token field with the token generated from the above.

6. Create additional-scrape-config secret:

```bash
oc create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

7. Deploy prometheus

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

8. Create Prometheus route

```bash
oc expose service prometheus-operated --hostname prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

9. Deploy grafana datasource

```bash
oc apply -f datasource-v5.yaml
```

10. Deploy grafana

```bash
oc apply -f grafana-v5.yaml
```

#### Migration from V4 to V5

1. Remove the Grafana Operator v4

At this stage, Grafana Operator will be removed but the v4 dashboards CRDs will remain intact. 

2. Install Grafana Operator v5

At this stage, if 3scale operator reconciliation loop was triggered, operator will recongnize that v5 is present and will create v5 dashboards, but won't remove the v4 dashboards.

3. Create Grafana v5 datasource (instructions above)
4. Create Grafana v5 grafana (instructions above)
5. Restart 3scale Operator

Restart might be required since by design, 3scale operator does not watch over the grafana dashboards, but reconciles them as part of regular reconciliation loops.

6. Delete the Grafana v4 custom resource
7. Expose Grafana v5 route

You might be required to re-create a Grafana route

8. At this point, you can remove the Grafana Dashboards of v4

At this point, all your Grafana v4 dashboards will also be removed
