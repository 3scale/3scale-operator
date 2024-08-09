## Enabling 3scale monitoring stack

This guide goes through the process of enabling 3scale monitoring stack using Prometheus and Grafana Operators.
Due to Grafana v4 deprecation on OpenShift 4.16+ this guide is divided into two sections, one covering OpenShift versions up to 4.15 (including), second section is specific to OpenShift 4.16+ and 3scale Operator 2.15+.
Grafana v5 is available on OpenShift 4.14+ and supported by 3scale Operator from version 2.15+, it is recommended to use Grafana v5 when possible.
The guide also goes through the migration steps required to migrate from Grafana v4 to Grafana v5.

#### OpenShift versions up to and including 4.15

1. Install Grafana operator v4 from the Operator Hub.

2. Install Prometheus operator v4 from the Operator Hub.

3. Enable monitoring in the APIManager CR
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata: ...
spec:
  monitoring:
    enabled: true # <------ here
  ...
```

4. Create new service account for Prometheus

```bash
oc create serviceaccount prometheus-monitoring
```

5. Create a ClusterRoleBinding to give Prometheus ServiceAccount the RBAC permissions required to scrape metrics. Update the ServiceAccount namespace before creating the ClusterRoleBinding.

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

6. Create a token for the prometheus-monitoring ServiceAccount

```bash
oc create token prometheus-monitoring
```
This token will expire, which means the Prometheus will lose access to required resource. You can add the `--duration X[s|m|h]` to specify how long will the token be valid for. Refer to official OpenShift documentation on managing token lifetime.

7. Update the file `3scale-scrape-configs.yaml` bearer_token field with the token generated from the above.

8. Create additional-scrape-config secret:

```bash
oc create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

9. Deploy prometheus

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

10. Create Prometheus route

```bash
oc expose service prometheus-operated --hostname prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

11. Deploy grafana datasource

```bash
oc apply -f datasource-v4.yaml
```

12. Deploy grafana

```bash
oc apply -f grafana-v4.yaml
```

#### OpenShift versions from 4.16+ and 3scale Operator version 2.15+

1. Install Grafana operator v5 from the Operator Hub.

2. Install Prometheus operator v4.10.0 from the Operator Hub.

3. Enable monitoring in the APIManager CR
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata: ...
spec:
  monitoring:
    enabled: true # <------ here
  ...
```

4. Create new service account for Prometheus

```bash
oc create serviceaccount prometheus-monitoring
```

5. Create a ClusterRoleBinding to give Prometheus ServiceAccount the RBAC permissions required to scrape metrics. Update the ServiceAccount namespace before creating the ClusterRoleBinding.

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

6. Create a token for the prometheus-monitoring ServiceAccount

```bash
oc create token prometheus-monitoring
```
This token will expire, which means the Prometheus will lose access to required resource. You can add the `--duration X[s|m|h]` to specify how long will the token be valid for. Refer to official OpenShift documentation on managing token lifetime.

7. Update the file `3scale-scrape-configs.yaml` bearer_token field with the token generated from the above.

8. Create additional-scrape-config secret:

```bash
oc create secret generic additional-scrape-configs --from-file=3scale-scrape-configs.yaml=./3scale-scrape-configs.yaml
```

9. Deploy prometheus

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

10. Create Prometheus route

```bash
oc expose service prometheus-operated --hostname prometheus.NAMESPACE_NAME.apps.CLUSTER_DOMAIN
```

11. Deploy grafana datasource

```bash
oc apply -f datasource-v5.yaml
```

12. Deploy grafana

```bash
oc apply -f grafana-v5.yaml
```

13. Expose grafana route

```bash
oc expose service example-grafana-service
```

#### Migration from v4 to v5

##### Grafana v4 removal

1. Remove the Grafana custom resource to trigger the deletion of Grafana v4 application along with the service and routes. Double check that the route is removed, if not, remove it manually.

2. Remove the Grafana v4 datasource custom resource

3. Remove the Grafana Operator v4

##### Grafana v5 installation

1. Install Grafana Operator v5

2. Create Grafana v5 datasource (instructions above in `OpenShift versions from 4.16+ and 3scale Operator version 2.15+` section)

3. Create Grafana v5 grafana (instructions above in `OpenShift versions from 4.16+ and 3scale Operator version 2.15+` section)

4. Expose Grafana v5 route (instructions above in `OpenShift versions from 4.16+ and 3scale Operator version 2.15+` section)

5. Create serviceAccount for Prometheus (instructions above in `OpenShift versions from 4.16+ and 3scale Operator version 2.15+` section) and update the token value in additional scrape configs.

6. Restart Prometheus instance

7. Migrate any custom dashboards to Grafana v5 CRDs.

##### Removal of v4 CRDs

1. Remove the GrafanaDashboards v4 CRDs

**ATTENTION** - Doing so will delete any existing v4 dashboards. This step is optional, if the CRDs are not removed, the 3scale operator will keep reconciling the v4 dashboards but they will have no effect on the monitoring stack.

2. Restart 3scale Operator

3. It might be required to restart Grafana Operator and Grafana deployment
