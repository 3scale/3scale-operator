# APIManagerBackup reference

The following Custom Resources are provided:

`APIManagerBackup`

This resource is the resource used to backup a 3scale API Management solution
deployed using an APIManager custom resource.

## Backup scenarios scope

Backup functionality is available when the following databases are
configured externally:
* System database (MySQL or PostgreSQL)
* Backend Redis database
* System Redis database

## Data that is backed up

* Secrets
  * system-smtp
  * system-seed
  * backend-internal-api
  * backend-listener
  * system-events-hook
  * system-app
  * system-recaptcha
  * zync
  * system-master-apicast
  * system-memcache
  * system-database
  * backend-redis
  * system-redis

* ConfigMaps
  * system-environment
  * apicast-environment

* APIManager
  * APIManager's custom resource Kubernetes object definition (json schema definition)

* System's FileStorage
  *  When the location of System's FileStorage is in a PersistentVolumeClaim (PVC)
  * **CURRENTLY UNSUPPORTED** When the location of System's FileStorage is in a S3 API-compatible storage

## Data that is not backed up

Backups of the external databases used by 3scale are not part of the
3scale-operator functionality and has to be performed by the user appropriately

## APIManagerBackup

| **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- |
| `spec` | [APIManagerBackupSpec](#APIManagerBackupSpec) | Yes | The specfication for APIManagerBackup custom resource |
| `status` | [APIManagerBackupStatusSpec](#APIManagerBackupStatusSpec) | No | The status of APIManagerBackup custom resource |

### APIManagerBackupSpec

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `apiManagerName` | string | No | Name of the APIManager deployed in the same namespace as the deployed APIManagerBackup | Name of the APIManager to backup |
| `backupDestination` | [APIManagerBackupDestinationSpec](#APIManagerBackupDestinationSpec) | Yes | See [APIManagerBackupDestinationSpec](#APIManagerBackupDestinationSpec) | Configuration related to where the backup is performed |

### APIManagerBackupDestinationSpec

This section controls where APIManager's backup is to be stored.
**One of the fields is mandatory to be set. Only one of the fields can be set. The fields are mutually exclusive.**

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `persistentVolumeClaim` | [PersistentVolumeClaimBackupDestination](#PersistentVolumeClaimBackupDestination) | No | nil | APIManager backup destination in PVC |

### PersistentVolumeClaimBackupDestination

There are two main ways to provide a PersistentVolumeClaim for the backup:
* Providing the volume name of an already existing Kubernetes PersistentVolume
  through the usage of the `volumeName` field. The already existing Kubernetes
  PersistentVolume has to be appropriately sized to contain
  all [data that is backed up](#data-that-is-backed-up). In this case storage
  size on the `resources` field has to be also specified, although it will be
  ignored as per K8s PersistentVolumeClaim requirements behavior
* Providing the desired size of the PersistentVolumeClaim through the
  `resources` field. In this case the `storageClass` field
  can be also set to specify what StorageClass will be used for the
  PersistentVolumeClaim if desired. Make sure enough resources are defined
  to contain all [data that is backed up](#data-that-is-backed-up)

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `resources` | [PersistentVolumeClaimResourcesSpec](#PersistentVolumeClaimResourcesSpec) | No | See [PersistentVolumeClaimResourcesSpec](#PersistentVolumeClaimResourcesSpec) | |
| `volumeName` | string | No | N/A | A binding reference to the PersistentVolume backing this claim. This is not the persistentVolumeClaim name. See the field `volumeName` in the [Kubernetes PersistentVolumeClaim API reference](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#persistentvolumeclaimspec-v1-core) for more information |
| `storageClass` | string | No | N/A | Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1 |

### PersistentVolumeClaimResourcesSpec

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `requests` | [v1 Quantity](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#quantity-resource-core) | Yes | N/A | Size of the PersistentVolumeClaim where the backup is to be performed. Set enough size to contain all [data that is backed up](#data-that-is-backed-up).

## APIManagerBackupStatusSpec

TODO complete status section with the status fields of the different steps. Not done at the moment as they are often changed
and they are not that important for the end-user point of view.

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `completed` | bool | No | false | `true` when APIManager's backup has finished |
| `apiManagerSourceName` | string | No | `""` | Name of the APIManager that APIManagerBackup handles |
| `startTime` | [meta/v1 Time](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#time-v1-meta) | No | N/A | Start time of the backup (in UTC) |
| `completionTime` | [meta/v1 Time](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#time-v1-meta) | No | `""` | Represents the time the backup was completed | 
| `backupPersistentVolumeClaimName` | string | No | `""` | Name of the PersistentVolumeClaim where the backup has been stored |