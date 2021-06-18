# CustomCustomPolicyDefinitionDefinition CRD Reference

## Table of Contents

* [CustomCustomPolicyDefinitionDefinition CRD Reference](#customcustompolicydefinitiondefinition-crd-reference)
   * [Table of Contents](#table-of-contents)
   * [CustomPolicyDefinition](#custompolicydefinition)
      * [CustomPolicyDefinitionSpec](#custompolicydefinitionspec)
         * [CustomPolicyDefinitionSchemaSpec](#custompolicydefinitionschemaspec)
         * [Provider Account Reference](#provider-account-reference)
      * [CustomPolicyDefinitionStatus](#custompolicydefinitionstatus)
         * [ConditionSpec](#conditionspec)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## CustomPolicyDefinition

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Spec | `spec` | [CustomPolicyDefinitionSpec](#custompolicydefinitionspec) | The specfication for the custom resource |
| Status | `status` | [CustomPolicyDefinitionStatus](#custompolicydefinitionstatus) | The status for the custom resource |

### CustomPolicyDefinitionSpec

`.spec`

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Name | **Yes** |
| Version | `version` | string | Version | **Yes** |
| Schema | `schema` | [CustomPolicyDefinitionSchemaSpec](#custompolicydefinitionschemaspec) | CustomPolicyDefinition schema definition | **Yes** |
| Provider Account Reference | `providerAccountRef` | object | [Provider account credentials secret reference](#provider-account-reference) | No |

Example:

```
apiVersion: capabilities.3scale.net/v1beta1
kind: CustomPolicyDefinition
metadata:
 name: custompolicydefinition-sample
spec:
  name: "MyCustomPolicy"
  version: "0.0.1"
  schema:
    name: "MyCustomPolicy"
    version: "0.0.1"
    summary: "some summary"
    $schema: "http://json-schema.org/draft-07/schema#"
    configuration:
      type: "object"
      properties:
        someAttr:
            description: "Some attribute"
            type: "integer"
```

#### CustomPolicyDefinitionSchemaSpec

`.spec.schema`

| **Field** | **json field**| **Type** | **Info** | **Required** |
| --- | --- | --- | --- | --- |
| Name | `name` | string | Schema Name | **Yes** |
| Version | `version` | string | Schema Version | **Yes** |
| Summary | `summary` | string | Schema Summary | **Yes** |
| Schema | `$schema` | string | `$schema` keyword is used to declare that this is a JSON Schema. Check [spec doc](https://json-schema.org/draft/2019-09/json-schema-core.html#rfc.section.8.1.1) for more info. | **Yes** |
| Description | `description` | array of string | Schema Description | No |
| Configuration | `configuration` | object | Schema configuration object | Yes. Minimum required is the empty object `{}` |

#### Provider Account Reference

Provider account credentials secret referenced by a [v1.LocalObjectReference](https://v1-15.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) type object.

The secret must have `adminURL` and `token` fields with tenant credentials.
Tenant controller will fetch the secret and read the following fields:

| **Field** | **Description** | **Required** |
| --- | --- | --- |
| *token* | Provider account access token with *Account Management API* scope and *Read & Write* permission | Yes |
| *adminURL* | Provider account's domain URL | Yes |

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: mytenant
type: Opaque
stringData:
  adminURL: https://my3scale-admin.example.com:443
  token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

### CustomPolicyDefinitionStatus

`.status`

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| ID | `policyID` | string | Internal 3scale ID |
| ProviderAccountHost | `providerAccountHost` | string | 3scale account's provider URL |
| Observed Generation | `observedGeneration` | string | helper field to see if status info is up to date with latest resource spec |
| Conditions | `conditions` | array of [condition](#ConditionSpec)s | resource conditions |

For example:

```
status:
  conditions:
  - lastTransitionTime: "2020-12-10T17:12:29Z"
    status: "False"
    type: Failed
  - lastTransitionTime: "2020-12-10T17:12:29Z"
    status: "False"
    type: Invalid
  - lastTransitionTime: "2020-12-10T17:12:29Z"
    status: "True"
    type: Ready
  observedGeneration: 1
  policyID: 20
  providerAccountHost: https://3scale.example.com
```

#### ConditionSpec

The status object has an array of Conditions through which the Product has or has not passed.
Each element of the Condition array has the following fields:

* The *lastTransitionTime* field provides a timestamp for when the entity last transitioned from one status to another.
* The *message* field is a human-readable message indicating details about the transition.
* The *reason* field is a unique, one-word, CamelCase reason for the condition’s last transition.
* The *status* field is a string, with possible values **True**, **False**, and **Unknown**.
* The *type* field is a string with the following possible values:
  * Invalid: Indicates that the combination of configuration in the CustomPolicyDefinitionSpec is not supported. This is not a transient error, but indicates a state that must be fixed before progress can be made;
  * Ready: Indicates the CustomPolicyDefinition resource has been successfully reconciled;
  * Failed: Indicates that an error occurred during reconcilliation;

| **Field** | **json field**| **Type** | **Info** |
| --- | --- | --- | --- |
| Type | `type` | string | Condition Type |
| Status | `status` | string | Status: True, False, Unknown |
| Reason | `reason` | string | Condition state reason |
| Message | `message` | string | Condition state description |
| LastTransitionTime | `lastTransitionTime` | timestamp | Last transition timestap |
