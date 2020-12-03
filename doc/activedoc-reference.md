# ActiveDoc CRD Reference

## Table of Contents

* [ActiveDoc CRD Reference](#activedoc-crd-reference)
   * [Table of Contents](#table-of-contents)
   * [ActiveDoc](#activedoc)
      * [ActiveDocSpec](#activedocspec)
         * [ActiveDocOpenAPIRefSpec](#activedocopenapirefspec)
         * [OpenAPI Secret Reference](#openapi-secret-reference)
         * [Provider Account Reference](#provider-account-reference)
      * [ActiveDocStatus](#activedocstatus)
         * [ConditionSpec](#conditionspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## ActiveDoc

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [ActiveDocSpec](#activedocspec) | The specfication for the custom resource |
| Status | `status` | [ActiveDocStatus](#activedocstatus) | The status for the custom resource |

### ActiveDocSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Name | **Yes** |
| ActiveDocOpenAPIRefSpec | `activeDocOpenAPIRef` | object | Reference to the OpenAPI Specification. See [ActiveDocOpenAPIRefSpec](#activedocopenapirefspec) | **Yes** |
| System Name | `systemName` | string | Name | No |
| Description | `description` | string | ActiveDoc description message | No |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |
| Product Reference | `productSystemName` | string | 3scale product's `system name`. The activedoc will be linked to this product | No |
| Published | `published` | bool | Switch to publish the activedoc. By default it will be `hidden` | No |
| SkipSwaggerValidations | `skipSwaggerValidations` | bool | Switch to skip OpenAPI validation. By default, the validation is enabled | No |

#### ActiveDocOpenAPIRefSpec

Reference to the OpenAPI Specification

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| SecretRef | `secretRef` | [v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectreference-v1-core) to [OpenAPI secret reference](#openapi-secret-reference) | The secret that contains the OpenAPI Document | No |
| URL | `url` | string | Remote URL from where to fetch the OpenAPI Document | No |

**NOTE**: Supported OpenAPI version is the [OpenAPI 3.0.2](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md) specification.

**NOTE**: Accepted formats are `json` and `yaml`

#### OpenAPI Secret Reference

The secret that contains the OpenAPI Document referenced by a [v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectreference-v1-core) type object.

The secret must have only **one field** with the value set to the openapi document content. The field name will not be read.

| **Field** | **Description** | **Required** |
| --- | --- | --- |
| *ANY* | OpenAPI Document | Yes |

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: my-openapi
type: Opaque
stringData:
  myopenapi.yaml: |
    ---
    openapi: "3.0.0"
    info:
    title: "some title"
    version: "1.0.0"
```

#### Provider Account Reference

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

### ActiveDocStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| ID | `activeDocId` | string | Internal ID |
| ProviderAccountHost | `providerAccountHost` | string | 3scale account's provider URL |
| ProductResourceName | `productResourceName` | [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | Reference to the linked 3scale product |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

For example:

```
status:
  activeDocId: 162943
  conditions:
  - lastTransitionTime: "2020-12-03T15:39:17Z"
    status: "False"
    type: Failed
  - lastTransitionTime: "2020-12-03T15:39:17Z"
    status: "False"
    type: Invalid
  - lastTransitionTime: "2020-12-03T15:39:17Z"
    status: "False"
    type: Orphan
  - lastTransitionTime: "2020-12-03T15:39:17Z"
    status: "True"
    type: Ready
  observedGeneration: 2
  providerAccountHost: https://3scale.example.com
```

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Invalid: Indicates that the combination of configuration in the ActiveDocSpec is not supported. This is not a transient error, but indicates a state that must be fixed before progress can be made;
  * Orphan: the ActiveDoc spec contains reference(s) to non existing resources;
  * Ready: Indicates the ActiveDoc resource has been successfully reconciled;
  * Failed: Indicates that an error occurred during reconcilliation;

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
