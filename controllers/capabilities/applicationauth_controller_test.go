package controllers

import (
	"bytes"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"net/http"
	"reflect"
	"strconv"
	"testing"
)

func TestApplicationAuthReconciler_applicationAuthReconciler(t *testing.T) {
	ap, _ := threescaleapi.NewAdminPortalFromStr("https://3scale-admin.test.3scale.net")
	applicationKey := "4efd48e3e2ecfdea1fc21eeddf0610b9"
	appID := int64(3)
	userAccountID := int64(3)
	applicationUpdate := &threescaleapi.ApplicationElem{
		Application: threescaleapi.Application{
			UserAccountID: strconv.FormatInt(userAccountID, 10),
			ID:            appID,
			AppName:       "newName",
		},
	}
	applicationKeyCreate := &threescaleapi.ApplicationElem{
		Application: threescaleapi.Application{
			ID: appID,
		},
	}
	applicationKeyList := &threescaleapi.ApplicationKeysElem{
		Keys: []threescaleapi.ApplicationKeyWrapper{
			{
				Key: threescaleapi.ApplicationKey{
					Value: applicationKey,
				},
			},
		},
	}

	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
	}
	type args struct {
		applicationAuth  *capabilitiesv1beta1.ApplicationAuth
		developerAccount *capabilitiesv1beta1.DeveloperAccount
		application      *capabilitiesv1beta1.Application
		product          *capabilitiesv1beta1.Product
		authSecret       AuthSecret
		threescaleClient *threescaleapi.ThreeScaleClient
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ApplicationAuthStatusReconciler
		wantErr bool
	}{
		{
			name: "Test generate secret",
			fields: fields{
				BaseReconciler: getBaseReconciler(getEmptyAuthSecretObj()),
			},
			args: args{
				applicationAuth:  getApplicationAuthGenerateSecret(),
				application:      getApplicationCR(),
				product:          getProductCR(),
				developerAccount: getApplicationDeveloperAccount(),
				authSecret:       getEmptyAuthSecret(),
				threescaleClient: threescaleapi.NewThreeScale(ap, "test", mockHttpApplicationAuthClient(applicationUpdate, applicationKeyCreate, applicationKeyList)),
			},
			want:    NewApplicationAuthStatusReconciler(getBaseReconciler(getApplicationAuthGenerateSecret()), getApplicationAuthGenerateSecret(), nil),
			wantErr: false,
		},
		{
			name: "Test populated secret",
			fields: fields{
				BaseReconciler: getBaseReconciler(getAuthSecretObj()),
			},
			args: args{
				applicationAuth:  getApplicationAuth(),
				application:      getApplicationCR(),
				product:          getProductCR(),
				developerAccount: getApplicationDeveloperAccount(),
				authSecret:       getAuthSecret(),
				threescaleClient: threescaleapi.NewThreeScale(ap, "test", mockHttpApplicationAuthClient(applicationUpdate, applicationKeyCreate, applicationKeyList)),
			},
			want:    NewApplicationAuthStatusReconciler(getBaseReconciler(getApplicationAuth()), getApplicationAuth(), nil),
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationAuthReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
			}
			got, err := r.applicationAuthReconciler(tt.args.applicationAuth, tt.args.developerAccount, tt.args.application, tt.args.product, tt.args.authSecret, tt.args.threescaleClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("applicationAuthReconciler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.reconcileError, tt.want.reconcileError) {
				t.Errorf("applicationAuthReconciler() got = %v, want %v", got.reconcileError, tt.want.reconcileError)
			}
			if !reflect.DeepEqual(got.resource, tt.want.resource) {
				t.Errorf("applicationAuthReconciler() got = %v, want %v", got.resource, tt.want.resource)
			}
		})
	}
}

func getApplicationAuthGenerateSecret() (CR *capabilitiesv1beta1.ApplicationAuth) {
	CR = &capabilitiesv1beta1.ApplicationAuth{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationAuthSpec{
			ApplicationCRName: "test",
			GenerateSecret:    pointer.Bool(true),
			AuthSecretRef: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProviderAccountRef: nil,
		},
	}
	return CR

}

func getApplicationAuth() (CR *capabilitiesv1beta1.ApplicationAuth) {
	CR = &capabilitiesv1beta1.ApplicationAuth{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationAuthSpec{
			ApplicationCRName: "test",
			GenerateSecret:    pointer.Bool(false),
			AuthSecretRef: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProviderAccountRef: nil,
		},
	}
	return CR

}

func getEmptyAuthSecretObj() *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Immutable: nil,
		Data: map[string][]byte{
			"UserKey":        []byte(""),
			"ApplicationKey": []byte(""),
		},
		StringData: nil,
		Type:       "",
	}
	return secret
}

func getEmptyAuthSecret() AuthSecret {
	authSecret := AuthSecret{
		UserKey:        "",
		ApplicationKey: "",
		ApplicationID:  "",
	}
	return authSecret
}

func getAuthSecretObj() *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Immutable: nil,
		Data: map[string][]byte{
			"UserKey":        []byte("testkey"),
			"ApplicationKey": []byte("testkey"),
		},
		StringData: nil,
		Type:       "",
	}
	return secret
}
func getAuthSecret() AuthSecret {
	authSecret := AuthSecret{
		UserKey:        "testkey",
		ApplicationKey: "testkey",
		ApplicationID:  "",
	}
	return authSecret
}

func mockHttpApplicationAuthClient(applicationUpdate *threescaleapi.ApplicationElem, applicationKeyCreate *threescaleapi.ApplicationElem, applicationKeyList *threescaleapi.ApplicationKeysElem) *http.Client {
	// override httpClient
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		if req.Method == http.MethodPut && req.URL.Path == "/admin/api/accounts/3/applications/3.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(applicationUpdate))),
			}
		}
		if req.Method == http.MethodPost && req.URL.Path == "/admin/api/accounts/3/applications/3/keys.json" {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(applicationKeyCreate))),
			}
		}
		if req.Method == http.MethodGet && req.URL.Path == "/admin/api/accounts/3/applications/3/keys.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBuffer(responseBody(applicationKeyList))),
			}
		}

		return nil
	})
	return httpClient
}
