package product

import (
	"errors"
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type methodData struct {
	item threescaleapi.MethodItem
	spec capabilitiesv1beta1.Methodpec
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
	resource            *capabilitiesv1beta1.Product
	entity              *helper.ProductEntity
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func NewThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.Product, threescaleAPIClient *threescaleapi.ThreeScaleClient) *ThreescaleReconciler {
	return &ThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		logger:              b.Logger().WithValues("3scale Reconciler", resource.Name),
	}
}

func (t *ThreescaleReconciler) Reconcile() (*helper.ProductEntity, error) {
	taskRunner := helper.NewTaskRunner(nil, t.logger)
	taskRunner.AddTask("SyncProduct", t.syncProduct)
	taskRunner.AddTask("SyncBackendUsage", t.syncBackendUsage)
	taskRunner.AddTask("SyncProxy", t.syncProxy)
	// First methods and metrics, then mapping rules.
	// Mapping rules reference methods and metrics.
	// When a method/metric is deleted,
	// any orphan mapping rule will be deleted automatically by 3scale
	taskRunner.AddTask("SyncMethods", t.syncMethods)
	taskRunner.AddTask("SyncMetrics", t.syncMetrics)
	taskRunner.AddTask("SyncMappingRules", t.syncMappingRules)
	taskRunner.AddTask("SyncApplicationPlans", t.syncApplicationPlans)
	taskRunner.AddTask("SyncPolicies", t.syncPolicies)
	// This should be the last step
	taskRunner.AddTask("BumbProxyVersion", t.bumpProxyVersion)

	err := taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.entity, nil
}

func (t *ThreescaleReconciler) syncProduct(_ interface{}) error {
	productList, err := t.threescaleAPIClient.ListProducts()
	if err != nil {
		return fmt.Errorf("Error sync product [%s]: %w", t.resource.Spec.SystemName, err)
	}

	// Find product in the list by system name
	idx, exists := func(pList []threescaleapi.Product) (int, bool) {
		for i, item := range pList {
			if item.Element.SystemName == t.resource.Spec.SystemName {
				return i, true
			}
		}
		return -1, false
	}(productList.Products)

	var productObj *threescaleapi.Product
	if exists {
		productObj = &productList.Products[idx]
	} else {
		// Create product using system_name.
		// it cannot be modified later
		params := threescaleapi.Params{
			"system_name": t.resource.Spec.SystemName,
		}
		product, err := t.threescaleAPIClient.CreateProduct(t.resource.Spec.Name, params)
		if err != nil {
			return fmt.Errorf("Error creating product %s: %w", t.resource.Spec.SystemName, err)
		}

		productObj = product
	}

	// Will be used by coming steps
	t.entity = helper.NewProductEntity(productObj, t.threescaleAPIClient, t.logger)

	params := threescaleapi.Params{}

	if productObj.Element.Name != t.resource.Spec.Name {
		params["name"] = t.resource.Spec.Name
	}

	if productObj.Element.Description != t.resource.Spec.Description {
		params["description"] = t.resource.Spec.Description
	}

	specDeploymentOption := t.resource.Spec.DeploymentOption()
	if specDeploymentOption != nil {
		if productObj.Element.DeploymentOption != *specDeploymentOption {
			params["deployment_option"] = *specDeploymentOption
		}
	} // only update deployment_option when set in the CR

	specAuthMode := t.resource.Spec.AuthenticationMode()
	if specAuthMode != nil {
		if productObj.Element.BackendVersion != *specAuthMode {
			params["backend_version"] = *specAuthMode
		}
	} // only update backend_version when set in the CR

	if len(params) > 0 {
		err = t.entity.Update(params)
		if err != nil {
			return fmt.Errorf("Error sync product [%s;%d]: %w", t.resource.Spec.SystemName, productObj.Element.ID, err)
		}
	}

	return nil
}

