# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [0.7.0] - 2021-05-12

### Added

- APIManager CRD 3scale 2.10
- Oracle support [#416](https://github.com/3scale/3scale-operator/pull/416) [#445](https://github.com/3scale/3scale-operator/pull/445)
- Application Capabilities: provider account from 3scale in the same namespace [#443](https://github.com/3scale/3scale-operator/pull/443)
- `last` attribute to MappingRule in Backend and Product types [#447](https://github.com/3scale/3scale-operator/pull/447)
- Resource requirementes for each 3scale component [#454](https://github.com/3scale/3scale-operator/pull/454)
- Resource requirementes for the 3scale operarator [#461](https://github.com/3scale/3scale-operator/pull/461)
- Upgraded grafana dependency to 3.5.0 [#475](https://github.com/3scale/3scale-operator/pull/475)
- System prometheus monitoring [#458](https://github.com/3scale/3scale-operator/pull/458)
- Grafana dashboad for the apicast metrics [#473](https://github.com/3scale/3scale-operator/pull/473)
- Resource pvc system storage and databases [#467](https://github.com/3scale/3scale-operator/pull/467)
- ConsoleLink for the 3scale master route [#462](https://github.com/3scale/3scale-operator/pull/462)
- Apicast workers monitoring alert [#484](https://github.com/3scale/3scale-operator/pull/484)
- Allow Zync Database to be configured as an external database on the operator [#489](https://github.com/3scale/3scale-operator/pull/489)
- Disconnected install support [#492](https://github.com/3scale/3scale-operator/pull/492)
- Added grafana dashboard for System Sidekiq metrics [#466](https://github.com/3scale/3scale-operator/pull/466)
- Reconcile monitoring resources [#599](https://github.com/3scale/3scale-operator/pull/599)
- APIManager: configurable ServiceAccount ImagePullSecrets for the managed DCs [#599](https://github.com/3scale/3scale-operator/pull/599)
- APIManager: apicast production worker field [#599](https://github.com/3scale/3scale-operator/pull/599)

### Fixed

- Regex validation in Product type for fields that specify USD [#446](https://github.com/3scale/3scale-operator/pull/446)
- Tolerations in postgresql pvc [#465](https://github.com/3scale/3scale-operator/pull/465)
- Reconciliate product type on backend type's metric deletion [#444](https://github.com/3scale/3scale-operator/pull/444)
- Reconcile system database secret when external databases enabled [#486](https://github.com/3scale/3scale-operator/pull/486)
- CVE-2020-14040 [#599](https://github.com/3scale/3scale-operator/pull/599)
- CVE-2020-9283 [#599](https://github.com/3scale/3scale-operator/pull/599)

### Changed
- Monitoring alerts severity [#481](https://github.com/3scale/3scale-operator/pull/481) [#488](https://github.com/3scale/3scale-operator/pull/488)

## [0.6.0] - 2020-08-28

### Added
- APIManager CRD 3scale 2.9
- Upgrade between 2.8 and 2.9 (no specific PR, implementation distributed in feature PR's)
- Metering labels [#367](https://github.com/3scale/3scale-operator/pull/367)
- Add configurable Storage Class for all PVCs [#386](https://github.com/3scale/3scale-operator/pull/386)
- backup/restore functionality - system deployment mode with PVC [#392](https://github.com/3scale/3scale-operator/pull/392)
- Add affinity and tolerations APIManager configurability for DeploymentConfigs [#384](https://github.com/3scale/3scale-operator/pull/384)
- Monitoring resources [#333](https://github.com/3scale/3scale-operator/pull/333)
- Operator capabilities V2: Product and Backend CRD's [#357](https://github.com/3scale/3scale-operator/pull/357)

### Changed
- Use new ImageStream tagging structure [#292](https://github.com/3scale/3scale-operator/pull/292)
- APIManager CRD system's app spec and sidekiq spec now optional [#394](https://github.com/3scale/3scale-operator/pull/394)

## [0.5.0] - 2020-04-02

### Added
- APIManager CRD 3scale 2.8
- Upgrade between 2.7 and 2.8 (no specific PR, implementation distributed in feature PR's)
- Metrics endpoint include 3scale operator and product version [#290](https://github.com/3scale/3scale-operator/pull/290)
- Move system-smtp to secret [#280](https://github.com/3scale/3scale-operator/pull/280)
- Support s3 apicompatible service [#302](https://github.com/3scale/3scale-operator/pull/302)
- Add support for PodDisruptionBudget [#308](https://github.com/3scale/3scale-operator/pull/308)
- Operator-sdk v0.15.2 Go version v1.13 [#331](https://github.com/3scale/3scale-operator/pull/331)

## [0.4.0] - 2019-11-28

### Added
- APIManager CRD 3scale 2.7
- Make replicas configurable in scalable DeploymentConfigs [#220](https://github.com/3scale/3scale-operator/pull/220)
- Upgrade between 2.6 and 2.7 [#203](https://github.com/3scale/3scale-operator/pull/203)

## [0.3.0] - 2019-11-06

### Added
- APIManager CRD 3scale 2.6
- Migrate DBs to imagestreams [#122](https://github.com/3scale/3scale-operator/pull/122)
- Operator-sdk v0.8.0 [#126](https://github.com/3scale/3scale-operator/pull/126)
- Authenticated registry.redhat.io [#124](https://github.com/3scale/3scale-operator/pull/124)
- PostgreSQL support [#129](https://github.com/3scale/3scale-operator/pull/129)
- Wildcard router removed [#135](https://github.com/3scale/3scale-operator/pull/135)
- Redis sentinels [#137](https://github.com/3scale/3scale-operator/pull/137)
- Zync-Que [#145](https://github.com/3scale/3scale-operator/pull/145)
- OLM catalog [#71](https://github.com/3scale/3scale-operator/pull/71)

## [0.2.0] - 2019-04-02

### Added
- APIManager CRD 3scale 2.5
- Tenants CRD
- Capabilities CRD (api, limit, plan, binding, metric, mappingrule)

### Added

[Unreleased]: https://github.com/3scale/3scale-operator/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.7.0
[0.6.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.6.0
[0.5.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.5.0
[0.4.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.4.0
[0.3.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.3.0
[0.2.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.2.0
