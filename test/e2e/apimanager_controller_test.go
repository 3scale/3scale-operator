package e2e

import (
	goctx "context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/3scale/3scale-operator/pkg/apis"
	appsgroup "github.com/3scale/3scale-operator/pkg/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	e2eutil "github.com/3scale/3scale-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	frameworke2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	clientappsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
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

func registryRedhatIoSecret(t *testing.T, namespace string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "threescale-registry-auth",
			Namespace: namespace,
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: []byte(encodedDockerConfigJSON(t)),
		},
	}
}

func encodedDockerConfigJSON(t *testing.T) string {
	redHatRegistryIOServer := os.Getenv("REGISTRY_REDHAT_IO_SERVER")
	redHatRegistryIOUser := os.Getenv("REGISTRY_REDHAT_IO_USERNAME")
	redHatRegistryIOPass := os.Getenv("REGISTRY_REDHAT_IO_PASSWORD")

	result := map[string]map[string]map[string]string{
		"auths": map[string]map[string]string{
			redHatRegistryIOServer: map[string]string{
				"username": redHatRegistryIOUser,
				"password": redHatRegistryIOPass,
				"auth":     base64.StdEncoding.EncodeToString([]byte(redHatRegistryIOUser + ":" + redHatRegistryIOPass)),
			},
		},
	}

	encodedres, err := json.Marshal(result)
	if err != nil {
		t.Fatal("Error marshaling JSON")
	}
	return string(encodedres)

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

	t.Log("Creating registry.redhat.io secret")
	imagesSecret := registryRedhatIoSecret(t, namespace)
	err = f.Client.Create(goctx.TODO(), imagesSecret, &framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("registry.redhat.io secret created")

	return
	t.Log("waiting until operator Deployment is ready...")

	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("operator Deployment is ready")

	enableResourceRequirements := false
	apicastNightlyImage := "quay.io/3scale/apicast:nightly"
	backendNightlyImage := "quay.io/3scale/apisonator:nightly"
	systemNightlyImage := "quay.io/3scale/porta:nightly"
	zyncNightlyImage := "quay.io/3scale/zync:nightly"
	apimanager := &appsv1alpha1.APIManager{
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:              "test1.127.0.0.1.nip.io",
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
}

func waitForAllApiManagerStandardDeploymentConfigs(t *testing.T, kubeclient kubernetes.Interface, osAppsV1Client clientappsv1.AppsV1Interface, namespace, name string, retryInterval, timeout time.Duration) error {
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
