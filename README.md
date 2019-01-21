# 3scale-operator

## Prerequisites

- [operator-sdk] version v0.2.1.
- [dep][dep_tool] version v0.5.0+.
- [git][git_tool]
- Access to a Openshift v3.11.0+ cluster.

## Quick Start

To download and prepare the environment for 3scale-operator:

```sh
$ mkdir -p $GOPATH/src/github.com/3scale
$ cd $GOPATH/src/github.com/3scale
$ git clone https://github.com/3scale/3scale-operator
$ cd 3scale-operator
$ git checkout master
$ make vendor
```

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[dep_tool]:https://golang.github.io/dep/docs/installation.html
