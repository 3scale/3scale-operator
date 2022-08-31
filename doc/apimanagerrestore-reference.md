# APIManagerRestore reference

The following Custom Resources are provided:

`APIManagerRestore`

This resource is the resource used to restore a 3scale API Management solution
previously deployed using an `APIManager` custom resource and the solution was
backed up by an `APIManagerBackup` custom resource.

## Table of Contents

* [Restore scenarios scope](#restore-scenarios-scope)
* [Data that is restored](#data-that-is-restored)
* [Data that is not restored](#data-that-is-not-restored)
* [APIManagerRestore](#apimanagerrestore)
   * [APIManagerRestoreSpec](#apimanagerrestorespec)
   * [APIManagerRestoreSourceSpec](#apimanagerrestoresourcespec)
   * [PersistentVolumeClaimRestoreSource](#persistentvolumeclaimrestoresource)
* [APIManagerRestoreStatusSpec](#apimanagerrestorestatusspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## Restore scenarios scope

Is in the scope of restore functionality of the operator:
* Restore using a back up that was generated using an `APIManagerBackup` custom
  resource. To see what 3scale API Management solution scenarios can be backed
  up please see the `APIManagerBackup` reference.

Is not the scope of restore functionality of the operator:
*  Restoring backed up data that was not performed using an `APIManagerBackup`
   custom resource
*  Restoring backed up data provided through an `APIManagerBackup` in a
   different 3scale version

## Data that is restored

* Secrets
  * system-smtp
  * system-seed
  * backend-internal-api
  * system-events-hook
  * system-app
  * system-recaptcha
  * zync
  * system-master-apicast

* ConfigMaps
  * system-environment
  * apicast-environment

* APIManager
  * APIManager's custom resource Kubernetes object definition (json schema definition)

* System's FileStorage
  * In a PersistentVolumeClaim
    * When the backed up System's FileStorage data was stored in a PersistentVolumeClaim
    * **CURRENTLY UNSUPPORTED**  When the backed up System's FileStorage data was stored in a S3 API-compatible storage

* 3scale related OpenShift routes (master, tenants, ...)

## Data that is not restored

Restore of the backed up external databases data used by 3scale is not part of
the 3scale-operator functionality and has to be performed by the user appropriately
before deploying the `APIManagerRestore` object

Restore of the following Secrets is not part of the 3scale-operator functionality
and has to be performed by the user appropriately:
  * system-database
  * backend-redis
  * system-redis

The reason for this is to allow the user to configure different database endpoints
than the ones used in the previous 3scale installation that was backed up

## APIManagerRestore

| **json/yaml field**| **Type** | **Required** | **Description** |
| --- | --- | --- | --- |
| `spec` | [APIManagerRestoreSpec](#APIManagerRestoreSpec) | Yes | The specfication for APIManagerBackup custom resource |
| `status` | [APIManagerRestoreStatusSpec](#APIManagerRestoreStatusSpec) | No | The status of APIManagerBackup custom resource |

### APIManagerRestoreSpec

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `restoreSource` | [APIManagerRestoreSourceSpec](#APIManagerRestoreSourceSpec) | Yes | See [APIManagerRestoreSourceSpec](#APIManagerRestoreSourceSpec) | Configuration related to from where the backup is restored |

### APIManagerRestoreSourceSpec

This section controls from where APIManager's is to be stored.

**One of the fields is mandatory to be set. Only one of the fields can be set. The fields are mutually exclusive.**

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `persistentVolumeClaim` | [PersistentVolumeClaimRestoreSource](#PersistentVolumeClaimRestoreSource) | No | nil | APIManager restore source from PVC |

### PersistentVolumeClaimRestoreSource
| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `claimSource` | [v1 PersistentVolumeClaimVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#persistentvolumeclaimvolumesource-v1-core) | Yes | N/A | PersistentvolumeClaim source where the backup is to be restored from |

## APIManagerRestoreStatusSpec

TODO complete status section with the status fields of the different steps. Not done at the moment as they are often changed
and they are not that important for the end-user point of view.

| **json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `completed` | bool | No | false | `true` when APIManager's restore has finished |
