package controllers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	testWildcardDomain = "test.example.com"
	testTenantName     = "3scale"
)

// getTestAPIManager returns a basic APIManager CR for testing.
func getTestAPIManager(namespace string) *appsv1alpha1.APIManager {
	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-apimanager",
			Namespace: namespace,
			UID:       "test-uid-123",
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:              testWildcardDomain,
				TenantName:                  ptr.To(testTenantName),
				ResourceRequirementsEnabled: ptr.To(true),
			},
			Backend: &appsv1alpha1.BackendSpec{
				ListenerSpec: &appsv1alpha1.BackendListenerSpec{},
				WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{},
			},
			Apicast: &appsv1alpha1.ApicastSpec{
				ProductionSpec: &appsv1alpha1.ApicastProductionSpec{},
				StagingSpec:    &appsv1alpha1.ApicastStagingSpec{},
			},
		},
		Status: appsv1alpha1.APIManagerStatus{
			Conditions: common.Conditions{},
		},
	}
}

// getTestDeployment returns a Deployment owned by the given APIManager UID, with availability controlled by available.
func getTestDeployment(name, namespace string, apimanagerUID string, available bool) *k8sappsv1.Deployment {
	replicas := int32(1)
	availableReplicas := int32(0)
	conditionStatus := corev1.ConditionFalse
	if available {
		availableReplicas = int32(1)
		conditionStatus = corev1.ConditionTrue
	}

	return &k8sappsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps.3scale.net/v1alpha1",
					Kind:       "APIManager",
					Name:       "example-apimanager",
					UID:        types.UID(apimanagerUID),
				},
			},
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: k8sappsv1.DeploymentStatus{
			Replicas:          replicas,
			UpdatedReplicas:   availableReplicas,
			ReadyReplicas:     availableReplicas,
			AvailableReplicas: availableReplicas,
			Conditions: []k8sappsv1.DeploymentCondition{
				{
					Type:   k8sappsv1.DeploymentAvailable,
					Status: conditionStatus,
				},
			},
		},
	}
}

// getAllStandardDeployments returns all standard APIManager deployments as a []runtime.Object slice.
func getAllStandardDeployments(namespace string, apimanagerUID string, available bool) []runtime.Object {
	deploymentNames := []string{
		"apicast-staging",
		"apicast-production",
		"backend-listener",
		"backend-worker",
		"backend-cron",
		"system-memcache",
		"system-app",
		"system-sidekiq",
		"system-searchd",
		"zync",
		"zync-que",
		"zync-database",
	}

	deployments := make([]runtime.Object, 0, len(deploymentNames))
	for _, name := range deploymentNames {
		deployments = append(deployments, getTestDeployment(name, namespace, apimanagerUID, available))
	}
	return deployments
}

// getAPIManagerBaseReconciler returns a BaseReconciler backed by a fake client pre-populated with objects.
func getAPIManagerBaseReconciler(objects ...runtime.Object) *reconcilers.BaseReconciler {
	s := scheme.Scheme
	err := appsv1alpha1.AddToScheme(s)
	if err != nil {
		return nil
	}
	err = routev1.AddToScheme(s)
	if err != nil {
		return nil
	}

	var clientObjects []client.Object
	for _, o := range objects {
		co, ok := o.(client.Object)
		if ok {
			clientObjects = append(clientObjects, co)
		}
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objects...).WithStatusSubresource(clientObjects...).Build()
	log := logf.Log.WithName("apimanager status reconciler test")
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)
	return reconcilers.NewBaseReconciler(context.TODO(), cl, s, cl, log, clientset.Discovery(), recorder)
}

// getRequiredSecrets returns the secrets that the default APIManager configuration references.
func getRequiredSecrets(namespace string) []runtime.Object {
	return []runtime.Object{
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "system-redis", Namespace: namespace},
			Data:       map[string][]byte{"URL": []byte("redis://system-redis:6379/1")},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "backend-redis", Namespace: namespace},
			Data: map[string][]byte{
				"REDIS_STORAGE_URL": []byte("redis://backend-redis:6379/0"),
				"REDIS_QUEUES_URL":  []byte("redis://backend-redis:6379/1"),
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "system-database", Namespace: namespace},
			Data:       map[string][]byte{"URL": []byte("mysql2://root:password@system-mysql:3306/system")},
		},
	}
}

