package operator

import (
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getInternalSecret(namespace string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.BackendSecretInternalApiSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.BackendSecretInternalApiUsernameFieldName: "usernameValue",
			component.BackendSecretInternalApiPasswordFieldName: "passwordValue",
		},
		Type: v1.SecretTypeOpaque,
	}
}

func getListenerSecret(namespace string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.BackendSecretBackendListenerSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.BackendSecretBackendListenerServiceEndpointFieldName: "serviceValue",
			component.BackendSecretBackendListenerRouteEndpointFieldName:   "routeValue",
		},
		Type: v1.SecretTypeOpaque,
	}
}

func getRedisSecret(namespace string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.BackendSecretBackendRedisSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.BackendSecretBackendRedisStorageURLFieldName:           "storageURLValue",
			component.BackendSecretBackendRedisQueuesURLFieldName:            "queueURLValue",
			component.BackendSecretBackendRedisStorageSentinelHostsFieldName: "storageSentinelHostsValue",
			component.BackendSecretBackendRedisStorageSentinelRoleFieldName:  "storageSentinelRoleValue",
			component.BackendSecretBackendRedisQueuesSentinelHostsFieldName:  "queueSentinelHostsValue",
			component.BackendSecretBackendRedisQueuesSentinelRoleFieldName:   "queueSentinelRoleValue",
		},
		Type: v1.SecretTypeOpaque,
	}
}

func TestGetBackendOptions(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true
	var oneValue int64 = 1

	cases := []struct {
		testName                    string
		resourceRequirementsEnabled bool
		internalSecret              *v1.Secret
		listenerSecret              *v1.Secret
		redisSecret                 *v1.Secret
	}{
		{"WithResourceRequirements", true, nil, nil, nil},
		{"InternalSecret", false, getInternalSecret(namespace), nil, nil},
		{"ListenerSecret", false, nil, getListenerSecret(namespace), nil},
		{"RedisSecret", false, nil, nil, getRedisSecret(namespace)},
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
					Backend: &appsv1alpha1.BackendSpec{
						ListenerSpec: &appsv1alpha1.BackendListenerSpec{Replicas: &oneValue},
						WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{Replicas: &oneValue},
						CronSpec:     &appsv1alpha1.BackendCronSpec{Replicas: &oneValue},
					},
				},
			}
			objs := []runtime.Object{apimanager}
			if tc.internalSecret != nil {
				objs = append(objs, tc.internalSecret)
			}
			if tc.listenerSecret != nil {
				objs = append(objs, tc.listenerSecret)
			}
			if tc.redisSecret != nil {
				objs = append(objs, tc.redisSecret)
			}

			cl := fake.NewFakeClient(objs...)
			optsProvider := NewOperatorBackendOptionsProvider(&apimanager.Spec, namespace, cl)
			_, err := optsProvider.GetBackendOptions()
			if err != nil {
				t.Error(err)
			}
			// created "opts" cannot be tested  here, it only has set methods
			// and cannot assert on setted values from a different package
			// TODO: refactor options provider structure
			// then validate setted resources
		})
	}
}
