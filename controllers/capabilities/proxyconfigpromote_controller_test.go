package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

func create(x int64) *int64 {
	return &x
}

type (
	fakeThreescaleClient struct{}
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
func getProviderAccount() (Secret *v1.Secret) {
	Secret = &v1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "threescale-provider-account",
			Namespace: "test",
		},
		Immutable: nil,
		Data: map[string][]byte{
			"adminURL": []byte("https://3scale-admin.test.3scale.net"),
			"token":    []byte("token"),
		},
		Type: "Opaque",
	}
	return Secret
}
func newTrue() *bool {
	b := true
	return &b
}
func getProxyConfigPromoteCRStaging() (CR *capabilitiesv1beta1.ProxyConfigPromote) {
	CR = &capabilitiesv1beta1.ProxyConfigPromote{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ProxyConfigPromoteSpec{
			ProductCRName: "test",
		},
	}
	return CR
}
func getProxyConfigPromoteCRProduction() (CR *capabilitiesv1beta1.ProxyConfigPromote) {
	CR = &capabilitiesv1beta1.ProxyConfigPromote{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ProxyConfigPromoteSpec{
			ProductCRName: "test",
			Production:    newTrue(),
		},
	}
	return CR
}

func getProductList() (productList *capabilitiesv1beta1.ProductList) {
	productList = &capabilitiesv1beta1.ProductList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []capabilitiesv1beta1.Product{
			{TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ProductSpec{
					Name:        "test",
					SystemName:  "test",
					Description: "test",
				},
				Status: capabilitiesv1beta1.ProductStatus{
					ID:                  create(3),
					ProviderAccountHost: "some string",
					ObservedGeneration:  1,
					Conditions:          nil,
				},
			},
		},
	}
	return productList
}

func getProductCR() (CR *capabilitiesv1beta1.Product) {

	CR = &capabilitiesv1beta1.Product{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ProductSpec{
			Name:        "test",
			SystemName:  "test",
			Description: "test",
		},
		Status: capabilitiesv1beta1.ProductStatus{
			ID:                  create(3),
			ProviderAccountHost: "some string",
			ObservedGeneration:  1,
			Conditions: common.Conditions{common.Condition{
				Type:   capabilitiesv1beta1.ProductSyncedConditionType,
				Status: v1.ConditionTrue,
			}},
		},
	}
	return CR
}
func mockHttpClient(proxyJson *client.ProxyJSON, productList *client.ProductList, proxyConfigElementSandbox *client.ProxyConfigElement, proxyConfigElementProduction *client.ProxyConfigElement) *http.Client {
	// override httpClient
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		// DeployProductProxy
		if req.Method == "POST" && req.URL.Path == "/admin/api/services/3/proxy/deploy.json" {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(proxyJson))),
			}
		}
		// ListProducts
		if req.Method == "GET" && req.URL.Path == "/admin/api/services.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(productList))),
			}
		}
		// GetLatestProxyConfig sandbox
		if req.Method == "GET" && req.URL.Path == "/admin/api/services/3/proxy/configs/sandbox/latest.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(proxyConfigElementSandbox))),
			}
		}
		// GetLatestProxyConfig production
		if req.Method == "GET" && req.URL.Path == "/admin/api/services/3/proxy/configs/production/latest.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(proxyConfigElementProduction))),
			}
		}
		// PromoteProxyConfig production
		if req.Method == "POST" && req.URL.Path == "/admin/api/services/3/proxy/configs/sandbox/1/promote.json" {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(proxyConfigElementProduction))),
			}
		}

		// PromoteProxyConfig production
		if req.Method == "POST" && req.URL.Path == "/admin/api/services/3/proxy/configs/sandbox/0/promote.json" {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(proxyConfigElementProduction))),
			}
		}

		return nil
	})
	return httpClient
}

func responseBody(class interface{}) (responseBodyBytes []byte) {
	responseBodyBytes, err := json.Marshal(class)
	if err != nil {
		fmt.Println("json marshal error", "err")
	}
	return responseBodyBytes
}