// makeAdmittedRoute returns a Route with the Admitted condition set to True.
func makeAdmittedRoute(name, namespace, host string) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       routev1.RouteSpec{Host: host},
		Status: routev1.RouteStatus{
			Ingress: []routev1.RouteIngress{
				{Conditions: []routev1.RouteIngressCondition{
					{Type: routev1.RouteAdmitted, Status: corev1.ConditionTrue},
				}},
			},
		},
	}
}

// getRequiredRoutes returns all routes expected by the default APIManager tenant configuration.
func getRequiredRoutes(namespace, wildcardDomain, tenantName string) []runtime.Object {
	return []runtime.Object{
		makeAdmittedRoute("backend-route", namespace, fmt.Sprintf("backend-%s.%s", tenantName, wildcardDomain)),
		makeAdmittedRoute("apicast-production-route", namespace, fmt.Sprintf("api-%s-apicast-production.%s", tenantName, wildcardDomain)),
		makeAdmittedRoute("apicast-staging-route", namespace, fmt.Sprintf("api-%s-apicast-staging.%s", tenantName, wildcardDomain)),
		makeAdmittedRoute("master-route", namespace, fmt.Sprintf("master.%s", wildcardDomain)),
		makeAdmittedRoute("developer-portal-route", namespace, fmt.Sprintf("%s.%s", tenantName, wildcardDomain)),
		makeAdmittedRoute("admin-portal-route", namespace, fmt.Sprintf("%s-admin.%s", tenantName, wildcardDomain)),
	}
}

