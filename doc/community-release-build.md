# 3scale-Operator Community Release

[Introduction](#introduction)  
[Main Steps Briefly](#main-steps-briefly)  
[Versioning Strategy](#versioning-strategy)  
[Steps to Perform Community Release in 3scale-operator repo](#steps-to-perform-community-release-in-3scale-operator-repo)  
[Prepare release in community-operators-prod repo](#prepare-release-in-community-operators-prod-repo)
[References](#references)


## Introduction
* Community Release created in parallel with Minor Product Releases or patch releases, and not for Major release.
  * Community Release is not required for every productized build or every product patch release.
  * In this document: Product Releases == Product Releases or patch releases
* Community release can be released after downstream release
* Community Release codebase is defined by Tag in the [3scale-operator](https://github.com/3scale/3scale-operator) repository. 
Tag format as v\<major>.\<minor>.<build#> (for example: v0.10.1, v0.11.0)
* Community Release bundles are defined in `redhat-openshift-ecosystem` in [3scale-community-operator repo](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/3scale-community-operator) . 


## Main Steps Briefly

Below are briefly the main steps for prepare Community Release

* Create a community release development branch in **3scale-operator** repo, based on the Product release CommitID. Do development.
* Create PR from community release development branch to a **3scale-2.X-stable** branch (PR example: https://github.com/3scale/3scale-operator/pull/950)
  * Do development, Test and Merge PR to `3scale-2.X-stable` branch
    * **Note**: Build of the PR will create a release image, like https://quay.io/repository/3scale/3scale-operator?tab=tags&tag=v0.11.0
* Testing - Following Testing will be done:
  * Initial installation from Index image,
  * Upgrade from previous release
* Tagging - Apply community release Tag (like v0.11.0) to the latest CommitId on the community release development branch.
* Prepare Community release - create PR in **community-operators-prod** repo (based on manifests from 3scale-operator repo), 
as in [PR example](https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/4150)
  * **Merge of this PR will publish Release**.
* Do sanity test of the Release after publishing

## Versioning Strategy
Semantic Versioning scheme will be used for Community releases:  MAJOR.MINOR.PATCH.
* MAJOR version: Increments when breaking changes are introduced
* MINOR version: Increments when backward-compatible functionality is added.
* PATCH version: Increments when backward-compatible bug fixes are made.
* Community releases Versions examples:  0.8.2,  0.9.0,  0.10.1,  0.11.0
* The Version is used in Community releases Tag name

```shell
$ git tag |grep -E "0.9.0|0.10.1|0.11.0"
v0.10.1
v0.11.0
v0.9.0
```

* Relation to Product release versions
  * It will be  defined in community-operators-prod/operators/3scale-community-operator, as bundle channel 
  in `annotations.yaml` and `budle.Dockerfile`; for example:
```
0.10.1/bundle.Dockerfile:LABEL operators.operatorframework.io.bundle.channels.v1=threescale-2.13
0.10.1/metadata/annotations.yaml:  operators.operatorframework.io.bundle.channels.v1: threescale-2.13
```

* It’s recommended also to add the annotation to release Tag, as in example:
```  
$ git tag -n v0.11.0
v0.11.0         Community Release branch 3scale-community-v0.11.0, based on product 3scale-2.14.1-GA
```

## Steps to Perform Community Release in 3scale-operator repo
1. Community Release Development Branch

* Create a Community release development branch **based on Product Release CommitID**
* Create PR and target it to **3scale-2.X-stable** branch

```
$ cd 3scale-operator
$ git checkout -b 3scale-community-v0.11.0 bf398d34ae9378befc4e6e8bf447adbeef37c054
```
2. 3scale Components Images
* Community release is using its own Component images,for:
    - apisonator
    - apicast
    - porta
    - zync

3. Create PR in 3scale-operator upstream for new bundle
   * Work to have E2E tests passed in PR
   * Merge PR to **3scale-2.X-stable** branch
  
4. Testing
   
   * Testing of `3scale-operator` must be completed before open PR in `community-operators-prod` repo
   * Testing must be completed before Tagging
   * Testing must be done for `Fresh Install` and `Upgrade`
  

5. Tagging

- Community Release is defined by Tag.  
- Tag will be applied to the Community Release Branch
- Tag will be applied after Testing completion.

```
$ cd 3scale-operator
$ git tag -a v0.11.0  -m "Community Release branch 3scale-community-v0.11.0, based on product 3scale-2.14.1-GA" ab1783b4207e43480bf7538d62ecbd83636ecf0c
$ git push myfork v0.11.0
```

* Compare Product and Community release tags

```
$ git diff v0.11.0..3scale-2.14.1-GA --name-only
.circleci/config.yml
Makefile
bundle/manifests/3scale-operator.clusterserviceversion.yaml
config/manager/kustomization.yaml
config/manager/manager.yaml
config/manifests/bases/3scale-operator.clusterserviceversion.yaml
pkg/3scale/amp/component/images.go
```

## Prepare release in community-operators-prod repo
See Documentation in [Pull Requests in community operators project](https://github.com/operator-framework/community-operators/blob/master/docs/contributing-via-pr.md)

**IMPORTANT. All previouse steps, including Testing of 3scale-operator (Fresh install and Upgrade) must be completed before opening a PR in the community-operators-prod repo**

* Fork & Clone community-operators-prod repo if you don't have it yet
* Get latest changes

```
$ git clone git@github.com:redhat-openshift-ecosystem/community-operators-prod.git
$ cd community-operators-prod

Example for myfork:
$ git remote -v
myfork        git@github.com:valerymo/community-operators-prod.git (fetch)
myfork        git@github.com:valerymo/community-operators-prod.git (push)
origin        git@github.com:redhat-openshift-ecosystem/community-operators-prod.git (fetch)
origin        git@github.com:redhat-openshift-ecosystem/community-operators-prod.git (push)
```

* Create branch in community-operators-prod repo
```
$ git checkout -b 3scale-community-operator-v0.11.0
```

* Create a new release folder (similar to previous release)
```
community-operators-prod/operators/3scale-community-operator/0.11.0
```

* Copy manifests from 3scale-operator Community release to community-operators-prod, as for example From  [3scale-operator PR](https://github.com/3scale/3scale-operator/pull/950) To [community-operators-prod/operators/3scale-community-operator/0.11.0](community-operators-prod/operators/3scale-community-operator/0.11.0)

* Update CSV. These are things that need pay attention:
  * Components images
  * Databases images
  *  CVS version
  *  CSV replaces
  *  CSV name
  *  CSV description
  *  CSV urls
  *  CSV rht.prod_ver
  *  Annotations default channel
  *  Annotations package
  *  Annotations channel
  * * Update metadata/annotations  and bundle.Dockerfile
  

* Compare 3scale-community-operator bundle with previous version, and with 3scale-operator
* Finally you will have update release bundle structure.
* Commit your changes and Create PR 
  * Do Signed-off  - git commit -s … , It’s required for PR test Pipeline

```
[community-operators-prod] (3scale-community-operator-v0.11.0)$ git log
commit 10c2d3dee56a8b12ddb021dad4e1b13b301a72d7 (HEAD -> 3scale-community-operator-v0.11.0)
Author: xxx xxx <xxx@xxx.com>
Date:   Thu Mar 7 15:37:42 2024 +0200
3scale-community-operator release v0.11.0
Signed-off-by: xxx xxx <xxx@xxx.com>
```

```
$ git push -u myfork 3scale-community-operator-v0.11.0
To github.com:valerymo/community-operators-prod.git
....
```

**IMPORTANT. Merging of community-operators-prod PR will trigger Release creation/publishing**

* Check and confirm all questions in PR description
* Merge
* Test of created release in OSD cluster / OperatorHub
________________


## References

* [Community operators project documentation](https://github.com/operator-framework/community-operators/blob/master/docs/contributing-via-pr.md)
* [Community operators repository](https://github.com/k8s-operatorhub/community-operators)
* [K8S PR example](https://github.com/k8s-operatorhub/community-operators/pull/2472)
* [OCP PR example](https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/2382)
* *Check for the latest guidelines about what it needs to be done; they change from time to time*
  * *Community operator release process moved* [here - operator-release-process](https://github.com/operator-framework/community-operators/blob/master/docs/operator-release-process.md)
