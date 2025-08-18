package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestApplicationAuthReconciler_syncApplicationAuth(t *testing.T) {
	appID := int64(3)
	userAccountID := int64(3)
	logger := logf.Log.WithName("ApplicationAuth reconciler")

	tests := []struct {
		name        string
		mockServer  *mockApplicationAuthServer
		authMode    string
		authSecret  AuthSecret
		expectedKey string
		wantErr     bool
	}{
		{
			name: "Empty userkey with empty secret",
			mockServer: &mockApplicationAuthServer{
				authMode:      "1",
				userKey:       "",
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "1",
			authSecret:  getEmptyAuthSecret(),
			expectedKey: "",
			wantErr:     false,
		},
		{
			name: "update empty user_key with value from secret",
			mockServer: &mockApplicationAuthServer{
				authMode:      "1",
				userKey:       "",
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "1",
			authSecret:  getAuthSecret(),
			expectedKey: "testkey",
			wantErr:     false,
		},
		{
			name: "update existing user_key with value from secret",
			mockServer: &mockApplicationAuthServer{
				authMode:      "1",
				userKey:       "initalkey",
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "1",
			authSecret:  getAuthSecret(),
			expectedKey: "testkey",
			wantErr:     false,
		},
		{
			name: "update existing user_key with the same value should not return error",
			mockServer: &mockApplicationAuthServer{
				authMode:      "1",
				userKey:       "testkey",
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:   "1",
			authSecret: getAuthSecret(), expectedKey: "testkey",
			wantErr: false,
		},
		{
			name: "returns error with empty application_key with empty secret",
			mockServer: &mockApplicationAuthServer{
				authMode:      "2",
				keys:          []string{},
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "2",
			authSecret:  getEmptyAuthSecret(),
			expectedKey: "",
			wantErr:     true,
		},
		{
			name: "update existing app_key with value from secret",
			mockServer: &mockApplicationAuthServer{
				authMode:      "2",
				keys:          []string{"initalkey"},
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "2",
			authSecret:  getAuthSecret(),
			expectedKey: "testkey",
			wantErr:     false,
		},
		{
			name: "update existing app_key with the same value should not return error",
			mockServer: &mockApplicationAuthServer{
				authMode:      "2",
				keys:          []string{"testkey"},
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "2",
			authSecret:  getAuthSecret(),
			expectedKey: "testkey",
			wantErr:     false,
		},
		{
			name: "update existing app_key with the same value should not return error",
			mockServer: &mockApplicationAuthServer{
				authMode:      "2",
				keys:          []string{"testkey"},
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode:    "2",
			authSecret:  getAuthSecret(),
			expectedKey: "testkey",
			wantErr:     false,
		},
		{
			name: "update with secret contains multiple app_key",
			mockServer: &mockApplicationAuthServer{
				authMode:      "2",
				keys:          []string{"testkey"},
				userAccountID: appID,
				appID:         userAccountID,
			},
			authMode: "2",
			authSecret: AuthSecret{
				ApplicationKey: "testkey1,testkey2,testkey3",
			},
			expectedKey: "testkey1,testkey2,testkey3",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.mockServer.GetServer()
			ap, _ := threescaleapi.NewAdminPortalFromStr(srv.URL)
			threescaleClient := threescaleapi.NewThreeScale(ap, "test", srv.Client())

			err := syncApplicationAuth(userAccountID, appID, tt.authMode, tt.authSecret, threescaleClient, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncApplicationAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			newKey := tt.mockServer.GetKey(tt.authMode)
			if newKey != tt.expectedKey {
				t.Fatalf("mismatch keys, expected: %s - got: %s", tt.expectedKey, newKey)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("syncApplicationAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplicationAuthReconciler_authSecretReferenceSource(t *testing.T) {
	logger := logf.Log.WithName("applicationAuth")
	ns := "test"

	tests := []struct {
		name           string
		authMode       string
		generateSecret bool
		secretData     map[string][]byte
		wantErr        bool
		err            string
	}{
		{
			name:           "return error when secret is empty",
			authMode:       "1",
			generateSecret: true,
			secretData:     map[string][]byte{},
			wantErr:        true,
			err:            "secret field 'UserKey' is required in secret 'test'",
		},
		{
			name:           "generate user_key when secret is empty",
			authMode:       "1",
			generateSecret: true,
			secretData:     map[string][]byte{"UserKey": []byte("")},
			wantErr:        false,
			err:            "",
		},
		{
			name:           "use user_key value in secret is empty",
			authMode:       "1",
			generateSecret: true,
			secretData:     map[string][]byte{"UserKey": []byte("testkey")},
			wantErr:        false,
			err:            "",
		},
		{
			name:           "return error when secret is empty",
			authMode:       "2",
			generateSecret: true,
			secretData:     map[string][]byte{},
			wantErr:        true,
			err:            "secret field 'ApplicationKey' is required in secret 'test'",
		},
		{
			name:           "generate app_key when secret is empty",
			authMode:       "2",
			generateSecret: true,
			secretData:     map[string][]byte{"ApplicationKey": []byte("")},
			wantErr:        false,
			err:            "",
		},
		{
			name:           "use app_key value in secret is empty",
			authMode:       "2",
			generateSecret: true,
			secretData:     map[string][]byte{"ApplicationKey": []byte("testkey")},
			wantErr:        false,
			err:            "",
		},
		{
			name:           "return error with unknown authMode",
			authMode:       "unknown",
			generateSecret: true,
			secretData:     map[string][]byte{},
			wantErr:        true,
			err:            "unknown authentication mode",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Immutable:  nil,
				Data:       tt.secretData,
				StringData: nil,
				Type:       "",
			}

			secretRef := &corev1.LocalObjectReference{
				Name: "test",
			}

			reconciler := getBaseReconciler(secret)
			client := reconciler.Client()
			authSecret, err := authSecretReferenceSource(client, ns, secretRef, tt.generateSecret, tt.authMode, logger)

			if (err != nil) != tt.wantErr {
				t.Fatalf("authSecretReferenceSource() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.err != err.Error() {
					t.Fatalf("authSecretReferenceSource() error = %v, wantErr %v", err, tt.err)
				}
			} else {
				newSecret := &corev1.Secret{}
				err = client.Get(context.Background(), types.NamespacedName{
					Name:      secretRef.Name,
					Namespace: ns,
				}, newSecret)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				switch tt.authMode {
				case "1":
					if authSecret.UserKey != string(newSecret.Data["UserKey"]) {
						t.Fatalf("mismatch user_key expected = '%s', got '%s'", authSecret.UserKey, newSecret.Data["UserKey"])
					}
				case "2":
					if authSecret.ApplicationKey != string(newSecret.Data["ApplicationKey"]) {
						t.Fatalf("mismatch user_key expected = '%s', got '%s'", authSecret.ApplicationKey, newSecret.Data["ApplicationKey"])
					}
				}
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
			GenerateSecret:    ptr.To(true),
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
			GenerateSecret:    ptr.To(false),
			AuthSecretRef: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProviderAccountRef: nil,
		},
	}
	return CR
}

func getEmptyAuthSecret() AuthSecret {
	authSecret := AuthSecret{
		UserKey:        "",
		ApplicationKey: "",
		ApplicationID:  "",
	}
	return authSecret
}

func getAuthSecret() AuthSecret {
	authSecret := AuthSecret{
		UserKey:        "testkey",
		ApplicationKey: "testkey",
		ApplicationID:  "",
	}
	return authSecret
}

type mockApplicationAuthServer struct {
	authMode      string
	appID         int64
	userAccountID int64
	userKey       string
	keys          []string
}

func (m *mockApplicationAuthServer) GetServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/api/accounts/{accoundID}/applications/{applicationID}", m.applicationHandler)
	mux.HandleFunc("PUT /admin/api/accounts/{accoundID}/applications/{applicationID}", m.applicationHandler)
	mux.HandleFunc("GET /admin/api/accounts/{accoundID}/applications/{applicationID}/keys.json", m.applicationKeysHandler)
	mux.HandleFunc("DELETE /admin/api/accounts/{accoundID}/applications/{applicationID}/keys/{key}", m.applicationKeysHandler)
	mux.HandleFunc("POST /admin/api/accounts/{accoundID}/applications/{applicationID}/keys.json", m.applicationKeysHandler)

	return httptest.NewServer(mux)
}

func (m *mockApplicationAuthServer) GetKey(mode string) string {
	switch mode {
	case "1":
		return m.userKey
	case "2":
		return strings.Join(m.keys, ",")
	default:
		return ""
	}
}

func (m *mockApplicationAuthServer) applicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		r.ParseForm()
		userKey := r.FormValue("user_key")
		if userKey != "" {
			if userKey == m.userKey {
				errorResponse(w, "user_key", []string{"has already been taken"})
				return
			} else {
				m.userKey = userKey
			}
		}
	}

	data := threescaleapi.ApplicationElem{
		Application: threescaleapi.Application{
			UserAccountID: strconv.FormatInt(m.userAccountID, 10),
			ID:            m.appID,
			AppName:       "newName",
			UserKey:       m.userKey,
		},
	}

	json, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func (m *mockApplicationAuthServer) applicationKeysHandler(w http.ResponseWriter, r *http.Request) {
	var keyLimit int

	switch r.Method {
	case http.MethodPost:
		r.ParseForm()
		key := r.FormValue("key")

		if len(key) < 5 {
			errorResponse(w, "value", []string{"is too short (minimum is 5 characters)"})
			return
		}

		// if key already existed, returns error
		if helper.ArrayContains(m.keys, key) {
			errorResponse(w, "value", []string{"has already been taken"})
			return
		}

		if m.authMode == "2" {
			keyLimit = 5
		}

		// Check if the current lenght does not exceed 5 keys limit
		if len(m.keys) == keyLimit {
			errorResponse(w, "base", []string{"Limit reached"})
			return
		}

		m.keys = append(m.keys, key)

		data := &threescaleapi.ApplicationElem{
			Application: threescaleapi.Application{
				UserAccountID: strconv.FormatInt(m.userAccountID, 10),
				ID:            m.appID,
				AppName:       "newName",
			},
		}

		json, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	case http.MethodDelete:
		key := strings.TrimSuffix(r.PathValue("key"), ".json")

		newKeys := slices.DeleteFunc(m.keys, func(existingKey string) bool {
			return existingKey == key
		})
		m.keys = newKeys
		return
	case http.MethodGet:
		keysObj := []threescaleapi.ApplicationKeyWrapper{}

		for _, key := range m.keys {
			keyObj := threescaleapi.ApplicationKeyWrapper{
				Key: threescaleapi.ApplicationKey{
					Value: key,
				},
			}
			keysObj = append(keysObj, keyObj)
		}

		data := &threescaleapi.ApplicationKeysElem{
			Keys: keysObj,
		}

		json, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(json)
	}
}

func errorResponse(w http.ResponseWriter, key string, value []string) {
	errObj := struct {
		Errors map[string][]string `json:"errors"`
	}{
		Errors: map[string][]string{key: value},
	}

	data, _ := json.Marshal(errObj)
	http.Error(w, string(data), http.StatusUnprocessableEntity)
}
