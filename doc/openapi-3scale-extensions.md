# OpenAPI 3.0.x 3scale Extensions

This reference information shows examples of how to add 3scale extensions at the root, path, or operation level in an OpenAPI 3.0.x definition.

## Root-level 3scale extension

You can optionally add a 3scale extension at the root level of an OpenAPI definition to add custom metrics, policies, and application plans to a product. 

The `metrics` block adheres to the [MetricSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#metricspec), the `policies` block adheres to the [PolicyConfigSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#policyconfigspec), the `applicationPlans` block adheres to the [ApplicationPlanSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#applicationplanspec), the `pricingRules` block adheres to the [PricingRuleSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#PricingRuleSpec), and the `limits` block adheres to the [LimitSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#LimitSpec).

The following example shows an extension to configure custom policies and application plans:

```yaml
x-3scale-product:
  metrics:  ## map[string]github.com/3scale/3scale-operator/apis/capabilities/v1beta1.MetricSpec
    metric01:  ## This system name needs to be unique across all methods AND metrics
      friendlyName: "My Metric 01"
      unit: "hits"
      description: "This is a custom metric"
  policies:  ## []github.com/3scale/3scale-operator/apis/capabilities/v1beta1.PolicyConfig
    - name: "myPolicy1"
      version: "0.1"
      enabled: true
      configuration:
        http_proxy: http://example.com
        https_proxy: https://example.com
    - name: "myPolicy2"
      version: "2.0"
      enabled: true
      configurationRef:
        name: "my-config-policy-secret"
        namespace: "my-3scale-namespace"
  applicationPlans:  ## map[string]github.com/3scale/3scale-operator/apis/capabilities/v1beta1.ApplicationPlanSpec
    plan01:
      name: "My Plan 01"
      appsRequireApproval: false
      trialPeriod: 3
      setupFee: "3.00"
      costMonth: "2.00"
      pricingRules:  ## []github.com/3scale/3scale-operator/apis/capabilities/v1beta1.PricingRuleSpec
        - from: 1
          to: 100
          pricePerUnit: "0.05"
          metricMethodRef:
            systemName: "metric01"
      limits:  ## []github.com/3scale/3scale-operator/apis/capabilities/v1beta1.LimitSpec
        - period: "week"
          value: 100
          metricMethodRef:
            systemName: "hits"
            backend: "backendA"
```

## Operation-level 3scale extension

You can optionally add a 3scale extension at the operation level of an OpenAPI definition to specify additional fields for a mapping rule.
The `mappingRule` fields listed in the example below (`metricMethodRef`, `increment`, and `last`) adhere to the [MappingRuleSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#mappingrulespec).

The following example shows an extension added for a `get` operation for a `petstore` app:

```yaml
paths:
  /pets:
    get:
      operationId: listPets
      x-3scale-operation:  ## Operation-level 3scale extension
        mappingRule:
          metricMethodRef: "metric01"  ## Optional. If let unset, metricMethodRef will be set to the operationId by default.
          increment: 2
          last: true
```