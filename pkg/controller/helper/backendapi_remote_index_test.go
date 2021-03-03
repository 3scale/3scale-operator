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

func TestBackendAPIRemoteIndexFindByID(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.BackendApiList{
			Backends: []threescaleapi.BackendApi{
				{
					Element: threescaleapi.BackendApiItem{
						ID:              int64(1),
						Name:            "Backend 01",
						SystemName:      "backend_01",
						Description:     "some descr 01",
						PrivateEndpoint: "https://example.com",
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
	remoteIndex, err := NewBackendAPIRemoteIndex(client, logrtesting.NullLogger{})
	ok(t, err)

	backendEntity, ok := remoteIndex.FindByID(int64(1))
	assert(t, ok, "backend 1 not found")
	assert(t, backendEntity != nil, "backend entity returned nil")
	equals(t, int64(1), backendEntity.ID())

	backendEntity, ok = remoteIndex.FindByID(int64(2))
	assert(t, !ok, "backend 2 found")
}

func TestBackendAPIRemoteIndexFindBySystemName(t *testing.T) {
	token := "12345"

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		respObject := &threescaleapi.BackendApiList{
			Backends: []threescaleapi.BackendApi{
				{
					Element: threescaleapi.BackendApiItem{
						ID:              int64(1),
						Name:            "Backend 01",
						SystemName:      "backend_01",
						Description:     "some descr 01",
						PrivateEndpoint: "https://example.com",
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
	remoteIndex, err := NewBackendAPIRemoteIndex(client, logrtesting.NullLogger{})
	ok(t, err)

	backendEntity, ok := remoteIndex.FindBySystemName("backend_01")
	assert(t, ok, "backend_01 not found")
	assert(t, backendEntity != nil, "backend entity returned nil")
	equals(t, "backend_01", backendEntity.SystemName())

	backendEntity, ok = remoteIndex.FindBySystemName("not_existing_system_name")
	assert(t, !ok, "unexpected backend found")
}

func TestBackendAPIRemoteIndexCreateBackendAPI(t *testing.T) {
	token := "12345"

	listBackendHandler := func(req *http.Request) *http.Response {
		respObject := &threescaleapi.BackendApiList{
			Backends: []threescaleapi.BackendApi{
				{
					Element: threescaleapi.BackendApiItem{
						ID:              int64(1),
						Name:            "Backend 01",
						SystemName:      "backend_01",
						Description:     "some descr 01",
						PrivateEndpoint: "https://example.com",
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
	}

	createBackendHandler := func(req *http.Request) *http.Response {
		respObject := threescaleapi.BackendApi{
			Element: threescaleapi.BackendApiItem{
				ID:              int64(2),
				Name:            "New Backend",
				SystemName:      "new_backend",
				Description:     "some descr",
				PrivateEndpoint: "https://example.com",
			},
		}

		responseBodyBytes, err := json.Marshal(respObject)
		ok(t, err)

		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewBuffer(responseBodyBytes)),
			Header:     make(http.Header),
		}
	}

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		switch req.Method {
		case "GET":
			return listBackendHandler(req)
		default:
			return createBackendHandler(req)
		}
	})

	client := threescaleapi.NewThreeScale(NewTestAdminPortal(t), token, httpClient)
	remoteIndex, err := NewBackendAPIRemoteIndex(client, logrtesting.NullLogger{})
	ok(t, err)

	backendEntity, err := remoteIndex.CreateBackendAPI(threescaleapi.Params{})
	ok(t, err)
	assert(t, backendEntity != nil, "backend entity returned nil")
	equals(t, "new_backend", backendEntity.SystemName())

	backendEntity, ok := remoteIndex.FindBySystemName("new_backend")
	assert(t, ok, "new backend not found")
	assert(t, backendEntity != nil, "backend entity returned nil")
	equals(t, "new_backend", backendEntity.SystemName())
}