// concat joins multiple []runtime.Object slices into one.
func concat(slices ...[]runtime.Object) []runtime.Object {
	var result []runtime.Object
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func TestAPIManagerStatusReconciler_Reconcile_statusConditions(t *testing.T) {
	namespace := "test-namespace"
	// WATCH_NAMESPACE is read by non-bypassed preflight cases; harmless for all others.
	t.Setenv("WATCH_NAMESPACE", namespace)

	am := getTestAPIManager(namespace)
	amUID := string(am.UID)

	// Namespace is environment-derived, not configurable.
	operatorNs, err := helper.GetOperatorNamespace()
	if err != nil {
		t.Fatalf("GetOperatorNamespace: %v", err)
	}

	requirementsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helper.OperatorRequirementsConfigMapName,
			Namespace: operatorNs,
		},
		Data: map[string]string{
			helper.RHTThreescaleVersion: version.ThreescaleVersionMajorMinor(),
		},
	}

	healthy := func(apimanager *appsv1alpha1.APIManager) []runtime.Object {
		return concat(
			getAllStandardDeployments(namespace, amUID, true),
			getRequiredSecrets(namespace),
			getRequiredRoutes(namespace, testWildcardDomain, testTenantName),
			[]runtime.Object{apimanager},
		)
	}

	condTrue := func(t *testing.T, updated *appsv1alpha1.APIManager, condType common.ConditionType) {
		t.Helper()
		cond := updated.Status.Conditions.GetCondition(condType)
		if cond == nil {
			t.Errorf("condition %q not found", condType)
			return
		}
		if !cond.IsTrue() {
			t.Errorf("condition %q: got False, want True", condType)
		}
	}
	condFalse := func(t *testing.T, updated *appsv1alpha1.APIManager, condType common.ConditionType) {
		t.Helper()
		cond := updated.Status.Conditions.GetCondition(condType)
		if cond == nil {
			t.Errorf("condition %q not found", condType)
			return
		}
		if cond.IsTrue() {
			t.Errorf("condition %q: got True, want False", condType)
		}
	}

	tests := []struct {
		name          string
		objects       []runtime.Object
		preflightsErr error
		bypass        bool
		assert        func(t *testing.T, updated *appsv1alpha1.APIManager)
	}{
		{
			name:    "All healthy - Available and all sub-conditions are True",
			objects: healthy(am),
			bypass:  true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				condTrue(t, updated, appsv1alpha1.APIManagerAvailableConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerDeploymentsAvailableConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerRoutesReadyConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerSecretsAvailableConditionType)
				if len(updated.Status.Deployments.Ready) == 0 && len(updated.Status.Deployments.Starting) == 0 && len(updated.Status.Deployments.Stopped) == 0 {
					t.Error("Deployments status is empty")
				}
			},
		},
		{
			name: "All deployments unavailable - Available and DeploymentsAvailable are False",
			objects: concat(
				getAllStandardDeployments(namespace, amUID, false),
				getRequiredSecrets(namespace),
				getRequiredRoutes(namespace, testWildcardDomain, testTenantName),
				[]runtime.Object{am},
			),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				condFalse(t, updated, appsv1alpha1.APIManagerAvailableConditionType)
				condFalse(t, updated, appsv1alpha1.APIManagerDeploymentsAvailableConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerRoutesReadyConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerSecretsAvailableConditionType)
			},
		},
		{
			name: "One deployment unavailable - DeploymentsAvailable condition names the failing deployment",
			objects: func() []runtime.Object {
				deps := getAllStandardDeployments(namespace, amUID, true)
				for i, obj := range deps {
					if d, ok := obj.(*k8sappsv1.Deployment); ok && d.Name == "apicast-staging" {
						deps[i] = getTestDeployment("apicast-staging", namespace, amUID, false)
						break
					}
				}
				return concat(deps, getRequiredSecrets(namespace), getRequiredRoutes(namespace, testWildcardDomain, testTenantName), []runtime.Object{am})
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerDeploymentsAvailableConditionType)
				if cond == nil {
					t.Fatal("DeploymentsAvailable condition not found")
				}
				if cond.IsTrue() {
					t.Error("DeploymentsAvailable: got True, want False")
				}
				if cond.Reason != "DeploymentsNotAvailable" {
					t.Errorf("Reason = %q, want %q", cond.Reason, "DeploymentsNotAvailable")
				}
				if !strings.Contains(cond.Message, "apicast-staging") {
					t.Errorf("Message = %q, want it to contain %q", cond.Message, "apicast-staging")
				}
			},
		},
		{
			name: "One deployment missing from cluster - DeploymentsAvailable condition names the missing deployment",
			objects: func() []runtime.Object {
				var deps []runtime.Object
				for _, obj := range getAllStandardDeployments(namespace, amUID, true) {
					if d, ok := obj.(*k8sappsv1.Deployment); ok && d.Name == "apicast-staging" {
						continue
					}
					deps = append(deps, obj)
				}
				return concat(deps, getRequiredSecrets(namespace), getRequiredRoutes(namespace, testWildcardDomain, testTenantName), []runtime.Object{am})
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerDeploymentsAvailableConditionType)
				if cond == nil {
					t.Fatal("DeploymentsAvailable condition not found")
				}
				if cond.IsTrue() {
					t.Error("DeploymentsAvailable: got True, want False")
				}
				if !strings.Contains(cond.Message, "apicast-staging") {
					t.Errorf("Message = %q, want it to contain %q", cond.Message, "apicast-staging")
				}
			},
		},
		{
			name: "Route not admitted - RoutesReady is False with host in message",
			objects: func() []runtime.Object {
				routes := getRequiredRoutes(namespace, testWildcardDomain, testTenantName)
				for i, obj := range routes {
					if r, ok := obj.(*routev1.Route); ok && r.Name == "backend-route" {
						r.Status.Ingress[0].Conditions[0].Status = corev1.ConditionFalse
						routes[i] = r
					}
				}
				return concat(getAllStandardDeployments(namespace, amUID, true), getRequiredSecrets(namespace), routes, []runtime.Object{am})
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				condFalse(t, updated, appsv1alpha1.APIManagerAvailableConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerDeploymentsAvailableConditionType)
				condTrue(t, updated, appsv1alpha1.APIManagerSecretsAvailableConditionType)
				cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerRoutesReadyConditionType)
				if cond == nil {
					t.Fatal("RoutesReady condition not found")
				}
				if cond.IsTrue() {
					t.Error("RoutesReady: got True, want False")
				}
				host := fmt.Sprintf("backend-%s.%s", testTenantName, testWildcardDomain)
				if !strings.Contains(cond.Message, host) {
					t.Errorf("RoutesReady Message = %q, want it to contain %q", cond.Message, host)
				}
			},
		},
		{
			name: "Route missing from cluster - RoutesReady is False with host in message",
			objects: func() []runtime.Object {
				var routes []runtime.Object
				for _, obj := range getRequiredRoutes(namespace, testWildcardDomain, testTenantName) {
					if r, ok := obj.(*routev1.Route); ok && r.Name == "backend-route" {
						continue
					}
					routes = append(routes, obj)
				}
				return concat(getAllStandardDeployments(namespace, amUID, true), getRequiredSecrets(namespace), routes, []runtime.Object{am})
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerRoutesReadyConditionType)
				if cond == nil {
					t.Fatal("RoutesReady condition not found")
				}
				if cond.IsTrue() {
					t.Error("RoutesReady: got True, want False")
				}
				host := fmt.Sprintf("backend-%s.%s", testTenantName, testWildcardDomain)
				if !strings.Contains(cond.Message, host) {
					t.Errorf("RoutesReady Message = %q, want it to contain %q", cond.Message, host)
				}
			},
		},
		{
			// SecretsAvailable=False is also asserted to carry its Reason and Message up to the
			// top-level Available condition for backwards compatibility with automation that reads
			// failure details from Available rather than the dedicated sub-condition.
			name: "Watched secret missing - SecretsAvailable is False; Available carries its Reason and Message",
			objects: func() []runtime.Object {
				amSecret := getTestAPIManager(namespace)
				amSecret.Spec.Apicast = &appsv1alpha1.ApicastSpec{
					ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
						CustomEnvironments: []appsv1alpha1.CustomEnvironmentSpec{
							{SecretRef: &corev1.LocalObjectReference{Name: "missing-secret"}},
						},
					},
					StagingSpec: &appsv1alpha1.ApicastStagingSpec{},
				}
				return concat(
					getAllStandardDeployments(namespace, amUID, true),
					getRequiredRoutes(namespace, testWildcardDomain, testTenantName),
					[]runtime.Object{amSecret},
				)
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				secretsCond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerSecretsAvailableConditionType)
				if secretsCond == nil {
					t.Fatal("SecretsAvailable condition not found")
				}
				if secretsCond.IsTrue() {
					t.Error("SecretsAvailable: got True, want False")
				}
				if secretsCond.Reason != "MissingWatchedSecrets" {
					t.Errorf("SecretsAvailable Reason = %q, want %q", secretsCond.Reason, "MissingWatchedSecrets")
				}
				if !strings.Contains(secretsCond.Message, "missing-secret") {
					t.Errorf("SecretsAvailable Message = %q, want it to contain %q", secretsCond.Message, "missing-secret")
				}
				// availCond to match secretsCond for backward compatibility
				availCond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerAvailableConditionType)
				if availCond == nil {
					t.Fatal("Available condition not found")
				}
				if availCond.Reason != secretsCond.Reason {
					t.Errorf("Available.Reason = %q, want %q (from SecretsAvailable)", availCond.Reason, secretsCond.Reason)
				}
				if availCond.Message != secretsCond.Message {
					t.Errorf("Available.Message = %q, want %q (from SecretsAvailable)", availCond.Message, secretsCond.Message)
				}
			},
		},
		{
			name: "HPA with ResourceRequirements disabled - Warning conditions emitted",
			objects: func() []runtime.Object {
				amHPA := getTestAPIManager(namespace)
				amHPA.Spec.ResourceRequirementsEnabled = ptr.To(false)
				amHPA.Spec.Backend.ListenerSpec.Hpa = true
				amHPA.Spec.Backend.WorkerSpec.Hpa = true
				amHPA.Spec.Apicast.ProductionSpec.Hpa = true
				return healthy(amHPA)
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				wantReasons := map[string]bool{
					"HPA & ResourceRequirementsEnabled false": true,
					"HPA": true,
				}
				var actualReasons []string
				for _, cond := range updated.Status.Conditions {
					if cond.Type == appsv1alpha1.APIManagerWarningConditionType {
						actualReasons = append(actualReasons, string(cond.Reason))
					}
				}
				if len(actualReasons) != len(wantReasons) {
					t.Errorf("warning reasons: got %v, want %v", actualReasons, wantReasons)
					return
				}
				for _, r := range actualReasons {
					if !wantReasons[r] {
						t.Errorf("unexpected warning reason %q", r)
					}
				}
			},
		},
		{
			name: "OpenTracing enabled - deprecation Warning conditions emitted",
			objects: func() []runtime.Object {
				amOT := getTestAPIManager(namespace)
				amOT.Spec.Apicast.ProductionSpec.OpenTracing = &appsv1alpha1.APIcastOpenTracingSpec{Enabled: ptr.To(true)}
				amOT.Spec.Apicast.StagingSpec.OpenTracing = &appsv1alpha1.APIcastOpenTracingSpec{Enabled: ptr.To(true)}
				return healthy(amOT)
			}(),
			bypass: true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				wantReasons := map[string]bool{
					"Apicast Staging OpenTracing Deprecation":    true,
					"Apicast Production OpenTracing Deprecation": true,
				}
				var actualReasons []string
				for _, cond := range updated.Status.Conditions {
					if cond.Type == appsv1alpha1.APIManagerWarningConditionType {
						actualReasons = append(actualReasons, string(cond.Reason))
					}
				}
				if len(actualReasons) != len(wantReasons) {
					t.Errorf("warning reasons: got %v, want %v", actualReasons, wantReasons)
					return
				}
				for _, r := range actualReasons {
					if !wantReasons[r] {
						t.Errorf("unexpected warning reason %q", r)
					}
				}
			},
		},
		{
			name:    "Preflights bypassed - condition not written",
			objects: healthy(am),
			bypass:  true,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				if cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerPreflightsConditionType); cond != nil {
					t.Error("Preflights condition present, want absent")
				}
			},
		},
		{
			name:    "Preflights met - condition is True",
			objects: append(healthy(am), requirementsConfigMap),
			bypass:  false,
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerPreflightsConditionType)
				if cond == nil {
					t.Fatal("Preflights condition not found")
				}
				if !cond.IsTrue() {
					t.Errorf("Preflights: got False, want True (message: %q)", cond.Message)
				}
				if !strings.Contains(cond.Message, "All requirements") {
					t.Errorf("Preflights Message = %q, want to contain %q", cond.Message, "All requirements")
				}
			},
		},
		{
			name:          "Preflights failed - condition is False with error in message",
			objects:       append(healthy(am), requirementsConfigMap),
			bypass:        false,
			preflightsErr: fmt.Errorf("system-database secret missing required key"),
			assert: func(t *testing.T, updated *appsv1alpha1.APIManager) {
				cond := updated.Status.Conditions.GetCondition(appsv1alpha1.APIManagerPreflightsConditionType)
				if cond == nil {
					t.Fatal("Preflights condition not found")
				}
				if cond.IsTrue() {
					t.Error("Preflights: got True, want False")
				}
				if !strings.Contains(cond.Message, "Preflights failed") {
					t.Errorf("Preflights Message = %q, want to contain %q", cond.Message, "Preflights failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bypass {
				t.Setenv("PREFLIGHT_CHECKS_BYPASS", "true")
			}

			base := getAPIManagerBaseReconciler(tt.objects...)

			fetched := &appsv1alpha1.APIManager{}
			if err := base.Client().Get(context.TODO(), types.NamespacedName{Name: "example-apimanager", Namespace: namespace}, fetched); err != nil {
				t.Fatalf("get APIManager: %v", err)
			}

			s := &APIManagerStatusReconciler{
				BaseReconciler:     base,
				apimanagerResource: fetched,
				logger:             logr.Discard(),
				preflightsErr:      tt.preflightsErr,
			}

			if _, err := s.Reconcile(); err != nil {
				t.Fatalf("Reconcile() unexpected error: %v", err)
			}

			updated := &appsv1alpha1.APIManager{}
			if err := base.Client().Get(context.TODO(), types.NamespacedName{Name: "example-apimanager", Namespace: namespace}, updated); err != nil {
				t.Fatalf("get updated APIManager: %v", err)
			}

			tt.assert(t, updated)
		})
	}
}

