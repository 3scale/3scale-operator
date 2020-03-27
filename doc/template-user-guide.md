# User Guide: Deploy 3scale using templates

## Introduction

Several deployment profiles are available

Profile Goals:
* **Eval**:
  * Small memory footprint
  * Fast startup
  * Runnable on laptop
  * Suitable for presale/sales demos
* **Default**:
  * 3scale works out of the box with no additional installs
    * The customer has to make sure a RWX already exists in the environment.
    * The user has to make sure the other required volumes are available.
  * Suitable for PoC or evaluation by a customer
* **External DB** (also known as HA):
  * Suitable for production use where customer wants HA or to re-use DB of their own
* **S3**
  * Same as **Default** profile, but with System’s FileStorage being in S3 instead of in a PVC (including some changes in parameters of the template)

## Deployment profile Index
* [Default](#default)
* [Eval](#eval)
* [S3](#s3)
* [External Databases](#external-databases)
* [Eval S3](#eval-s3)
* [Default Postgresql](#default-postgresql)

## Default

#### Features

3scale works out of the box with no additional installs

#### Template file
```
pkg/3scale/amp/auto-generated-templates/amp/amp.yml
```

#### Deploy template
```
oc new-app --file pkg/3scale/amp/auto-generated-templates/amp/amp.yml \
           --param WILDCARD_DOMAIN=lvh.me
```

#### Required Parameters
| Parameter Name | Description | Example |
| :--- | :---| :--- |
| **WILDCARD_DOMAIN** | Root domain for the wildcard routes | example.com |

#### Configurable

| Parameter Name | Description | Default |
| :--- | :---| :--- |
| **AMP_RELEASE** | AMP release tag | 2.8 |
| **APP_LABEL** | Used for object app labels | 3scale-api-management |
| **TENANT_NAME** | Default tenant prefix name. *-admin* suffix will be appended | 3scale |
| **RWX_STORAGE_CLASS** | The Storage Class to be used by ReadWriteMany PVCs | 'null' |
| **AMP_BACKEND_IMAGE** | 3scale Backend component docker image URL | quay.io/3scale/3scale28:apisonator-3scale-2.8.0-GA |
| **AMP_ZYNC_IMAGE** | 3scale Zync component docker image URL | quay.io/3scale/3scale28:zync-3scale-2.8.0-GA |
| **AMP_APICAST_IMAGE** | 3scale Apicast component docker image URL | quay.io/3scale/3scale28:apicast-3scale-2.8.0-GA |
| **AMP_SYSTEM_IMAGE** | 3scale System component docker image URL | quay.io/3scale/3scale28:porta-3scale-2.8.0-GA |
| **ZYNC_DATABASE_IMAGE** | Zync's PostgreSQL image to use | centos/postgresql-10-centos7 |
| **MEMCACHED_IMAGE** | Memcached image to use | memcached:1.5 |
| **IMAGESTREAM_TAG_IMPORT_INSECURE** | the server may bypass certificate verification | false |
| **SYSTEM_DATABASE_IMAGE** | System MySQL image URL | centos/mysql-57-centos7 |
| **REDIS_IMAGE** | Redis image to use | centos/redis-32-centos7 |
| **SYSTEM_DATABASE_USER** | System MySQL User | mysql |
| **SYSTEM_DATABASE_PASSWORD** | System MySQL Password | random value |
| **SYSTEM_DATABASE** | System MySQL Database Name | system |
| **SYSTEM_DATABASE_ROOT_PASSWORD** | System MySQL Root password | random value |
| **SYSTEM_BACKEND_USERNAME** | Internal 3scale API username for internal 3scale api auth | 3scale_api_user |
| **SYSTEM_BACKEND_PASSWORD** | Internal 3scale API password for internal 3scale api auth | random value |
| **SYSTEM_BACKEND_SHARED_SECRET** | Shared secret to import events from backend to system | random value |
| **SYSTEM_APP_SECRET_KEY_BASE** | System application secret key base | random value |
| **ADMIN_PASSWORD** | Default 3scale tenant administrator account password | random value |
| **ADMIN_USERNAME** | Default 3scale tenant administrator account username | admin |
| **ADMIN_EMAIL** | Default 3scale tenant administrator account email | - |
| **ADMIN_ACCESS_TOKEN** | Default 3scale tenant administrator account access token | random value |
| **MASTER_NAME** | 3scale _MASTER_ account name | master |
| **MASTER_USER** | 3scale _MASTER_ account administrator username | master |
| **MASTER_PASSWORD** | 3scale _MASTER_ account administrator password | random value |
| **MASTER_ACCESS_TOKEN** | 3scale _MASTER_ account administrator access token | random value |
| **RECAPTCHA_PUBLIC_KEY** | reCAPTCHA site key (used in spam protection) | - |
| **RECAPTCHA_PRIVATE_KEY** | reCAPTCHA secret key (used in spam protection) | - |
| **SYSTEM_REDIS_URL** | Define the external system-redis to connect to | redis://system-redis:6379/1 |
| **SYSTEM_MESSAGE_BUS_REDIS_URL** | Define the external system-redis message bus to connect to | see note<sup>[1](#note1)</sup> |
| **SYSTEM_REDIS_NAMESPACE** | namespace to be used by System's Redis Database | none |
| **SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE** | namespace to be used by System's Message Bus Redis Database | none |
| **ZYNC_DATABASE_PASSWORD** | Zync Database PostgreSQL Connection Password | random value |
| **ZYNC_SECRET_KEY_BASE** | Zync application secret key base | random value |
| **ZYNC_AUTHENTICATION_TOKEN** | Zync application authentication token | random value |
| **APICAST_ACCESS_TOKEN** | Read Only Access Token that is APIcast going to use to download its configuration | random value |
| **APICAST_MANAGEMENT_API** | Scope of the APIcast Management API. Can be disabled, status or debug | status |
| **APICAST_OPENSSL_VERIFY** | OpenSSL peer verification when downloading the configuration | false |
| **APICAST_RESPONSE_CODES** | Enable logging response codes in APIcast | true |
| **APICAST_REGISTRY_URL** | The URL to point to APIcast policies registry management | http://apicast-staging:8090/policies |

## Eval

#### Features

Memory/CPU resource limits/requests specification for the DeploymentConfigs removed

#### Template file
```
pkg/3scale/amp/auto-generated-templates/amp/amp-eval.yml
```

#### Deploy template
```
oc new-app --file pkg/3scale/amp/auto-generated-templates/amp/amp-eval.yml \
           --param WILDCARD_DOMAIN=lvh.me
```

#### Required Parameters
| Parameter Name | Description | Example |
| :--- | :---| :--- |
| **WILDCARD_DOMAIN** | Root domain for the wildcard routes | example.com |

#### Configurable

| Parameter Name | Description | Default |
| :--- | :---| :--- |
| **AMP_RELEASE** | AMP release tag | 2.8 |
| **APP_LABEL** | Used for object app labels | 3scale-api-management |
| **TENANT_NAME** | Default tenant prefix name. *-admin* suffix will be appended | 3scale |
| **RWX_STORAGE_CLASS** | The Storage Class to be used by ReadWriteMany PVCs | 'null' |
| **AMP_BACKEND_IMAGE** | 3scale Backend component docker image URL | quay.io/3scale/3scale28:apisonator-3scale-2.8.0-GA |
| **AMP_ZYNC_IMAGE** | 3scale Zync component docker image URL | quay.io/3scale/3scale28:zync-3scale-2.8.0-GA |
| **AMP_APICAST_IMAGE** | 3scale Apicast component docker image URL | quay.io/3scale/3scale28:apicast-3scale-2.8.0-GA |
| **AMP_SYSTEM_IMAGE** | 3scale System component docker image URL | quay.io/3scale/3scale28:porta-3scale-2.8.0-GA |
| **ZYNC_DATABASE_IMAGE** | Zync's PostgreSQL image to use | centos/postgresql-10-centos7 |
| **MEMCACHED_IMAGE** | Memcached image to use | memcached:1.5 |
| **IMAGESTREAM_TAG_IMPORT_INSECURE** | the server may bypass certificate verification | false |
| **SYSTEM_DATABASE_IMAGE** | System PostgreSQL image URL | centos/postgresql-10-centos7 |
| **REDIS_IMAGE** | Redis image to use | centos/redis-32-centos7 |
| **SYSTEM_DATABASE_USER** | System PostgreSQL User | system |
| **SYSTEM_DATABASE_PASSWORD** | System PostgreSQL Password | random value |
| **SYSTEM_DATABASE** | System PostgreSQL Database Name | system |
| **SYSTEM_BACKEND_USERNAME** | Internal 3scale API username for internal 3scale api auth | 3scale_api_user |
| **SYSTEM_BACKEND_PASSWORD** | Internal 3scale API password for internal 3scale api auth | random value |
| **SYSTEM_BACKEND_SHARED_SECRET** | Shared secret to import events from backend to system | random value |
| **SYSTEM_APP_SECRET_KEY_BASE** | System application secret key base | random value |
| **ADMIN_PASSWORD** | Default 3scale tenant administrator account password | random value |
| **ADMIN_USERNAME** | Default 3scale tenant administrator account username | admin |
| **ADMIN_EMAIL** | Default 3scale tenant administrator account email | - |
| **ADMIN_ACCESS_TOKEN** | Default 3scale tenant administrator account access token | random value |
| **MASTER_NAME** | 3scale _MASTER_ account name | master |
| **MASTER_USER** | 3scale _MASTER_ account administrator username | master |
| **MASTER_PASSWORD** | 3scale _MASTER_ account administrator password | random value |
| **MASTER_ACCESS_TOKEN** | 3scale _MASTER_ account administrator access token | random value |
| **RECAPTCHA_PUBLIC_KEY** | reCAPTCHA site key (used in spam protection) | - |
| **RECAPTCHA_PRIVATE_KEY** | reCAPTCHA secret key (used in spam protection) | - |
| **SYSTEM_REDIS_URL** | Define the external system-redis to connect to | redis://system-redis:6379/1 |
| **SYSTEM_MESSAGE_BUS_REDIS_URL** | Define the external system-redis message bus to connect to | see note<sup>[1](#note1)</sup> |
| **SYSTEM_REDIS_NAMESPACE** | namespace to be used by System's Redis Database | none |
| **SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE** | namespace to be used by System's Message Bus Redis Database | none |
| **ZYNC_DATABASE_PASSWORD** | Zync Database PostgreSQL Connection Password | random value |
| **ZYNC_SECRET_KEY_BASE** | Zync application secret key base | random value |
| **ZYNC_AUTHENTICATION_TOKEN** | Zync application authentication token | random value |
| **APICAST_ACCESS_TOKEN** | Read Only Access Token that is APIcast going to use to download its configuration | random value |
| **APICAST_MANAGEMENT_API** | Scope of the APIcast Management API. Can be disabled, status or debug | status |
| **APICAST_OPENSSL_VERIFY** | OpenSSL peer verification when downloading the configuration | false |
| **APICAST_RESPONSE_CODES** | Enable logging response codes in APIcast | true |
| **APICAST_REGISTRY_URL** | The URL to point to APIcast policies registry management | http://apicast-staging:8090/policies |

## S3

#### Features

System’s FileStorage being in S3 instead of in a PVC. No need for RWX persistence volume provisioning.

#### Template file
```
pkg/3scale/amp/auto-generated-templates/amp/amp-s3.yml
```

#### Deploy template
```
oc new-app --file pkg/3scale/amp/auto-generated-templates/amp/amp-s3.yml \
           --param AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY_ID \
           --param AWS_SECRET_ACCESS_KEY=YOUR_SECRET_ACCESS_KEY \
           --param AWS_BUCKET=YOUR_AWS_BUCKET \
           --param AWS_REGION=YOUR_AWS_REGION \
           --param WILDCARD_DOMAIN=lvh.me
```

#### Required Parameters
| Parameter Name | Description | Example |
| :--- | :---| :--- |
| **WILDCARD_DOMAIN** | Root domain for the wildcard routes | example.com |
| **AWS_ACCESS_KEY_ID** | AWS Access Key ID to use in S3 Storage for assets | - |
| **AWS_SECRET_ACCESS_KEY** | AWS Access Key Secret to use in S3 Storage for assets | - |
| **AWS_BUCKET** | AWS S3 Bucket Name to use in S3 Storage for assets | - |
| **AWS_REGION** | AWS Region to use in S3 Storage for assets | eu-west-1 |

#### Configurable

| Parameter Name | Description | Default |
| :--- | :---| :--- |
| **AMP_RELEASE** | AMP release tag | 2.8 |
| **APP_LABEL** | Used for object app labels | 3scale-api-management |
| **TENANT_NAME** | Default tenant prefix name. *-admin* suffix will be appended | 3scale |
| **AMP_BACKEND_IMAGE** | 3scale Backend component docker image URL | quay.io/3scale/3scale28:apisonator-3scale-2.8.0-GA |
| **AMP_ZYNC_IMAGE** | 3scale Zync component docker image URL | quay.io/3scale/3scale28:zync-3scale-2.8.0-GA |
| **AMP_APICAST_IMAGE** | 3scale Apicast component docker image URL | quay.io/3scale/3scale28:apicast-3scale-2.8.0-GA |
| **AMP_SYSTEM_IMAGE** | 3scale System component docker image URL | quay.io/3scale/3scale28:porta-3scale-2.8.0-GA |
| **ZYNC_DATABASE_IMAGE** | Zync's PostgreSQL image to use | centos/postgresql-10-centos7 |
| **MEMCACHED_IMAGE** | Memcached image to use | memcached:1.5 |
| **IMAGESTREAM_TAG_IMPORT_INSECURE** | the server may bypass certificate verification | false |
| **SYSTEM_DATABASE_IMAGE** | System MySQL image URL | centos/mysql-57-centos7 |
| **REDIS_IMAGE** | Redis image to use | centos/redis-32-centos7 |
| **SYSTEM_DATABASE_USER** | System MySQL User | mysql |
| **SYSTEM_DATABASE_PASSWORD** | System MySQL Password | random value |
| **SYSTEM_DATABASE** | System MySQL Database Name | system |
| **SYSTEM_DATABASE_ROOT_PASSWORD** | System MySQL Root password | random value |
| **SYSTEM_BACKEND_USERNAME** | Internal 3scale API username for internal 3scale api auth | 3scale_api_user |
| **SYSTEM_BACKEND_PASSWORD** | Internal 3scale API password for internal 3scale api auth | random value |
| **SYSTEM_BACKEND_SHARED_SECRET** | Shared secret to import events from backend to system | random value |
| **SYSTEM_APP_SECRET_KEY_BASE** | System application secret key base | random value |
| **ADMIN_PASSWORD** | Default 3scale tenant administrator account password | random value |
| **ADMIN_USERNAME** | Default 3scale tenant administrator account username | admin |
| **ADMIN_EMAIL** | Default 3scale tenant administrator account email | - |
| **ADMIN_ACCESS_TOKEN** | Default 3scale tenant administrator account access token | random value |
| **MASTER_NAME** | 3scale _MASTER_ account name | master |
| **MASTER_USER** | 3scale _MASTER_ account administrator username | master |
| **MASTER_PASSWORD** | 3scale _MASTER_ account administrator password | random value |
| **MASTER_ACCESS_TOKEN** | 3scale _MASTER_ account administrator access token | random value |
| **RECAPTCHA_PUBLIC_KEY** | reCAPTCHA site key (used in spam protection) | - |
| **RECAPTCHA_PRIVATE_KEY** | reCAPTCHA secret key (used in spam protection) | - |
| **SYSTEM_REDIS_URL** | Define the external system-redis to connect to | redis://system-redis:6379/1 |
| **SYSTEM_MESSAGE_BUS_REDIS_URL** | Define the external system-redis message bus to connect to | see note<sup>[1](#note1)</sup> |
| **SYSTEM_REDIS_NAMESPACE** | namespace to be used by System's Redis Database | none |
| **SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE** | namespace to be used by System's Message Bus Redis Database | none |
| **ZYNC_DATABASE_PASSWORD** | Zync Database PostgreSQL Connection Password | random value |
| **ZYNC_SECRET_KEY_BASE** | Zync application secret key base | random value |
| **ZYNC_AUTHENTICATION_TOKEN** | Zync application authentication token | random value |
| **APICAST_ACCESS_TOKEN** | Read Only Access Token that is APIcast going to use to download its configuration | random value |
| **APICAST_MANAGEMENT_API** | Scope of the APIcast Management API. Can be disabled, status or debug | status |
| **APICAST_OPENSSL_VERIFY** | OpenSSL peer verification when downloading the configuration | false |
| **APICAST_RESPONSE_CODES** | Enable logging response codes in APIcast | true |
| **APICAST_REGISTRY_URL** | The URL to point to APIcast policies registry management | http://apicast-staging:8090/policies |

## External Databases

#### Features

3scale deployment will rely on externally managed databases.
Suitable for production use where customer wants high availability or to re-use DB of their own

#### Template file
```
pkg/3scale/amp/auto-generated-templates/amp/amp-ha.yml
```

#### Deploy template
```
oc new-app --file pkg/3scale/amp/auto-generated-templates/amp/amp-ha.yml \
           --param BACKEND_REDIS_STORAGE_ENDPOINT=redis://backend-redis:6379/0 \
           --param BACKEND_REDIS_QUEUES_ENDPOINT=redis://backend-redis:6379/1 \
           --param SYSTEM_DATABASE_URL=mysql2://root:password1@system-mysql/system \
           --param SYSTEM_REDIS_URL=redis://system-redis:6379/0 \
           --param SYSTEM_MESSAGE_BUS_REDIS_URL=redis://system-redis:6379/1 \
           --param WILDCARD_DOMAIN=lvh.me
```

#### Required Parameters
| Parameter Name | Description | Example |
| :--- | :---| :--- |
| **SYSTEM_REDIS_URL** | Define the external system-redis to connect to | redis://system-redis:6379/0 |
| **SYSTEM_MESSAGE_BUS_REDIS_URL** | Define the external system-redis message bus to connect to | redis://system-redis:6379/1 |
| **SYSTEM_DATABASE_URL** | Define the external system-mysql to connect to | mysql2://root:password1@system-mysql/system |
| **BACKEND_REDIS_STORAGE_ENDPOINT** | Define the external backend-redis storage endpoint to connect to | redis://backend-redis:6379/0 |
| **BACKEND_REDIS_QUEUES_ENDPOINT** | Define the external backend-redis queues endpoint to connect to | redis://backend-redis:6379/1 |
| **WILDCARD_DOMAIN** | Root domain for the wildcard routes | example.com |

#### Configurable

| Parameter Name | Description | Default |
| :--- | :---| :--- |
| **AMP_RELEASE** | AMP release tag | 2.8 |
| **APP_LABEL** | Used for object app labels | 3scale-api-management |
| **TENANT_NAME** | Default tenant prefix name. *-admin* suffix will be appended | 3scale |
| **RWX_STORAGE_CLASS** | The Storage Class to be used by ReadWriteMany PVCs | 'null' |
| **AMP_BACKEND_IMAGE** | 3scale Backend component docker image URL | quay.io/3scale/3scale28:apisonator-3scale-2.8.0-GA |
| **AMP_ZYNC_IMAGE** | 3scale Zync component docker image URL | quay.io/3scale/3scale28:zync-3scale-2.8.0-GA |
| **AMP_APICAST_IMAGE** | 3scale Apicast component docker image URL | quay.io/3scale/3scale28:apicast-3scale-2.8.0-GA |
| **AMP_SYSTEM_IMAGE** | 3scale System component docker image URL | quay.io/3scale/3scale28:porta-3scale-2.8.0-GA |
| **ZYNC_DATABASE_IMAGE** | Zync's PostgreSQL image to use | centos/postgresql-10-centos7 |
| **MEMCACHED_IMAGE** | Memcached image to use | memcached:1.5 |
| **IMAGESTREAM_TAG_IMPORT_INSECURE** | the server may bypass certificate verification | false |
| **SYSTEM_BACKEND_USERNAME** | Internal 3scale API username for internal 3scale api auth | 3scale_api_user |
| **SYSTEM_BACKEND_PASSWORD** | Internal 3scale API password for internal 3scale api auth | random value |
| **SYSTEM_BACKEND_SHARED_SECRET** | Shared secret to import events from backend to system | random value |
| **SYSTEM_APP_SECRET_KEY_BASE** | System application secret key base | random value |
| **ADMIN_PASSWORD** | Default 3scale tenant administrator account password | random value |
| **ADMIN_USERNAME** | Default 3scale tenant administrator account username | admin |
| **ADMIN_EMAIL** | Default 3scale tenant administrator account email | - |
| **ADMIN_ACCESS_TOKEN** | Default 3scale tenant administrator account access token | random value |
| **MASTER_NAME** | 3scale _MASTER_ account name | master |
| **MASTER_USER** | 3scale _MASTER_ account administrator username | master |
| **MASTER_PASSWORD** | 3scale _MASTER_ account administrator password | random value |
| **MASTER_ACCESS_TOKEN** | 3scale _MASTER_ account administrator access token | random value |
| **RECAPTCHA_PUBLIC_KEY** | reCAPTCHA site key (used in spam protection) | - |
| **RECAPTCHA_PRIVATE_KEY** | reCAPTCHA secret key (used in spam protection) | - |
| **SYSTEM_REDIS_NAMESPACE** | namespace to be used by System's Redis Database | none |
| **SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE** | namespace to be used by System's Message Bus Redis Database | none |
| **ZYNC_DATABASE_PASSWORD** | Zync Database PostgreSQL Connection Password | random value |
| **ZYNC_SECRET_KEY_BASE** | Zync application secret key base | random value |
| **ZYNC_AUTHENTICATION_TOKEN** | Zync application authentication token | random value |
| **APICAST_ACCESS_TOKEN** | Read Only Access Token that is APIcast going to use to download its configuration | random value |
| **APICAST_MANAGEMENT_API** | Scope of the APIcast Management API. Can be disabled, status or debug | status |
| **APICAST_OPENSSL_VERIFY** | OpenSSL peer verification when downloading the configuration | false |
| **APICAST_RESPONSE_CODES** | Enable logging response codes in APIcast | true |
| **APICAST_REGISTRY_URL** | The URL to point to APIcast policies registry management | http://apicast-staging:8090/policies |
| **SYSTEM_MESSAGE_BUS_REDIS_SENTINEL_HOSTS** | Define the external system message bus sentinel hosts | none |
| **SYSTEM_MESSAGE_BUS_REDIS_SENTINEL_ROLE** | Define the external system message bus sentinel role | none |
| **SYSTEM_REDIS_SENTINEL_HOSTS** | Define the external system redis sentinel hosts | none |
| **SYSTEM_REDIS_SENTINEL_ROLE** | Define the external system redis sentinel role | none |
| **BACKEND_REDIS_QUEUE_SENTINEL_HOSTS** | Define the external backend redis queue sentinel hosts | none |
| **BACKEND_REDIS_QUEUE_SENTINEL_ROLE** | Define the external backend redis queue sentinel role | none |
| **BACKEND_REDIS_STORAGE_SENTINEL_HOSTS** | Define the external backend redis storage sentinel hosts | none |
| **BACKEND_REDIS_STORAGE_SENTINEL_ROLE** | Define the external backend redis storage sentinel role | none |

## Eval S3

#### Features

* Memory/CPU resource limits/requests specification for the DeploymentConfigs removed
* System’s FileStorage being in S3 instead of in a PVC. No need for RWX persistence volume provisioning.

#### Template file
```
pkg/3scale/amp/auto-generated-templates/amp/amp-eval-s3.yml
```

#### Deploy template
```
oc new-app --file pkg/3scale/amp/auto-generated-templates/amp/amp-eval-s3.yml \
           --param AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY_ID \
           --param AWS_SECRET_ACCESS_KEY=YOUR_SECRET_ACCESS_KEY \
           --param AWS_BUCKET=YOUR_AWS_BUCKET \
           --param AWS_REGION=YOUR_AWS_REGION \
           --param WILDCARD_DOMAIN=lvh.me
```

#### Required Parameters
| Parameter Name | Description | Example |
| :--- | :---| :--- |
| **WILDCARD_DOMAIN** | Root domain for the wildcard routes | example.com |
| **AWS_ACCESS_KEY_ID** | AWS Access Key ID to use in S3 Storage for assets | - |
| **AWS_SECRET_ACCESS_KEY** | AWS Access Key Secret to use in S3 Storage for assets | - |
| **AWS_BUCKET** | AWS S3 Bucket Name to use in S3 Storage for assets | - |
| **AWS_REGION** | AWS Region to use in S3 Storage for assets | eu-west-1 |

#### Configurable

| Parameter Name | Description | Default |
| :--- | :---| :--- |
| **AMP_RELEASE** | AMP release tag | 2.8 |
| **APP_LABEL** | Used for object app labels | 3scale-api-management |
| **TENANT_NAME** | Default tenant prefix name. *-admin* suffix will be appended | 3scale |
| **AMP_BACKEND_IMAGE** | 3scale Backend component docker image URL | quay.io/3scale/3scale28:apisonator-3scale-2.8.0-GA |
| **AMP_ZYNC_IMAGE** | 3scale Zync component docker image URL | quay.io/3scale/3scale28:zync-3scale-2.8.0-GA |
| **AMP_APICAST_IMAGE** | 3scale Apicast component docker image URL | quay.io/3scale/3scale28:apicast-3scale-2.8.0-GA |
| **AMP_SYSTEM_IMAGE** | 3scale System component docker image URL | quay.io/3scale/3scale28:porta-3scale-2.8.0-GA |
| **ZYNC_DATABASE_IMAGE** | Zync's PostgreSQL image to use | centos/postgresql-10-centos7 |
| **MEMCACHED_IMAGE** | Memcached image to use | memcached:1.5 |
| **IMAGESTREAM_TAG_IMPORT_INSECURE** | the server may bypass certificate verification | false |
| **SYSTEM_DATABASE_IMAGE** | System MySQL image URL | centos/mysql-57-centos7 |
| **REDIS_IMAGE** | Redis image to use | centos/redis-32-centos7 |
| **SYSTEM_DATABASE_USER** | System MySQL User | mysql |
| **SYSTEM_DATABASE_PASSWORD** | System MySQL Password | random value |
| **SYSTEM_DATABASE** | System MySQL Database Name | system |
| **SYSTEM_DATABASE_ROOT_PASSWORD** | System MySQL Root password | random value |
| **SYSTEM_BACKEND_USERNAME** | Internal 3scale API username for internal 3scale api auth | 3scale_api_user |
| **SYSTEM_BACKEND_PASSWORD** | Internal 3scale API password for internal 3scale api auth | random value |
| **SYSTEM_BACKEND_SHARED_SECRET** | Shared secret to import events from backend to system | random value |
| **SYSTEM_APP_SECRET_KEY_BASE** | System application secret key base | random value |
| **ADMIN_PASSWORD** | Default 3scale tenant administrator account password | random value |
| **ADMIN_USERNAME** | Default 3scale tenant administrator account username | admin |
| **ADMIN_EMAIL** | Default 3scale tenant administrator account email | - |
| **ADMIN_ACCESS_TOKEN** | Default 3scale tenant administrator account access token | random value |
| **MASTER_NAME** | 3scale _MASTER_ account name | master |
| **MASTER_USER** | 3scale _MASTER_ account administrator username | master |
| **MASTER_PASSWORD** | 3scale _MASTER_ account administrator password | random value |
| **MASTER_ACCESS_TOKEN** | 3scale _MASTER_ account administrator access token | random value |
| **RECAPTCHA_PUBLIC_KEY** | reCAPTCHA site key (used in spam protection) | - |
| **RECAPTCHA_PRIVATE_KEY** | reCAPTCHA secret key (used in spam protection) | - |
| **SYSTEM_REDIS_URL** | Define the external system-redis to connect to | redis://system-redis:6379/1 |
| **SYSTEM_MESSAGE_BUS_REDIS_URL** | Define the external system-redis message bus to connect to | see note<sup>[1](#note1)</sup> |
| **SYSTEM_REDIS_NAMESPACE** | namespace to be used by System's Redis Database | none |
| **SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE** | namespace to be used by System's Message Bus Redis Database | none |
| **ZYNC_DATABASE_PASSWORD** | Zync Database PostgreSQL Connection Password | random value |
| **ZYNC_SECRET_KEY_BASE** | Zync application secret key base | random value |
| **ZYNC_AUTHENTICATION_TOKEN** | Zync application authentication token | random value |
| **APICAST_ACCESS_TOKEN** | Read Only Access Token that is APIcast going to use to download its configuration | random value |
| **APICAST_MANAGEMENT_API** | Scope of the APIcast Management API. Can be disabled, status or debug | status |
| **APICAST_OPENSSL_VERIFY** | OpenSSL peer verification when downloading the configuration | false |
| **APICAST_RESPONSE_CODES** | Enable logging response codes in APIcast | true |
| **APICAST_REGISTRY_URL** | The URL to point to APIcast policies registry management | http://apicast-staging:8090/policies |

## Default Postgresql

#### Features

Same profile as **Default** profile, but **PostgreSQL** will be used as internal _System_ database.

#### Template file
```
pkg/3scale/amp/auto-generated-templates/amp/amp-postgresql.yml
```

#### Deploy template
```
oc new-app --file pkg/3scale/amp/auto-generated-templates/amp/amp-postgresql.yml \
           --param WILDCARD_DOMAIN=lvh.me
```

#### Required Parameters
| Parameter Name | Description | Example |
| :--- | :---| :--- |
| **WILDCARD_DOMAIN** | Root domain for the wildcard routes | example.com |

#### Configurable

| Parameter Name | Description | Default |
| :--- | :---| :--- |
| **AMP_RELEASE** | AMP release tag | 2.8 |
| **APP_LABEL** | Used for object app labels | 3scale-api-management |
| **TENANT_NAME** | Default tenant prefix name. *-admin* suffix will be appended | 3scale |
| **RWX_STORAGE_CLASS** | The Storage Class to be used by ReadWriteMany PVCs | 'null' |
| **AMP_BACKEND_IMAGE** | 3scale Backend component docker image URL | quay.io/3scale/3scale28:apisonator-3scale-2.8.0-GA |
| **AMP_ZYNC_IMAGE** | 3scale Zync component docker image URL | quay.io/3scale/3scale28:zync-3scale-2.8.0-GA |
| **AMP_APICAST_IMAGE** | 3scale Apicast component docker image URL | quay.io/3scale/3scale28:apicast-3scale-2.8.0-GA |
| **AMP_SYSTEM_IMAGE** | 3scale System component docker image URL | quay.io/3scale/3scale28:porta-3scale-2.8.0-GA |
| **ZYNC_DATABASE_IMAGE** | Zync's PostgreSQL image to use | centos/postgresql-10-centos7 |
| **MEMCACHED_IMAGE** | Memcached image to use | memcached:1.5 |
| **IMAGESTREAM_TAG_IMPORT_INSECURE** | the server may bypass certificate verification | false |
| **SYSTEM_DATABASE_IMAGE** | System MySQL image URL | centos/mysql-57-centos7 |
| **REDIS_IMAGE** | Redis image to use | centos/redis-32-centos7 |
| **SYSTEM_DATABASE_USER** | System MySQL User | mysql |
| **SYSTEM_DATABASE_PASSWORD** | System MySQL Password | random value |
| **SYSTEM_DATABASE** | System MySQL Database Name | system |
| **SYSTEM_DATABASE_ROOT_PASSWORD** | System MySQL Root password | random value |
| **SYSTEM_BACKEND_USERNAME** | Internal 3scale API username for internal 3scale api auth | 3scale_api_user |
| **SYSTEM_BACKEND_PASSWORD** | Internal 3scale API password for internal 3scale api auth | random value |
| **SYSTEM_BACKEND_SHARED_SECRET** | Shared secret to import events from backend to system | random value |
| **SYSTEM_APP_SECRET_KEY_BASE** | System application secret key base | random value |
| **ADMIN_PASSWORD** | Default 3scale tenant administrator account password | random value |
| **ADMIN_USERNAME** | Default 3scale tenant administrator account username | admin |
| **ADMIN_EMAIL** | Default 3scale tenant administrator account email | - |
| **ADMIN_ACCESS_TOKEN** | Default 3scale tenant administrator account access token | random value |
| **MASTER_NAME** | 3scale _MASTER_ account name | master |
| **MASTER_USER** | 3scale _MASTER_ account administrator username | master |
| **MASTER_PASSWORD** | 3scale _MASTER_ account administrator password | random value |
| **MASTER_ACCESS_TOKEN** | 3scale _MASTER_ account administrator access token | random value |
| **RECAPTCHA_PUBLIC_KEY** | reCAPTCHA site key (used in spam protection) | - |
| **RECAPTCHA_PRIVATE_KEY** | reCAPTCHA secret key (used in spam protection) | - |
| **SYSTEM_REDIS_URL** | Define the external system-redis to connect to | redis://system-redis:6379/1 |
| **SYSTEM_MESSAGE_BUS_REDIS_URL** | Define the external system-redis message bus to connect to | see note<sup>[1](#note1)</sup> |
| **SYSTEM_REDIS_NAMESPACE** | namespace to be used by System's Redis Database | none |
| **SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE** | namespace to be used by System's Message Bus Redis Database | none |
| **ZYNC_DATABASE_PASSWORD** | Zync Database PostgreSQL Connection Password | random value |
| **ZYNC_SECRET_KEY_BASE** | Zync application secret key base | random value |
| **ZYNC_AUTHENTICATION_TOKEN** | Zync application authentication token | random value |
| **APICAST_ACCESS_TOKEN** | Read Only Access Token that is APIcast going to use to download its configuration | random value |
| **APICAST_MANAGEMENT_API** | Scope of the APIcast Management API. Can be disabled, status or debug | status |
| **APICAST_OPENSSL_VERIFY** | OpenSSL peer verification when downloading the configuration | false |
| **APICAST_RESPONSE_CODES** | Enable logging response codes in APIcast | true |
| **APICAST_REGISTRY_URL** | The URL to point to APIcast policies registry management | http://apicast-staging:8090/policies |

## Notes
<a name="note1">1</a>: *SYSTEM_MESSAGE_BUS_REDIS_URL* by default is the same value as *SYSTEM_REDIS_URL* but with the logical database incremented by 1 and the result applied mod 16
