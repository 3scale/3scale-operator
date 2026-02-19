package operator

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/version"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	policyv1 "k8s.io/api/policy/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	wildcardDomain = "test.3scale.net"
	appLabel       = "someLabel"
	apimanagerName = "example-apimanager"
	namespace      = "someNS"
	tenantName     = "someTenant"
	trueValue      = true
)

func addExpectedMeteringLabels(src map[string]string, componentName string, componentType helper.ComponentType) {
	labels := []struct {
		k string
		v string
	}{
		{"com.company", "Red_Hat"},
		{"rht.prod_name", "Red_Hat_Integration"},
		{"rht.prod_ver", "master"},
		{"rht.comp", "3scale"},
		{"rht.comp_ver", version.ThreescaleVersionMajorMinor()},
		{"rht.subcomp", componentName},
		{"rht.subcomp_t", string(componentType)},
	}
	for _, label := range labels {
		src[label.k] = label.v
	}
}

func basicApimanager() *appsv1alpha1.APIManager {
	tmpAppLabel := appLabel
	tmpTenantName := tenantName
	tmpTrueValue := trueValue

	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:              wildcardDomain,
				AppLabel:                    &tmpAppLabel,
				TenantName:                  &tmpTenantName,
				ResourceRequirementsEnabled: &tmpTrueValue,
			},
			System: &appsv1alpha1.SystemSpec{},
		},
	}

	_, err := apimanager.SetDefaults()
	if err != nil {
		panic(fmt.Errorf("Error creating Basic APIManager: %v", err))
	}
	return apimanager
}

func GetTestSecret(namespace, secretName string, data map[string]string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: data,
		Type:       v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getSystemDBSecret(databaseURL, username, password string) *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemDatabaseUserFieldName:     username,
		component.SystemSecretSystemDatabasePasswordFieldName: password,
		component.SystemSecretSystemDatabaseURLFieldName:      databaseURL,
	}
	return GetTestSecret(namespace, component.SystemSecretSystemDatabaseSecretName, data)
}

func getTestAffinity(prefix string) *v1.Affinity {
	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchFields: []v1.NodeSelectorRequirement{
							{
								Key:      fmt.Sprintf("%s-%s", prefix, "key2"),
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{fmt.Sprintf("%s-%s", prefix, "val2")},
							},
						},
					},
				},
			},
		},
	}
}

func getTestTolerations(prefix string) []v1.Toleration {
	return []v1.Toleration{
		{
			Key:      fmt.Sprintf("%s-%s", prefix, "key1"),
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    fmt.Sprintf("%s-%s", prefix, "val1"),
		},
		{
			Key:      fmt.Sprintf("%s-%s", prefix, "key2"),
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    fmt.Sprintf("%s-%s", prefix, "val2"),
		},
	}
}

// Generic test helpers - reusable across all reconciler tests

// setupScheme registers all required Kubernetes types into the scheme
func setupScheme(t *testing.T) *runtime.Scheme {
	s := scheme.Scheme
	if err := appsv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := v1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := k8sappsv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := batchv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := imagev1.Install(s); err != nil {
		t.Fatal(err)
	}
	if err := routev1.Install(s); err != nil {
		t.Fatal(err)
	}
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := configv1.Install(s); err != nil {
		t.Fatal(err)
	}
	if err := policyv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

// jobExists checks if a job exists in the cluster
func jobExists(t *testing.T, client k8sclient.Client, name, namespace string) bool {
	job := &batchv1.Job{}
	err := client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, job)
	if err != nil {
		if k8serr.IsNotFound(err) {
			return false
		}
		t.Fatalf("unexpected error checking if job %s exists: %v", name, err)
	}
	return true
}

// getJob retrieves a job from the cluster
func getJob(t *testing.T, client k8sclient.Client, name, namespace string) *batchv1.Job {
	job := &batchv1.Job{}
	err := client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, job)
	if err != nil {
		t.Fatalf("failed to get job %s: %v", name, err)
	}
	return job
}

// deploymentExists checks if a deployment exists in the cluster
func deploymentExists(t *testing.T, client k8sclient.Client, name, namespace string) bool {
	deployment := &k8sappsv1.Deployment{}
	err := client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if err != nil {
		if k8serr.IsNotFound(err) {
			return false
		}
		t.Fatalf("unexpected error checking if deployment %s exists: %v", name, err)
	}
	return true
}

func createJob(name, namespace, image string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "hook",
							Image: image,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
		Status: batchv1.JobStatus{},
	}
}

// createCompletedJob creates a job fixture that is marked as completed
func createCompletedJob(name, namespace, image string, revision int64) *batchv1.Job {
	job := createJob(name, namespace, image)

	// this fixture simulates a completed job already on the cluster created by a
	// specific version of the operator - the annotation must be set as a literal string
	job.Annotations = map[string]string{
		"apimanager.apps.3scale.net/system-app-deployment-revision": strconv.FormatInt(revision, 10),
	}
	job.Status.Conditions = []batchv1.JobCondition{
		{
			Type:   batchv1.JobComplete,
			Status: v1.ConditionTrue,
		},
	}

	return job
}

// createIncompleteJob creates a job fixture that is still running
func createIncompleteJob(name, namespace, image string, revision int64) *batchv1.Job {
	job := createJob(name, namespace, image)

	job.Annotations = map[string]string{
		component.SystemAppRevisionAnnotation: strconv.FormatInt(revision, 10),
	}

	return job
}

// createJobWithoutAnnotation creates a job fixture without revision annotation
// This simulates jobs created manually or by older operator versions
func createJobWithoutAnnotation(name, namespace, image string, completed bool) *batchv1.Job {
	job := createJob(name, namespace, image)

	if completed {
		job.Status.Conditions = []batchv1.JobCondition{
			{
				Type:   batchv1.JobComplete,
				Status: v1.ConditionTrue,
			},
		}
	}

	return job
}
