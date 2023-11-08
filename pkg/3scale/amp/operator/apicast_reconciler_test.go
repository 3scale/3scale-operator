package operator

import (
	"context"
	k8sappsv1 "k8s.io/api/apps/v1"
	"testing"

	"github.com/3scale/3scale-operator/pkg/common"
	"k8s.io/apimachinery/pkg/util/intstr"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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

	"github.com/3scale/3scale-operator/apis/apps"
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

	apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
	_, err = apicastReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      k8sclient.Object
	}{
		{"stagingDeployment", "apicast-staging", &k8sappsv1.Deployment{}},
		{"productionDeployment", "apicast-production", &k8sappsv1.Deployment{}},
		{"stagingService", "apicast-staging", &v1.Service{}},
		{"productionService", "apicast-production", &v1.Service{}},
		{"envConfigMap", "apicast-environment", &v1.ConfigMap{}},
		{"stagingPDB", "apicast-staging", &policyv1.PodDisruptionBudget{}},
		{"productionPDB", "apicast-production", &policyv1.PodDisruptionBudget{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, tc.obj)
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

		p1Secret = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "someSecretP1",
				Namespace: namespace, // Update this with the appropriate namespace
			},
			Data: map[string][]byte{
				"apicast-policy.json": []byte("testApicastPolicy"),
				"example.lua":         []byte("testExampleLua"),
				"init.lua":            []byte("testInitLua"),
			},
			Type: v1.SecretTypeOpaque,
		}

		p1CustomPolicy = component.CustomPolicy{
			Name:    "P1",
			Version: "0.1.0",
			Secret:  p1Secret,
		}

		p2CustomPolicy = component.CustomPolicy{
			Name:    "P2",
			Version: "0.1.0",
			Secret:  p1Secret,
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
							SecretRef: &v1.LocalObjectReference{Name: "someSecretP1"},
						},
					},
				},
			},
		},
	}

	// Existing Deployment has 1 custom policy defined: P1
	// Desired Deployment has 1 custom policy defined: P2
	// P2 should be added to existing Deployment
	// P1 should be deleted from existing Deployment
	apicastOptions := &component.ApicastOptions{

		ProductionCustomPolicies: []component.CustomPolicy{p1CustomPolicy},
		StagingTracingConfig:     &component.APIcastTracingConfig{},
		ProductionTracingConfig:  &component.APIcastTracingConfig{},
	}
	apicast := component.NewApicast(apicastOptions)
	existingProdDeployment := apicast.ProductionDeployment()
	existingProdDeployment.Namespace = namespace

	// - Policy annotation for P1 added
	p1Found := false
	for key := range existingProdDeployment.Annotations {
		if p1CustomPolicy.AnnotationKey() == key {
			p1Found = true
		}
	}

	if !p1Found {
		t.Fatal("P1 policy annotation not found. Should have been created")
	}

	p2Secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: p2CustomPolicy.Secret.Name, Namespace: namespace},
		Data: map[string][]byte{
			"init.lua":            []byte("some lua code"),
			"apicast-policy.json": []byte("{}"),
		},
		Type: v1.SecretTypeOpaque,
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager, existingProdDeployment, p2Secret}
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

	apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
	_, err = apicastReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "apicast-production",
		Namespace: namespace,
	}
	existing := &k8sappsv1.Deployment{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	// Assert existing Deployment:
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
			TracingLibrary:          apps.APIcastDefaultTracingLibrary,
			Enabled:                 true,
			TracingConfigSecretName: &existingTracingConfig1SecretName,
		}

		desiredTracingConfig1 = component.APIcastTracingConfig{
			TracingLibrary:          apps.APIcastDefaultTracingLibrary,
			Enabled:                 true,
			TracingConfigSecretName: &desiredTracingConfig1SecretName,
		}
	)

	apicastOptions := &component.ApicastOptions{
		StagingTracingConfig:    &component.APIcastTracingConfig{},
		ProductionTracingConfig: &existingTracingConfig1,
	}
	apicast := component.NewApicast(apicastOptions)
	existingProdDeployment := apicast.ProductionDeployment()
	existingProdDeployment.Namespace = namespace

	// - Tracing Configuration 1 added into the Production Deployment with the expected key
	existingTracingConfig1Found := false
	for key := range existingProdDeployment.Annotations {
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
	objs := []runtime.Object{apimanager, existingProdDeployment, existingTc1Secret, desiredTc1Secret}
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
	existing := &k8sappsv1.Deployment{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	// // Assert existing Deployment:
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

func TestApicastServicePortMutator(t *testing.T) {
	var (
		name                       = "example-apimanager"
		namespace                  = "operator-unittest"
		port                 int32 = 1111
		targetPort                 = intstr.FromInt(1111)
		wildcardDomain             = "test.3scale.net"
		log                        = logf.Log.WithName("operator_test")
		appLabel                   = "someLabel"
		tenantName                 = "someTenant"
		apicastManagementAPI       = "disabled"
		trueValue                  = true
		oneValue             int64 = 1
	)

	ctx := context.TODO()
	existingStagingService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apicast-staging",
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: port, TargetPort: targetPort},
			},
		},
	}
	existingProductionService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apicast-production",
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: port, TargetPort: targetPort},
			},
		},
	}

	productionServiceKey := common.ObjectKey(existingProductionService)
	stagingServiceKey := common.ObjectKey(existingStagingService)

	cases := []struct {
		testName           string
		apim               *appsv1alpha1.APIManager
		expectedPort       int32
		expectedTargetPort intstr.IntOrString
	}{
		{
			"annotationTrue",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   namespace,
					Annotations: map[string]string{"apps.3scale.net/disable-apicast-service-reconciler": "true"},
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
				},
			},
			port,
			targetPort,
		},
		{
			"annotationFalse",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   namespace,
					Annotations: map[string]string{"apps.3scale.net/disable-apicast-service-reconciler": "false"},
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
				},
			},
			8080,
			intstr.FromInt(8080),
		},
		{
			"annotationAbsent",
			&appsv1alpha1.APIManager{
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
				},
			},
			8080,
			intstr.FromInt(8080),
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			// Objects to track in the fake client.
			objs := []runtime.Object{tc.apim, existingStagingService, existingProductionService}
			s := scheme.Scheme
			s.AddKnownTypes(appsv1alpha1.GroupVersion, tc.apim)
			if err := appsv1.AddToScheme(s); err != nil {
				t.Fatal(err)
			}
			if err := imagev1.AddToScheme(s); err != nil {
				t.Fatal(err)
			}
			if err := routev1.AddToScheme(s); err != nil {
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
			baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, tc.apim)

			apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
			if _, err := apicastReconciler.Reconcile(); err != nil {
				t.Fatal(err)
			}
			prodSvc := &v1.Service{}
			stageSvc := &v1.Service{}

			// Fetch both services with fake client
			if err := cl.Get(context.TODO(), productionServiceKey, prodSvc); err != nil {
				t.Fatal(err)
			}
			if err := cl.Get(context.TODO(), stagingServiceKey, stageSvc); err != nil {
				t.Fatal(err)
			}

			if stageSvc.Spec.Ports[0].Port != tc.expectedPort || stageSvc.Spec.Ports[0].TargetPort != tc.expectedTargetPort {
				t.Fatal("Apicast service Ports do not match the expected service port based on the annotation")
			}
			if prodSvc.Spec.Ports[0].Port != tc.expectedPort || prodSvc.Spec.Ports[0].TargetPort != tc.expectedTargetPort {
				t.Fatal("Apicast service Target Ports do not match the expected service targetPort based on the annotation")
			}
		})
	}
}

