package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestNewBackendReconciler(t *testing.T) {
	var (
		name                 = "example-apimanager"
		namespace            = "operator-unittest"
		wildcardDomain       = "test.3scale.net"
		log                  = logf.Log.WithName("operator_test")
		appLabel             = "someLabel"
		tenantName           = "someTenant"
		trueValue            = true
		oneValue       int64 = 1
	)

	ctx := context.TODO()

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
			Backend: &appsv1alpha1.BackendSpec{
				ListenerSpec: &appsv1alpha1.BackendListenerSpec{Replicas: &oneValue},
				WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{Replicas: &oneValue},
				CronSpec:     &appsv1alpha1.BackendCronSpec{Replicas: &oneValue},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
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

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	BaseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	backendReconciler := NewBackendReconciler(BaseAPIManagerLogicReconciler)
	_, err = backendReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      runtime.Object
	}{
		{"cronDC", "backend-cron", &appsv1.DeploymentConfig{}},
		{"listenerDC", "backend-listener", &appsv1.DeploymentConfig{}},
		{"listenerService", "backend-listener", &v1.Service{}},
		{"listenerRoute", "backend", &routev1.Route{}},
		{"workerDC", "backend-worker", &appsv1.DeploymentConfig{}},
		{"environmentCM", "backend-environment", &v1.ConfigMap{}},
		{"internalAPISecret", component.BackendSecretInternalApiSecretName, &v1.Secret{}},
		{"listenerSecret", component.BackendSecretBackendListenerSecretName, &v1.Secret{}},
		{"workerPDB", "backend-worker", &v1beta1.PodDisruptionBudget{}},
		{"cronPDB", "backend-cron", &v1beta1.PodDisruptionBudget{}},
		{"listenerPDB", "backend-listener", &v1beta1.PodDisruptionBudget{}},
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

func TestAnnotationsBackendReconciler(t *testing.T) {
	var (
		namespace = "operator-unittest"
		log       = logf.Log.WithName("operator_test")
		// oneValue             int64 = 1
		twoValue int32 = 2
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
	err = routev1.AddToScheme(s)
	if err != nil {
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
		{"cronDC-annotation not present", "backend-cron", &appsv1.DeploymentConfig{}, backendApiManagerCreator("someAnnotation", "false"), disableCronInstancesSyncing, "dummy", int32(1), confirmReplicasWhenAnnotationIsNotPresent},
		{"cronDC-annotation false", "backend-cron", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableCronInstancesSyncing, "false"), disableCronInstancesSyncing, "false", int32(1), confirmReplicasWhenAnnotationPresent},
		{"cronDC-annotation true", "backend-cron", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableCronInstancesSyncing, "true"), disableCronInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"cronDC-annotation true of dummy value", "backend-cron", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableCronInstancesSyncing, "true"), disableCronInstancesSyncing, "someDummyValue", int32(1), confirmReplicasWhenAnnotationPresent},

		{"listenerDC-annotation not present", "backend-listener", &appsv1.DeploymentConfig{}, backendApiManagerCreator("someAnnotation", "false"), disableBackendListenerInstancesSyncing, "dummy", int32(1), confirmReplicasWhenAnnotationIsNotPresent},
		{"listenerDC-annotation false", "backend-listener", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableBackendListenerInstancesSyncing, "false"), disableBackendListenerInstancesSyncing, "false", int32(1), confirmReplicasWhenAnnotationPresent},
		{"listenerDC-annotation true", "backend-listener", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableBackendListenerInstancesSyncing, "true"), disableBackendListenerInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"listenerDC-annotation true of dummy value", "backend-listener", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableBackendListenerInstancesSyncing, "true"), disableBackendListenerInstancesSyncing, "someDummyValue", int32(1), confirmReplicasWhenAnnotationPresent},

		{"workerDC-annotation not present", "backend-worker", &appsv1.DeploymentConfig{}, backendApiManagerCreator("someAnnotation", "false"), disableBackendWorkerInstancesSyncing, "dummy", int32(1), confirmReplicasWhenAnnotationIsNotPresent},
		{"workerDC-annotation false", "backend-worker", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableBackendWorkerInstancesSyncing, "false"), disableBackendWorkerInstancesSyncing, "false", int32(1), confirmReplicasWhenAnnotationPresent},
		{"workerDC-annotation true", "backend-worker", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableBackendWorkerInstancesSyncing, "true"), disableBackendWorkerInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"workerDC-annotation true of dummy value", "backend-worker", &appsv1.DeploymentConfig{}, backendApiManagerCreator(disableBackendWorkerInstancesSyncing, "true"), disableBackendWorkerInstancesSyncing, "someDummyValue", int32(1), confirmReplicasWhenAnnotationPresent},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{tc.apimanager}
			cl := fake.NewFakeClient(objs...)
			clientAPIReader := fake.NewFakeClient(objs...)
			clientset := fakeclientset.NewSimpleClientset()
			recorder := record.NewFakeRecorder(10000)
			baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
			baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, tc.apimanager)

			backendReconciler := NewBackendReconciler(baseAPIManagerLogicReconciler)
			_, err = backendReconciler.Reconcile()
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
			_, err = backendReconciler.Reconcile()
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

func backendApiManagerCreator(disableSyncAnnotation string, disableSyncAnnotationValue string) *appsv1alpha1.APIManager {
	var (
		name                 = "example-apimanager"
		namespace            = "operator-unittest"
		wildcardDomain       = "test.3scale.net"
		appLabel             = "someLabel"
		tenantName           = "someTenant"
		trueValue            = true
		oneValue       int64 = 1
	)

	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{disableSyncAnnotation: disableSyncAnnotationValue},
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				WildcardDomain:               wildcardDomain,
				TenantName:                   &tenantName,
				ResourceRequirementsEnabled:  &trueValue,
			},
			Backend: &appsv1alpha1.BackendSpec{
				ListenerSpec: &appsv1alpha1.BackendListenerSpec{Replicas: &oneValue},
				WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{Replicas: &oneValue},
				CronSpec:     &appsv1alpha1.BackendCronSpec{Replicas: &oneValue},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
}

func confirmReplicasWhenAnnotationIsNotPresent(apiManager *appsv1alpha1.APIManager, dc *appsv1.DeploymentConfig, annotation string, annotationValue string, expectedValue int32) bool {
	if !metav1.HasAnnotation(apiManager.ObjectMeta, annotation) {
		if dc.Spec.Replicas != expectedValue {
			return false
		}
	}

	return true
}

func confirmReplicasWhenAnnotationPresent(apiManager *appsv1alpha1.APIManager, dc *appsv1.DeploymentConfig, annotation string, annotationValue string, expectedValue int32) bool {
	if metav1.HasAnnotation(apiManager.ObjectMeta, annotation) && apiManager.Annotations[annotation] == annotationValue {
		if dc.Spec.Replicas != expectedValue {
			return false
		}
	}

	return true
}
