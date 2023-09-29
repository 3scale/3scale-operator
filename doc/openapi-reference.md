# OpenAPI CRD Reference

## Table of Contents

* [OpenAPI](#openapi)
   * [OpenAPIAnnotations](#openapiannotations)
   * [OpenAPISpec](#openapispec)
      * [OpenAPIRef](#openapiref)
      * [Provider Account Reference](#provider-account-reference)
   * [OpenAPIStatus](#openapistatus)
      * [ConditionSpec](#conditionspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## OpenAPI

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [OpenAPISpec](#openapispec) | The specfication for the custom resource |
| Status | `status` | [OpenAPIStatus](#openapistatus) | The status for the custom resource |

### OpenAPIAnnotations

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| insecure_skip_verify | `insecure_skip_verify` | boolean | 3scale client skips certificate verification when reconciling a backend and product object created via OpenAPI - defaults to "false" | No |

### OpenAPISpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| OpenAPIRef | `openapiRef` | object | Reference to the OpenAPI Specification. See [OpenAPIRef](#openapiref) | Yes |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |
| ProductionPublicBaseURL | `productionPublicBaseURL` | string | Custom public production URL | No |
| StagingPublicBaseURL | `stagingPublicBaseURL` | string | Custom public staging URL | No |
| ProductSystemName | `productSystemName` | string | Custom 3scale product system name | No |
| PrivateBaseURL | `privateBaseURL` | string | Custom private base URL | No |
| PrefixMatching | `prefixMatching` | boolean | Use prefix matching instead of strict matching on mapping rules derived from openapi operations. Defaults to strict matching. | No |
| PrivateAPIHostHeader | `privateAPIHostHeader` | string | Custom host header sent by the API gateway to the private API | No |
| PrivateAPISecretToken | `privateAPISecretToken` | string | Custom secret token sent by the API gateway to the private API | No |
| OIDC | `oidc` | [*OIDCSpec](https://github.com/3scale/3scale-operator/blob/master/doc/product-reference.md#oidcspec) | OIDCSpec defines the desired configuration of OpenID Connect Authentication | No |

#### OpenAPIRef

Reference to the OpenAPI Specification

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| SecretRef | `secretRef` | [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) to [OpenAPI secret reference](#openapi-secret-reference) | The secret that contains the OpenAPI Document | No |
| URL | `url` | string | Remote URL from where to fetch the OpenAPI Document | No |

**NOTE**: Supported OpenAPI version is the [OpenAPI 3.0.2](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md) specification.

**NOTE**: Accepted formats are `json` and `yaml`

#### OpenAPI Secret Reference

The secret that contains the OpenAPI Document referenced by a [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) type object.

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

### OpenAPIStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| ProviderAccountHost | `providerAccountHost` | string | 3scale account's provider URL |
| ProductResourceName | `productResourceName` | [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | Reference to the managed 3scale product |
| BackendResourceNames | `backendResourceNames` | array of [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | List of references to the managed 3scale backend |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

For example:

```
status:
    backendResourceNames:
    - name: swaggerpetstore-4686df5a-da77-4099-904a-edc3d273aa53
    conditions:
    - lastTransitionTime: "2020-10-23T16:59:22Z"
      status: "False"
      type: Failed
    - lastTransitionTime: "2020-10-23T16:59:22Z"
      status: "False"
      type: Invalid
    - lastTransitionTime: "2020-10-23T16:59:22Z"
      status: "True"
      type: Ready
    observedGeneration: 1
    productResourceName:
      name: swaggerpetstore-4686df5a-da77-4099-904a-edc3d273aa53
    providerAccountHost: https://3scale-admin.example.net
```

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Invalid: Indicates that the combination of configuration in the OpenAPISpec is not supported. This is not a transient error, but indicates a state that must be fixed before progress can be made;
  * Ready: Indicates the openapi resource has been successfully reconciled;
  * Failed: Indicates that an error occurred during reconcilliation;

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
