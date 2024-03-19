package operator

import (
	"context"
	"fmt"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestHighAvailabilityReconciler(t *testing.T) {
	var (
		log                     = logf.Log.WithName("operator_test")
		systemMysqlRootPassword = "rootPassw1"
		systemMysqlDatabaseName = "myDatabaseName"
		databaseURL             = fmt.Sprintf("mysql2://root:%s@system-mysql/%s", systemMysqlRootPassword, systemMysqlDatabaseName)
	)

	ctx := context.TODO()

	apimanager := basicApimanagerTestHA()
	backendRedisSecret := testBackendRedisSecret()
	systemRedisSecret := testSystemRedisSecret()
	systemDBSecret := getSystemDBSecret(databaseURL)
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager, backendRedisSecret, systemRedisSecret, systemDBSecret}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
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

	cases := []struct {
		reconcilerConstructor DependencyReconcilerConstructor
		testName              string
		secretName            string
	}{
		{NewBackendExternalRedisReconciler, "backendRedisTest", component.BackendSecretBackendRedisSecretName},
		{NewSystemExternalRedisReconciler, "systemRedisTest", component.SystemSecretSystemRedisSecretName},
		{NewSystemExternalDatabaseReconciler, "systemDatabaseTest", component.SystemSecretSystemDatabaseSecretName},
	}
	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			reconciler := tc.reconcilerConstructor(BaseAPIManagerLogicReconciler)
			_, err = reconciler.Reconcile()
			if err != nil {
				subT.Fatal(err)
			}

			secret := &v1.Secret{}
			namespacedName := types.NamespacedName{
				Name:      tc.secretName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, secret)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.secretName, err)
			}

		})
	}
}
