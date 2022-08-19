package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func NewTestAdminPortal(t *testing.T) *threescaleapi.AdminPortal {
	t.Helper()
	ap, err := threescaleapi.NewAdminPortalFromStr("https://www.test.com:443")
	ok(t, err)
	return ap
}

func GetTestSecret(namespace, secretName string, data map[string]string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: data,
		Type:       v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func GetMethodsMetricsRoundTripFunc(req *http.Request) *http.Response {
	metricList := &threescaleapi.MetricJSONList{
		Metrics: []threescaleapi.MetricJSON{
			{
				Element: threescaleapi.MetricItem{
					ID:         int64(1),
					Name:       "Hits",
					SystemName: "hits",
					Unit:       "hit",
				},
			},
			{
				Element: threescaleapi.MetricItem{
					ID:         int64(2),
					Name:       "Metric 01",
					SystemName: "metric_01",
					Unit:       "1",
				},
			},
			{
				Element: threescaleapi.MetricItem{
					ID:         int64(3),
					Name:       "Method 01",
					SystemName: "method_01",
					Unit:       "hit",
				},
			},
		},
	}

	methodList := &threescaleapi.MethodList{
		Methods: []threescaleapi.Method{
			{
				Element: threescaleapi.MethodItem{
					ID:         int64(3),
					Name:       "Method 01",
					ParentID:   int64(1),
					SystemName: "method_01",
				},
			},
		},
	}

	var respObject interface{}

	if req.Method == "GET" && regexp.MustCompile("metrics.json").FindString(req.URL.Path) != "" {
		respObject = metricList
	}

	if req.Method == "GET" && regexp.MustCompile("methods.json").FindString(req.URL.Path) != "" {
		respObject = methodList
	}

	responseBodyBytes, _ := json.Marshal(respObject)

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
		Header:     make(http.Header),
	}
}

func FindMetric(l *threescaleapi.MetricJSONList, systemName string) bool {
	for _, n := range l.Metrics {
		if systemName == n.Element.SystemName {
			return true
		}
	}
	return false
}

func FindBackendUsage(l threescaleapi.BackendAPIUsageList, path string) bool {
	for _, n := range l {
		if path == n.Element.Path {
			return true
		}
	}
	return false
}

func FindApplicationPlan(plans []threescaleapi.ApplicationPlan, systemName string) bool {
	for _, n := range plans {
		if systemName == n.Element.SystemName {
			return true
		}
	}
	return false
}

func FindPolicy(policies []threescaleapi.PolicyConfig, name string) bool {
	for _, n := range policies {
		if name == n.Name {
			return true
		}
	}
	return false
}
