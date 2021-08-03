package helper

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/3scale/3scale-operator/pkg/common"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// NOTE: remove when templates are gone
func WrapRawExtensions(objects []common.KubernetesObject) []runtime.RawExtension {
	var rawExtensions []runtime.RawExtension
	for index := range objects {
		rawExtensions = append(rawExtensions, WrapRawExtension(objects[index]))
	}
	return rawExtensions
}

// NOTE: remove when templates are gone
func WrapRawExtension(object runtime.Object) runtime.RawExtension {
	return runtime.RawExtension{Object: object}
}

// NOTE: remove when templates are gone
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
