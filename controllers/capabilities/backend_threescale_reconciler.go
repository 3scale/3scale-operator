package controllers

import (
	"errors"
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type BackendThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	backendResource     *capabilitiesv1beta1.Backend
	backendAPIEntity    *controllerhelper.BackendAPIEntity
	backendRemoteIndex  *controllerhelper.BackendAPIRemoteIndex
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	providerAccount     *controllerhelper.ProviderAccount
	logger              logr.Logger
}

func NewThreescaleReconciler(b *reconcilers.BaseReconciler,
	backendResource *capabilitiesv1beta1.Backend,
	threescaleAPIClient *threescaleapi.ThreeScaleClient,
	backendRemoteIndex *controllerhelper.BackendAPIRemoteIndex,
	providerAccount *controllerhelper.ProviderAccount,
) *BackendThreescaleReconciler {

	return &BackendThreescaleReconciler{
		BaseReconciler:      b,
		backendResource:     backendResource,
		backendRemoteIndex:  backendRemoteIndex,
		threescaleAPIClient: threescaleAPIClient,
		providerAccount:     providerAccount,
		logger:              b.Logger().WithValues("3scale Reconciler", backendResource.Name),
	}
}

func (t *BackendThreescaleReconciler) Reconcile() (*controllerhelper.BackendAPIEntity, error) {
	taskRunner := helper.NewTaskRunner(nil, t.logger)
	taskRunner.AddTask("SyncBackend", t.syncBackend)
	// First methods and metrics, then mapping rules.
	// Mapping rules reference methods and metrics.
	// When a method/metric is deleted,
	// any orphan mapping rule will be deleted automatically by 3scale
	taskRunner.AddTask("SyncMethods", t.syncMethods)
	taskRunner.AddTask("SyncMetrics", t.syncMetrics)
	taskRunner.AddTask("SyncMappingRules", t.syncMappingRules)

	err := taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.backendAPIEntity, nil
}

func (t *BackendThreescaleReconciler) syncBackend(_ interface{}) error {
	var (
		err              error
		backendAPIEntity *controllerhelper.BackendAPIEntity
	)

	backendAPIEntity, exists := t.backendRemoteIndex.FindBySystemName(t.backendResource.Spec.SystemName)

	if !exists {
		// Create backend using system_name.
		// it cannot be modified later
		params := threescaleapi.Params{
			"system_name":      t.backendResource.Spec.SystemName,
			"name":             t.backendResource.Spec.Name,
			"private_endpoint": t.backendResource.Spec.PrivateBaseURL,
		}
		backendAPIEntity, err = t.backendRemoteIndex.CreateBackendAPI(params)
		if err != nil {
			return fmt.Errorf("Error sync backend [%s]: %w", t.backendResource.Spec.SystemName, err)
		}
	}

	// Will be used by coming steps
	t.backendAPIEntity = backendAPIEntity

	updatedParams := threescaleapi.Params{}

	if t.backendAPIEntity.Name() != t.backendResource.Spec.Name {
		updatedParams["name"] = t.backendResource.Spec.Name
	}

	if t.backendAPIEntity.Description() != t.backendResource.Spec.Description {
		updatedParams["description"] = t.backendResource.Spec.Description
	}

	if t.backendAPIEntity.PrivateEndpoint() != t.backendResource.Spec.PrivateBaseURL {
		updatedParams["private_endpoint"] = t.backendResource.Spec.PrivateBaseURL
	}

	if len(updatedParams) > 0 {
		err = t.backendAPIEntity.Update(updatedParams)
		if err != nil {
			return fmt.Errorf("Error sync backend [%s]: %w", t.backendResource.Spec.SystemName, err)
		}
	}

	return nil
}

