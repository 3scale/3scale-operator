# 3scale Operator


## Capabilities CRD

The 3scale Operator uses different CRDs that map into 3scale Porta objects:

* **Binding**: The Binding object relates a Secret (Containing the tenant credentials) with a set of APIs via a label selector
* **API**: Defines a 3scale API, defining the API backend URL, the desired Integration Method  (Apicast OnPrem/Hosted, Code Plugins...), and references to "mapping rules", Plans and Metrics using several label selectors. 
* **MappingRule**: A MappingRule is a combination of an HTTP Path, a Metric, an HTTP Verb, and an increment value. It's used by Apicast and other integrations, to increase a metric counter depending on the user usage.  
* **Metric**: Defines a Metric in 3scale.
* **Plan**: Plans map into Application Plans of 3scale Porta, define a set of usage limits. References Limits using a label Selector.
* **Limit**: A limit defines a max value for a given metric in a determined set of time. References a Metric object via an ObjectRef

CRD Diagram:
```


     ┌──────────────┐                                         
     │   Secret:    │            ┌───────────┐                      
     │              │            │           │                      
     │    Tenant    │◀─ObjectRef-│  Binding  │                      
     │ Credentials  │            │           │                      
     └──────────────┘            └───────────┘                      
                                       │                            
                                       │                            
                                    Selector                            
                                       │                            
                                       │                            
                                       ▼                            
       ┌──────────────┐           ┌─────────┐                       
       │┌─────────────┴┐          │         │                       
       ││              │ Selector │         │                       
       ││ Mapping Rule │◀─────────│   API   │───Selector────┐       
       └┤              │          │         │               │       
        └──────────────┘          │         │               │       
            ObjectRef             └─────────┘               │       
                │                      │                    │       
                ▼                      │                    │       
         ┌────────────┐            Selector                 │       
         │┌───────────┴┐               │                    │       
         └┤┌───────────┴◀──────────────┘                    │       
          └┤┌───────────┴┐                          ┌───────▼────┐  
           └┤   Metric   │                          │┌───────────┴┐ 
            └────────────┘                          └┤┌───────────┴┐
                   ▲                                 └┤    Plan    │
                   │           ┌───────────┐          └────────────┘
                   └─ObjectRef-│┌──────────┴┐                │      
                               └┤┌──────────┴┐           Selector     
                                └┤   Limit   │◀──────────────┘      
                                 └───────────┘    
                                 
```