func (t *ThreescaleReconciler) syncMethods(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.Methods))
	for systemName := range t.resource.Spec.Methods {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MethodItem{}
	existingList, err := t.entity.Methods()
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Methods))
	for _, existing := range existingList.Methods {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.Methodpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Method map key set
		desiredNewMap[systemName] = t.resource.Spec.Methods[systemName]
	}
	err = t.createNewMethods(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	notDesiredMap := map[string]threescaleapi.MethodItem{}
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredMap[systemName] = existingMap[systemName]
	}
	err = t.processNotDesiredMethods(notDesiredMap)
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]methodData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = methodData{
			item: existingMap[systemName],
			spec: t.resource.Spec.Methods[systemName],
		}
	}

	err = t.reconcileMatchedMethods(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMethods(desiredNewMap map[string]capabilitiesv1beta1.Methodpec) error {
	for systemName, method := range desiredNewMap {
		params := threescaleapi.Params{
			"friendly_name": method.Name,
			"system_name":   systemName,
		}
		if len(method.Description) > 0 {
			params["description"] = method.Description
		}
		err := t.entity.CreateMethod(params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMethods(notDesiredMap map[string]threescaleapi.MethodItem) error {
	for _, notDesiredMethod := range notDesiredMap {
		err := t.entity.DeleteMethod(notDesiredMethod.ID)
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
			err := t.entity.UpdateMethod(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error reconcile product methods: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) syncMetrics(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.Metrics))
	for systemName := range t.resource.Spec.Metrics {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MetricItem{}
	existingList, err := t.entity.Metrics()
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
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
	err = t.processNotDesiredMetrics(notDesiredMap)
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Reconcile existing and changed metrics
	//

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]metricData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = metricData{
			item: existingMap[systemName],
			spec: t.resource.Spec.Metrics[systemName],
		}
	}

	err = t.reconcileMatchedMetrics(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired metrics
	//

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.MetricSpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Metrics map key set
		desiredNewMap[systemName] = t.resource.Spec.Metrics[systemName]
	}
	err = t.createNewMetrics(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
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
		err := t.entity.CreateMetric(params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMetrics(notDesiredMap map[string]threescaleapi.MetricItem) error {
	for _, metric := range notDesiredMap {
		err := t.entity.DeleteMetric(metric.ID)
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
			err := t.entity.UpdateMetric(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error updating product metric: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) syncMappingRules(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.MappingRules))
	desiredMap := map[string]capabilitiesv1beta1.MappingRuleSpec{}
	for _, spec := range t.resource.Spec.MappingRules {
		key := fmt.Sprintf("%s:%s", spec.HTTPMethod, spec.Pattern)
		desiredKeys = append(desiredKeys, key)
		desiredMap[key] = spec
	}

	existingKeys := []string{}
	existingMap := map[string]threescaleapi.MappingRuleItem{}
	existingList, err := t.entity.MappingRules()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
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
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
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
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
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
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMappingRules(notDesiredList []threescaleapi.MappingRuleItem) error {
	for _, mappingRule := range notDesiredList {
		err := t.entity.DeleteMappingRule(mappingRule.ID)
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
		metricID, err := t.entity.FindMethodMetricIDBySystemName(data.spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("Error reconcile product mapping rule: %w", err)
		}

		if metricID < 0 {
			// Should not happen as metric and method references have been validated and should exists
			return errors.New("product metric method ref for mapping rule not found")
		}

		if metricID != data.item.MetricID {
			params["metric_id"] = strconv.FormatInt(metricID, 10)
		}

		//
		// Reconcile delta
		//
		if *data.spec.Increment != data.item.Delta {
			params["delta"] = strconv.Itoa(*data.spec.Increment)
		}

		if len(params) > 0 {
			err := t.entity.UpdateMappingRule(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error reconcile product mapping rule: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMappingRules(desiredList []capabilitiesv1beta1.MappingRuleSpec) error {
	for _, spec := range desiredList {
		metricID, err := t.entity.FindMethodMetricIDBySystemName(spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("Error creating product [%s] mappingrule: %w", t.resource.Spec.SystemName, err)
		}

		if metricID < 0 {
			// Should not happen as metric and method references have been validated and should exists
			return errors.New("product metric method ref for mapping rule not found")
		}

		params := threescaleapi.Params{
			"pattern":     spec.Pattern,
			"http_method": spec.HTTPMethod,
			"metric_id":   strconv.FormatInt(metricID, 10),
			// Defaults are set in the spec, should not be nil
			"delta": strconv.Itoa(*spec.Increment),
		}

		err = t.entity.CreateMappingRule(params)
		if err != nil {
			return fmt.Errorf("Error creating product [%s] mappingrule: %w", t.resource.Spec.SystemName, err)
		}
	}
	return nil
}

func (t *ThreescaleReconciler) syncBackendUsage(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) syncProxy(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) syncApplicationPlans(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) syncPolicies(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) bumpProxyVersion(_ interface{}) error {
	return nil
}
