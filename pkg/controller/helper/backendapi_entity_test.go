package helper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	logrtesting "github.com/go-logr/logr/testing"
	"github.com/google/go-cmp/cmp"
)

func TestSanitizeBackendSystemName(t *testing.T) {
	cases := []struct {
		name               string
		systemName         string
		expectedSystemName string
	}{
		{"test01", "hits.45498", "hits"},
		{"test02", "hits.something.45498", "hits.something"},
		{"test03", "hits", "hits"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			newName := SanitizeBackendSystemName(tc.systemName)
			if newName != tc.expectedSystemName {
				diff := cmp.Diff(newName, tc.expectedSystemName)
				subT.Errorf("diff %s", diff)
			}
		})
	}
}

func TestBackendAPIEntityBasics(t *testing.T) {
	token := "12345"
	backendItem := &threescaleapi.BackendApi{
		Element: threescaleapi.BackendApiItem{
			ID:              int64(4567),
			Name:            "some backend",
			SystemName:      "my_backend",
			Description:     "some descr",
			PrivateEndpoint: "https://example.com",
			AccountID:       int64(12),
		},
	}

	client := threescaleapi.NewThreeScale(nil, token, nil)

	backendEntity := NewBackendAPIEntity(backendItem, client, logrtesting.NullLogger{})
	equals(t, backendEntity.ID(), backendItem.Element.ID)
	equals(t, backendEntity.Name(), backendItem.Element.Name)
	equals(t, backendEntity.SystemName(), backendItem.Element.SystemName)
	equals(t, backendEntity.Description(), backendItem.Element.Description)
	equals(t, backendEntity.PrivateEndpoint(), backendItem.Element.PrivateEndpoint)
}

