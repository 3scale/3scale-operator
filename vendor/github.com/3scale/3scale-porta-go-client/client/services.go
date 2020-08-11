package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	serviceCreateList   = "/admin/api/services.xml"
	serviceUpdateDelete = "/admin/api/services/%s.xml"
)

func (c *ThreeScaleClient) CreateService(name string) (Service, error) {
	var s Service

	endpoint := serviceCreateList
	values := url.Values{}
	values.Add("name", name)
	values.Add("system_name", name)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return s, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusCreated, &s)
	return s, err
}

// UpdateService         - Update the service. Valid params keys and their purpose are as follows:
// "name"                - Name of the service.
// "description"         - Description of the service
// "support_email"       - New support email.
// "tech_support_email"  - New tech support email.
// "admin_support_email" - New admin support email.
// "deployment_option"   - Deployment option for the gateway: 'hosted' for APIcast hosted, 'self-managed' for APIcast Self-managed option
// "backend_version"     - Authentication mode: '1' for API key, '2' for App Id / App Key, 'oauth' for OAuth mode, 'oidc' for OpenID Connect
func (c *ThreeScaleClient) UpdateService(id string, params Params) (Service, error) {
	var s Service

	endpoint := fmt.Sprintf(serviceUpdateDelete, id)

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return s, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &s)
	return s, err
}

// DeleteService - Delete the service.
// Deleting a service removes all applications and service subscriptions.
func (c *ThreeScaleClient) DeleteService(id string) error {
	endpoint := fmt.Sprintf(serviceUpdateDelete, id)

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

func (c *ThreeScaleClient) ListServices() (ServiceList, error) {
	var sl ServiceList

	ep := serviceCreateList

	req, err := c.buildGetReq(ep)
	if err != nil {
		return sl, httpReqError
	}

	values := url.Values{}
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return sl, err
	}
	defer resp.Body.Close()

	err = handleXMLResp(resp, http.StatusOK, &sl)
	return sl, err
}
