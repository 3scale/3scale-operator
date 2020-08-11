package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	limitAppPlanCreate           = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	limitAppPlanList             = "/admin/api/application_plans/%s/limits.xml"
	limitAppPlanUpdateDelete     = "/admin/api/application_plans/%s/metrics/%s/limits/%s.xml "
	limitAppPlanMetricList       = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	limitEndUserPlanCreateList   = "/admin/api/end_user_plans/%s/metrics/%s/limits.xml"
	limitEndUserPlanUpdateDelete = "/admin/api/end_user_plans/%s/metrics/%s/limits/%s.xml"

	// JSON endpoints
	appPlanLimitListResourceEndpoint          = "/admin/api/application_plans/%d/limits.json"
	appPlanLimitListPerMetricResourceEndpoint = "/admin/api/application_plans/%d/metrics/%d/limits.json"
	appPlanLimitPerMetricResourceEndpoint     = "/admin/api/application_plans/%d/metrics/%d/limits/%d.json"
)

// CreateLimitAppPlan - Adds a limit to a metric of an application plan.
// All applications with the application plan (application_plan_id) will be constrained by this new limit on the metric (metric_id).
// Deprecated. Use CreateApplicationPlanLimit instead
func (c *ThreeScaleClient) CreateLimitAppPlan(appPlanId string, metricId string, period string, value int) (Limit, error) {
	endpoint := fmt.Sprintf(limitAppPlanCreate, appPlanId, metricId)

	values := url.Values{}
	values.Add("application_plan_id", appPlanId)

	return c.limitCreate(endpoint, metricId, period, value, values)
}

// CreateLimitEndUserPlan - Adds a limit to a metric of an end user plan
// All applications with the application plan (end_user_plan_id) will be constrained by this new limit on the metric (metric_id).
// Deprecated. End User plans are deprecated
func (c *ThreeScaleClient) CreateLimitEndUserPlan(endUserPlanId string, metricId string, period string, value int) (Limit, error) {
	endpoint := fmt.Sprintf(limitEndUserPlanCreateList, endUserPlanId, metricId)

	values := url.Values{}
	values.Add("end_user_plan_id", endUserPlanId)

	return c.limitCreate(endpoint, metricId, period, value, values)
}

// UpdateLimitsPerPlan - Updates a limit on a metric of an end user plan
// Valid params keys and their purpose are as follows:
// "period" - Period of the limit
// "value"  - Value of the limit
// Deprecated. Use UpdateApplicationPlanLimit instead
func (c *ThreeScaleClient) UpdateLimitPerAppPlan(appPlanId string, metricId string, limitId string, p Params) (Limit, error) {
	endpoint := fmt.Sprintf(limitAppPlanUpdateDelete, appPlanId, metricId, limitId)
	return c.updateLimit(endpoint, p)
}

// UpdateLimitsPerMetric - Updates a limit on a metric of an application plan
// Valid params keys and their purpose are as follows:
// "period" - Period of the limit
// "value"  - Value of the limit
// Deprecated. End User plans are deprecated
func (c *ThreeScaleClient) UpdateLimitPerEndUserPlan(userPlanId string, metricId string, limitId string, p Params) (Limit, error) {
	endpoint := fmt.Sprintf(limitEndUserPlanUpdateDelete, userPlanId, metricId, limitId)
	return c.updateLimit(endpoint, p)
}

// DeleteLimitPerAppPlan - Deletes a limit on a metric of an application plan
// Deprecated. Use DeleteApplicationPlanLimit instead
func (c *ThreeScaleClient) DeleteLimitPerAppPlan(appPlanId string, metricId string, limitId string) error {
	endpoint := fmt.Sprintf(limitAppPlanUpdateDelete, appPlanId, metricId, limitId)
	return c.deleteLimit(endpoint)
}

// DeleteLimitPerEndUserPlan - Deletes a limit on a metric of an end user plan
// Deprecated. End User plans are deprecated
func (c *ThreeScaleClient) DeleteLimitPerEndUserPlan(userPlanId string, metricId string, limitId string) error {
	endpoint := fmt.Sprintf(limitEndUserPlanUpdateDelete, userPlanId, metricId, limitId)
	return c.deleteLimit(endpoint)
}