func TestReplicaApicastReconciler(t *testing.T) {
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
		apimanager               *appsv1alpha1.APIManager
		expectedAmountOfReplicas int32
	}{
		{"apicast-staging replicas set", "apicast-staging", testApicastAPIManagerCreator(&oneValue64, nil), oneValue},
		{"apicast-staging replicas not set", "apicast-staging", testApicastAPIManagerCreator(nil, nil), twoValue},

		{"apicast-production replicas set", "apicast-production", testApicastAPIManagerCreator(nil, &oneValue64), oneValue},
		{"apicast-production replicas not set", "apicast-production", testApicastAPIManagerCreator(nil, nil), twoValue},
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

			apicastReconciler := NewApicastReconciler(baseAPIManagerLogicReconciler)
			_, err = apicastReconciler.Reconcile()
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
			_, err = apicastReconciler.Reconcile()
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

func testApicastAPIManagerCreator(stagingReplicas, productionReplicas *int64) *appsv1alpha1.APIManager {
	var (
		name                 = "example-apimanager"
		namespace            = "operator-unittest"
		wildcardDomain       = "test.3scale.net"
		appLabel             = "someLabel"
		tenantName           = "someTenant"
		trueValue            = true
		apicastManagementAPI = "disabled"
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
			Apicast: &appsv1alpha1.ApicastSpec{
				ApicastManagementAPI: &apicastManagementAPI,
				OpenSSLVerify:        &trueValue,
				IncludeResponseCodes: &trueValue,
				StagingSpec:          &appsv1alpha1.ApicastStagingSpec{Replicas: stagingReplicas},
				ProductionSpec:       &appsv1alpha1.ApicastProductionSpec{Replicas: productionReplicas},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
}
