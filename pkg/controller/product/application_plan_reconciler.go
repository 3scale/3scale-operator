package product

import (
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type limitKey struct {
	Period   string
	Value    int
	MetricID int64
}

type pricingRuleKey struct {
	From         int
	To           int
	PricePerUnit string
	MetricID     int64
}

type applicationPlanReconciler struct {
	*reconcilers.BaseReconciler
	systemName          string
	resource            capabilitiesv1beta1.ApplicationPlanSpec
	productEntity       *controllerhelper.ProductEntity
	backendRemoteIndex  *controllerhelper.BackendAPIRemoteIndex
	planEntity          *controllerhelper.ApplicationPlanEntity
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func newApplicationPlanReconciler(b *reconcilers.BaseReconciler,
	systemName string,
	resource capabilitiesv1beta1.ApplicationPlanSpec,
	threescaleAPIClient *threescaleapi.ThreeScaleClient,
	productEntity *controllerhelper.ProductEntity,
	backendRemoteIndex *controllerhelper.BackendAPIRemoteIndex,
	planEntity *controllerhelper.ApplicationPlanEntity,
	logger logr.Logger,
) *applicationPlanReconciler {

	return &applicationPlanReconciler{
		BaseReconciler:      b,
		systemName:          systemName,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		productEntity:       productEntity,
		backendRemoteIndex:  backendRemoteIndex,
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
	// desired Limits
	desiredList := a.resource.Limits

	// existing Limits
	existingList, err := a.planEntity.Limits()
	if err != nil {
		return fmt.Errorf("Error sync plan [%s] limits: %w", a.systemName, err)
	}

	// item is not updated, either created or deleted.
	// computeUnDesiredLimits should match the entire object
	undesiredLimits, err := a.computeUnDesiredLimits(existingList.Limits, desiredList)
	if err != nil {
		return fmt.Errorf("Error sync plan [%s] limits: %w", a.systemName, err)
	}
	for idx := range undesiredLimits {
		err := a.planEntity.DeleteLimit(undesiredLimits[idx].Element.MetricID, undesiredLimits[idx].Element.ID)
		if err != nil {
			return err
		}
	}

	// item is not updated, either created or deleted.
	// computeDesiredLimits should match the entire object
	desiredLimits, err := a.computeDesiredLimits(desiredList, existingList.Limits)
	if err != nil {
		return fmt.Errorf("Error sync plan [%s] limits: %w", a.systemName, err)
	}

	for idx := range desiredLimits {
		params := threescaleapi.Params{
			"period": desiredLimits[idx].Period,
			"value":  strconv.Itoa(desiredLimits[idx].Value),
		}

		metricID, err := a.findID(desiredLimits[idx].MetricMethodRef)
		if err != nil {
			return err
		}

		err = a.planEntity.CreateLimit(metricID, params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *applicationPlanReconciler) syncPricingRules(_ interface{}) error {
	// desired pricing rules
	desiredList := a.resource.PricingRules

	// existing pricing rules
	existingList, err := a.planEntity.PricingRules()
	if err != nil {
		return fmt.Errorf("Error sync plan [%s] pricing rules: %w", a.systemName, err)
	}

	// item is not updated, either created or deleted.
	// computeUnDesiredPricingRules should match the entire object
	undesiredRules, err := a.computeUnDesiredPricingRules(existingList.Rules, desiredList)
	if err != nil {
		return fmt.Errorf("Error sync plan [%s] pricing rules: %w", a.systemName, err)
	}
	for idx := range undesiredRules {
		err := a.planEntity.DeletePricingRule(undesiredRules[idx].Element.MetricID, undesiredRules[idx].Element.ID)
		if err != nil {
			return err
		}
	}

	// item is not updated, either created or deleted.
	// computeDesiredPricingRules should match the entire object
	desiredRules, err := a.computeDesiredPricingRules(desiredList, existingList.Rules)
	if err != nil {
		return fmt.Errorf("Error sync plan [%s] pricing rules: %w", a.systemName, err)
	}

	for idx := range desiredRules {
		params := threescaleapi.Params{
			"min":           strconv.Itoa(desiredRules[idx].From),
			"max":           strconv.Itoa(desiredRules[idx].To),
			"cost_per_unit": desiredRules[idx].PricePerUnit,
		}

		metricID, err := a.findID(desiredRules[idx].MetricMethodRef)
		if err != nil {
			return err
		}

		err = a.planEntity.CreatePricingRule(metricID, params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *applicationPlanReconciler) computeUnDesiredLimits(
	existingList []threescaleapi.ApplicationPlanLimit,
	desiredList []capabilitiesv1beta1.LimitSpec) ([]threescaleapi.ApplicationPlanLimit, error) {

	target := map[limitKey]bool{}
	for _, desired := range desiredList {
		metricID, err := a.findID(desired.MetricMethodRef)
		if err != nil {
			return nil, err
		}

		desiredKey := limitKey{
			Period:   desired.Period,
			Value:    desired.Value,
			MetricID: metricID,
		}

		target[desiredKey] = true
	}

	result := make([]threescaleapi.ApplicationPlanLimit, 0)
	for _, existing := range existingList {
		existingKey := limitKey{
			Period:   existing.Element.Period,
			Value:    existing.Element.Value,
			MetricID: existing.Element.MetricID,
		}

		if _, ok := target[existingKey]; !ok {
			result = append(result, existing)
		}
	}

	return result, nil
}

func (a *applicationPlanReconciler) computeDesiredLimits(
	desiredList []capabilitiesv1beta1.LimitSpec,
	existingList []threescaleapi.ApplicationPlanLimit) ([]capabilitiesv1beta1.LimitSpec, error) {

	target := map[limitKey]bool{}
	for _, existing := range existingList {
		existingKey := limitKey{
			Period:   existing.Element.Period,
			Value:    existing.Element.Value,
			MetricID: existing.Element.MetricID,
		}

		target[existingKey] = true
	}

	result := make([]capabilitiesv1beta1.LimitSpec, 0)
	for _, desired := range desiredList {
		metricID, err := a.findID(desired.MetricMethodRef)
		if err != nil {
			return nil, err
		}
		desiredKey := limitKey{
			Period:   desired.Period,
			Value:    desired.Value,
			MetricID: metricID,
		}

		if _, ok := target[desiredKey]; !ok {
			result = append(result, desired)
		}
	}

	return result, nil
}

func (a *applicationPlanReconciler) findID(ref capabilitiesv1beta1.MetricMethodRefSpec) (int64, error) {
	var (
		metricID int64
		err      error
	)
	if ref.BackendSystemName == nil {
		metricID, err = a.productEntity.FindMethodMetricIDBySystemName(ref.SystemName)
		if err != nil {
			return metricID, err
		}
	} else {
		backendEntity, ok := a.backendRemoteIndex.FindBySystemName(*ref.BackendSystemName)
		if !ok {
			panic(fmt.Sprintf("Backend SystemName %s not found in backend index", *ref.BackendSystemName))
		}
		metricID, err = backendEntity.FindMethodMetricIDBySystemName(ref.SystemName)
		if err != nil {
			return metricID, err
		}
	}
	return metricID, nil
}

func (a *applicationPlanReconciler) computeUnDesiredPricingRules(
	existingList []threescaleapi.ApplicationPlanPricingRule,
	desiredList []capabilitiesv1beta1.PricingRuleSpec) ([]threescaleapi.ApplicationPlanPricingRule, error) {

	target := map[pricingRuleKey]bool{}
	for _, desired := range desiredList {
		metricID, err := a.findID(desired.MetricMethodRef)
		if err != nil {
			return nil, err
		}

		desiredKey := pricingRuleKey{
			From:         desired.From,
			To:           desired.To,
			PricePerUnit: desired.PricePerUnit,
			MetricID:     metricID,
		}

		target[desiredKey] = true
	}

	result := make([]threescaleapi.ApplicationPlanPricingRule, 0)
	for _, existing := range existingList {
		existingKey := pricingRuleKey{
			From:         existing.Element.Min,
			To:           existing.Element.Max,
			PricePerUnit: existing.Element.CostPerUnit,
			MetricID:     existing.Element.MetricID,
		}

		if _, ok := target[existingKey]; !ok {
			result = append(result, existing)
		}
	}

	return result, nil
}

func (a *applicationPlanReconciler) computeDesiredPricingRules(
	desiredList []capabilitiesv1beta1.PricingRuleSpec,
	existingList []threescaleapi.ApplicationPlanPricingRule) ([]capabilitiesv1beta1.PricingRuleSpec, error) {

	target := map[pricingRuleKey]bool{}
	for _, existing := range existingList {
		existingKey := pricingRuleKey{
			From:         existing.Element.Min,
			To:           existing.Element.Max,
			PricePerUnit: existing.Element.CostPerUnit,
			MetricID:     existing.Element.MetricID,
		}

		target[existingKey] = true
	}

	result := make([]capabilitiesv1beta1.PricingRuleSpec, 0)
	for _, desired := range desiredList {
		metricID, err := a.findID(desired.MetricMethodRef)
		if err != nil {
			return nil, err
		}
		desiredKey := pricingRuleKey{
			From:         desired.From,
			To:           desired.To,
			PricePerUnit: desired.PricePerUnit,
			MetricID:     metricID,
		}

		if _, ok := target[desiredKey]; !ok {
			result = append(result, desired)
		}
	}

	return result, nil
}
