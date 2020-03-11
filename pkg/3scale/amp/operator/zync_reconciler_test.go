package operator

import (
	"context"
	"testing"

	"k8s.io/api/policy/v1beta1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Unable to get config: (%v)", err)
	}
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
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	BaseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)

	zyncReconciler := NewZyncReconciler(BaseAPIManagerLogicReconciler)
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
