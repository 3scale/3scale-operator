package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	appPlanRuleListResourceEndpoint          = "/admin/api/application_plans/%d/pricing_rules.json"
	appPlanRuleListPerMetricResourceEndpoint = "/admin/api/application_plans/%d/metrics/%d/pricing_rules.json"
	appPlanRulePerMetricResourceEndpoint     = "/admin/api/application_plans/%d/metrics/%d/pricing_rules/%d.json"
)

// ListApplicationPlansPricingRules List existing application plans pricing rules for a given application plan
func (c *ThreeScaleClient) ListApplicationPlansPricingRules(planID int64) (*ApplicationPlanPricingRuleList, error) {
	endpoint := fmt.Sprintf(appPlanRuleListResourceEndpoint, planID)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &ApplicationPlanPricingRuleList{}
	err = handleJsonResp(resp, http.StatusOK, list)
	return list, err
}

// CreateApplicationPlanPricingRule Create 3scale application plan pricing rule
func (c *ThreeScaleClient) CreateApplicationPlanPricingRule(planID, metricID int64, params Params) (*ApplicationPlanPricingRule, error) {
	endpoint := fmt.Sprintf(appPlanRuleListPerMetricResourceEndpoint, planID, metricID)

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

	item := &ApplicationPlanPricingRule{}
	err = handleJsonResp(resp, http.StatusCreated, item)
	return item, err
}

// DeleteApplicationPlanPricingRule Delete 3scale application plan pricing rule
func (c *ThreeScaleClient) DeleteApplicationPlanPricingRule(planID, metricID, ruleID int64) error {
	endpoint := fmt.Sprintf(appPlanRulePerMetricResourceEndpoint, planID, metricID, ruleID)

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