// ListLimitsPerAppPlan - Returns the list of all limits associated to an application plan.
// Deprecated. Use ListApplicationPlansLimits instead
func (c *ThreeScaleClient) ListLimitsPerAppPlan(appPlanId string) (LimitList, error) {
	endpoint := fmt.Sprintf(limitAppPlanList, appPlanId)
	return c.listLimits(endpoint)
}

// ListLimitsPerEndUserPlan - Returns the list of all limits associated to an end user plan.
// Deprecated. End User plans are deprecated
func (c *ThreeScaleClient) ListLimitsPerEndUserPlan(endUserPlanId string, metricId string) (LimitList, error) {
	endpoint := fmt.Sprintf(limitEndUserPlanCreateList, endUserPlanId, metricId)
	return c.listLimits(endpoint)
}

// ListLimitsPerMetric - Returns the list of all limits associated to a metric of an application plan
func (c *ThreeScaleClient) ListLimitsPerMetric(appPlanId string, metricId string) (LimitList, error) {
	endpoint := fmt.Sprintf(limitAppPlanMetricList, appPlanId, metricId)
	return c.listLimits(endpoint)
}

func (c *ThreeScaleClient) limitCreate(ep string, metricId string, period string, value int, values url.Values) (Limit, error) {
	var apiResp Limit

	values.Add("metric_id", metricId)
	values.Add("period", period)
	values.Add("value", strconv.Itoa(value))

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return apiResp, httpReqError
	}
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return apiResp, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusCreated, &apiResp)
	return apiResp, err
}

func (c *ThreeScaleClient) updateLimit(ep string, p Params) (Limit, error) {
	var l Limit
	values := url.Values{}
	for k, v := range p {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(ep, body)
	if err != nil {
		return l, httpReqError
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return l, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &l)
	return l, err
}

func (c *ThreeScaleClient) deleteLimit(ep string) error {
	values := url.Values{}
	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(ep, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return handleXMLResp(resp, http.StatusCreated, nil)
}

// listLimits takes an endpoint and returns a list of limits
func (c *ThreeScaleClient) listLimits(ep string) (LimitList, error) {
	var ml LimitList

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

// ListApplicationPlansLimits List existing application plan limits for a given application plan
func (c *ThreeScaleClient) ListApplicationPlansLimits(planID int64) (*ApplicationPlanLimitList, error) {
	endpoint := fmt.Sprintf(appPlanLimitListResourceEndpoint, planID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &ApplicationPlanLimitList{}
	err = handleJsonResp(resp, http.StatusOK, list)
	return list, err
}

// CreateApplicationPlanLimit Create 3scale application plan limit
func (c *ThreeScaleClient) CreateApplicationPlanLimit(planID, metricID int64, params Params) (*ApplicationPlanLimit, error) {
	endpoint := fmt.Sprintf(appPlanLimitListPerMetricResourceEndpoint, planID, metricID)

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

	item := &ApplicationPlanLimit{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteApplicationPlanLimit Delete 3scale application plan limit
func (c *ThreeScaleClient) DeleteApplicationPlanLimit(planID, metricID, limitID int64) error {
	endpoint := fmt.Sprintf(appPlanLimitPerMetricResourceEndpoint, planID, metricID, limitID)

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

// ApplicationPlanLimit Read 3scale application plan limit
func (c *ThreeScaleClient) ApplicationPlanLimit(planID, metricID, limitID int64) (*ApplicationPlanLimit, error) {
	endpoint := fmt.Sprintf(appPlanLimitPerMetricResourceEndpoint, planID, metricID, limitID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &ApplicationPlanLimit{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// UpdateApplicationPlanLimit Update 3scale application plan limit
func (c *ThreeScaleClient) UpdateApplicationPlanLimit(planID, metricID, limitID int64, params Params) (*ApplicationPlanLimit, error) {
	endpoint := fmt.Sprintf(appPlanLimitPerMetricResourceEndpoint, planID, metricID, limitID)

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

	item := &ApplicationPlanLimit{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}
