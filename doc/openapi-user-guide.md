# OpenAPI Custom Resource

The [OpenAPI CRD](openapi-reference.md) is used as the source of truth to reconciliate
one [3scale Product custom resource](product-reference.md) and
one [3scale Backend custom resource](backend-reference.md).

## Table of contents

* [OpenAPI Custom Resource](#openapi-custom-resource)
   * [Table of contents](#table-of-contents)
   * [OpenAPI document sources](#openapi-document-sources)
      * [Secret OpenAPI spec source](#secret-openapi-spec-source)
      * [URL OpenAPI spec source](#url-openapi-spec-source)
   * [Supported OpenAPI spec version and limitations](#supported-openapi-spec-version-and-limitations)
   * [OpenAPI importing rules](#openapi-importing-rules)
      * [Product name](#product-name)
      * [Private Base URL](#private-base-url)
      * [3scale Methods](#3scale-methods)
      * [3scale Mapping Rules](#3scale-mapping-rules)
      * [Authentication](#authentication)
      * [ActiveDocs](#activedocs)
      * [3scale Product Policy Chain](#3scale-product-policy-chain)
      * [3scale Deployment Mode](#3scale-deployment-mode)
   * [Minimum required OAS doc](#minimum-required-oas-doc)
   * [Link your OpenAPI spec to your 3scale tenant or provider account](#link-your-openapi-spec-to-your-3scale-tenant-or-provider-account)

Generated using [github-markdown-toc](https://github.com/ekalinin/github-markdown-toc)

## OpenAPI document sources

The OpenAPI document <OAS> can be read from different sources:
* Kubernetes secret
* URL. Supported schemes are `http` and `https`.

*Note*: Accepted OpenAPI spec document formats are `json` and `yaml`.

### Secret OpenAPI spec source

Create a secret with the OpenAPI spec document. The name of the secret object will be referenced in the OpenAPI CR.

The following example shows how to create a secret out of a file:

```yaml
$ cat myopenapi.yaml
---
openapi: "3.0.2"
info:
  title: "some title"
  description: "some description"
  version: "1.0.0"
paths:
  /pet:
    get:
      operationId: "getPet"
      responses:
        405:
          description: "invalid input"


$ oc create secret generic myopenapi --from-file myopenapi.yaml
secret/myopenapi created
```

**NOTE** The filename used as key inside the secret is not read by the operator. Only the content is read.

Then, create your OpenAPI CR providing reference to the secret holding the OpenAPI document.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: OpenAPI
metadata:
  name: openapi1
spec:
  openapiRef:
    secretRef:
      name: myopenapi
```

[OpenAPI CRD Reference](openapi-reference.md) for more info.

### URL OpenAPI spec source

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: OpenAPI
metadata:
  name: openapi1
spec:
  openapiRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.yaml"
```

[OpenAPI CRD Reference](openapi-reference.md) for more info.

## Supported OpenAPI spec version and limitations

* [OpenAPI __3.0.2__ specification](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md) with some limitations:
  * `info.title` field value must not exceed `253-38 = 215` character length. It will be used to create some openshift object names with some length [limitations](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/).
  * Only first `servers[0].url` element in `servers` list parsed as *private base url*. As OpenAPI specification `basePath` property, `servers[0].url` URL's base path component will be used.
  * `servers` element in path item or operation items are not supported.
  * Just a single top level security requirement supported. Operation level security requirements not supported.
  * Supported security schemes: `apiKey`.

## OpenAPI importing rules

### Product name

The default product system name is taken from the `info.title` field in the OpenAPI definition.
However, you can override this product name using the `spec.productSystemName` field
of the [OpenAPI CRD](openapi-reference.md).

### Private Base URL

Private base URL is read from OpenAPI `servers[0].url` field.
You can override this using the `spec.privateBaseURL` field
of the [OpenAPI CRD](openapi-reference.md).

### 3scale Methods

Each OpenAPI defined operation will translate in one 3scale method at product level.
The method name is read from the [operationId](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md#operationObject) field of the operation object.

### 3scale Mapping Rules

Each OpenAPI defined operation will translate in one 3scale mapping rule at product level.
Previously existing mapping rules will be replaced by those imported from the OpenAPI.

OpenAPI [paths](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md#pathsObject) object provides mapping rules *Verb* and *Pattern* properties. 3scale methods will be associated accordingly to the [operationId](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md#operationObject)

*Delta* value is hard-coded to `1`.

By default, *Strict matching* policy is being configured.
Matching policy can be switched to **Prefix matching** using the `spec.PrefixMatching` field
of the [OpenAPI CRD](openapi-reference.md).

### Authentication

Just one top level security requirement supported.
Operation level security requirements are not supported.

Supported security schemes: `apiKey`.

For the `apiKey` security scheme type:
* *credentials location* will be read from the OpenAPI [in](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md#security-scheme-object) field of the security scheme object.
* *Auth user key* will be read from the OpenAPI [name](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md#security-scheme-object) field of the security scheme object.

Partial example of OpenAPI (3.0.2) with `apiKey` security requirement

```yaml
---
openapi: "3.0.2"
security:
  - petstore_api_key: []
components:
  securitySchemes:
    petstore_api_key:
      type: apiKey
      name: api_key
      in: header
```

When OpenAPI does not specify any security requirements:
* The product authentication will be configured for `apiKey`.
* *credentials location* will default to 3scale value `As query parameters (GET) or body parameters (POST/PUT/DELETE)`.
* *Auth user key* will default to 3scale value `user_key`

3scale *Authentication Security* can be set using the `spec.privateAPIHostHeader` and the `spec.privateAPISecretToken` fields of the [OpenAPI CR](openapi-reference.md).

### ActiveDocs

No 3scale ActiveDoc is created.

### 3scale Product Policy Chain

3scale policy chain will be the default one created by 3scale.

### 3scale Deployment Mode

By default, the configured 3scale deployment mode will be `APIcast 3scale managed`.
However, when the `spec.productionPublicBaseURL` or the `spec.stagingPublicBaseURL` (or both)
fields are provided in the [OpenAPI custom resource](openapi-reference.md),
the product's deployment mode will be `APIcast self-managed`.

Example of a [OpenAPI custom resource](openapi-reference.md) with custom public base URL:

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: OpenAPI
metadata:
  name: openapi1
spec:
  openapiRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.yaml"
  productionPublicBaseURL: "https://production.my-gateway.example.com"
  stagingPublicBaseURL: "https://staging.my-gateway.example.com"
```

## Minimum required OAS doc

In [OAS 3.0.2](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md#oasDocument),
the minimum **valid** OpenAPI document just contains `info` and `paths` fields.

For instance:

```yaml
---
openapi: "3.0.2"
info:
  title: "some title"
  description: "some description"
  version: "1.0.0"
paths:
  /pet:
    get:
      operationId: "getPet"
      responses:
        405:
          description: "invalid input"
```

However, with this OpenAPI document, there is critical 3scale configuration lacking and
it must be provided for a working 3scale configuration:
* `Private Base URL` filling the `spec.privateBaseURL` field of the [OpenAPI CRD](openapi-reference.md)

The minimum valid [OpenAPI custom resource](openapi-reference.md) for a working 3scale product is:

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: OpenAPI
metadata:
  name: openapi1
spec:
  openapiRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.yaml"
```

*Note*: The referenced OpenAPI document should include the `servers[0].url` field. For instance:

```yaml
---
openapi: "3.0.2"
info:
  title: "some title"
  description: "some description"
  version: "1.0.0"
servers:
  - url: https://petstore.swagger.io/v1
paths:
  /pet:
    get:
      operationId: "getPet"
      responses:
        405:
          description: "invalid input"
```

*Note*: 3scale still requires creating the application key, but this is out of scope.

## Link your OpenAPI spec to your 3scale tenant or provider account

When some [OpenAPI custom resource](openapi-reference.md) is found by the 3scale operator,
the *LookupProviderAccount* process is started to figure out the tenant owning the resource.

The process will check the following tenant credential sources. If none is found, an error is raised.

* Read credentials from *providerAccountRef* resource attribute. This is a secret local reference, for instance `mytenant`

```
apiVersion: capabilities.3scale.net/v1beta1
kind: OpenAPI
metadata:
  name: openapi1
spec:
  openapiRef:
    url: "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/petstore.yaml"
  providerAccountRef:
    name: mytenant
```

[OpenAPI CRD Reference](openapi-reference.md) for more info about fields.

The `mytenant` secret must have`adminURL` and `token` fields with tenant credentials. For example:

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

* Default `threescale-provider-account` secret

For example: `adminURL=https://3scale-admin.example.com` and `token=123456`.

```
oc create secret generic threescale-provider-account --from-literal=adminURL=https://3scale-admin.example.com --from-literal=token=123456
```

* Default provider account in the same namespace 3scale deployment

The operator will gather required credentials automatically for the default 3scale tenant (provider account) if 3scale installation is found in the same namespace as the custom resource.
