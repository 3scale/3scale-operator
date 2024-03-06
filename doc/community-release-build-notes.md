# 3SCALE-OPERATOR Community Release - process notes


## Introduction

- Community Release created in parallel with Minor Product Releases
- Community Release created after QE sign off but before full productization (product release shipping)
- Community Release Not required for every productized builds or every product release patch
- Community Release can be synced with specific Product release, and can be Not synced
- Community Release is defined by **Tag** (like v0.09.0, v0.10.1, v0.11.0) 
in 3scale-operator codebase repo https://github.com/3scale/3scale-operator  
Please see [Community Release Tagging](#community-release-tagging)
- OLM Channel definition (for Operator Hub, for v0.10.1): in https://github.com/redhat-openshift-ecosystem/community-operators-prod/blob/main/operators/3scale-community-operator/0.10.1/metadata/annotations.yaml


## Steps to Perform Community Release

**1. Create Branch in 3scale-operator repo (if required)**
https://github.com/3scale/3scale-operator

Branching process may differ for Community releases, depends on following cases:
- Community Release is identical to minor Product release
- Community Release is based on minor Product releases, but has additional features that are not in Product yet.
- Community Release is not synced with Product release

- If Community Release is identical to minor Product release - no need create Branch, Tag will apply to CommitID on existing branch 
- If Community Release is based on minor Product releases, but has additional features, that not in Product yet
  - Community release Branch needs to be created based on Product release Tag (if already available), or stable branch (for example: 3scale-2.14-stable).
  - Branch name will be like 3scale-2.14-stable-community-v0.11
    - This Community release Branch could be used for all patches of this community release; that will have tags like v0.11.0, v0.11.1,..
- Community Release is not synced with Product release    
  - Notes. If community release can be based on CommitID in master, and does not have additional features - the branch is not required, 
  otherwise - it is similar to the previous case - create a community branch from a stable Product release branch.

**2. Development**
- If required - community release can include `3scale-operator` feature, that not in Product release. The feature can require development.

**3. Tagging**
- Tag is applied to **CommitID**
- Tag examples: v0.09.0, v0.10.1, v0.11.0 (see [3scale-operator repo - Tags](https://github.com/3scale/3scale-operator)

**4. Testing**  
- Community release should pass development testing.
- TODO
- **Notes**. Community 3scale-operator Release is a release which has not been vetted or verified by Red Hat. 
  Red Hat provides no support for community Operators.  
  [Learn more about Red Hat’s third party software support policy](https://access.redhat.com/third-party-software-support)

**5. Prepare release in community-operators-prod repo**
- see 3scale Release configuration in [3scale-community-operator](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/3scale-community-operator) 
**IMPORTANT**  _This section contains step for PR creation in community-operators-prod repo. 
This PR will be merged ONLY after completion of Step 4-Testing. 
Merging of community-operators-prod PR will trigger Release creation/publishing.
See next section for details_

- You need to have your Fork and local replica of [3scale-community-operator](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/3scale-community-operator)
```
git clone git@github.com:redhat-openshift-ecosystem/community-operators-prod.git
git remote add myfork git@github.com:<my user>/community-operators-prod.git
```
- Create your branch (if required) 

```shell
$ git checkout -b 3scale-community-operator-v0.11.0
Switched to a new branch '3scale-community-operator-v0.11.0'
```
- Create a new release folder similar to below. 
_You can copy from existing one and do required configuration_,   
```shell 
$cd community-operators-prod/operators/3scale-community-operator
$ cp -r 0.10.1 0.11.0
```
- Update manifests, metadata and docker file in new release folder (0.11.0 in this example)

- Create PR - see [references and examples](#references-and-examples)
- For more Process details:  [contributing-via-pr.md](https://github.com/redhat-openshift-ecosystem/community-operators-prod/blob/main/docs/contributing-via-pr.md)

- **IMPORTANT**
  - Merge PR in community-operators-prod repo only after Testing completed for 3scale-operator Release in 3scale-operator repo!***  See section [Testing](#4-testing)
  - When the PR is merged, the community release pipeline will be triggered  
  
_You can find some Useful commands interacting with the pipeline [here]( https://github.com/redhat-openshift-ecosystem/community-operators-prod/blob/main/docs/contributing-via-pr.md#useful-commands-interacting-with-the-pipeline)_

- Pipeline creation 
  - Please see  [Documentation for community-operator-pipeline](https://github.com/redhat-openshift-ecosystem/community-operators-prod/blob/main/docs/community-operator-pipeline.md)
  The community operator pipeline is divided into two workflows (please see documentation):
    - The **community-hosted-pipeline** to test and validate the submitted operator bundle in the PR.
    - The **community-release-pipeline** to release the operator bundle to the catalog after the PR is merged.


### References and examples

- repo managing the catalog of operatorhub.io for **K8S**: https://github.com/k8s-operatorhub/community-operators
- repo managing the catalog of community operators for **OCP** https://github.com/redhat-openshift-ecosystem/community-operators-prod
  - _**Note**_:  
    - _apicast is available both in K8S and OCP repos_
    - _3scale  operator is available in OCP repo_
- K8S PR example https://github.com/k8s-operatorhub/community-operators/pull/2472  
- OCP PR example https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/2382
  - _**Note**: check for the latest guidelines about what it needs to be done; they change from time to time_



