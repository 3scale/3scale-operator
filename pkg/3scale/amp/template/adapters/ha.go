package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	templatev1 "github.com/openshift/api/template/v1"
)

type HAAdapter struct {
}

func NewHAAdapter(options []string) Adapter {
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
	hob := component.HighAvailabilityOptionsBuilder{}
	hob.AppLabel("${APP_LABEL}")
	hob.BackendRedisQueuesEndpoint("${BACKEND_REDIS_QUEUES_ENDPOINT}")
	hob.BackendRedisStorageEndpoint("${BACKEND_REDIS_STORAGE_ENDPOINT}")
	hob.SystemDatabaseURL("${SYSTEM_DATABASE_URL}")
	hob.SystemRedisURL("${SYSTEM_REDIS_URL}")
	hob.SystemMessageBusRedisURL("${SYSTEM_MESSAGE_BUS_REDIS_URL}")

	return hob.Build()
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
	}
}
