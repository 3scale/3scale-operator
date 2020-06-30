# 3scale-operator

[![CircleCI](https://circleci.com/gh/3scale/3scale-operator/tree/master.svg?style=svg)](https://circleci.com/gh/3scale/3scale-operator/tree/master)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![GitHub release](https://img.shields.io/github/v/release/3scale/3scale-operator.svg)](https://github.com/3scale/3scale-operator/releases/latest)
[![codecov](https://codecov.io/gh/3scale/3scale-operator/branch/master/graph/badge.svg)](https://codecov.io/gh/3scale/3scale-operator)

## Overview

The 3scale Operator creates and maintains the Red Hat 3scale API Management on [OpenShift](https://www.openshift.com/) in various deployment configurations.
[3scale API Management](https://www.redhat.com/en/technologies/jboss-middleware/3scale) makes it easy to manage your APIs.
Share, secure, distribute, control, and monetize your APIs on an infrastructure platform built for performance, customer control, and future growth.

## Quickstart

To get up and running quickly, check our [Quickstart guides](doc/quickstart-guide.md).

## Features

Current *capabilities* state: **Full Lifecycle**

* Stable
  * **Installer**: A way to install a 3scale API Management solution, providing configurability options at the time of installation
  * **Upgrade**: Upgrade from previously installed 3scale API Management solution
  * **Reconciliation**: Tunable CRD parameters after 3scale API Management solution has been installed
* Tech Preview
  * **Application Capabilities via Operator**: Allow interacting with underlying 3scale API Management solution. Expose objects like *tenant*, *product*, *backend* as  _Custom Resource_ objects.

## User Guide

* Check our [Operator user guide](doc/operator-user-guide.md) for interacting with the 3scale operator.
* Check our [Template user guide](doc/template-user-guide.md) for deploying 3scale in various deployment profiles.

## Contributing
You can contribute by:

* Raising any issues you find using 3scale Operator
* Fixing issues by opening [Pull Requests](https://github.com/3scale/3scale-operator/pulls)
* Submitting a patch or opening a PR
* Improving documentation
* Talking about 3scale Operator

All bugs, tasks or enhancements are tracked as [GitHub issues](https://github.com/3scale/3scale-operator/issues).

The [Development guide](doc/development.md) describes how to build the 3scale Operator and how to test your changes before submitting a patch or opening a PR.

## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.
