package operator

import (
	"context"
	"testing"

	"k8s.io/api/policy/v1beta1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemReconcilerCreate(t *testing.T) {
	var (
		log = logf.Log.WithName("operator_test")
	)
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Unable to get config: (%v)", err)
	}
	apimanager := basicApimanagerSpecTestSystemOptions(name, namespace)
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = routev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	BaseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)

	reconciler := NewSystemReconciler(BaseAPIManagerLogicReconciler)
	_, err = reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      runtime.Object
	}{
		{"systemPVC", "system-storage", &v1.PersistentVolumeClaim{}},
		{"systemProviderService", "system-provider", &v1.Service{}},
		{"systemMasterService", "system-master", &v1.Service{}},
		{"systemDeveloperService", "system-developer", &v1.Service{}},
		{"systemSphinxService", "system-sphinx", &v1.Service{}},
		{"systemMemcacheService", "system-memcache", &v1.Service{}},
		{"systemAppDC", "system-app", &appsv1.DeploymentConfig{}},
		{"systemSideKiqDC", "system-sidekiq", &appsv1.DeploymentConfig{}},
		{"systemSphinxqDC", "system-sphinx", &appsv1.DeploymentConfig{}},
		{"systemCM", "system", &v1.ConfigMap{}},
		{"systemEnvironmentCM", "system-environment", &v1.ConfigMap{}},
		{"systemSMTPSecret", "system-smtp", &v1.Secret{}},
		{"systemEventsHookSecret", component.SystemSecretSystemEventsHookSecretName, &v1.Secret{}},
		{"systemRedisSecret", component.SystemSecretSystemRedisSecretName, &v1.Secret{}},
		{"systemMasterApicastSecret", component.SystemSecretSystemMasterApicastSecretName, &v1.Secret{}},
		{"systemSeedSecret", component.SystemSecretSystemSeedSecretName, &v1.Secret{}},
		{"systemRecaptchaSecret", component.SystemSecretSystemRecaptchaSecretName, &v1.Secret{}},
		{"systemAppSecret", component.SystemSecretSystemAppSecretName, &v1.Secret{}},
		{"systemMemcachedSecret", component.SystemSecretSystemMemcachedSecretName, &v1.Secret{}},
		{"systemMemcachedSecret", component.SystemSecretSystemMemcachedSecretName, &v1.Secret{}},
		{"systemAppPDB", "system-app", &v1beta1.PodDisruptionBudget{}},
		{"systemSidekiqPDB", "system-sidekiq", &v1beta1.PodDisruptionBudget{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}
		})
	}
}
