package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	appCreate = "/admin/api/accounts/%s/applications.json"
	appList   = "/admin/api/accounts/%d/applications.json"
)

// CreateApp - Create an application.
// The application object can be extended with Fields Definitions in the Admin Portal where you can add/remove fields
func (c *ThreeScaleClient) CreateApp(accountId string, planId string, name string, description string) (Application, error) {
	var app Application
	endpoint := fmt.Sprintf(appCreate, accountId)

	values := url.Values{}
	values.Add("account_id", accountId)
	values.Add("plan_id", planId)
	values.Add("name", name)
	values.Add("description", description)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return app, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return app, err
	}
	defer resp.Body.Close()

	apiResp := &ApplicationElem{}
	err = handleJsonResp(resp, http.StatusCreated, apiResp)
	if err != nil {
		return app, err
	}
	return apiResp.Application, nil
}

// ListApplications - List of applications for a given account.
func (c *ThreeScaleClient) ListApplications(accountID int64) (*ApplicationList, error) {
	endpoint := fmt.Sprintf(appList, accountID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	applicationList := &ApplicationList{}
	err = handleJsonResp(resp, http.StatusOK, applicationList)
	return applicationList, err
}
