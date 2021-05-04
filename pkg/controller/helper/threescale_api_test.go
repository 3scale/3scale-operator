package helper

import (
	"net/url"
	"testing"
)

func TestPortaClientInvalidURL(t *testing.T) {
	providerAccount := &ProviderAccount{AdminURLStr: ":foo", Token: "some token"}
	_, err := PortaClient(providerAccount)
	assert(t, err != nil, "error should not be nil")
}

func TestPortaClient(t *testing.T) {
	providerAccount := &ProviderAccount{AdminURLStr: "http://somedomain.example.com", Token: "some token"}
	_, err := PortaClient(providerAccount)
	ok(t, err)
}

func TestPortaClientFromURLStringInvalidURL(t *testing.T) {
	_, err := PortaClientFromURLString(":foo", "some token")
	assert(t, err != nil, "error should not be nil")
}

func TestPortaClientFromURLString(t *testing.T) {
	_, err := PortaClientFromURLString("http://somedomain.example.com", "some token")
	ok(t, err)
}

func TestPortaClientFromURL(t *testing.T) {
	url := &url.URL{}
	_, err := PortaClientFromURL(url, "some token")
	assert(t, err != nil, "error should not be nil")
}
