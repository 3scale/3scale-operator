package client

import (
	"net/http"
	"net/url"
)

const (
	accountList = "/admin/api/accounts.json"
)

func (c *ThreeScaleClient) ListAccounts() (*AccountList, error) {
	req, err := c.buildGetReq(accountList)
	if err != nil {
		return nil, err
	}

	urlValues := url.Values{}
	req.URL.RawQuery = urlValues.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	accountList := &AccountList{}
	err = handleJsonResp(resp, http.StatusOK, accountList)
	return accountList, err
}
