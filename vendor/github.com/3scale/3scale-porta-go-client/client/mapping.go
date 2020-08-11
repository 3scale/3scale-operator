package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateMappingRule - Create API for Mapping Rule endpoint
func (c *ThreeScaleClient) CreateMappingRule(
	svcId string, method string,
	pattern string, delta int, metricId string) (MappingRule, error) {

	var mr MappingRule
	ep := genMrEp(svcId)

	values := url.Values{}
	values.Add("service_id", svcId)
	values.Add("http_method", method)
	values.Add("pattern", pattern)
	values.Add("delta", strconv.Itoa(delta))
	values.Add("metric_id", metricId)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return mr, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return mr, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusCreated, &mr)
	return mr, err
}

// UpdateMetric - Updates a Proxy Mapping Rule
// The proxy object must be updated after a mapping rule update to apply the change to proxy config
// Valid params keys and their purpose are as follows:
// "http_method" - HTTP method
// "pattern"     - Mapping Rule pattern
// "delta"       - Increase the metric by this delta
// "metric_id"   - The metric ID
func (c *ThreeScaleClient) UpdateMappingRule(svcId string, id string, params Params) (MappingRule, error) {
	var m MappingRule

	ep := genMrUpdateEp(svcId, id)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(ep, body)
	if err != nil {
		return m, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return m, err
	}

	err = handleXMLResp(resp, http.StatusOK, &m)
	return m, err
}

// DeleteMappingRule - Deletes a Proxy Mapping Rule.
// The proxy object must be updated after a mapping rule deletion to apply the change to proxy config
func (c *ThreeScaleClient) DeleteMappingRule(svcId string, id string) error {
	ep := genMrUpdateEp(svcId, id)

	body := strings.NewReader("")
	req, err := c.buildDeleteReq(ep, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	return handleXMLResp(resp, http.StatusOK, nil)
}

// ListMappingRule - List API for Mapping Rule endpoint
func (c *ThreeScaleClient) ListMappingRule(svcId string) (MappingRuleList, error) {
	var mrl MappingRuleList
	ep := genMrEp(svcId)

	req, err := c.buildGetReq(ep)
	if err != nil {
		return mrl, httpReqError
	}

	values := url.Values{}
	values.Add("service_id", svcId)

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return mrl, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &mrl)
	return mrl, err
}

func genMrEp(svcId string) string {
	return fmt.Sprintf(mappingRuleEndpoint, svcId)
}

func genMrUpdateEp(svcId string, id string) string {
	return fmt.Sprintf(updateDeleteMappingRuleEndpoint, svcId, id)
}
