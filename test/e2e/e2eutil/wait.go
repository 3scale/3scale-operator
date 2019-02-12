package e2eutil

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	appsv1 "github.com/openshift/api/apps/v1"
	clientappsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WaitForDeploymentConfig(t *testing.T, kubeclient kubernetes.Interface, osAppsV1Client clientappsv1.AppsV1Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		dcInterface := osAppsV1Client.DeploymentConfigs(namespace)
		dc, err := dcInterface.Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of '%s' DeploymentConfig\n", name)
				return false, nil
			}
			return false, err
		}

		isReady := false
		dcConditions := dc.Status.Conditions
		for _, dcCondition := range dcConditions {
			if dcCondition.Type == appsv1.DeploymentAvailable && dcCondition.Status == corev1.ConditionTrue {
				isReady = true
			}
		}
		if isReady {
			t.Logf("DeploymentConfig '%s' available\n", name)
			return true, nil
		}
		availableReplicas := dc.Status.AvailableReplicas
		desiredReplicas := dc.Spec.Replicas
		t.Logf("Waiting for full availability of %s DeploymentConfig (%d/%d)\n", name, availableReplicas, desiredReplicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}
