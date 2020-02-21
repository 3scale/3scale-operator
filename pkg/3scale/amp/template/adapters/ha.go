package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	templatev1 "github.com/openshift/api/template/v1"
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
	haComponent.IncreaseReplicasNumber(res)
	res = haComponent.DeleteInternalDatabasesObjects(res)
	haComponent.UpdateDatabasesURLS(res)
	haComponent.DeleteDBRelatedParameters(template)
	haComponent.UnsetSystemRedisDBDefaultValues(template)
	template.Objects = helper.WrapRawExtensions(res)
}

func (h *HAAdapter) addObjects(template *templatev1.Template, haComponent *component.HighAvailability) {
	template.Objects = append(template.Objects, helper.WrapRawExtensions(haComponent.Objects())...)
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
