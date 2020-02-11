# Development Guide

## Table of contents
* [Prerequisites](#prerequisites)
* [Clone repository](#clone-repository)
* [Building 3scale operator image](#building-3scale-operator-image)
* [Run 3scale Operator](#run-3scale-operator)
  * [Run 3scale Operator Locally](#run-3scale-operator-locally)
  * [Deploy 3scale Operator Manually](#deploy-3scale-operator-manually)
    * [Cleanup manually deployed operator](#cleanup-manually-deployed-operator)
  * [Deploy custom 3scale Operator using OLM](#deploy-custom-3scale-operator-using-olm)
* [Run tests](#run-tests)
* [Building 3scale templates](#building-3scale-templates)
* [Manifest management](#manifest-management)
  * [Verify operator manifest](#verify-operator-manifest)
  * [Push an operator bundle into external app registry](#push-an-operator-bundle-into-external-app-registry)
* [Licenses management](#licenses-management)
  * [Adding manually a new license](#adding-manually-a-new-license)

## Prerequisites

* [operator-sdk] version v0.8.0
* [git][git_tool]
* [go] version 1.13+
* [kubernetes] version v1.13.0+
* [oc] version v4.1+
* Access to a Openshift v4.1.0+ cluster.
* A user with administrative privileges in the OpenShift cluster.

## Clone repository

```sh
mkdir -p $GOPATH/src/github.com/3scale
cd $GOPATH/src/github.com/3scale
git clone https://github.com/3scale/3scale-operator
cd 3scale-operator
git checkout master
```

## Building 3scale operator image

[Clone the repository](#clone-repository)

Build operator image

```sh
make build IMAGE=quay.io/myorg/3scale-operator VERSION=test
```

## Run 3scale Operator

### Run 3scale Operator Locally

Run operator from command line, it will not be deployed as pod.

* [Clone the repository](#clone-repository)

* Register the 3scale-operator CRDs in the OpenShift API Server

```sh
// As a cluster admin
for i in `ls deploy/crds/*_crd.yaml`; do oc create -f $i ; done
```

* Create a new OpenShift project (optional)

```sh
export NAMESPACE=operator-test
oc new-project $NAMESPACE
```

* Run operator

```sh
make local
```

### Deploy 3scale Operator Manually

Build operator image and deploy manually as a pod.

* [Build 3scale operator image](#building-3scale-operator-image)

* Push image to public repo (for instance `quay.io`)

```sh
make push IMAGE=quay.io/myorg/3scale-operator VERSION=test
```

* Register the 3scale-operator CRDs in the OpenShift API Server

```sh
// As a cluster admin
for i in `ls deploy/crds/*_crd.yaml`; do oc create -f $i ; done
```

* Create a new OpenShift project (optional)

```sh
export NAMESPACE=operator-test
oc new-project $NAMESPACE
```

* Deploy the needed roles and ServiceAccounts

```sh
oc create -f deploy/service_account.yaml
oc create -f deploy/role.yaml
oc create -f deploy/role_binding.yaml
```

* Deploy the operator

```sh
sed -i 's|REPLACE_IMAGE|quay.io/myorg/3scale-operator:test|g' deploy/operator.yaml
oc create -f deploy/operator.yaml
```

#### Cleanup manually deployed operator

* Delete all `apimanager` custom resources:

```sh
oc delete apimanagers --all
```

* Delete the 3scale-operator operator, its associated roles and service accounts

```sh
oc delete -f deploy/operator.yaml
oc delete -f deploy/role_binding.yaml
oc delete -f deploy/service_account.yaml
oc delete -f deploy/role.yaml
```

* Delete the APIManager CRD:

```sh
oc delete crds apimanagers.apps.3scale.net
```

### Deploy custom 3scale Operator using OLM

To install this operator on OpenShift 4 using OLM for end-to-end testing, 

* [Push an operator bundle into external app registry](#push-an-operator-bundle-into-external-app-registry).

* Create the [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#4-create-the-operatorsource)
provided in `deploy/olm-catalog/3scale-operatorsource.yaml` to load your operator bundle in OpenShift.

```bash
oc create -f deploy/olm-catalog/3scale-operatorsource.yaml
```

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of
the OpenShift console _Catalog_. It can be easily found by filtering the provider type to _Custom_.

### Run tests

#### Run unittests

No need access to OCP cluster

```sh
make unit
make test-crds
```

#### Run integration tests

Access to a Openshift v4.1.0+ cluster required

* Run tests locally deploying image
```sh
export NAMESPACE=operator-test
make e2e-run
```

* Run tests locally running operator with go run instead of as an image in the cluster
```sh
export NAMESPACE=operator-test
make e2e-local-run
```

## Building 3scale templates

[Clone the repository](#clone-repository)

```sh
cd cd pkg/3scale/amp && make clean all
```

The location of the templates:
```
pkg/3scale/amp/auto-generated-templates
```

**NOTE**: If you want to use supported and stable templates you should go to the
[official repository](https://github.com/3scale/3scale-amp-openshift-templates)

## Manifest management

`operator-courier` is used for metadata syntax checking and validation.
This can be installed directly from pip:

```sh
pip3 install operator-courier
```

### Verify operator manifest

Check [Required fields within your CSV](https://github.com/operator-framework/community-operators/blob/master/docs/required-fields.md)

`operator-courier` will verify the fields included in the Operator metadata (CSV)

```sh
make verify-manifest
```

### Push an operator bundle into external app registry

* Get quay token

Detailed information on this [guide](https://github.com/operator-framework/operator-courier/#authentication)

```bash
curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '{"user": {"username": "YOURUSERNAME", "password": "YOURPASSWORD"}}' | jq '.token'
```

* Push bundle to Quay.io

Detailed information on this [guide](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#push-to-quayio).

```bash
make push-manifest APPLICATION_REPOSITORY_NAMESPACE=YOUR_QUAY_NAMESPACE MANIFEST_RELEASE=1.0.0 TOKEN=YOUR_TOKEN
```

## Licenses management

It is a requirement that a file describing all the licenses used in the product is included,
so that users can examine it.

* Check licenses when dependencies change.

```sh
make licenses-check
```

* Update `licenses.xml` file.

```sh
make licenses.xml
```

### Adding manually a new license

When licenses check does not parse correctly licensing information, it will complain.
In that case, you need to add manually license information.

There are two options: a)specify dependency license (recommended) or b)add exception for that dependency.

* Specify dependency license:

```sh
license_finder dependencies add YOURLIBRARY --decisions-file=doc/dependency_decisions.yml LICENSE --project-path "PROJECT URL"
```

For instance

```sh
license_finder dependencies add k8s.io/klog --decisions-file=doc/dependency_decisions.yml "Apache 2.0" --project-path "https://github.com/kubernetes/klog"
```

* Adding exception for a dependency:

```sh
license_finder approval add YOURLIBRARY --decisions-file=doc/dependency_decisions.yml --why "LICENSE_TYPE LINK_TO_LICENSE"
```

For instance

```sh
license_finder approval add github.com/golang/glog --decisions-file=doc/dependency_decisions.yml --why "Apache 2.0 License https://github.com/golang/glog/blob/master/LICENSE"
```

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[oc]:https://github.com/openshift/origin/releases
