# 3scale-operator

[![CircleCI](https://circleci.com/gh/3scale/3scale-operator/tree/master.svg?style=svg)](https://circleci.com/gh/3scale/3scale-operator/tree/master)

### Project Status: alpha

The project is currently alpha which means that there are still new feautres
and APIs planned that will be added in the future.
Due to this, breaking changes may still happen.

Only use for short-lived testing clusters. Do not deploy it in the same
OpenShift project than one having an already existing
3scale API Management solution, as it could potentially alter/delete the
existing elements in the project.

## Overview

This project contains the 3scale operator software. 3scale operator is a
software based on [Kubernetes operators](https://coreos.com/operators/) that
provides:
* A way to install a 3scale API Management solution, providing configurability
  options at the time of installation
* Ability to define 3scale API definitions and set
  them into a 3scale API Management solution

This functionalities definitions are provided via Kubernetes custom resources
that the operator is able to understand and process.

## Prerequisites

* [operator-sdk] version v0.5.0.
* [dep][dep_tool] version v0.5.0+.
* [git][git_tool]
* [go] version 1.11+
* [kubernetes] version v1.11.0+
* [oc] version v3.11+
* Access to a Openshift v3.11.0+ cluster.
* A user with administrative privileges in the OpenShift cluster.

## Getting started

Download the 3scale-operator project:

```sh
mkdir -p $GOPATH/src/github.com/3scale
cd $GOPATH/src/github.com/3scale
git clone https://github.com/3scale/3scale-operator
cd 3scale-operator
git checkout master
```

Create and deploy a 3scale-operator and the custom resources needed
to install a sample 3scale API Management solution and API definitions:

```sh
# As an OpenShift administrative user create the 3scale-operator CRDs:
for i in `ls deploy/crds/*_crd.yaml`; do oc create -f $i ; done

# Create a new empty project (this can be done with any desired OpenShift user)
# ** It is very important to deploy all the elements in this new unique project,
# because deploying the resources in existing infrastructure could
# potentially alter/delete existing 3scale elements **
export NAMESPACE="operator-test"
oc new-project $NAMESPACE
oc project $NAMESPACE

# Create the 3scale-operator ServiceAccount
oc create -f deploy/service_account.yaml

# Create the roles and role bindings associated to the 3scale-operator
# to be deployed
oc create -f deploy/role.yaml
oc create -f deploy/role_binding.yaml

# Set the desired operator image in the operator YAML. For example,
# the latest available one
sed -i 's|REPLACE_IMAGE|quay.io/3scale/3scale-operator:latest|g' deploy/operator.yaml

# Deploy the 3scale-operator in the created project
oc create -f deploy/operator.yaml

# Verify that the operator is deployed and ready. Execute the
# following command until it shows the Deployment is ready
oc get deployment 3scale-operator
```

At this point, the 3scale-operator is deployed and ready to accept the
3scale-operator custom resource creation requests that it can process to
perform the provide the functionalities described in the [Overview](#Overview)
section.

To see a deploy example of the 3scale-operator, how to deploy example custom
resources to deploy a 3scale API Management solution and 3scale API definitions,
and how to cleanup the operator and the custom resources
refer to the [User guide](doc/user-guide.md).

## Development

Assuming you have already downloaded the 3scale-operator project (see
[Getting Started](#Getting-started)), install the project dependencies:

```sh
make vendor
```

After performing the desired changes in the code, the operator can be executed
locally via the following Makefile rule:
`make local`

## Testing

Local tests can be executed with the following commands:
```
make e2e-setup
make e2e-local-run
```

## Documentation

* [User guide](doc/user-guide.md)
* [APIManager reference](doc/apimanager-reference.md)
* [Tenant reference](doc/tenant-reference.md)
* [Capabilities reference](doc/api-crd-reference.md) 

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[oc]:https://docs.okd.io/3.11/cli_reference/get_started_cli.html#cli-reference-get-started-cli
