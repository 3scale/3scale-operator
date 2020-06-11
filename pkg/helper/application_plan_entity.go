package helper

import (
	"fmt"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type ApplicationPlanEntity struct {
	productID int64
	client    *threescaleapi.ThreeScaleClient
	obj       threescaleapi.ApplicationPlanItem
	logger    logr.Logger
}

func NewApplicationPlanEntity(productID int64, obj threescaleapi.ApplicationPlanItem, cl *threescaleapi.ThreeScaleClient, logger logr.Logger) *ApplicationPlanEntity {
	return &ApplicationPlanEntity{
		productID: productID,
		obj:       obj,
		client:    cl,
		logger:    logger.WithValues("ApplicationPlanEntity", obj.ID),
	}
}

func (b *ApplicationPlanEntity) ID() int64 {
	return b.obj.ID
}

func (b *ApplicationPlanEntity) Name() string {
	return b.obj.Name
}

func (b *ApplicationPlanEntity) ApprovalRequired() bool {
	return b.obj.ApprovalRequired
}

func (b *ApplicationPlanEntity) TrialPeriodDays() int {
	return b.obj.TrialPeriodDays
}

func (b *ApplicationPlanEntity) SetupFee() float64 {
	return b.obj.SetupFee
}

func (b *ApplicationPlanEntity) CostPerMonth() float64 {
	return b.obj.CostPerMonth
}

func (b *ApplicationPlanEntity) Update(params threescaleapi.Params) error {
	b.logger.V(1).Info("Update", "params", params)
	updated, err := b.client.UpdateApplicationPlan(b.productID, b.obj.ID, params)
	if err != nil {
		return fmt.Errorf("product [%s] plan [%s] update: %w", b.productID, b.obj.SystemName, err)
	}

	b.obj = updated.Element

	return nil
}
