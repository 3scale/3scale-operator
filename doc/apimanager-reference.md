# APIManager CRD reference

This resource is used to deploy a 3scale API Management solution.

One APIManager custom resource per project is allowed.

## Table of Contents

* [APIManager](#apimanager)
   * [APIManagerSpec](#apimanagerspec)
   * [ApicastSpec](#apicastspec)
   * [ApicastProductionSpec](#apicastproductionspec)
   * [ApicastStagingSpec](#apicaststagingspec)
   * [CustomPolicySpec](#custompolicyspec)
   * [CustomPolicySecret](#custompolicysecret)
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
   * [DeprecatedSystemS3Spec](#deprecatedsystems3spec)
   * [DatabaseSpec](#databasespec)
   * [MySQLSpec](#mysqlspec)
   * [SystemMySQLPVCSpec](#systemmysqlpvcspec)
   * [PostgreSQLSpec](#postgresqlspec)
   * [SystemPostgreSQLPVCSpec](#systempostgresqlpvcspec)
   * [SystemAppSpec](#systemappspec)
   * [SystemSidekiqSpec](#systemsidekiqspec)
   * [SystemSphinxSpec](#systemsphinxspec)
   * [ZyncSpec](#zyncspec)
   * [ZyncAppSpec](#zyncappspec)
   * [ZyncQueSpec](#zyncquespec)
   * [HighAvailabilitySpec](#highavailabilityspec)
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

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## APIManager

| **Field** | **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- | --- |
| Spec | `spec` | [APIManagerSpec](#APIManagerSpec) | Yes | The specfication for APIManager custom resource |
| Status | `status` | [APIManagerStatus](#APIManagerStatus) | No | The status for the custom resource |

### APIManagerSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| WildcardDomain | `wildcardDomain` | string | Yes | N/A | Root domain for the wildcard routes. Eg. example.com will generate 3scale-admin.example.com. |
| AppLabel | `appLabel` | string | No | `3scale-api-management` | The value of the `app` label that will be applied to the API management solution
| TenantName | `tenantName` | string | No | `3scale` | Tenant name under the root that Admin UI will be available with -admin suffix.
| ImageStreamTagImportInsecure | `imageStreamTagImportInsecure` | bool | No | `false` | Set to true if the server may bypass certificate verification or connect directly over HTTP during image import |
| ImagePullSecrets | `imagePullSecrets` | \[\][corev1.LocalObjectReference](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core) | No | `[ { name: "threescale-registry-auth" } ]` | List of image pull secrets to be used on the managed DeploymentConfigs ServiceAccounts. See [imagePullSecrets field in K8s ServiceAccount documentation](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#serviceaccount-v1-core) for details on Image pull secrets. If not specified, `threescale-registry-auth` is used. Secret names that contain `dockercfg-` or `token-` anywhere in part of its name cannot be specified. If an update to this attribute is performed the corresponding DeploymentConfig pods have to be redeployed by the user to make the changes effective |
| ResourceRequirementsEnabled | `resourceRequirementsEnabled` | bool | No | `true` | When true, 3Scale API management solution is deployed with the optimal resource requirements and limits. Setting this to false removes those resource requirements. ***Warning*** Only set it to false for development and evaluation environments. When set to `true`, default compute resources are set for the APIManager components. See [Default APIManager components compute resources](#Default-APIManager-components-compute-resources) to see the default assigned values |
| ApicastSpec | `apicast` | \*ApicastSpec | No | See [ApicastSpec](#ApicastSpec) | Spec of the Apicast part |
| BackendSpec | `backend` | \*BackendSpec | No | See [BackendSpec](#BackendSpec) reference | Spec of the Backend part |
| SystemSpec  | `system`  | \*SystemSpec  | No | See [SystemSpec](#SystemSpec) reference | Spec of the System part |
| ZyncSpec    | `zync`    | \*ZyncSpec    | No | See [ZyncSpec](#ZyncSpec) reference | Spec of the Zync part    |
| HighAvailabilitySpec | `highAvailability` | \*HighAvailabilitySpec | No | See [HighAvailabilitySpec](#HighAvailabilitySpec) reference | Spec of the HighAvailability part |
| PodDisruptionBudgetSpec | `podDisruptionBudget` | \*PodDisruptionBudgetSpec | No | See [PodDisruptionBudgetSpec](#PodDisruptionBudgetSpec) reference | Spec of the PodDisruptionBudgetSpec part |
| MonitoringSpec | `monitoring` | \*MonitoringSpec | No | Disabled | [MonitoringSpec](#MonitoringSpec) reference |

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
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| Workers | `workers` | integer | No | Automatically computed. Check [apicast doc](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_workers) for further info. | Defines the number of worker processes |
| LogLevel | `logLevel` | string | No | N/A | Log level for the OpenResty logs  (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_log_level)) |
| CustomPolicies | `customPolicies` | [][CustomPolicySpec](#CustomPolicySpec) | No | N/A | List of custom policies |

### ApicastStagingSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `apicast-staging` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| LogLevel | `logLevel` | string | No | N/A | Log level for the OpenResty logs  (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_log_level)) |
| CustomPolicies | `customPolicies` | [][CustomPolicySpec](#CustomPolicySpec) | No | N/A | List of custom policies |

### CustomPolicySpec

| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `name` | string | Yes | N/A | Name |
| `version` | string | Yes | N/A | Version |
| `SecretRef` | LocalObjectReference | Yes | N/A | Secret reference with the policy content. See [CustomPolicySecret](#CustomPolicySecret) for more information.

### CustomPolicySecret

Contains custom policy specific content. Two files,  `init.lua` and `apicast-policy.json`, are required, but more can be added optionally.

Some examples are available [here](/doc/adding-custom-policies.md)

| **Field** | **Description** |
| --- | --- |
| `init.lua` | Custom policy lua code entry point |
| `apicast-policy.json` | Custom policy metadata |

### BackendSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for Backend |
| RedisImage | `redisImage` | string | No | nil | Used to overwrite the desired Redis image for the Redis used by backend. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| RedisAffinity | `redisAffinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| RedisTolerations | `redisTolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| RedisResources | `redisResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | RedisResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| RedisPersistentVolumeClaimSpec | `redisPersistentVolumeClaim` | \*[BackendRedisPersistentVolumeClaimSpec](#BackendRedisPersistentVolumeClaimSpec) | No | nil | Backend's Redis PersistentVolumeClaim configuration options. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| ListenerSpec | `listenerSpec` | \*BackendListenerSpec | No | See [BackendListenerSpec](#BackendListenerSpec) reference | Spec of Backend Listener part |
| WorkerSpec | `workerSpec` | \*BackendWorkerSpec | No | See [BackendWorkerSpec](#BackendWorkerSpec) reference | Spec of Backend Worker part |
| CronSpec | `cronSpec` | \*BackendCronSpec | No | See [BackendCronSpec](#BackendCronSpec) reference | Spec of Backend Cron part |

### BackendRedisPersistentVolumeClaimSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |

### BackendListenerSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `backend-listener` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### BackendWorkerSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `backend-worker` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### BackendCronSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `backend-cron` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### SystemSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for System |
| RedisImage | `redisImage` | string | No | nil | Used to overwrite the desired Redis image for the Redis used by System. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| RedisPersistentVolumeClaimSpec | `redisPersistentVolumeClaim` | \*[SystemRedisPersistentVolumeClaimSpec](#SystemRedisPersistentVolumeClaimSpec) | No | nil | System's Redis PersistentVolumeClaim configuration options. Only takes effect when `.spec.highAvailability.enabled` is not set to true  |
| RedisAffinity | `redisAffinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| RedisTolerations | `redisTolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| RedisResources | `redisResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | RedisResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| MemcachedImage | `memcachedImage` | string | No | nil | Used to overwrite the desired Memcached image for the Memcached used by System. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| MemcachedAffinity | `memcachedAffinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Only takes effect when `.spec.highAvailability.enabled` is not set to true | |
| MemcachedTolerations | `memcachedTolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Only takes effect when `.spec.highAvailability.enabled` is not set to true |
| MemcachedResources | `memcachedResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | MemcachedResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| FileStorageSpec | `fileStorage` | \*SystemFileStorageSpec | No | See [FileStorageSpec](#FileStorageSpec) specification | Spec of the System's File Storage part |
| DatabaseSpec | `database` | \*SystemDatabaseSpec | No | See [DatabaseSpec](#DatabaseSpec) specification | Spec of the System's Database part |
| AppSpec | `appSpec` | \*SystemAppSpec | No | See [SystemAppSpec](#SystemAppSpec) reference | Spec of System App part |
| SidekiqSpec | `sidekiqSpec` | \*SystemSidekiqSpec | No | See [SystemSidekiqSpec](#SystemSidekiqSpec) reference | Spec of System Sidekiq part |
| SphinxSpec | `sphinxSpec` | \*SystemSphinxSpex | No | See [SystemSphinxSpec](#SystemSphinxSpec) reference | Spec of System's Sphinx part |

### SystemRedisPersistentVolumeClaimSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| StorageClassName | `storageClassName` | string | No | nil | The Storage Class to be used by the PVC |

### FileStorageSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| PVC | `persistentVolumeClaim` | \*SystemPVCSpec | No | nil | Used to use a PersistentVolumeClaim as the System's file storage. See [SystemPVCSpec](#SystemPVCSpec) |
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

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Configuration | `configurationSecretRef` | [corev1.LocalObjectReference](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core) | Yes | N/A | Local object reference to the secret to be used where the AWS configuration is stored. See [LocalObjectReference](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core) on how to specify the local object reference to the secret |

The secret name specified in the `configurationSecretRef` field must be
pre-created by the user before creating the APIManager custom resource.
Otherwise the operator will complain about it. See the
[fileStorage S3 credentials secret](#fileStorage-S3-credentials-secret)
specification to see what fields the secret should have and the values
that should be set on it.

### DeprecatedSystemS3Spec
  **DEPRECATED** Setting fields here has no effect. Use [SystemS3Spec](#SystemS3Spec) instead

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| AWSBucket | `awsBucket` | string | Yes | N/A | **DEPRECATED** Use [SystemS3Spec](#SystemS3Spec) instead |
| AWSRegion | `awsRegion` | string | Yes | N/A | **DEPRECATED** Use [SystemS3Spec](#SystemS3Spec) instead |
| AWSCredentials | `awsCredentialsSecret` | **DEPRECATED** Use [SystemS3Spec](#SystemS3Spec) instead |

### DatabaseSpec

Note: Deploying databases internally with this section is meant for evaluation purposes. Check [HighAvailabilitySpec](#highavailabilityspec) for production ready recommended deployments.

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| MySQL | `mysql`| \*SystemMySQLSpec | No | nil | Enable MySQL database as System's database. Only takes effect when `.spec.highAvailability.enabled` is not set to true. See [MySQLSpec](#MySQLSpec) specification |
| PostgreSQL | `postgresql` | \*SystemPostgreSQLSpec | No | nil | Enable PostgreSQL database as System's database. Only takes effect when `.spec.highAvailability.enabled` is not set to true. See [PostgreSQLSpec](#PostgreSQLSpec)

### MySQLSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for System's MySQL database |
| PersistentVolumeClaimSpec | `persistentVolumeClaim` | \*[SystemMySQLPVCSpec](#SystemMySQLPVCSpec) | No | nil | System's MySQL PersistentVolumeClaim configuration options |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

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
| PersistentVolumeClaimSpec | `persistentVolumeClaim` | \*[SystemPostgreSQLPVCSpec](#SystemPostgreSQLPVCSpec) | No | nil | System's PostgreSQL PersistentVolumeClaim configuration options |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

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
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| MasterContainerResources | `masterContainerResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| ProviderContainerResources | `providerContainerResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |
| DeveloperContainerResources | `developerContainerResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### SystemSidekiqSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `system-sidekiq` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### SystemSphinxSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### ZyncSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Image | `image` | string | No | nil | Used to overwrite the desired container image for Zync |
| PostgreSQLImage | `postgreSQLImage` | string | No | nil | Used to overwrite the desired PostgreSQL image for the PostgreSQL used by Zync. Does not take effect when `.spec.highAvailability.enabled` and `spec.highAvailability.externalZyncDatabase` are set to true |
| AppSpec | `appSpec` | \*ZyncAppSpec | No | See [ZyncAppSpec](#ZyncAppSpec) reference | Spec of Zync App part |
| QueSpec | `queSpec` | \*ZyncQueSpec | No | See [ZyncQueSpec](#ZyncQueSpec) reference | Spec of Zync Que part |
| DatabaseAffinity | `databaseAffinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules. Does not take effect when `.spec.highAvailability.enabled` and `spec.highAvailability.externalZyncDatabase` are set to true |
| DatabaseTolerations | `databaseTolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints. Does not take effect when `.spec.highAvailability.enabled` and `spec.highAvailability.externalZyncDatabase` are set to true |
| DatabaseResources | `databaseResources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | DatabaseResources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior. Does not take effect when `.spec.highAvailability.enabled` and `spec.highAvailability.externalZyncDatabase` are set to true |

### ZyncAppSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `zync` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### ZyncQueSpec

| **Field** | **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- | --- |
| Replicas | `replicas` | integer | No | 1 | Number of Pod replicas of the `zync-que` deployment |
| Affinity | `affinity` | [v1.Affinity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#affinity-v1-core) | No | `nil` | Affinity is a group of affinity scheduling rules |
| Tolerations | `tolerations` | \[\][v1.Tolerations](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#toleration-v1-core) | No | `nil` | Tolerations allow pods to schedule onto nodes with matching taints |
| Resources | `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | `nil` | Resources describes the compute resource requirements. Takes precedence over `spec.resourceRequirementsEnabled` with replace behavior |

### HighAvailabilitySpec

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
* [system-redis](#system-redis) with the `URL` and `MESSAGE_BUS_URL` fields
  with the value pointing to the desired external databases. The databases
  should be configured in high-availability mode

Additionally, when HighAvailability is enabled, if the `externalZyncDatabaseEnabled` field is
also enabled the user has to pre-create the following secret too:
* [zync](#zync) with the `DATABASE_URL` and `DATABASE_PASSWORD` fields
  with the values pointing to the desired external database settings.
  The database should be configured in high-availability mode

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
    * All expected DeploymentConfigs to be deployed exist and have the `Available` condition set to true
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
| `requests` | [v1 Quantity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#quantity-resource-core) | Yes | N/A | Requested size of the PersistentVolumeClaim. |

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
| REDIS_STORAGE_URL | Backend's redis storage database URL. | Mandatory when `.spec.highAvailability.enabled` is `true`. Otherwise the default value is: `redis://backend-redis:6379/0` |
| REDIS_STORAGE_SENTINEL_ROLE | Backend's redis storage sentinel role name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| REDIS_STORAGE_SENTINEL_HOSTS | Backend's redis storage sentinel hosts name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| REDIS_QUEUES_URL | Backend's redis queues database URL  | Mandatory when `.spec.highAvailability.enabled` is `true`. Otherwise the default value is: `redis://backend-redis:6379/1` |
| REDIS_QUEUES_SENTINEL_ROLE | Backend's redis queues sentinel role name. Used only when Redis sentinel is configured in the Redis database being used | `""` |
| REDIS_QUEUES_SENTINEL_HOSTS | Backend's redis queues sentinel hosts name. Used only when Redis sentinel is configured in the Redis database being used | `""` |

### system-app

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| SECRET_KEY_BASE | System application secret key base | Autogenerated value |
| USER_SESSION_TTL | System user's login session length (TTL) in seconds. The empty value means 2 weeks | `""` |

### system-database

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | URL of the Porta database. When `.spec.highAvailability.enabled` is `true` this parameter is mandatory and the format of the URL must be: `<DBScheme>://<AdminUser>:<AdminPassword>@<DatabaseHost>/<DatabaseName>`, where `<AdminUser>` must be an already existing user in the external database with full permissions on the specified `<DatabaseName>` logical database and `<DatabaseName>` must be an already existing logical database in the external database.<br/>When `.spec.highAvailability.enabled` is set to `false` the `<AdminPassword>` part of the URL can be set instead of an autogenerated one. In order to do that (if desired) set `URL` as the value specified in the `Default Value` column with the `<AutoGeneratedValue>` placeholder replaced with the desired value.| A default is only set when `spec.highAvailability.enabled` is `false`.Otherwise it is a mandatory field.<br/>When `spec.highAvailability.enabled` is `false`:<br/>If MySQL is used as System's database (the default behavior):  `mysql2://root:<AutogeneratedValue>@system-mysql/mysql`.<br/>If PostgreSQL is used(when `.spec.system.database.postgresql` is set): `postgresql://root:<AutoGeneratedValue>@system-postgresql/system`. |
| DB_USER | Non-administrative database username. Not used when `spec.highAvailability.enabled` is set to `false` | `mysql` |
| DB_PASSWORD | Password of the non-administrative database user. Not used when `spec.highAvailability.enabled` is set to `false` | Autogenerated value |
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
| SECRET_KEY | reCAPTCHA secret key (used in spam protection) for System| `""` |

### system-redis

| **Field** | **Description** | **Default value** |
| --- | --- | --- |
| URL | System's Redis database URL | Mandatory when `.spec.highAvailability.enabled` is `true`. Otherwise the default value is: `redis://system-redis:6379/1` |
| MESSAGE_BUS_URL | System's Message Bus Redis database URL | Mandatory when `.spec.highAvailability.enabled` is `true`. Otherwise the default value is: `redis://system-redis:6379/8` |
| NAMESPACE | Define the namespace to be used by System's Redis Database. The empty value means not namespaced | `""` |
| MESSAGE_BUS_NAMESPACE | Define the namespace to be used by System's Message Bus Redis Database. The empty value means not namespaced | `""` |
| SENTINEL_HOSTS | System's Redis sentinel hosts. Used only when Redis sentinel is configured | `""` |
| SENTINEL_ROLE | System's Redis sentinel role name. Used only when Redis sentinel is configured | `""` |
| MESSAGE_BUS_SENTINEL_HOSTS | System's Message Bus Redis sentinel hosts. Used only when Redis sentinel is configured | `""` |
| MESSAGE_BUS_SENTINEL_ROLE | System's Message Bus Redis sentinel role name. Used only when Redis sentinel is configured | `""` |

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
| DATABASE_URL | PostgreSQL database used by Zync. Not configurable when `spec.highAvailability.enabled` is false | When `.spec.highAvailability.enabled` and `.spec.highAvailability.zyncExternalDatabaseEnabled` are set to `true` this parameter is mandatory and has to follow the format: `postgresql://<zync-db-username>:<ZYNC_DATABASE_PASSWORD>@<zync-db-host>:<zync-db-port>/zync_production`, where `<zync-db-username>` must be an already existing user in the external database with full permissions on the `zync_production` logical database, `zync_production` logical database must be an already existing logical database in the external database and the specified value of `<ZYNC_DATABASE_PASSWORD>` must be the same as the `ZYNC_DATABASE_PASSWORD` parameter in this secret. Otherwise it has a default value, which is `postgresql://zync:<ZYNC_DATABASE_PASSWORD>@zync-database:5432/zync_production` |
| ZYNC_DATABASE_PASSWORD | Database password associated to the user specified in the `DATABASE_URL` parameter | When `.spec.highAvailability.enabled` and `.spec.highAvailability.zyncExternalDatabaseEnabled` are set to `true` this parameter is mandatory and must have the same value as the password part of the `DATABASE_URL` parameter in this secret . Otherwise the default value is an autogenerated value if not defined |
| SECRET_KEY_BASE | Zync's application key generator to encrypt communications | Autogenerated value |
| ZYNC_AUTHENTICATION_TOKEN | Authentication token used to authenticate System when calling Zync | Autogenerated value |

### fileStorage-S3-credentials-secret

The name of this secret can be any name as long as does not collide with other
existing secret names.

| **Field** | **Description** | **Required** |
| --- | --- | --- |
| AWS_ACCESS_KEY_ID | AWS Access Key ID to use in S3 Storage for System's file storage | Y |
| AWS_SECRET_ACCESS_KEY | AWS Access Key Secret to use in S3 Storage for System's file storage | Y |
| AWS_BUCKET | S3 bucket to be used as System's FileStorage for assets | Y |
| AWS_REGION | Region of the S3 bucket to be used as System's FileStorage for assets | Y |
| AWS_HOSTNAME | Default: Amazon endpoints - AWS S3 compatible provider endpoint hostname | N |
| AWS_PROTOCOL | Default: HTTPS - AWS S3 compatible provider endpoint protocol | N |
| AWS_PATH_STYLE | Default: false - When set to true, the bucket name is always left in the request URI and never moved to the host as a sub-domain | N |

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
| system-sphinx | 80m | 1000m | 250Mi | 512Mi |
| system-redis | 150m | 500m | 256Mi | 32Gi |
| system-mysql | 250m | No limit | 512Mi | 2Gi |
| system-postgresql | 250m | No limit | 512Mi | 2Gi |
| backend-listener | 500m | 1000m | 550Mi | 700Mi |
| backend-worker | 150m | 1000m | 50Mi | 300Mi |
| backend-cron | 50m | 150m | 40Mi | 150Mi |
| backend-redis | 1000m | 2000m | 1024Mi | 32Gi |
| apicast-production | 500m | 1000m | 64Mi | 128Mi |
| apicast-staging | 50m | 100m | 64Mi | 128Mi |
| zync | 150m | 1 | 250M | 512Mi |
| zync-que | 250m | 1 | 250M | 512Mi |
| zync-database | 50m | 250m | 250M | 2G |
