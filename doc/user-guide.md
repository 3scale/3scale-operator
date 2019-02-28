# User Guide

This guide shows an example of how to:

* Deploy the 3scale-operator
* Deploy an APIManager custom resource. An APIManager custom resource allows
   you to deploy a 3scale API Management solution
* Deploy the Capabilities custom resources. The Capabilities custom resources
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

## Prerequisites

* [operator-sdk] version v0.5.0.
* [dep][dep_tool] version v0.5.0+.
* [git][git_tool]
* [go] version 1.11+
* [kubernetes] version v1.11.0+
* [oc] version v3.11+
* Access to a Openshift v3.11.0+ cluster.
* A user with administrative privileges in the OpenShift cluster.


## Warning

The 3scale-operator is not in a stable status. Deploy the 3scale-operator and
its related custom resources in a non-productive environment. It is important
to deploy it in a namespace where no other elements exists.

**It is important to deploy all the elements in a separated namespace/project because deploying the resources in existing namespaces containing infrastructure could potentially alter/delete existing elements if those are together with the 3scale-operator and custom resources**

## Install the 3scale Operator

Download the 3scale-operator project into your machine:

```sh
mkdir -p $GOPATH/src/github.com/3scale
cd $GOPATH/src/github.com/3scale
git clone https://github.com/3scale/3scale-operator
cd 3scale-operator
git checkout master
```
## Create a new OpenShift project

```sh
oc new-project operator-test
```

This creates a new OpenShift project where the operator, the APIManager custom
resource and the Capabilities custom resources will be installed.

**It is very important to deploy all the elements in this new unique project, because deploying the resources in existing infrastructure could potentially alter existing elements**

## Register the 3scale-operator CRDs in the OpenShift API Server

Logged as an administrative user in the cluster, deploy all the
3scale-operator CRDs:

```sh
// As a cluster admin
for i in `ls deploy/crds/*_crd.yaml`; do oc create -f $i ; done
```

This will register the APIManager CRD and the CRDs related to the
Capabilities functionality of the operator in the OpenShift API Server.

If everything is ok then you should be able to query the resource types
defined by this CRDs via `oc get`.

For example, to verify that the APIManager CRD has been correctly
registered you can execute:

```sh
oc get apimanagers
```

At this moment you should see as the output:

```
No resources found.
```

## Deploy the needed roles and ServiceAccounts for the 3scale-operator

Go to the `operator-test` OpenShift project and make sure that no
other elements exist:

```sh
export NAMESPACE="operator-test"
oc project ${NAMESPACE}
oc get all // This shouldn't return any result
```

Deploy the ServiceAccount that will be used by the 3scale-operator:

```sh
oc create -f deploy/service_account.yaml
```

Log as an administrative user in the cluster and deploy
the 3scale-operator Role and the RoleBinding that will attach that role
to the created ServiceAccount:

```sh
// As a cluster admin
export NAMESPACE="operator-test"
oc project ${NAMESPACE}
oc create -f deploy/role.yaml
oc create -f deploy/role_binding.yaml
```

## Obtain and Set the desired operator image in the operator YAML

Set the operator's container image into the operator YAML. For example,
if you want to use the latest available operator image use:

```
sed -i 's|REPLACE_IMAGE|quay.io/3scale/3scale-operator:latest|g' deploy/operator.yaml
```

## Deploy the 3scale-operator

```
export NAMESPACE="operator-test"
oc project ${NAMESPACE}
oc create -f deploy/operator.yaml
```

This will create a Deployment that will contain a Pod with the Operator code and
will start listening to incoming APIManager and Capabilities resources.

## Deploy the APIManager custom resource

Deploying the APIManager custom resource will make the Operator begin
the processing of it and will deploy a 3scale API Management solution from it.

To deploy an APIManager, create a new YAML file with the following content:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  productVersion: <productVersion>
  wildcardDomain: <wildcardDomain>
  wildcardPolicy: <None|Subdomain>
  resourceRequirementsEnabled: true
```

If wildcardPolicy is `Subdomain`, wildcard routes at the OpenShift router
level need to be enabled. You can do so by executing
`oc set env dc/router ROUTER_ALLOW_WILDCARD_ROUTES=true -n default`

To look at more information on what the APIManager fields are
refer to the [Reference](apimanager-reference.md) documentation.

```sh
export NAMESPACE="operator-test"
oc project ${NAMESPACE}
oc create -f <yaml-name>
```

This should trigger the deployment of a 3scale API Management
solution in the "operator-test" project

## Deploy Tenants custom resource

Deploying the *APIManager* custom resource (see section above) creates a default tenant.
Optionally, you may create other tenants deploying **Tenant custom resource** objects.

To deploy a new tenant in your 3scale instance, create a new YAML file with the following content:

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
the [Tenant CRD Reference](tenant-crd-reference.md) documentation.

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
  aprovalRequired: false
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
 
## Cleanup

Delete the created custom resources:

Delete the APIManager custom resource and the 3scale API Management solution
elements that have been deployed from it. Deleting the APIManager will delete
all 3Scale API Management related objects in where it has been deployed:

```sh
oc delete -f <yaml-name-of-the-apimanager-custom-resource>
```

Delete the 3scale-operator operator, its associated roles and
service accounts

```sh
oc delete -f deploy/operator.yaml
oc delete -f deploy/role_binding.yaml
oc delete -f deploy/service_account.yaml
oc delete -f deploy/role.yaml
```

Delete the APIManager and Capabilities related CRDs:

```sh
oc delete -f deploy/crds/
```

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[oc]:https://docs.okd.io/3.11/cli_reference/get_started_cli.html#cli-reference-get-started-cli
