# Application Capabilities


Featured capabilities:

* Allow interaction with the underlying 3scale API Management solution.
* Manage the 3scale application declaratively using openshift (custom) resources.

The following diagram shows 3scale entities and relations that will be eligible for management using openshift (custom) resources in a declarative way.

![3scale Object types](3scale-diagram.png)

The following diagram shows available custom resource definitions and their relations provided by the 3scale operator.

![3scale Custom Resources](capabilities-diagram.png)

## Table of contents

<!--ts-->
* [Application Capabilities](#application-capabilities)
   * [Table of contents](#table-of-contents)
   * [CRD Index](#crd-index)
   * [Quickstart Guide](#quickstart-guide)
   * [Backend custom resource](#backend-custom-resource)
      * [Backend metrics](#backend-metrics)
      * [Backend methods](#backend-methods)
      * [Backend mapping rules](#backend-mapping-rules)
      * [Backend custom resource status field](#backend-custom-resource-status-field)
      * [Link your 3scale backend to your 3scale tenant or provider account](#link-your-3scale-backend-to-your-3scale-tenant-or-provider-account)
   * [Product custom resource](#product-custom-resource)
      * [Product Deployment: Apicast Hosted](#product-deployment-apicast-hosted)
      * [Product Deployment: Apicast Self Managed](#product-deployment-apicast-self-managed)
      * [Product authentication types](#product-authentication-types)
         * [User Key](#user-key)
         * [AppID and AppKey pair](#appid-and-appkey-pair)
         * [OIDC](#oidc)
      * [Product metrics](#product-metrics)
      * [Product methods](#product-methods)
      * [Product mapping rules](#product-mapping-rules)
      * [Product application plans](#product-application-plans)
      * [Product application plan limits](#product-application-plan-limits)
      * [Product application plan pricing rules](#product-application-plan-pricing-rules)
      * [Product backend usages](#product-backend-usages)
      * [Product policy chain](#product-policy-chain)
      * [Product custom gateway response on errors](#product-custom-gateway-response-on-errors)
      * [Product custom resource status field](#product-custom-resource-status-field)
      * [Link your 3scale product to your 3scale tenant or provider account](#link-your-3scale-product-to-your-3scale-tenant-or-provider-account)
   * [OpenAPI custom resource](#openapi-custom-resource)
   * [ActiveDoc custom resource](#activedoc-custom-resource)
      * [Features](#features)
      * [Reference your OpenAPI document using secret source](#reference-your-openapi-document-using-secret-source)
      * [Reference your OpenAPI document using URL source](#reference-your-openapi-document-using-url-source)
      * [ActiveDoc spec source linked with a 3scale product](#activedoc-spec-source-linked-with-a-3scale-product)
      * [Link your ActiveDoc spec to your 3scale tenant or provider account](#link-your-activedoc-spec-to-your-3scale-tenant-or-provider-account)
   * [CustomPolicyDefinition Custom Resource](#custompolicydefinition-custom-resource)
      * [Link your CustomPolicyDefinition spec to your 3scale tenant or provider account](#link-your-custompolicydefinition-spec-to-your-3scale-tenant-or-provider-account)
   * [Tenant custom resource](#tenant-custom-resource)
      * [Preparation before deploying the new tenant](#preparation-before-deploying-the-new-tenant)
      * [Deploy the new tenant custom resource](#deploy-the-new-tenant-custom-resource)
      * [Tenant deletion](#tenant-deletion)
   * [DeveloperAccount custom resource](#developeraccount-custom-resource)
      * [DeveloperAccount custom resource status field](#developeraccount-custom-resource-status-field)
      * [Link your DeveloperAccount to your 3scale tenant or provider account](#link-your-developeraccount-to-your-3scale-tenant-or-provider-account)
   * [DeveloperUser custom resource](#developeruser-custom-resource)
      * [Create developer user with member role](#create-developer-user-with-member-role)
      * [Create developer user with admin role](#create-developer-user-with-admin-role)
      * [DeveloperUser custom resource status field](#developeruser-custom-resource-status-field)
      * [Link your DeveloperUser to your 3scale tenant or provider account](#link-your-developeruser-to-your-3scale-tenant-or-provider-account)
   * [Application Custom Resource](#application-custom-resource)
      * [Application Custom Resource Status Fields](#application-custom-resource-status-fields)
      * [Application Misconfiguration Errors](#application-misconfiguration-errors)
   * [Limitations and unimplemented functionalities](#limitations-and-unimplemented-functionalities)
<!--te-->

## CRD Index

* [Application CRD reference](application-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_application.yaml)
* [Backend CRD reference](backend-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_backend.yaml)
* [Product CRD reference](product-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_product.yaml) [\[2\]](cr_samples/product/)
* [Tenant CRD reference](tenant-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1alpha1_tenant.yaml)
* [OpenAPI CRD reference](openapi-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_openapi_url.yaml) [\[2\]](cr_samples/openapi/)
* [DeveloperAccount CRD reference](developeraccount-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_developeraccount.yaml)
* [DeveloperUser CRD reference](developeruser-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_developeruser_admin.yaml) [\[2\]](cr_samples/developeruser/)
* [ActiveDoc CRD reference](tenant-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_activedoc_url.yaml) [\[2\]](cr_samples/activedoc/)
* [CustomPolicyDefinition CRD reference](custompolicydefinition-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_custompolicydefinition.yaml)
* [ProxyConfigPromote CRD reference](proxyConfigPromote-reference.md)
    * CR samples [\[1\]](../config/samples/capabilities_v1beta1_proxyconfigpromote.yaml)

## Quickstart Guide

To get up and running quickly, this quickstart guide will show how to deploy your first 3scale product and backend with the minimum required configuration.

Requirements

* Access to an OpenShift Container Platform 4.2 cluster.
* 3scale operator up and running. [Installing through the OLM is quick and easy](quickstart-guide.md).
* Access to 3scale admin portal. Local in the working openshift namespace or remote 3scale. All you need is *3scale Admin URL* and *access token*.

**Steps**

**A)** Create `threescale-provider-account` secret with 3scale admin portal credentials. For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.
```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

**B)** Setup 3scale backend with upstream api `https://api.example.com`.

Create yaml file with the following content:

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Backend
metadata:
  name: backend1
spec:
  name: "Operated Backend 1"
  systemName: "backend1"
  privateBaseURL: "https://api.example.com"
```

Check on the fields of **Backend** custom resource and possible values in the [Backend CRD Reference](backend-reference.md) documentation.

Create a custom resource:

```
oc create -f backend1.yaml
```

**C)** Setup 3scale product with all default settings using previously created backend

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  systemName: "operatedproduct1"
  backendUsages:
    backend1:
      path: /
```

Check on the fields of **Product** custom resource and possible values in the [Product CRD Reference](product-reference.md) documentation.

Create a custom resource:

```
oc create -f product1.yaml
```

Created custom resources will take a few seconds to setup your 3scale instance. You can check when resources are synchronized checking object's `status` field conditions.
Or directly using `oc wait` command:

```
oc wait --for=condition=Synced --timeout=-1s backend/backend1
oc wait --for=condition=Synced --timeout=-1s product/product1
```

## Backend custom resource

It is assumed the reader is familiarized with [3scale backends](https://access.redhat.com/documentation/en-us/red_hat_3scale_api_management/2.8/html/glossary/threescale_glossary#api-backend).

The minimum configuration required to deploy and manage one 3scale backend is the *Private Base URL* and a name.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Backend
metadata:
  name: backend-1
spec:
  name: "My Backend Name"
  privateBaseURL: "https://api.example.com"
```

### Backend metrics

Define desired backend metrics in your backend custom resource.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Backend
metadata:
  name: backend-1
spec:
  name: "My Backend Name"
  privateBaseURL: "https://api.example.com"
  metrics:
    metric01:
      friendlyName: Metric01
      unit: "1"
    metric02:
      friendlyName: Metric02
      unit: "1"
    hits:
      description: Number of API hits
      friendlyName: Hits
      unit: "hit"
```

Check on the fields of **Backend** custom resource and possible values in the [Backend CRD Reference](backend-reference.md) documentation.

* **NOTE 1**: `metrics` map key names will be used as `system_name`. In the example: `metric01`, `metric02` and `hits`.
* **NOTE 2**: `metrics` map key names must be unique among all metrics **AND** methods.
* **NOTE 3**: `unit` and `friendlyName` fields are required.
* **NOTE 4**: `hits` metric will be created by the operator for you if not present.

### Backend methods

Define desired backend methods in your backend custom resource.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Backend
metadata:
  name: backend-1
spec:
  name: "My Backend Name"
  privateBaseURL: "https://api.example.com"
  methods:
    method01:
      friendlyName: Method01
    method02:
      friendlyName: Method02
```

Check on the fields of **Backend** custom resource and possible values in the [Backend CRD Reference](backend-reference.md) documentation.

* **NOTE 1**: `methods` map key names will be used as `system_name`. In the example: `method01` and `method02`.
* **NOTE 2**: `methods` map key names must be unique among all metrics **AND** methods.
* **NOTE 3**: `friendlyName` field is required.

### Backend mapping rules

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Backend
metadata:
  name: backend-1
spec:
  name: "My Backend Name"
  privateBaseURL: "https://api.example.com"
  mappingRules:
    - httpMethod: GET
      pattern: "/pets"
      increment: 1
      metricMethodRef: hits
    - httpMethod: GET
      pattern: "/pets/id"
      increment: 1
      metricMethodRef: hits
  metrics:
    hits:
      description: Number of API hits
      friendlyName: Hits
      unit: "hit"
```

Check on the fields of **Backend** custom resource and possible values in the [Backend CRD Reference](backend-reference.md) documentation.

* **NOTE 1**: `httpMethod`, `pattern`, `increment` and `metricMethodRef` fields are required.
* **NOTE 2**: `metricMethodRef` holds a reference to the existing metric or method map key name `system_name`. In the example, `hits`.

### Backend custom resource status field

The status field shows resource information useful for the end user.
It is not regarded to be updated manually and it is being reconciled on every change of the resource.

Fields:

* **backendId**: 3scale bakend internal ID
* **conditions**: status.Conditions k8s common pattern. States:
  * *Failed*: Indicates that an error occurred during synchronization. The operator will retry.
  * *Synced*: Indicates the backend has been successfully synchronized.
  * *Invalid*: Invalid object. This is not a transient error, but it reports about invalid spec and should be changed. The operator will not retry.
* **observedGeneration**: helper field to see if status info is up to date with latest resource spec.
* **providerAccountHost**: 3scale provider account URL to which the backend is synchronized.

Example of *Synced* resource.

```yaml
status:
  backendId: 59978
  conditions:
  - lastTransitionTime: "2020-06-22T10:50:33Z"
    status: "False"
    type: Failed
  - lastTransitionTime: "2020-06-22T10:50:33Z"
    status: "False"
    type: Invalid
  - lastTransitionTime: "2020-06-22T10:50:33Z"
    status: "True"
    type: Synced
  observedGeneration: 2
  providerAccountHost: https://3scale-admin.example.com
```

### Link your 3scale backend to your 3scale tenant or provider account

When some 3scale resource is found by the 3scale operator,
*LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Backend
metadata:
  name: backend-1
spec:
  name: "My Backend Name"
  privateBaseURL: "https://api.example.com"
  providerAccountRef:
    name: mytenant
```

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.

## Product custom resource

It is assumed the reader is familiarized with [3scale products](https://access.redhat.com/documentation/en-us/red_hat_3scale_api_management/2.8/html/glossary/threescale_glossary#product).

The minimum configuration required to deploy and manage one 3scale product is the name.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
```

### Product Deployment: Apicast Hosted

Configure your product with *Apicast Hosted* deployment mode

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  deployment:
    apicastHosted: {}
```

### Product Deployment: Apicast Self Managed

Configure your product with *Apicast Self Managed* deployment mode

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  deployment:
    apicastSelfManaged:
      stagingPublicBaseURL: "https://staging.api.example.com"
      productionPublicBaseURL: "https://production.api.example.com"
```

### Product authentication types

#### User Key

The application is identified & authenticated via a single string.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  deployment:
    <any>:
      authentication:
        userkey:
          authUserKey: myKey
```

Check [Product CRD Reference](product-reference.md) documentation for all the details.

#### AppID and AppKey pair

The application is identified via the App_ID and authenticated via the App_Key.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  deployment:
    <any>:
      authentication:
        appKeyAppID:
          appID: myAppID
          appKey: myAppKey
```

* **NOTE 1**: `appID` is the name of the parameter that acts of behalf of app id.
* **NOTE 2**: `appKey` is the name of the parameter that acts of behalf of app key.

Check [Product CRD Reference](product-reference.md) documentation for all the details.

#### OIDC

Use OpenID Connect for any OAuth 2.0 flow.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  deployment:
    <any>:
      authentication:
        oidc:
          issuerType: "keycloak"
          issuerEndpoint: "https://myclientid:myclientsecret@mykeycloack.example.com/auth/realms/myrealm"
          authenticationFlow:
            standardFlowEnabled: false
            implicitFlowEnabled: true
            serviceAccountsEnabled: true
            directAccessGrantsEnabled: true
          jwtClaimWithClientID: "azp"
          jwtClaimWithClientIDType: "plain"
```

* **NOTE 1**: `issuerType` and `issuerEndpoint` fields are required.
* **NOTE 2**: `issuerType` Defines the type of the issuer with the following valid types:
  * `keycloak`: Red Hat Single Sign-On
  * `rest`: Rest API
* **NOTE 3**: `issuerEndpoint` defines the location of your OpenID Provider. The format of this endpoint is determined on your OpenID Provider setup. A common guidance would be `https://<CLIENT_ID>:<CLIENT_SECRET>@<HOST>:<PORT>/auth/realms/<REALM_NAME>`
* **NOTE 4**: The credentials (*CLIENT_ID* and *CLIENT_CREDENTIALS*) provided in `issuerEndpoint` should have sufficient permissions to manage other clients in the realm.

Check [Product CRD Reference](product-reference.md) documentation for all the details.

### Product metrics

Define desired product metrics using the *metrics* object.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  metrics:
    metric01:
      friendlyName: Metric01
      unit: "1"
    hits:
      description: Number of API hits
      friendlyName: Hits
      unit: "hit"
```

* **NOTE 1**: `metrics` map key names will be used as `system_name`. In the example: `metric01` and `hits`.
* **NOTE 2**: `hits` metric will be created by the operator for you if not present.
* **NOTE 3**: `metrics` map key names must be unique among all metrics **AND** methods.
* **NOTE 4**: `unit` and `friendlyName` fields are required.

### Product methods

Define desired product methods using the *methods* object.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  methods:
    method01:
      friendlyName: Method01
    method02:
      friendlyName: Method02
```

* **NOTE 1**: `methods` map key names will be used as `system_name`. In the example: `method01` and `method02`.
* **NOTE 2**: `methods` map key names must be unique among all metrics **AND** methods.
* **NOTE 3**: `friendlyName` field is required.

### Product mapping rules

Define desired product mapping rules declaratively using the `mappingRules` object.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  metrics:
    hits:
      description: Number of API hits
      friendlyName: Hits
      unit: "hit"
  methods:
    method01:
      friendlyName: Method01
  mappingRules:
    - httpMethod: GET
      pattern: "/pets"
      increment: 1
      metricMethodRef: hits
    - httpMethod: GET
      pattern: "/cars"
      increment: 1
      metricMethodRef: method01
```

* **NOTE 1**: `httpMethod`, `pattern`, `increment` and `metricMethodRef` fields are required.
* **NOTE 2**: `metricMethodRef` holds a reference to the existing metric or method map key name `system_name`. In the example, `hits`.

### Product application plans

Define desired product application plans declaratively using the `applicationPlans` object.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  applicationPlans:
    plan01:
      name: "My Plan 01"
      setupFee: "14.56"
    plan02:
      name: "My Plan 02"
      trialPeriod: 3
      costMonth: 3
```

* **NOTE 1**: `applicationPlans` map key names will be used as `system_name`. In the example: `plan01` and `plan02`.

### Product application plan limits

Define the desired product application plan limits declaratively using the `applicationPlans.limits` list.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  metrics:
    hits:
      description: Number of API hits
      friendlyName: Hits
      unit: "hit"
  applicationPlans:
    plan01:
      name: "My Plan 01"
      limits:
        - period: month
          value: 300
          metricMethodRef:
            systemName: hits
            backend: backendA
        - period: week
          value: 100
          metricMethodRef:
            systemName: hits
```

* **NOTE 1**: `period`, `value` and `metricMethodRef` fields are required.
* **NOTE 2**: `metricMethodRef` reference can be product or backend reference. Use `backend` optional field to reference metric's backend owner.

### Product application plan pricing rules

Define desired product application plan pricing rules declaratively using the `applicationPlans.pricingRules` list.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  metrics:
    hits:
      description: Number of API hits
      friendlyName: Hits
      unit: "hit"
  applicationPlans:
    plan01:
      name: "My Plan 01"
      pricingRules:
        - from: 1
          to: 100
          pricePerUnit: "15.45"
          metricMethodRef:
            systemName: hits
        - from: 1
          to: 300
          pricePerUnit: "15.45"
          metricMethodRef:
            systemName: hits
            backend: backendA
```

* **NOTE 1**: `from`, `to`, `pricePerUnit` and `metricMethodRef` fields are required.
* **NOTE 2**: `metricMethodRef` reference can be product or backend reference. Use `backend` optional field to reference metric's backend owner.
* **NOTE 3**: `from` and `to` will be validated. `from` < `to` for any rule and overlapping ranges for the same metric is not allowed.

### Product backend usages

Define desired product backend usages declaratively using the `backendUsages` object.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  backendUsages:
    backendA:
      path: /A
    backendB:
      path: /B
```

* **NOTE 1**: `backendUsages` map key names are references to `Backend system_name`. In the example: `backendA` and `backendB`.
* **NOTE 2**: `path` field is required.

### Product policy chain

Define desired product policy chain declaratively using the `policies` object.

The policy configuration can be defined in plain text using the `configuration` field, for example:

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  policies:
  - configuration:
      http_proxy: http://example.com
      https_proxy: https://example.com
    enabled: true
    name: camel
    version: builtin
  - configuration: {}
    enabled: true
    name: apicast
    version: builtin
```

* **NOTE 1**: `apicast` policy item will be added by the operator if not included.

Alternatively, the configuration for the policy can be referenced in a secret, for example:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: config-policy
type: Opaque
stringData:
  configuration: "{\"http_proxy\":\"http://secret.com\"}"
```
* **NOTE 1**: `configuration` field must be used to contain the policy configuration.

This secret can then be referenced using the `configurationRef` field in the policy.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  policies:
  - configurationRef:
       name: config-policy
    name: camel
    version: builtin
    enabled: true
```

Policy chain of a 3scale product can be exported using the 3scale Toolbox [export command](https://github.com/3scale/3scale_toolbox/blob/master/docs/export-import-policy-chain.md)

```
$ 3scale policies export <MY-3SCALE-PROVIDER-ACCOUNT> <MY-PRODUCT>
---
- name: apicast
  version: builtin
  configuration: {}
  enabled: true
- name: content_caching
  version: builtin
  configuration:
    rules:
    - header: X-Cache-Status
      condition:
        operations:
        - left_type: plain
          right_type: plain
          op: "=="
          right: GET
          left: ein
        - left_type: plain
          right_type: plain
          op: "!="
          right: RIGHTein
          left: elftein
        combine_op: and
      cache: true
  enabled: true
- name: camel
  version: builtin
  configuration:
    https_proxy: https://example.com
    http_proxy: http://example.com
  enabled: true
```

### Product custom gateway response on errors

Define desired product custom gateway reponse on errors declaratively using the `gatewayResponse` object.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  deployment:
    apicastHosted:
      authentication:
        userkey:
          gatewayResponse:
            errorStatusAuthFailed: 500
            errorHeadersAuthFailed: "text/plain; charset=mycharset"
            errorAuthFailed: "My custom reponse body"
            errorStatusAuthMissing: 500
            errorHeadersAuthMissing: "text/plain; charset=mycharset"
            errorAuthMissing: "My custom reponse body"
            errorStatusNoMatch: 501
            errorHeadersNoMatch: "text/plain; charset=mycharset"
            errorNoMatch: "My custom reponse body"
            errorStatusLimitsExceeded: 502
            errorHeadersLimitsExceeded: "text/plain; charset=mycharset"
            errorLimitsExceeded: "My custom reponse body"
```

* **NOTE 1**: The `gatewayResponse` optional field my be set in several different deployment options.
The example just shows it for the apicast hosted deployment option and the authentication mode called UserKey.

Check [Product CRD Reference](product-reference.md) documentation for all the details.


### Product custom resource status field

The status field shows resource information useful for the end user.
It is not regarded to be updated manually and it is being reconciled on every change of the resource.

Fields:

* **productId**: 3scale product internal ID
* **conditions**: status.Conditions k8s common pattern. States:
  * *Failed*: Indicates that an error occurred during synchronization. The operator will retry.
  * *Synced*: Indicates the product has been successfully synchronized.
  * *Invalid*: Invalid object. This is not a transient error, but it reports about invalid spec and should be changed. The operator will not retry.
  * *Orphan*: Spec references non existing resource. The operator will retry.
* **observedGeneration**: helper field to see if status info is up to date with latest resource spec.
* **state**: 3scale product internal state read from 3scale API.
* **providerAccountHost**: 3scale provider account URL to which the backend is synchronized.

Example of *Synced* resource.

```yaml
status:
  conditions:
  - lastTransitionTime: "2020-10-21T18:07:01Z"
    status: "False"
    type: Failed
  - lastTransitionTime: "2020-10-21T18:06:54Z"
    status: "False"
    type: Invalid
  - lastTransitionTime: "2020-10-21T18:07:01Z"
    status: "False"
    type: Orphan
  - lastTransitionTime: "2020-10-21T18:07:01Z"
    status: "True"
    type: Synced
  observedGeneration: 1
  productId: 2555417872138
  providerAccountHost: https://3scale-admin.example.com
  state: incomplete
```

### Link your 3scale product to your 3scale tenant or provider account

When some 3scale resource is found by the 3scale operator,
*LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  providerAccountRef:
    name: mytenant
```

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.

## OpenAPI custom resource
* [OpenAPI custom resource](openapi-user-guide.md)

## ActiveDoc custom resource

### Features

* [OpenAPI 3.X](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md) specification
* Accepted OpenAPI spec document formats are `json` and `yaml`.
* OpenAPI spec document can be read from:
  * Secret
  * URL. Supported schemes are `http` and `https`
* Optionally, link the activedoc with a 3scale product using the `productSystemName` field. The value must be the `system_name` of the 3scale product's CR. The referenced 3scale product must exist as CR, it cannot be some unmanaged 3scale product's system name.
* Publish or hide the activedoc using the `published` field. By default, it will be `hidden`.
* Skip OpenAPI 3.0 validations using the `skipSwaggerValidations` field. By default, the activedoc will be validated.

### Reference your OpenAPI document using secret source

Create a secret with the OpenAPI spec document. The name of the secret object will be referenced in the ActiveDoc CR.

The following example shows how to create a secret out of a file:

```
$ cat myopenapi.yaml
---
openapi: "3.0.0"
info:
title: "some title"
version: "1.0.0"

$ oc create secret generic myopenapi --from-file myopenapi.yaml
secret/myopenapi created
```

**NOTE** The field name inside the secret is not read by the operator. Only the content is read.

**NOTE** The secret is not monitored for updates. If the secret is updated after being created, the ActiveDoc custom resource has to be updated to force reconcilliation. It is safe to delete the `status` field to force the reconcilliation and the operator will re-create the `status` field again.

Then, create your ActiveDoc CR providing reference to the secret holding the OpenAPI document.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: ActiveDoc
metadata:
  name: activedoc-secret
spec:
  name: "Operated ActiveDoc From secret"
  activeDocOpenAPIRef:
    secretRef:
      name: myopenapi
```

[ActiveDoc CRD Reference](activedoc-reference.md) for more info about fields.

### Reference your OpenAPI document using URL source

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: ActiveDoc
metadata:
  name: activedoc-from-url
spec:
  name: "Operated ActiveDoc From URL"
  activeDocOpenAPIRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.json"
```

[ActiveDoc CRD Reference](activedoc-reference.md) for more info about fields.

### ActiveDoc spec source linked with a 3scale product

One Product custom resource can be linked to the ActiveDoc custom resource.
The ActiveDoc custom resource cannot be linked to an unmanaged 3scale product by the system name.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: ActiveDoc
metadata:
  name: activedoc-with-product-link
spec:
  name: "Operated ActiveDoc"
  openapiRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.yaml"
  productSystemName: myPetProduct
```

[ActiveDoc CRD Reference](activedoc-reference.md) for more info about fields.

### Link your ActiveDoc spec to your 3scale tenant or provider account

When some ActiveDoc custom resource is found by the 3scale operator,
*LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: ActiveDoc
metadata:
  name: activedoc1
spec:
  name: "Operated ActiveDoc"
  openapiRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.yaml"
  providerAccountRef:
    name: mytenant
```

[ActiveDoc CRD Reference](activedoc-reference.md) for more info about fields.

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.

## CustomPolicyDefinition Custom Resource

Example:

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: CustomPolicyDefinition
metadata:
  name: custompolicydefinition-sample
spec:
  name: "MyCustomPolicy"
  version: "0.0.1"
  schema:
    name: "MyCustomPolicy"
    version: "0.0.1"
    summary: "some summary"
    $schema: "http://json-schema.org/draft-07/schema#"
    configuration:
      type: "object"
      properties:
        someAttr:
            description: "Some attribute"
            type: "integer"
```

[CustomPolicyDefinition CRD Reference](custompolicydefinition-reference.md) for more info about fields.

### Link your CustomPolicyDefinition spec to your 3scale tenant or provider account

When some CustomPolicyDefinition custom resource is found by the 3scale operator,
*LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: CustomPolicyDefinition
metadata:
  name: custompolicydefinition-sample
spec:
  name: "MyCustomPolicy"
  version: "0.0.1"
  schema:
    name: "MyCustomPolicy"
    version: "0.0.1"
    summary: "some summary"
    $schema: "http://json-schema.org/draft-07/schema#"
    configuration:
      type: "object"
      properties:
        someAttr:
            description: "Some attribute"
            type: "integer"
  providerAccountRef:
    name: mytenant
```

[CustomPolicyDefinition CRD Reference](custompolicydefinition-reference.md) for more info about fields.

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.

## Tenant custom resource

Tenant is also known as Provider Account.

Creating the [*APIManager*](apimanager-reference.md) custom resource tells the operator to deploy 3scale.
Default 3scale installation includes a default tenant ready to be used. Optionally,
you may create other tenants creating [Tenant](tenant_reference.md) custom resource objects.

### Preparation before deploying the new tenant

To deploy a new tenant in your 3scale instance, first, you need some preparation steps:

* Create or local 3scale Master credentials secret: *MASTER_SECRET*
* Create a new secret to store the password for the admin account of the new tenant: *ADMIN_SECRET*
* Get the 3scale master account hostname: *MASTER_HOSTNAME*


A) *3scale Master credentials secret: MASTER_SECRET*

Tenant management can only be done using 3scale *master* account. You need *master* account credentials (preferably an access token).

* If the tenant resource is created in the same namespace as 3scale,
the secret with *master* account credentials has been created already and it is called **system-seed**.

* If the tenant resource is not created in the same namespace as 3scale,
you need to create a secret with the *master* account credentials.

```sh
oc create secret generic system-seed --from-literal=MASTER_ACCESS_TOKEN=<master access token>
```

Note: the name of the secret is optional. The secret name will be used in the tenant custom resource.

B) *Create a new secret to store the password for the admin account of the new tenant: ADMIN_SECRET*

```sh
oc create secret generic ecorp-admin-secret --from-literal=admin_password=<admin password value>
```

Note: the name of the secret is optional. The secret name will be used in the tenant custom resource.

C) *Get 3scale master account hostname: MASTER_HOSTNAME*

When you deploy 3scale using the operator, the master account has a fixed URL: `master.${wildcardDomain}`

* If you have access to the namespace where 3scale is installed,
the master account hostname can be easily obtained:

```
oc get routes --field-selector=spec.to.name==system-master -o jsonpath="{.items[].spec.host}"
```

### Deploy the new tenant custom resource

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Tenant
metadata:
  name: ecorp-tenant
spec:
  username: admin
  systemMasterUrl: https://<MASTER_HOSTNAME>
  email: admin@ecorp.com
  organizationName: ECorp
  masterCredentialsRef:
    name: <MASTER_SECRET>
  passwordCredentialsRef:
    name: <ADMIN_SECRET*>
  tenantSecretRef:
    name: tenant-secret
```

Check on the fields of Tenant Custom Resource and possible values in the [Tenant CRD Reference](tenant-reference.md) documentation.

Create the tenant resource:

```sh
oc create -f <yaml-name>
```

This should trigger the creation of a new tenant in your 3scale API Management solution.

The 3scale operator will create a new secret and store the new tenant's credentials in it. The new tenant *provider_key* and *admin domain url* will be stored in a secret.
The secret location can be specified using *tenantSecretRef* tenant spec key.

Example of the created secret content:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tenant-secret
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

Refer to [Tenant CRD Reference](tenant-reference.md) documentation for more information.

### Tenant deletion

If a tenant has been created via CR it can be marked for deletion in 3scale API Management solution by deleting the tenant CR.

## DeveloperAccount custom resource

The minimum configuration required to deploy and manage one 3scale developer account is:
* Provide, at least, the organization name in the `spec.OrgName` field.
* Create one [DeveloperUser CR](#developeruser-custom-resource) with the `admin` role. Without any admin developer user custom resource deployed, the account cannot be created.
Like any other tenant owned entities, the developer account needs to be linked to some 3scale tenant or provider account.

Custom resource example:

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: DeveloperAccount
metadata:
  name: developeraccount-simple-sample
spec:
  orgName: Ecorp
```

### DeveloperAccount custom resource status field

The status field shows resource information useful for the end user.
It is not regarded to be updated manually and it is being reconciled on every change of the resource.

Fields:

* **accountID**: developer account internal ID
* **accountState**: developer account state
* **creditCardStored**: info about credit card
* **conditions**: status.Conditions k8s common pattern. States:
  * *Invalid*: Invalid object. This is not a transient error, but it reports about invalid spec and should be changed. The operator will not retry.
  * *Failed*: Indicates that an error occurred during synchronization. The operator will retry.
  * *Ready*: Indicates the account has been successfully synchronized.
  * *Waiting*: Indicates the account is waiting for some event to happen. The operator will retry.
* **observedGeneration**: helper field to see if status info is up to date with latest resource spec.
* **providerAccountHost**: 3scale provider account URL to which the backend is synchronized.

Example of *Ready* resource.

```yaml
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

### Link your DeveloperAccount to your 3scale tenant or provider account

When some openapi custom resource is found by the 3scale operator,
*LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: DeveloperAccount
metadata:
  name: developeraccount-simple-sample
spec:
  orgName: Ecorp
  providerAccountRef:
    name: mytenant
```

[DeveloperAccont CRD reference](developeraccount-reference.md) for more info about fields.

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.

## DeveloperUser custom resource

Notes:

* 3scale developer users belong to some developer account. Therefore, the `DeveloperUser` custom resource requires a reference to one [DeveloperAccount CR](#developeraccount-custom-resource)
* `email` and `username` fields are unique among all developer users of the tenant.
* The password for the developer user will be provided in a referenced secret in the `passwordCredentialsRef` field.
* Developer users have the role of `admin` or `member`.

Before creating the developer user custom resource, create a new secret to store the password

```sh
$ cat secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: developeruserpassword
stringData:
  password: <password value>

$ oc apply -f secret.yaml
```

Alternatively

```sh
oc create secret generic developeruserpassword --from-literal=password=<password value>
```

### Create developer user with member role

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: DeveloperUser
metadata:
  name: developeruser-member-sample
spec:
  username: myusername1
  email: myusername1@example.com
  role: member
  passwordCredentialsRef:
    name: developeruserpassword
  developerAccountRef:
    name: developeraccount-simple-sample
```

### Create developer user with admin role

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: DeveloperUser
metadata:
  name: developeruser-member-sample
spec:
  username: myusername1
  email: myusername1@example.com
  role: admin
  passwordCredentialsRef:
    name: developeruserpassword
  developerAccountRef:
    name: developeraccount-simple-sample
```

### DeveloperUser custom resource status field

The status field shows resource information useful for the end user.
It is not regarded to be updated manually and it is being reconciled on every change of the resource.

Fields:

* **developerUserID**: developer user internal ID
* **developerUserState**: developer user state
* **accountID**: developer account internal ID to which developer user is linked
* **conditions**: status.Conditions k8s common pattern. States:
  * *Invalid*: Invalid object. This is not a transient error, but it reports about invalid spec and should be changed. The operator will not retry.
  * *Failed*: Indicates that an error occurred during synchronization. The operator will retry.
  * *Ready*: Indicates the account has been successfully synchronized.
  * *Orphan*: Spec references non existing resource. The operator will retry.
* **observedGeneration**: helper field to see if status info is up to date with latest resource spec.
* **providerAccountHost**: 3scale provider account URL to which the backend is synchronized.

Example of *Ready* resource.

```yaml
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

### Link your DeveloperUser to your 3scale tenant or provider account

When some openapi custom resource is found by the 3scale operator,
*LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: DeveloperUser
metadata:
  name: developeruser-member-sample
spec:
  username: myusername1
  email: myusername1@example.com
  passwordCredentialsRef:
    name: developeruserpassword
  developerAccountRef:
    name: developeraccount-simple-sample
  providerAccountRef:
    name: mytenant
```

[DeveloperUser CRD reference](developeruser-reference.md) for more info about fields.

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.

## Application Custom Resource

Notes:

* 3scale applications belong to some DeveloperAccount account.
* 3scale applications are linked directly to a product and applicationPlan

Consider we have the following product which is connected to a backend
```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1-cr
spec:
  applicationPlans:
    plan01:
      name: "My Plan 01"
      limits:
        - period: month
          value: 300
          metricMethodRef:
            systemName: hits
            backend: backend1
    plan02:
      name: "My Plan 02"
      limits:
        - period: month
          value: 300
          metricMethodRef:
            systemName: hits
            backend: backend1
  name: product1
  backendUsages:
    backend1:
      path: /
```

And the following developerAccount

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: DeveloperAccount
metadata:
  name: developeraccount01
spec:
  orgName: Ecorp
  providerAccountRef:
    name: mytenant
```
You will need the product CR name, developerAccount CR name and the application plan name from the product CR to create an application as follows, giving it a name and description

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Application
metadata:
  name: example
spec:
  accountCR:
    name: developeraccount01
  applicationPlanName: plan01
  productCR:
    name: product1-cr
  name: application-name
  description: description of application
```
You can suspend an existing application by updating the `spec.suspend` bool in the application CR

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Application
metadata:
  name: example
spec:
  accountCR:
    name: developeraccount01
  applicationPlanName: plan02
  productCR:
    name: product1-cr
  name: testApp12
  description: further testing12
  suspend: true
```
You will see the `status.state` update from live to suspended
```yaml
status:
  applicationID: 1
  conditions:
    - lastTransitionTime: '2022-11-01T14:22:14Z'
      status: 'True'
      type: Ready
  observedGeneration: 2
  providerAccountHost: 'https://3scale-admin.example.com'
  state: suspended
```
[Application CRD reference](application-reference.md) for more info about fields.

### Application Custom Resource Status Fields

Fields:

* **applicationID**: application internal ID
* **conditions**: status.Conditions k8s common pattern. States:
    * *Ready*: Indicates the account has been successfully synchronized.
* **observedGeneration**: helper field to see if status info is up to date with latest resource spec.
* **providerAccountHost**: 3scale provider account URL to which the backend is synchronized.
* **state**: either live or suspended depending on the `spec.suspend` bool

e.g. of a Successful status
```yaml
status:
  applicationID: 1
  conditions:
    - lastTransitionTime: '2022-11-01T14:22:14Z'
      status: 'True'
      type: Ready
  observedGeneration: 1
  providerAccountHost: 'https://3scale-admin.example.com'
  state: live
```

[Application CRD reference](application-reference.md) for more info about fields.

### Application Misconfiguration Errors

DeveloperAccount doesn't exist
The `status.message` field will inform you if the spec.accountCR.name developerAccount is incorrect or not found
```yaml
status:
  conditions:
    - lastTransitionTime: '2022-11-01T15:17:58Z'
      message: DeveloperAccount.capabilities.3scale.net "developeraccount02" not found
      status: 'False'
      type: Ready
  observedGeneration: 7
```
Product doesn't exist
The `status.message` field will inform you if the product CR name is incorrect or not found
```yaml
status:
  conditions:
    - lastTransitionTime: '2022-11-01T15:17:58Z'
      message: Product.capabilities.3scale.net "product2-cr" not found
      status: 'False'
      type: Ready
  observedGeneration: 8
```
ApplicationPlan doesn't exist
The `status.message` field will inform you if the applicationPlanName is not found in the product CR
```yaml
status:
  conditions:
    - lastTransitionTime: '2022-11-01T15:17:58Z'
      message: >-
        reconcile3scaleApplication application [testApp]: Plan [plan03] doesnt
        exist in product [product1]
      status: 'False'
      type: Ready
  observedGeneration: 9
```

## Limitations and unimplemented functionalities

* Single sign on (SSO) authentication for the admin portal
* Single sign on (SSO) authentication for the developers portal
