package e2e

import (
	goctx "context"
	"testing"
	"time"

	"github.com/3scale/3scale-operator/pkg/apis"
	appsgroup "github.com/3scale/3scale-operator/pkg/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	e2eutil "github.com/3scale/3scale-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	frameworke2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	clientappsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	clientroutev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
)

func TestApiManagerController(t *testing.T) {
	var err error

	apimanagerList := &appsv1alpha1.APIManagerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       appsgroup.APIManagerKind,
			APIVersion: appsv1alpha1.SchemeGroupVersion.String(),
		},
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, apimanagerList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// Run subtests
	t.Run("apimanager-group", func(t *testing.T) {
		t.Run("StandardDeploy", productizedUnconstrainedDeploymentSubtest)
	})
}

func newAPIManagerCluster(t *testing.T) (*framework.Framework, *framework.TestCtx) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: cleanupRetryInterval})
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
	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("operator Deployment is ready")

	return f, ctx
}

func productizedUnconstrainedDeploymentSubtest(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: cleanupRetryInterval})
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

	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("operator Deployment is ready")

	enableResourceRequirements := false
	wildcardDomain := "test1.127.0.0.1.nip.io"
	apimanager := &appsv1alpha1.APIManager{
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:              wildcardDomain,
				ResourceRequirementsEnabled: &enableResourceRequirements,
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

	err = waitForAllAPIManagerStandardDeploymentConfigs(t, f.KubeClient, osAppsV1Client, namespace, retryInterval, time.Minute*15)
	if err != nil {
		t.Fatal(err)
	}

	osRouteV1Client, err := clientroutev1.NewForConfig(f.KubeConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = waitForAllAPIManagerStandardRoutes(t, f.KubeClient, osRouteV1Client, namespace, retryInterval, time.Minute*15, wildcardDomain)
	if err != nil {
		t.Fatal(err)
	}

	elapsed = time.Since(start)
	t.Logf("APIManager creation and availability took %s seconds", elapsed)
}

func waitForAllAPIManagerStandardDeploymentConfigs(t *testing.T, kubeclient kubernetes.Interface, osAppsV1Client clientappsv1.AppsV1Interface, namespace string, retryInterval, timeout time.Duration) error {
	deploymentConfigNames := []string{ // TODO gather this from constants/somewhere centralized
		"apicast-production",
		"apicast-staging",
		"backend-cron",
		"backend-listener",
		"backend-redis",
		"backend-worker",
		"system-app",
		"system-memcache",
		"system-mysql",
		"system-redis",
		"system-sidekiq",
		"system-sphinx",
		"zync",
		"zync-que",
		"zync-database",
	}

	for _, dcName := range deploymentConfigNames {
		err := e2eutil.WaitForDeploymentConfig(t, kubeclient, osAppsV1Client, namespace, dcName, retryInterval, time.Minute*15)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForAllAPIManagerStandardRoutes(t *testing.T, kubeclient kubernetes.Interface, osRouteV1Client clientroutev1.RouteV1Interface, namespace string, retryInterval, timeout time.Duration, wildcardDomain string) error {
	routeHosts := []string{
		"backend-3scale." + wildcardDomain,                // Backend Listener route
		"api-3scale-apicast-production." + wildcardDomain, // Apicast Production '3scale' tenant Route
		"api-3scale-apicast-staging." + wildcardDomain,    // Apicast Staging '3scale' tenant Route
		"master." + wildcardDomain,                        // System's Master Portal Route
		"3scale." + wildcardDomain,                        // System's '3scale' tenant Developer Portal Route
		"3scale-admin." + wildcardDomain,                  // System's '3scale' tenant Admin Portal Route
	}
	for _, routeHost := range routeHosts {
		err := e2eutil.WaitForRouteFromHost(t, kubeclient, osRouteV1Client, namespace, routeHost, retryInterval, time.Minute*15)
		if err != nil {
			return err
		}
	}

	return nil
}
