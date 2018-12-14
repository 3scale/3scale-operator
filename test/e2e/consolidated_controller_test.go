package e2e

import (
	goctx "context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis"
	operator "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v12 "k8s.io/api/core/v1"
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
			Kind:       "API",
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

	t.Run("binding-group", func(t *testing.T) {
		t.Run("ConsolidatedObject", BindingController)
	})
}

func BindingController(t *testing.T) {
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

	if err = BindingCreationTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func BindingCreationTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
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
				Namespace: "myproject",
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
			Namespace: "myproject",
			Labels:    map[string]string{"api": "myapi"},
		},
		Spec: operator.APISpec{
			Description: "",
			IntegrationMethod: operator.IntegrationMethod{
				ApicastOnPrem: &operator.ApicastOnPrem{
					AuthenticationSettings: operator.ApicastAuthenticationSettings{
						Credentials: operator.IntegrationCredentials{
							APIKey: &operator.APIKey{
								AuthParameterName:   "query",
								CredentialsLocation: "user_key",
							},
						},
					},
				},
			},
			PlanSelector:   v1.LabelSelector{},
			MetricSelector: v1.LabelSelector{},
		},
		Status: operator.APIStatus{},
	}

	err = f.Client.Create(goctx.TODO(), myAPI, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = f.Client.Create(goctx.TODO(), exampleBinding, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	consolidated := &operator.Consolidated{}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "test01-consolidated", Namespace: namespace}, consolidated)
	if err != nil {
		return err
	}

	if consolidated.Spec.APIs[0].Name != "myapi" {
		return fmt.Errorf("expected API in consolidated object named: myapi, got: %s", consolidated.Spec.APIs[0].Name)
	}

	return nil
}
