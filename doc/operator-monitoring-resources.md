# 3scale Monitoring Resources

The 3scale monitoring resources are (optionally) installed when 3scale is installed on Openshift using the 3scale Operator.

## Prerequirements

* [prometheus-operator](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus#quickstart) needs to be deployed in the cluster.

The prometheus operator is an operator that creates, configures, and manages Prometheus clusters atop Kubernetes. It provides `ServiceMonitor`, `PodMonitor` and `PrometheusRule` custom resources required by 3scale monitoring.

* [grafana-operator](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus#quickstart) needs to be deployed in the cluster.

The grafana operator is an operator for creating and managing Grafana instances. It provides `GrafanaDashboard` custom resources required by 3scale monitoring.

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

## Monitored components

* Kubernetes resources at namespace level where 3scale is installed
* Apicast Staging
* Apicast Production
* 3scale Backend worker
* 3scale Backend listener
* System sidekiq
* Zync
* Zync-que

## Exposing monitoring resources

`GrafanaDashboard` resources will all be labeled with

```
monitoring-key: middleware
```

Make sure the prometheus services and grafana services created by respective operators are configured to monitor resources with that label.

Depending on the prometheus and grafana service configuration, the namespace where 3scale is installed might require labels too. Check your monitoring provider configuration like grafana and prometheus servers.
