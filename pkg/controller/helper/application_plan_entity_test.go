package helper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	logrtesting "github.com/go-logr/logr/testing"
)

func TestApplicationPlanEntityBasics(t *testing.T) {
	var productID int64 = 1293
	token := "12345"
	planItem := threescaleapi.ApplicationPlanItem{
		ID:               4567,
		Name:             "some plan",
		ApprovalRequired: true,
		TrialPeriodDays:  3,
		SetupFee:         5.67,
		CostPerMonth:     8.67,
	}

	client := threescaleapi.NewThreeScale(nil, token, nil)

	appPlanEntity := NewApplicationPlanEntity(productID, planItem, client, logrtesting.NullLogger{})
	equals(t, appPlanEntity.ID(), planItem.ID)
	equals(t, appPlanEntity.Name(), planItem.Name)
	equals(t, appPlanEntity.ApprovalRequired(), planItem.ApprovalRequired)
	equals(t, appPlanEntity.TrialPeriodDays(), planItem.TrialPeriodDays)
	equals(t, appPlanEntity.SetupFee(), planItem.SetupFee)
	equals(t, appPlanEntity.CostPerMonth(), planItem.CostPerMonth)
}

func TestApplicationPlanEntityUpdate(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationPlan{
			Element: threescaleapi.ApplicationPlanItem{
				ID:               4567,
				Name:             "some plan",
				ApprovalRequired: true,
				TrialPeriodDays:  3,
				SetupFee:         5.67,
				CostPerMonth:     8.67,
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.Update(threescaleapi.Params{})
	ok(t, err)
	equals(t, appPlanEntity.ID(), int64(4567))
	equals(t, appPlanEntity.Name(), "some plan")
	equals(t, appPlanEntity.ApprovalRequired(), true)
	equals(t, appPlanEntity.TrialPeriodDays(), 3)
	equals(t, appPlanEntity.SetupFee(), 5.67)
	equals(t, appPlanEntity.CostPerMonth(), 8.67)
}

func TestApplicationPlanEntityUpdateError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.Update(threescaleapi.Params{})
	assert(t, err != nil, "update did not return error")
}

func TestApplicationPlanEntityLimits(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationPlanLimitList{
			Limits: []threescaleapi.ApplicationPlanLimit{
				{
					Element: threescaleapi.ApplicationPlanLimitItem{
						ID:       int64(1),
						Period:   "eternity",
						MetricID: int64(10),
						Value:    1000,
					},
				},
				{
					Element: threescaleapi.ApplicationPlanLimitItem{
						ID:       int64(2),
						Period:   "year",
						MetricID: int64(10),
						Value:    1000,
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	limits, err := appPlanEntity.Limits()
	ok(t, err)
	assert(t, limits != nil, "Limits returned nil")
	equals(t, len(limits.Limits), 2)
}

func TestApplicationPlanEntityLimitsError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	_, err := appPlanEntity.Limits()
	assert(t, err != nil, "Limits did not return error")
}

func TestApplicationPlanEntityDeleteLimit(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.DeleteLimit(int64(1234), int64(10))
	ok(t, err)
}

func TestApplicationPlanEntityDeleteLimitError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.DeleteLimit(int64(1234), int64(10))
	assert(t, err != nil, "DeleteLimit did not return error")
}

func TestApplicationPlanEntityCreateLimit(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationPlanLimit{
			Element: threescaleapi.ApplicationPlanLimitItem{
				ID:       int64(1),
				Period:   "eternity",
				MetricID: int64(10),
				Value:    1000,
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.CreateLimit(int64(1234), threescaleapi.Params{})
	ok(t, err)
}

func TestApplicationPlanEntityCreateLimitError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.CreateLimit(int64(1234), threescaleapi.Params{})
	assert(t, err != nil, "CreateLimit did not return error")
}

func TestApplicationPlanEntityPricingRules(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationPlanPricingRuleList{
			Rules: []threescaleapi.ApplicationPlanPricingRule{
				{
					Element: threescaleapi.ApplicationPlanPricingRuleItem{
						ID:          int64(1),
						MetricID:    int64(10),
						CostPerUnit: "123",
						Min:         1,
						Max:         10,
					},
				},
				{
					Element: threescaleapi.ApplicationPlanPricingRuleItem{
						ID:          int64(1),
						MetricID:    int64(10),
						CostPerUnit: "123",
						Min:         11,
						Max:         20,
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	rules, err := appPlanEntity.PricingRules()
	ok(t, err)
	assert(t, rules != nil, "PricingRules returned nil")
	equals(t, len(rules.Rules), 2)
}

func TestApplicationPlanEntityPricingRulesError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	_, err := appPlanEntity.PricingRules()
	assert(t, err != nil, "PricingRules did not return error")
}

func TestApplicationPlanEntityDeletePricingRule(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.DeletePricingRule(int64(1234), int64(10))
	ok(t, err)
}

func TestApplicationPlanEntityDeletePricingRuleError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.DeletePricingRule(int64(1234), int64(10))
	assert(t, err != nil, "DeletePricingRule did not return error")
}

func TestApplicationPlanEntityCreatePricingRule(t *testing.T) {
	var productID int64 = 1293
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationPlanPricingRule{
			Element: threescaleapi.ApplicationPlanPricingRuleItem{
				ID:          int64(1),
				MetricID:    int64(10),
				CostPerUnit: "123",
				Min:         11,
				Max:         20,
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.CreatePricingRule(int64(1234), threescaleapi.Params{})
	ok(t, err)
}

func TestApplicationPlanEntityCreatePricingRuleError(t *testing.T) {
	var productID int64 = 1293
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

	appPlanEntity := NewApplicationPlanEntity(productID, threescaleapi.ApplicationPlanItem{}, client, logrtesting.NullLogger{})
	err := appPlanEntity.CreatePricingRule(int64(1234), threescaleapi.Params{})
	assert(t, err != nil, "CreatePricingRule did not return error")
}
