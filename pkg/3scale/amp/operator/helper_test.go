package operator

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

const (
	wildcardDomain       = "test.3scale.net"
	appLabel             = "someLabel"
	apimanagerName       = "example-apimanager"
	namespace            = "someNS"
	tenantName           = "someTenant"
	insecureImportPolicy = false
	trueValue            = true
)

func addExpectedMeteringLabels(src map[string]string, componentName string, componentType helper.ComponentType) {
	labels := []struct {
		k string
		v string
	}{
		{"com.company", "Red_Hat"},
		{"rht.prod_name", "Red_Hat_Integration"},
		{"rht.prod_ver", "2021.Q4"},
		{"rht.comp", "3scale"},
		{"rht.comp_ver", product.ThreescaleRelease},
		{"rht.subcomp", componentName},
		{"rht.subcomp_t", string(componentType)},
	}
	for _, label := range labels {
		src[label.k] = label.v
	}
}

func basicApimanager() *appsv1alpha1.APIManager {
	tmpAppLabel := appLabel
	tmpTenantName := tenantName
	tmpInsecureImportPolicy := insecureImportPolicy
	tmpTrueValue := trueValue

	apimanager := &appsv1alpha1.APIManager{
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

	_, err := apimanager.SetDefaults()
	if err != nil {
		panic(fmt.Errorf("Error creating Basic APIManager: %v", err))
	}
	return apimanager
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

func getTestAffinity(prefix string) *v1.Affinity {
	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					v1.NodeSelectorTerm{
						MatchFields: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      fmt.Sprintf("%s-%s", prefix, "key2"),
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{fmt.Sprintf("%s-%s", prefix, "val2")},
							},
						},
					},
				},
			},
		},
	}
}

func getTestTolerations(prefix string) []v1.Toleration {
	return []v1.Toleration{
		v1.Toleration{
			Key:      fmt.Sprintf("%s-%s", prefix, "key1"),
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    fmt.Sprintf("%s-%s", prefix, "val1"),
		},
		v1.Toleration{
			Key:      fmt.Sprintf("%s-%s", prefix, "key2"),
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    fmt.Sprintf("%s-%s", prefix, "val2"),
		},
	}
}
