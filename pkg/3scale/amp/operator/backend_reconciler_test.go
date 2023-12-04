package operator

import (
	"context"
	k8sappsv1 "k8s.io/api/apps/v1"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
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

	backendRedisSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-redis",
			Namespace: namespace,
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager, backendRedisSecret}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.Install(s)
	if err != nil {
		t.Fatal(err)
	}
	err = imagev1.Install(s)
	if err != nil {
		t.Fatal(err)
	}
	err = routev1.Install(s)
	if err != nil {
		t.Fatal(err)
	}
	err = configv1.Install(s)
	if err != nil {
		t.Fatal(err)
	}

	// 3scale 2.14 -> 2.15
	err = appsv1.Install(s)
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
		obj      k8sclient.Object
	}{
		{"cronDeployment", "backend-cron", &k8sappsv1.Deployment{}},
		{"listenerDeployment", "backend-listener", &k8sappsv1.Deployment{}},
		{"listenerService", "backend-listener", &v1.Service{}},
		{"listenerRoute", "backend", &routev1.Route{}},
		{"workerDeployment", "backend-worker", &k8sappsv1.Deployment{}},
		{"environmentCM", "backend-environment", &v1.ConfigMap{}},
		{"internalAPISecret", component.BackendSecretInternalApiSecretName, &v1.Secret{}},
		{"listenerSecret", component.BackendSecretBackendListenerSecretName, &v1.Secret{}},
		{"workerPDB", "backend-worker", &policyv1.PodDisruptionBudget{}},
		{"cronPDB", "backend-cron", &policyv1.PodDisruptionBudget{}},
		{"listenerPDB", "backend-listener", &policyv1.PodDisruptionBudget{}},
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

func TestReplicaBackendReconciler(t *testing.T) {
	var (
		namespace        = "operator-unittest"
		log              = logf.Log.WithName("operator_test")
		twoValue   int32 = 2
		oneValue   int32 = 1
		oneValue64 int64 = 1
	)
	ctx := context.TODO()
	s := scheme.Scheme

	err := appsv1alpha1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = appsv1.Install(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := configv1.Install(s); err != nil {
		t.Fatal(err)
	}
	err = routev1.Install(s)
	if err != nil {
		t.Fatal(err)
	}

	backendRedisSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-redis",
			Namespace: namespace,
		},
	}

	cases := []struct {
		testName                 string
		objName                  string
		apimanager               *appsv1alpha1.APIManager
		expectedAmountOfReplicas int32
	}{
		{"cron replicas set", "backend-cron", backendApiManagerCreator(nil, &oneValue64, nil), oneValue},
		{"cron replicas not set", "backend-cron", backendApiManagerCreator(nil, nil, nil), twoValue},

		{"listener replicas set", "backend-listener", backendApiManagerCreator(&oneValue64, nil, nil), oneValue},
		{"listener replicas not set", "backend-listener", backendApiManagerCreator(nil, nil, nil), twoValue},

		{"worker replicas set", "backend-worker", backendApiManagerCreator(nil, nil, &oneValue64), oneValue},
		{"worker replicas not set", "backend-worker", backendApiManagerCreator(nil, nil, nil), twoValue},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{tc.apimanager, backendRedisSecret}
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

			deployment := &k8sappsv1.Deployment{}
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, deployment)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			// bump the amount of replicas in the deployment
			deployment.Spec.Replicas = &twoValue
			err = cl.Update(context.TODO(), deployment)
			if err != nil {
				subT.Errorf("error updating deployment of %s: %v", tc.objName, err)
			}

			// re-run the reconciler
			_, err = backendReconciler.Reconcile()
			if err != nil {
				t.Fatal(err)
			}

			err = cl.Get(context.TODO(), namespacedName, deployment)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			if tc.expectedAmountOfReplicas != *deployment.Spec.Replicas {
				subT.Errorf("expected replicas do not match. expected: %d actual: %d", tc.expectedAmountOfReplicas, deployment.Spec.Replicas)
			}
		})
	}
}

func backendApiManagerCreator(listenerReplicas, cronReplicas, workerReplicas *int64) *appsv1alpha1.APIManager {
	var (
		name           = "example-apimanager"
		namespace      = "operator-unittest"
		wildcardDomain = "test.3scale.net"
		appLabel       = "someLabel"
		tenantName     = "someTenant"
		trueValue      = true
	)

	return &appsv1alpha1.APIManager{
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
				ListenerSpec: &appsv1alpha1.BackendListenerSpec{Replicas: listenerReplicas},
				WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{Replicas: workerReplicas},
				CronSpec:     &appsv1alpha1.BackendCronSpec{Replicas: cronReplicas},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
}
