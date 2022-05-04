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
|----------------------------|----------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------| --- |
| ProductCRName              | `productCRName`      | string   | Name of product Cr                                                                                                                                         | Yes |
| Production                 | `production`         | bool     | If true promotes to production, if false promotes to staging                                                                                               | Yes |
| DeleteCR                   | `deleteCR`           | bool     | If true deletes the resource after a succesfull promotion                                                                                                  | No |
| Provider Account Reference | `providerAccountRef` | object   | [Provider account credentials secret reference](#provider-account-reference)                                                                               | No |



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

| **Field**           | **json field**       | **Type** | **Info**                                                                    |
|---------------------|----------------------| --- |-----------------------------------------------------------------------------|
| Product ID          | `productId`          | string | Internal ID of promted product                                              |
| Promote Environment | `promoteEnvironment` | string | Environment to which product should be promoted to                          |
| Status              | `status`             | string | Completed if promotion is successful , Failed if promotion is unsuccessful |