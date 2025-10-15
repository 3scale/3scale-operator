package helper

import (
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type ApplicationEntity struct {
	client          *threescaleapi.ThreeScaleClient
	ApplicationObj  *threescaleapi.Application
	ApplicationList *threescaleapi.ApplicationList
	*threescaleapi.ApplicationPlanJSONList
	logger logr.Logger
}

func NewApplicationEntity(ApplicationObj *threescaleapi.Application, client *threescaleapi.ThreeScaleClient, logger logr.Logger) *ApplicationEntity {
	return &ApplicationEntity{
		ApplicationObj: ApplicationObj,
		client:         client,
		logger:         logger.WithValues("Application", ApplicationObj.ID),
	}
}

func (b *ApplicationEntity) ID() int64 {
	return b.ApplicationObj.ID
}

func (b *ApplicationEntity) AppName() string {
	return b.ApplicationObj.AppName
}

func (b *ApplicationEntity) Description() string {
	return b.ApplicationObj.Description
}

func (b *ApplicationEntity) UserAccountID() int64 {
	return b.ApplicationObj.UserAccountID
}

func (b *ApplicationEntity) PlanID() int64 {
	return b.ApplicationObj.PlanID
}

func (b *ApplicationEntity) ProductId() int64 {
	return b.ApplicationObj.ServiceID
}

func (b *ApplicationEntity) ApplicationState() string {
	return b.ApplicationObj.State
}
