# User Guide

<!--ts-->
* [User Guide](#user-guide)
   * [Installing 3scale](#installing-3scale)
      * [Prerequisites](#prerequisites)
      * [Basic installation](#basic-installation)
      * [Deployment Configuration Options](#deployment-configuration-options)
         * [Evaluation Installation](#evaluation-installation)
         * [External Databases Installation](#external-databases-installation)
         * [S3 Filestorage Installation](#s3-filestorage-installation)
            * [Long Term S3 IAM credentials](#long-term-s3-iam-credentials)
            * [Manual mode with STS](#manual-mode-with-sts)
            * [AWS S3 compatible provider](#aws-s3-compatible-provider)
         * [Setting a custom Storage Class for System FileStorage RWX PVC-based installations](#setting-a-custom-storage-class-for-system-filestorage-rwx-pvc-based-installations)
         * [PostgreSQL Installation](#postgresql-installation)
         * [Enabling Pod Disruption Budgets](#enabling-pod-disruption-budgets)
         * [Setting custom affinity and tolerations](#setting-custom-affinity-and-tolerations)
         * [Setting custom compute resource requirements at component level](#setting-custom-compute-resource-requirements-at-component-level)
         * [Setting custom storage resource requirements](#setting-custom-storage-resource-requirements)
         * [Setting custom PriorityClassName](#setting-custom-priorityclassname)
         * [Setting Horizontal Pod Autoscaling](#setting-horizontal-pod-autoscaling)
         * [Setting custom TopologySpreadConstraints](#setting-custom-topologyspreadconstraints)
         * [Setting custom labels](#setting-custom-labels)
         * [Setting custom Annotations](#setting-custom-annotations)
         * [Setting porta client to skip certificate verification](#setting-porta-client-to-skip-certificate-verification)
         * [Gateway instrumentation](#gateway-instrumentation)
      * [Preflight checks](#preflights)
      * [Reconciliation](#reconciliation)
         * [Resources](#resources)
         * [Backend replicas](#backend-replicas)
         * [Apicast replicas](#apicast-replicas)
         * [System replicas](#system-replicas)
         * [Pod Disruption Budget](#pod-disruption-budget)
      * [Upgrading 3scale](#upgrading-3scale)
      * [3scale installation Backup and Restore](#3scale-installation-backup-and-restore)
      * [Application Capabilities](#application-capabilities)
      * [APIManager CRD reference](#apimanager-crd-reference)
         * [CR Samples](#cr-samples)
<!--te-->

## Installing 3scale

This section will take you through installing and deploying the 3scale solution via the 3scale operator,
using the [*APIManager*](apimanager-reference.md) custom resource.

Deploying the APIManager custom resource will make the operator begin processing and will deploy a
3scale solution from it.

### Prerequisites

* OpenShift Container Platform 4.1
* Deploying 3scale using the operator first requires that you follow the steps
in [quickstart guide](quickstart-guide.md) about *Install the 3scale operator*
* Some [Deployment Configuration Options](#deployment-configuration-options) require OpenShift infrastructure to provide availability for the following persistent volumes (PV):
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
For instance, when the `<wildCardDomain>` is `example.com`, then the Admin Portal URL would be:

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

One, many or all of the default configuration options can be overridden with specific field values in
the [*APIManager*](apimanager-reference.md) custom resource.

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
Suitable for production use where customers want self-managed databases.

3scale API Management has been tested and it’s supported with the following database versions:

| Database | Version |
| :--- | :--- |
| MySQL | 8.0 |
| Redis | 5 |
| PostgreSQL | 10.6 |


3scale API Management requires the following database instances:

* Backend Redis (two instances: storage and queue)
* System Redis
* System RDBMS
* Zync RDBMS

The [*APIManager External Component Spec*](apimanager-reference.md#ExternalComponentsSpec)
allows to pick which databases will be externally managed and with databases will be managed by the
3scale operator. The following example helps to illustrate:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: lvh.me
  externalComponents:
    backend:
      redis: true
    system:
      database: false
      redis: true
    zync:
      database: false
```

In the example above, backend redis and system redis will be externally managed and the 3scale
operator will deploy and configure system database and zync database components.

The 3scale operator requires a secret with the connection string for every single external database
component. If a externally managed database secret does not exist, the operator will not deploy 3scale.

Use of slaves for Internal Redis is not supported.

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

* **Zync database secret**

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

#### S3 Filestorage Installation
3scale’s FileStorage being in a S3 service instead of in a PVC.

Two S3 authentication methods are supported:
* [Manual mode with STS](https://docs.openshift.com/container-platform/4.9/authentication/managing_cloud_provider_credentials/cco-mode-sts.html): Secure Token Service for  short-term, limited-privilege security credentials
* Long Term S3 IAM credentials

##### Long Term S3 IAM credentials

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

Check [S3 secret reference](apimanager-reference.md#fileStorage-S3-credentials-secret) for reference.

**ApiManager**

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager-sample
  namespace: 3scale-test
spec:
  wildcardDomain: <wildcardDomain>
  system:
    fileStorage:
      simpleStorageService:
        configurationSecretRef:
          name: s3-credentials
```

##### Manual mode with STS

**ApiManager**
STS authentication mode needs to be explicitly enabled from the APIManager CR.

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager-sample
  namespace: 3scale-test
spec:
  wildcardDomain: <wildcardDomain>
  system:
    fileStorage:
      simpleStorageService:
        configurationSecretRef:
          name: s3-credentials
        sts:
          enabled: true
          audience: openshift
```
Users are allowed to define their own **audience** if necessary; the default value is `openshift`.

With the new support for STS authentication (Secure Token Service for  short-term, limited-privilege security credentials),
the secret generated by the Cloud Credential tooling looks differ from IAM (Identity and Access Management) secret.
There are two new fields `AWS_ROLE_ARN` and `AWS_WEB_IDENTITY_TOKEN_FILE` are present instead of
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

**STS Secret sample**
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: s3-credentials
  namespace: 3scale-test
data:
  AWS_ROLE_ARN: XXXXX=
  AWS_WEB_IDENTITY_TOKEN_FILE: XXXXX=
  AWS_BUCKET: XXXXX=
  AWS_REGION: XXX
type: Opaque
```

Secret name can be anyone, as it will be referenced in the *APIManager* custom resource.

Check [S3 secret reference](apimanager-reference.md#fileStorage-S3-credentials-secret) for reference.

**Summary for keys for each secret "type"**

|Secret key                 |Required for IAM|Required for STS|
|---------------------------|---|---|
AWS_ACCESS_KEY_ID           |Y|N|
AWS_SECRET_ACCESS_KEY       |Y|N|
AWS_ROLE_ARN                |N|Y|
AWS_WEB_IDENTITY_TOKEN_FILE |N|Y|
AWS_BUCKET                  |Y|Y|
AWS_REGION                  |Y|Y|
AWS_HOSTNAME                |N|N|
AWS_PROTOCOL                |N|N|
AWS_PATH_STYLE              |N|N|

**In case of STS - the operator will add a projected volume to request the token**
Following pods will have projected volume in case of STS:
- system-app
- system-app hook pre
- system-sidekiq

**Pod example for STS**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: system-sidekiq-1-zncrz
  namespace: 3scale-test
spec:
  containers:
  ....
    volumeMounts:
    - mountPath: /var/run/secrets/openshift/serviceaccount
      name: s3-credentials
      readOnly: true
  .....
  volumes:
  - name: s3-credentials
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          audience: openshift
          expirationSeconds: 3600
          path: token
```

_Reference to STS configured cluster pre-requisite:_
- https://docs.openshift.com/container-platform/4.11/authentication/managing_cloud_provider_credentials/cco-mode-sts.html
- https://github.com/openshift/cloud-credential-operator/blob/master/docs/sts.md

##### AWS S3 compatible provider

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
The 3scale API Management solution Deployments deployed and managed by the
APIManager will be configured with Kubernetes Pod Disruption Budgets
enabled.

A Pod Disruption Budget limits the number of pods related to an application
(in this case, pods of a Deployment) that are down simultaneously
from **voluntary disruptions**.

When enabling the Pod Disruption Budgets for non-database Deployments will
be set with a setting of maximum of 1 unavailable pod at any given time.
Database-related Deployments are excluded from this configuration.
Additionally, `system-sphinx` Deployment is also excluded.

For details about the behavior of Pod Disruption Budgets, what they perform and
what constitutes a 'voluntary disruption' see the following
[Kubernetes Documentation](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)

Pods which are deleted or unavailable due to a rolling upgrade to an application
do count against the disruption budget, but the Deployments are not
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

#### Setting custom PriorityClassName
PriorityClassName specifies the Pod priority.  See [here](https://docs.openshift.com/container-platform/4.13/nodes/pods/nodes-pods-priority.html) for more information.   
It be can be customized through APIManager CR `priorityClassName` attribute for each Deployment.  
Example for apicast-staging and backend-listener:
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
    name: example-apimanager
spec:
    wildcardDomain: example.com
    resourceRequirementsEnabled: false
    apicast:
        stagingSpec:
            priorityClassName: openshift-user-critical
    backend:
        listenerSpec:
            priorityClassName: openshift-user-critical
```
#### Setting Horizontal Pod Autoscaling 
Horizontal Pod Autoscaling(HPA) is available for Apicast-production, Backend-listener and Backend-worker. The backend
components require Redis running [async mode](https://github.com/3scale/apisonator/blob/master/docs/openshift_horizontal_scaling.md#async). 
Async is enabled by default by the operator provided you aren't using [logical Redis databases](https://github.com/3scale/apisonator/blob/master/docs/openshift_horizontal_scaling.md#redis-databases).
If you are not running in Async mode you won't be able to enable HPA for the backend.

Provided you are running in Async mode, you can enable hpa for the components and accept the default configuration which
will give you a HPA with 90% resources set and max and min pods set to 5 and 1. The following is an example of the 
output HPA for backend-worker using the defaults. 

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata: 
  name: backend-worker
  namespace: 3scale-test
spec: 
  scaleTargetRef: 
    apiVersion: apps/v1
    kind: Deployment
    name: backend-worker
  minReplicas: 1
  maxReplicas: 5
  metrics: 
    - type: Resource
      resource: 
        name: cpu
        target: 
          averageUtilization: 90
          type: Utilization
    - type: Resource
      resource: 
        name: memory
        target: 
          averageUtilization: 90
          type: Utilization
```
Here is an example of the APIManager CR set with backend-worker, backend-listener and apicast-production set to default 
HPA values e.g.
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
    name: example-apimanager
spec:
    wildcardDomain: example.com
    resourceRequirementsEnabled: false
    apicast:
        productionSpec:
          hpa: true
    backend:
        listenerSpec:
          hpa: true
        workerSpec:
          hpa: true
```
Removing hpa field or setting enabled to false will remove the HPA for the component. 
Once `hpa: true` is set, instances of HPA will be created with the default values. You can manually edit these HPA 
instances to optimize your [configuration](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).
With hpa enabled it overrides and ignores values for replicas and resources for apicast production, backend listener and worker. 


#### Setting custom TopologySpreadConstraints
TopologySpreadConstraints specifies how to spread matching pods among the given topology.  See [here](https://docs.openshift.com/container-platform/4.13/nodes/scheduling/nodes-scheduler-pod-topology-spread-constraints.html) for more information.  
It be can be customized through APIManager CR `topologySpreadConstraints` attribute for each Deployment.  
Example for apicast-staging and backend-listener:
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
    name: example-apimanager
spec:
    wildcardDomain: example.com
    resourceRequirementsEnabled: false
    apicast:
        stagingSpec:
            priorityClassName: openshift-user-critical
            topologySpreadConstraints:
            - maxSkew: 1
              topologyKey: topology.kubernetes.io/zone
              whenUnsatisfiable: ScheduleAnyway
              labelSelector:
                matchLabels:
                  app: 3scale-api-management
    backend:
        listenerSpec:
            priorityClassName: openshift-user-critical
            topologySpreadConstraints:
            - maxSkew: 1
              topologyKey: topology.kubernetes.io/zone
              whenUnsatisfiable: ScheduleAnyway
              labelSelector:
                matchLabels:
                  app: 3scale-api-management
```

#### Setting custom labels
[Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) can be customized through the APIManager CR `labels` attribute for each Deployment and are applied to  their pods.
Example for apicast-staging and backend-listener:
```yaml
 apiVersion: apps.3scale.net/v1alpha1
 kind: APIManager
 metadata:
   name: example-apimanager
 spec:
   wildcardDomain: example.com
   resourceRequirementsEnabled: false
   backend: 
    listenerSpec:
       labels:
         backendLabel1: sample-label1
         backendLabel2: sample-label2
   apicast:
     stagingSpec:
       labels:
         apicastStagingLabel1: sample-label1
         apicastStagingLabel2: sample-label2
```


#### Setting custom Annotations

3scale components pods can have Annotations - key/value pairs. See [here](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) 
for more information. Annotations can be customized via APIManager CR for any 3scale component.    
Example for apicast-staging and backend-listener:  
```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-apimanager
spec:
  wildcardDomain: example.com
  apicast:
    stagingSpec:
      annotations:
        anno-sample1: anno1
  backend:
    listenerSpec:
      annotations:
        anno-sample2: anno2
```

#### Setting porta client to skip certificate verification
Whenever a controller reconciles an object it creates a new porta client to make API calls. That client is configured to verify the server's certificate chain by default. For development/testing purposes, you may want the client to skip certificate verification when reconciling an object. This can be done using the annotation `insecure_skip_verify: true`, which can be added to the following objects:
* ActiveDoc
* Application
* Backend
* CustomPolicyDefinition
* DeveloperAccount
* DeveloperUser
* OpenAPI - backend and product
* Product
* ProxyConfigPromote
* Tenant

#### Gateway instrumentation

Please refer to [Gateway instrumentation](gateway-instrumentation.md) document

### Preflights

Operator will perform a set of preflight checks to ensure that:
- the database versions are of minimum required versions
- in the event of upgrades, the upgrade on APIManager instance can be performed without breaking existing APIManager instance

Operator will create a config map called "3scale-api-management-operator-requirements" which will list the required 
versions of the databases, which include:
- system database - both PostgreSQL and MySQL (Oracle databases are not currently checked)
- backend redis - both, queues and storage databases will be checked
- system redis

Once the verification is successful the operator will annotate the APIManager with "apps.3scale.net/apimanager-confirmed-requirements-version"
and the resource version of the config map. 

If the verification fails, the operator will not continue with the installation or upgrade.

When a new upgrade of 3scale Operator is applied on top of existing installation, preflight checks will perform the database version check to ensure that the upgrade 
version requirements are met before allowing the upgrade to proceed. If the upgrade requirements are not met, the operator will continue to work as normal, but will not 
allow the upgrade and will check the databases versions every 10 minutes or, on any other change to deployment or APIManager resource. Note that the operator will prevent
the upgrade even if the upgrade is approved by the user until the requirements are confirmed.

Preflight checks will also prevent multi-minor version hops which 3scale Operator does not support. For example, it's not allowed to go from 2.14 to 2.16 in a single hop.
In the event of this happening, the user will have to revert back to the previous version of the operator and follow supported upgrade path.

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
Whether Pod Disruption Budgets are enabled for non-database Deployments

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

### 3scale installation Backup and Restore
* [3scale installation Backup and Restore](operator-backup-and-restore.md)

### Application Capabilities
* [Application Capabilities](operator-application-capabilities.md)

### APIManager CRD reference
* [APIManager CRD reference](apimanager-reference.md)

#### CR Samples

* [\[1\]](../config/samples/apps_v1alpha1_apimanager_simple.yaml)
* [\[2\]](cr_samples/apimanager/)
