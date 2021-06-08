# 3scale Monitoring Resources

The 3scale monitoring resources are (optionally) installed when 3scale is installed on Openshift using the 3scale Operator.

## TOC

* [Enabling 3scale monitoring](#enabling-3scale-monitoring)
* [Monitored components](#monitored-components)
* [3scale Prometheus Rules](/doc/prometheusrules)
* [Monitoring stack](#monitoring-stack)
   * [Prometheus](#prometheus)
   * [Grafana](#grafana)

## Enabling 3scale monitoring

3scale monitoring is disabled by default. It can be enabled by setting monitoring to `true` in the [APIManager CR](apimanager-reference.md).

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager1
spec:
  wildcardDomain: example.com
  monitoring:
    enabled: true
```

NOTE: All monitoring resources will be created by the operator using *Create only* reconciliation policy. That means that PrometheusRules and GrafanaDashboards objects can be updated preventing the operator to revert the changes. This policy allows us to tune, for instance, the alert thresholds, to your needs.

Optionally, *PrometheusRules* deployment can be disabled. By default, *PrometheusRules* will be deployed.

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager1
spec:
  wildcardDomain: example.com
  monitoring:
    enabled: true
    enablePrometheusRules: false
```

Check available [3scale Prometheus Rules](/doc/prometheusrules).

## Monitored components

* Kubernetes resources at pod and namespace level where 3scale is installed
* Apicast Staging
* Apicast Production
* Backend worker
* Backend listener
* System
* Zync
* Zync-que

See:
* [APIcast metrics](https://github.com/3scale/APIcast/blob/master/doc/prometheus-metrics.md)
* [Backend metics](https://github.com/3scale/apisonator/blob/master/docs/prometheus_metrics.md)


## Monitoring stack

3scale monitoring is leveraged by [prometheus](https://prometheus.io/) and [grafana](https://grafana.com/) monitoring solutions. They need to be up and running in the cluster and configured to watch for monitoring resources.

**Monitoring stack setup and running is beyond 3scale operator responsibilities, cluster administrator must provide it.**
There are many ways to provide the required monitoring stack. Few examples are given:

* Openshift new proposal [User Workload Monitoring](https://github.com/openshift/enhancements/blob/master/enhancements/monitoring/user-workload-monitoring.md)

* [kube-prometheus](https://github.com/coreos/kube-prometheus): This repository collects Kubernetes manifests, Grafana dashboards, and Prometheus rules combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.

* For devtesting purposes only, *quickstart steps* to deploy minimum required monitoring stack in [monitoring-stack-deployment](monitoring-stack-deployment/README.md).
*Warning*: no *authentication* setup included. Protect the prometheus and grafana services when you want to expose them on the internet.
In our dev clusters we configure authentication to be managed by Openshift (so htpassword or github)

### Prometheus

3scale monitoring requires [Prometheus Operator](https://github.com/coreos/prometheus-operator) to be deployed in the cluster.
The prometheus operator is an operator that creates, configures, and manages Prometheus clusters atop Kubernetes. It provides `PodMonitor`, `ServiceMonitor` and `PrometheusRule` custom resources definitions required by 3scale monitoring.

Tested releases:
* Prometheus operator `v0.37.0`
* Prometheus image: `quay.io/prometheus/prometheus:v2.16.0`

Make sure prometheus services are configured to monitor 3scale monitoring resources.
When the prometheus services are deployed in the same namespace as 3scale, the simplest configuration is *catch-all* config in `Prometheus` custom resource spec:

```
podMonitorSelector: {}
ruleSelector: {}
```

Optionally, you can filter by labels. 3scale operator created `PodMonitors` and `PrometheusRules` will all be labeled, by default, with

```
app: 3scale-api-management
```

The `app=3scale-api-management` label value can be overriden in the [APIManager CR](apimanager-reference.md#APIManagerSpec).


`Prometheus` custom resource spec to filter podmonitors and rules:

```
podMonitorSelector:
  matchExpressions:
  - key: app
      operator: In
      values:
      - 3scale-api-management
ruleSelector:
  matchExpressions:
  - key: app
      operator: In
      values:
      - 3scale-api-management
```

*NOTE*: If the prometheus operator is installed in a different namespace than 3scale, then configure it accordingly to watch for resources outside the namespace.
For instance, label the namespace where 3scale in installed with `MYLABELKEY=MYLABELVALUE`, then setup a *namespace selector*:

```
  podMonitorNamespaceSelector:
    matchExpressions:
      - key: MYLABELKEY
        operator: In
        values:
          - MYLABELVALUE
```

Do not forget to provide required RBAC permissions. Check operator [doc](https://github.com/coreos/prometheus-operator/blob/v0.32.0/Documentation/api.md#prometheusspec) regarding this issue.

**Kubernetes metrics: kube-state-metrics**

3scale monitoring requires [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) for dashboards and alerts. There are several options available:

* Federate the prometheus instance with cluster default prometheus instance to gather required metrics.
* Configure your own scraping jobs to get metrics from kubelet, etcd...

### Grafana

3scale monitoring requires [Grafana Operator](https://github.com/integr8ly/grafana-operator) to be deployed in the cluster.
The grafana operator is an operator for creating and managing Grafana instances. It provides `GrafanaDashboard` custom resource definition required by 3scale monitoring.

Tested releases:
* Grafana operator `v3.6.0`
* Grafana image: `grafana/grafana:7.1.1`

Make sure grafana services are configured to monitor 3scale monitoring resources:
* `GrafanaDashboards`

3scale operator created `GrafanaDashboards` will all be labeled with

```
app: 3scale-api-management
monitoring-key: middleware
```

The `app=3scale-api-management` label value can be overriden in the [APIManager CR](apimanager-reference.md#APIManagerSpec).

This `dashboardLabelSelector` configuration in the `Grafana` custom resource spec should do that:

```
apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: grafana
spec:
  dashboardLabelSelector:
  - matchExpressions:
    - key: app
      operator: In
      values:
      - 3scale-api-management
```

*NOTE*: If the grafana operator is installed in a different namespace than 3scale, then configure accordingly it to watch for resources outside the namespace.
Use `--namespaces` or `--scan-all` operator flags to enable watching for dashboards in a list of namespaces.
Do not forget to provide required RBAC permissions.
Check operator [doc](https://github.com/integr8ly/grafana-operator/blob/v2.0.0/documentation/deploy_grafana.md#operator-flags) regarding this issue.

To set Grafana to gather its data from Prometheus as its storage data
source make sure you create a `GrafanaDataSource` with the type `prometheus`.
You can find more information in the grafana operator's
[GrafanaDataSource documentation](https://github.com/integr8ly/grafana-operator/blob/v2.0.0/documentation/datasources.md)
