# 3scale Monitoring Resources

The 3scale monitoring resources are (optionally) installed when 3scale is installed on Openshift using the 3scale Operator.

## TOC

* [Enabling 3scale monitoring](#enabling-3scale-monitoring)
* [Monitored components](#monitored-components)
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

NOTE: All monitoring resources will be created by the operator using *Create only* reconcilliation policy. That means that PrometheusRules and GrafanaDashboards objects can be updated preventing the operator to revert the changes back. This policy allows to tune, for instance, the alert thresholds, to your needs.

## Monitored components

* Kubernetes resources at pod and namespace level where 3scale is installed
* Apicast Staging
* Apicast Production
* 3scale Backend worker
* 3scale Backend listener
* System
* Zync
* Zync-que

## Monitoring stack

3scale monitoring is leveraged by [prometheus](https://prometheus.io/) and [grafana](https://grafana.com/) monitoring solutions. They need to be up and running in the cluster and configured to watch for monitoring resources.

Monitoring stack setup and running is beyond 3scale operator reponsabilities.
There are many ways to provide the required monitoring stack. Few examples are given:

* Openshift new proposal [User Workload Monitoring](https://github.com/openshift/enhancements/blob/master/enhancements/monitoring/user-workload-monitoring.md)

* [kube-prometheus](https://github.com/coreos/kube-prometheus): This repository collects Kubernetes manifests, Grafana dashboards, and Prometheus rules combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.

* For devtesting purposes only, *quickstart steps* to deploy minimum required monitoring stack in [monitoring-stack-deployment](monitoring-stack-deployment/README.md)

### Prometheus

3scale monitoring requires [Prometheus Operator](https://github.com/coreos/prometheus-operator) to be deployed in the cluster.
The prometheus operator is an operator that creates, configures, and manages Prometheus clusters atop Kubernetes. It provides `PodMonitor`, `ServiceMonitor` and `PrometheusRule` custom resources definitions required by 3scale monitoring.

Tested releases:
* Prometheus operator `v0.32.0`
* Prometheus image: `quay.io/openshift/origin-prometheus: 4.2`

Make sure prometheus services are configured to monitor 3scale monitoring resources.
The simplest configuration is catch-all config in `Prometheus` custom resource spec :

```
podMonitorSelector: {}
ruleSelector: {}
```

Optionally, you can filter by labels. 3scale operator created `PodMonitors` and `PrometheusRules` will all be labeled with

```
app: 3scale-api-management
```

`Prometheus` custom resource spec to filter podmonitors and rules:

```
podMonitorSelector:
  matchExpressions:
  - key: monitoring-key
      operator: In
      values:
      - middleware
ruleSelector:
  matchExpressions:
  - key: monitoring-key
    operator: In
    values:
    - middleware
```

Note: If the prometheus operator is installed in a different namespace than 3scale, then configure it accordingly to watch for resources outside the namespace. Check operator [doc](https://github.com/coreos/prometheus-operator/blob/v0.32.0/Documentation/api.md#prometheusspec) regarding this issue.

**Kubernetes metrics: kube-state-metrics**

3scale monitoring requires [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) for dashboards and alerts. There are several options available:

* Federate the prometheus instance with cluster default prometheus instance to gather required metrics.
* Configure your own scraping jobs to get metrics from kubelet, etcd...

### Grafana

3scale monitoring requires [Grafana Operator](https://github.com/integr8ly/grafana-operator) to be deployed in the cluster.
The grafana operator is an operator for creating and managing Grafana instances. It provides `GrafanaDashboard` custom resource definition required by 3scale monitoring.

Tested releases:
* Grafana operator `v3.4.0`
* Grafana image: `quay.io/openshift/origin-grafana:4.2`

Make sure grafana services are configured to monitor 3scale monitoring resources:
* `GrafanaDashboards`

3scale operator created `GrafanaDashboards` will all be labeled with

```
app: 3scale-api-management
monitoring-key: middleware
```

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

Note: If the grafana operator is installed in a different namespace than 3scale, then configure accordingly it to watch for resources outside the namespace. Check operator [doc](https://github.com/integr8ly/grafana-operator/blob/v2.0.0/documentation/deploy_grafana.md#operator-flags) regarding this issue.
