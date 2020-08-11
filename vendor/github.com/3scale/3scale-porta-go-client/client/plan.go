package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	appPlanCreate         = "/admin/api/services/%s/application_plans.xml"
	appPlanUpdateDelete   = "/admin/api/services/%s/application_plans/%s.xml"
	appPlansList          = "/admin/api/application_plans.xml"
	appPlansByServiceList = "/admin/api/services/%s/application_plans.xml"
	appPlanSetDefault     = "/admin/api/services/%s/application_plans/%s/default.xml"
)

// CreateAppPlan - Creates an application plan.
// Deprecated. Use CreateApplicationPlan instead
func (c *ThreeScaleClient) CreateAppPlan(svcId string, name string, stateEvent string) (Plan, error) {
	var apiResp Plan
	endpoint := fmt.Sprintf(appPlanCreate, svcId)

	values := url.Values{}
	values.Add("service_id", svcId)
	values.Add("name", name)
	values.Add("state_event", stateEvent)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
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

// UpdateAppPlan - Updates an application plan
// Deprecated. Use UpdateApplicationPlan instead
func (c *ThreeScaleClient) UpdateAppPlan(svcId string, appPlanId string, name string, stateEvent string, params Params) (Plan, error) {
	endpoint := fmt.Sprintf(appPlanUpdateDelete, svcId, appPlanId)

	values := url.Values{}
	values.Add("service_id", svcId)
	values.Add("name", name)

	if stateEvent != "" {
		values.Add("state_event", stateEvent)
	}

	for k, v := range params {
		values.Add(k, v)
	}

	return c.updatePlan(endpoint, values)
}

// DeleteAppPlan - Deletes an application plan
// Deprecated. Use DeleteApplicationPlan instead
func (c *ThreeScaleClient) DeleteAppPlan(svcId string, appPlanId string) error {
	endpoint := fmt.Sprintf(appPlanUpdateDelete, svcId, appPlanId)

	values := url.Values{}

	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(endpoint, body)
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

// ListAppPlanByServiceId - Lists all application plans, filtering on service id
// Deprecated. Use ListApplicationPlansByProduct instead
func (c *ThreeScaleClient) ListAppPlanByServiceId(svcId string) (ApplicationPlansList, error) {
	var appPlans ApplicationPlansList
	endpoint := fmt.Sprintf(appPlansByServiceList, svcId)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return appPlans, httpReqError
	}

	values := url.Values{}
	values.Add("service_id", svcId)

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return appPlans, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &appPlans)
	return appPlans, err
}

// ListAppPlan - List all application plans
func (c *ThreeScaleClient) ListAppPlan() (ApplicationPlansList, error) {
	var appPlans ApplicationPlansList
	endpoint := appPlansList

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return appPlans, httpReqError
	}

	values := url.Values{}

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return appPlans, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &appPlans)
	return appPlans, err
}

// SetDefaultPlan - Makes the application plan the default one
func (c *ThreeScaleClient) SetDefaultPlan(svcId string, id string) (Plan, error) {
	endpoint := fmt.Sprintf(appPlanSetDefault, svcId, id)

	values := url.Values{}
	return c.updatePlan(endpoint, values)
}

func (c *ThreeScaleClient) updatePlan(endpoint string, values url.Values) (Plan, error) {
	var apiResp Plan
	body := strings.NewReader(values.Encode())
	req, err := c.buildPutReq(endpoint, body)
	if err != nil {
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiResp, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &apiResp)
	return apiResp, err
}
