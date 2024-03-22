# ApplicationAuth CRD Reference

## Table of Contents

* [ApplicationAuth](#ApplicationAuth)
    * [ApplicationAuthSpec](#ApplicationAuthspec)
        * [Auth secret reference](#Auth-secret-reference)
        * [Provider Account Reference](#provider-account-reference)
    * [ApplicationAuthStatus](#ApplicationAuthstatus)
        * [ConditionSpec](#conditionspec)


Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## ApplicationAuth

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [ApplicationAuthSpec](#ApplicationAuthSpec) | The specfication for the custom resource |
| Status | `status` | [ApplicationAuthStatus](#ApplicationAuthStatus) | The status for the custom resource |

### ApplicationAuthSpec

| **Field**                  | **json field**      | **Type** | **Info**                                                                     | **Required** |
| --- | --- | --- | --- | --- |
| ApplicationCRName          | `applicationCRName` | string | Name of application Custom Resource                                          | Yes          |
| GenerateSecret             | `generateSecret`    | bool | If true ApplicationKey and UserKey are generatedd                            | No           |
| AuthSecretRef              | `authSecretRef`     | object | [Auth secret reference](#Auth-secret-reference)                              | Yes          |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No           |

#### Auth secret reference

Auth secret reference by a [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) type object.

The secret must have `UserKey` and `ApplicationKey` fields.

| **Field**               | **Description**                                                                                                    | **Required** |
| --- | --- | --- |
| *UserKey*               | UserKey field can be populated with a secret key or left an empty string if you wish to generate the secret        | Yes |
| *ApplicationKey* | ApplicationKey field can be populated with a secret key or left an empty string if you wish to generate the secret | Yes |

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: authsecret
type: Opaque
stringData:
  UserKey: "testUserKey"
  ApplicationKey: "testApplicationKey"
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


### ApplicationAuthStatus

| **Field** | **json field** | **Type** | **Info** |
| --- | --- | --- | --- |
| Conditions | `conditions` | array of [conditions](#ConditionSpec) | resource conditions |

For example:

```yaml
status:
  conditions:
    - lastTransitionTime: '2022-05-11T13:41:00Z'
      status: 'True'
      type: Ready
      message: "Application authentication has been successfully pushed, any further interactions with this CR will not be applied"
```

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
    * Ready: Indicates the ApplicationAuth resource has been successfully reconciled;
    * Failed: Indicates the ApplicationAuth resource is in failed state;

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
