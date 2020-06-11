package helper

import (
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type ApplicationPlanEntity struct {
	client *threescaleapi.ThreeScaleClient
	obj    threescaleapi.ApplicationPlanItem
	logger logr.Logger
}

func NewApplicationPlanEntity(obj threescaleapi.ApplicationPlanItem, cl *threescaleapi.ThreeScaleClient, logger logr.Logger) *ApplicationPlanEntity {
	return &ApplicationPlanEntity{
		obj:    obj,
		client: cl,
		logger: logger.WithValues("ApplicationPlanEntity", obj.ID),
	}
}

func (b *ApplicationPlanEntity) ID() int64 {
	return b.obj.ID
}

func (b *ApplicationPlanEntity) Name() string {
	return b.obj.Name
}
