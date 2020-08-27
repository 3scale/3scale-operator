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
| Increment | `increment` | int | Increase the metric by this delta | Yes |
| Position | `position` | int | Mapping Rule position | No |
| Last | `last` | \*bool | Last matched Mapping Rule to process | No |

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

The status object has an array of Conditions through which the Backend has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Synced: the backend has been synchronized with 3scale;
  * Invalid: the backend spec is semantically wrong and has to be changed;
  * Failed: An error occurred during synchronization.

| **Field** | **json field**| **Type** | **Info** 
| --- | --- | --- | --- | 
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap | Yes |
