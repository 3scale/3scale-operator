# User Guide

## Table of contents
* [Installing 3scale](#installing-3scale)
  * [Prerequisites](#prerequisites)
  * [Basic Installation](#basic-installation)
  * [Deployment Configuration Options](#deployment-configuration-options)
    * [Evaluation Installation](#evaluation-installation)
    * [External Databases Installation](#external-databases-installation)
    * [S3 Filestorage Installation](#s3-filestorage-installation)
    * [Setting a custom Storage Class for System FileStorage RWX PVC-based installations](#setting-a-custom-storage-class-for-system-filestorage-rwx-pvc-based-installations)
    * [PostgreSQL Installation](#postgresql-installation)
    * [Enabling Pod Disruption Budgets](#enabling-pod-disruption-budgets)
    * [Setting custom affinity and tolerations](#setting-custom-affinity-and-tolerations)
    * [Setting custom compute resource requirements at component level](#setting-custom-compute-resource-requirements-at-component-level)
    * [Setting custom storage resource requirements](#setting-custom-storage-resource-requirements)
    * [Enabling monitoring resources](operator-monitoring-resources.md)
    * [Adding custom policies](adding-custom-policies.md)
* [Reconciliation](#reconciliation)
* [Upgrading 3scale](#upgrading-3scale)
* [3scale installation Backup and Restore using the operator (in *TechPreview*)](operator-backup-and-restore.md)
* [Application Capabilities (in *TechPreview*)](operator-application-capabilities.md)
* [APIManager CRD reference](apimanager-reference.md)
    * CR samples [\[1\]](../config/samples/apps_v1alpha1_apimanager_simple.yaml) [\[2\]](cr_samples/apimanager/)

## Installing 3scale

This section will take you through installing and deploying the 3scale solution via the 3scale operator,
using the [*APIManager*](apimanager-reference.md) custom resource.

Deploying the APIManager custom resource will make the operator begin processing and will deploy a
3scale solution from it.

### Prerequisites

* OpenShift Container Platform 4.1
* Deploying 3scale using the operator first requires that you follow the steps
in [quickstart guide](quickstart-guide.md) about *Install the 3scale operator*
* Some [Deployment Configuration Options](#deployment-configuration-options) require OpenShift infrastructure to provide availablity for the following persistent volumes (PV):
  * 3 RWO (ReadWriteOnce) persistent volumes
  * 1 RWX (ReadWriteMany) persistent volume
    * 3scale's System component needs a RWX(ReadWriteMany) PersistentVolume for
      its FileStorage when System's FileStorage is configured to be
      a PVC (default behavior). System's FileStorage characteristics:
      * Contains configuration files read by the System component at run-time
      * Stores Static files (HTML, CSS, JS, etc) uploaded to System by its
        CMS feature, for the purpose of creating a Developer Portal
      * System can be scaled horizontally with multiple pods uploading and
        reading said static files, hence the need for a RWX PersistentVolume
        when APIManager is configured to use PVC as System's FileStorage


The RWX persistent volume must be configured to be group writable.
For a list of persistent volume types that support the required access modes,
see the [OpenShift documentation](https://access.redhat.com/documentation/en-us/openshift_container_platform/4.1/html-single/storage/index#persistent-volumes_understanding-persistent-storage)

### Basic installation

To deploy the minimal APIManager object with all default values, follow the following procedure:
1. Click *Catalog > Installed Operators*. From the list of *Installed Operator*s, click _3scale Operator_.
1. Click *API Manager > Create APIManager*
1. Create *APIManager* object with the following content.

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: <wildcardDomain>
```

The `wildcardDomain` parameter can be any desired name you wish to give that resolves to the IP addresses
of OpenShift router nodes. Be sure to remove the placeholder marks for your parameters: `< >`.

When 3scale has been installed, a default *tenant* is created for you ready to be used,
with a fixed URL: `3scale-admin.${wildcardDomain}`.
For instance, when the *<wildCardDomain>* is `example.com`, then the Admin Portal URL would be:

```
https://3scale-admin.example.com
```

Optionally, you can create new tenants on the _MASTER portal URL_, with a fixed URL:

```
https://master.example.com
```

All required access credentials are stored in `system-seed` secret.

### Deployment Configuration Options

By default, the following deployment configuration options will be applied:
* Containers will have [k8s resources limits and requests](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) specified. The advantage is twofold: ensure a minimum performance level and limit resources to allow external services and solutions to be allocated.
* Internal databases will be deployed.
* Filestorage will be based on *Persistent Volumes*, one of them requiring
  *RWX* access mode. Openshift must be configured to provide them when
  requested. For the  RWX persistent volume, a preexisting custom storage
  class to be used can be specified by the user if desired
  (see [Setting a custom Storage Class for System FileStorage RWX PVC-based installations](#setting-a-custom-storage-class-for-system-filestorage-rwx-pvc-based-installations))
* Mysql will be the internal relational database deployed.

Default configuration option is suitable for PoC or evaluation by a customer.

One, many or all of the default configuration options can be overriden with specific field values in
the [*APIManager*](apimanager-reference.md) custom resource.
The 3scale operator allows all available combinations whereas templates only fixed deployment profiles.
For instance, the operator allows you to deploy 3scale in evaluation mode and external databases mode.
Templates do not allow this specific deployment configuration. There are templates available only for the most common configuration options.

#### Evaluation Installation
Containers will not have [k8s resources limits and requests](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) specified.

* Small memory footprint
* Fast startup
* Runnable on laptop
* Suitable for presale/sales demos

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  resourceRequirementsEnabled: false
```

Check [*APIManager*](apimanager-reference.md) custom resource for reference.

#### External Databases Installation
Suitable for production use where customer wants HA or to re-use DB of their own.

3scale API Management has been tested and it’s supported with the following databases:

| Database | Version |
| :--- | :--- |
| MySQL | 5.7 |
| Redis | 3.2 |
| PostgreSQL | 10.6 |

Before creating *APIManager* custom resource to deploy 3scale,
connection settings for the external databases needs to be provided using openshift secrets.

* **Backend redis secret**

Two external redis instances must be deployed by the customer. Then fill connection settings.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backend-redis
stringData:
  REDIS_STORAGE_URL: "redis://backend-redis-storage"
  REDIS_STORAGE_SENTINEL_HOSTS: "redis://sentinel-0.example.com:26379,redis://sentinel-1.example.com:26379, redis://sentinel-2.example.com:26379"
  REDIS_STORAGE_SENTINEL_ROLE: "master"
  REDIS_QUEUES_URL: "redis://backend-redis-queues"
  REDIS_QUEUES_SENTINEL_HOSTS: "redis://sentinel-0.example.com:26379,redis://sentinel-1.example.com:26379, redis://sentinel-2.example.com:26379"
  REDIS_QUEUES_SENTINEL_ROLE: "master"
type: Opaque
```

Secret name must be `backend-redis`.

See [Backend redis secret](apimanager-reference.md#backend-redis) for reference.

* **System redis secret**

Two external redis instances must be deployed by the customer. Then fill connection settings.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: system-redis
stringData:
  URL: "redis://system-redis"
  SENTINEL_HOSTS: "redis://sentinel-0.example.com:26379,redis://sentinel-1.example.com:26379, redis://sentinel-2.example.com:26379"
  SENTINEL_ROLE: "master"
  NAMESPACE: ""
  MESSAGE_BUS_URL: "redis://system-redis-messagebus"
  MESSAGE_BUS_SENTINEL_HOSTS: "redis://sentinel-0.example.com:26379,redis://sentinel-1.example.com:26379, redis://sentinel-2.example.com:26379"
  MESSAGE_BUS_SENTINEL_ROLE: "master"
  MESSAGE_BUS_NAMESPACE: ""
type: Opaque
```

Secret name must be `system-redis`.

See [System redis secret](apimanager-reference.md#system-redis) for reference.

* **System database secret**
MySQL or PostgreSQL database instance must be deployed by the customer. Then fill connection settings.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: system-database
stringData:
  URL: "mysql2://root:password0@system-mysql/system"
  DB_USER: "mysql"
  DB_PASSWORD: "password1"
type: Opaque
```

Secret name must be `system-database`.

See [System database secret](apimanager-reference.md#system-database) for reference.

Finally, create *APIManager* custom resource to deploy 3scale

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  highAvailability:
    enabled: true
```

Additionally, Zync Database can also be configured as an external database.

* **Zync secret**

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: zync
stringData:
  DATABASE_URL: postgresql://<zync-db-user>:<zync-db-password>@<zync-db-host>:<zync-db-port>/zync_production
  ZYNC_DATABASE_PASSWORD: <zync-db-password>
type: Opaque
```

Secret name must be `zync`.

See [Zync secret](apimanager-reference.md#zync) for reference.


In this case, finally create *APIManager* custom resource to deploy 3scale
with the previous external databases and zync database externally too:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  highAvailability:
    enabled: true
    externalZyncDatabaseEnabled: true
```

Check [*APIManager HighAvailabilitySpec*](apimanager-reference.md#HighAvailabilitySpec) for reference.

#### S3 Filestorage Installation
3scale’s FileStorage being in a S3 service instead of in a PVC.

Before creating *APIManager* custom resource to deploy 3scale,
connection settings for the S3 service needs to be provided using an openshift secret.

* **S3 secret**

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-auth
stringData:
  AWS_ACCESS_KEY_ID: 1234567
  AWS_SECRET_ACCESS_KEY: 987654321
  AWS_BUCKET: mybucket.example.com
  AWS_REGION: eu-west-1
type: Opaque
```

Secret name can be anyone, as it will be referenced in the *APIManager* custom resource.

**AWS S3 compatible provider**

AWS S3 compatible provider can be configured in the S3 secret with *AWS_HOSTNAME*,
*AWS_PATH_STYLE* and *AWS_PROTOCOL* optional keys.
Check [S3 secret reference](apimanager-reference.md#fileStorage-S3-credentials-secret) for reference.

Finally, create *APIManager* custom resource to deploy 3scale

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  system:
    fileStorage:
      simpleStorageService:
        configurationSecretRef:
          name: aws-auth
```

Note that S3 secret name is provided directly in the APIManager custom resource.

Check [*APIManager SystemS3Spec*](apimanager-reference.md#SystemS3Spec) for reference.

#### Setting a custom Storage Class for System FileStorage RWX PVC-based installations

When deploying an APIManager using PVC as System's FileStorage (default behavior), the
[default storage class configured in the user's cluster](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#defaultstorageclass) is automatically used to provision System's FileStorage
RWX(ReadWriteMany) PVC.

It's sometimes the case that a user might want to provision System's FileStorage
PVC with another [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/) different than the default one, either because it
does not want to use the default storage class or because the default storage
class does not provision persistent volumes compatible with ReadWriteMany(RWX)
access modes.

For this reason, APIManager allows a user to configure an existing custom storage
class for System's FileStorage PVC.

Important: When specifying a custom storage class for System's PVC, The
specified storage class must be able to provision ReadWriteMany(RWX) Persistent
Volumes (see [Prerequisites](#Prerequisites))

To configure System's FileStorage PVC Storage Class to be used:
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  system:
    fileStorage:
      persistentVolumeClaim:
        storageClassName: <existing-storage-class-name>
```

For example, if a user has deployed and configured a storage class that
provisions PVC volumes through NFS, and has named this storage class `nfs`,
the value of `<existing-storage-class-name>` should be `nfs`

#### PostgreSQL Installation

By default, Mysql will be the internal relational database deployed.
This deployment configuration can be overrided to use PostgreSQL instead.

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  system:
    database:
      postgresql: {}
```

Check [*APIManager DatabaseSpec*](apimanager-reference.md#DatabaseSpec) for reference.

#### Enabling Pod Disruption Budgets
The 3scale API Management solution DeploymentConfigs deployed and managed by the
APIManager will be configured with Kubernetes Pod Disruption Budgets
enabled.

A Pod Disruption Budget limits the number of pods related to an application
(in this case, pods of a DeploymentConfig) that are down simultaneously
from **voluntary disruptions**.

When enabling the Pod Disruption Budgets for non-database DeploymentConfigs will
be set with a setting of maximum of 1 unavailable pod at any given time.
Database-related DeploymentConfigs are excluded from this configuration.
Additionally, `system-sphinx` DeploymentConfig is also excluded.

For details about the behavior of Pod Disruption Budgets, what they perform and
what constitutes a 'voluntary disruption' see the following
[Kubernetes Documentation](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)

Pods which are deleted or unavailable due to a rolling upgrade to an application
do count against the disruption budget, but the DeploymentConfigs are not
limited by Pod Disruption Budgets when doing rolling upgrades or they are
scaled up/down.

In order for the Pod Disruption Budget setting to be effective the number of
replicas of each non-database component has to be set to a value greater than 1.

Example:
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  apicast:
    stagingSpec:
      replicas: 2
    productionSpec:
      replicas: 2
  backend:
    listenerSpec:
      replicas: 2
    workerSpec:
      replicas: 2
    cronSpec:
      replicas: 2
  system:
    appSpec:
      replicas: 2
    sidekiqSpec:
      replicas: 2
  zync:
    appSpec:
      replicas: 2
    queSpec:
      replicas: 2
  podDisruptionBudget:
    enabled: true
```

#### Setting custom affinity and tolerations

Kubernetes [Affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity
) and [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)
can be customized in a 3scale API Management solution through APIManager
CR attributes in order to customize where/how the different 3scale components of
an installation are scheduled onto Kubernetes Nodes.

For example, setting a custom node affinity for backend listener
and custom tolerations for system's memcached would be done in the
following way:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  backend:
    listenerSpec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: "kubernetes.io/hostname"
                operator: In
                values:
                - ip-10-96-1-105
              - key: "beta.kubernetes.io/arch"
                operator: In
                values:
                - amd64
  system:
    memcachedTolerations:
    - key: key1
      value: value1
      operator: Equal
      effect: NoSchedule
    - key: key2
      value: value2
      operator: Equal
      effect: NoSchedule
```

See [APIManager reference](apimanager-reference.md) for a full list of
attributes related to affinity and tolerations.

#### Setting custom compute resource requirements at component level

Kubernetes [Compute Resource Requirements](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
can be customized in a 3scale API Management solution through APIManager
CR attributes in order to customize compute resource requirements (this is, CPU
and Memory) assigned to a specific APIManager's component.

For example, setting custom compute resource requirements for system-master's
system-provider container, for backend-listener and for zync-database can be
done in the following way:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  backend:
    listenerSpec:
      resources:
        requests:
          memory: "150Mi"
          cpu: "300m"
        limits:
          memory: "500Mi"
          cpu: "1000m"
  system:
    appSpec:
      providerContainerResources:
        requests:
          memory: "111Mi"
          cpu: "222m"
        limits:
          memory: "333Mi"
          cpu: "444m"
  zync:
    databaseResources:
      requests:
        memory: "111Mi"
        cpu: "222m"
      limits:
        memory: "333Mi"
        cpu: "444m"
```

See [APIManager reference](apimanager-reference.md) for a full list of
attributes related to compute resource requirements.

#### Setting custom storage resource requirements

Openshift [storage resource requirements](https://docs.openshift.com/container-platform/4.5/storage/persistent_storage/persistent-storage-local.html#create-local-pvc_persistent-storage-local)
can be customized through APIManager CR attributes.

This is the list of 3scale PVC's resources that can be customized with examples.

* *System Shared (RWX) Storage PVC*

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager1
spec:
  wildcardDomain: example.com
  system:
    fileStorage:
      persistentVolumeClaim:
        resources:
          requests: 2Gi
```

* *MySQL (RWO) PVC*
```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager1
spec:
  wildcardDomain: example.com
  system:
    database:
      mysql:
        persistentVolumeClaim:
          resources:
            requests: 2Gi
```

* *PostgreSQL (RWO) PVC*
```
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager1
spec:
  wildcardDomain: example.com
  system:
    database:
      postgresql:
        persistentVolumeClaim:
          resources:
            requests: 2Gi
```

*IMPORTANT NOTE*: Storage resource requirements are **usually** install only attributes.
Only when the underlying PersistentVolume's storageclass allows resizing, storage resource requirements can be modified after installation.
Check [Expanding persistent volumes](https://docs.openshift.com/container-platform/4.5/storage/expanding-persistent-volumes.html) official doc for more information.

### Reconciliation
After 3scale API Management solution has been installed, 3scale Operator enables updating a given set
of parameters from the custom resource in order to modify system configuration options.
Modifications are performed in a hot swapping way, i.e., without stopping or shutting down the system.

**Not all the parameters of the [APIManager CRD](apimanager-reference.md) are reconciliable**

The following is a list of reconciliable parameters.

* [Resources](#resources)
* [Backend replicas](#backend-replicas)
* [Apicast replicas](#apicast-replicas)
* [System replicas](#system-replicas)
* [Pod Disruption Budget](#pod-disruption-budget)

#### Resources
Resource limits and requests for all 3scale components

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  ResourceRequirementsEnabled: true/false
```

#### Backend replicas
Backend components pod count

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  backend:
    listenerSpec:
      replicas: X
    workerSpec:
      replicas: Y
    cronSpec:
      replicas: Z
```

#### Apicast replicas
Apicast staging and production components pod count

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  apicast:
    productionSpec:
      replicas: X
    stagingSpec:
      replicas: Z
```

#### System replicas
System app and system sidekiq components pod count

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  system:
    appSpec:
      replicas: X
    sidekiqSpec:
      replicas: Z
```

#### Pod Disruption Budget
Whether Pod Disruption Budgets are enabled for non-database DeploymentConfigs

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  ...
  podDisruptionBudget:
    enabled: true/false
  ...
```

### Upgrading 3scale
Upgrading 3scale API Management solution requires upgrading 3scale operator.
However, upgrading 3scale operator does not necessarily imply upgrading 3scale API Management solution.
The operator could be upgraded because there are bugfixes or security fixes.

The recommended way to upgrade the 3scale operator is via the Operator Lifecycle Manager (OLM).

If you selected *Automatic updates* when 3scale operator was installed,
when a new version of the operator is available, the Operator Lifecycle Manager (OLM)
automatically upgrades the running instance of the operator without human intervention.

If you selected *Manual updates*, when a newer version of the Operator is available,
the OLM creates an update request. As a cluster administrator, you must then manually approve
that update request to have the Operator updated to the new version.
