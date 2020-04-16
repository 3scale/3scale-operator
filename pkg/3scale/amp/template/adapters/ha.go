package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	HAHighlyAvailableReplicas = 2
)

type HAAdapter struct {
}

func NewHAAdapter() Adapter {
	return &HAAdapter{}
}

func (h *HAAdapter) Adapt(template *templatev1.Template) {
	options, err := h.options()
	if err != nil {
		panic(err)
	}
	haComponent := component.NewHighAvailability(options)

	h.addParameters(template)
	h.addObjects(template, haComponent)
	h.postProcess(template, haComponent)

	// update metadata
	template.Name = "3scale-api-management-ha"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system (High Availability)"
}

func (h *HAAdapter) postProcess(template *templatev1.Template, haComponent *component.HighAvailability) {
	res := helper.UnwrapRawExtensions(template.Objects)
	h.increaseReplicasNumber(res)
	res = h.deleteInternalDatabasesObjects(res)
	h.updateDatabasesURLS(haComponent, res)
	h.deleteDBRelatedParameters(template)
	h.unsetSystemRedisDBDefaultValues(template)
	template.Objects = helper.WrapRawExtensions(res)
}

func (h *HAAdapter) addObjects(template *templatev1.Template, haComponent *component.HighAvailability) {
	componentObjects := h.componentObjects(haComponent)
	template.Objects = append(template.Objects, helper.WrapRawExtensions(componentObjects)...)
}

func (h *HAAdapter) componentObjects(c *component.HighAvailability) []common.KubernetesObject {
	systemDatabaseSecret := c.SystemDatabaseSecret()

	objects := []common.KubernetesObject{
		systemDatabaseSecret,
	}

	return objects
}

func (h *HAAdapter) increaseReplicasNumber(objects []common.KubernetesObject) {
	// We do not increase the number of replicas in database DeploymentConfigs
	excludedDeploymentConfigs := map[string]bool{
		"system-memcache": true,
		"system-sphinx":   true,
		"zync-database":   true,
	}

	for _, obj := range objects {
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			if _, isExcluded := excludedDeploymentConfigs[dc.Name]; !isExcluded {
				dc.Spec.Replicas = HAHighlyAvailableReplicas
			}
		}
	}
}

