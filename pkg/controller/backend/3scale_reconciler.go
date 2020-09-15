package backend

import (
	"errors"
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type methodData struct {
	item threescaleapi.MethodItem
	spec capabilitiesv1beta1.MethodSpec
}

type metricData struct {
	item threescaleapi.MetricItem
	spec capabilitiesv1beta1.MetricSpec
}

type mappingRuleData struct {
	item threescaleapi.MappingRuleItem
	spec capabilitiesv1beta1.MappingRuleSpec
}

type ThreescaleReconciler struct {
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
) *ThreescaleReconciler {

	return &ThreescaleReconciler{
		BaseReconciler:      b,
		backendResource:     backendResource,
		backendRemoteIndex:  backendRemoteIndex,
		threescaleAPIClient: threescaleAPIClient,
		providerAccount:     providerAccount,
		logger:              b.Logger().WithValues("3scale Reconciler", backendResource.Name),
	}
}

func (t *ThreescaleReconciler) Reconcile() (*controllerhelper.BackendAPIEntity, error) {
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

func (t *ThreescaleReconciler) syncBackend(_ interface{}) error {
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

func (t *ThreescaleReconciler) syncMethods(_ interface{}) error {
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

func (t *ThreescaleReconciler) createNewMethods(desiredNewMap map[string]capabilitiesv1beta1.MethodSpec) error {
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

func (t *ThreescaleReconciler) deleteNotDesiredMethodsFrom3scale(notDesiredMap map[string]threescaleapi.MethodItem) error {
	for _, notDesiredMethod := range notDesiredMap {
		err := t.backendAPIEntity.DeleteMethod(notDesiredMethod.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// valid for metrics and methods as long as 3scale ensures system_names are unique among methods and metrics
func (t *ThreescaleReconciler) deleteExternalMetricReferences(notDesiredMetrics []string) error {
	productList, err := controllerhelper.ProductList(t.backendResource.Namespace, t.Client(), t.providerAccount, t.logger)
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
func (t *ThreescaleReconciler) deleteExternalMetricReferencesOnProduct(notDesiredMetrics []string, productRef capabilitiesv1beta1.Product) error {
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

func (t *ThreescaleReconciler) reconcileMatchedMethods(matchedMap map[string]methodData) error {
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

func (t *ThreescaleReconciler) syncMetrics(_ interface{}) error {
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

func (t *ThreescaleReconciler) createNewMetrics(desiredNewMap map[string]capabilitiesv1beta1.MetricSpec) error {
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

func (t *ThreescaleReconciler) deleteNotDesiredMetricsFrom3scale(notDesiredMap map[string]threescaleapi.MetricItem) error {
	for _, metric := range notDesiredMap {
		err := t.backendAPIEntity.DeleteMetric(metric.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedMetrics(matchedMap map[string]metricData) error {
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

func (t *ThreescaleReconciler) syncMappingRules(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.backendResource.Spec.MappingRules))
	desiredMap := map[string]capabilitiesv1beta1.MappingRuleSpec{}
	for _, spec := range t.backendResource.Spec.MappingRules {
		key := fmt.Sprintf("%s:%s", spec.HTTPMethod, spec.Pattern)
		desiredKeys = append(desiredKeys, key)
		desiredMap[key] = spec
	}

	existingKeys := []string{}
	existingMap := map[string]threescaleapi.MappingRuleItem{}
	existingList, err := t.backendAPIEntity.MappingRules()
	if err != nil {
		return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
	}
	for _, item := range existingList.MappingRules {
		key := fmt.Sprintf("%s:%s", item.Element.HTTPMethod, item.Element.Pattern)
		existingKeys = append(existingKeys, key)
		existingMap[key] = item.Element
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

	//
	// Reconcile existing and changed mapping rules
	//

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedList := make([]mappingRuleData, 0, len(matchedKeys))
	for _, key := range matchedKeys {
		matchedList = append(matchedList, mappingRuleData{
			item: existingMap[key],
			spec: desiredMap[key],
		})
	}

	err = t.reconcileMatchedMappingRules(matchedList)
	if err != nil {
		return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired mapping rules
	//

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewList := make([]capabilitiesv1beta1.MappingRuleSpec, 0, len(desiredNewKeys))
	for _, key := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the desiredKeys set
		desiredNewList = append(desiredNewList, desiredMap[key])
	}
	err = t.createNewMappingRules(desiredNewList)
	if err != nil {
		return fmt.Errorf("Error sync backend [%s] mappingrules: %w", t.backendResource.Spec.SystemName, err)
	}

	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMappingRules(notDesiredList []threescaleapi.MappingRuleItem) error {
	for _, mappingRule := range notDesiredList {
		err := t.backendAPIEntity.DeleteMappingRule(mappingRule.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedMappingRules(matchedList []mappingRuleData) error {
	for _, data := range matchedList {
		params := threescaleapi.Params{}

		//
		// Reconcile metric or method
		//
		metricID, err := t.backendAPIEntity.FindMethodMetricIDBySystemName(data.spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("Error reconcile backend mapping rule: %w", err)
		}

		if metricID < 0 {
			// Should not happen as metric and method references have been validated and should exists
			return errors.New("backend metric method ref for mapping rule not found")
		}

		if metricID != data.item.MetricID {
			params["metric_id"] = strconv.FormatInt(metricID, 10)
		}

		//
		// Reconcile delta
		//
		if data.spec.Increment != data.item.Delta {
			params["delta"] = strconv.Itoa(data.spec.Increment)
		}

		//
		// Reconcile last
		//
		desiredLastAttribute := false
		if data.spec.Last != nil {
			desiredLastAttribute = *data.spec.Last
		}

		if desiredLastAttribute != data.item.Last {
			params["last"] = strconv.FormatBool(desiredLastAttribute)
		}

		if len(params) > 0 {
			err := t.backendAPIEntity.UpdateMappingRule(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error reconcile backend mapping rule: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMappingRules(desiredList []capabilitiesv1beta1.MappingRuleSpec) error {
	for _, spec := range desiredList {
		metricID, err := t.backendAPIEntity.FindMethodMetricIDBySystemName(spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("Error creating backend [%s] mappingrule: %w", t.backendResource.Spec.SystemName, err)
		}

		if metricID < 0 {
			// Should not happen as metric and method references have been validated and should exists
			return errors.New("backend metric method ref for mapping rule not found")
		}

		params := threescaleapi.Params{
			"pattern":     spec.Pattern,
			"http_method": spec.HTTPMethod,
			"metric_id":   strconv.FormatInt(metricID, 10),
			"delta":       strconv.Itoa(spec.Increment),
		}

		if spec.Last != nil {
			params["last"] = strconv.FormatBool(*spec.Last)
		}

		err = t.backendAPIEntity.CreateMappingRule(params)
		if err != nil {
			return fmt.Errorf("Error creating backend [%s] mappingrule: %w", t.backendResource.Spec.SystemName, err)
		}
	}
	return nil
}
