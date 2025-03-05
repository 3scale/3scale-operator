package helper

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
)

var (
	// InvalidDNS1123Regexp not alphanumeric
	InvalidDNS1123Regexp = regexp.MustCompile(`[^0-9A-Za-z-]`)
)

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

// CmpResources returns true if the resource requirements a is equal to b,
func CmpResources(a, b *v1.ResourceRequirements) bool {
	return CmpResourceList(&a.Limits, &b.Limits) && CmpResourceList(&a.Requests, &b.Requests)
}

// CmpResourceList returns true if the resourceList a is equal to b,
func CmpResourceList(a, b *v1.ResourceList) bool {
	return a.Cpu().Cmp(*b.Cpu()) == 0 &&
		a.Memory().Cmp(*b.Memory()) == 0 &&
		b.Pods().Cmp(*b.Pods()) == 0 &&
		b.StorageEphemeral().Cmp(*b.StorageEphemeral()) == 0
}

func GetEnvVar(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func GetStringPointerValueOrDefault(val *string, def string) string {
	if val != nil {
		return *val
	}
	return def
}

func DNS1123Name(in string) string {
	tmp := strings.ToLower(in)
	return InvalidDNS1123Regexp.ReplaceAllString(tmp, "")
}

type TLSConfig struct {
	Enabled       bool
	CACertificate string
	Certificate   string
	Key           string
}

// HasCA returns whether the configuration has a certificate authority or not.
func (c *TLSConfig) HasCA() bool {
	return c.CACertificate != ""
}

// HasCertAuth returns whether the configuration has certificate authentication or not.
func (c *TLSConfig) HasCertAuth() bool {
	return (c.Certificate != "" && c.Key != "")
}

func LoadCerts(cfg *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{ //nolint:gosec
		MinVersion: tls.VersionTLS12,
	}

	if cfg.HasCA() {
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM([]byte(cfg.CACertificate))
		if !ok {
			return nil, fmt.Errorf("unable to load root certificate")
		}
		tlsConfig.RootCAs = certPool
	}

	// If key/cert were provided
	if cfg.HasCertAuth() {
		cert, err := tls.X509KeyPair([]byte(cfg.Certificate), []byte(cfg.Key))
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
