package operator

import (
	"context"
	"testing"

	"github.com/3scale/3scale-operator/pkg/common"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

func TestApicastReconciler(t *testing.T) {
	var (
		name                       = "example-apimanager"
		namespace                  = "operator-unittest"
		wildcardDomain             = "test.3scale.net"
		log                        = logf.Log.WithName("operator_test")
		appLabel                   = "someLabel"
		tenantName                 = "someTenant"
		trueValue                  = true
		apicastManagementAPI       = "disabled"
		oneValue             int64 = 1
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
			Apicast: &appsv1alpha1.ApicastSpec{
				ApicastManagementAPI: &apicastManagementAPI,
				OpenSSLVerify:        &trueValue,
				IncludeResponseCodes: &trueValue,
				StagingSpec: &appsv1alpha1.ApicastStagingSpec{
					Replicas: &oneValue,
				},
				ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
					Replicas: &oneValue,
				},
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

	apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
	_, err = apicastReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      runtime.Object
	}{
		{"stagingDeployment", "apicast-staging", &appsv1.DeploymentConfig{}},
		{"productionDeployment", "apicast-production", &appsv1.DeploymentConfig{}},
		{"stagingService", "apicast-staging", &v1.Service{}},
		{"productionService", "apicast-production", &v1.Service{}},
		{"envConfigMap", "apicast-environment", &v1.ConfigMap{}},
		{"stagingPDB", "apicast-staging", &policyv1.PodDisruptionBudget{}},
		{"productionPDB", "apicast-production", &policyv1.PodDisruptionBudget{}},
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

func TestApicastReconcilerCustomPolicyParts(t *testing.T) {
	var (
		name                       = "example-apimanager"
		namespace                  = "operator-unittest"
		wildcardDomain             = "test.3scale.net"
		log                        = logf.Log.WithName("operator_test")
		appLabel                   = "someLabel"
		tenantName                 = "someTenant"
		apicastManagementAPI       = "disabled"
		trueValue                  = true
		oneValue             int64 = 1

		p1CustomPolicy = component.CustomPolicy{
			Name:      "P1",
			Version:   "0.1.0",
			SecretRef: v1.LocalObjectReference{Name: "someSecretP1"},
		}

		p2CustomPolicy = component.CustomPolicy{
			Name:      "P2",
			Version:   "0.1.0",
			SecretRef: v1.LocalObjectReference{Name: "someSecretP2"},
		}
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
			Apicast: &appsv1alpha1.ApicastSpec{
				ApicastManagementAPI: &apicastManagementAPI,
				OpenSSLVerify:        &trueValue,
				IncludeResponseCodes: &trueValue,
				StagingSpec: &appsv1alpha1.ApicastStagingSpec{
					Replicas: &oneValue,
				},
				ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
					Replicas: &oneValue,
					CustomPolicies: []appsv1alpha1.CustomPolicySpec{
						{
							Name:      p2CustomPolicy.Name,
							Version:   p2CustomPolicy.Version,
							SecretRef: &p2CustomPolicy.SecretRef,
						},
					},
				},
			},
		},
	}

	// Existing DC has 1 custom policy defined: P1
	// Desired DC has 1 custom policy defined: P2
	// P2 should be added to existing DC
	// P1 should be deleted from existing DC
	apicastOptions := &component.ApicastOptions{
		ProductionCustomPolicies: []component.CustomPolicy{p1CustomPolicy},
		StagingTracingConfig:     &component.APIcastTracingConfig{},
		ProductionTracingConfig:  &component.APIcastTracingConfig{},
	}
	apicast := component.NewApicast(apicastOptions)
	existingProdDC := apicast.ProductionDeploymentConfig()
	existingProdDC.Namespace = namespace

	// - Policy annotation for P1 added
	p1Found := false
	for key := range existingProdDC.Annotations {
		if p1CustomPolicy.AnnotationKey() == key {
			p1Found = true
		}
	}

	if !p1Found {
		t.Fatal("P1 policy annotation not found. Should have been created")
	}

	p2Secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: p2CustomPolicy.SecretRef.Name, Namespace: namespace},
		Data: map[string][]byte{
			"init.lua":            []byte("some lua code"),
			"apicast-policy.json": []byte("{}"),
		},
		Type: v1.SecretTypeOpaque,
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager, existingProdDC, p2Secret}
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

	apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
	_, err = apicastReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "apicast-production",
		Namespace: namespace,
	}
	existing := &appsv1.DeploymentConfig{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	// Assert existing DC:
	// - Volume for P1 deleted
	for idx := range existing.Spec.Template.Spec.Volumes {
		if existing.Spec.Template.Spec.Volumes[idx].Name == p1CustomPolicy.VolumeName() {
			t.Fatal("P1 volume found. Should have been deleted")
		}
	}
	// - VolumeMount for P1 deleted
	for idx := range existing.Spec.Template.Spec.Containers[0].VolumeMounts {
		if existing.Spec.Template.Spec.Containers[0].VolumeMounts[idx].Name == p1CustomPolicy.VolumeName() {
			t.Fatal("P1 volumemount found. Should have been deleted")
		}
	}

	// - Volume for P2 added
	p2Found := false
	for idx := range existing.Spec.Template.Spec.Volumes {
		if existing.Spec.Template.Spec.Volumes[idx].Name == p2CustomPolicy.VolumeName() {
			p2Found = true
		}
	}

	if !p2Found {
		t.Fatal("P2 volume not found. Should have been created")
	}

	// - VolumeMount for P2 added
	p2Found = false
	for idx := range existing.Spec.Template.Spec.Containers[0].VolumeMounts {
		if existing.Spec.Template.Spec.Containers[0].VolumeMounts[idx].Name == p2CustomPolicy.VolumeName() {
			p2Found = true
		}
	}

	if !p2Found {
		t.Fatal("P2 volumemount not found. Should have been created")
	}

	// - Policy annotation for P1 deleted
	for key := range existing.Annotations {
		if p1CustomPolicy.AnnotationKey() == key {
			t.Fatal("P1 annotation found. Should have been deleted")
		}
	}

	// - Policy annotation for P2 added
	p2Found = false
	for key := range existing.Annotations {
		if p2CustomPolicy.AnnotationKey() == key {
			p2Found = true
		}
	}

	if !p2Found {
		t.Fatal("P2 policy annotation not found. Should have been created")
	}
}

