# Application Capabilities via Operator

Featured capabilities:

* Allow interacting with the underlying 3scale API Management solution.
* Manage the 3scale application declaratively using openshift (custom) resources.
* ...

Existing 3scale custom object types and how they are related to each other is shown in the following diagram:

![3scale Object types and ](capabilities-diagram.png)

## CRD reference

<table>
  <tr>
    <td>[Account Provider](tenant.md)</td>
    <td>[Product](tenant.md)</td>
    <td>[Backend](tenant.md)</td>
    <td>[Account](tenant.md)</td>
  </tr>
  <tr>
    <td>[ActiveDoc](tenant.md)</td>
    <td>[Backend Usage](tenant.md)</td>
    <td>[Application Plan](tenant.md)</td>
    <td>[Application](tenant.md)</td>
  </tr>
</table>

|  | [Product](tenant.md) | [Backend](tenant.md) | [MappingRule](tenant.md) |
| --- | --- | --- | --- |
| [Metric](tenant.md) | [Method](tenant.md) | [Application](tenant.md) | [ApplicationPlan](tenant.md) |


* [Tenant reference](tenant-reference.md)
* [Capabilities reference](api-crd-reference.md)

Deploy the Capabilities custom resources. The Capabilities custom resources
allow you to define 3scale [Porta](https://github.com/3scale/porta) API definitions and set
them into a Porta installation. This Porta installation does not necessarily
need to be the same than the one deployed from the deployment of
an APIManager resource. The available Capabilities custom resources are:
* API
* Binding
* Limit
* MappingRule
* Metric
* Plan
* Tenant

## Deploy Tenants custom resource

Deploying the *APIManager* custom resource (see section above) creates a default tenant.
Optionally, you may create other tenants deploying **Tenant custom resource** objects.

To deploy a new tenant in your 3scale instance, first, create secret to store admin password:

```sh
$ cat ecorp-admin-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: ecorp-admin-secret
type: Opaque
stringData:
  admin_password: <admin password value>


$ oc create -f ecorp-admin-secret.yaml
secret/ecorp-admin-secret created
```

then, create a new Tenant CR YAML file with the following content:

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Tenant
metadata:
  name: ecorp-tenant
spec:
  username: admin
  systemMasterUrl: https://master.<wildcardDomain>
  email: admin@ecorp.com
  organizationName: ECorp
  masterCredentialsRef:
    name: system-seed
  passwordCredentialsRef:
    name: ecorp-admin-secret
  tenantSecretRef:
    name: ecorp-tenant-secret
    namespace: operator-test
```

To look at more information on what the Tenant Custom Resource fields and
possible values are refer to
the [Tenant CRD Reference](tenant-reference.md) documentation.

```sh
export NAMESPACE="operator-test"
oc project ${NAMESPACE}
oc create -f <yaml-name>
```

This should trigger the creation of a new tenant in your 3scale API Management
solution in the "operator-test" project.

Tenant *provider_key* and *admin domain url* will be stored in a secret.
The secret location can be specified using *tenantSecretRef* tenant spec key.
Refer to [Tenant CRD Reference](tenant-reference.md) documentation for more information.

## Deploy the Capabilities related custom resources

Here, we will start to configure APIs, metrics, mappingrules... in our newly created tenant by only using Openshift Objects!

So let's create our first API:

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: API
metadata:
  creationTimestamp: 2019-01-25T13:28:41Z
  generation: 1
  labels:
    environment: testing
  name: api01
spec:
  planSelector:
    matchLabels:
      api: api01
  description: api01
  integrationMethod:
    apicastHosted:
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
        secretToken: Shared_secret_sent_from_proxy_to_API_backend_9603f637ca51ccfe
      mappingRulesSelector:
        matchLabels:
          api: api01
      privateBaseURL: https://echo-api.3scale.net:443
  metricSelector:
    matchLabels:
      api: api01
```

In all the Selectors (metric, plan, mappingrules...) we use a specific label "api: api01", you can change that and add as many labels and play with the selectors to cover really complex scenarios.

We should add a Plan:

```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Plan
metadata:
  labels:
    api: api01
  name: plan01
spec:
  approvalRequired: false
  default: true
  costs:
    costMonth: 0
    setupFee: 0
  limitSelector:
    matchLabels:
      api: api01
  trialPeriod: 0
```

A metric called metric01:

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
  incrementHits: false

```

A simple limit with a limit of 10 hits per day for the previous metric:

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

And a MappingRule to increment the metric01:

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

And now, let's "bind" all together with the binding object, we will use the credential created by the Tenant Controller:
```yaml
apiVersion: capabilities.3scale.net/v1alpha1
kind: Binding
metadata:
  name: mytestingbinding
spec:
  credentialsRef:
    name: ecorp-tenant-secret
  APISelector:
    matchLabels:
      environment: testing
```

As you can see, the binding object will reference the `ecorp-tenant-secret` and just create the API objects that are labeled as "environment: staging

Now, navigate to your new created 3scale Tenant, and check that everything has been created!

For more information, check the reference doc: [Capabilities CRD Reference](api-crd-reference.md)
