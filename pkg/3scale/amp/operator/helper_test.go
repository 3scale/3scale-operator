package operator

import (
	"os"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	wildcardDomain       = "test.3scale.net"
	appLabel             = "someLabel"
	apimanagerName       = "example-apimanager"
	namespace            = "someNS"
	tenantName           = "someTenant"
	insecureImportPolicy = false
	trueValue            = true
	falseValue           = false
)

func basicApimanager() *appsv1alpha1.APIManager {
	tmpAppLabel := appLabel
	tmpTenantName := tenantName
	tmpInsecureImportPolicy := insecureImportPolicy
	tmpTrueValue := trueValue

	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:               wildcardDomain,
				AppLabel:                     &tmpAppLabel,
				ImageStreamTagImportInsecure: &tmpInsecureImportPolicy,
				TenantName:                   &tmpTenantName,
				ResourceRequirementsEnabled:  &tmpTrueValue,
			},
			System: &appsv1alpha1.SystemSpec{},
		},
	}
}

func TestMain(m *testing.M) {
	logf.SetLogger(zap.Logger())
	os.Exit(m.Run())
}

func GetTestSecret(namespace, secretName string, data map[string]string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: data,
		Type:       v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getSystemDBSecret(databaseURL, username, password string) *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemDatabaseUserFieldName:     username,
		component.SystemSecretSystemDatabasePasswordFieldName: password,
		component.SystemSecretSystemDatabaseURLFieldName:      databaseURL,
	}
	return GetTestSecret(namespace, component.SystemSecretSystemDatabaseSecretName, data)
}
