## Adding custom policies

If you have created custom policies, you must add them to APIcast.
Go to [write your own custom policy](https://github.com/3scale/APIcast/blob/master/doc/policies.md#write-your-own-policy)
for more info about creating custom policies.

### Prerequisites

* Custom policy metadata included in `apicast-policy.json`:

```json
{
  "$schema": "http://apicast.io/policy-v1/schema#manifest#",
  "name": "APIcast Example Policy",
  "summary": "This is just an example.",
  "description": "This policy is just an example how to write your custom policy.",
  "version": "0.1",
  "configuration": {
    "type": "object",
    "properties": { }
  }
}
```
* Custom policy lua code. The `init.lua` file is required. Optionally, add more lua files. For this specific example, `init.lua` and `example.lua` files are used.

`init.lua`:
```lua
return require('example')
```

`example.lua`: Will add a request header called `X-CustomPolicy` with value `customValue` to the upstream API.
```lua
local setmetatable = setmetatable

local _M = require('apicast.policy').new('Example', '0.1')
local mt = { __index = _M }

function _M.new()
  return setmetatable({}, mt)
end

function _M:init()
  -- do work when nginx master process starts
end

function _M:init_worker()
  -- do work when nginx worker process is forked from master
end

function _M:rewrite()
  -- change the request before it reaches upstream
    ngx.req.set_header('X-CustomPolicy', 'customValue')
end

function _M:access()
  -- ability to deny the request before it is sent upstream
end

function _M:content()
  -- can create content instead of connecting to upstream
end

function _M:post_action()
  -- do something after the response was sent to the client
end

function _M:header_filter()
  -- can change response headers
end

function _M:body_filter()
  -- can read and change response body
  -- https://github.com/openresty/lua-nginx-module/blob/master/README.markdown#body_filter_by_lua
end

function _M:log()
  -- can do extra logging
end

function _M:balancer()
  -- use for example require('resty.balancer.round_robin').call to do load balancing
end

return _M
```
### Adding custom policy

#### Create secret with the custom policy content

```
oc create secret generic custom-policy-example-1 \
  --from-file=./apicast-policy.json \
  --from-file=./init.lua \
  --from-file=./example.lua
```

By default, content changes in the secret will not be noticed by the 3scale operator.
The 3scale operator allows the monitoring of secret changes, this can be achieved by adding the 
`apimanager.apps.3scale.net/watched-by=apimanager` label to the required secret.
With the label in place, when the content of the secret changes, the operator will update the deployment of the apicast 
where that secret is used (staging or production).  
The operator will not take *ownership* of the secret in any way.
```
oc label secret custom-policy-example-1 apimanager.apps.3scale.net/watched-by=apimanager
```

#### Configure and deploy APIManager CR with the custom policy

`apimanager.yaml` content (only relevant content shown):

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: example-3scale
spec:
  apicast:
    stagingSpec:
      customPolicies:
        - name: custom-policy1
          version: "0.1"
          secretRef:
             name: custom-policy-example-1
        - name: custom-policy2
          version: "0.1"
          secretRef:
              name: custom-policy-example-2
    productionSpec:
      customPolicies:
        - name: custom-policy1
          version: "0.1"
          secretRef:
             name: custom-policy-example-1
        - name: custom-policy2
          version: "0.1"
          secretRef:
              name: custom-policy-example-2
```

```
oc apply -f apimanager.yaml
```

The APIManager custom resource allows adding multiple custom policies.

**NOTE**: The tuple (`name`, `version`) has to be unique in the `customPolicies` array.

**NOTE**: If secret does not exist, the operator will mark the APIManager CustomResource as failed. 

#### Add the custom policy metadata to 3scale policy registry

The 3scale policy registry will be missing any custom policies loaded on gateways,
and therefore those policies won't be available in the Admin Portal when configuring the Integration settings for a Product.

* Using 3scale Registry API

The 3scale Registry API endpoints can be used to create, edit, list and delete custom policies definitions from/to the Admin Portal's registry.
The ActiveDocs are available from the Admin Portal at `https://{ADMIN_PORTAL}/p/admin/api_docs.`

```
$ cat policy-registry-item.json
{
  "name": "Example",
  "version": "0.1",
  "schema" : {
    "name" :"APIcast Example Policy",
    "version" : "0.1",
    "$schema" : "http://apicast.io/policy-v1/schema#manifest#",
    "summary" : "This is just an example.",
    "description" : "This policy is just an example how to write your custom policy.",
    "configuration": {
      "type" : "object",
      "properties" : {}
    }
  }
}


$ curl -v -X POST -u ":{ACCESS_TOKEN}" -H "Content-Type: application/json" \
    -d @policy-registry-item.json \
    https://{ADMIN_PORTAL}/admin/api/registry/policies.json
```

* Using 3scale operator [Application Capabilities](https://github.com/3scale/3scale-operator/blob/master/doc/operator-application-capabilities.md)

[CustomPolicyDefinition Custom Resource](https://github.com/3scale/3scale-operator/blob/master/doc/operator-application-capabilities.md#custompolicydefinition-custom-resource)
can be used. Just, deploy the `CustomPolicyDefinition` resource and the 3scale operator will do the actual registration process.

```yaml
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

#### Adding the custom policy to a policy chain in 3scale

Configure 3scale product to include the new custom policy in the gateway policy chain to be used by APIcast.
The custom policy needs to be added to the policy registry before adding it to some 3scale product's policy chain.

* Using the 3scale admin portal UI

Follow these steps to modify the policy chain in the Admin Portal:

- Log in to 3scale.
- Navigate to the API product you want to configure the policy chain for.
- In `[your_product_name]` > `Integration` > `Policies`, click Add policy.
- Under the Policy Chain section, use the arrow icons to reorder policies in the policy chain. Always place the 3scale APIcast policy last in the policy chain.
- Click the Update Policy Chain button to save the policy chain.

* Using 3scale operator [Application Capabilities](https://github.com/3scale/3scale-operator/blob/master/doc/operator-application-capabilities.md)

[Product Custom Resource](https://github.com/3scale/3scale-operator/blob/master/doc/operator-application-capabilities.md#product-policy-chain)
can be used to define desired policy chain declaratively.

```yaml
apiVersion: capabilities.3scale.net/v1beta1
kind: Product
metadata:
  name: product1
spec:
  name: "OperatedProduct 1"
  policies:
    - configuration: {}
      enabled: true
      name: Example
      version: 0.1
```
