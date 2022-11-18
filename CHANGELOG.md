# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

### Added

- Bumped Go to 1.17 [#785](https://github.com/3scale/3scale-operator/pull/785)
- Bumped k8s deps to 0.24.3 [#785](https://github.com/3scale/3scale-operator/pull/785)
- Bumped controller-runtime v0.12.2 [#785](https://github.com/3scale/3scale-operator/pull/785)
- OCP 4.12 support  [#785](https://github.com/3scale/3scale-operator/pull/785)
- Application CRD [#778](https://github.com/3scale/3scale-operator/pull/778)
- Ability to set APICAST_SERVICE_CACHE_SIZE  [#793](https://github.com/3scale/3scale-operator/pull/793)
- Expose apicast metric ports in service [#791](https://github.com/3scale/3scale-operator/pull/791)

### Fixed

- Fixed replica reconciliation [#784](https://github.com/3scale/3scale-operator/pull/784)

## [0.10.0] - 2022-11-18

### Added

- Support removal of a developer account CR [#741](https://github.com/3scale/3scale-operator/pull/741)
- Support removal of a developer user CR [#751](https://github.com/3scale/3scale-operator/pull/741)
- Skipping apicast service port reconcile based on annotation [#739](https://github.com/3scale/3scale-operator/pull/739)
- Disable replicas reconciliation when annotation is present [#736](https://github.com/3scale/3scale-operator/pull/736)
- Add sop_url as annotation to alerts [#526](https://github.com/3scale/3scale-operator/pull/526)
- ProxyConfigPromote CRD [#742](https://github.com/3scale/3scale-operator/pull/742)
- `secret_key_base` to system sphinx [#762](https://github.com/3scale/3scale-operator/pull/762) [#783](https://github.com/3scale/3scale-operator/pull/783)

### Changed

- Use `sum_irate` instead of `sum_rate` [#740](https://github.com/3scale/3scale-operator/pull/740)
- APImanager CRD: preserve unknown fields to support multiple version in the same cluster [#754](https://github.com/3scale/3scale-operator/pull/754)
- :warning: Templates dropped [#764](https://github.com/3scale/3scale-operator/pull/764)
- Upgrade moved to regular reconciliation logic. It no longer depends on the operator version. Support for multiple release streams. [#781](https://github.com/3scale/3scale-operator/pull/781)

### Fixed

- Product reconciler panic when backend was deleted from the UI [#743](https://github.com/3scale/3scale-operator/pull/743)
- Tenant cascade deletion [#747](https://github.com/3scale/3scale-operator/pull/747)
- Reqlogger panic that caused operator to crash [#748](https://github.com/3scale/3scale-operator/pull/748)
- Grant clusterversions.config.openshift.io resource add and list verbs [#745](https://github.com/3scale/3scale-operator/pull/748)
- Issues with database granular options management [#756](https://github.com/3scale/3scale-operator/pull/756) [#757](https://github.com/3scale/3scale-operator/pull/757)

## [0.9.0] - 2022-06-14

### Added

- Cluster scope install mode [#713](https://github.com/3scale/3scale-operator/pull/713)
- Tenant CRD: delete tenant [#715](https://github.com/3scale/3scale-operator/pull/715)
- APIManager CRD: reconcile apicast environment vars [#720](https://github.com/3scale/3scale-operator/pull/720)
- Tenant CRD: store access token instead of the provider key [#725](https://github.com/3scale/3scale-operator/pull/725)
- APIManager CRD: upgrade mysql to 8 [#682](https://github.com/3scale/3scale-operator/pull/682)
- Product CRD: delete product [#723](https://github.com/3scale/3scale-operator/pull/723)
- Backend CRD: delete backend [#723](https://github.com/3scale/3scale-operator/pull/723)
- Tenant ownership to products and backends [#727](https://github.com/3scale/3scale-operator/pull/727)
- APIManager CRD: external components optional [#733](https://github.com/3scale/3scale-operator/pull/733) [#737](https://github.com/3scale/3scale-operator/pull/737)

### Changed

- APIManager CRD: remove AMP_RELEASE from system env var set [#654](https://github.com/3scale/3scale-operator/pull/654)
- APIManager CRD: Cleanup Redis message bus references [#686](https://github.com/3scale/3scale-operator/pull/686)
- APIManager CRD: New metering labels [#679](https://github.com/3scale/3scale-operator/pull/679)

### Fixed

- OpenAPI CRD: disable openapi schema format validation [#712](https://github.com/3scale/3scale-operator/pull/712)
- APIManagerRestore CRD: fix controller [#761](https://github.com/3scale/3scale-operator/pull/761)

## [0.8.1] - 2021-12-13

### Added

- Enable operator metric service and servicemonitor [#667](https://github.com/3scale/3scale-operator/pull/667)
- APIManager CRD: Add proxy-related attributes to APIcast Staging and Production [#668](https://github.com/3scale/3scale-operator/pull/668)
- HTTP client respects http prxy env vars [#683](https://github.com/3scale/3scale-operator/pull/683)
- APIManager CRD: add x-kubernetes-preserve-unknown-fields to disable pruning [#683](https://github.com/3scale/3scale-operator/pull/683)

### Changed

- Delete kube-rbac-proxy container from controller-manager [#695](https://github.com/3scale/3scale-operator/pull/695)

## [0.8.0] - 2021-09-22

- APIManager CRD 3scale 2.11
- OpenAPI CRD [#496](https://github.com/3scale/3scale-operator/pull/496)
- Product CRD: policy chain [#523](https://github.com/3scale/3scale-operator/pull/523)
- Product CRD: oidc auth type [#531](https://github.com/3scale/3scale-operator/pull/531)
- ActiveDoc CRD [#539](https://github.com/3scale/3scale-operator/pull/539)
- CustomPolicyDefinition CRD [#546](https://github.com/3scale/3scale-operator/pull/546)
- APIManager CRD: Available condition [#549](https://github.com/3scale/3scale-operator/pull/549)
- CRD upgraded to v1 [#535](https://github.com/3scale/3scale-operator/pull/535)
- Account CRD [#551](https://github.com/3scale/3scale-operator/pull/551)
- APIManager CRD: Add configurable noreply FROM mail address [#566](https://github.com/3scale/3scale-operator/pull/566)
- APIManager CRD: Add System's USER_SESSION_TTL configurability [#621](https://github.com/3scale/3scale-operator/pull/621)
- Digest pinning [#640](https://github.com/3scale/3scale-operator/pull/640)
- APIManager CRD: Apicast custom policies [#645](https://github.com/3scale/3scale-operator/pull/645)
- Enable builds/ testing on ppc64le architecture [#646](https://github.com/3scale/3scale-operator/pull/646)
- Product CRD: Make application plans publication state configurable [#648](https://github.com/3scale/3scale-operator/pull/648)
- APIManager CRD: Add OpenTracing support for APIcast environments [#651](https://github.com/3scale/3scale-operator/pull/651)
- APIManager CRD: APIcast custom environments [#652](https://github.com/3scale/3scale-operator/pull/652)
- APIManager CRD: APIcast TLS at pod level [#653](https://github.com/3scale/3scale-operator/pull/653)

### Changed

- Operator-sdk upgrade to 1.2 [#514](https://github.com/3scale/3scale-operator/pull/514) [#516](https://github.com/3scale/3scale-operator/pull/516)
- APIManager CRD: remove secret ownership [#575](https://github.com/3scale/3scale-operator/pull/575)
- APIManager CRD: Increase default backend-cron DeploymentConfig memory limits [#592](https://github.com/3scale/3scale-operator/pull/592)
- APIManager CRD: Redis upgrade to ver 5 [#596](https://github.com/3scale/3scale-operator/pull/596)
- APIManager CRD: prometheus rules as optin feature [#622](https://github.com/3scale/3scale-operator/pull/622)

### Fixed

- APIManager CRD: system-app BACKEND_ROUTE env var from service endpoint [#536](https://github.com/3scale/3scale-operator/pull/536)
- CVE-2020-15257 [#559](https://github.com/3scale/3scale-operator/pull/559)
- CVE-2020-8912 [#603](https://github.com/3scale/3scale-operator/pull/603)

And many many small fixes. Check [PRs](https://github.com/3scale/3scale-operator/compare/3scale-2.10.0-GA...3scale-2.11.3-GA)

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

[Unreleased]: https://github.com/3scale/3scale-operator/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.9.0
[0.8.1]: https://github.com/3scale/3scale-operator/releases/tag/v0.8.1
[0.8.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.8.0
[0.7.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.7.0
[0.6.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.6.0
[0.5.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.5.0
[0.4.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.4.0
[0.3.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.3.0
[0.2.0]: https://github.com/3scale/3scale-operator/releases/tag/v0.2.0
