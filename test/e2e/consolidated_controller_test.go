package e2e

import (
	goctx "context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis"
	operator "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestBindingController(t *testing.T) {

	var err error

	BindingList := &operator.BindingList{
		TypeMeta: v1.TypeMeta{
			Kind:       "Binding",
			APIVersion: "v1alpha1",
		},
	}

	APIList := &operator.APIList{
		TypeMeta: v1.TypeMeta{
			Kind:       "API",
			APIVersion: "v1alpha1",
		},
	}

	ConsolidatedList := &operator.ConsolidatedList{
		TypeMeta: v1.TypeMeta{
			Kind:       "Consolidated",
			APIVersion: "v1alpha1",
		},
	}

	MetricList := &operator.MetricList{
		TypeMeta: v1.TypeMeta{
			Kind:       "Metric",
			APIVersion: "v1alpha1",
		},
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, BindingList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, APIList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, ConsolidatedList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, MetricList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	t.Run("binding-group", func(t *testing.T) {
		t.Run("BasicBinding", BasicBindingController)
	})
}

func BasicBindingController(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	// get global framework variables
	f := framework.Global
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = BasicBinding(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func BasicBinding(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// create binding custom resource
	exampleBinding := &operator.Binding{
		TypeMeta: v1.TypeMeta{
			Kind:       "Binding",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test01",
			Namespace: namespace,
		},
		Spec: operator.BindingSpec{
			CredentialsRef: v12.SecretReference{
				Name:      "test",
				Namespace: namespace,
			},
			APISelector: v1.LabelSelector{
				MatchLabels:      map[string]string{"api": "myapi"},
				MatchExpressions: nil,
			},
		},
	}

	myAPI := &operator.API{
		TypeMeta: v1.TypeMeta{
			Kind:       "API",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "myapi",
			Labels:    map[string]string{"api": "myapi"},
			Namespace: namespace,
		},
		Spec: operator.APISpec{
			APIBase: operator.APIBase{
				Description: "test",
				IntegrationMethod: operator.IntegrationMethod{
					ApicastOnPrem: &operator.ApicastOnPrem{
						APIcastBaseOptions: operator.APIcastBaseOptions{
							PrivateBaseURL:    "a",
							APITestGetRequest: "a",
							AuthenticationSettings: operator.ApicastAuthenticationSettings{
								HostHeader:  "",
								SecretToken: "",
								Credentials: operator.IntegrationCredentials{
									APIKey: &operator.APIKey{
										AuthParameterName:   "query",
										CredentialsLocation: "user-key",
									},
								},
								Errors: operator.Errors{
									AuthenticationFailed: operator.Authentication{
										ResponseCode: 0,
										ContentType:  "",
										ResponseBody: "",
									},
									AuthenticationMissing: operator.Authentication{
										ResponseCode: 0,
										ContentType:  "",
										ResponseBody: "",
									},
								},
							},
						},
						StagingPublicBaseURL:    "a",
						ProductionPublicBaseURL: "a",
						APIcastBaseSelectors: operator.APIcastBaseSelectors{
							MappingRulesSelector: v1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
							PoliciesSelector: v1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
						},
					},
				},
			},
			APISelectors: operator.APISelectors{
				PlanSelector:   v1.LabelSelector{},
				MetricSelector: v1.LabelSelector{},
			},
		},
		Status: operator.APIStatus{},
	}

	exampleSecret := &v12.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"access_token":     []byte("test"),
			"admin_portal_url": []byte("test"),
		},
		StringData: nil,
		Type:       v12.SecretTypeOpaque,
	}

	exampleMetric := &operator.Metric{
		TypeMeta: v1.TypeMeta{
			Kind:       "Metric",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test01-metric",
			Namespace: namespace,
		},
		Spec: operator.MetricSpec{
			Unit:           "hits",
			Description:    "test",
			IncrementsHits: false,
		},
		Status: operator.MetricStatus{},
	}

	err = f.Client.Create(goctx.TODO(), myAPI, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = f.Client.Create(goctx.TODO(), exampleSecret, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = f.Client.Create(goctx.TODO(), exampleBinding, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = f.Client.Create(goctx.TODO(), exampleMetric, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	consolidated := &operator.Consolidated{}

	tries := 1
Retry:
	tries++
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "test01-consolidated", Namespace: namespace}, consolidated)
	if err != nil && errors.IsNotFound(err) && tries < 3 {
		time.Sleep(2 * time.Second)
		goto Retry
	} else if err != nil {
		return err
	}

	if len(consolidated.Spec.APIs) == 0 {
		return fmt.Errorf("APIs for consolidated object are empty: %#v", consolidated)
	}

	if consolidated.Spec.APIs[0].Name != "myapi" {
		return fmt.Errorf("expected API in consolidated object named: myapi, got: %s", consolidated.Spec.APIs[0].Name)
	}
	if consolidated.Spec.APIs[0].IntegrationMethod.ApicastOnPrem == nil {
		return fmt.Errorf("expected API integration ApicastOnPrem")
	}

	if len(consolidated.Spec.APIs[0].Metrics) == 0 {
		return fmt.Errorf("metrics for API are empty")
	}

	if consolidated.Spec.APIs[0].Metrics[0].Name != "test01-metric" {
		return fmt.Errorf("expected metric name to be test01-metric, got %s", consolidated.Spec.APIs[0].Metrics[0].Name)
	}

	return nil
}
