# Backend CRD field Reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [BackendSpec](#BackendSpec) | The specfication for the custom resource |
| Status | `status` | [BackendStatus](#BackendStatus) | The status for the custom resource |

## BackendSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Name | Yes |
| Private Base URL | `privateBaseURL` | string | The private endpoint | Yes |
| System Name | `systemName` | string | Name | No |
| Description | `description` | string | Backend description message | No |
| Mapping Rules | `mappingRules` | object | See [MappingRules Spec](#MappingRuleSpec) | No |
| Metrics | `metrics` | object | Map with key as metric system name and value as [Metric Spec](#MetricSpec) | No |
| Methods | `methods` | object | Map with key as method system name and value as [Method Spec](#MethodSpec) | No |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |

### MappingRuleSpec

Specifies backend mapping rule

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| HTTPMethod | `httpMethod` | string | Valid values: GET;HEAD;POST;PUT;DELETE;OPTIONS;TRACE;PATCH;CONNECT | Yes |
| Pattern | `pattern` | string | Mapping Rule pattern | Yes |
| Metric Method Reference | `metricMethodRef` | string | Existing method or metric **system name** | Yes |
| Increment | `increment` | int | Increase the metric by this delta | No |
| Position | `position` | int | Mapping Rule position | No |

### MetricSpec

Specifies backend metric

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `friendlyName` | string | Metric name | Yes |
| Unit | `unit` | string | Metric unit | Yes |
| Description | `description` | string | Metric description message | No |

### MethodSpec

Specifies backend method

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `friendlyName` | string | Method name | Yes |
| Description | `description` | string | Method description message | No |

### Provider Account Reference

Provider account credentials secret referenced by a [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) type object. 

The secret must have `adminURL` and `token` fields with tenant credentials.
Tenant controller will fetch the secret and read the following fields:

| **Field** | **Description** | **Required** |
| --- | --- | --- |
| *token* | Provider account access token with *Account Management API* scope and *Read & Write* permission | Yes |
| *adminURL* | Provider account's domain URL | Yes |

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

## BackendStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- | --- |
| Backend ID | `backendId` | string | Internal ID |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Error Reason | `errorReason` | string | error code |
| Error Message | `errorMessage` | string | error message |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

### ConditionSpec

Specifies backend status condition

| **Field** | **json field**| **Type** | **Info** 
| --- | --- | --- | --- | 
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap | Yes |
