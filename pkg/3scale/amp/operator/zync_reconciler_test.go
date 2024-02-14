package operator

import (
	"context"
	k8sappsv1 "k8s.io/api/apps/v1"
	"testing"

	policyv1 "k8s.io/api/policy/v1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestNewZyncReconciler(t *testing.T) {
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
			Zync: &appsv1alpha1.ZyncSpec{
				AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: &oneValue},
				QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: &oneValue},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := k8sappsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = configv1.Install(s)
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
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
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
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	zyncReconciler := NewZyncReconciler(baseAPIManagerLogicReconciler)
	_, err = zyncReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      client.Object
	}{
		{"queRole", "zync-que-role", &rbacv1.Role{}},
		{"queServiceAccount", "zync-que-sa", &v1.ServiceAccount{}},
		{"queRoleBinding", "zync-que-rolebinding", &rbacv1.RoleBinding{}},
		{"zyncDeployment", "zync", &k8sappsv1.Deployment{}},
		{"zyncQueDeployment", "zync-que", &k8sappsv1.Deployment{}},
		{"zyncDatabaseDeployment", "zync-database", &k8sappsv1.Deployment{}},
		{"zyncService", "zync", &v1.Service{}},
		{"zyncDatabaseService", "zync-database", &v1.Service{}},
		{"zyncSecret", component.ZyncSecretName, &v1.Secret{}},
		{"zyncPDB", "zync", &policyv1.PodDisruptionBudget{}},
		{"quePDB", "zync-que", &policyv1.PodDisruptionBudget{}},
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

func TestNewZyncReconcilerWithAllExternalDatabases(t *testing.T) {
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
			Zync: &appsv1alpha1.ZyncSpec{
				AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: &oneValue},
				QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: &oneValue},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
			HighAvailability: &appsv1alpha1.HighAvailabilitySpec{
				Enabled:                     true,
				ExternalZyncDatabaseEnabled: &trueValue,
			},
			ExternalComponents: appsv1alpha1.AllComponentsExternal(),
		},
	}

	zyncExternalDatabaseSecret := getZyncSecretExternalDatabase(namespace)

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager, zyncExternalDatabaseSecret}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := k8sappsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = configv1.Install(s)
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
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
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
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	zyncReconciler := NewZyncReconciler(baseAPIManagerLogicReconciler)
	_, err = zyncReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName   string
		objName    string
		obj        client.Object
		hasToExist bool
	}{
		{"queRole", "zync-que-role", &rbacv1.Role{}, true},
		{"queServiceAccount", "zync-que-sa", &v1.ServiceAccount{}, true},
		{"queRoleBinding", "zync-que-rolebinding", &rbacv1.RoleBinding{}, true},
		{"zyncDeployment", "zync", &k8sappsv1.Deployment{}, true},
		{"zyncQueDeployment", "zync-que", &k8sappsv1.Deployment{}, true},
		{"zyncDatabaseDeployment", "zync-database", &k8sappsv1.Deployment{}, false},
		{"zyncService", "zync", &v1.Service{}, true},
		{"zyncDatabaseService", "zync-database", &v1.Service{}, false},
		{"zyncSecret", component.ZyncSecretName, &v1.Secret{}, true},
		{"zyncPDB", "zync", &policyv1.PodDisruptionBudget{}, true},
		{"quePDB", "zync-que", &policyv1.PodDisruptionBudget{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			if tc.hasToExist {
				if err != nil {
					subT.Errorf("error fetching object %s: %v", tc.objName, err)
				}
			} else {
				if err == nil || !errors.IsNotFound(err) {
					subT.Errorf("object %s that shouldn't exist exists or different error than NotFound returned: %v", tc.objName, err)
				}
			}
		})
	}
}

func TestReplicaZyncReconciler(t *testing.T) {
	var (
		namespace        = "operator-unittest"
		log              = logf.Log.WithName("operator_test")
		oneValue   int32 = 1
		oneValue64 int64 = 1
		twoValue   int32 = 2
	)
	ctx := context.TODO()
	s := scheme.Scheme

	err := appsv1alpha1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = k8sappsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	if err := configv1.Install(s); err != nil {
		t.Fatal(err)
	}

	// 3scale 2.14 -> 2.15
	err = appsv1.Install(s)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName                 string
		objName                  string
		apimanager               *appsv1alpha1.APIManager
		expectedAmountOfReplicas int32
	}{
		{"zync replicas set", "zync", testZyncAPIManagerCreator(&oneValue64, nil), oneValue},
		{"zync replicas not set", "zync", testZyncAPIManagerCreator(nil, nil), twoValue},

		{"zync-que replicas set", "zync-que", testZyncAPIManagerCreator(nil, &oneValue64), oneValue},
		{"zync-que replicas not set", "zync-que", testZyncAPIManagerCreator(nil, nil), twoValue},
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

			reconciler := NewZyncReconciler(baseAPIManagerLogicReconciler)
			_, err = reconciler.Reconcile()
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

			// bump the amount of replicas in the dc
			deployment.Spec.Replicas = &twoValue
			err = cl.Update(context.TODO(), deployment)
			if err != nil {
				subT.Errorf("error updating dc of %s: %v", tc.objName, err)
			}

			// re-run the reconciler
			_, err = reconciler.Reconcile()
			if err != nil {
				t.Fatal(err)
			}

			err = cl.Get(context.TODO(), namespacedName, deployment)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			if tc.expectedAmountOfReplicas != *deployment.Spec.Replicas {
				subT.Errorf("expected replicas do not match. expected: %d actual: %d", tc.expectedAmountOfReplicas, *deployment.Spec.Replicas)
			}
		})
	}
}

func testZyncAPIManagerCreator(zyncReplicas, zyncQueReplicas *int64) *appsv1alpha1.APIManager {
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
			Zync: &appsv1alpha1.ZyncSpec{
				AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: zyncReplicas},
				QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: zyncQueReplicas},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
}
