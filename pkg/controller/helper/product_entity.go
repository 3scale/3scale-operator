package helper

import (
	"fmt"
	"strings"

	"github.com/3scale/3scale-operator/pkg/helper"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type ProductEntity struct {
	client            *threescaleapi.ThreeScaleClient
	productObj        *threescaleapi.Product
	metrics           *threescaleapi.MetricJSONList
	metricsAndMethods *threescaleapi.MetricJSONList
	methods           *threescaleapi.MethodList
	mappingRules      *threescaleapi.MappingRuleJSONList
	backendUsages     threescaleapi.BackendAPIUsageList
	proxy             *threescaleapi.ProxyJSON
	plans             *threescaleapi.ApplicationPlanJSONList
	logger            logr.Logger
}

func NewProductEntity(obj *threescaleapi.Product, cl *threescaleapi.ThreeScaleClient, logger logr.Logger) *ProductEntity {
	return &ProductEntity{
		productObj: obj,
		client:     cl,
		logger:     logger.WithValues("ProductEntity", obj.Element.ID),
	}
}

func (b *ProductEntity) ID() int64 {
	return b.productObj.Element.ID
}

func (b *ProductEntity) Name() string {
	return b.productObj.Element.Name
}

func (b *ProductEntity) State() string {
	return b.productObj.Element.State
}

func (b *ProductEntity) Description() string {
	return b.productObj.Element.Description
}

func (b *ProductEntity) DeploymentOption() string {
	return b.productObj.Element.DeploymentOption
}

func (b *ProductEntity) BackendVersion() string {
	return b.productObj.Element.DeploymentOption
}

func (b *ProductEntity) Update(params threescaleapi.Params) error {
	b.logger.V(1).Info("Update", "params", params)
	updated, err := b.client.UpdateProduct(b.productObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("product [%s] update request: %w", b.productObj.Element.SystemName, err)
	}

	b.productObj = updated

	return nil
}

func (b *ProductEntity) Methods() (*threescaleapi.MethodList, error) {
	b.logger.V(1).Info("Methods")
	if b.methods == nil {
		methods, err := b.getMethods()
		if err != nil {
			return nil, err
		}
		b.methods = methods
	}
	return b.methods, nil
}

func (b *ProductEntity) CreateMethod(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMethod", "params", params)
	hitsID, err := b.getHitsID()
	if err != nil {
		return err
	}
	_, err = b.client.CreateProductMethod(b.productObj.Element.ID, hitsID, params)
	if err != nil {
		return fmt.Errorf("product [%s] create method: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMethods()
	return nil
}

func (b *ProductEntity) DeleteMethod(id int64) error {
	b.logger.V(1).Info("DeleteMethod", "ID", id)
	hitsID, err := b.getHitsID()
	if err != nil {
		return err
	}
	err = b.client.DeleteProductMethod(b.productObj.Element.ID, hitsID, id)
	if err != nil {
		return fmt.Errorf("product [%s] delete method: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMethods()
	return nil
}

func (b *ProductEntity) UpdateMethod(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMethod", "ID", id, "params", params)
	hitsID, err := b.getHitsID()
	if err != nil {
		return err
	}
	_, err = b.client.UpdateProductMethod(b.productObj.Element.ID, hitsID, id, params)
	if err != nil {
		return fmt.Errorf("product [%s] update method: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMethods()
	return nil
}

func (b *ProductEntity) CreateMetric(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMetric", "params", params)
	_, err := b.client.CreateProductMetric(b.productObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("product [%s] create metric: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMetrics()
	return nil
}

func (b *ProductEntity) DeleteMetric(id int64) error {
	b.logger.V(1).Info("DeleteMetric", "ID", id)
	err := b.client.DeleteProductMetric(b.productObj.Element.ID, id)
	if err != nil {
		return fmt.Errorf("product [%s] delete metric: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMetrics()
	return nil
}

func (b *ProductEntity) UpdateMetric(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMethod", "ID", id, "params", params)
	_, err := b.client.UpdateProductMetric(b.productObj.Element.ID, id, params)
	if err != nil {
		return fmt.Errorf("product [%s] update metric: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMetrics()
	return nil
}

func (b *ProductEntity) MetricsAndMethods() (*threescaleapi.MetricJSONList, error) {
	if b.metricsAndMethods == nil {
		metricsAndMethods, err := b.getMetricsAndMethods()
		if err != nil {
			return nil, err
		}
		b.metricsAndMethods = metricsAndMethods
	}
	return b.metricsAndMethods, nil
}

func (b *ProductEntity) Metrics() (*threescaleapi.MetricJSONList, error) {
	if b.metrics == nil {
		metrics, err := b.getMetrics()
		if err != nil {
			return nil, err
		}
		b.metrics = metrics
	}
	return b.metrics, nil
}

func SanitizeProductSystemName(systemName string) string {
	lastIndex := strings.LastIndex(systemName, ".")
	if lastIndex < 0 {
		return systemName
	}

	return systemName[:lastIndex]
}

func (b *ProductEntity) MappingRules() (*threescaleapi.MappingRuleJSONList, error) {
	if b.mappingRules == nil {
		mappingRules, err := b.getMappingRules()
		if err != nil {
			return nil, err
		}
		b.mappingRules = mappingRules
	}
	return b.mappingRules, nil
}

func (b *ProductEntity) DeleteMappingRule(id int64) error {
	b.logger.V(1).Info("DeleteMappingRule", "ID", id)
	err := b.client.DeleteProductMappingRule(b.productObj.Element.ID, id)
	if err != nil {
		return fmt.Errorf("product [%s] delete mapping rule: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMappingRules()
	return nil
}

func (b *ProductEntity) CreateMappingRule(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMappingRule", "params", params)
	_, err := b.client.CreateProductMappingRule(b.productObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("product [%s] create mappingrule: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMappingRules()
	return nil
}

func (b *ProductEntity) UpdateMappingRule(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMappingRule", "ID", id, "params", params)
	_, err := b.client.UpdateProductMappingRule(b.productObj.Element.ID, id, params)
	if err != nil {
		return fmt.Errorf("product [%s] update mappingrule: %w", b.productObj.Element.SystemName, err)
	}
	b.resetMappingRules()
	return nil
}

// FindMethodMetricIDBySystemName returns metric or method ID from system name.
// -1 if metric and method is not found
func (b *ProductEntity) FindMethodMetricIDBySystemName(systemName string) (int64, error) {
	metricsMethodList, err := b.MetricsAndMethods()
	if err != nil {
		return -1, err
	}

	for _, metric := range metricsMethodList.Metrics {
		if metric.Element.SystemName == systemName {
			return metric.Element.ID, nil
		}
	}

	return -1, nil
}

func (b *ProductEntity) BackendUsages() (threescaleapi.BackendAPIUsageList, error) {
	b.logger.V(1).Info("BackendUsages")
	if b.backendUsages == nil {
		backendUsages, err := b.getBackendUsages()
		if err != nil {
			return nil, err
		}
		b.backendUsages = backendUsages
	}
	return b.backendUsages, nil
}

func (b *ProductEntity) DeleteBackendUsage(id int64) error {
	b.logger.V(1).Info("DeleteBackendUsage", "ID", id)
	err := b.client.DeleteBackendapiUsage(b.productObj.Element.ID, id)
	if err != nil {
		return fmt.Errorf("product [%s] delete backendusage: %w", b.productObj.Element.SystemName, err)
	}
	b.resetBackendUsages()
	return nil
}

func (b *ProductEntity) UpdateBackendUsage(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateBackendUsage", "ID", id, "params", params)
	_, err := b.client.UpdateBackendapiUsage(b.productObj.Element.ID, id, params)
	if err != nil {
		return fmt.Errorf("product [%s] update backendusage: %w", b.productObj.Element.SystemName, err)
	}
	b.resetBackendUsages()
	return nil
}

func (b *ProductEntity) CreateBackendUsage(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateBackendUsage", "params", params)
	_, err := b.client.CreateBackendapiUsage(b.productObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("product [%s] update backendusage: %w", b.productObj.Element.SystemName, err)
	}
	b.resetBackendUsages()
	return nil
}

func (b *ProductEntity) Proxy() (*threescaleapi.ProxyJSON, error) {
	b.logger.V(1).Info("Proxy")
	if b.proxy == nil {
		proxy, err := b.getProxy()
		if err != nil {
			return nil, err
		}
		b.proxy = proxy
	}
	return b.proxy, nil
}

func (b *ProductEntity) UpdateProxy(params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateProxy", "params", params)
	updated, err := b.client.UpdateProductProxy(b.productObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("product [%s] update proxy: %w", b.productObj.Element.SystemName, err)
	}

	b.proxy = updated
	return nil
}

func (b *ProductEntity) ApplicationPlans() (*threescaleapi.ApplicationPlanJSONList, error) {
	b.logger.V(1).Info("ApplicationPlans")
	if b.plans == nil {
		plans, err := b.getApplicationPlans()
		if err != nil {
			return nil, err
		}
		b.plans = plans
	}
	return b.plans, nil
}

func (b *ProductEntity) DeleteApplicationPlan(id int64) error {
	b.logger.V(1).Info("DeleteApplicationPlan", "ID", id)
	err := b.client.DeleteApplicationPlan(b.productObj.Element.ID, id)
	if err != nil {
		return fmt.Errorf("product [%s] delete applicationPlan: %w", b.productObj.Element.SystemName, err)
	}
	b.resetApplicationPlans()
	return nil
}

func (b *ProductEntity) CreateApplicationPlan(params threescaleapi.Params) (*threescaleapi.ApplicationPlan, error) {
	b.logger.V(1).Info("CreateApplicationPlan", "params", params)
	obj, err := b.client.CreateApplicationPlan(b.productObj.Element.ID, params)
	if err != nil {
		return nil, fmt.Errorf("product [%s] create plan: %w", b.productObj.Element.SystemName, err)
	}
	b.resetApplicationPlans()
	return obj, nil
}

func (b *ProductEntity) PromoteProxyToStaging() error {
	b.logger.V(1).Info("PromoteProxyToStaging")
	proxyObj, err := b.client.DeployProductProxy(b.productObj.Element.ID)
	if err != nil {
		return fmt.Errorf("product [%s] promote proxy to staging: %w", b.productObj.Element.SystemName, err)
	}

	b.proxy = proxyObj
	return nil
}

//
// PRIVATE
//
//

func (b *ProductEntity) resetBackendUsages() {
	b.backendUsages = nil
}

func (b *ProductEntity) resetMethods() {
	b.metricsAndMethods = nil
	b.methods = nil
}

func (b *ProductEntity) resetMetrics() {
	b.metricsAndMethods = nil
	b.metrics = nil
}

func (b *ProductEntity) resetMappingRules() {
	b.mappingRules = nil
}

func (b *ProductEntity) resetApplicationPlans() {
	b.plans = nil
}

func (b *ProductEntity) getMethods() (*threescaleapi.MethodList, error) {
	b.logger.V(1).Info("getMethods")
	hitsID, err := b.getHitsID()
	if err != nil {
		return nil, err
	}
	methodList, err := b.client.ListProductMethods(b.productObj.Element.ID, hitsID)
	if err != nil {
		return nil, fmt.Errorf("product [%s] get methods: %w", b.productObj.Element.SystemName, err)
	}

	return methodList, nil
}

func (b *ProductEntity) getMetrics() (*threescaleapi.MetricJSONList, error) {
	b.logger.V(1).Info("getMetrics")
	metricsAndMethods, err := b.MetricsAndMethods()
	if err != nil {
		return nil, err
	}
	metricsAndMethodsKeys := make([]string, 0, len(metricsAndMethods.Metrics))
	for _, metric := range metricsAndMethods.Metrics {
		metricsAndMethodsKeys = append(metricsAndMethodsKeys, metric.Element.SystemName)
	}

	methods, err := b.Methods()
	if err != nil {
		return nil, err
	}
	methodsKeys := make([]string, 0, len(methods.Methods))
	for _, method := range methods.Methods {
		methodsKeys = append(methodsKeys, method.Element.SystemName)
	}

	metricKeys := helper.ArrayStringDifference(metricsAndMethodsKeys, methodsKeys)
	metricList := &threescaleapi.MetricJSONList{
		Metrics: make([]threescaleapi.MetricJSON, 0, len(metricKeys)),
	}
	for _, systemName := range metricKeys {
		for _, metric := range metricsAndMethods.Metrics {
			if metric.Element.SystemName == systemName {
				metricList.Metrics = append(metricList.Metrics, metric)
			}
		}
	}

	return metricList, nil
}

func (b *ProductEntity) getMetricsAndMethods() (*threescaleapi.MetricJSONList, error) {
	b.logger.V(1).Info("getMetricsAndMethods")
	metricList, err := b.client.ListProductMetrics(b.productObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("product [%s] get metrics: %w", b.productObj.Element.SystemName, err)
	}

	return metricList, nil
}

func (b *ProductEntity) getHitsID() (int64, error) {
	b.logger.V(1).Info("getHitsID")
	list, err := b.MetricsAndMethods()
	if err != nil {
		return 0, err
	}

	for _, metric := range list.Metrics {
		if metric.Element.SystemName == "hits" {
			return metric.Element.ID, nil
		}
	}

	return 0, fmt.Errorf("product [%s] hits not found", b.productObj.Element.SystemName)
}

func (b *ProductEntity) getMappingRules() (*threescaleapi.MappingRuleJSONList, error) {
	b.logger.V(1).Info("getMappingRules")
	list, err := b.client.ListProductMappingRules(b.productObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("product [%s] get mapping rules: %w", b.productObj.Element.SystemName, err)
	}

	return list, nil
}

func (b *ProductEntity) getBackendUsages() (threescaleapi.BackendAPIUsageList, error) {
	b.logger.V(1).Info("getBackendUsages")
	list, err := b.client.ListBackendapiUsages(b.productObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("product [%s] get backendUsages: %w", b.productObj.Element.SystemName, err)
	}

	return list, nil
}

func (b *ProductEntity) getProxy() (*threescaleapi.ProxyJSON, error) {
	b.logger.V(1).Info("getProxy")
	obj, err := b.client.ProductProxy(b.productObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("product [%s] get proxy: %w", b.productObj.Element.SystemName, err)
	}

	return obj, nil
}

func (b *ProductEntity) getApplicationPlans() (*threescaleapi.ApplicationPlanJSONList, error) {
	b.logger.V(1).Info("getApplicationPlans")
	list, err := b.client.ListApplicationPlansByProduct(b.productObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("product [%s] get plans: %w", b.productObj.Element.SystemName, err)
	}

	return list, nil
}
