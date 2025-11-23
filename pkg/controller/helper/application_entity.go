package helper

import (
	"fmt"

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

func (b *ApplicationEntity) ChangeApplicationPlan(planID int64) error {
	b.logger.V(1).Info("ChangeApplicationPlan", "planID", planID)

	updatedApplication, err := b.client.ChangeApplicationPlan(b.UserAccountID(), b.ID(), planID)
	if err != nil {
		return fmt.Errorf("application [%s] changeApplicationPlan: %w", b.ApplicationObj.AppName, err)
	}

	b.ApplicationObj.PlanID = updatedApplication.PlanID
	return nil
}

func (b *ApplicationEntity) Suspend() error {
	b.logger.V(1).Info("Suspend")

	updatedApplication, err := b.client.ApplicationSuspend(b.UserAccountID(), b.ID())
	if err != nil {
		return fmt.Errorf("application [%s] suspend: %w", b.ApplicationObj.AppName, err)
	}
	b.ApplicationObj.State = updatedApplication.State

	return nil
}

func (b *ApplicationEntity) Resume() error {
	b.logger.V(1).Info("Resume")

	updatedApplication, err := b.client.ApplicationResume(b.UserAccountID(), b.ID())
	if err != nil {
		return fmt.Errorf("application [%s] resume: %w", b.ApplicationObj.AppName, err)
	}
	b.ApplicationObj.State = updatedApplication.State

	return nil
}

func (b *ApplicationEntity) Update(params threescaleapi.Params) error {
	b.logger.V(1).Info("Update", "params", params)

	updatedApplication, err := b.client.UpdateApplication(b.UserAccountID(), b.ID(), params)
	if err != nil {
		return fmt.Errorf("application [%s] update: %w", b.ApplicationObj.AppName, err)
	}

	b.ApplicationObj = updatedApplication

	return nil
}