func TestBackendAPIEntityUpdate(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.BackendApi{
			Element: threescaleapi.BackendApiItem{
				ID:              int64(4567),
				Name:            "some backend",
				SystemName:      "my_backend",
				Description:     "some descr",
				PrivateEndpoint: "https://example.com",
				AccountID:       int64(12),
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.Update(threescaleapi.Params{})
	ok(t, err)
	equals(t, int64(4567), backendEntity.ID())
	equals(t, "some backend", backendEntity.Name())
	equals(t, "my_backend", backendEntity.SystemName())
	equals(t, "some descr", backendEntity.Description())
	equals(t, "https://example.com", backendEntity.PrivateEndpoint())
}

func TestBackendAPIEntityUpdateError(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := struct {
			Errors map[string][]string `json:"errors"`
		}{
			map[string][]string{"some_attr": []string{"not valid"}},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.Update(threescaleapi.Params{})
	assert(t, err != nil, "update did not return error")
}

func TestBackendAPIEntityMethods(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	methodList, err := backendEntity.Methods()
	ok(t, err)
	assert(t, methodList != nil, "method list returned nil")
	equals(t, 1, len(methodList.Methods))
	equals(t, "method_01", methodList.Methods[0].Element.SystemName)
}

func TestBackendAPIEntityMethodsError(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		errObj := struct {
			Errors map[string][]string `json:"errors"`
		}{
			map[string][]string{"some_attr": []string{"not valid"}},
		}

		responseBodyBytes, err := json.Marshal(errObj)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)
	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	_, err := backendEntity.Methods()
	assert(t, err != nil, "Methods did not return error")
}

func TestBackendAPIEntityCreateMethod(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		if req.Method == "GET" && regexp.MustCompile("metrics.json").FindString(req.URL.Path) != "" {
			return GetMethodsMetricsRoundTripFunc(req)
		}

		respObject := &threescaleapi.Method{
			Element: threescaleapi.MethodItem{
				ID:         int64(4),
				Name:       "Method 02",
				ParentID:   int64(1),
				SystemName: "method_02",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.CreateMethod(threescaleapi.Params{})
	ok(t, err)
}

func TestBackendAPIEntityDeleteMethod(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		if req.Method == "GET" && regexp.MustCompile("metrics.json").FindString(req.URL.Path) != "" {
			return GetMethodsMetricsRoundTripFunc(req)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.DeleteMethod(int64(3))
	ok(t, err)
}

func TestBackendAPIEntityUpdateMethod(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		if req.Method == "GET" && regexp.MustCompile("metrics.json").FindString(req.URL.Path) != "" {
			return GetMethodsMetricsRoundTripFunc(req)
		}

		respObject := &threescaleapi.Method{
			Element: threescaleapi.MethodItem{
				ID:         int64(2),
				Name:       "Method 02",
				ParentID:   int64(1),
				SystemName: "method_02",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.UpdateMethod(int64(3), threescaleapi.Params{})
	ok(t, err)
}

func TestBackendAPIEntityCreateMetric(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.MetricJSON{
			Element: threescaleapi.MetricItem{
				ID:         int64(5),
				Name:       "Metric 02",
				SystemName: "metric_02",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.CreateMetric(threescaleapi.Params{})
	ok(t, err)
}

func TestBackendAPIEntityDeleteMetric(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.DeleteMetric(int64(5))
	ok(t, err)
}

func TestBackendAPIEntityUpdateMetric(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.MetricJSON{
			Element: threescaleapi.MetricItem{
				ID:         int64(5),
				Name:       "Metric 02",
				SystemName: "metric_02",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.UpdateMetric(int64(3), threescaleapi.Params{})
	ok(t, err)
}

func TestBackendAPIEntityMetricsAndMethods(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	metricList, err := backendEntity.MetricsAndMethods()
	ok(t, err)
	assert(t, metricList != nil, "metric list returned nil")
	equals(t, 3, len(metricList.Metrics))
	assert(t, FindMetric(metricList, "hits"), "hits metric not found")
	assert(t, FindMetric(metricList, "metric_01"), "metric_01 metric not found")
	assert(t, FindMetric(metricList, "method_01"), "method_01 metric not found")
}

func TestBackendAPIEntityMetrics(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	metricList, err := backendEntity.Metrics()
	ok(t, err)
	assert(t, metricList != nil, "metric list returned nil")
	equals(t, 2, len(metricList.Metrics))
	assert(t, FindMetric(metricList, "hits"), "hits metric not found")
	assert(t, FindMetric(metricList, "metric_01"), "metric_01 metric not found")
}

func TestBackendAPIEntityMappingRules(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.MappingRuleJSONList{
			MappingRules: []threescaleapi.MappingRuleJSON{
				{
					Element: threescaleapi.MappingRuleItem{
						MetricID:   int64(1),
						Pattern:    "/pets",
						HTTPMethod: "GET",
						Delta:      1,
						Position:   1,
					},
				},
				{
					Element: threescaleapi.MappingRuleItem{
						MetricID:   int64(1),
						Pattern:    "/cats",
						HTTPMethod: "GET",
						Delta:      1,
						Position:   1,
					},
				},
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	ruleList, err := backendEntity.MappingRules()
	ok(t, err)
	assert(t, ruleList != nil, "mapping rule list returned nil")
	equals(t, 2, len(ruleList.MappingRules))
}

func TestBackendAPIEntityDeleteMappingRule(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.DeleteMappingRule(int64(3))
	ok(t, err)
}

func TestBackendAPIEntityCreateMappingRule(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.MappingRuleJSON{
			Element: threescaleapi.MappingRuleItem{
				MetricID:   int64(1),
				Pattern:    "/pets",
				HTTPMethod: "GET",
				Delta:      1,
				Position:   1,
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.CreateMappingRule(threescaleapi.Params{})
	ok(t, err)
}

func TestBackendAPIEntityUpdateMappingRule(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.MappingRuleJSON{
			Element: threescaleapi.MappingRuleItem{
				MetricID:   int64(1),
				Pattern:    "/pets",
				HTTPMethod: "GET",
				Delta:      1,
				Position:   1,
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})
	err := backendEntity.UpdateMappingRule(int64(1), threescaleapi.Params{})
	ok(t, err)
}

func TestBackendAPIEntityFindMethodMetricIDBySystemName(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	backendEntity := NewBackendAPIEntity(&threescaleapi.BackendApi{}, client, logrtesting.NullLogger{})

	cases := []struct {
		name       string
		systemName string
		expectedID int64
	}{
		{"hits exists", "hits", 1},
		{"metric_01 exists", "metric_01", 2},
		{"method_01 exists", "method_01", 3},
		{"method_02 does not exist", "method_02", -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			metricID, err := backendEntity.FindMethodMetricIDBySystemName(tc.systemName)
			ok(subT, err)
			equals(subT, tc.expectedID, metricID)
		})
	}
}
