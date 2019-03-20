package helper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/3scale/3scale-porta-go-client/client"
)

// PortaClientFromURLString instantiate porta_client.ThreeScaleClient from admin url string
func PortaClientFromURLString(adminURLStr, masterAccessToken string) (*client.ThreeScaleClient, error) {
	adminURL, err := url.Parse(adminURLStr)
	if err != nil {
		return nil, err
	}
	return PortaClient(adminURL, masterAccessToken)
}

// PortaClient instantiates porta_client.ThreeScaleClient from admin url object
func PortaClient(url *url.URL, masterAccessToken string) (*client.ThreeScaleClient, error) {
	adminPortal, err := client.NewAdminPortal(url.Scheme, url.Hostname(), PortFromURL(url))
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return client.NewThreeScale(adminPortal, masterAccessToken, &http.Client{Transport: tr}), nil
}

// PortFromURL infers port number if it is not explict
func PortFromURL(url *url.URL) int {
	if url.Port() != "" {
		if portNum, err := strconv.Atoi(url.Port()); err == nil {
			return portNum
		}
	}

	// Default HTTP port numbers
	portNum := 80
	// Scheme is always lowercase
	if url.Scheme == "https" {
		portNum = 443
	}
	return portNum
}

// SetURLDefaultPort adds the default Port if not set
func SetURLDefaultPort(rawurl string) string {

	urlObj, _ := url.Parse(rawurl)

	if urlObj.Port() != "" {
		return urlObj.String()
	}

	portNum := PortFromURL(urlObj)
	return fmt.Sprintf("%s:%d", urlObj.String(), portNum)
}
