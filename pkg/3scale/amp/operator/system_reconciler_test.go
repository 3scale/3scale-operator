package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemReconcilerCreate(t *testing.T) {
	var (
		log = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()

	apimanager := basicApimanagerSpecTestSystemOptions()
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
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
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := configv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	reconciler := NewSystemReconciler(baseAPIManagerLogicReconciler)
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

func TestSystemReconcilerDisableReplicaSyncingAnnotations(t *testing.T) {
	var (
		namespace       = "someNS"
		log             = logf.Log.WithName("operator_test")
		twoValue  int32 = 2
	)
	ctx := context.TODO()
	s := scheme.Scheme

	err := appsv1alpha1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := configv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName                 string
		objName                  string
		obj                      runtime.Object
		apimanager               *appsv1alpha1.APIManager
		annotation               string
		annotationValue          string
		expectedAmountOfReplicas int32
		validatingFunction       func(*appsv1alpha1.APIManager, *appsv1.DeploymentConfig, string, string, int32) bool
	}{
		{"systemAppDC-annotation not present", "system-app", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem("someAnnotation", "false"), disableSystemAppInstancesSyncing, "dummy", int32(3), confirmReplicasWhenAnnotationIsNotPresent},
		{"systemAppDC-annotation false", "system-app", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem(disableSystemAppInstancesSyncing, "false"), disableSystemAppInstancesSyncing, "false", int32(3), confirmReplicasWhenAnnotationPresent},
		{"systemAppDC-annotation true", "system-app", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem(disableSystemAppInstancesSyncing, "true"), disableSystemAppInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"systemAppDC-annotation true of dummy value", "system-app", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem(disableSystemAppInstancesSyncing, "true"), disableSystemAppInstancesSyncing, "someDummyValue", int32(3), confirmReplicasWhenAnnotationPresent},

		{"systemSideKiqDC-annotation not present", "system-sidekiq", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem("someAnnotation", "false"), disableSidekiqInstancesSyncing, "dummy", int32(4), confirmReplicasWhenAnnotationIsNotPresent},
		{"systemSideKiqDC-annotation false", "system-sidekiq", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem(disableSidekiqInstancesSyncing, "false"), disableSidekiqInstancesSyncing, "false", int32(4), confirmReplicasWhenAnnotationPresent},
		{"systemSideKiqDC-annotation true", "system-sidekiq", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem(disableSidekiqInstancesSyncing, "true"), disableSidekiqInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"systemSideKiqDC-annotation true of dummy value", "system-sidekiq", &appsv1.DeploymentConfig{}, apiManagerCreatorSystem(disableSidekiqInstancesSyncing, "true"), disableSidekiqInstancesSyncing, "someDummyValue", int32(4), confirmReplicasWhenAnnotationPresent},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{tc.apimanager}
			// Create a fake client to mock API calls.
			cl := fake.NewFakeClient(objs...)
			clientAPIReader := fake.NewFakeClient(objs...)
			clientset := fakeclientset.NewSimpleClientset()
			recorder := record.NewFakeRecorder(10000)
			baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
			baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, tc.apimanager)

			systemReconciler := NewSystemReconciler(baseAPIManagerLogicReconciler)
			_, err = systemReconciler.Reconcile()
			if err != nil {
				t.Fatal(err)
			}

			dc := &appsv1.DeploymentConfig{}
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}

			err = cl.Get(context.TODO(), namespacedName, dc)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			// bump the amount of replicas in the dc
			dc.Spec.Replicas = twoValue
			err = cl.Update(context.TODO(), dc)
			if err != nil {
				subT.Errorf("error updating dc of %s: %v", tc.objName, err)
			}

			// re-run the reconciler
			_, err = systemReconciler.Reconcile()
			if err != nil {
				t.Fatal(err)
			}

			err = cl.Get(context.TODO(), namespacedName, dc)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			correct := tc.validatingFunction(tc.apimanager, dc, tc.annotation, tc.annotationValue, tc.expectedAmountOfReplicas)
			if !correct {
				subT.Errorf("value of expteced replicas does not match for %s. expected: %v actual: %v", tc.objName, tc.expectedAmountOfReplicas, dc.Spec.Replicas)
			}
		})
	}
}

func apiManagerCreatorSystem(disableSyncAnnotation string, disableSyncAnnotationValue string) *appsv1alpha1.APIManager {
	tmpSystemAppReplicas := systemAppReplicas
	tmpSystemSideKiqReplicas := systemSidekiqReplicas
	tmpApicastRegistryURL := apicastRegistryURL

	apimanager := basicApimanager()
	apimanager.Annotations = map[string]string{disableSyncAnnotation: disableSyncAnnotationValue}
	apimanager.Spec.Apicast = &appsv1alpha1.ApicastSpec{RegistryURL: &tmpApicastRegistryURL}
	apimanager.Spec.System = &appsv1alpha1.SystemSpec{
		FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{},
		AppSpec:         &appsv1alpha1.SystemAppSpec{Replicas: &tmpSystemAppReplicas},
		SidekiqSpec:     &appsv1alpha1.SystemSidekiqSpec{Replicas: &tmpSystemSideKiqReplicas},
		SphinxSpec:      &appsv1alpha1.SystemSphinxSpec{},
	}
	apimanager.Spec.PodDisruptionBudget = &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true}
	return apimanager
}
