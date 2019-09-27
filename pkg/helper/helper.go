package helper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

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

func WrapRawExtensions(objects []common.KubernetesObject) []runtime.RawExtension {
	var rawExtensions []runtime.RawExtension
	for index := range objects {
		rawExtensions = append(rawExtensions, WrapRawExtension(objects[index]))
	}
	return rawExtensions
}

func WrapRawExtension(object runtime.Object) runtime.RawExtension {
	return runtime.RawExtension{Object: object}
}

func UnwrapRawExtensions(rawExts []runtime.RawExtension) []common.KubernetesObject {
	var objects []common.KubernetesObject
	for index := range rawExts {
		rawObject := rawExts[index].Object
		obj, ok := rawObject.(common.KubernetesObject)
		if ok {
			objects = append(objects, obj)
		} else {
			panic(fmt.Sprintf("Expected RawExtension to wrap a KubernetesObject, but instead found %v", rawObject))
		}
	}
	return objects
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
