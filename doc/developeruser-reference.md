# DeveloperUser CRD Reference

## Table of Contents

* [DeveloperUser](#developeruser)
   * [DeveloperUserSpec](#developeruserspec)
      * [Password secret reference](#password-secret-reference)
      * [Provider Account Reference](#provider-account-reference)
   * [DeveloperUserStatus](#developeruserstatus)
      * [ConditionSpec](#conditionspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## DeveloperUser

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [DeveloperUserSpec](#developeruserspec) | The specfication for the custom resource |
| Status | `status` | [DeveloperUserStatus](#developeruserstatus) | The status for the custom resource |

### DeveloperUserSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Username | `username` | string | Username  | Yes |
| Email | `email` | string | Email | Yes |
| PasswordCredentialsRef | `passwordCredentialsRef` | [v1.SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#secretreference-v1-core) to [Password secret reference](#password-secret-reference)] | The secret that contains password | Yes |
| DeveloperAccountRef | `developerAccountRef` | [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | Local reference to the parent [DeveloperAccount CR](developeraccount-reference.md) | Yes |
| Suspended | `suspended` | bool | Defines the desired state. Defaults to "false" | No |
| Role | `role` | string | Defines the desired role. Valid values are `member` or `admin`. Defaults to `member` | No |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |

#### Password secret reference

The secret that contains the password referenced by a [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) type object.

| **Field** | **Description** | **Required** |
| --- | --- | --- |
| `password` | The field containing the password value | Yes |

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: my-user-password
type: Opaque
stringData:
  password: <password value>
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

### DeveloperUserStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| ID | `developerUserID` | int | Developer user internal ID |
| AccountID | `accoundID` | int | Parent developer account internal ID |
| DeveloperUserState | `developerUserState` | string | Developer user state |
| ProviderAccountHost | `providerAccountHost` | string | 3scale account's provider URL |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

For example:

```
status:
  accoundID: 2445583436906
  conditions:
  - lastTransitionTime: "2021-02-17T23:38:48Z"
    status: "False"
    type: Failed
  - lastTransitionTime: "2021-02-17T23:38:48Z"
    status: "False"
    type: Invalid
  - lastTransitionTime: "2021-02-17T23:39:09Z"
    status: "False"
    type: Orphan
  - lastTransitionTime: "2021-02-17T23:39:09Z"
    status: "True"
    type: Ready
  developerUserID: 2445583628982
  developerUserState: active
  observedGeneration: 1
  providerAccountHost: https://3scale-admin.example.com
```

#### ConditionSpec

The status object has an array of Conditions through which the DeveloperUser has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * *Invalid*: Invalid object. This is not a transient error, but it reports about invalid spec and should be changed. The operator will not retry.
  * *Failed*: Indicates that an error occurred during synchronization. The operator will retry.
  * *Ready*: Indicates the user has been successfully synchronized.
  * *Orphan*: The spec contains reference(s) to non existing resources.

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
