# Install the 3scale operator through Operator Lifecycle Manager (OLM).

You will need access to an OpenShift Container Platform 4.1 cluster.

Procedure
1. In the OpenShift Container Platform console, log in using an account with administrator privileges.
1. Create new project `operator-test` in *Projects > Create Project*.
1. Click *Catalog > OperatorHub*.
1. In the Filter by keyword box, type 3scale operator to find the 3scale operator.
1. Click the 3scale operator. Information about the Operator is displayed.
1. Click *Install*. The Create Operator Subscription page opens.
1. On the *Create Operator Subscription* page, accept all of the default selections and click Subscribe.
1. After the subscription *upgrade status* is shown as *Up to date*, click *Catalog > Installed Operators* to verify that the 3scale operator ClusterServiceVersion (CSV) is displayed and its Status ultimately resolves to _InstallSucceeded_ in the `operator-test` project.

# Deploying 3scale using the operator

## Prerequisites

* An OpenShift secret with the name "threescale-registry-auth" has to be created
  in every OpenShift project where an APIManager is desired to be deployed. This
  secret is used to be able to retrieve images from the registry.redhat.io
  container registry. The secret must have the credentials of a Service Account
  created in the registry.redhat.io container registry. The secret contents
  and more information about the registry can be found here: https://access.redhat.com/RegistryAuthentication.
  To create the secret: `oc create -f threescale-registry-auth-secret.yml` in
  the project where an APIManager is to be deployed

## Deployment

Deploying the *APIManager* custom resource will make the operator begin processing and will deploy a 3scale solution from it.

1. Click *Catalog > Installed Operators*. From the list of *Installed Operator*s, click _3scale Operator_. 
1. Click *API Manager > Create APIManager*
1. Create *APIManager* object with the following content.

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: <wildcardDomain>
```

The wildcardDomain parameter can be any desired name you wish to give that resolves to an IP address, which is a valid DNS domain. Be sure to remove the placeholder marks for your parameters: < >.

# Start using 3scale

When you deploy 3scale using the operator, a default tenant is created, with a fixed URL: `https://3scale-admin.${wildcardDomain}`
