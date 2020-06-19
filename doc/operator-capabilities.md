# Application Capabilities via Operator

Featured capabilities:

* Allow interacting with the underlying 3scale API Management solution.
* Manage the 3scale application declaratively using openshift (custom) resources.

The following diagram shows 3scale entities and relations that will be eligible for management using openshift (custom) resources in  a declarative way.

![3scale Object types](3scale-diagram.png)

The following diagram shows available custom resource definitions and their relations provided by the 3scale operator.

![3scale Custom Resources](capabilities-diagram.png)

## Table of contents

* [Tenant custom resource](#tenant-custom-resource)
  * [Tenant CRD reference](tenant-reference.md)
* [Backend custom resource](#backend-custom-resource)
  * [Backend CRD reference](backend-reference.md)
* WIP [Product](product_reference.md)
* WIP [Account](Account_reference.md)
* WIP [ActiveDoc](activedoc_reference.md)

## Tenant custom resource

Tenant is also known as Provider Account.

Creating the [*APIManager*](apimanager-reference.md) custom resource tells the operator to deploy 3scale.
Default 3scale installation includes a default tenant ready to be used. Optionally,
you may create other tenants creating [Tenant](tenant_reference.md) custom resource objects.

### Preparation before deploying the new tenant

To deploy a new tenant in your 3scale instance, first you need some preparation steps:

* Create or local 3scale Master credentials secret: *MASTER_SECRET*
* Create a new secret to store the password for the admin account of the new tenant: *ADMIN_SECRET*
* Get the 3scale master account hostname: *MASTER_HOSTNAME*


A) *3scale Master credentials secret: MASTER_SECRET*

Tenant management can only be done using 3scale *master* account. You need *master* account credentials (preferably and access token). 

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

The 3scale operator will create a new secret and store new tenant's credentials in it. The new tenant *provider_key* and *admin domain url* will be stored in a secret.
The secret location can be specified using *tenantSecretRef* tenant spec key.

Example of the created secret content:

```
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
