# 3scale installation Backup and Restore using the operator

## Table of contents
* [General description](#general-description)
* [Backing up 3scale](#backing-up-3scale)
  * [Backup compatible scenarios](#restore-compatible-scenarios)
  * [Backup workflow](#backup-workflow)
* [Restoring 3scale](#restoring-3scale)
  * [Restore compatible scenarios](#restore-compatible-scenarios)
  * [Restore workflow](#restore-workflow)
* [APIManagerBackup CRD reference](apimanagerbackup-reference.md)
* [APIManagerRestore CRD reference](apimanagerrestore-reference.md)

## General description

Backup and Restore functionality of a 3scale installation is provided for
3scale installations deployed using the [APIManager](apimanager-reference.md)
custom resource definition provided by the 3scale-operator.

Operator capabilities custom resources are not part of the 3scale Installation
functionality and are thus not included as part of the 3scale installation
Backup and Restore functionality.

To see how to back up an APIManager 3scale based installation see [#Backing up 3scale](#backing-up-3scale)

To see how to restore a previously backed up APIManager 3scale based installation
using the operator backup functionality see [restoring 3scale](#restoring-3scale)

## Backing up 3scale

The backup functionality of a 3scale installation deployed by an `APIManager`
custom resource is provided.

### Backup compatible scenarios

To see what 3scale installation configurations can be backed up see the
[APIManagerBackup reference](apimanagerbackup-reference.md). Specifically the following sections should be
consulted:
* [Backup scenarios scope](apimanagerbackup-reference.md#backup-scenarios-scope)
* [Data that is backed up](apimanagerbackup-reference.md#data-that-is-backed-up)

### Backup workflow

To backup a 3scale installation deployed with an existing APIManager the
workflow is the following one:

1. Perform a backup of the 3scale external databases:
   * backend-redis
   * system-redis
   * system database (MySQL or PostgreSQL)
1. Perform a backup of the following Kubernetes secrets:
   * backend-redis
   * system-redis
   * system-database
1. Create the APIManagerBackup Custom resource in the same namespace
   as where the 3scale installation managed by the APIManager object
   is deployed. See the [APIManagerBackup reference](apimanagerbackup-reference.md)
   to see the available fields that can be configured. An example would be:
   ```
     apiVersion: apps.3scale.net/v1alpha1
     kind: APIManagerBackup
     metadata:
      name: example-apimanagerbackup-pvc
     spec:
       backupDestination:
         persistentVolumeClaim:
           resources:
             requests: "10Gi"
   ```
1. Wait until APIManagerBackup finishes. You can check this by obtaining
   the content of APIManagerBackup and waiting until the `.status.completed` field
   is set to true.
1. At this point the backup has finished. The backup contents are detailed in
   the [APIManagerBackup reference](apimanagerbackup-reference.md#data-that-is-backed-up).
   Other fields in the `status` section of the APIManagerBackup show details of the backup,
   like the name of the PersistentVolumeClaim where the data has been backed up when
   the configured backup destination has been a PersistentVolumeClaim. Make sure
   you take note of the value of `status.backupPersistentVolumeClaimName` field

## Restoring 3scale

The restore functionality of a 3scale installation previously deployed by an `APIManager` custom
resource and backed up by an `APIManagerBackup` scenario is provided

### Compatible scenarios

To see what 3scale installation configurations can be restored see the
[APIManagerRestore reference](apimanagerrestore-reference.md). Specifically the
following sections should be consulted:
* [Restore scenarios scope](apimanagerrestore-reference.md#restore-scenarios-scope)
* [Data that is restored](apimanagerrestore-reference.md#data-that-is-restored)

### Restore workflow

To restore a 3scale installation previously deployed with an APIManager that was
backed using an APIManagerBackup custom resource the workflow is the
following one:

1. Make sure that there is no APIManager (and its corresponding 3scale installation)
   custom resource created in the namespace where 3scale is to be restored
1. Perform a restore of the 3scale external databases:
   * backend-redis
   * system-redis
   * system database (MySQL or PostgreSQL)
1. Perform a restore of the following Kubernetes secrets:
   * backend-redis
   * system-redis
   * system-database
1. Create the APIManagerRestore custom resource. Configuration of the APIManagerRestore
   has to specify backed up data of the same installation that was backed up
   by an APIManagerBackup custom resource. See the [APIManagerRestore reference](apimanagerrestore-reference.md)
   to see the available fields that can be configured. An example would be:
   ```
     apiVersion: apps.3scale.net/v1alpha1
     kind: APIManagerRestore
     metadata:
       name: example-apimanagerrestore-pvc
     spec:
      restoreSource:
        persistentVolumeClaim:
          claimSource:
            claimName: example-apimanagerbackup-pvc # Name of the PVC produced as the backup result of an APIManagerBackup
            readOnly: true
   ```
1. Wait until APIManagerRestore finishes. You can check this by obtaining
   the content of APIManagerRestore and waiting until the `.status.completed` field
   is set to true.
1. At this point the restore has finished. You should see a new APIManager custom
   resource has been created and a 3scale installation deployed by it being
   deployed and eventually running.