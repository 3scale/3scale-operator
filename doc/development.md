# Development Guide

## Table of contents
* [Prerequisites](#prerequisites)
* [Clone repository](#clone-repository)
* [Building 3scale operator image](#building-3scale-operator-image)
* [Run 3scale Operator](#run-3scale-operator)
   * [Run 3scale Operator Locally](#run-3scale-operator-locally)
   * [Deploy custom 3scale Operator using OLM](#deploy-custom-3scale-operator-using-olm)
   * [Environment Variables](#3scale-operator-environment-variables)
   * [Run tests](#run-tests)
      * [Run all tests](#run-all-tests)
      * [Run unit tests](#run-unit-tests)
      * [Run end-to-end tests](#run-end-to-end-tests)
* [Building 3scale prometheus rules](#building-3scale-prometheus-rules)
* [Bundle management](#bundle-management)
   * [Generate an operator bundle image](#generate-an-operator-bundle-image)
   * [Push an operator bundle into an external container repository](#push-an-operator-bundle-into-an-external-container-repository)
   * [Validate an operator bundle image](#validate-an-operator-bundle-image)
* [Licenses management](#licenses-management)
   * [Adding manually a new license](#adding-manually-a-new-license)
* [Adding new watched secrets](#adding-new-watched-secrets)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## Prerequisites

* [operator-sdk] version v1.2.0
* [docker] version 17.03+
* [git][git_tool]
* [go] version 1.22+
* [kubernetes] version v1.13.0+
* [oc] version v4.1+
* Access to a Openshift v4.8.0+ cluster.
* A user with administrative privileges in the OpenShift cluster.
* Make sure that the `DOCKER_ORG` and `DOCKER_REGISTRY` environment variables are set to the same value as
  your username on the container registry, and the container registry you are using.

```sh
export DOCKER_ORG=docker_hub_username
export DOCKER_REGISTRY=quay.io
```

## Clone repository

```sh
git clone https://github.com/3scale/3scale-operator
cd 3scale-operator
```

## Building 3scale operator image

Build the operator image

```sh
make docker-build-only IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator:myversiontag
```

## Run 3scale Operator

### Run 3scale Operator Locally

Run operator from the command line, it will not be deployed as a pod.

* Register the 3scale-operator CRDs in the OpenShift API Server

```sh
make install
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

* Alternatively run:

```sh
make cluster/prepare/local
```
This command will
- create namespace called 3scale-test (namespace name can be adjusted with `NAMESPACE` envar)
- install the CRDs
- provision backend Redis instance and create required by 3scale secret
- provision system Redis instance and create required by 3scale secret
- provision MySQL database instance and create required by 3scale secret (alternatively, you can choose PostgreSQL by setting envar: `DEV_SYSTEM_DB_POSTGRES=true`)

### Deploy custom 3scale Operator using OLM

* Build and upload custom operator image
```
make docker-build-only IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator:myversiontag
make operator-image-push IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator:myversiontag
```

* Build and upload custom operator bundle image. Changes to avoid conflicts will be made by the makefile.
```
make bundle-custom-build IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator:myversiontag BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator-bundles:myversiontag
make bundle-image-push BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator-bundles:myversiontag
```

* Deploy the operator in your currently configured and active cluster in $HOME/.kube/config

```
make bundle-run BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/3scale-operator-bundles:myversiontag
```

**Note**: The _catalogsource_ will be installed in the `openshift-marketplace` namespace
[issue](https://bugzilla.redhat.com/show_bug.cgi?id=1779080). By default, cluster scoped
subscription will be created in the namespace `openshift-marketplace`.
Feel free to delete the operator (from the UI **OperatorHub -> Installed Operators**)
and install it namespace or cluster scoped.

It will take a few minutes for the operator to become visible under
the _OperatorHub_ section of the OpenShift console _Catalog_. It can be
easily found by filtering the provider type to _Custom_.

### 3scale Operator Environment Variables
There are environment variables that may be used to aid in development. Refer to the table below for details:

| Variable                    | Options    |   Type   | Default | Details                                                                                                                                                    |
|-----------------------------|------------|:--------:|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| THREESCALE_DEBUG            | `1` or `0` | Optional | `0`     | If `1`, sets the porta client logging to be more verbose.                                                                                                  |

### Run tests

#### Run all tests

Access to a Openshift v4.8.0+ cluster required

```sh
make test
```

#### Run unit tests

```sh
make test-unit
```

#### Run end-to-end tests

Access to a Openshift v4.8.0+ cluster required

```sh
WATCH_NAMESPACE=3scale-test make test-e2e
```
Note that the value of the WATCH_NAMESPACE is irrelevant but required. E2E test will create an new namespace regardless of the value of WATCH_NAMESPACE 

## Building 3scale prometheus rules

[Clone the repository](#clone-repository)

```sh
make prometheus-rules
```

Optionally, specify the namespace. By default, the namespace `__NAMESPACE__` will be used.

```sh
make prometheus-rules PROMETHEUS_RULES_NAMESPACE=my-custom-namespace
```

## Bundle management

### Generate an operator bundle image

```sh
make bundle-build BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/myrepo:myversiontag
```

### Push an operator bundle into an external container repository

```sh
make bundle-image-push BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/myrepo:myversiontag
```

### Validate an operator bundle image

NOTE: if validating an image, the image must exist in a remote registry, not just locally.

```
make bundle-validate-image BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/myrepo:myversiontag
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

## Adding new watched secrets
After adding a new secret to the APIManager CRD make sure to also update the following files if you want the 3scale-operator to watch the new secret:
1. [apis/apps/v1alpha1/apimanager_types.go](../apis/apps/v1alpha1/apimanager_types.go)
    - Add a new `GetXYZSecretRef()` function that returns the secret ref
2. [apis/apps/v1alpha1/apimanager_types.go](../apis/apps/v1alpha1/apimanager_types.go)
    - Update the `Get3scaleSecretRefs()` to call the new `GetXYZSecretRef()` function from step 1
3. [pkg/3scale/amp/operator/apicast_reconciler.go](../pkg/3scale/amp/operator/apicast_reconciler.go)
    - Add the new secret to the `getSecretUIDs()` function
4. [pkg/3scale/amp/component/deployment_annotations.go](../pkg/3scale/amp/component/deployment_annotations.go)
    - Add the new secret to the `getWatchedSecretAnnotations()` function
5. [pkg/3scale/amp/component/deployment_annotations.go](../pkg/3scale/amp/component/deployment_annotations.go)
    - Add the new secret to the switch in the `HasSecretHashChanged()` function
6. pkg/3scale/amp/component/{component_name}.go
    - Add a new const called `XYZSecretResverAnnotationPrefix` that can be referenced throughout the code
      - The const should be in the `component` package but the exact file will depend on which deployment the new watched secret relates to. For example if the secret is relevant to the `apicast` deployments, the const belongs in [pkg/3scale/amp/component/apicast.go](../pkg/3scale/amp/component/apicast.go)
7. pkg/3scale/amp/component/{component_name}.go
   - Add an annotation for the new watched secret to the relevant deployment's `.spec.template.metadata.annotations`
      - The exact file that needs changing will depend on which deployment the new watched secret relates to. See the `StagingDeployment()` function in [pkg/3scale/amp/component/apicast.go](../pkg/3scale/amp/component/apicast.go) for an example