## Binding CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [Binding](#BindingSpec) | The specification for the Binding custom resource |
| Status | `status` | [Status](#BindingStatus) | The status specification for the Binding custom resource |

### BindingSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Credentials Reference | `credentialsRef` | SecretRef | Reference to a Secret that contains the tenant credentials. See [Tenant Secret](#Tenant-Secret) for more details | Yes |
| API Selector | `APISelector` | LabelSelector | Selects the desired APIs to be created with the previous credentials, if empty, selects all the API object in the current namespace/project. | No |

### BindingStatus

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Desired State | `desiredState` | string | Contains the desired state of the system serialized in json | No |
| Current State | `currentState` | string |  Contains the current state of the system serialized in json  | No |
| Previous State | `previousState` | string |  Contains the previous state of the system serialized in json  | No |
| Last Successful Sync | `lastSync` | Timestamp |  Timestamp of the last successful sync | No |

### Tenant Secret

The credentials are typically provided by the [Tenant Controller](tenant-reference.md)
and stored in a secret, defined by the tenant CR.

But this Secret can also be created by the user following this schema: 

| **Field** | **Description** |
| --- | --- |
| *adminURL* | The URL of the admin portal for the target 3scale Tenant, with protocol and port. |
| *token* | Access Token with Read & Write permissions, or Provider Key |

#### Example Secret: 

```sh
apiVersion: v1
kind: Secret
metadata:
  name: ecorp-master-secret
type: Opaque
stringData:
  adminURL: https://my3scale-admin.3scale.net:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"

```


### Example Binding CR:

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Binding
metadata:
  name: myStagingCluster
spec:
  credentialsRef:
    name: staging-credentials
  APISelector:
    matchLabels:
      environment: staging
```

## API CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [APISpec](#APISpec) | The specification for the API custom resource |
| Status | `status` | TODO | The status for the API custom resource |

### APISpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Description | `description` | string | API Description | Yes |
| Integration Method | `integrationMethod` | Object | See [Integration Method](#IntegrationMethod) for more details | Yes |
| Plan Selector | `planSelector` | LabelSelector | Selects the desired Plan objects, if empty, selects all the Plan objects in the same namespace| No |
| Metric Selector | `metricSelector` | LabelSelector | Selects the desired Metric objects, if empty, selects all the Plan objects in the same namespace | No |

#### IntegrationMethod

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Apicast Hosted | `apicastHosted` | Object | Configures the API to use the included Apicast instance. See [ApicastHosted](#ApicastHosted) for more details |  Yes*  |
| Apicast OnPrem | `apicastOnPrem` | Object | Configures the API to use a user deployed Apicast instance. See [ApicastOnPrem](#ApicastOnPrem) for more details |  Yes*  |
| CodePlugin | `codePlugin` | Object | Configures the API to any of the code plugins libraries. See [CodePlugin](#CodePlugin) for more details |  Yes*  |

\* Only One Integration Method must be set.

##### ApicastHosted

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| API Test Get Request | `apiTestGetRequest` | string | The API path to use for the initial test request. Example: "/" |  Yes  |
| Authentication Settings | `authenticationSettings` | Object | See [Authentication Settings](#Authentication-Settings) for more details |  Yes  |
| MappingRules Selector | `mappingRulesSelector` | LabelSelector | Selects the desired MappingRule objects, if empty, selects all the MappingRule objects in the same namespace | No |
| Private Base URL | `privateBaseURL` | string | The URL of the private API to expose with 3scale. For example: "https://echo-api.3scale.net:443" |  Yes  |

##### ApicastOnPrem

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| API Test Get Request | `apiTestGetRequest` | string | The API path to use for the initial test request. Example: "/" |  Yes  |
| Authentication Settings | `authenticationSettings` | Object | See [Authentication Settings](#Authentication-Settings) for more details |  Yes  |
| MappingRules Selector | `mappingRulesSelector` | LabelSelector | Selects the desired MappingRule objects, if empty, selects all the MappingRule objects in the same namespace | No |
| Private Base URL | `privateBaseURL` | string | The URL of the API to expose with 3scale. For example: "https://echo-api.3scale.net:443" |  Yes  |
| Staging Public Base URL | `stagingPublicBaseURL` | string | The endpoint where the staging config will be exposed |  Yes  |
| Production Public Base URL | `productionPublicBaseURL` | string | The endpoint where the production config will be exposed. This is the URL that will be used by the final production users of the API  |  Yes  |

##### CodePlugin

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Authentication Settings | `authenticationSettings` | Object | See [Authentication Settings](#Authentication-Settings) for more details |  Yes  |

###### Authentication Settings

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Credentials | `credentials` | Object | See [Credentials](#Credentials) for more details |  Yes  |
| Errors | `errors` | Object | See [Errors](#Errors) for more details |  Yes  |
| Host Header | `hostHeader` | string | Override for the Host header when contacting the Private Base URL |  Yes  |
| Secret Token | `secretToken` | string | Secret token used to communicate with the Private Base URL |  Yes  |

###### Credentials

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| API Key | `apicastHosted` | Object | Configures the API to use the included Apicast instance. See [ApicastHosted](#ApicastHosted) for more details |  Yes*  |

###### Errors

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Authentication Failed | `authenticationFailed` | Object | See [Authentication Failed](#Authentication Failed) for more details |  Yes  |
| Authentication Missing | `authenticationMissing` | Object | See [Authentication Missing](#Authentication Missing) for more details |  Yes  |

###### Authentication Failed

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Content Type | `contentType` | string | The Content-Type to use, when returning the HTTP message to the client if authentication fails |  Yes  |
| Response Body | `responseBody` | string | The Response Body to use when returning the HTTP message to the client if authentication fails |  Yes  |
| Response Code | `responseCode` | int | The Response Code to use when returning the HTTP message to the client if authentication fails |  Yes  |

###### Authentication Missing

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Content Type | `contentType` | string | The Content-Type to use, when returning the HTTP message to the client if authentication parameters are missing |  Yes  |
| Response Body | `responseBody` | string | The Response Body to use when returning the HTTP message to the client if authentication parameters are missing |  Yes  |
| Response Code | `responseCode` | int | The Response Code to use when returning the HTTP message to the client if authentication parameters are missing |  Yes  |

### Example API CR: 

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: API
metadata:
  labels:
    environment: staging
  name: api01
spec:
  planSelector:
    matchLabels:
      api: api01
  description: api01
  integrationMethod:
    apicastOnPrem:
      apiTestGetRequest: /
      authenticationSettings:
        credentials:
          apiKey:
            authParameterName: user-key
            credentialsLocation: headers
        errors:
          authenticationFailed:
            contentType: text/plain; charset=us-ascii
            responseBody: Authentication failed
            responseCode: 403
          authenticationMissing:
            contentType: text/plain; charset=us-ascii
            responseBody: Authentication Missing
            responseCode: 403
        hostHeader: ""
        secretToken: MySecretTokenBetweenApicastAndMyBackend_1237120312
      mappingRulesSelector:
        matchLabels:
          api: api01
      privateBaseURL: https://echo-api.3scale.net:443
      productionPublicBaseURL: https://api.testing.com:443
      stagingPublicBaseURL: https://api.testing.com:443
  metricSelector:
    matchLabels:
      api: api01
```


## MappingRule CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [MappingRuleSpec](#MappingRuleSpec) | The specification for the MappingRule custom resource |
| Status | `status` | TODO | The status for the MappingRule custom resource |

### MappingRuleSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Increment | `increment` | int | Amount to increment the desired metric, for example: 1 | Yes |
| HTTP Method | `method` | string | The HTTP Method to match, for example: GET | Yes |
| Metric Reference | `metricRef` | ObjectRef | A kubernetes Object Reference to the desired Metric  | Yes |
| Path | `path` | string | The HTTP path to match to increment the desired Metric. | Yes |

#### Example MappingRule CR:

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: MappingRule
metadata:
  labels:
    api: api01
  name: metric01-get-path01
spec:
  increment: 1
  method: GET
  metricRef:
    name: metric01
  path: /path01
  ```

## Metric CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [MetricSpec](#MetricSpec) | The specification for the Metric custom resource |
| Status | `status` | TODO | The status for the Metric custom resource |

### MetricSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Description | `description` | string | Description for the metric | Yes |
| Unit | `unit` | string | The unit of the metric, for display purposes, for example: hits | Yes |

#### Example Metric CR:

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Metric
metadata:
  labels:
    api: api01
   name: metric01
spec:
  description: metric01
  unit: hit
```

## Plan CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [PlanSpec](#PlanSpec) | The specification for the Plan custom resource |
| Status | `status` | TODO | The status for the Plan custom resource |

### PlanSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Default | `default` | boolean | Sets the Plan as the default for developers to sign-up | Yes |
| Approval Required | `approvalRequired` | boolean | Defines if a final user requires approval from the admin to sign up for a plan | Yes |
| Costs | `costs` | Object | See [Costs](#Costs) | Yes |
| Limit Selector | `limitSelector` | LabelSelector | Selects the desired Limit objects, if empty, selects all the Limit objects in the same namespace | No |
| Trial Period | `trialPeriod` | int | See [Master Secret](tenant-reference.md#Master-Secret) for more details | Yes |

#### Costs

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Cost Month | `costMonth` | int | Monthly cost | Yes |
| Setup Fee | `setupFee` | int | Setup Fee | Yes |

#### Example Plan CR: 

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Plan
metadata:
  labels:
    api: api01
  name: plan01
spec:
  default: true
  approvalRequired: false
  costs:
    costMonth: 0
    setupFee: 0
  limitSelector:
    matchLabels:
      plan: plan01
  trialPeriod: 0
```
## Limit CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [LimitSpec](#LimitSpec) | The specification for the Limit custom resource |
| Status | `status` | TODO | The status for the Limit custom resource |

### LimitSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Description | `description` | string | Limit description | Yes |
| Max Value | `maxValue` | string | Max value for limit | Yes |
| Metric Reference | `metricRef` | ObjectRef | A kubernetes Object Reference to the desired Metric  | Yes |
| Period | `period` | string | Period of the limit: minute, day, month, week, eternity | Yes |

#### Example Limit CR: 

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Limit
metadata:
  labels:
    api: api01
  name: plan01-metric01-day-10
spec:
  description: Limit for metric01 in plan01
  maxValue: 10
  metricRef:
    name: metric01
  period: day
```

