package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	appPlanListResourceEndpoint = "/admin/api/services/%d/application_plans.json"
	appPlanResourceEndpoint     = "/admin/api/services/%d/application_plans/%d.json"
)

// ListApplicationPlansByProduct List existing application plans for a given product
func (c *ThreeScaleClient) ListApplicationPlansByProduct(productID int64) (*ApplicationPlanJSONList, error) {
	endpoint := fmt.Sprintf(appPlanListResourceEndpoint, productID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &ApplicationPlanJSONList{}
	err = handleJsonResp(resp, http.StatusOK, list)
	return list, err
}

// CreateApplicationPlan Create 3scale product application plan
func (c *ThreeScaleClient) CreateApplicationPlan(productID int64, params Params) (*ApplicationPlan, error) {
	endpoint := fmt.Sprintf(appPlanListResourceEndpoint, productID)

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

	item := &ApplicationPlan{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteApplicationPlan Delete 3scale product plan
func (c *ThreeScaleClient) DeleteApplicationPlan(productID, id int64) error {
	endpoint := fmt.Sprintf(appPlanResourceEndpoint, productID, id)

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

// ApplicationPlan Read 3scale product application plan
func (c *ThreeScaleClient) ApplicationPlan(productID, id int64) (*ApplicationPlan, error) {
	endpoint := fmt.Sprintf(appPlanResourceEndpoint, productID, id)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	item := &ApplicationPlan{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}

// UpdateApplicationPlan Update 3scale product application plan
func (c *ThreeScaleClient) UpdateApplicationPlan(productID, id int64, params Params) (*ApplicationPlan, error) {
	endpoint := fmt.Sprintf(appPlanResourceEndpoint, productID, id)

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

	item := &ApplicationPlan{}
	err = handleJsonResp(resp, http.StatusOK, item)
	return item, err
}
