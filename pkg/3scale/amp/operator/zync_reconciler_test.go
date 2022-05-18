package operator

import (
	"context"
	"testing"

	"k8s.io/api/policy/v1beta1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
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
		obj      runtime.Object
	}{
		{"queRole", "zync-que-role", &rbacv1.Role{}},
		{"queServiceAccount", "zync-que-sa", &v1.ServiceAccount{}},
		{"queRoleBinding", "zync-que-rolebinding", &rbacv1.RoleBinding{}},
		{"zyncDC", "zync", &appsv1.DeploymentConfig{}},
		{"zyncQueDC", "zync-que", &appsv1.DeploymentConfig{}},
		{"zyncDatabaseDC", "zync-database", &appsv1.DeploymentConfig{}},
		{"zyncService", "zync", &v1.Service{}},
		{"zyncDatabaseService", "zync-database", &v1.Service{}},
		{"zyncSecret", component.ZyncSecretName, &v1.Secret{}},
		{"zyncPDB", "zync", &v1beta1.PodDisruptionBudget{}},
		{"quePDB", "zync-que", &v1beta1.PodDisruptionBudget{}},
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
		obj        runtime.Object
		hasToExist bool
	}{
		{"queRole", "zync-que-role", &rbacv1.Role{}, true},
		{"queServiceAccount", "zync-que-sa", &v1.ServiceAccount{}, true},
		{"queRoleBinding", "zync-que-rolebinding", &rbacv1.RoleBinding{}, true},
		{"zyncDC", "zync", &appsv1.DeploymentConfig{}, true},
		{"zyncQueDC", "zync-que", &appsv1.DeploymentConfig{}, true},
		{"zyncDatabaseDC", "zync-database", &appsv1.DeploymentConfig{}, false},
		{"zyncService", "zync", &v1.Service{}, true},
		{"zyncDatabaseService", "zync-database", &v1.Service{}, false},
		{"zyncSecret", component.ZyncSecretName, &v1.Secret{}, true},
		{"zyncPDB", "zync", &v1beta1.PodDisruptionBudget{}, true},
		{"quePDB", "zync-que", &v1beta1.PodDisruptionBudget{}, true},
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

func TestZyncReconcilerDisableReplicaSyncingAnnotations(t *testing.T) {
	var (
		namespace                  = "operator-unittest"
		log                        = logf.Log.WithName("operator_test")
		twoValue             int32 = 2
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

	cases := []struct {
		testName string
		objName  string
		obj      runtime.Object
		apimanager *appsv1alpha1.APIManager
		annotation string
		annotationValue string
		expectedAmountOfReplicas int32
		validatingFunction func(*appsv1alpha1.APIManager, *appsv1.DeploymentConfig, string, string, int32) bool
	}{
		{"zyncQueDC-annotation not present", "zync-que", &appsv1.DeploymentConfig{}, zyncApiManagerCreator("someAnnotation", "false"), disableZyncQueInstancesSyncing, "dummy", int32(1), confirmReplicasWhenAnnotationIsNotPresent},
		{"zyncQueDC-annotation false", "zync-que", &appsv1.DeploymentConfig{}, zyncApiManagerCreator(disableZyncQueInstancesSyncing, "false"), disableZyncQueInstancesSyncing, "false", int32(1), confirmReplicasWhenAnnotationPresent},
		{"zyncQueDC-annotation true", "zync-que", &appsv1.DeploymentConfig{}, zyncApiManagerCreator(disableZyncQueInstancesSyncing, "true"), disableZyncQueInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"zyncQueDC-annotation true of dummy value", "zync-que", &appsv1.DeploymentConfig{}, zyncApiManagerCreator(disableZyncQueInstancesSyncing, "true"), disableZyncQueInstancesSyncing, "someDummyValue", int32(1), confirmReplicasWhenAnnotationPresent},		
		
		{"zyncDC-annotation not present", "zync", &appsv1.DeploymentConfig{}, zyncApiManagerCreator("someAnnotation", "false"), disableZyncInstancesSyncing, "dummy", int32(1), confirmReplicasWhenAnnotationIsNotPresent},
		{"zyncDC-annotation false", "zync", &appsv1.DeploymentConfig{}, zyncApiManagerCreator(disableZyncInstancesSyncing, "false"), disableZyncInstancesSyncing, "false", int32(1), confirmReplicasWhenAnnotationPresent},
		{"zyncDC-annotation true", "zync", &appsv1.DeploymentConfig{}, zyncApiManagerCreator(disableZyncInstancesSyncing, "true"), disableZyncInstancesSyncing, "true", int32(2), confirmReplicasWhenAnnotationPresent},
		{"zyncDC-annotation true of dummy value", "zync", &appsv1.DeploymentConfig{}, zyncApiManagerCreator(disableZyncInstancesSyncing, "true"), disableZyncInstancesSyncing, "someDummyValue", int32(1), confirmReplicasWhenAnnotationPresent},
		
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			zyncExternalDatabaseSecret := getZyncSecretExternalDatabase(namespace)
			// Objects to track in the fake client.
			objs := []runtime.Object{tc.apimanager, zyncExternalDatabaseSecret}
			// Create a fake client to mock API calls.
			cl := fake.NewFakeClient(objs...)
			clientAPIReader := fake.NewFakeClient(objs...)
			clientset := fakeclientset.NewSimpleClientset()
			recorder := record.NewFakeRecorder(10000)
			baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
			baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, tc.apimanager)

			zyncReconciler := NewZyncReconciler(baseAPIManagerLogicReconciler)
			_, err = zyncReconciler.Reconcile()
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
			_, err = zyncReconciler.Reconcile()
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

func zyncApiManagerCreator(disableSyncAnnotation string, disableSyncAnnotationValue string) *appsv1alpha1.APIManager {
	var (
		name                       = "example-apimanager"
		namespace                  = "operator-unittest"
		wildcardDomain             = "test.3scale.net"
		appLabel                   = "someLabel"
		tenantName                 = "someTenant"
		trueValue                  = true
		oneValue             int64 = 1
	)

	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
			Zync: &appsv1alpha1.ZyncSpec{
				AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: &oneValue},
				QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: &oneValue},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
			HighAvailability: &appsv1alpha1.HighAvailabilitySpec{
				Enabled:                     true,
				ExternalZyncDatabaseEnabled: &trueValue,
			},
		},
	}
}