func (h *HAAdapter) deleteInternalDatabasesObjects(objects []common.KubernetesObject) []common.KubernetesObject {
	keepObjects := []common.KubernetesObject{}

	for objIdx, object := range objects {
		switch obj := (object).(type) {
		case *appsv1.DeploymentConfig:
			if _, ok := component.HighlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				// We create a new array and add to it the elements that will
				//NOT have to be deleted
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *v1.Service:
			if _, ok := component.HighlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *v1.PersistentVolumeClaim:
			if obj.ObjectMeta.Name != "backend-redis-storage" && obj.ObjectMeta.Name != "system-redis-storage" &&
				obj.ObjectMeta.Name != "mysql-storage" {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *v1.ConfigMap:
			if obj.ObjectMeta.Name != "mysql-main-conf" && obj.ObjectMeta.Name != "mysql-extra-conf" &&
				obj.ObjectMeta.Name != "redis-config" {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *imagev1.ImageStream:
			if _, ok := component.HighlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		default:
			keepObjects = append(keepObjects, objects[objIdx])
		}
	}

	return keepObjects
}

func (h *HAAdapter) updateDatabasesURLS(c *component.HighAvailability, objects []common.KubernetesObject) {
	for objIdx := range objects {
		obj := objects[objIdx]
		secret, ok := obj.(*v1.Secret)
		if ok {
			switch secret.Name {
			case "system-redis":
				secret.StringData["URL"] = c.Options.SystemRedisURL
				secret.StringData["MESSAGE_BUS_URL"] = c.Options.SystemMessageBusRedisURL
				secret.StringData[component.SystemSecretSystemRedisSentinelHosts] = c.Options.SystemRedisSentinelsHosts
				secret.StringData[component.SystemSecretSystemRedisSentinelRole] = c.Options.SystemRedisSentinelsRole
				secret.StringData[component.SystemSecretSystemRedisMessageBusSentinelHosts] = c.Options.SystemMessageBusRedisSentinelsHosts
				secret.StringData[component.SystemSecretSystemRedisMessageBusSentinelRole] = c.Options.SystemMessageBusRedisSentinelsRole
			case "backend-redis":
				secret.StringData["REDIS_STORAGE_URL"] = c.Options.BackendRedisStorageEndpoint
				secret.StringData["REDIS_QUEUES_URL"] = c.Options.BackendRedisQueuesEndpoint
				secret.StringData[component.BackendSecretBackendRedisStorageSentinelHostsFieldName] = c.Options.BackendRedisStorageSentinelHosts
				secret.StringData[component.BackendSecretBackendRedisStorageSentinelRoleFieldName] = c.Options.BackendRedisStorageSentinelRole
				secret.StringData[component.BackendSecretBackendRedisQueuesSentinelHostsFieldName] = c.Options.BackendRedisQueuesSentinelHosts
				secret.StringData[component.BackendSecretBackendRedisQueuesSentinelRoleFieldName] = c.Options.BackendRedisQueuesSentinelRole
			}
		}
	}
}

func (h *HAAdapter) deleteDBRelatedParameters(template *templatev1.Template) {
	keepParams := []templatev1.Parameter{}
	dbParamsToDelete := map[string]bool{
		"REDIS_IMAGE":                   true,
		"SYSTEM_DATABASE_IMAGE":         true,
		"SYSTEM_DATABASE_USER":          true,
		"SYSTEM_DATABASE_PASSWORD":      true,
		"SYSTEM_DATABASE":               true,
		"SYSTEM_DATABASE_ROOT_PASSWORD": true,
	}

	for paramIdx := range template.Parameters {
		paramName := template.Parameters[paramIdx].Name
		if _, ok := dbParamsToDelete[paramName]; !ok {
			keepParams = append(keepParams, template.Parameters[paramIdx])
		}
	}

	template.Parameters = keepParams
}

func (h *HAAdapter) unsetSystemRedisDBDefaultValues(template *templatev1.Template) {
	dbParamsToUpdate := map[string]bool{
		"SYSTEM_REDIS_URL":             true,
		"SYSTEM_MESSAGE_BUS_REDIS_URL": true,
	}

	for paramIdx := range template.Parameters {
		paramName := template.Parameters[paramIdx].Name
		if _, ok := dbParamsToUpdate[paramName]; ok {
			template.Parameters[paramIdx].Value = ""
		}
	}
}

func (h *HAAdapter) addParameters(template *templatev1.Template) {
	template.Parameters = append(template.Parameters, h.parameters()...)
}

func (h *HAAdapter) options() (*component.HighAvailabilityOptions, error) {
	o := component.NewHighAvailabilityOptions()
	o.AppLabel = "${APP_LABEL}"
	o.BackendRedisQueuesEndpoint = "${BACKEND_REDIS_QUEUES_ENDPOINT}"
	o.BackendRedisQueuesSentinelHosts = "${BACKEND_REDIS_QUEUE_SENTINEL_HOSTS}"
	o.BackendRedisQueuesSentinelRole = "${BACKEND_REDIS_QUEUE_SENTINEL_ROLE}"
	o.BackendRedisStorageEndpoint = "${BACKEND_REDIS_STORAGE_ENDPOINT}"
	o.BackendRedisStorageSentinelHosts = "${BACKEND_REDIS_STORAGE_SENTINEL_HOSTS}"
	o.BackendRedisStorageSentinelRole = "${BACKEND_REDIS_STORAGE_SENTINEL_ROLE}"
	o.SystemDatabaseURL = "${SYSTEM_DATABASE_URL}"
	o.SystemRedisURL = "${SYSTEM_REDIS_URL}"
	o.SystemRedisSentinelsHosts = "${SYSTEM_REDIS_SENTINEL_HOSTS}"
	o.SystemRedisSentinelsRole = "${SYSTEM_REDIS_SENTINEL_ROLE}"
	o.SystemMessageBusRedisSentinelsHosts = "${SYSTEM_MESSAGE_BUS_REDIS_SENTINEL_HOSTS}"
	o.SystemMessageBusRedisSentinelsRole = "${SYSTEM_MESSAGE_BUS_REDIS_SENTINEL_ROLE}"
	o.SystemMessageBusRedisURL = "${SYSTEM_MESSAGE_BUS_REDIS_URL}"

	err := o.Validate()
	return o, err
}

func (h *HAAdapter) parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_STORAGE_ENDPOINT",
			Description: "Define the external backend-redis storage endpoint to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_QUEUES_ENDPOINT",
			Description: "Define the external backend-redis queues endpoint to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_URL",
			Description: "Define the external system-mysql to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_MESSAGE_BUS_REDIS_SENTINEL_HOSTS",
			Description: "Define the external system message bus sentinel hosts",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_MESSAGE_BUS_REDIS_SENTINEL_ROLE",
			Description: "Define the external system message bus sentinel role",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_SENTINEL_HOSTS",
			Description: "Define the external system redis sentinel hosts",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_SENTINEL_ROLE",
			Description: "Define the external system redis sentinel role",
		},
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_QUEUE_SENTINEL_HOSTS",
			Description: "Define the external backend redis queue sentinel hosts",
		},
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_QUEUE_SENTINEL_ROLE",
			Description: "Define the external backend redis queue sentinel role",
		},
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_STORAGE_SENTINEL_HOSTS",
			Description: "Define the external backend redis storage sentinel hosts",
		},
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_STORAGE_SENTINEL_ROLE",
			Description: "Define the external backend redis storage sentinel role",
		},
	}
}