func (t *BackendThreescaleReconciler) syncMethods(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.backendResource.Spec.Methods))
	for systemName := range t.backendResource.Spec.Methods {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MethodItem{}
	existingList, err := t.backendAPIEntity.Methods()
	if err != nil {
		return fmt.Errorf("Error sync backend methods [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Methods))
	for _, existing := range existingList.Methods {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	//
	// Deleted existing and not desired
	//
	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	notDesiredMap := map[string]threescaleapi.MethodItem{}
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredMap[systemName] = existingMap[systemName]
	}
	err = t.deleteNotDesiredMethodsFrom3scale(notDesiredMap)
	if err != nil {
		return fmt.Errorf("Error sync backend methods [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	err = t.deleteExternalMetricReferences(notDesiredExistingKeys)
	if err != nil {
		return fmt.Errorf("Error sync backend methods [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	//
	// Reconcile existing and changed
	//
	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]methodData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = methodData{
			item: existingMap[systemName],
			spec: t.backendResource.Spec.Methods[systemName],
		}
	}

	err = t.reconcileMatchedMethods(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync backend methods [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired
	//
	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.MethodSpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Method map key set
		desiredNewMap[systemName] = t.backendResource.Spec.Methods[systemName]
	}
	err = t.createNewMethods(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error sync backend methods [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	return nil
}

func (t *BackendThreescaleReconciler) createNewMethods(desiredNewMap map[string]capabilitiesv1beta1.MethodSpec) error {
	for systemName, method := range desiredNewMap {
		params := threescaleapi.Params{
			"friendly_name": method.Name,
			"system_name":   systemName,
		}
		if len(method.Description) > 0 {
			params["description"] = method.Description
		}
		err := t.backendAPIEntity.CreateMethod(params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *BackendThreescaleReconciler) deleteNotDesiredMethodsFrom3scale(notDesiredMap map[string]threescaleapi.MethodItem) error {
	for _, notDesiredMethod := range notDesiredMap {
		err := t.backendAPIEntity.DeleteMethod(notDesiredMethod.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// valid for metrics and methods as long as 3scale ensures system_names are unique among methods and metrics
func (t *BackendThreescaleReconciler) deleteExternalMetricReferences(notDesiredMetrics []string) error {
	productList, err := controllerhelper.ProductList(t.backendResource.Namespace, t.Client(), t.providerAccount.AdminURLStr, t.logger)
	if err != nil {
		return fmt.Errorf("deleteExternalMetricReferences: %w", err)
	}

	// filter products referencing current backend resource
	linkedProductList := make([]capabilitiesv1beta1.Product, 0)
	for _, product := range productList {
		if _, ok := product.Spec.BackendUsages[t.backendResource.Spec.SystemName]; ok {
			linkedProductList = append(linkedProductList, product)
		}
	}

	t.logger.V(1).Info("Product linked to backend", "total", len(linkedProductList))

	for productIdx := range linkedProductList {
		err = t.deleteExternalMetricReferencesOnProduct(notDesiredMetrics, linkedProductList[productIdx])
		if err != nil {
			return fmt.Errorf("deleteExternalMetricReferences: %w", err)
		}
	}

	return nil
}

// valid for metrics and methods as long as 3scale ensures system_names are unique among methods and metrics
func (t *BackendThreescaleReconciler) deleteExternalMetricReferencesOnProduct(notDesiredMetrics []string, productRef capabilitiesv1beta1.Product) error {
	productUpdated := false
	product := productRef.DeepCopy()

	for planSystemName, planSpec := range product.Spec.ApplicationPlans {
		planSpecUpdated := false

		// Check limits with external references to the current backend
		newLimits := make([]capabilitiesv1beta1.LimitSpec, 0)
		for limitIdx, limitSpec := range planSpec.Limits {
			// Check if the limit belongs to the current backend
			// Check if the limit is marked for deletion in notDesiredMap
			if limitSpec.MetricMethodRef.BackendSystemName == nil ||
				*limitSpec.MetricMethodRef.BackendSystemName != t.backendResource.Spec.SystemName ||
				!helper.ArrayContains(notDesiredMetrics, limitSpec.MetricMethodRef.SystemName) {
				newLimits = append(newLimits, planSpec.Limits[limitIdx])
			}
		}

		if len(newLimits) != len(planSpec.Limits) {
			planSpecUpdated = true
			planSpec.Limits = newLimits
		}

		// Check pricingRules with external references to the current backend
		newRules := make([]capabilitiesv1beta1.PricingRuleSpec, 0)
		for ruleIdx, ruleSpec := range planSpec.PricingRules {
			// Check if the current rule belongs to the current backend
			if ruleSpec.MetricMethodRef.BackendSystemName == nil ||
				*ruleSpec.MetricMethodRef.BackendSystemName != t.backendResource.Spec.SystemName ||
				!helper.ArrayContains(notDesiredMetrics, ruleSpec.MetricMethodRef.SystemName) {
				newRules = append(newRules, planSpec.PricingRules[ruleIdx])
			}
		}

		if len(newRules) != len(planSpec.PricingRules) {
			planSpecUpdated = true
			planSpec.PricingRules = newRules
		}

		if planSpecUpdated {
			productUpdated = true
			product.Spec.ApplicationPlans[planSystemName] = planSpec
		}
	}

	if productUpdated {
		err := t.UpdateResource(product)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *BackendThreescaleReconciler) reconcileMatchedMethods(matchedMap map[string]methodData) error {
	for _, data := range matchedMap {
		params := threescaleapi.Params{}
		if data.spec.Name != data.item.Name {
			params["friendly_name"] = data.spec.Name
		}

		if data.spec.Description != data.item.Description {
			params["description"] = data.spec.Description
		}

		if len(params) > 0 {
			err := t.backendAPIEntity.UpdateMethod(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error reconcile backendAPI methods: %w", err)
			}
		}
	}

	return nil
}

func (t *BackendThreescaleReconciler) syncMetrics(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.backendResource.Spec.Metrics))
	for systemName := range t.backendResource.Spec.Metrics {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MetricItem{}
	existingList, err := t.backendAPIEntity.Metrics()
	if err != nil {
		return fmt.Errorf("Error sync backend metrics [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Metrics))
	for _, existing := range existingList.Metrics {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	//
	// Deleted existing and not desired metrics
	//

	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	notDesiredMap := map[string]threescaleapi.MetricItem{}
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredMap[systemName] = existingMap[systemName]
	}
	err = t.deleteNotDesiredMetricsFrom3scale(notDesiredMap)
	if err != nil {
		return fmt.Errorf("Error sync backend metrics [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	err = t.deleteExternalMetricReferences(notDesiredExistingKeys)
	if err != nil {
		return fmt.Errorf("Error sync backend metrics [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	//
	// Reconcile existing and changed metrics
	//

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]metricData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = metricData{
			item: existingMap[systemName],
			spec: t.backendResource.Spec.Metrics[systemName],
		}
	}

	err = t.reconcileMatchedMetrics(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync backend metrics [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired metrics
	//

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.MetricSpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Metrics map key set
		desiredNewMap[systemName] = t.backendResource.Spec.Metrics[systemName]
	}
	err = t.createNewMetrics(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error sync backend metrics [%s]: %w", t.backendResource.Spec.SystemName, err)
	}

	return nil
}

func (t *BackendThreescaleReconciler) createNewMetrics(desiredNewMap map[string]capabilitiesv1beta1.MetricSpec) error {
	for systemName, metric := range desiredNewMap {
		params := threescaleapi.Params{
			"friendly_name": metric.Name,
			"unit":          metric.Unit,
			"system_name":   systemName,
		}
		if len(metric.Description) > 0 {
			params["description"] = metric.Description
		}
		err := t.backendAPIEntity.CreateMetric(params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *BackendThreescaleReconciler) deleteNotDesiredMetricsFrom3scale(notDesiredMap map[string]threescaleapi.MetricItem) error {
	for _, metric := range notDesiredMap {
		err := t.backendAPIEntity.DeleteMetric(metric.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *BackendThreescaleReconciler) reconcileMatchedMetrics(matchedMap map[string]metricData) error {
	for _, data := range matchedMap {
		params := threescaleapi.Params{}
		if data.spec.Name != data.item.Name {
			params["friendly_name"] = data.spec.Name
		}

		if data.spec.Unit != data.item.Unit {
			params["unit"] = data.spec.Unit
		}

		if data.spec.Description != data.item.Description {
			params["description"] = data.spec.Description
		}

		if len(params) > 0 {
			err := t.backendAPIEntity.UpdateMetric(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error updating backendAPI metric: %w", err)
			}
		}
	}

	return nil
}

func (t *BackendThreescaleReconciler) syncMappingRules(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.backendResource.Spec.MappingRules))
	desiredMap := map[string]capabilitiesv1beta1.MappingRuleSpec{}
	for _, spec := range t.backendResource.Spec.MappingRules {
		key := fmt.Sprintf("%s:%s", spec.HTTPMethod, spec.Pattern)
		desiredKeys = append(desiredKeys, key)
		desiredMap[key] = spec
	}

	existingMap, err := t.getExistingMappingRules()
	if err != nil {
		return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
	}
	existingKeys := make([]string, 0, len(existingMap))
	for existingKey := range existingMap {
		existingKeys = append(existingKeys, existingKey)
	}

	//
	// Deleted existing and not desired mapping rules
	//
	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	notDesiredList := make([]threescaleapi.MappingRuleItem, 0, len(notDesiredExistingKeys))
	for _, key := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingKeys set
		notDesiredList = append(notDesiredList, existingMap[key])
	}
	err = t.processNotDesiredMappingRules(notDesiredList)
	if err != nil {
		return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
	}

	// If existing non-desired mapping rules have been detected we refetch
	// the existing list after deletion to get a more consistent view
	// of the existing rules attributes after the deletion.
	if len(notDesiredList) > 0 {
		existingMap, err = t.getExistingMappingRules()
		if err != nil {
			return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
		}
	}

	// Reconcile desired mapping rules
	// In order of definition in the custom resource. Create or update
	// the MappingRule being processed depending on whether it already exists
	// in the 3scale API. Specified 'position' attribute of the MappingRule
	// always corresponds to the position in the CR's MappingRules array.
	// Even though when creating/updating a MappingRule the existing MappingRule
	// positions in 3scale change, we always compare the desired keys with the
	// existing MappingRule positions at this point. We do not refetch the list.
	// Although this is sort of temporarily inconsistent the result is consistent
	// because we always update positions in ascending order. The worst case of
	// this is that some updates are performed on MappingRules whose positions
	// are already reconciled, which is unneeded. The alternative would be to
	// refetch all the MappingRules each time we create/update a MappingRule,
	// which would be more inefficient than doing potential unneeded updates as
	// it is being done with this implementation.
	// Additionally, in case the relative position MappingRules' between
	// unmodified MappingRules happens temporarily during the reconciliation that
	// is not an issue due to changes are not effective until the user promotes
	// the configuration
	t.logger.V(1).Info("syncMappingRules", "desiredKeys", desiredKeys)
	for desiredIdxZeroBased, desiredKey := range desiredKeys {
		desiredMappingRule := desiredMap[desiredKey]
		// We define the position sent to System starting from one (one-based array)
		// instead of zero-based. The reason for that is that System API does not
		// allow to overwrite an existing MappingRule setting it the position 0.
		// By starting with a minimum value of position 1 we avoid that special
		// implemented behavior by system
		desiredIdx := desiredIdxZeroBased + 1
		if existingMappingRule, ok := existingMap[desiredKey]; ok {
			// Reconcile MappingRule
			t.logger.V(1).Info("syncMappingRules", "desiredMappingRuleToReconcile", desiredKey, "position", desiredIdx)
			err := t.reconcileMappingRuleWithPosition(desiredMappingRule, desiredIdx, existingMappingRule)
			if err != nil {
				return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
			}
		} else {
			// Create MappingRule
			t.logger.V(1).Info("syncMappingRules", "desiredMappingRuleToCreate", desiredKey, "position", desiredIdx)
			err := t.createNewMappingRuleWithPosition(desiredMappingRule, desiredIdx)
			if err != nil {
				return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
			}
		}
	}

	return nil
}

func (t *BackendThreescaleReconciler) processNotDesiredMappingRules(notDesiredList []threescaleapi.MappingRuleItem) error {
	for _, mappingRule := range notDesiredList {
		err := t.backendAPIEntity.DeleteMappingRule(mappingRule.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *BackendThreescaleReconciler) getExistingMappingRules() (map[string]threescaleapi.MappingRuleItem, error) {
	existingMap := map[string]threescaleapi.MappingRuleItem{}
	existingList, err := t.backendAPIEntity.MappingRules()
	if err != nil {
		return nil, fmt.Errorf("Error getting backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
	}
	for _, item := range existingList.MappingRules {
		key := fmt.Sprintf("%s:%s", item.Element.HTTPMethod, item.Element.Pattern)
		existingMap[key] = item.Element
	}

	return existingMap, nil
}

func (t *BackendThreescaleReconciler) reconcileMappingRuleWithPosition(desired capabilitiesv1beta1.MappingRuleSpec, desiredPosition int, existing threescaleapi.MappingRuleItem) error {
	params := threescaleapi.Params{}

	//
	// Reconcile metric or method
	//
	metricID, err := t.backendAPIEntity.FindMethodMetricIDBySystemName(desired.MetricMethodRef)
	if err != nil {
		return fmt.Errorf("Error reconcile backend mapping rule: %w", err)
	}

	if metricID < 0 {
		// Should not happen as metric and method references have been validated and should exists
		return errors.New("backend metric method ref for mapping rule not found")
	}

	if metricID != existing.MetricID {
		params["metric_id"] = strconv.FormatInt(metricID, 10)
	}

	//
	// Reconcile delta
	//
	if desired.Increment != existing.Delta {
		params["delta"] = strconv.Itoa(desired.Increment)
	}

	//
	// Reconcile last
	//
	desiredLastAttribute := false
	if desired.Last != nil {
		desiredLastAttribute = *desired.Last
	}

	if desiredLastAttribute != existing.Last {
		params["last"] = strconv.FormatBool(desiredLastAttribute)
	}

	//
	// Reconcile Position
	//
	if desiredPosition != existing.Position {
		params["position"] = strconv.FormatInt(int64(desiredPosition), 10)
	}

	if len(params) > 0 {
		err := t.backendAPIEntity.UpdateMappingRule(existing.ID, params)
		if err != nil {
			return fmt.Errorf("Error reconcile backend mapping rule: %w", err)
		}
	}

	return nil
}

func (t *BackendThreescaleReconciler) createNewMappingRuleWithPosition(desired capabilitiesv1beta1.MappingRuleSpec, desiredPosition int) error {
	metricID, err := t.backendAPIEntity.FindMethodMetricIDBySystemName(desired.MetricMethodRef)
	if err != nil {
		return fmt.Errorf("Error creating backend [%s] mappingrule: %w", t.backendResource.Spec.SystemName, err)
	}

	if metricID < 0 {
		// Should not happen as metric and method references have been validated and should exists
		return errors.New("backend metric method ref for mapping rule not found")
	}

	params := threescaleapi.Params{
		"pattern":     desired.Pattern,
		"http_method": desired.HTTPMethod,
		"metric_id":   strconv.FormatInt(metricID, 10),
		"delta":       strconv.Itoa(desired.Increment),
	}

	if desired.Last != nil {
		params["last"] = strconv.FormatBool(*desired.Last)
	}

	params["position"] = strconv.FormatInt(int64(desiredPosition), 10)

	err = t.backendAPIEntity.CreateMappingRule(params)
	if err != nil {
		return fmt.Errorf("Error creating backend [%s] mappingrule: %w", t.backendResource.Spec.SystemName, err)
	}
	return nil
}
