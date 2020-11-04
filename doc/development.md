# Development Guide

## Table of contents
* [Prerequisites](#prerequisites)
* [Clone repository](#clone-repository)
* [Building 3scale operator image](#building-3scale-operator-image)
* [Run 3scale Operator](#run-3scale-operator)
  * [Run 3scale Operator Locally](#run-3scale-operator-locally)
  * [Deploy custom 3scale Operator using OLM](#deploy-custom-3scale-operator-using-olm)
* [Run tests](#run-tests)
  * [Run all tests](#run-all-tests)
  * [Run unit tests](#run-unit-tests)
  * [Run end-to-end tests](#run-end-to-end-tests)
* [Building 3scale templates](#building-3scale-templates)
* [Bundle management](#bundle-management)
  * [(re)Generate an operator bundle image](#generate-an-operator-bundle-image)
  * [Validate an operator bundle image](#validate-an-operator-bundle-image)
  * [Push an operator bundle into an external container repository](#push-an-operator-bundle-into-an-external-container-repository)
* [Licenses management](#licenses-management)
  * [Adding manually a new license](#adding-manually-a-new-license)

## Prerequisites

* [operator-sdk] version v1.1.0
* [docker] version 17.03+
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

Build the operator image

```sh
make docker-build IMG=quay.io/myorg/3scale-operator:myversiontag
```

## Run 3scale Operator

### Run 3scale Operator Locally

Run operator from the command line, it will not be deployed as a pod.

* [Clone the repository](#clone-repository)

* Register the 3scale-operator CRDs in the OpenShift API Server

```sh
// As a cluster admin
for i in `ls bundle/manifests/**apps.3scale.net_*.yaml`; do oc create -f $i ; done
for i in `ls bundle/manifests/**capabilities.3scale.net_*.yaml`; do oc create -f $i ; done
```

* Create a new OpenShift project (optional)

```sh
export NAMESPACE=operator-test
oc new-project $NAMESPACE
```

* Install the dependencies

```sh
make download
```

* Run operator

```sh
make run
```

### Deploy custom 3scale Operator using OLM

To install this operator on an OpenShift 4.5+ cluster using OLM for end-to-end testing:

* Perform naming changes to avoid collision with existing 3scale Operator
  official public operators catalog entries:
  * Edit the `bundle/manifests/3scale-operator.clusterserviceversion.yaml` file
    and perform the following changes:
      * Change the current value of `.metadata.name` to a different name
        than `3scale-operator.v*`. For example to `myorg-3scale-operator.v0.0.1`
      * Change the current value of `.spec.displayName` to a value that helps you
        identify the catalog entry name from other operators and the official
        3scale operator entries. For example to `"MyOrg 3scale operator"`
      * Change the current value of `.spec.provider.Name` to a value that helps
        you identify the catalog entry name from other operators and the official
        3scale operator entries. For example, to `MyOrg`
  * Edit the `bundle.Dockerfile` file and change the value of
    the Dockerfile label `LABEL operators.operatorframework.io.bundle.package.v1`
    to a different value than `3scale-operator`. For example to
    `myorg-3scale-operator`
  * Edit the `bundle/metadata/annotations.yaml` file and change the value of
    `.annotations.operators.operatorframework.io.bundle.package.v1` to a
    different value than `3scale-operator`. For example to
    `myorg-3scale-operator`. The new value should match the
    Dockerfile label `LABEL operators.operatorframework.io.bundle.package.v1`
    in the `bundle.Dockerfile` as explained in the point above

  It is really important that all the previously shown fields are changed
  to avoid overwriting the 3scale operator official public operator
  catalog entry in your cluster and to avoid confusion having two equal entries
  on it.

  * [Create an operator bundle image](#generate-an-operator-bundle-image) using the
  changed contents above

  * [Push the operator bundle into an external container repository](#push-an-operator-bundle-into-an-external-container-repository).

  * Run the following command to deploy the operator in your currently configured
    and active cluster in $HOME/.kube/config:
    ```sh
    operator-sdk run bundle --namespace <mynamespace> <BUNDLE_IMAGE_URL>
    ```

    Additionally, a specific kubeconfig file with a desired Kubernetes
    configuration can be provided too:
    ```sh
    operator-sdk run bundle --namespace <mynamespace> --kubeconfig <path> <BUNDLE_IMAGE_URL>
    ```

It will take a few minutes for the operator to become visible under
the _OperatorHub_ section of the OpenShift console _Catalog_. It can be
easily found by filtering the provider type to _Custom_.

### Run tests

#### Run all tests

Access to a Openshift v4.1.0+ cluster required

```sh
make test
```

#### Run unit tests

```sh
make test-unit
```

#### Run end-to-end tests

Access to a Openshift v4.1.0+ cluster required

```sh
make test-e2e
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

## Bundle management

### Generate an operator bundle image

```sh
make bundle
```

The generated output will be saved in the `bundle` directory

### Validate an operator bundle image

```sh
make bundle-validate
```

### Push an operator bundle into an external container repository

```sh
make docker-push IMG=quay.io/myorg/3scale-operator:myversiontag
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

There are two options: a)specify dependency license (recommended) or b)add an exception for that dependency.

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
[docker]:https://docs.docker.com/install/
[kubernetes]:https://kubernetes.io/
[oc]:https://github.com/openshift/origin/releases
