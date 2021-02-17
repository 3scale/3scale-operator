# DeveloperAccount CRD Reference

## Table of Contents

* [DeveloperAccount CRD Reference](#developeraccount-crd-reference)
   * [Table of Contents](#table-of-contents)
   * [DeveloperAccount](#developeraccount)
      * [DeveloperAccountSpec](#developeraccountspec)
         * [Provider Account Reference](#provider-account-reference)
      * [DeveloperAccountStatus](#developeraccountstatus)
         * [ConditionSpec](#conditionspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## DeveloperAccount

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [DeveloperAccountSpec](#developeraccountspec) | The specfication for the custom resource |
| Status | `status` | [DeveloperAccountStatus](#developeraccountstatus) | The status for the custom resource |

### DeveloperAccountSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| OrgName | `orgName` | string | Group/Org  | Yes |
| MonthlyBillingEnabled | `monthlyBillingEnabled` | bool | The billing status. Defaults to `true` | No |
| MonthlyChargingEnabled | `monthlyChargingEnabled` | bool | Defaults to `true` | No |
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

### DeveloperAccountStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| ID | `accountID` | int | Developer account internal ID |
| AccountState | `accountState` | string | Developer account state |
| CreditCardStored | `creditCardStored` | bool | Info about credit card |
| ProviderAccountHost | `providerAccountHost` | string | 3scale account's provider URL |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

For example:

```
status:
  accountID: 2445583436906
  accountState: approved
  conditions:
  - lastTransitionTime: "2021-02-17T23:39:00Z"
    status: "False"
    type: Failed
  - lastTransitionTime: "2021-02-17T23:39:00Z"
    status: "False"
    type: Invalid
  - lastTransitionTime: "2021-02-17T23:39:00Z"
    status: "True"
    type: Ready
  - lastTransitionTime: "2021-02-17T23:39:00Z"
    status: "False"
    type: Waiting
  creditCardStored: false
  observedGeneration: 1
  providerAccountHost: https://3scale-admin.example.com
```

#### ConditionSpec

The status object has an array of Conditions through which the DeveloperAccount has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * *Invalid*: Invalid object. This is not a transient error, but it reports about invalid spec and should be changed. The operator will not retry.
  * *Failed*: Indicates that an error occurred during synchronization. The operator will retry.
  * *Ready*: Indicates the account has been successfully synchronized.
  * *Waiting*: Indicates the account is waiting for some event to happen. The operator will retry.

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
