# ProxyConfigPromote CRD Reference

## Table of Contents

* [ProxyConfigPromote](#proxyconfigpromote)
    * [ProxyConfigPromoteSpec](#proxyconfigpromotespec)
        * [Provider Account Reference](#provider-account-reference)
    * [ProxyConfigPromoteStatus](#proxyconfigpromotestatus)
        * [ConditionSpec](#conditionspec)


Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## ProxyConfigPromote

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [ProxyConfigPromoteSpec](#ProxyConfigPromoteSpec) | The specfication for the custom resource |
| Status | `status` | [ProxyConfigPromoteStatus](#ProxyConfigPromoteStatus) | The status for the custom resource |

### ProxyConfigPromoteSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| ProductCRName | `productCRName` | string | Name of product Cr| Yes |
| Production | `production` | bool | If true promotes to production, if false promotes to staging | No |
| DeleteCR | `deleteCR` | bool | If true deletes the resource after a succesfull promotion | No |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |

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

### ProxyConfigPromoteStatus

| **Field** | **json field** | **Type** | **Info** |
| --- | --- | --- | --- |
| ProductId | `productId` | string | Internal ID of promted product |
| LatestProductionVersion | `latestProductionVersion` | string | int with the current version in the production environment |
| LatestStagingVersion | `latestStagingVersion` | string | int with the current version in the staging environment |
| Conditions | `conditions` | array of [conditions](#ConditionSpec) | resource conditions |

For example:

```yaml
status:
  conditions:
    - lastTransitionTime: '2022-05-11T13:41:00Z'
      message: >-
        3scale product has been successfully promoted, any further interactions
        with this CR (apart from deletion) won't be applied
      status: 'True'
      type: Ready
  latestProductionVersion: 7
  latestStagingVersion: 8
  productId: '3'
```

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Ready: Indicates the ProxyConfigPromote resource has been successfully reconciled;
  * Failed: Indicates the ProxyConfigPromote resource is in failed state;

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
