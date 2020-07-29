package helper

import (
	"fmt"
	"strings"

	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type BackendAPIEntity struct {
	client            *threescaleapi.ThreeScaleClient
	backendAPIObj     *threescaleapi.BackendApi
	metrics           *threescaleapi.MetricJSONList
	metricsAndMethods *threescaleapi.MetricJSONList
	methods           *threescaleapi.MethodList
	mappingRules      *threescaleapi.MappingRuleJSONList
	logger            logr.Logger
}

func NewBackendAPIEntity(backendAPIObj *threescaleapi.BackendApi, client *threescaleapi.ThreeScaleClient, logger logr.Logger) *BackendAPIEntity {
	return &BackendAPIEntity{
		backendAPIObj: backendAPIObj,
		client:        client,
		logger:        logger.WithValues("BackendAPI", backendAPIObj.Element.ID),
	}
}

func (b *BackendAPIEntity) ID() int64 {
	return b.backendAPIObj.Element.ID
}

func (b *BackendAPIEntity) SystemName() string {
	return b.backendAPIObj.Element.SystemName
}

func (b *BackendAPIEntity) Name() string {
	return b.backendAPIObj.Element.Name
}

func (b *BackendAPIEntity) Description() string {
	return b.backendAPIObj.Element.Description
}

func (b *BackendAPIEntity) PrivateEndpoint() string {
	return b.backendAPIObj.Element.PrivateEndpoint
}

func (b *BackendAPIEntity) Update(params threescaleapi.Params) error {
	b.logger.V(1).Info("Update", "params", params)
	updatedBackendAPI, err := b.client.UpdateBackendApi(b.backendAPIObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("backend [%s] update request: %w", b.backendAPIObj.Element.SystemName, err)
	}

	b.backendAPIObj = updatedBackendAPI

	return nil
}

func (b *BackendAPIEntity) Methods() (*threescaleapi.MethodList, error) {
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

func (b *BackendAPIEntity) CreateMethod(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMethod", "params", params)
	hitsID, err := b.getHitsID()
	if err != nil {
		return err
	}
	_, err = b.client.CreateBackendApiMethod(b.backendAPIObj.Element.ID, hitsID, params)
	if err != nil {
		return fmt.Errorf("backend [%s] create method: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMethods()
	return nil
}

func (b *BackendAPIEntity) DeleteMethod(id int64) error {
	b.logger.V(1).Info("DeleteMethod", "ID", id)
	hitsID, err := b.getHitsID()
	if err != nil {
		return err
	}
	err = b.client.DeleteBackendApiMethod(b.backendAPIObj.Element.ID, hitsID, id)
	if err != nil {
		return fmt.Errorf("backend [%s] delete method: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMethods()
	return nil
}

func (b *BackendAPIEntity) UpdateMethod(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMethod", "ID", id, "params", params)
	hitsID, err := b.getHitsID()
	if err != nil {
		return err
	}
	_, err = b.client.UpdateBackendApiMethod(b.backendAPIObj.Element.ID, hitsID, id, params)
	if err != nil {
		return fmt.Errorf("backend [%s] update method: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMethods()
	return nil
}

func (b *BackendAPIEntity) CreateMetric(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMetric", "params", params)
	_, err := b.client.CreateBackendApiMetric(b.backendAPIObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("backend [%s] create metric: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMetrics()
	return nil
}

func (b *BackendAPIEntity) DeleteMetric(id int64) error {
	b.logger.V(1).Info("DeleteMetric", "ID", id)
	err := b.client.DeleteBackendApiMetric(b.backendAPIObj.Element.ID, id)
	if err != nil {
		return fmt.Errorf("backend [%s] delete metric: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMetrics()
	return nil
}

func (b *BackendAPIEntity) UpdateMetric(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMethod", "ID", id, "params", params)
	_, err := b.client.UpdateBackendApiMetric(b.backendAPIObj.Element.ID, id, params)
	if err != nil {
		return fmt.Errorf("backend [%s] update metric: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMetrics()
	return nil
}

func (b *BackendAPIEntity) MetricsAndMethods() (*threescaleapi.MetricJSONList, error) {
	b.logger.V(1).Info("metricsAndMethods")
	if b.metricsAndMethods == nil {
		metricsAndMethods, err := b.getMetricsAndMethods()
		if err != nil {
			return nil, err
		}
		b.metricsAndMethods = metricsAndMethods
	}
	return b.metricsAndMethods, nil
}

func (b *BackendAPIEntity) Metrics() (*threescaleapi.MetricJSONList, error) {
	b.logger.V(1).Info("Metrics")
	if b.metrics == nil {
		metrics, err := b.getMetrics()
		if err != nil {
			return nil, err
		}
		b.metrics = metrics
	}
	return b.metrics, nil
}

func SanitizeBackendSystemName(systemName string) string {
	lastIndex := strings.LastIndex(systemName, ".")
	if lastIndex < 0 {
		return systemName
	}

	return systemName[:lastIndex]
}

func (b *BackendAPIEntity) MappingRules() (*threescaleapi.MappingRuleJSONList, error) {
	b.logger.V(1).Info("MappingRules")
	if b.mappingRules == nil {
		mappingRules, err := b.getMappingRules()
		if err != nil {
			return nil, err
		}
		b.mappingRules = mappingRules
	}
	return b.mappingRules, nil
}

func (b *BackendAPIEntity) DeleteMappingRule(id int64) error {
	b.logger.V(1).Info("DeleteMappingRule", "ID", id)
	err := b.client.DeleteBackendapiMappingRule(b.backendAPIObj.Element.ID, id)
	if err != nil {
		return fmt.Errorf("backend [%s] delete mapping rule: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMappingRules()
	return nil
}

func (b *BackendAPIEntity) CreateMappingRule(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMappingRule", "params", params)
	_, err := b.client.CreateBackendapiMappingRule(b.backendAPIObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("backend [%s] create mappingrule: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMappingRules()
	return nil
}

func (b *BackendAPIEntity) UpdateMappingRule(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMappingRule", "ID", id, "params", params)
	_, err := b.client.UpdateBackendapiMappingRule(b.backendAPIObj.Element.ID, id, params)
	if err != nil {
		return fmt.Errorf("backend [%s] update mappingrule: %w", b.backendAPIObj.Element.SystemName, err)
	}
	b.resetMappingRules()
	return nil
}

// FindMethodMetricIDBySystemName returns metric or method ID from system name.
// -1 if metric and method is not found
func (b *BackendAPIEntity) FindMethodMetricIDBySystemName(systemName string) (int64, error) {
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

//
// PRIVATE
//
//

func (b *BackendAPIEntity) resetMethods() {
	b.metricsAndMethods = nil
	b.methods = nil
}

func (b *BackendAPIEntity) resetMetrics() {
	b.metricsAndMethods = nil
	b.metrics = nil
}

func (b *BackendAPIEntity) resetMappingRules() {
	b.mappingRules = nil
}

func (b *BackendAPIEntity) getMethods() (*threescaleapi.MethodList, error) {
	b.logger.V(1).Info("getMethods")
	hitsID, err := b.getHitsID()
	if err != nil {
		return nil, err
	}
	methodList, err := b.client.ListBackendapiMethods(b.backendAPIObj.Element.ID, hitsID)
	if err != nil {
		return nil, fmt.Errorf("backend [%s] get methods: %w", b.backendAPIObj.Element.SystemName, err)
	}
	sanitizeBackendMethodList(methodList)

	return methodList, nil
}

func (b *BackendAPIEntity) getMetrics() (*threescaleapi.MetricJSONList, error) {
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

func (b *BackendAPIEntity) getMetricsAndMethods() (*threescaleapi.MetricJSONList, error) {
	b.logger.V(1).Info("getMetricsAndMethods")
	metricList, err := b.client.ListBackendapiMetrics(b.backendAPIObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("backend [%s] get metrics: %w", b.backendAPIObj.Element.SystemName, err)
	}

	sanitizeBackendMetricList(metricList)
	return metricList, nil
}

func (b *BackendAPIEntity) getHitsID() (int64, error) {
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

	return 0, fmt.Errorf("backend [%s] hits not found", b.backendAPIObj.Element.SystemName)
}

// sanitizeBackendMetricList sanitizes systemName from backend metrics
// Update is made in place
// "system_name": "hits.45498" -> "system_name": "hits"
func sanitizeBackendMetricList(list *threescaleapi.MetricJSONList) {
	for i := range list.Metrics {
		list.Metrics[i].Element.SystemName = SanitizeBackendSystemName(list.Metrics[i].Element.SystemName)
	}
}

// sanitizeBackendMethodList sanitizes systemName from backend methods
// Update is made in place
// "system_name": "backend01.45498" -> "system_name": "backend01"
func sanitizeBackendMethodList(list *threescaleapi.MethodList) {
	for i := range list.Methods {
		list.Methods[i].Element.SystemName = SanitizeBackendSystemName(list.Methods[i].Element.SystemName)
	}
}

func (b *BackendAPIEntity) getMappingRules() (*threescaleapi.MappingRuleJSONList, error) {
	b.logger.V(1).Info("getMappingRules")
	list, err := b.client.ListBackendapiMappingRules(b.backendAPIObj.Element.ID)
	if err != nil {
		return nil, fmt.Errorf("backend [%s] get mapping rules: %w", b.backendAPIObj.Element.SystemName, err)
	}

	return list, nil
}
