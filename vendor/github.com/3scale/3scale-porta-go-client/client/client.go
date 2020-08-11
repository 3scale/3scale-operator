package client

// This package provides bare minimum functionality for all the endpoints it exposes,
// which is a subset of the Account Management API.

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	mappingRuleEndpoint             = "/admin/api/services/%s/proxy/mapping_rules.xml"
	createListMetricEndpoint        = "/admin/api/services/%s/metrics.xml"
	updateDeleteMetricEndpoint      = "/admin/api/services/%s/metrics/%s.xml"
	updateDeleteMappingRuleEndpoint = "/admin/api/services/%s/proxy/mapping_rules/%s.xml"
)

var httpReqError = errors.New("error building http request")

// Returns a custom AdminPortal which integrates with the users Account Management API.
// Supported schemes are http and https
func NewAdminPortal(scheme string, host string, port int) (*AdminPortal, error) {
	url2, err := verifyUrl(fmt.Sprintf("%s://%s:%d", scheme, host, port))
	if err != nil {
		return nil, err
	}
	return &AdminPortal{scheme, host, port, url2}, nil
}

// Creates a ThreeScaleClient to communicate with Account Management API.
// If http Client is nil, the default http client will be used
func NewThreeScale(backEnd *AdminPortal, credential string, httpClient *http.Client) *ThreeScaleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ThreeScaleClient{backEnd, credential, httpClient}
}

func NewParams() Params {
	params := make(map[string]string)
	return params
}

func (p Params) AddParam(key string, value string) {
	p[key] = value
}

// SetCredentials allow the user to set the client credentials
func (c *ThreeScaleClient) SetCredentials(credential string) {
	c.credential = credential
}

// Request builder for GET request to the provided endpoint
func (c *ThreeScaleClient) buildGetReq(ep string) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("GET", c.adminPortal.baseUrl.ResolveReference(path).String(), nil)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Authorization", "Basic "+basicAuth("", c.credential))
	return req, err
}

// Request builder for POST request to the provided endpoint
func (c *ThreeScaleClient) buildPostReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("POST", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth("", c.credential))
	return req, err
}

// Request builder for PUT request to the provided endpoint
func (c *ThreeScaleClient) buildUpdateReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("PUT", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth("", c.credential))
	return req, err
}

// Request builder for DELETE request to the provided endpoint
func (c *ThreeScaleClient) buildDeleteReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("DELETE", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth("", c.credential))
	return req, err
}

// Request builder for PUT request to the provided endpoint
func (c *ThreeScaleClient) buildPutReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("PUT", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth("", c.credential))
	return req, err
}

// Verifies a custom admin portal is valid
func verifyUrl(urlToCheck string) (*url.URL, error) {
	url2, err := url.ParseRequestURI(urlToCheck)
	if err == nil {
		if url2.Scheme != "http" && url2.Scheme != "https" {
			return url2, fmt.Errorf("unsupported schema %s passed to adminPortal", url2.Scheme)
		}

		if url2.Hostname() == "" {
			return url2, fmt.Errorf("hostname empty after parsing")
		}

	}
	return url2, err
}

// handleXMLResp takes a http response and validates it against an expected status code
// if response code is unexpected or it fails to decode into the interface provided
// by the caller, an error of type ApiErr is returned
func handleXMLResp(resp *http.Response, expectCode int, decodeInto interface{}) error {
	if resp.StatusCode != expectCode {
		return handleXMLErrResp(resp)
	}

	if decodeInto == nil {
		return nil
	}

	if err := xml.NewDecoder(resp.Body).Decode(decodeInto); err != nil {
		return createApiErr(resp.StatusCode, createDecodingErrorMessage(err))

	}
	return nil
}

// handleJsonResp takes a http response and validates it against an expected status code
// if response code is unexpected or it fails to decode into the interface provided
// by the caller, an error of type ApiErr is returned
func handleJsonResp(resp *http.Response, expectCode int, decodeInto interface{}) error {
	if resp.StatusCode != expectCode {
		return handleJsonErrResp(resp)
	}

	if decodeInto == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(decodeInto); err != nil {
		return createApiErr(resp.StatusCode, createDecodingErrorMessage(err))
	}

	return nil
}

// handleXMLErrResp decodes an XML response from 3scale system
// into an error of type ApiErr
func handleXMLErrResp(resp *http.Response) error {
	var errResp ErrorResp

	if err := xml.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return createApiErr(resp.StatusCode, createDecodingErrorMessage(err))
	}

	return ApiErr{resp.StatusCode, errResp.Text}
}

// handleJsonErrResp decodes a JSON response from 3scale system
// into an error of type APiErr
func handleJsonErrResp(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnprocessableEntity:
		return parseUnprocessableEntityError(resp)
	default:
		return parseUnexpectedError(resp)
	}
}

func parseUnexpectedError(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return createApiErr(resp.StatusCode, string(body))
}

func parseUnprocessableEntityError(resp *http.Response) error {
	errObj := struct {
		Errors map[string][]string `json:"errors"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&errObj); err != nil {
		return createApiErr(resp.StatusCode, createDecodingErrorMessage(err))
	}

	msg, err := json.Marshal(errObj.Errors)
	if err != nil {
		return createApiErr(resp.StatusCode, createDecodingErrorMessage(err))
	}

	return createApiErr(resp.StatusCode, string(msg))
}

func createApiErr(statusCode int, message string) ApiErr {
	return ApiErr{
		code: statusCode,
		err:  message,
	}
}

func createDecodingErrorMessage(err error) string {
	return fmt.Sprintf("decoding error - %s", err.Error())
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