// TestAPIManagerStatusReconciler_Reconcile_requeueOnTrueToFalseTransition is a regression
// test for the stale-read requeue bug: when Available transitions from True to False the
// reconciler must requeue, not settle silently at Available=False.
func TestAPIManagerStatusReconciler_Reconcile_requeueOnTrueToFalseTransition(t *testing.T) {
	namespace := "test-namespace"
	t.Setenv("PREFLIGHT_CHECKS_BYPASS", "true")

	// Seed the CR with Available=True in its current status (old state).
	am := getTestAPIManager(namespace)
	am.Status.Conditions = common.Conditions{
		{Type: appsv1alpha1.APIManagerAvailableConditionType, Status: corev1.ConditionTrue},
	}

	// Cluster state: deployments not available, so calculateStatus() will return Available=False.
	objects := concat(
		getAllStandardDeployments(namespace, string(am.UID), false),
		getRequiredSecrets(namespace),
		getRequiredRoutes(namespace, testWildcardDomain, testTenantName),
		[]runtime.Object{am},
	)

	s := &APIManagerStatusReconciler{
		BaseReconciler:     getAPIManagerBaseReconciler(objects...),
		apimanagerResource: am,
		logger:             logr.Discard(),
	}

	result, err := s.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile() unexpected error: %v", err)
	}
	if !result.Requeue {
		t.Errorf("Reconcile() Requeue = false, want true on Available True-to-False transition")
	}
}
