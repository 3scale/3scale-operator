package operator

import (
	"strings"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetSystemPostgreSQLOptions(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true
	databaseURL := "postgresql://user:password@postgresql.example.com/system"

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
			optsProvider := NewSystemPostgresqlOptionsProvider(apimanager, namespace, cl)
			_, err := optsProvider.GetSystemPostgreSQLOptions()
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

func TestGetPostgreSQLOptionsInvalidURL(t *testing.T) {
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
		// prefix must be postgresql
		{"invalidURL01", "mysql://root:password1@system-postgresql/system", "'postgresql'"},
		// missing user
		{"invalidURL02", "postgresql://system-postgresql/system", "authentication information"},
		// missing user
		{"invalidURL03", "postgresql://:password1@system-postgresql/system", "secret must contain a username"},
		// missing passwd
		{"invalidURL04", "postgresql://user@system-postgresql/system", "secret must contain a password"},
		// path missing
		{"invalidURL05", "postgresql://user:password1@system-postgresql", "database name"},
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
			optsProvider := NewSystemPostgresqlOptionsProvider(apimanager, namespace, cl)
			_, err := optsProvider.GetSystemPostgreSQLOptions()
			if err == nil {
				subT.Fatal("expected to fail for invalid URL")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				subT.Fatalf("expected error regexp: %s, got: (%v)", tc.errSubstr, err)
			}
		})
	}
}
