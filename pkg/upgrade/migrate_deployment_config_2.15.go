package upgrade

import (
	"context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/helper"
	appsv1 "github.com/openshift/api/apps/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MigrateDeploymentConfigToDeployment verifies the Deployment is healthy and then deletes the corresponding DeploymentConfig
// 3scale 2.14 -> 2.15
func MigrateDeploymentConfigToDeployment(dName string, dNamespace string, client k8sclient.Client) (bool, error) {
	deploymentConfig := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dName,
			Namespace: dNamespace,
		},
	}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{Name: deploymentConfig.Name, Namespace: deploymentConfig.Namespace}, deploymentConfig)

	// Breakout if the DeploymentConfig has already been deleted
	if k8serr.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("error getting deploymentconfig %s: %v", deploymentConfig.Name, err)
	}

	// Verify that the Deployment is healthy
	deployment := &k8sappsv1.Deployment{}
	err = client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: dNamespace,
		Name:      dName,
	}, deployment)
	if err != nil {
		return false, fmt.Errorf("error getting deployment %s: %v", deployment.Name, err)
	}
	if !helper.IsDeploymentAvailable(deployment) {
		log.V(1).Info(fmt.Sprintf("deployment %s is not yet available", deployment.Name))
		return false, nil
	}

	// Delete the DeploymentConfig because the Deployment replacing it is healthy
	err = client.Delete(context.TODO(), deploymentConfig)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return false, fmt.Errorf("error deleting deploymentconfig %s: %v", deploymentConfig.Name, err)
		}
	}

	log.Info(fmt.Sprintf("%s Deployment has replaced its corresponding DeploymentConfig", deployment.Name))
	return true, nil
}