func TestApicastReconcilerTracingConfigParts(t *testing.T) {
	var (
		name                                   = "example-apimanager"
		namespace                              = "operator-unittest"
		wildcardDomain                         = "test.3scale.net"
		log                                    = logf.Log.WithName("operator_test")
		appLabel                               = "someLabel"
		tenantName                             = "someTenant"
		apicastManagementAPI                   = "disabled"
		trueValue                              = true
		falseValue                             = false
		oneValue                         int64 = 1
		existingTracingConfig1SecretName       = "mysecretnameone"
		desiredTracingConfig1SecretName        = "mysecretnametwo"

		existingTracingConfig1 = component.APIcastTracingConfig{
			TracingLibrary:          component.APIcastDefaultTracingLibrary,
			Enabled:                 true,
			TracingConfigSecretName: &existingTracingConfig1SecretName,
		}

		desiredTracingConfig1 = component.APIcastTracingConfig{
			TracingLibrary:          component.APIcastDefaultTracingLibrary,
			Enabled:                 true,
			TracingConfigSecretName: &desiredTracingConfig1SecretName,
		}
	)

	apicastOptions := &component.ApicastOptions{
		StagingTracingConfig:    &component.APIcastTracingConfig{},
		ProductionTracingConfig: &existingTracingConfig1,
	}
	apicast := component.NewApicast(apicastOptions)
	existingProdDC := apicast.ProductionDeploymentConfig()
	existingProdDC.Namespace = namespace

	// - Tracing Configuration 1 added into the Production DC with the expected key
	existingTracingConfig1Found := false
	for key := range existingProdDC.Annotations {
		if existingTracingConfig1.AnnotationKey() == key {
			existingTracingConfig1Found = true
		}
	}

	if !existingTracingConfig1Found {
		t.Fatal("tracing config 1 annotation not found. Should have been created")
	}

	existingTc1Secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: *existingTracingConfig1.TracingConfigSecretName, Namespace: namespace},
		Data: map[string][]byte{
			"config": []byte("some existing tracing config"),
		},
		Type: v1.SecretTypeOpaque,
	}

	desiredTc1Secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: *desiredTracingConfig1.TracingConfigSecretName, Namespace: namespace},
		Data: map[string][]byte{
			"config": []byte("some desired tracing config"),
		},
		Type: v1.SecretTypeOpaque,
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
			Apicast: &appsv1alpha1.ApicastSpec{
				ApicastManagementAPI: &apicastManagementAPI,
				OpenSSLVerify:        &trueValue,
				IncludeResponseCodes: &trueValue,
				StagingSpec: &appsv1alpha1.ApicastStagingSpec{
					Replicas: &oneValue,
					OpenTracing: &appsv1alpha1.APIcastOpenTracingSpec{
						Enabled: &falseValue,
						TracingConfigSecretRef: &v1.LocalObjectReference{
							Name: "anothersecret",
						},
					},
				},
				ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
					Replicas: &oneValue,
					OpenTracing: &appsv1alpha1.APIcastOpenTracingSpec{
						Enabled: &trueValue,
						TracingConfigSecretRef: &v1.LocalObjectReference{
							Name: desiredTracingConfig1SecretName,
						},
					},
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager, existingProdDC, existingTc1Secret, desiredTc1Secret}
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

	ctx := context.TODO()
	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
	_, err = apicastReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "apicast-production",
		Namespace: namespace,
	}
	existing := &appsv1.DeploymentConfig{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	// // Assert existing DC:
	// // - Volume for existingTracingConfig1 deleted
	for idx := range existing.Spec.Template.Spec.Volumes {
		if existing.Spec.Template.Spec.Volumes[idx].Name == existingTracingConfig1.VolumeName() {
			t.Fatal("existingTracingConfig1 volume found. Should have been deleted")
		}
	}
	// // - VolumeMount for existingTracingConfig1 deleted
	for idx := range existing.Spec.Template.Spec.Containers[0].VolumeMounts {
		if existing.Spec.Template.Spec.Containers[0].VolumeMounts[idx].Name == existingTracingConfig1.VolumeName() {
			t.Fatal("existingTracingConfig1 volumemount found. Should have been deleted")
		}
	}

	// // - Volume for desiredTracingConfig1 added
	desiredTracingConfig1Found := false
	for idx := range existing.Spec.Template.Spec.Volumes {
		if existing.Spec.Template.Spec.Volumes[idx].Name == desiredTracingConfig1.VolumeName() {
			desiredTracingConfig1Found = true
		}
	}

	if !desiredTracingConfig1Found {
		t.Fatal("desiredTracingConfig1 volume not found. Should have been created")
	}

	// // - VolumeMount for desiredTracingConfig1 added
	desiredTracingConfig1Found = false
	for idx := range existing.Spec.Template.Spec.Containers[0].VolumeMounts {
		if existing.Spec.Template.Spec.Containers[0].VolumeMounts[idx].Name == desiredTracingConfig1.VolumeName() {
			desiredTracingConfig1Found = true
		}
	}

	if !desiredTracingConfig1Found {
		t.Fatal("desiredTracingConfig1 volumemount not found. Should have been created")
	}

	// // - Tracing config annotation for existingTracingConfig1 deleted
	for key := range existing.Annotations {
		if existingTracingConfig1.AnnotationKey() == key {
			t.Fatal("existingTracingConfig1 annotation found. Should have been deleted")
		}
	}

	// // - Tracing config annotation for desiredTracingConfig1 added
	desiredTracingConfig1Found = false
	for key := range existing.Annotations {
		if desiredTracingConfig1.AnnotationKey() == key {
			desiredTracingConfig1Found = true
		}
	}

	if !desiredTracingConfig1Found {
		t.Fatal("desiredTracingConfig1 tracing config annotation not found. Should have been created")
	}
}
