package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

func TestApplicationBasic(t *testing.T) {
	token := "12345"
	application := &threescaleapi.Application{
		ID:                      3,
		CreatedAt:               "",
		UpdatedAt:               "",
		State:                   "live",
		UserAccountID:           3,
		FirstTrafficAt:          "",
		FirstDailyTrafficAt:     "",
		EndUserRequired:         false,
		ServiceID:               2,
		UserKey:                 "",
		ProviderVerificationKey: "",
		PlanID:                  8,
		AppName:                 "test",
		Description:             "test",
		ExtraFields:             "",
		Error:                   "",
	}

	client := threescaleapi.NewThreeScale(nil, token, nil)

	appEntity := NewApplicationEntity(application, client, logr.Discard())
	require.Equal(t, appEntity.ID(), application.ID)
	require.Equal(t, appEntity.AppName(), application.AppName)
	require.Equal(t, appEntity.Description(), application.Description)
	require.Equal(t, appEntity.UserAccountID(), application.UserAccountID)
	require.Equal(t, appEntity.PlanID(), application.PlanID)
	require.Equal(t, appEntity.ApplicationState(), application.State)
}

func TestChangeApplicationPlan(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationElem{
			Application: threescaleapi.Application{
				ID:              3,
				State:           "live",
				UserAccountID:   3,
				EndUserRequired: false,
				ServiceID:       2,
				PlanID:          2,
				AppName:         "test",
				Description:     "test",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)
	appEntity := NewApplicationEntity(&threescaleapi.Application{}, client, logr.Discard())

	err := appEntity.ChangeApplicationPlan(2)
	require.NoError(t, err)
	require.Equal(t, int64(2), appEntity.PlanID())
}

func TestSuspend(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationElem{
			Application: threescaleapi.Application{
				ID:              3,
				State:           "suspended",
				UserAccountID:   3,
				EndUserRequired: false,
				ServiceID:       2,
				PlanID:          2,
				AppName:         "test",
				Description:     "test",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)
	appEntity := NewApplicationEntity(&threescaleapi.Application{}, client, logr.Discard())

	err := appEntity.Suspend()
	require.NoError(t, err)
	require.Equal(t, "suspended", appEntity.ApplicationState())
}

func TestResume(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationElem{
			Application: threescaleapi.Application{
				ID:              3,
				State:           "live",
				UserAccountID:   3,
				EndUserRequired: false,
				ServiceID:       2,
				PlanID:          2,
				AppName:         "test",
				Description:     "test",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)
	appEntity := NewApplicationEntity(&threescaleapi.Application{}, client, logr.Discard())

	err := appEntity.Resume()
	require.NoError(t, err)
	require.Equal(t, "live", appEntity.ApplicationState())
}

func TestUpdate(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := threescaleapi.ApplicationElem{
			Application: threescaleapi.Application{
				ID:              3,
				State:           "live",
				UserAccountID:   3,
				EndUserRequired: false,
				ServiceID:       2,
				PlanID:          2,
				AppName:         "test",
				Description:     "test",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)
	appEntity := NewApplicationEntity(&threescaleapi.Application{}, client, logr.Discard())

	err := appEntity.Update(threescaleapi.Params{})
	require.NoError(t, err)
	require.Equal(t, appEntity.ID(), int64(3))
	require.Equal(t, appEntity.AppName(), "test")
	require.Equal(t, appEntity.Description(), "test")
	require.Equal(t, appEntity.UserAccountID(), int64(3))
	require.Equal(t, appEntity.PlanID(), int64(2))
	require.Equal(t, appEntity.ApplicationState(), "live")
}
