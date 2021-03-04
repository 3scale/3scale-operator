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
)

func TestProductEntityBasics(t *testing.T) {
	token := "12345"
	product := &threescaleapi.Product{
		Element: threescaleapi.ProductItem{
			ID:               int64(4567),
			Name:             "some product",
			SystemName:       "my_product",
			State:            "active",
			Description:      "some descr",
			DeploymentOption: "hosted",
			BackendVersion:   "1",
		},
	}

	client := threescaleapi.NewThreeScale(nil, token, nil)
	productEntity := NewProductEntity(product, client, logrtesting.NullLogger{})
	equals(t, product.Element.ID, productEntity.ID())
	equals(t, product.Element.Name, productEntity.Name())
	equals(t, product.Element.State, productEntity.State())
	equals(t, product.Element.Description, productEntity.Description())
	equals(t, product.Element.DeploymentOption, productEntity.DeploymentOption())
	// TODO FIX
	equals(t, product.Element.DeploymentOption, productEntity.BackendVersion())
}

func TestProductEntityUpdate(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.Product{
			Element: threescaleapi.ProductItem{
				ID: int64(4567), Name: "some product", SystemName: "my_product", State: "active",
				Description: "some descr", DeploymentOption: "hosted", BackendVersion: "1",
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.Update(threescaleapi.Params{})
	ok(t, err)
	equals(t, int64(4567), productEntity.ID())
	equals(t, "some product", productEntity.Name())
	equals(t, "some descr", productEntity.Description())
}

func TestProductEntityMethods(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	methodList, err := productEntity.Methods()
	ok(t, err)
	assert(t, methodList != nil, "method list returned nil")
	equals(t, 1, len(methodList.Methods))
	equals(t, "method_01", methodList.Methods[0].Element.SystemName)
}

func TestProductEntityCreateMethod(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.CreateMethod(threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityDeleteMethod(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.DeleteMethod(int64(3))
	ok(t, err)
}

func TestProductEntityUpdateMethod(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdateMethod(int64(3), threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityCreateMetric(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.CreateMetric(threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityDeleteMetric(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.DeleteMetric(int64(5))
	ok(t, err)
}

func TestProductEntityUpdateMetric(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdateMetric(int64(3), threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityMetricsAndMethods(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	metricList, err := productEntity.MetricsAndMethods()
	ok(t, err)
	assert(t, metricList != nil, "metric list returned nil")
	equals(t, 3, len(metricList.Metrics))
	assert(t, FindMetric(metricList, "hits"), "hits metric not found")
	assert(t, FindMetric(metricList, "metric_01"), "metric_01 metric not found")
	assert(t, FindMetric(metricList, "method_01"), "method_01 metric not found")
}

func TestProductEntityMetrics(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	metricList, err := productEntity.Metrics()
	ok(t, err)
	assert(t, metricList != nil, "metric list returned nil")
	equals(t, 2, len(metricList.Metrics))
	assert(t, FindMetric(metricList, "hits"), "hits metric not found")
	assert(t, FindMetric(metricList, "metric_01"), "metric_01 metric not found")
}

func TestProductEntityMappingRules(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	ruleList, err := productEntity.MappingRules()
	ok(t, err)
	assert(t, ruleList != nil, "mapping rule list returned nil")
	equals(t, 2, len(ruleList.MappingRules))
}

func TestProductEntityDeleteMappingRule(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.DeleteMappingRule(int64(3))
	ok(t, err)
}

func TestProductEntityCreateMappingRule(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.CreateMappingRule(threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityUpdateMappingRule(t *testing.T) {
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdateMappingRule(int64(1), threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityFindMethodMetricIDBySystemName(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(GetMethodsMetricsRoundTripFunc)
	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})

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
			metricID, err := productEntity.FindMethodMetricIDBySystemName(tc.systemName)
			ok(subT, err)
			equals(subT, tc.expectedID, metricID)
		})
	}
}

func TestProductEntityBackendUsages(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.BackendAPIUsageList{
			{
				Element: threescaleapi.BackendAPIUsageItem{
					Path: "/v1", ProductID: int64(1), BackendAPIID: int64(1),
				},
			},
			{
				Element: threescaleapi.BackendAPIUsageItem{
					Path: "/v2", ProductID: int64(1), BackendAPIID: int64(2),
				},
			},
			{
				Element: threescaleapi.BackendAPIUsageItem{
					Path: "/v3", ProductID: int64(1), BackendAPIID: int64(3),
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	backendUsageList, err := productEntity.BackendUsages()
	ok(t, err)
	equals(t, 3, len(backendUsageList))
	assert(t, FindBackendUsage(backendUsageList, "/v1"), "/v1 backend usage not found")
	assert(t, FindBackendUsage(backendUsageList, "/v2"), "/v2 backend usage not found")
	assert(t, FindBackendUsage(backendUsageList, "/v3"), "/v3 backend usage not found")
}

func TestProductEntityDeleteBackendUsage(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.DeleteBackendUsage(int64(3))
	ok(t, err)
}

func TestProductEntityUpdateBackendUsage(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.BackendAPIUsage{
			Element: threescaleapi.BackendAPIUsageItem{
				Path: "/v1", ProductID: int64(1), BackendAPIID: int64(1),
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdateBackendUsage(int64(1), threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityCreateBackendUsage(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.BackendAPIUsage{
			Element: threescaleapi.BackendAPIUsageItem{
				Path: "/v1", ProductID: int64(1), BackendAPIID: int64(1),
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.CreateBackendUsage(threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityProxy(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.ProxyJSON{
			Element: threescaleapi.ProxyItem{Endpoint: "https://gw.example.com"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	proxy, err := productEntity.Proxy()
	ok(t, err)
	assert(t, proxy != nil, "proxy returned nil")
	equals(t, "https://gw.example.com", proxy.Element.Endpoint)
}

func TestProductEntityUpdateProxy(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.ProxyJSON{
			Element: threescaleapi.ProxyItem{Endpoint: "https://gw.example.com"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdateProxy(threescaleapi.Params{})
	ok(t, err)
}

func TestProductEntityApplicationPlans(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.ApplicationPlanJSONList{
			Plans: []threescaleapi.ApplicationPlan{
				{
					Element: threescaleapi.ApplicationPlanItem{SystemName: "plan01"},
				},
				{
					Element: threescaleapi.ApplicationPlanItem{SystemName: "plan02"},
				},
				{
					Element: threescaleapi.ApplicationPlanItem{SystemName: "plan03"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	plans, err := productEntity.ApplicationPlans()
	ok(t, err)
	assert(t, plans != nil, "plans returned nil")
	equals(t, 3, len(plans.Plans))
	assert(t, FindApplicationPlan(plans.Plans, "plan01"), "plan01 plan not found")
	assert(t, FindApplicationPlan(plans.Plans, "plan02"), "plan02 plan not found")
	assert(t, FindApplicationPlan(plans.Plans, "plan03"), "plan03 plan not found")
}

func TestProductEntityDeleteApplicationPlan(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.DeleteApplicationPlan(int64(1))
	ok(t, err)
}

func TestProductEntityCreateApplicationPlan(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.ApplicationPlan{
			Element: threescaleapi.ApplicationPlanItem{SystemName: "plan01"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	plan, err := productEntity.CreateApplicationPlan(threescaleapi.Params{})
	ok(t, err)
	assert(t, plan != nil, "plan returned nil")
	equals(t, "plan01", plan.Element.SystemName)
}

func TestProductEntityPromoteProxyToStaging(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.ProxyJSON{
			Element: threescaleapi.ProxyItem{Endpoint: "https://gw.example.com"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.PromoteProxyToStaging()
	ok(t, err)
}

func TestProductEntityPolicies(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.PoliciesConfigList{
			Policies: []threescaleapi.PolicyConfig{
				{Name: "policy01"},
				{Name: "policy02"},
				{Name: "policy03"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	policies, err := productEntity.Policies()
	ok(t, err)
	assert(t, policies != nil, "policies returned nil")
	equals(t, 3, len(policies.Policies))
	assert(t, FindPolicy(policies.Policies, "policy01"), "policy01 policy not found")
	assert(t, FindPolicy(policies.Policies, "policy02"), "policy02 policy not found")
	assert(t, FindPolicy(policies.Policies, "policy03"), "policy03 policy not found")
}

func TestProductEntityUpdatePolicies(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.PoliciesConfigList{
			Policies: []threescaleapi.PolicyConfig{
				{Name: "policy01"},
				{Name: "policy02"},
				{Name: "policy03"},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdatePolicies(&threescaleapi.PoliciesConfigList{})
	ok(t, err)
}

func TestProductEntityOIDCConfiguration(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.OIDCConfiguration{
			Element: threescaleapi.OIDCConfigurationItem{ID: int64(1)},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	oidcConf, err := productEntity.OIDCConfiguration()
	ok(t, err)
	assert(t, oidcConf != nil, "oidcConf returned nil")
	equals(t, int64(1), oidcConf.Element.ID)
}

func TestProductEntityUpdateOIDCConfiguration(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.OIDCConfiguration{
			Element: threescaleapi.OIDCConfigurationItem{ID: int64(1)},
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

	productEntity := NewProductEntity(&threescaleapi.Product{}, client, logrtesting.NullLogger{})
	err := productEntity.UpdateOIDCConfiguration(&threescaleapi.OIDCConfiguration{})
	ok(t, err)
}
