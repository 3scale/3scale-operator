package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	backendListResourceEndpoint       = "/admin/api/backend_apis.json"
	backendResourceEndpoint           = "/admin/api/backend_apis/%d.json"
	backendMethodListResourceEndpoint = "/admin/api/backend_apis/%d/metrics/%d/methods.json"
	backendMethodResourceEndpoint     = "/admin/api/backend_apis/%d/metrics/%d/methods/%d.json"
	backendMetricListResourceEndpoint = "/admin/api/backend_apis/%d/metrics.json"
	backendMetricResourceEndpoint     = "/admin/api/backend_apis/%d/metrics/%d.json"
	backendMRListResourceEndpoint     = "/admin/api/backend_apis/%d/mapping_rules.json"
	backendMRResourceEndpoint         = "/admin/api/backend_apis/%d/mapping_rules/%d.json"
	backendUsageListResourceEndpoint  = "/admin/api/services/%d/backend_usages.json"
	backendUsageResourceEndpoint      = "/admin/api/services/%d/backend_usages/%d.json"
)

// ListBackends List existing backends
func (c *ThreeScaleClient) ListBackendApis() (*BackendApiList, error) {
	req, err := c.buildGetReq(backendListResourceEndpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	backendList := &BackendApiList{}
	err = handleJsonResp(resp, http.StatusOK, backendList)
	return backendList, err
}

// CreateBackendApi Create 3scale Backend
func (c *ThreeScaleClient) CreateBackendApi(params Params) (*BackendApi, error) {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(backendListResourceEndpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	backendApi := &BackendApi{}
	err = handleJsonResp(resp, http.StatusCreated, backendApi)
	return backendApi, err
}

// DeleteBackendApi Delete existing backend
func (c *ThreeScaleClient) DeleteBackendApi(id int64) error {
	backendEndpoint := fmt.Sprintf(backendResourceEndpoint, id)

	req, err := c.buildDeleteReq(backendEndpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleJsonResp(resp, http.StatusOK, nil)
}

// BackendApi Read 3scale Backend
func (c *ThreeScaleClient) BackendApi(id int64) (*BackendApi, error) {
	backendEndpoint := fmt.Sprintf(backendResourceEndpoint, id)

	req, err := c.buildGetReq(backendEndpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	backendAPI := &BackendApi{}
	err = handleJsonResp(resp, http.StatusOK, backendAPI)
	return backendAPI, err
}

// UpdateBackendApi Update 3scale Backend
func (c *ThreeScaleClient) UpdateBackendApi(id int64, params Params) (*BackendApi, error) {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	backendEndpoint := fmt.Sprintf(backendResourceEndpoint, id)

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(backendEndpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	backendAPI := &BackendApi{}
	err = handleJsonResp(resp, http.StatusOK, backendAPI)
	return backendAPI, err
}

// ListBackendapiMethods List existing backend methods
func (c *ThreeScaleClient) ListBackendapiMethods(backendapiID, hitsID int64) (*MethodList, error) {
	endpoint := fmt.Sprintf(backendMethodListResourceEndpoint, backendapiID, hitsID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &MethodList{}
	err = handleJsonResp(resp, http.StatusOK, list)
	return list, err
}

// CreateBackendApiMethod Create 3scale Backend method
func (c *ThreeScaleClient) CreateBackendApiMethod(backendapiID, hitsID int64, params Params) (*Method, error) {
	endpoint := fmt.Sprintf(backendMethodListResourceEndpoint, backendapiID, hitsID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &Method{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteBackendApiMethod Delete 3scale Backend method
func (c *ThreeScaleClient) DeleteBackendApiMethod(backendapiID, hitsID, methodID int64) error {
	endpoint := fmt.Sprintf(backendMethodResourceEndpoint, backendapiID, hitsID, methodID)

	req, err := c.buildDeleteReq(endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleJsonResp(resp, http.StatusOK, nil)
}

// BackendApiMethod Read 3scale Backend method
func (c *ThreeScaleClient) BackendApiMethod(backendapiID, hitsID, methodID int64) (*Method, error) {
	endpoint := fmt.Sprintf(backendMethodResourceEndpoint, backendapiID, hitsID, methodID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &Method{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// UpdateBackendApiMethod Update 3scale Backend method
func (c *ThreeScaleClient) UpdateBackendApiMethod(backendapiID, hitsID, methodID int64, params Params) (*Method, error) {
	endpoint := fmt.Sprintf(backendMethodResourceEndpoint, backendapiID, hitsID, methodID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &Method{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// ListBackendapiMetrics List existing backend metric
func (c *ThreeScaleClient) ListBackendapiMetrics(backendapiID int64) (*MetricJSONList, error) {
	endpoint := fmt.Sprintf(backendMetricListResourceEndpoint, backendapiID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &MetricJSONList{}
	err = handleJsonResp(resp, http.StatusOK, list)
	return list, err
}

// CreateBackendApiMetric Create 3scale Backend metric
func (c *ThreeScaleClient) CreateBackendApiMetric(backendapiID int64, params Params) (*MetricJSON, error) {
	endpoint := fmt.Sprintf(backendMetricListResourceEndpoint, backendapiID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &MetricJSON{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteBackendApiMetric Delete 3scale Backend metric
func (c *ThreeScaleClient) DeleteBackendApiMetric(backendapiID, metricID int64) error {
	endpoint := fmt.Sprintf(backendMetricResourceEndpoint, backendapiID, metricID)

	req, err := c.buildDeleteReq(endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleJsonResp(resp, http.StatusOK, nil)
}

// BackendApiMetric Read 3scale Backend metric
func (c *ThreeScaleClient) BackendApiMetric(backendapiID, metricID int64) (*MetricJSON, error) {
	endpoint := fmt.Sprintf(backendMetricResourceEndpoint, backendapiID, metricID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &MetricJSON{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// UpdateBackendApiMetric Update 3scale Backend metric
func (c *ThreeScaleClient) UpdateBackendApiMetric(backendapiID, metricID int64, params Params) (*MetricJSON, error) {
	endpoint := fmt.Sprintf(backendMetricResourceEndpoint, backendapiID, metricID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &MetricJSON{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// ListBackendapiMappingRules List existing backend mapping rules
func (c *ThreeScaleClient) ListBackendapiMappingRules(backendapiID int64) (*MappingRuleJSONList, error) {
	endpoint := fmt.Sprintf(backendMRListResourceEndpoint, backendapiID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &MappingRuleJSONList{}
	err = handleJsonResp(resp, http.StatusOK, list)
	return list, err
}

// CreateBackendapiMappingRule Create 3scale Backend mappingrule
func (c *ThreeScaleClient) CreateBackendapiMappingRule(backendapiID int64, params Params) (*MappingRuleJSON, error) {
	endpoint := fmt.Sprintf(backendMRListResourceEndpoint, backendapiID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &MappingRuleJSON{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteBackendapiMappingRule Delete 3scale Backend mapping rule
func (c *ThreeScaleClient) DeleteBackendapiMappingRule(backendapiID, mrID int64) error {
	endpoint := fmt.Sprintf(backendMRResourceEndpoint, backendapiID, mrID)

	req, err := c.buildDeleteReq(endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleJsonResp(resp, http.StatusOK, nil)
}

// BackendapiMappingRule Read 3scale Backend mapping rule
func (c *ThreeScaleClient) BackendapiMappingRule(backendapiID, mrID int64) (*MappingRuleJSON, error) {
	endpoint := fmt.Sprintf(backendMRResourceEndpoint, backendapiID, mrID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &MappingRuleJSON{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// UpdateBackendapiMappingRule Update 3scale Backend mapping rule
func (c *ThreeScaleClient) UpdateBackendapiMappingRule(backendapiID, mrID int64, params Params) (*MappingRuleJSON, error) {
	endpoint := fmt.Sprintf(backendMRResourceEndpoint, backendapiID, mrID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &MappingRuleJSON{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// ListBackendapiUsages List existing backend usages for a given product
func (c *ThreeScaleClient) ListBackendapiUsages(productID int64) (BackendAPIUsageList, error) {
	endpoint := fmt.Sprintf(backendUsageListResourceEndpoint, productID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := BackendAPIUsageList{}
	err = handleJsonResp(resp, http.StatusOK, &list)
	return list, err
}

// CreateBackendapiUsage Create 3scale Backend usage
func (c *ThreeScaleClient) CreateBackendapiUsage(productID int64, params Params) (*BackendAPIUsage, error) {
	endpoint := fmt.Sprintf(backendUsageListResourceEndpoint, productID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &BackendAPIUsage{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteBackendapiUsage Delete 3scale Backend usage
func (c *ThreeScaleClient) DeleteBackendapiUsage(productID, backendUsageID int64) error {
	endpoint := fmt.Sprintf(backendUsageResourceEndpoint, productID, backendUsageID)

	req, err := c.buildDeleteReq(endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleJsonResp(resp, http.StatusOK, nil)
}

// BackendapiUsage Read 3scale Backend usage
func (c *ThreeScaleClient) BackendapiUsage(productID, backendUsageID int64) (*BackendAPIUsage, error) {
	endpoint := fmt.Sprintf(backendUsageResourceEndpoint, productID, backendUsageID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &BackendAPIUsage{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// UpdateBackendapiUsage Update 3scale Backend usage
func (c *ThreeScaleClient) UpdateBackendapiUsage(productID, backendUsageID int64, params Params) (*BackendAPIUsage, error) {
	endpoint := fmt.Sprintf(backendUsageResourceEndpoint, productID, backendUsageID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &BackendAPIUsage{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}
