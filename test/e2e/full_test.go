package e2e

import (
	"bytes"
	goctx "context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/apis"
	appsgroup "github.com/3scale/3scale-operator/pkg/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/controller/tenant"
	"github.com/3scale/3scale-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	frameworke2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientappsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
)

func TestFullHappyPath(t *testing.T) {
	var err error

	apimanagerList := apiManagerList()
	tenantList := tenantList()

	err = framework.AddToFrameworkScheme(apis.AddToScheme, apimanagerList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, tenantList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	f := framework.Global
	t.Log("waiting until operator Deployment is ready...")

	cfgHost := f.KubeConfig.Host
	clusterURL, err := url.Parse(cfgHost)
	if err != nil {
		t.Fatal(err)
	}
	if clusterURL.Scheme == "" {
		clusterURL.Scheme = "https"
	}
	clusterHost := clusterURL.Host
	clusterHost = strings.Split(clusterHost, ":")[0]

	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("operator Deployment is ready")

	// Deploy APIManager resource
	enableResourceRequirements := false
	wildcardPolicy := string(routev1.WildcardPolicySubdomain)
	apiManagerWildcardDomain := fmt.Sprintf("test1.%s.nip.io", clusterHost)
	apicastNightlyImage := "quay.io/3scale/apicast:nightly"
	backendNightlyImage := "quay.io/3scale/apisonator:nightly"
	systemNightlyImage := "quay.io/3scale/porta:nightly"
	wildcardRouterNightlyImage := "quay.io/3scale/wildcard-router:nightly"
	zyncNightlyImage := "quay.io/3scale/zync:nightly"
	apimanager := &appsv1alpha1.APIManager{
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:              apiManagerWildcardDomain,
				WildcardPolicy:              &wildcardPolicy,
				ResourceRequirementsEnabled: &enableResourceRequirements,
			},
			Apicast: &appsv1alpha1.ApicastSpec{
				Image: &apicastNightlyImage,
			},
			Backend: &appsv1alpha1.BackendSpec{
				Image: &backendNightlyImage,
			},
			System: &appsv1alpha1.SystemSpec{
				Image: &systemNightlyImage,
			},
			WildcardRouter: &appsv1alpha1.WildcardRouterSpec{
				Image: &wildcardRouterNightlyImage,
			},
			Zync: &appsv1alpha1.ZyncSpec{
				Image: &zyncNightlyImage,
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-apimanager",
			Namespace: namespace,
		},
	}

	var start time.Time
	var elapsed time.Duration

	start = time.Now()

	err = f.Client.Create(goctx.TODO(), apimanager, &framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	osAppsV1Client, err := clientappsv1.NewForConfig(f.KubeConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = waitForAllApiManagerStandardDeploymentConfigs(t, f.KubeClient, osAppsV1Client, namespace, "3scale-operator", retryInterval, time.Minute*15)
	if err != nil {
		t.Fatal(err)
	}

	elapsed = time.Since(start)
	t.Logf("APIManager creation and availability took %s seconds", elapsed)

	start = time.Now()
	// Deploy Tenant resource
	// - Deploy AdminPass secret
	adminPassSecretName := "tenant01adminsecretname"
	adminPass := "thisisapass"
	adminPassSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      adminPassSecretName,
			Labels:    map[string]string{"app": "3scale-operator"},
		},
		StringData: map[string]string{tenant.TenantAdminPasswordSecretField: adminPass},
		Type:       v1.SecretTypeOpaque,
	}
	err = f.Client.Create(goctx.TODO(), adminPassSecret, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Creating tenant admin pass secret")
	err = e2eutil.WaitForSecret(t, f.KubeClient, namespace, adminPassSecretName, retryInterval, time.Minute*2)
	if err != nil {
		t.Fatal(err)
	}

	systemSeedSecret, err := f.KubeClient.CoreV1().Secrets(namespace).Get(component.SystemSecretSystemSeedSecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	masterDomainByteArray, ok := systemSeedSecret.Data[component.SystemSecretSystemSeedMasterDomainFieldName]
	if !ok {
		t.Fatalf("field %s not found in systemseed secret", component.SystemSecretSystemSeedMasterDomainFieldName)
	}

	masterDomain := bytes.NewBuffer(masterDomainByteArray).String()

	// deploy tenant resource
	tenantSecretName := "tenantproviderkeysecret"
	systemMasterURL := fmt.Sprintf("https://%s.%s", masterDomain, apimanager.Spec.WildcardDomain)
	tenant := &apiv1alpha1.Tenant{
		Spec: apiv1alpha1.TenantSpec{
			Username:         "admin",
			Email:            "admin@example.com",
			OrganizationName: "ECorp",
			SystemMasterUrl:  systemMasterURL,
			PasswordCredentialsRef: v1.SecretReference{
				Name:      adminPassSecretName,
				Namespace: namespace,
			},
			MasterCredentialsRef: v1.SecretReference{
				Name:      component.SystemSecretSystemSeedSecretName,
				Namespace: namespace,
			},
			TenantSecretRef: v1.SecretReference{
				Name:      tenantSecretName,
				Namespace: namespace,
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant01",
			Namespace: namespace,
		},
	}
	t.Log("Creating tenant resource")
	err = f.Client.Create(goctx.TODO(), tenant, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}
	err = e2eutil.WaitForSecret(t, f.KubeClient, namespace, tenantSecretName, retryInterval, time.Minute*2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Tenant reconciliation DONE")
	elapsed = time.Since(start)
	t.Logf("Tenant creation and availability took %s seconds", elapsed)

	start = time.Now()

	api := &apiv1alpha1.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testapi",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.APISpec{
			APIBase: apiv1alpha1.APIBase{
				Description: "testapi created by 3scale operator",
				IntegrationMethod: apiv1alpha1.IntegrationMethod{
					ApicastHosted: &apiv1alpha1.ApicastHosted{
						APIcastBaseOptions: apiv1alpha1.APIcastBaseOptions{
							PrivateBaseURL:    "https://echo-api.3scale.net:443",
							APITestGetRequest: "/",
							AuthenticationSettings: apiv1alpha1.ApicastAuthenticationSettings{
								HostHeader:  "",
								SecretToken: "Shared_secret_sent_from_proxy_to_API_backend_7c2229057468d5fd",
								Credentials: apiv1alpha1.IntegrationCredentials{
									APIKey: &apiv1alpha1.APIKey{
										AuthParameterName:   "user_key",
										CredentialsLocation: "query",
									},
								},
								Errors: apiv1alpha1.Errors{
									AuthenticationFailed: apiv1alpha1.Authentication{
										ResponseCode: 403,
										ContentType:  "text/plain; charset=us-ascii",
										ResponseBody: "Authentication failed",
									},
									AuthenticationMissing: apiv1alpha1.Authentication{
										ResponseCode: 403,
										ContentType:  "text/plain; charset=us-ascii",
										ResponseBody: "Authentication parameters missing",
									},
								},
							},
						},
						APIcastBaseSelectors: apiv1alpha1.APIcastBaseSelectors{
							MappingRulesSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"environment": "testing"},
							},
						},
					},
				},
			},
			APISelectors: apiv1alpha1.APISelectors{
				PlanSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"environment": "testing"},
				},
				MetricSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"environment": "testing"},
				},
			},
		},
		Status: apiv1alpha1.APIStatus{},
	}

	err = f.Client.Create(goctx.TODO(), api, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	metric01 := &apiv1alpha1.Metric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric01-test",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.MetricSpec{
			Unit:          "hits",
			Description:   "metric 01",
			IncrementHits: false,
		},
		Status: apiv1alpha1.MetricStatus{},
	}

	err = f.Client.Create(goctx.TODO(), metric01, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	metric02 := &apiv1alpha1.Metric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric02",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.MetricSpec{
			Unit:          "hits",
			Description:   "metric 02",
			IncrementHits: false,
		},
		Status: apiv1alpha1.MetricStatus{},
	}

	err = f.Client.Create(goctx.TODO(), metric02, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	plan := &apiv1alpha1.Plan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "plan01",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.PlanSpec{
			PlanBase: apiv1alpha1.PlanBase{
				Default:          true,
				TrialPeriod:      0,
				ApprovalRequired: false,
				Costs: apiv1alpha1.PlanCost{
					SetupFee:  1,
					CostMonth: 1,
				},
			},
			PlanSelectors: apiv1alpha1.PlanSelectors{
				LimitSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"environment": "testing"},
				},
			},
		},
		Status: apiv1alpha1.PlanStatus{},
	}

	err = f.Client.Create(goctx.TODO(), plan, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	limit01 := &apiv1alpha1.Limit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "limit01",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.LimitSpec{
			LimitBase: apiv1alpha1.LimitBase{
				Period:   "eternity",
				MaxValue: 100,
			},
			LimitObjectRef: apiv1alpha1.LimitObjectRef{
				Metric: v1.ObjectReference{
					Namespace: namespace,
					Name:      "metric01-test",
				},
			},
		},
		Status: apiv1alpha1.LimitStatus{},
	}

	err = f.Client.Create(goctx.TODO(), limit01, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	limit02 := &apiv1alpha1.Limit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "limit02",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.LimitSpec{
			LimitBase: apiv1alpha1.LimitBase{
				Period:   "day",
				MaxValue: 100,
			},
			LimitObjectRef: apiv1alpha1.LimitObjectRef{
				Metric: v1.ObjectReference{
					Namespace: namespace,
					Name:      "Hits",
				},
			},
		},
		Status: apiv1alpha1.LimitStatus{},
	}

	err = f.Client.Create(goctx.TODO(), limit02, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	mappingRule01 := &apiv1alpha1.MappingRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mappingrule01",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.MappingRuleSpec{
			MappingRuleBase: apiv1alpha1.MappingRuleBase{
				Path:      "/testing",
				Method:    "GET",
				Increment: 1,
			},
			MappingRuleMetricRef: apiv1alpha1.MappingRuleMetricRef{
				MetricRef: v1.ObjectReference{
					Namespace: namespace,
					Name:      "Hits",
				},
			},
		},
		Status: apiv1alpha1.MappingRuleStatus{},
	}

	err = f.Client.Create(goctx.TODO(), mappingRule01, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	mappingRule02 := &apiv1alpha1.MappingRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mappingrule02",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.MappingRuleSpec{
			MappingRuleBase: apiv1alpha1.MappingRuleBase{
				Path:      "/metric01",
				Method:    "POST",
				Increment: 10,
			},
			MappingRuleMetricRef: apiv1alpha1.MappingRuleMetricRef{
				MetricRef: v1.ObjectReference{
					Namespace: namespace,
					Name:      "metric01-test",
				},
			},
		},
		Status: apiv1alpha1.MappingRuleStatus{},
	}

	err = f.Client.Create(goctx.TODO(), mappingRule02, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	binding := &apiv1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testbinding",
			Namespace: namespace,
			Labels:    map[string]string{"environment": "testing"},
		},
		Spec: apiv1alpha1.BindingSpec{
			CredentialsRef: v1.SecretReference{
				Name:      tenantSecretName,
				Namespace: namespace,
			},
			APISelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"environment": "testing"},
			},
		},
		Status: apiv1alpha1.BindingStatus{},
	}

	err = f.Client.Create(goctx.TODO(), binding, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: binding.Name}, binding)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Checking for the binding object to be in sync with 3scale")

	err = e2eutil.WaitForReconciliationWith3scale(t, f.Client, *binding, 30*time.Second, 240*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	elapsed = time.Since(start)
	t.Logf("Binding in sync took %s seconds", elapsed)

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: binding.Name}, binding)
	if err != nil {
		t.Fatal(err)
	}

	currentState, err := binding.GetCurrentState()
	if err != nil {
		t.Fatal(err)
	}
	desiredState, err := binding.GetDesiredState()
	if err != nil {
		t.Fatal(err)
	}
	if !apiv1alpha1.CompareStates(*currentState, *desiredState) {
		t.Fatalf("States are not in sync: \n\ncurrent: %#v\n\ndesired: %#v\n\n", *currentState, *desiredState)
	}

}

func tenantList() *apiv1alpha1.TenantList {
	return &apiv1alpha1.TenantList{
		TypeMeta: metav1.TypeMeta{
			Kind:       apiv1alpha1.TenantKind,
			APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
		},
	}
}

func apiManagerList() *appsv1alpha1.APIManagerList {
	return &appsv1alpha1.APIManagerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       appsgroup.APIManagerKind,
			APIVersion: appsv1alpha1.SchemeGroupVersion.String(),
		},
	}
}
