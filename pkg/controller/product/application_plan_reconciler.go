package product

import (
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type applicationPlanReconciler struct {
	*reconcilers.BaseReconciler
	systemName          string
	resource            capabilitiesv1beta1.ApplicationPlanSpec
	productEntity       *helper.ProductEntity
	planEntity          *helper.ApplicationPlanEntity
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func newApplicationPlanReconciler(b *reconcilers.BaseReconciler,
	systemName string,
	resource capabilitiesv1beta1.ApplicationPlanSpec,
	threescaleAPIClient *threescaleapi.ThreeScaleClient,
	productEntity *helper.ProductEntity,
	planEntity *helper.ApplicationPlanEntity,
	logger logr.Logger,
) *applicationPlanReconciler {
	return &applicationPlanReconciler{
		BaseReconciler:      b,
		systemName:          systemName,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		productEntity:       productEntity,
		planEntity:          planEntity,
		logger:              logger.WithValues("Plan", systemName),
	}
}

// Reconcile ensures plan attrs, limits and pricingRules are reconciled
func (a *applicationPlanReconciler) Reconcile() error {
	taskRunner := helper.NewTaskRunner(nil, a.logger)
	taskRunner.AddTask("SyncPlan", a.syncPlan)
	taskRunner.AddTask("SyncLimits", a.syncLimits)
	taskRunner.AddTask("SyncPricingRules", a.syncPricingRules)

	err := taskRunner.Run()
	if err != nil {
		return err
	}

	return nil
}

func (a *applicationPlanReconciler) syncPlan(_ interface{}) error {
	params := threescaleapi.Params{}

	if a.resource.Name != nil {
		if a.planEntity.Name() != *a.resource.Name {
			params["name"] = *a.resource.Name
		}
	}

	if a.resource.AppsRequireApproval != nil {
		if a.planEntity.ApprovalRequired() != *a.resource.AppsRequireApproval {
			params["approval_required"] = strconv.FormatBool(*a.resource.AppsRequireApproval)
		}
	}

	if a.resource.TrialPeriod != nil {
		if a.planEntity.TrialPeriodDays() != *a.resource.TrialPeriod {
			params["trial_period_days"] = strconv.Itoa(*a.resource.TrialPeriod)
		}
	}

	if a.resource.SetupFee != nil {
		// Field CRD openapiV3 validation should ensure no error parsing
		desiredValue, _ := strconv.ParseFloat(*a.resource.SetupFee, 10)
		if a.planEntity.SetupFee() != desiredValue {
			params["setup_fee"] = *a.resource.SetupFee
		}
	}

	if a.resource.CostMonth != nil {
		// Field CRD openapiV3 validation should ensure no error parsing
		desiredValue, _ := strconv.ParseFloat(*a.resource.CostMonth, 10)
		if a.planEntity.CostPerMonth() != desiredValue {
			params["cost_per_month"] = *a.resource.CostMonth
		}
	}

	if len(params) > 0 {
		err := a.planEntity.Update(params)
		if err != nil {
			return fmt.Errorf("Error sync plan [%s;%d]: %w", a.systemName, a.planEntity.ID(), err)
		}
	}

	return nil
}

func (a *applicationPlanReconciler) syncLimits(_ interface{}) error {
	return nil
}

func (a *applicationPlanReconciler) syncPricingRules(_ interface{}) error {
	return nil
}
