package e2eutil

import (
	"context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/test"
	"k8s.io/apimachinery/pkg/types"
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

func WaitForSecret(t *testing.T, kubeClient kubernetes.Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		_, secretErr := kubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		if secretErr != nil {
			if apierrors.IsNotFound(secretErr) {
				t.Logf("Waiting for availability of secret '%s'\n", name)
				return false, nil
			}
			return false, secretErr
		}

		t.Logf("Secret [%s] available\n", name)
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func WaitForConsolidated(t *testing.T, client test.FrameworkClient, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		consolidated := v1alpha1.Consolidated{}
		err = client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name,}, &consolidated)
		if err != nil {
			t.Logf("Waiting for consolidated object\n")
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func WaitForReconciliationWith3scale(t *testing.T, existingConsolidated v1alpha1.Consolidated, retryInterval, timeout time.Duration) error {

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {

		desiredConsolidated, err := v1alpha1.NewConsolidatedFrom3scale(existingConsolidated.Spec.Credentials, existingConsolidated.Spec.APIs)
		if err != nil {
			t.Fatal(err)
		}
		if !v1alpha1.CompareConsolidated(existingConsolidated, *desiredConsolidated) {
			t.Log("Consolidated object is not yet reconcile, retrying.")
			return false, fmt.Errorf("Reconciliation is not finished")
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil

}