func TestProxyConfigPromoteReconciler_proxyConfigPromoteReconciler(t *testing.T) {
	// product List
	productList := &client.ProductList{
		Products: []client.Product{
			{
				Element: client.ProductItem{
					ID:         3,
					Name:       "test",
					SystemName: "test",
				},
			}, {
				Element: client.ProductItem{
					ID:         4,
					Name:       "test2",
					SystemName: "test2",
				},
			},
		},
	}
	// Complete proxyJson
	proxyJson := &client.ProxyJSON{
		Element: client.ProxyItem{
			Endpoint:        "productionEndpoint.example.com",
			SandboxEndpoint: "staging.example.com",
			UpdatedAt:       "2009-11-17 20:35:59.651387237 +0000",
		},
	}
	proxyConfigElementSandbox := &client.ProxyConfigElement{
		ProxyConfig: client.ProxyConfig{
			ID:          3,
			Version:     1,
			Environment: "",
			Content: client.Content{
				ID:        3,
				CreatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
				UpdatedAt: time.Date(2009, 11, 17, 20, 34, 59, 651387237, time.UTC),
				Proxy: client.ContentProxy{
					ID:        3,
					UpdatedAt: "2009-11-17 20:34:59.651387237 +0000",
				},
			},
		},
	}
	failedProxyConfigElementSandbox := &client.ProxyConfigElement{
		ProxyConfig: client.ProxyConfig{
			ID:          3,
			Version:     0,
			Environment: "",
			Content: client.Content{
				ID:        3,
				CreatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
				UpdatedAt: time.Date(2009, 11, 17, 20, 34, 59, 651387237, time.UTC),
				Proxy: client.ContentProxy{
					ID:        3,
					UpdatedAt: "2009-11-17 20:34:59.651387237 +0000",
				},
			},
		},
	}
	failedProxyConfigElementProduction := &client.ProxyConfigElement{
		ProxyConfig: client.ProxyConfig{
			ID:          3,
			Version:     0,
			Environment: "",
			Content: client.Content{
				ID:        3,
				CreatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
				UpdatedAt: time.Date(2009, 11, 17, 20, 34, 59, 651387237, time.UTC),
				Proxy: client.ContentProxy{
					ID:        3,
					UpdatedAt: "2009-11-17 20:34:59.651387237 +0000",
				},
			},
		},
	}

	proxyConfigElementProduction := &client.ProxyConfigElement{
		ProxyConfig: client.ProxyConfig{
			ID:          3,
			Version:     1,
			Environment: "",
			Content: client.Content{
				ID:        3,
				CreatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
				UpdatedAt: time.Date(2009, 11, 17, 20, 34, 59, 651387237, time.UTC),
				Proxy: client.ContentProxy{
					ID:        3,
					UpdatedAt: "2009-11-17 20:34:59.651387237 +0000",
				},
			},
		},
	}

	// new adminportal
	ap, _ := client.NewAdminPortalFromStr("https://3scale-admin.test.3scale.net")

	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
	}
	type args struct {
		proxyConfigPromote  *capabilitiesv1beta1.ProxyConfigPromote
		reqLogger           logr.Logger
		threescaleAPIClient *client.ThreeScaleClient
		product             *capabilitiesv1beta1.Product
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ProxyConfigPromoteStatusReconciler
		wantErr bool
	}{
		{
			name: "Test promotion to staging Completed",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
			},
			args: args{
				proxyConfigPromote:  getProxyConfigPromoteCRStaging(),
				reqLogger:           logf.Log.WithName("test reqlogger"),
				threescaleAPIClient: client.NewThreeScale(ap, "test", mockHttpClient(proxyJson, productList, proxyConfigElementSandbox, failedProxyConfigElementProduction)),
				product:             getProductCR(),
			},
			want: &ProxyConfigPromoteStatusReconciler{
				BaseReconciler:          getBaseReconciler(),
				resource:                getProxyConfigPromoteCRStaging(),
				state:                   "Completed",
				productID:               "3",
				latestProductionVersion: 0,
				latestStagingVersion:    1,
				reconcileError:          nil,
				logger:                  logf.Log.WithValues("Status Reconciler", "test"),
			},
			wantErr: false,
		},
		{
			name: "Test promotion to Production Completed",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
			},
			args: args{
				proxyConfigPromote:  getProxyConfigPromoteCRProduction(),
				reqLogger:           logf.Log.WithName("test reqlogger"),
				threescaleAPIClient: client.NewThreeScale(ap, "test", mockHttpClient(proxyJson, productList, proxyConfigElementSandbox, proxyConfigElementProduction)),
				product:             getProductCR(),
			},
			want: &ProxyConfigPromoteStatusReconciler{
				BaseReconciler:          getBaseReconciler(),
				resource:                getProxyConfigPromoteCRProduction(),
				state:                   "Completed",
				productID:               "3",
				latestProductionVersion: 1,
				latestStagingVersion:    1,
				reconcileError:          nil,
				logger:                  logf.Log.WithValues("Status Reconciler", "test"),
			},
			wantErr: false,
		},
		{
			name: "Test promotion to Production Failed",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
			},
			args: args{
				proxyConfigPromote:  getProxyConfigPromoteCRProduction(),
				reqLogger:           logf.Log.WithName("test reqlogger"),
				threescaleAPIClient: client.NewThreeScale(ap, "test", mockHttpClient(proxyJson, productList, failedProxyConfigElementSandbox, failedProxyConfigElementProduction)),
				product:             getProductCR(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ProxyConfigPromoteReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
			}
			got, err := r.proxyConfigPromoteReconciler(tt.args.proxyConfigPromote, tt.args.reqLogger, tt.args.threescaleAPIClient, tt.args.product)
			if (err != nil) && tt.wantErr {
				t.Logf("proxyConfigPromoteReconciler(), wantErr %v", tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.state, tt.want.state) {
				t.Errorf("proxyConfigPromoteReconciler() got.state = %v, want.state %v", got.state, tt.want.state)
			}
			if !reflect.DeepEqual(got.productID, tt.want.productID) {
				t.Errorf("proxyConfigPromoteReconciler() got.productID = %v, want.productID %v", got.productID, tt.want.productID)
			}
			if !reflect.DeepEqual(got.latestProductionVersion, tt.want.latestProductionVersion) {
				t.Errorf("proxyConfigPromoteReconciler() got.latestProductionVersion = %v, want.latestProductionVersion %v", got.latestProductionVersion, tt.want.latestProductionVersion)
			}
			if !reflect.DeepEqual(got.latestStagingVersion, tt.want.latestStagingVersion) {
				t.Errorf("proxyConfigPromoteReconciler() got.latestStagingVersion = %v, want.latestStagingVersion %v", got.latestStagingVersion, tt.want.latestStagingVersion)
			}
		})
	}
}
