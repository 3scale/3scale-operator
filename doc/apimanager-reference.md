# APIManager CRD reference

This resource is used to deploy a 3scale API Management solution.

One APIManager custom resource per project is allowed.

## Table of Contents

<!--ts-->
* [APIManager CRD reference](#apimanager-crd-reference)
   * [Table of Contents](#table-of-contents)
   * [APIManager](#apimanager)
      * [APIManagerSpec](#apimanagerspec)
      * [APIManagerMetaData](#apimanagermetadata)
      * [ApicastSpec](#apicastspec)
      * [ApicastProductionSpec](#apicastproductionspec)
      * [ApicastStagingSpec](#apicaststagingspec)
      * [CustomPolicySpec](#custompolicyspec)
      * [CustomPolicySecret](#custompolicysecret)
      * [APIcastOpenTracingSpec](#apicastopentracingspec)
      * [OpenTelemetrySpec](#opentelemetryspec)
      * [APIcastTracingConfigSecret](#apicasttracingconfigsecret)
         * [CustomEnvironmentSpec](#customenvironmentspec)
         * [CustomEnvironmentSecret](#customenvironmentsecret)
      * [BackendSpec](#backendspec)
      * [BackendRedisPersistentVolumeClaimSpec](#backendredispersistentvolumeclaimspec)
      * [BackendListenerSpec](#backendlistenerspec)
      * [BackendWorkerSpec](#backendworkerspec)
      * [BackendCronSpec](#backendcronspec)
      * [SystemSpec](#systemspec)
      * [SystemRedisPersistentVolumeClaimSpec](#systemredispersistentvolumeclaimspec)
      * [FileStorageSpec](#filestoragespec)
      * [SystemPVCSpec](#systempvcspec)
      * [SystemS3Spec](#systems3spec)
      * [STSSpec](#stsspec)
      * [DeprecatedSystemS3Spec](#deprecatedsystems3spec)
      * [DatabaseSpec](#databasespec)
      * [MySQLSpec](#mysqlspec)
      * [SystemMySQLPVCSpec](#systemmysqlpvcspec)
      * [PostgreSQLSpec](#postgresqlspec)
      * [SystemPostgreSQLPVCSpec](#systempostgresqlpvcspec)
      * [SystemAppSpec](#systemappspec)
      * [SystemSidekiqSpec](#systemsidekiqspec)
      * [SystemSphinxSpec](#systemsphinxspec)
      * [SystemSearchdSpec](#systemsearchdspec)
      * [PVCGenericSpec](#pvcgenericspec)
      * [ZyncSpec](#zyncspec)
      * [ZyncAppSpec](#zyncappspec)
      * [ZyncQueSpec](#zyncquespec)
      * [HighAvailabilitySpec](#highavailabilityspec)
      * [ExternalComponentsSpec](#externalcomponentsspec)
      * [ExternalSystemComponents](#externalsystemcomponents)
      * [ExternalBackendComponents](#externalbackendcomponents)
      * [ExternalZyncComponents](#externalzynccomponents)
      * [PodDisruptionBudgetSpec](#poddisruptionbudgetspec)
      * [MonitoringSpec](#monitoringspec)
      * [APIManagerStatus](#apimanagerstatus)
         * [ConditionSpec](#conditionspec)
   * [PersistentVolumeClaimResourcesSpec](#persistentvolumeclaimresourcesspec)
   * [APIManager Secrets](#apimanager-secrets)
      * [backend-internal-api](#backend-internal-api)
      * [backend-listener](#backend-listener)
      * [backend-redis](#backend-redis)
      * [system-app](#system-app)
      * [system-database](#system-database)
      * [system-events-hook](#system-events-hook)
      * [system-master-apicast](#system-master-apicast)
      * [system-memcache](#system-memcache)
      * [system-recaptcha](#system-recaptcha)
      * [system-redis](#system-redis)
      * [system-seed](#system-seed)
      * [zync](#zync)
      * [fileStorage-S3-credentials-secret](#filestorage-s3-credentials-secret)
      * [system-smtp](#system-smtp)
   * [Default APIManager components compute resources](#default-apimanager-components-compute-resources)
<!--te-->

## APIManager

| **Field** | **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- | --- |
| Spec | `spec` | [APIManagerSpec](#APIManagerSpec) | Yes | The specfication for APIManager custom resource |
| Status | `status` | [APIManagerStatus](#APIManagerStatus) | No | The status for the custom resource  |
| MetaData | `metadata` | [APIManagerMetaData](#APIManagerMetaData) | No | The meta data for APIManager custom resource    |

### APIManagerSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| WildcardDomain | `wildcardDomain` | string | Yes | N/A | Root domain for the wildcard routes. Eg. example.com will generate 3scale-admin.example.com. |
| AppLabel | `appLabel` | string | No | `3scale-api-management` | The value of the `app` label that will be applied to the API management solution
| TenantName | `tenantName` | string | No | `3scale` | Tenant name under the root that Admin UI will be available with -admin suffix.
| ImagePullSecrets | `imagePullSecrets` | \[\][corev1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#localobjectreference-v1-core) | No | `[ { name: "threescale-registry-auth" } ]` | List of image pull secrets to be used on the managed Deployments ServiceAccounts. See [imagePullSecrets field in K8s ServiceAccount documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#serviceaccount-v1-core) for details on Image pull secrets. If not specified, `threescale-registry-auth` is used. Secret names that contain `dockercfg-` or `token-` anywhere in part of its name cannot be specified. If an update to this attribute is performed the corresponding Deployment pods have to be redeployed by the user to make the changes effective |
| ResourceRequirementsEnabled | `resourceRequirementsEnabled` | bool | No | `true` | When true, 3Scale API management solution is deployed with the optimal resource requirements and limits. Setting this to false removes those resource requirements. ***Warning*** Only set it to false for development and evaluation environments. When set to `true`, default compute resources are set for the APIManager components. See [Default APIManager components compute resources](#Default-APIManager-components-compute-resources) to see the default assigned values |
| ApicastSpec | `apicast` | \*ApicastSpec | No | See [ApicastSpec](#ApicastSpec) | Spec of the Apicast part |
| BackendSpec | `backend` | \*BackendSpec | No | See [BackendSpec](#BackendSpec) reference | Spec of the Backend part |
| SystemSpec  | `system`  | \*SystemSpec  | No | See [SystemSpec](#SystemSpec) reference | Spec of the System part |
| ZyncSpec    | `zync`    | \*ZyncSpec    | No | See [ZyncSpec](#ZyncSpec) reference | Spec of the Zync part    |
| HighAvailabilitySpec | `highAvailability` | \*HighAvailabilitySpec | No | **[DEPRECATED**] See [ExternalComponentsSpec](#ExternalComponentsSpec) reference | |
| ExternalComponentsSpec | `externalComponents` | \*ExternalComponentsSpec | No | See [ExternalComponentsSpec](#ExternalComponentsSpec) reference | Spec of the ExternalComponentsSpec part |
| PodDisruptionBudgetSpec | `podDisruptionBudget` | \*PodDisruptionBudgetSpec | No | See [PodDisruptionBudgetSpec](#PodDisruptionBudgetSpec) reference | Spec of the PodDisruptionBudgetSpec part |
| MonitoringSpec | `monitoring` | \*MonitoringSpec | No | Disabled | [MonitoringSpec](#MonitoringSpec) reference |

### APIManagerMetaData

| **Annotations**  | **Name** | **Default value** | **Description** |
| --- | --- | --- | --- |
| `apps.3scale.net/disable-apicast-service-reconciler` | disableApicastPortReconcile | `false` | Can be `true` or `false` - will disable apicast service port reconcile when true |

### ApicastSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| ApicastManagementAPI | `managementAPI` | string | No | `status` | Scope of the APIcast Management API. Can be disabled, status or debug. At least status required for health checks |
| OpenSSLVerify | `openSSLVerify` | bool | No | `false` | Turn on/off the OpenSSL peer verification when downloading the configuration |
| IncludeResponseCodes  | `responseCodes` | bool | No | `true` | Enable logging response codes in APIcast |
| RegistryURL | `registryURL` | string | No | `http://apicast-staging:8090/policies` | The URL to point to APIcast policies registry management |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for Apicast |
| ProductionSpec | `productionSpec` | \*ApicastProductionSpec | No | See [ApicastProductionSpec](#ApicastProductionSpec) reference | Spec of APIcast production part |
| StagingSpec | `stagingSpec` | \*ApicastStagingSpec | No | See [ApicastStagingSpec](#ApicastStagingSpec) reference | Spec of APIcast staging part |

### ApicastProductionSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `apicast-production` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| Workers | `workers` | integer | No | Automatically computed. Check [apicast doc](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_workers) for further info. | Defines the number of worker processes |
| LogLevel | `logLevel` | string | No | N/A | Log level for the OpenResty logs  (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_log_level)) |
| CustomPolicies | `customPolicies` | [][CustomPolicySpec](#CustomPolicySpec) | No | N/A | List of custom policies |
| OpenTracing | `openTracing` | [APIcastOpenTracingSpec](#APIcastOpenTracingSpec) | No | N/A | **[DEPRECATED]** Use `openTelementry` instead. Contains the OpenTracing integration configuration |
| OpenTelemetry | `openTelemetry` | [OpenTelemetrySpec](#OpenTelemetrySpec) | No | N/A | contains the OpenTelemetry integration configuration |
| CustomEnvironments | `customEnvironments` | [][CustomEnvironmentSpec](#CustomEnvironmentSpec) | No | N/A | List of custom environments |
| HTTPSPort | `httpsPort` | int | No | **8443** only when `httpsCertificateSecretRef` is provided | Controls on which port APIcast should start listening for HTTPS connections. Do not use `8080` as HTTPS port (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_https_port)) |
| HTTPSVerifyDepth | `httpsVerifyDepth` | int | No | N/A | Defines the maximum length of the client certificate chain. (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_https_verify_depth)) |
| HTTPSCertificateSecretRef | `httpsCertificateSecretRef` | LocalObjectReference | No | APIcast has a default certificate used when `httpsPort` is provided | References secret containing the X.509 certificate in the PEM format and the X.509 certificate secret key |
| AllProxy | `allProxy` | string | No | N/A | Specifies a HTTP(S) proxy to be used for connecting to services if a protocol-specific proxy is not specified. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#all_proxy-all_proxy)) |
| HTTPProxy | `httpProxy` | string | No | N/A | Specifies a HTTP(S) Proxy to be used for connecting to HTTP services. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#http_proxy-http_proxy)) |
| HTTPSProxy | `httpsProxy` | string | No | N/A | Specifies a HTTP(S) Proxy to be used for connecting to HTTPS services. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#https_proxy-https_proxy)) |
| NoProxy | `noProxy` | string | No | N/A | Specifies a comma-separated list of hostnames and domain names for which the requests should not be proxied. Setting to a single `*` character, which matches all hosts, effectively disables the proxy (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#no_proxy-no_proxy)) |
| ServiceCacheSize | `serviceCacheSize` | int | No | N/A | Specifies the number of services that APICast can store in the internal cache (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_service_cache_size)) |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                   | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))  |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No | `nil` | Specifies how to spread matching pods among the given topology |
| Labels | `labels` | map[string]string | No | `nil ` | Specifies labels that should be added to component |
| Annotations | `annotations` | map[string]string | No | `nil ` | Specifies Annotations that should be added to component |
| Hpa | `hpa` | bool | No | `nil` | Enables the horizontal pod autoscaling with default values |



### ApicastStagingSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `apicast-staging` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| LogLevel | `logLevel` | string | No | N/A | Log level for the OpenResty logs  (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_log_level)) |
| CustomPolicies | `customPolicies` | [][CustomPolicySpec](#CustomPolicySpec) | No | N/A | List of custom policies |
| OpenTracing | `openTracing` | [APIcastOpenTracingSpec](#APIcastOpenTracingSpec) | No | N/A | **[DEPRECATED]** Use `openTelementry` instead. Contains the OpenTracing integration configuration |
| OpenTelemetry | `openTelemetry` | [OpenTelemetrySpec](#OpenTelemetrySpec) | No | N/A | contains the OpenTelemetry integration configuration |
| CustomEnvironments | `customEnvironments` | [][CustomEnvironmentSpec](#CustomEnvironmentSpec) | No | N/A | List of custom environments |
| HTTPSPort | `httpsPort` | int | No | **8443** only when `httpsCertificateSecretRef` is provided | Controls on which port APIcast should start listening for HTTPS connections. Do not use `8080` as HTTPS port (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_https_port)) |
| HTTPSVerifyDepth | `httpsVerifyDepth` | int | No | N/A | Defines the maximum length of the client certificate chain. (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_https_verify_depth)) |
| HTTPSCertificateSecretRef | `httpsCertificateSecretRef` | LocalObjectReference | No | APIcast has a default certificate used when `httpsPort` is provided | References secret containing the X.509 certificate in the PEM format and the X.509 certificate secret key |
| AllProxy | `allProxy` | string | No | N/A | Specifies a HTTP(S) proxy to be used for connecting to services if a protocol-specific proxy is not specified. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#all_proxy-all_proxy)) |
| HTTPProxy | `httpProxy` | string | No | N/A | Specifies a HTTP(S) Proxy to be used for connecting to HTTP services. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#http_proxy-http_proxy)) |
| HTTPSProxy | `httpsProxy` | string | No | N/A | Specifies a HTTP(S) Proxy to be used for connecting to HTTPS services. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#https_proxy-https_proxy)) |
| NoProxy | `noProxy` | string | No | N/A | Specifies a comma-separated list of hostnames and domain names for which the requests should not be proxied. Setting to a single `*` character, which matches all hosts, effectively disables the proxy (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#no_proxy-no_proxy)) |
| ServiceCacheSize | `serviceCacheSize` | int | No | N/A | Specifies the number of services that APICast can store in the internal cache (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_service_cache_size)) |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                   | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### CustomPolicySpec

| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `name` | string | Yes | N/A | Name |
| `version` | string | Yes | N/A | Version |
| `secretRef` | LocalObjectReference | Yes | N/A | Secret reference with the policy content. See [CustomPolicySecret](#CustomPolicySecret) for more information.

### CustomPolicySecret

Contains custom policy specific content. Two files,  `init.lua` and `apicast-policy.json`, are required, but more can be added optionally.

Some examples are available [here](/doc/adding-custom-policies.md)

| **Field** | **Description** |
| --- | --- |
| `init.lua` | Custom policy lua code entry point |
| `apicast-policy.json` | Custom policy metadata |

### APIcastOpenTracingSpec
| **Field** | **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Enabled | `enabled` | bool | No | `false` | Controls whether OpenTracing integration with APIcast is enabled. By default it is not enabled |
| TracingLibrary | `tracingLibrary` | string | No | `jaeger` | Controls which OpenTracing library is loaded. At the moment the supported values are: `jaeger`. If not set, `jaeger` will be used |
| TracingConfigSecretRef | `tracingConfigSecretRef` | LocalObjectReference | No | tracing library-specific default | Secret reference with the tracing library-specific configuration. Each supported tracing library provides a default configuration file which is used if `tracingConfigSecretRef` is not specified. See [APIcastTracingConfigSecret](#APIcastTracingConfigSecret) for more information. |

### OpenTelemetrySpec
| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `enabled` | bool | No | `false` | Controls whether opentelemetry based gateway instrumentation is enabled or not. By default it is **disabled** |
| `tracingConfigSecretRef` | *LocalObjectReference* | No | None | Secret reference with the [opentelemetry tracing configuration](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx). |
| `tracingConfigSecretKey` | string | No | If unspecified, the first secret key in lexicographical order will be referenced as tracing configuration | The secret key used as tracing configuration |

**Watch for secret changes**

By default, content changes in the secret will not be noticed by the 3scale operator.
The 3scale operator allows monitoring the secret for changes adding the `apimanager.apps.3scale.net/watched-by=apimanager` label.
With that label in place, when the content of the secret is changed, the operator will get notified.
Then, the operator will rollout apicast deployment to make the changes effective.
The operator will not take *ownership* of the secret in any way.

### APIcastTracingConfigSecret

| **Field** | **Description** |
| --- | --- |
| `config` | Tracing library-specific configuration |

*NOTE*: Once apicast has been deployed, the content of the secret should not be updated externally.
If the content of the secret is updated externally, after apicast has been deployed, the container can automatically see the changes.
However, apicast has the environment already loaded and it does not change the behavior.

If the custom environment content needs to be changed, there are two options:

* [**recommended way**] Create another secret with a different name and update the APIcast custom resource field `spec.apicast.<apicast-environment>.openTracing.tracingConfigSecretRef.name`. The operator will trigger a rolling update loading the new custom environment content.
* Update the existing secret content and redeploy apicast turning `spec.replicas` to 0 and then back to the previous value.

#### CustomEnvironmentSpec

| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `secretRef` | LocalObjectReference | Yes | N/A | Secret reference with the custom environment content. See [CustomEnvironmentSecret](#CustomEnvironmentSecret) for more information.

#### CustomEnvironmentSecret

Generic (`opaque`) type secret holding one or more custom environments.
The operator will load in the APIcast container all the files (keys) found in the secret.

Some examples are available [here](/doc/adding-apicast-custom-environments.md)

| **Field** | **Description** |
| --- | --- |
| *filename* | Custom environment lua code |


### BackendSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for Backend |
| RedisImage | `redisImage` | string | No | nil | Used to overwrite the desired Redis image for the Redis used by backend. Only takes effect when redis is not managed externally |
| RedisAffinity | `redisAffinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Only takes effect when redis is not managed externally |
| RedisTolerations | `redisTolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Only takes effect when redis is not managed externally |
| RedisResources | `redisResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | RedisResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| RedisPersistentVolumeClaimSpec | `redisPersistentVolumeClaim` | \*[BackendRedisPersistentVolumeClaimSpec](#BackendRedisPersistentVolumeClaimSpec) | No | nil | Backend's Redis PersistentVolumeClaim configuration options. Only takes effect when redis is not managed externally |
| ListenerSpec | `listenerSpec` | \*BackendListenerSpec | No | See [BackendListenerSpec](#BackendListenerSpec) reference | Spec of Backend Listener part |
| WorkerSpec | `workerSpec` | \*BackendWorkerSpec | No | See [BackendWorkerSpec](#BackendWorkerSpec) reference | Spec of Backend Worker part |
| CronSpec | `cronSpec` | \*BackendCronSpec | No | See [BackendCronSpec](#BackendCronSpec) reference | Spec of Backend Cron part |
| RedisPriorityClassName         | `redisPriorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| RedisTopologySpreadConstraints | `redisTopologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| RedisLabels                    | `redisLabels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| RedisAnnotations                    | `redisAnnotations`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies Annotations that should be added to component   |

### BackendRedisPersistentVolumeClaimSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |

### BackendListenerSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `backend-listener` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName | `priorityClassName` | string | No | N/A | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No | `nil` | Specifies how to spread matching pods among the given topology |
| Labels | `labels` | map[string]string | No | `nil ` | Specifies labels that should be added to component |
| Annotations | `annotations` | map[string]string | No | `nil ` | Specifies Annotations that should be added to component |
| HpaSpec | `hpa` | bool | No | `nil` | Enables the horizontal pod autoscaling with default values |


### BackendWorkerSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `backend-worker` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName | `priorityClassName` | string | No | N/A | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No | `nil` | Specifies how to spread matching pods among the given topology | 
| Labels | `labels` | map[string]string | No | `nil ` | Specifies labels that should be added to component |
| Annotations | `annotations` | map[string]string | No | `nil ` | Specifies Annotations that should be added to component |
| HpaSpec | `hpa` | bool | No | `nil` | Enable the horizontal pod autoscaling with default values |

### BackendCronSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `backend-cron` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### SystemSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for System |
| RedisImage | `redisImage` | string | No | nil | Used to overwrite the desired Redis image for the Redis used by System. Only takes effect when redis is not managed externally |
| RedisPersistentVolumeClaimSpec | `redisPersistentVolumeClaim` | \*[SystemRedisPersistentVolumeClaimSpec](#SystemRedisPersistentVolumeClaimSpec) | No | nil | System's Redis PersistentVolumeClaim configuration options. Only takes effect when redis is not managed externally |
| RedisAffinity | `redisAffinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Only takes effect when redis is not managed externally |
| RedisTolerations | `redisTolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Only takes effect when redis is not managed externally |
| RedisResources | `redisResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | RedisResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| MemcachedImage | `memcachedImage` | string | No | nil | Used to overwrite the desired Memcached image for the Memcached used by System |
| MemcachedAffinity | `memcachedAffinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| MemcachedTolerations | `memcachedTolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| MemcachedResources | `memcachedResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | MemcachedResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| FileStorageSpec | `fileStorage` | \*SystemFileStorageSpec | No | See [FileStorageSpec](#FileStorageSpec) specification | Spec of the System's File Storage part |
| DatabaseSpec | `database` | \*SystemDatabaseSpec | No | See [DatabaseSpec](#DatabaseSpec) specification | Spec of the System's Database part |
| AppSpec | `appSpec` | \*SystemAppSpec | No | See [SystemAppSpec](#SystemAppSpec) reference | Spec of System App part |
| SidekiqSpec | `sidekiqSpec` | \*SystemSidekiqSpec | No | See [SystemSidekiqSpec](#SystemSidekiqSpec) reference | Spec of System Sidekiq part |
| SphinxSpec | `sphinxSpec` | \*SystemSphinxSpex | No | **DEPRECATED** Use `SearchdSpec` instead. See [SystemSphinxSpec](#SystemSphinxSpec) reference | Spec of System's Sphinx part |
| SearchdSpec | `searchdSpec` | [SystemSearchdSpec](#SystemSearchdSpec) | No | See [SystemSearchdSpec](#SystemSearchdSpec) reference | Spec of System's Searchd component |
| MemcachedPriorityClassName | `memcachedPriorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| MemcachedTopologySpreadConstraints | `memcachedTopologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| MemcachedLabels                    | `memcachedLabels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| RedisPriorityClassName             | `redisPriorityClassName`             | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| RedisTopologySpreadConstraints     | `redisTopologySpreadConstraints`     | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| RedisLabels                        | `redisLabels`                        | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| RedisAnnotations          | `redisAnnotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### SystemRedisPersistentVolumeClaimSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |

### FileStorageSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| PVC | `persistentVolumeClaim` | [PVCGenericSpec](#PVCGenericSpec) | No | nil | PersistentVolumeClaim spec for the System shared storage |
| DeprecatedS3  | `amazonSimpleStorageService` | \*DeprecatedSystemS3Spec | No | nil | DEPRECATED [DeprecatedSystemS3Spec](#DeprecatedSystemS3Spec). Used to use S3 as the System's file storage. See [SystemS3Spec](#SystemS3Spec) |
| S3  | `simpleStorageService` | \*SystemS3Spec | No | nil | Used to use S3 as the System's file storage. See [SystemS3Spec](#SystemS3Spec) |

Only one of the fields can be chosen. If no field is specified then PVC is used.

### SystemPVCSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |
| Resources | `resources` | [PersistentVolumeClaimResourcesSpec](#PersistentVolumeClaimResourcesSpec) | No | nil | The minimum resources the volume should have. Resources will not take any effect when VolumeName is provided. This parameter is not updateable when the underlying PV is not resizable. |
| VolumeName | `volumeName` | string | No | nil | The binding reference to the existing PersistentVolume backing this claim |

### SystemS3Spec

| **Field**     | **json/yaml field**      | **Type**                                                                                                                         | **Required** | **Default value** | **Description**                                                                                                                                                                                                                                                                 |
|---------------|--------------------------|----------------------------------------------------------------------------------------------------------------------------------|--------------|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Configuration | `configurationSecretRef` | [corev1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#localobjectreference-v1-core) | Yes          | N/A               | Local object reference to the secret to be used where the AWS configuration is stored. See [LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#localobjectreference-v1-core) on how to specify the local object reference to the secret |
| STS           | `sts`                    | [STSSpec](#STSSpec) | No | IAM S3 authentication | STS spec object |

The secret name specified in the `configurationSecretRef` field must be
pre-created by the user before creating the APIManager custom resource.
Otherwise the operator will complain about it. See the
[fileStorage S3 credentials secret](#fileStorage-S3-credentials-secret)
specification to see what fields the secret should have and the values
that should be set on it.

### STSSpec

| *yaml Field* | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `enabled` | bool | No | `false` | Enable Secure Token Service for  short-term, limited-privilege security credentials |
| `audience` | string | No | `openshift` | The ID the token is intended for. This field does not have any effect when STS is not enabled. |


### DeprecatedSystemS3Spec
**DEPRECATED** Setting fields here has no effect. Use [SystemS3Spec](#SystemS3Spec) instead

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| AWSBucket | `awsBucket` | string | Yes | N/A | **DEPRECATED** Use [SystemS3Spec](#SystemS3Spec) instead |
| AWSRegion | `awsRegion` | string | Yes | N/A | **DEPRECATED** Use [SystemS3Spec](#SystemS3Spec) instead |
| AWSCredentials | `awsCredentialsSecret` | **DEPRECATED** Use [SystemS3Spec](#SystemS3Spec) instead |

### DatabaseSpec

Note: Deploying databases internally with this section is meant for evaluation purposes. Check [ExternalComponentsSpec](#ExternalComponentsSpec) for production ready recommended deployments.

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| MySQL | `mysql`| \*SystemMySQLSpec | No | nil | Enable MySQL database as System's database. Only takes effect when the instance is not managed externally. See [MySQLSpec](#MySQLSpec) specification |
| PostgreSQL | `postgresql` | \*SystemPostgreSQLSpec | No | nil | Enable PostgreSQL database as System's database. Only takes effect when the instance is not managed externally. See [PostgreSQLSpec](#PostgreSQLSpec)

### MySQLSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for System's MySQL database |
| PersistentVolumeClaimSpec | `persistentVolumeClaim` | [PVCGenericSpec](#PVCGenericSpec) | No | nil | System's MySQL PersistentVolumeClaim configuration options |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### SystemMySQLPVCSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |
| Resources | `resources` | [PersistentVolumeClaimResourcesSpec](#PersistentVolumeClaimResourcesSpec) | No | nil | The minimum resources the volume should have. Resources will not take any effect when VolumeName is provided. This parameter is not updateable when the underlying PV is not resizable. |
| VolumeName | `volumeName` | string | No | nil | The binding reference to the existing PersistentVolume backing this claim |

### PostgreSQLSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for System's PostgreSQL database |
| PersistentVolumeClaimSpec | `persistentVolumeClaim` | [PVCGenericSpec](#PVCGenericSpec) | No | nil | System's PostgreSQL PersistentVolumeClaim configuration options |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |


### SystemPostgreSQLPVCSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |
| Resources | `resources` | [PersistentVolumeClaimResourcesSpec](#PersistentVolumeClaimResourcesSpec) | No | nil | The minimum resources the volume should have. Resources will not take any effect when VolumeName is provided. This parameter is not updateable when the underlying PV is not resizable. |
| VolumeName | `volumeName` | string | No | nil | The binding reference to the existing PersistentVolume backing this claim |

### SystemAppSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `system-app` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| MasterContainerResources | `masterContainerResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| ProviderContainerResources | `providerContainerResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| DeveloperContainerResources | `developerContainerResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### SystemSidekiqSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `system-sidekiq` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### SystemSphinxSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |

### SystemSearchdSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired custom image for the Searchd server used by System |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PVC | `persistentVolumeClaim` | [PVCGenericSpec](#PVCGenericSpec) | No | nil | PersistentVolumeClaim spec for the System Searchd component |
| PriorityClassName         | `priorityClassName`         | string | No           | N/A | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No | `nil` | Specifies how to spread matching pods among the given topology |
| Labels | `labels` | map[string]string | No | `nil ` | Specifies labels that should be added to component |
| Annotations | `annotations` | map[string]string  | No | `nil ` | Specifies Annotations that should be added to component |


### PVCGenericSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |
| Resources | `resources` | [PersistentVolumeClaimResourcesSpec](#PersistentVolumeClaimResourcesSpec) | No | nil | The minimum resources the volume should have. Resources will not take any effect when VolumeName is provided. This parameter is not updateable when the underlying PV is not resizable. |
| VolumeName | `volumeName` | string | No | nil | The binding reference to the existing PersistentVolume backing this claim |

### ZyncSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for Zync |
| PostgreSQLImage | `postgreSQLImage` | string | No | nil | Used to overwrite the desired PostgreSQL image for the PostgreSQL used by Zync. Does not take effect when the database is managed externally |
| AppSpec | `appSpec` | \*ZyncAppSpec | No | See [ZyncAppSpec](#ZyncAppSpec) reference | Spec of Zync App part |
| QueSpec | `queSpec` | \*ZyncQueSpec | No | See [ZyncQueSpec](#ZyncQueSpec) reference | Spec of Zync Que part |
| DatabaseAffinity | `databaseAffinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Does not take effect when the database is managed externally |
| DatabaseTolerations | `databaseTolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Does not take effect when the database is managed externally |
| DatabaseResources | `databaseResources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | DatabaseResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior. Does not take effect when the database is managed externally |
| DatabasePriorityClassName         | `databasePriorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| DatabaseTopologySpreadConstraints | `databaseTopologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No | `nil` | Specifies how to spread matching pods among the given topology |
| DatabaseLabels | `databaseLabels` | map[string]string | No | `nil ` | Specifies labels that should be added to component |
| DatabaseAnnotations | `databaseAnnotations` | map[string]string  | No | `nil ` | Specifies Annotations that should be added to component |

### ZyncAppSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `zync` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName | `priorityClassName` | string | No | N/A | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)) |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No | `nil` | Specifies how to spread matching pods among the given topology |
| Labels | `labels` | map[string]string | No | `nil ` | Specifies labels that should be added to component |
| Annotations | `annotations` | map[string]string  | No | `nil ` | Specifies Annotations that should be added to component |

### ZyncQueSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `zync-que` deployment |
| Affinity | `affinity` | [v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| PriorityClassName         | `priorityClassName`         | string                                                                                                                                    | No           | N/A                                                                                                                                            | If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. (see [docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/))                                                                                                                                                                                                                                                                              |
| TopologySpreadConstraints | `topologySpreadConstraints` | \[\][v1.TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#topologyspreadconstraint-v1-core) | No           | `nil`                                                                                                                                          | Specifies how to spread matching pods among the given topology                                                                                                                                                                                                                                          |
| Labels                    | `labels`                    | map[string]string                                                                                                                        | No           | `nil `                                                                                                                                         | Specifies labels that should be added to component                                                                                                                                                                                                                                                                                   |
| Annotations          | `annotations`                    | map[string]string  | No           | `nil `  | Specifies Annotations that should be added to component   |

### HighAvailabilitySpec

[**DEPRECATED**] See [ExternalComponentsSpec](#ExternalComponentsSpec) reference

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Enabled | `enabled` | bool | No | `false` | Enable to use external system database, backend redis, and system redis databases|
| ExternalZyncDatabaseEnabled | `externalZyncDatabaseEnabled` | bool | No | `false` | Enable to user external zync database. The value of this field only takes effect when `spec.highAvailability.enabled` is set to `true` |

When HighAvailability is enabled the following secrets have to be pre-created by the user:

* [backend-redis](#backend-redis) with the `REDIS_STORAGE_URL` and
  `REDIS_QUEUES_URL` fields with values pointing to the desired external
  databases. The databases should be configured
  in high-availability mode
* [system-database](#system-database) with the `URL` field with the value
  pointing to the desired external database. The database should be configured
  in high-availability mode
* [system-redis](#system-redis) with the `URL` field
  with the value pointing to the desired external database. The database
  should be configured in high-availability mode

Additionally, when HighAvailability is enabled, if the `externalZyncDatabaseEnabled` field is
also enabled the user has to pre-create the following secret too:
* [zync](#zync) with the `DATABASE_URL` and `DATABASE_PASSWORD` fields
  with the values pointing to the desired external database settings.
  The database should be configured in high-availability mode

Use of slaves for Internal Redis is not supported.

### ExternalComponentsSpec

| **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- |
| `system` | \*[ExternalSystemComponents](#ExternalSystemComponents) | No | Use external system databases |
| `backend` | \*[ExternalBackendComponents](#ExternalBackendComponents) | No | Use external backend databases |
| `zync` | \*[ExternalZyncComponents](#ExternalZyncComponents) | No | Use external zync databases |

### ExternalSystemComponents

| **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- |
| `redis` | `bool` | No | Use external redis databases. Defaults to `false` |
| `database` | `bool` | No | Use external RDBMS database. Defaults to `false` |

When system `redis` is enabled the following secret has to be pre-created by the user:

* [system-redis](#system-redis) with the `URL` field
  with the value pointing to the desired external database.

When system `database` is enabled the following secret has to be pre-created by the user:

* [system-database](#system-database) with the `URL` field with the value
  pointing to the desired external database.

### ExternalBackendComponents

| **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- |
| `redis` | `bool` | No | Use external redis databases. Defaults to `false` |

When backend `redis` is enabled the following secret has to be pre-created by the user:

* [backend-redis](#backend-redis) with the `REDIS_STORAGE_URL` and
  `REDIS_QUEUES_URL` fields with values pointing to the desired external
  databases.

### ExternalZyncComponents

| **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- |
| `database` | `bool` | No | Use external RDBMS database. Defaults to `false` |

When zync `database` is enabled the following secret has to be pre-created by the user:

* [zync](#zync) with the `DATABASE_URL` and `DATABASE_PASSWORD` fields
  with the values pointing to the desired external database settings.

### PodDisruptionBudgetSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Enabled | `enabled` | bool | No | `false` | Enable to automatically create [PodDisruptionBudgets](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/) for components that can scale. Not including any of the databases or redis services.|

### MonitoringSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Enabled | `enabled` | bool | No | `false` | [Enable to automatically create monitoring resources](operator-monitoring-resources.md) |
| EnablePrometheusRules | `enablePrometheusRules` | bool | No | `true` | Activate/Disable *PrometheusRules* deployment |

### APIManagerStatus

Used by the Operator/Kubernetes to control the state of the APIManager.
an `APIManager` status field should never be modified by the user.

| **Field** | **json/yaml field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Available | `available` | v1.Condition | Indicates whether the APIManager is in `Available` state. See [ConditionSpec](#ConditionSpec) for a description on the meaning of `Available`|

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string indicating the type of the condition. The types are:
  * `Available`: An APIManager is in `Available` state when *all* of the following scenarios are true:
    * All expected Deployments to be deployed exist and have the `Available` condition set to true
    * All 3scale default OpenShift routes exist and have the Admitted condition set to true. The default routes are:
      * Master route
      * Backend Listener route
      * Default tenant admin route, developer route, APIcast staging and production routes beloinging to the default tenant


| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |



## PersistentVolumeClaimResourcesSpec

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `requests` | [v1 Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#quantity-resource-core) | Yes | N/A | Requested size of the PersistentVolumeClaim. |

## APIManager Secrets

Additionally, if desired, several sensitive APIManager configuration options
can be controlled by pre-creating some Kubernetes secrets before deploying the
APIManager Custom Resource.

The available configurable secrets are:

### backend-internal-api

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| username | Backend internal API username. Backend internal API is used by System | `3scale_api_user` |
| password | Backend internal API password. Backend internal API is used by System | Autogenerated value |

### backend-listener

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| service_endpoint | Backend listener service endpoint. Used by System and Apicast | `http://backend-listener:3000` |
| route_endpoint | Backend listener route endpoint. Used by System | `https://backend-<tenantName>.<wildcardDomain>` |

### backend-redis

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| REDIS_STORAGE_URL | Backend's redis storage database URL. | Mandatory when the instance is managed externally. Otherwise the default value is: `redis://backend-redis:6379/0` |
| REDIS_STORAGE_SENTINEL_ROLE | Backend's redis storage sentinel role name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| REDIS_STORAGE_SENTINEL_HOSTS | Backend's redis storage sentinel hosts name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| REDIS_QUEUES_URL | Backend's redis queues database URL  | Mandatory when the instance is managed externally. Otherwise the default value is: `redis://backend-redis:6379/1` |
| REDIS_QUEUES_SENTINEL_ROLE | Backend's redis queues sentinel role name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| REDIS_QUEUES_SENTINEL_HOSTS | Backend's redis queues sentinel hosts name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| CONFIG_REDIS_CA_FILE | Backend's redis file that contains the Certificate Authority (CA) certificate| `""`                                                                                                              |
| CONFIG_REDIS_CERT | Backend's redis certificate | `""`                                                                                                              |
| CONFIG_REDIS_PRIVATE_KEY | Backend's redis private key used for authentication in SSL/TLS communication| `""`                                                                                                              |
| CONFIG_REDIS_SSL | This field is "true" if any of the other fields (CONFIG_REDIS_CA_FILE, CONFIG_REDIS_CERT, CONFIG_REDIS_PRIVATE_KEY) are not empty. Otherwise it is "false"| `false`                                                                                                           |
| CONFIG_QUEUES_CA_FILE | Backend's redis file with configuration setting for the CA certificate that is used for secure communications in the context of a Redis queueing. | `""`                                                                                                              |
| CONFIG_QUEUES_CERT | Backend's redis certificate used for establishing secure connections in a Redis queuing | `""`                                                                                                              |
| CONFIG_QUEUES_PRIVATE_KEY | Backend's redis private key used for establishing secure connections in the context of Redis queuing. | `""`                                                                                                              |
| CONFIG_QUEUES_SSL | This field is "true" if any of the other fields (CONFIG_QUEUES_CA_FILE, CONFIG_QUEUES_CERT, CONFIG_QUEUES_PRIVATE_KEY) are not empty. Otherwise it is "false"|`""`| `false`|

### system-app

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| SECRET_KEY_BASE | System application secret key base | Autogenerated value |
| USER_SESSION_TTL | System user's login session length (TTL) in seconds. The empty value means 2 weeks | `""` |

### system-database

For Mysql:

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | URL of the Porta database. Format: `mysql2://<AdminUser>:<AdminPassword>@<DatabaseHost>/<DatabaseName>`, where `<AdminUser>` must be an already existing user in the external database with full permissions on the specified `<DatabaseName>` logical database and `<DatabaseName>` must be an already existing logical database in the external database.| Mandatory when the instance is managed externally. A default is only set when database is managed internally.<br/>When managed internally:<br/>`mysql2://root:<AutogeneratedValue>@system-mysql/mysql`.|
| DB_USER | Not used by 3scale components. Only used when the database is managed internally to create a new user granted with superuser permissions for the database specified in the `URL` field. | `mysql` |
| DB_PASSWORD | Not used by 3scale components. Only used when the database is managed internally to create a new user granted with superuser permissions for the database specified in the `URL` field. | Autogenerated value |

For Postgresql:

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | URL of the Porta database. Format: `postgresql://<AdminUser>:<AdminPassword>@<DatabaseHost>/<DatabaseName>`, where `<AdminUser>` must be an already existing user in the external database with full permissions on the specified `<DatabaseName>` logical database and `<DatabaseName>` must be an already existing logical database in the external database.| Mandatory when the instance is managed externally. A default is only set when database is managed internally.<br/>When managed internally:<br/>`postgresql://system:<AutoGeneratedValue>@system-postgresql/system`.|
| DB_USER | Not used by 3scale components. Only used when the database is managed internally to create a user with superuser power. | `system` |
| DB_PASSWORD | Not used by 3scale components. Only used when the database is managed internally to create a user with superuser power. | Autogenerated value |

For Oracle:

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | URL of the Porta database. Mandatory for Oracle based deployments. Format: `oracle-enhanced://<AdminUser>:<AdminPassword>@<DatabaseHost>/<DatabaseName>`, where `<AdminUser>` must be an already existing user in the external database with full permissions on the specified `<DatabaseName>` logical database and `<DatabaseName>` must be an already existing logical database in the external database.| - |
| ORACLE_SYSTEM_PASSWORD | Password of Oracle's `SYSTEM` administrative user. Required and only used when system's database provided in `URL` field is an external Oracle database | N/A |

### system-events-hook

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | The URL to System's event reports endpoint, used by Backend to report its events | `http://system-master:3000/master/events/import` |
| PASSWORD | Shared secret to import events from backend to system | Autogenerated value |

### system-master-apicast

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| ACCESS_TOKEN | Read only access token that APIcast uses to download its configuration from System | Autogenerated value |
| BASE_URL | URL of the 3scale portal admin endpoint with authentication part | `http://<ACCESS_TOKEN>@system-master:3000` |
| PROXY_CONFIGS_ENDPOINT | URL of the available configs for all System's services | `http://<ACCESS_TOKEN>@system-master:3000/master/api/proxy/configs` |

### system-memcache

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| SERVERS | System's Memcached URL | `system-memcache:11211` |

### system-recaptcha

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| PUBLIC_KEY | reCAPTCHA site key (used in spam protection) for System| `""` |
| PRIVATE_KEY | reCAPTCHA secret key (used in spam protection) for System| `""` |

### system-redis

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | System's Redis database URL | Mandatory when instance is managed externally. Otherwise the default value is: `redis://system-redis:6379/1` |
| NAMESPACE | Define the namespace to be used by System's Redis Database. The empty value means not namespaced | `""` |
| SENTINEL_HOSTS | System's Redis sentinel hosts. Used only when Redis sentinel is configured | `""` |
| SENTINEL_ROLE | System's Redis sentinel role name. Used only when Redis sentinel is configured | `""` |
| REDIS_CA_FILE | System's redis file that contains the Certificate Authority (CA) certificate | `""` |
| REDIS_CLIENT_CERT | System's Redis Client certificate | `""`|
| REDIS_PRIVATE_KEY | System's redis private key used for authentication in SSL/TLS communication | `""`|
| REDIS_SSL | This field is "true" if any of the other fields (REDIS_CA_FILE, REDIS_CLIENT_CERT, REDIS_PRIVATE_KEY) are not emp``ty. Otherwise it is "false"| `false`|

### system-seed

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| MASTER_USER | System's master username | `master` |
| MASTER_PASSWORD | System's master password | Autogenerated value |
| MASTER_ACCESS_TOKEN | System's master access token | Autogenerated value |
| MASTER_DOMAIN | System's master domain name | `master` |
| ADMIN_USER | System's admin user of the tenant created by default | `admin` |
| ADMIN_PASSWORD | System's admin password of the tenant created by default | Autogenerated value |
| ADMIN_ACCESS_TOKEN | System's admin access token of the tenant created by default | Autogenerated value |
| TENANT_NAME | Tenant name under the root that Admin UI will be available with -admin suffix | `<tenantName>` |

### zync

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| DATABASE_URL | PostgreSQL database used by Zync. Only used when the database is managed externally | Format: `postgresql://<zync-db-username>:<ZYNC_DATABASE_PASSWORD>@<zync-db-host>:<zync-db-port>/zync_production`, where `<zync-db-username>` must be an already existing user in the external database with full permissions on the `zync_production` logical database, `zync_production` logical database must be an already existing logical database in the external database and the specified value of `<ZYNC_DATABASE_PASSWORD>` must be the same as the `ZYNC_DATABASE_PASSWORD` parameter in this secret. Otherwise it has a default value, which is `postgresql://zync:<ZYNC_DATABASE_PASSWORD>@zync-database:5432/zync_production` |
| ZYNC_DATABASE_PASSWORD | Database password associated to the user specified in the `DATABASE_URL` parameter | When the database is managed externally, this parameter is mandatory and must have the same value as the password part of the `DATABASE_URL` parameter in this secret. Otherwise the default value is an autogenerated value if not defined |
| SECRET_KEY_BASE | Zync's application key generator to encrypt communications | Autogenerated value |
| ZYNC_AUTHENTICATION_TOKEN | Authentication token used to authenticate System when calling Zync | Autogenerated value |

### fileStorage-S3-credentials-secret

The name of this secret can be any name as long as does not collide with other
existing secret names.

| **Field** | **Description** | **Required for IAM** | **Required for STS** |
| --- | --- | --- |----------------------|
| AWS_ACCESS_KEY_ID | AWS Access Key ID to use in S3 Storage for System's file storage | Y | N                    |
| AWS_SECRET_ACCESS_KEY | AWS Access Key Secret to use in S3 Storage for System's file storage | Y | N                    |
| AWS_BUCKET | S3 bucket to be used as System's FileStorage for assets | Y | Y                    |
| AWS_REGION | Region of the S3 bucket to be used as System's FileStorage for assets | Y | Y                    |
| AWS_HOSTNAME | Default: Amazon endpoints - AWS S3 compatible provider endpoint hostname | N | N                    |
| AWS_PROTOCOL | Default: HTTPS - AWS S3 compatible provider endpoint protocol | N | N                    |
| AWS_PATH_STYLE | Default: false - When set to true, the bucket name is always left in the request URI and never moved to the host as a sub-domain | N | N                    |
| AWS_ROLE_ARN | ARN of the Role which has a policy attached to authenticate using AWS STS | N | Y                    |
| AWS_WEB_IDENTITY_TOKEN_FILE | Path to mounted token file location e.g. /var/run/secrets/openshift/serviceaccount/token | N | Y                    |

### system-smtp

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| address | Address (hostname or IP) of the remote mail server to use. If set to a value different than `""` System will use the mail server to send mails related to events that happen in the API management solution |  `""` |
| port | Port of the remote mail server to use | `""` |
| domain | In case the mail server requires a HELO domain | `""` |
| authentication | In case the mail server requires authentication, set this setting to the authentication type here. `plain` to send the password in the clear, `login` to send password Base64 encoded, or `cram_md5` to combine a Challenge/Response mechanism based on the HMAC-MD5 algorithm | `""` |
| username | In case the mail server requires authentication and the authentication type requires it | `""` |
| password | In case the mail server requires authentication and the authentication type requires it | `""` |
| openssl.verify.mode | When using TLS, you can set how OpenSSL checks the certificate. This is really useful if you need to validate a self-signed and/or a wildcard certificate. You can use the name of an OpenSSL verify constant: `none` or `peer` | `""` |
| from_address | `from` address value for the no-reply mail | `""` |

## Default APIManager components compute resources

When APIManager's `spec.resourceRequirementsEnabled` attribute is set to
`true`, default compute resources are set for the APIManager components.

The specific compute resources default values that are set for the
APIManager components are the following ones:


| **Component** | **CPU Requests**| **CPU Limits** | **Memory Requests** | **Memory Limits** |
| --- | --- | --- | --- | --- |
| system-app's system-master | 50m | 1000m | 600Mi | 800Mi |
| system-app's system-provider | 50m | 1000m | 600Mi | 800Mi |
| system-app's system-developer | 50m | 1000m | 600Mi | 800Mi |
| system-sidekiq | 100m | 1000m  | 500Mi  | 2Gi |
| system-searchd | 80m | 1000m | 250Mi | 512Mi |
| system-redis | 150m | 500m | 256Mi | 32Gi |
| system-mysql | 250m | No limit | 512Mi | 2Gi |
| system-postgresql | 250m | No limit | 512Mi | 2Gi |
| backend-listener | 500m | 1000m | 550Mi | 700Mi |
| backend-worker | 150m | 1000m | 50Mi | 300Mi |
| backend-cron | 100m | 500m | 100Mi | 500Mi |
| backend-redis | 1000m | 2000m | 1024Mi | 32Gi |
| apicast-production | 500m | 1000m | 64Mi | 128Mi |
| apicast-staging | 50m | 100m | 64Mi | 128Mi |
| zync | 150m | 1 | 250M | 512Mi |
| zync-que | 250m | 1 | 250M | 512Mi |
| zync-database | 50m | 250m | 250M | 2G |
