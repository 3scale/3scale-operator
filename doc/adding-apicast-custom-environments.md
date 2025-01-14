## Adding custom environments

Add custom environment(s) loaded in all 3scale products.

Here is an example of a environment that is loaded in all services: `custom_env.lua`

```lua
local cjson = require('cjson')
local PolicyChain = require('apicast.policy_chain')
local policy_chain = context.policy_chain

local logging_policy_config = cjson.decode([[
{
  "enable_access_logs": false,
  "custom_logging": "\"{{request}}\" to service {{service.id}} and {{service.name}}"
}
]])

policy_chain:insert( PolicyChain.load_policy('logging', 'builtin', logging_policy_config), 1)

return {
  policy_chain = policy_chain,
  port = { metrics = 9421 },
}
```

### Prerequisites

* One or more custom environment(s) in lua code.

### Adding custom environment

#### Create secret with the custom environment content

```
oc create secret generic custom-env-1 --from-file=./custom_env.lua
```

By default, content changes in the secret will not be noticed by the 3scale operator.
The 3scale operator allows the monitoring of secret changes, this can be achieved by adding the
`apimanager.apps.3scale.net/watched-by=apimanager` label to the required secret.
With the label in place, when the content of the secret changes, the operator will update the deployment of the apicast
where that secret is used (staging or production).  
The operator will not take *ownership* of the secret in any way.
```
oc label secret custom-env-1 apimanager.apps.3scale.net/watched-by=apimanager
```

**NOTE**: a secret can host multiple custom environments. The operator will load each one of them.

#### Configure and deploy APIManager CR with the apicast custom environment

`apimanager.yaml` content (only relevant content shown):

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIManager
metadata:
  name: apimanager-apicast-custom-environment
spec:
  apicast:
    productionSpec:
      customEnvironments:
        - secretRef:
            name: custom-env-1
        - secretRef:
            name: custom-env-2
    stagingSpec:
      customEnvironments:
        - secretRef:
            name: custom-env-3
```

**NOTE**: Multiple custom environment secrets can be added. The operator will load each one of them.

Check [APIManager CRD Reference](apimanager-reference.md) documentation for more information.

```
oc apply -f apimanager.yaml
```

The APIManager custom resource allows adding multiple custom environments per secret.

**NOTE**: If the referenced secret does not exist, the operator will mark the APIManager CustomResource as failed. The apicast Deployment object will also fail if the referenced secret does not exist.