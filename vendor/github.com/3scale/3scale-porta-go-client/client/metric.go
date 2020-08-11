package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CreateMetric - Creates a metric on a service. All metrics are scoped by service.
func (c *ThreeScaleClient) CreateMetric(svcId string, name string, description string, unit string) (Metric, error) {
	var m Metric

	ep := genMetricCreateListEp(svcId)

	values := url.Values{}
	values.Add("service_id", svcId)
	values.Add("friendly_name", name)
	values.Add("description", description)
	values.Add("unit", unit)

	body := strings.NewReader(values.Encode())

	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return m, httpReqError
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return m, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusCreated, &m)
	return m, err
}

// UpdateMetric - Updates the metric of a service. Valid params keys and their purpose are as follows:
// "friendly_name" - Name of the metric.
// "unit" - Measure unit of the metric.
// "description" - Description of the metric.
func (c *ThreeScaleClient) UpdateMetric(svcId string, id string, params Params) (Metric, error) {
	var m Metric

	ep := genMetricUpdateDeleteEp(svcId, id)

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
	if err != nil {
		return m, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &m)
	return m, err
}

// DeleteMetric - Deletes the metric of a service.
// When a metric is deleted, the associated limits across application plans are removed
func (c *ThreeScaleClient) DeleteMetric(svcId string, id string) error {
	ep := genMetricUpdateDeleteEp(svcId, id)

	body := strings.NewReader("")
	req, err := c.buildDeleteReq(ep, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleXMLResp(resp, http.StatusOK, nil)
}

// ListMetric - Returns the list of metrics of a service
func (c *ThreeScaleClient) ListMetrics(svcId string) (MetricList, error) {
	var ml MetricList

	ep := genMetricCreateListEp(svcId)

	req, err := c.buildGetReq(ep)
	if err != nil {
		return ml, httpReqError
	}

	values := url.Values{}
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return ml, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &ml)
	return ml, err
}

func genMetricCreateListEp(svcID string) string {
	return fmt.Sprintf(createListMetricEndpoint, svcID)
}

func genMetricUpdateDeleteEp(svcID string, metricId string) string {
	return fmt.Sprintf(updateDeleteMetricEndpoint, svcID, metricId)
}
