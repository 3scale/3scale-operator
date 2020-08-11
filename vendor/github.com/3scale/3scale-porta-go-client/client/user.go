package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	userActivate = "/admin/api/accounts/%d/users/%d/activate.json"
	userRead     = "/admin/api/accounts/%d/users/%d.json"
	userList     = "/admin/api/accounts/%d/users.json"
	userUpdate   = "/admin/api/accounts/%d/users/%d.json"
)

// ActivateUser activates user of a given account from pending state to active
func (c *ThreeScaleClient) ActivateUser(accountID, userID int64) error {
	endpoint := fmt.Sprintf(userActivate, accountID, userID)

	req, err := c.buildUpdateReq(endpoint, nil)
	if err != nil {
		return err
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

// ReadUser reads user of a given account
func (c *ThreeScaleClient) ReadUser(accountID, userID int64) (*User, error) {
	endpoint := fmt.Sprintf(userRead, accountID, userID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	userElem := &UserElem{}
	err = handleJsonResp(resp, http.StatusOK, userElem)
	if err != nil {
		return nil, err
	}
	return &userElem.User, nil
}

// ListUser list users of a given account and a given filter params
func (c *ThreeScaleClient) ListUsers(accountID int64, filterParams Params) (*UserList, error) {
	endpoint := fmt.Sprintf(userList, accountID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return nil, httpReqError
	}

	values := url.Values{}
	for k, v := range filterParams {
		values.Add(k, v)
	}
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	userList := &UserList{}
	err = handleJsonResp(resp, http.StatusOK, userList)
	if err != nil {
		return nil, err
	}
	return userList, nil
}

// UpdateUser updates user of a given account
func (c *ThreeScaleClient) UpdateUser(accountID int64, userID int64, userParams Params) (*User, error) {
	endpoint := fmt.Sprintf(userUpdate, accountID, userID)

	values := url.Values{}
	for k, v := range userParams {
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

	userElem := &UserElem{}
	err = handleJsonResp(resp, http.StatusOK, userElem)
	if err != nil {
		return nil, err
	}
	return &userElem.User, nil
}
