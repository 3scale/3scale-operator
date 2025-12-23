# Application CRD Reference

Table of Contents
=================

* [Application](#application)
    * [ApplicationSpec](#applicationspec)
        * [Provider Account Reference](#provider-account-reference)
    * [ApplicationStatus](#applicationstatus)
        * [ConditionSpec](#conditionspec)

Created by [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## Application

| **Field** | **json field**| **Type**                        | **Info** |
| --- | --- |---------------------------------| --- |
| Spec | `spec` | [ApplicationSpec](#ApplicationSpec) | The specfication for the custom resource |
| Status | `status` | [ApplicationStatus](#ApplicationStatus) | The status for the custom resource |

### ApplicationSpec

| **Field**           | **json field**        | **Type** | **Info**                                                                                                                                            | **Required** |
|---------------------|-----------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| Name                | `name`                | string   | Name                                                                                                                                                | Yes          |
| Description         | `description`         | string   | human-readable text of the application                                                                                                              | Yes          |
| AccountCR           | `accountCR`           | object   | name of account CR via [v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#localobjectreference-v1-core) | Yes          |
| ProductCR           | `productCR`           | object   | name of product CR via [v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#localobjectreference-v1-core) | Yes          |
| ApplicationPlanName | `applicationPlanName` | string   | name of application plan that the application will use                                                                                              | Yes          |
| Suspend             | `suspend`             | bool     | suspend application if true suspends application, if false resumes application                                                                      | No           |
| AuthSecretRef       | `authSecretRef`       | object   | [Auth secret reference](#Auth-secret-reference)                                                                                                     | No           |



#### Provider Account Reference

Application CR relies on the provider account reference for the [developer account](./developeruser-reference.md#provider-account-reference) 
and the [product](./product-reference.md#provider-account-reference) being the same. If not you will see an error in the status.

#### Auth secret reference

Auth secret reference by a [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) type object.

| **Field**         | **Description**                                                                                                    | **Required** |
| ----------------- | ------------------------------------------------------------------------------------------------------------------ | ------------ |
| *UserKey*         | UserKey field can be populated with a secret key or left an empty string if you wish to generate the secret        | Yes          |
| *ApplicationID*   | ApplicationID field can be populated with a secret key or left an empty string if you wish to generate the secret  | Yes          |
| *ApplicationKey*  | ApplicationKey field can be populated with a secret key or left an empty string if you wish to generate the secret | Yes          |
| *ClientID*        | ClientID field can be populated with a secret key or left an empty string if you wish to generate the secret       | Yes          |
| *ClientSecret*    | ClientSecret field can be populated with a secret key or left an empty string if you wish to generate the secret   | Yes          |


NOTE: ApplicationCR relies on ProductCR authentication mode to determine which fields to use from the secret

For example:
* With UserKey authentication mode
```
apiVersion: v1
kind: Secret
metadata:
  name: authsecret
type: Opaque
stringData:
  UserKey: "testApplicationUserKey"
```

* With AppID/AppKey authentication mode
```
apiVersion: v1
kind: Secret
metadata:
  name: authsecret
type: Opaque
stringData:
  ApplicationID: "testApplicationID"
  ApplicationKey: "testApplicationKey"
```

* With OIDC authentication mode
```
apiVersion: v1
kind: Secret
metadata:
  name: authsecret
type: Opaque
stringData:
  ClientID: "testApplicationClientID"
  ClientSecret: "testApplicationClientSecret"
```

### ApplicationStatus

| **Field**           | **json field**        | **Type**                              | **Info**                                                                   |
|---------------------|-----------------------|---------------------------------------|----------------------------------------------------------------------------|
| ID                  | `applicationID`       | int64                                 | Internal ID                                                                |
| Observed Generation | `observedGeneration`  | string                                | helper field to see if status info is up to date with latest resource spec |
| State               | `state`               | string                                | state message                                                              |
| ProviderAccountHost | `providerAccountHost` | string                                | 3scale control plane host                                                  |
| Conditions          | `conditions`          | array of [condition](#ConditionSpec)s | resource conditions                                                        |

#### ConditionSpec

The status object has an array of Conditions through which the Backend has or has not passed.
Each element of the Condition array has the following fields:

* The *ready* field is a string, with possible values **True**, **False**, and **Unknown**.

| **Field** | **json field** | **Type** | **Info**                    |
|-----------|----------------| --- |-----------------------------|
| Ready     | `ready`        | string | Ready: True, False, Unknown |

