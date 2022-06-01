# ProxyConfigPromote CRD Reference

## Table of Contents

* [ProxyConfigPromote](#proxyconfigpromote)
    * [ProxyConfigPromoteSpec](#proxyconfigpromotespec)
        * [Provider Account Reference](#provider-account-reference)
    * [ProxyConfigPromoteStatus](#proxyconfigpromotestatus)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## ProxyConfigPromote

| **Field** | **json field**| **Type**                                          | **Info** |
| --- | --- |---------------------------------------------------| --- |
| Spec | `spec` | [ProxyConfigPromoteSpec](#ProxyConfigPromoteSpec) | The specfication for the custom resource |
| Status | `status` | [ProxyConfigPromoteStatus](#ProxyConfigPromoteStatus)        | The status for the custom resource |

### ProxyConfigPromoteSpec

| **Field**                  | **json field**       | **Type** | **Info**                                                                                                                                                   | **Required** |
|----------------------------|----------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| ProductCRName              | `productCRName`      | string   | Name of product Cr                                                                                                                                         | Yes          |
| Production                 | `production`         | bool     | If true promotes to production, if false promotes to staging                                                                                               | No           |
| DeleteCR                   | `deleteCR`           | bool     | If true deletes the resource after a succesfull promotion                                                                                                  | No           |                                                                            | No |


### ProxyConfigPromoteStatus

| **Field**           | **json field**       | **Type** | **Info**                                                                    |
|---------------------|----------------------| --- |-----------------------------------------------------------------------------|
| ProductId          | `productId`          | string | Internal ID of promted product                                              |
| LatestProductionVersion | `latestProductionVersion` | string | int with the current version in the production environment      |
| LatestStagingVersion | `latestStagingVersion` | string | int with the current version in the staging environment      |
| Conditions | `conditions` | array of [conditions](#ConditionSpec) | resource conditions |

For example:

```yaml
status:
  conditions:
    - lastTransitionTime: '2022-05-11T13:41:00Z'
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
  * Ready: Indicates the ActiveDoc resource has been successfully reconciled;

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
