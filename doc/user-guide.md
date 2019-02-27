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
  * Binding (TODO)
  * Consolidated
  * Limit
  * MappingRule
  * Metric
  * Plan
  * Tenant

## Prerequisites

* [operator-sdk] version v0.2.1.
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

To look at more information on what the APIManager fields are
refer to the [Reference](reference.md) documentation.

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
  systemMasterUrl: https://master.amp24.127.0.0.1.nip.io
  email: admin@ecorp.com
  organizationName: ECorp
  masterCredentialsRef:
    name: system-seed
    namespace: operator-test
  passwordCredentialsRef:
    name: ecorp-admin-secret
    namespace: operator-test
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
Refer to [Tenant CRD Reference](tenant-crd-reference.md) documentation for more information.

## Deploy the Capabilities related custom resources

TODO

## Cleanup

Delete the created custom resources:

Delete the APIManager custom resource and the 3scale API Management solution
elements that have been deployed from it. Deleting the APIManager will delete
all 3Scale API Management related objects in where it has been deployed:

```sh
oc delete -f <yaml-name-of-the-apimanager-custom-resource>
```

Delete the capabilities custom resources

TODO

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
for i in `ls deploy/crds/*_crd.yaml`; do oc delete -f $i ; done
```

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[oc]:https://docs.okd.io/3.11/cli_reference/get_started_cli.html#cli-reference-get-started-cli