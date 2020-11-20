# Product CRD Reference

## Table of Contents

* [Product](#product)
  * [ProductSpec](#productspec)
    * [ProductDeploymentSpec](#productdeploymentspec)
      * [ApicastHostedSpec](#apicasthostedspec)
      * [ApicastSelfManagedSpec](#apicastselfmanagedspec)
    * [AuthenticationSpec](#authenticationspec)
      * [UserKeyAuthenticationSpec](#userkeyauthenticationspec)
      * [AppKeyAppIDAuthenticationSpec](#appkeyappidauthenticationspec)
      * [SecuritySpec](#securityspec)
    * [MappingRuleSpec](#mappingrulespec)
    * [MetricSpec](#metricspec)
    * [MethodSpec](#methodspec)
    * [GatewayResponseSpec](#gatewayresponsespec)
    * [Provider Account Reference](#provider-account-reference)
    * [BackendUsageSpec](#backendusagespec)
    * [ApplicationPlanSpec](#applicationplanspec)
    * [PricingRuleSpec](#pricingrulespec)
    * [MetricMethodRefSpec](#metricmethodrefspec)
    * [LimitSpec](#limitspec)
  * [ProductStatus](#productstatus)
    * [ConditionSpec](#conditionspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## Product

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [ProductSpec](#ProductSpec) | The specfication for the custom resource |
| Status | `status` | [ProductStatus](#ProductStatus) | The status for the custom resource |

### ProductSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Name | Yes |
| System Name | `systemName` | string | Name | No |
| Description | `description` | string | Product description message | No |
| Deployment | `deployment` | object | See [ProductDeploymentSpec](#ProductDeploymentSpec) | No |
| Mapping Rules | `mappingRules` | array | See [MappingRules Spec](#MappingRuleSpec). Order in the array matters. Rules are processed as defined in the array from more prioritary to less prioritary | No |
| Metrics | `metrics` | object | Map with key as metric system name and value as [Metric Spec](#MetricSpec) | No |
| Methods | `methods` | object | Map with key as method system name and value as [Method Spec](#MethodSpec) | No |
| Backend Usages | `backendUsages` | object | Map with key as backend system name and value as [BackendUsageSpec](#BackendUsageSpec) | No |
| Application Plans | `applicationPlans` | object | Map with key as plan's system name and value as [ApplicationPlanSpec](#ApplicationPlanSpec) | No |
| Policy Chain | `policies` | array | Array of [PolicyConfigSpec](#PolicyConfigSpec) objects | No |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |

#### ProductDeploymentSpec

Specifies product deployment mode

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| ApicastHosted | `apicastHosted` | object | See [ApicastHostedSpec](#ApicastHostedSpec) | No |
| ApicastSelfManaged | `apicastSelfManaged` | object | See [ApicastSelfManagedSpec](#ApicastSelfManagedSpec) | No |

##### ApicastHostedSpec

Specifies apicast hosted deployment mode

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Authentication | `authentication` | object | See [AuthenticationSpec](#AuthenticationSpec) | No |

##### ApicastSelfManagedSpec

Specifies apicast self managed deployment mode

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Authentication | `authentication` | object | See [AuthenticationSpec](#AuthenticationSpec) | No |
| StagingPublicBaseURL | `stagingPublicBaseURL` | string | Staging Public Base URL | No |
| ProductionPublicBaseURL | `productionPublicBaseURL` | string | Production Public Base URL | No |

#### AuthenticationSpec

Specifies product authentication

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| UserKeyAuthentication | `userkey` | object | See [UserKeyAuthenticationSpec](#UserKeyAuthenticationSpec) | No |
| AppKeyAppIDAuthentication | `appKeyAppID` | object | See [AppKeyAppIDAuthenticationSpec](#AppKeyAppIDAuthenticationSpec) | No |
| OIDC | `oidc` | object | See [OIDCSpec](#OIDCSpec) | No |

##### UserKeyAuthenticationSpec

Specifies product user key authentication mode

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Key | `authUserKey` | string | The application is identified & authenticated via a single string | No |
| CredentialsLoc | `credentials` | string | Credentials location. Valid values: *headers*, *query*, *authorization* | No |
| Security | `security` | object | See [SecuritySpec](#SecuritySpec) | No |
| GatewayResponse | `gatewayResponse` | object | See [GatewayResponseSpec](#GatewayResponseSpec) | No |

##### AppKeyAppIDAuthenticationSpec

Specifies product appKey & appID authentication mode

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| AppID | `appID` | string | The application is identified via the *App_ID* | No |
| AppKey | `appKey` | string | The application is authenticated via the *App_Key* | No |
| CredentialsLoc | `credentials` | string | Credentials location. Valid values: *headers*, *query*, *authorization* | No |
| Security | `security` | object | See [SecuritySpec](#SecuritySpec) | No |
| GatewayResponse | `gatewayResponse` | object | See [GatewayResponseSpec](#GatewayResponseSpec) | No |

##### OIDCSpec

Specifies product OIDC authentication mode

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| IssuerEndpoint | `issuerEndpoint` | string | Location of your OpenID Provider. The format of this endpoint is determined on your OpenID Provider setup. A common guidance would be `https://<CLIENT_ID>:<CLIENT_SECRET>@<HOST>:<PORT>/auth/realms/<REALM_NAME>`.  | **Yes** |
| IssuerType | `issuerType` | string | The type of the OIDC issuer. Valid values: *keycloak*, *rest* | **Yes** |
| AuthenticationFlow | `authenticationFlow` | object | See [OIDCAuthenticationFlowSpec](#OIDCAuthenticationFlowSpec) | No |
| JwtClaimWithClientID | `jwtClaimWithClientID` | string | The JSON Web Token (JWT) Claim with ClientID that contains the clientID. Defaults to 'azp'. | No |
| JwtClaimWithClientIDType | `jwtClaimWithClientIDType` | string | Sets to process the ClientID Token Claim value as a string or as a liquid template. Valid values: *plain*, *liquid* | No |
| CredentialsLoc | `credentials` | string | Credentials location. Valid values: *headers*, *query*, *authorization* | No |
| Security | `security` | object | See [SecuritySpec](#SecuritySpec) | No |
| GatewayResponse | `gatewayResponse` | object | See [GatewayResponseSpec](#GatewayResponseSpec) | No |

##### OIDCAuthenticationFlowSpec

Specifies OAuth2.0 authorization grant type

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| StandardFlowEnabled | `standardFlowEnabled` | bool | OAuth 2 grant type: [*Authorization Code*](https://oauth.net/2/grant-types/authorization-code/) | **Yes** |
| ImplicitFlowEnabled | `implicitFlowEnabled` | bool | OAuth 2 grant type: [*Implicit*](https://oauth.net/2/grant-types/implicit/) | **Yes** |
| ServiceAccountsEnabled | `serviceAccountsEnabled` | bool | OAuth 2 grant type: [*Client Credentials*](https://oauth.net/2/grant-types/client-credentials/) | **Yes** |
| DirectAccessGrantsEnabled | `directAccessGrantsEnabled` | bool | OAuth 2 grant type: [*Password Grant*](https://oauth.net/2/grant-types/password/) | **Yes** |

##### SecuritySpec

Specifies product security

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| HostHeader | `hostHeader` | string | Lets you define a custom Host request header | No |
| SecretToken | `secretToken` | string | Enables you to block any direct developer requests to your API backend | No |

#### MappingRuleSpec

Specifies product mapping rules

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| HTTPMethod | `httpMethod` | string | Valid values: GET;HEAD;POST;PUT;DELETE;OPTIONS;TRACE;PATCH;CONNECT | Yes |
| Pattern | `pattern` | string | Mapping Rule pattern | Yes |
| Metric Method Reference | `metricMethodRef` | string | Existing method or metric **system name** | Yes |
| Increment | `increment` | int | Increase the metric by this delta | Yes |
| Last | `last` | \*bool | Last matched Mapping Rule to process | No |

#### MetricSpec

Specifies product metric

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `friendlyName` | string | Metric name | Yes |
| Unit | `unit` | string | Metric unit | Yes |
| Description | `description` | string | Metric description message | No |

#### MethodSpec

Specifies product method

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `friendlyName` | string | Method name | Yes |
| Description | `description` | string | Method description message | No |

##### GatewayResponseSpec

Specifies custom gateway response on errors

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| ErrorStatusAuthFailed | `errorStatusAuthFailed` | int | The response code when authentication fails | No |
| ErrorHeadersAuthFailed | `errorHeadersAuthFailed` | string | The Content-Type header when authentication fails | No |
| ErrorAuthFailed | `errorAuthFailed` | string | The response body when authentication fails | No |
| ErrorStatusAuthMissing | `errorStatusAuthMissing` | int | The response code when authentication is missing | No |
| ErrorHeadersAuthMissing | `errorHeadersAuthMissing` | string | The Content-Type header when authentication is missing | No |
| ErrorAuthMissing | `errorAuthMissing` | string | The response body when authentication is missing | No |
| ErrorStatusNoMatch | `errorStatusNoMatch` | int | The response code when no match error | No |
| ErrorHeadersNoMatch | `errorHeadersNoMatch` | string | The Content-Type header when no match error | No |
| ErrorNoMatch | `errorNoMatch` | string | The response body when no match error | No |
| ErrorStatusLimitsExceeded | `errorStatusLimitsExceeded` | int | The response code when usage limit exdeeded | No |
| ErrorHeadersLimitsExceeded | `errorHeadersLimitsExceeded` | string | The Content-Type header when usage limit exceeded | No |
| ErrorLimitsExceeded | `errorLimitsExceeded` | string | The response body when usage limit exceeded | No |


#### PolicyConfigSpec

Specifies product policy config object

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Policy name | Yes |
| Version | `version` | string | Policy version | Yes |
| Enabled | `enabled` | boolean | Policy enabling switch | Yes |
| Configuration | `configuration` | object | Policy configuration object | Yes. Minimum required is the empty object `{}` |

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

#### BackendUsageSpec

Specifies product backend usage

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Path | `path` | string | The path where this Backend API and its methods are available within the context of this Product | Yes |

#### ApplicationPlanSpec

LimitSpec defines the maximum value a metric can take on a contract before the user is no longer authorized to use resources.
Once a limit has been passed in a given period, reject messages will be issued if the service is accessed under this contract.

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Friendly name | No |
| AppsRequireApproval | `appsRequireApproval` | bool | Set whether or not applications can be created on demand or if approval is required from you before they are activated | No |
| TrialPeriod | `trialPeriod` | int | Trial Period (days) | No |
| SetupFee | `setupFee` | string | Setup fee (USD) | No |
| CostMonth | `costMonth` | string | Cost per Month (USD) | No |
| PricingRules | `pricingRules` | array | Array of [PricingRuleSpec](#PricingRuleSpec) objects | No |
| Limits | `limits` | array | Array of [LimitSpec](#LimitSpec) objects | No |

#### PricingRuleSpec

PricingRuleSpec defines the cost of each operation performed on an API.
Multiple pricing rules on the same metric divide up the ranges of when a pricing rule applies.

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| From | `from` | int | Range From | Yes |
| To | `to` | int | Range To | Yes |
| PricePerUnit | `pricePerUnit` | string | Price per unit (USD) | Yes |
| Metric Reference | `metricMethodRef` | object | See [MetricMethodRefSpec](#MetricMethodRefSpec) | No |

#### MetricMethodRefSpec

MetricMethodRefSpec defines method or metric reference. Metric or method can optionally belong to used backends.

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| SystemName | `systemName` | string | Identifies uniquely the metric or method | Yes |
| Backend | `backend` | string | Identifies uniquely backend owning the metric or method | No |

#### LimitSpec

LimitSpec defines the maximum value a metric can take on a contract before the user is no longer authorized to use resources.
Once a limit has been passed in a given period, reject messages will be issued if the service is accessed under this contract.

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Period | `period` | string | Limit period. Valid values: *eternity*, *year*, *month*, *week*, *day*, *hour*, *minute* | Yes |
| Value | `value` | int | Limit value | Yes |
| Metric Reference | `metricMethodRef` | object | See [MetricMethodRefSpec](#MetricMethodRefSpec) | No |

### ProductStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| ID | `productID` | string | Internal ID |
| State | `state` | string | Internal 3scale product state description |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Error Reason | `errorReason` | string | error code |
| Error Message | `errorMessage` | string | error message |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Synced: the product has been synchronized with 3scale;
  * Orphan: the product spec contains reference(s) to non existing resources;
  * Invalid: the product spec is semantically wrong and has to be changed;
  * Failed: An error occurred during synchronization.

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
