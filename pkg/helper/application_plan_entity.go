package helper

import (
	"fmt"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type ApplicationPlanEntity struct {
	productID    int64
	client       *threescaleapi.ThreeScaleClient
	obj          threescaleapi.ApplicationPlanItem
	limits       *threescaleapi.ApplicationPlanLimitList
	pricingRules *threescaleapi.ApplicationPlanPricingRuleList
	logger       logr.Logger
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
		return fmt.Errorf("product [%d] plan [%s] update: %w", b.productID, b.obj.SystemName, err)
	}

	b.obj = updated.Element

	return nil
}

func (b *ApplicationPlanEntity) Limits() (*threescaleapi.ApplicationPlanLimitList, error) {
	if b.limits == nil {
		limits, err := b.getLimits()
		if err != nil {
			return nil, err
		}
		b.limits = limits
	}
	return b.limits, nil
}

func (b *ApplicationPlanEntity) getLimits() (*threescaleapi.ApplicationPlanLimitList, error) {
	b.logger.V(1).Info("getLimits")
	list, err := b.client.ListApplicationPlansLimits(b.obj.ID)
	if err != nil {
		return nil, fmt.Errorf("application plan [%s] get limits: %w", b.obj.SystemName, err)
	}

	return list, nil
}

func (b *ApplicationPlanEntity) DeleteLimit(metricID, id int64) error {
	b.logger.V(1).Info("DeleteLimit", "metricID", metricID, "ID", id)
	err := b.client.DeleteApplicationPlanLimit(b.obj.ID, metricID, id)
	if err != nil {
		return fmt.Errorf("application plan [%s] delete limit: %w", b.obj.SystemName, err)
	}
	b.resetLimits()
	return nil
}

func (b *ApplicationPlanEntity) CreateLimit(metricID int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateLimit", "metricID", metricID, "params", params)
	_, err := b.client.CreateApplicationPlanLimit(b.obj.ID, metricID, params)
	if err != nil {
		return fmt.Errorf("application plan [%s] create limit: %w", b.obj.SystemName, err)
	}
	b.resetLimits()
	return nil
}

func (b *ApplicationPlanEntity) resetLimits() {
	b.limits = nil
}

func (b *ApplicationPlanEntity) PricingRules() (*threescaleapi.ApplicationPlanPricingRuleList, error) {
	if b.pricingRules == nil {
		rules, err := b.getPricingRules()
		if err != nil {
			return nil, err
		}
		b.pricingRules = rules
	}
	return b.pricingRules, nil
}

func (b *ApplicationPlanEntity) getPricingRules() (*threescaleapi.ApplicationPlanPricingRuleList, error) {
	b.logger.V(1).Info("getPricingRules")
	list, err := b.client.ListApplicationPlansPricingRules(b.obj.ID)
	if err != nil {
		return nil, fmt.Errorf("application plan [%s] get pricing rules: %w", b.obj.SystemName, err)
	}

	return list, nil
}

func (b *ApplicationPlanEntity) DeletePricingRule(metricID, id int64) error {
	b.logger.V(1).Info("DeletePricingRule", "metricID", metricID, "ID", id)
	err := b.client.DeleteApplicationPlanPricingRule(b.obj.ID, metricID, id)
	if err != nil {
		return fmt.Errorf("application plan [%s] delete pricing rule: %w", b.obj.SystemName, err)
	}
	b.resetPricingRules()
	return nil
}

func (b *ApplicationPlanEntity) CreatePricingRule(metricID int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("CreatePricingRule", "metricID", metricID, "params", params)
	_, err := b.client.CreateApplicationPlanPricingRule(b.obj.ID, metricID, params)
	if err != nil {
		return fmt.Errorf("application plan [%s] create pricing rule: %w", b.obj.SystemName, err)
	}
	b.resetPricingRules()
	return nil
}

func (b *ApplicationPlanEntity) resetPricingRules() {
	b.pricingRules = nil
}
