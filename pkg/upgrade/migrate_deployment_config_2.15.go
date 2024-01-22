package upgrade

import (
	"context"
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
)

// MigrateDeploymentConfigToDeployment verifies the Deployment is healthy and then deletes the corresponding DeploymentConfig
// 3scale 2.14 -> 2.15
func MigrateDeploymentConfigToDeployment(dName string, dNamespace string, overrideDeploymentHealth bool, client k8sclient.Client, scheme *runtime.Scheme) (bool, error) {
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

	// Check Deployment health
	deployment := &k8sappsv1.Deployment{}
	err = client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: dNamespace,
		Name:      dName,
	}, deployment)

	// Return error if can't get Deployment
	if err != nil && !k8serr.IsNotFound(err) {
		return false, fmt.Errorf("error getting deployment %s: %w", deployment.Name, err)
	}

	// Requeue if Deployment doesn't exist yet
	if k8serr.IsNotFound(err) {
		log.V(1).Info(fmt.Sprintf("deployment %s does not exist", deployment.Name))
		return false, nil
	}

	// Requeue if Deployment isn't healthy and override is set to false, otherwise proceed
	if !helper.IsDeploymentAvailable(deployment) && !overrideDeploymentHealth {
		log.V(1).Info(fmt.Sprintf("deployment %s is not yet available and overrideDeploymentHealth is set to %t", deployment.Name, overrideDeploymentHealth))
		return false, nil
	}

	// Transfer zync-que route ownership
	if deployment.Name == component.ZyncQueDeploymentName && scheme != nil {
		err = transferZyncRoutesOwnership(deployment, client, scheme)
		if err != nil {
			return false, err
		}
	}

	// Delete the DeploymentConfig because the Deployment replacing it is healthy or overrideDeploymentHealth is set to true
	err = client.Delete(context.TODO(), deploymentConfig)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return false, fmt.Errorf("error deleting deploymentconfig %s: %v", deploymentConfig.Name, err)
		}
	}

	log.Info(fmt.Sprintf("%s Deployment has replaced its corresponding DeploymentConfig", deployment.Name))
	return true, nil
}

func transferZyncRoutesOwnership(deployment *k8sappsv1.Deployment, client k8sclient.Client, scheme *runtime.Scheme) error {
	listOps := []k8sclient.ListOption{
		k8sclient.InNamespace(deployment.Namespace),
	}

	routeList := &routev1.RouteList{}
	err := client.List(context.TODO(), routeList, listOps...)
	if err != nil {
		return fmt.Errorf("failed to list routes to transfer ownership: %w", err)
	}

	for _, rt := range routeList.Items {
		for _, ownerRef := range rt.ObjectMeta.OwnerReferences {
			if ownerRef.Name == component.ZyncQueDeploymentName {
				err = controllerutil.SetOwnerReference(deployment, &rt, scheme)
				if err != nil {
					return fmt.Errorf("failed to set ownerRef for route %v : %w", rt, err)
				}

				err = client.Update(context.TODO(), &rt)
				if err != nil {
					return fmt.Errorf("failed to update ownerRef for route %v : %w", rt, err)
				}
			}
		}
	}

	return nil
}
