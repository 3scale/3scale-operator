# 3scale Operator

## Tenant CRD field reference

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [TenantSpec](#TenantSpec) | The specfication for Tenant custom resource |
| Status | `status` | [TenantStatus](#TenantStatus) | The status for the Tenant custom resource |

### TenantSpec

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Organization Name | `organizationName` | string | Organization Name | Yes |
| Email | `email` | string | Admin email address | Yes |
| Admin Username | `username` | string | Admin credentials: username | Yes |
| Master Account Domain URL | `systemMasterUrl` | string | Master Account URL | Yes |
| Master Account Credentials Secret | `masterCredentialsRef` | object | See [Master Secret](#Master-Secret) for more details | Yes |
| Admin Secret | `passwordCredentialsRef` | object | See [Admin Secret](#Admin-Secret) for more details | Yes |
| Tenant Credentials Secret | `tenantSecretRef` | object | See [Tenant Secret](#Tenant-Secret) for more details | No |

#### Master Secret
Tenants can be managed using master provider account credentials. This secret provides those credentials to the 3scale operator.

The credentials are tipically provided by [APIManager](operator-user-guide.md#Basic-installation)
and stored in the secret name [system-seed](apimanager-reference.md#system-seed).
If this is the case, `masterCredentialsRef` object should look like:

```yaml
masterCredentialsRef:
  name: system-seed
```

Tenant controller will fetch the secret and read the following fields:

| **Field** | **Description** |
| --- | --- |
| *MASTER_ACCESS_TOKEN* | Master provider account access token with *Account Management API* scope and *Read & Write* permission|

If secret needs to be created manually, can be defined in the following way:

```sh
$ cat ecorp-master-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: ecorp-master-secret
type: Opaque
stringData:
  MASTER_ACCESS_TOKEN: <master access token>

$ oc create -f ecorp-master-secret.yaml
secret/ecorp-master-secret created
```

then, `masterCredentialsRef` object should look like:

```yaml
masterCredentialsRef:
  name: ecorp-master-secret
```

#### Admin Secret

Tenant creation requires Admin username, email and password. The password will be provided as a secret and referenced by `passwordCredentialsRef` object.

*IMPORTANT* This *Admin Secret* has to be created *before* **Tenant custom resource** is created. Otherwise, **3scale operator** will complain.

Secret required fields:

| **Field** | **Description** |
| --- | --- |
| *admin_password* | Tenant admin user password value |

Admin secret needs to be created manually, can be defined in the following way:

```sh
$ cat ecorp-admin-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: ecorp-admin-secret
type: Opaque
stringData:
  admin_password: <admin password value>


$ oc create -f ecorp-admin-secret.yaml
secret/ecorp-admin-secret created
```

then, `passwordCredentialsRef` object should look like:

```yaml
passwordCredentialsRef:
  name: ecorp-admin-secret
```

#### Tenant Secret
When tenant has been created, tenant level credentials will be created to operate on that particular tenant.
Those credentials will be stored by **tenant controller** in a secret.
`tenantSecretRef` in tenant's spec will reference this specific secret.

Note that `tenantSecretRef` attribute is optional. If not provided by tenant custom resource spec,
**tenant controller** will try to store tenant credentials in a secret with the following default values for name and namespace:

```yaml
tenantSecretRef:
  name: ${tenantName}-${tenantOrgName}
  namespace: YOUR-CURRENT-NAMESPACE
```

Fields available in tenant secret:

| **Field** | **Description** |
| --- | --- |
| *token* | Tenant's provider key |
| *adminURL* | Tenant's admin domain URL |

### TenantStatus

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Admin User ID | `adminID` | string | Internal ID for the admin user |
| Tenant ID | `tenantID` | string | Internal ID for the provider account |
| Tenant Admin Domain URL | `adminURL` | string | Tenant's admin domain URL |

