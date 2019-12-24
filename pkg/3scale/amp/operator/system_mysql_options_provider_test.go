package operator

import (
	"strings"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getSystemDBSecret(namespace, databaseURL string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemDatabaseSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemDatabaseUserFieldName:     "user",
			component.SystemSecretSystemDatabasePasswordFieldName: "password1",
			component.SystemSecretSystemDatabaseURLFieldName:      databaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func TestGetMysqlOptions(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true
	databaseURL := "mysql2://root:password1@system-mysql/system"

	cases := []struct {
		testName                    string
		resourceRequirementsEnabled bool
		systemDatabaseSecret        *v1.Secret
	}{
		{"WithResourceRequirements", true, nil},
		{"SystemDBSecret", false, getSystemDBSecret(namespace, databaseURL)},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			apimanager := &appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
						WildcardDomain:               wildcardDomain,
						TenantName:                   &tenantName,
						ResourceRequirementsEnabled:  &tc.resourceRequirementsEnabled,
					},
				},
			}
			objs := []runtime.Object{apimanager}
			if tc.systemDatabaseSecret != nil {
				objs = append(objs, tc.systemDatabaseSecret)
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewSystemMysqlOptionsProvider(&apimanager.Spec, namespace, cl)
			_, err := optsProvider.GetMysqlOptions()
			if err != nil {
				subT.Error(err)
			}
			// created "opts" cannot be tested  here, it only has set methods
			// and cannot assert on setted values from a different package
			// TODO: refactor options provider structure
			// then validate setted resources
		})
	}
}

func TestGetMysqlOptionsInvalidURL(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true

	cases := []struct {
		testName    string
		databaseURL string
		errSubstr   string
	}{
		// prefix must be mysql2
		{"invalidURL01", "mysql://root:password1@system-mysql/system", "'mysql2'"},
		// missing user
		{"invalidURL02", "mysql2://system-mysql/system", "authentication information"},
		// user not root
		{"invalidURL03", "mysql2://user:password1@system-mysql/system", "'root'"},
		// passwd missing
		{"invalidURL04", "mysql2://root@system-mysql/system", "secret must contain a password"},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			apimanager := &appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
						WildcardDomain:               wildcardDomain,
						TenantName:                   &tenantName,
						ResourceRequirementsEnabled:  &trueValue,
					},
				},
			}
			secret := getSystemDBSecret(namespace, tc.databaseURL)
			objs := []runtime.Object{apimanager, secret}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewSystemMysqlOptionsProvider(&apimanager.Spec, namespace, cl)
			_, err := optsProvider.GetMysqlOptions()
			if err == nil {
				subT.Fatal("expected to fail for invalid URL")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				subT.Fatalf("expected error regexp: %s, got: (%v)", tc.errSubstr, err)
			}
		})
	}
}
