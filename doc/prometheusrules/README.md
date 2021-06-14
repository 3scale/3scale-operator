## 3scale PrometheusRules

### Index

* [Apicast](apicast.yaml)
* [Backend Listener](backend-listener.yaml)
* [Backend Worker](backend-worker.yaml)
* [System App](system-app.yaml)
* [System Sidekiq](system-sidekiq.yaml)
* [3scale Kube State Metrics](threescale-kube-state-metrics.yaml)
* [Zync](zync.yaml)
* [Zync QUE](zync-que.yaml)

### Namespaced prometheus rules

Published prometheus rules are namespaced with the generic `__NAMESPACE__` token.
The namespacing avoids conflicts when multiple 3scale instances are deployed in a cluster.

Before deploying the prometheus rules, make sure you modify the prometheus rules resources with 
your desired namespace. It can be easily done, for instance, for the apicast prometheus rules:

```bash
sed -i 's/__NAMESPACE__/mynamespace/g' apicast.yaml
```

Optionally, you can generate the prometheus rules with your own namespace name.
Read the section about [Building 3scale prometheus rules](/doc/development.md#building-3scale-prometheus-rules)
to follow the steps and know about required development tools needed to be installed.

### Tune the prometheus rules based on your infraestructure

If you decided to not have the 3scale operator deploy the prometheus rules for you,
it is clear that you want to tune them to your own needs. Before deploying the published prometheus rules, 
make sure you pay attention to: 

* Rule expression conditions and thresholds 
* Duration of the rule (the `for` fieldp)
* Severity of the rule

