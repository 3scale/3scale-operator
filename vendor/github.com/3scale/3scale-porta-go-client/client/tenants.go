package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	tenantCreate = "/master/api/providers.json"
	tenantRead   = "/master/api/providers/%d.json"
	tenantUpdate = "/master/api/providers/%d.json"
)

// CreateTenant creates new tenant using 3scale API
func (c *ThreeScaleClient) CreateTenant(orgName, username, email, password string) (*Tenant, error) {
	values := url.Values{}
	values.Add("org_name", orgName)
	values.Add("username", username)
	values.Add("email", email)
	values.Add("password", password)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(tenantCreate, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tenant := &Tenant{}
	err = handleJsonResp(resp, http.StatusCreated, tenant)
	return tenant, err
}

// ShowTenant - Returns tenant info for the specified ID
func (c *ThreeScaleClient) ShowTenant(tenantID int64) (*Tenant, error) {
	endpoint := fmt.Sprintf(tenantRead, tenantID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tenant := &Tenant{}
	err = handleJsonResp(resp, http.StatusOK, tenant)
	return tenant, err
}

// UpdateTenant - Updates tenant info for the specified ID
func (c *ThreeScaleClient) UpdateTenant(tenantID int64, params Params) (*Tenant, error) {
	endpoint := fmt.Sprintf(tenantUpdate, tenantID)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	body := strings.NewReader(values.Encode())
	req, err := c.buildPutReq(endpoint, body)
	if err != nil {
		return nil, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tenant := &Tenant{}
	err = handleJsonResp(resp, http.StatusOK, tenant)
	return tenant, err
}

// DeleteTenant - Schedules a tenant account to be permanently deleted in X days (check Porta doc)
func (c *ThreeScaleClient) DeleteTenant(tenantID int64) error {
	endpoint := fmt.Sprintf(tenantUpdate, tenantID)

	req, err := c.buildDeleteReq(endpoint, nil)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleJsonErrResp(resp)
	}

	return nil
}
