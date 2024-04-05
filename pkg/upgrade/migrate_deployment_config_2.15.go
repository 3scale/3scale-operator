package upgrade

import (
	"context"
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

	// Copy any custom labels from the DeploymentConfig to the Deployment
	err = transferDeploymentConfigLabels(deploymentConfig, deployment, client)
	if err != nil {
		return false, fmt.Errorf("error transferring deploymentconfig labels to deployment %s: %v", deployment.Name, err)
	}

	// Copy any custom annotations from the DeploymentConfig to the Deployment
	err = transferDeploymentConfigAnnotations(deploymentConfig, deployment, client)
	if err != nil {
		return false, fmt.Errorf("error transferring deploymentconfig annotations to deployment %s: %v", deployment.Name, err)
	}

	// Copy any custom container env vars from the DeploymentConfig to the Deployment
	err = transferDeploymentConfigEnvVars(deploymentConfig, deployment, client)
	if err != nil {
		return false, fmt.Errorf("error transferring deploymentconfig env vars to deployment %s: %v", deployment.Name, err)
	}

	// Copy any custom volumes and volume mounts from the DeploymentConfig to the Deployment
	err = transferDeploymentConfigVolumes(deploymentConfig, deployment, client)
	if err != nil {
		return false, fmt.Errorf("error transferring deploymentconfig volumes to deployment %s: %v", deployment.Name, err)
	}

	// Update Deployment replica count to match that of the DeploymentConfig
	err = matchDeploymentConfigReplicaCount(deploymentConfig, deployment, client)
	if err != nil {
		return false, fmt.Errorf("error transferring deploymentconfig replica count to deployment %s: %v", deployment.Name, err)
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

func transferDeploymentConfigLabels(dc *appsv1.DeploymentConfig, deployment *k8sappsv1.Deployment, client k8sclient.Client) error {
	// Get the DeploymentConfig-level labels from the DeploymentConfig and Deployment
	dcLabels := dc.ObjectMeta.Labels
	dLabels := deployment.ObjectMeta.Labels

	// Determine which DeploymentConfig-level labels exist in dc but not in deployment
	missingLabels := make(map[string]string)
	for key, value := range dcLabels {
		if _, exists := dLabels[key]; !exists {
			missingLabels[key] = value
		}
	}

	// Add the missing DeploymentConfig-level labels to deployment
	if len(missingLabels) > 0 {
		for key, value := range missingLabels {
			if deployment.ObjectMeta.Labels == nil {
				deployment.ObjectMeta.Labels = map[string]string{}
			}
			deployment.ObjectMeta.Labels[key] = value
		}
	}

	// Extract the Pod-level labels from the DeploymentConfig and Deployment
	dcPodLabels := dc.Spec.Template.Labels
	dPodLabels := deployment.Spec.Template.Labels

	// Determine which Pod-level labels exist in dc but not in deployment
	missingPodLabels := make(map[string]string)
	for key, value := range dcPodLabels {
		if _, exists := dPodLabels[key]; !exists {
			// Skip adding the old deploymentConfig label
			if key != "deploymentConfig" {
				missingPodLabels[key] = value
			}
		}
	}

	// Add any missing Pod-level labels to deployment
	if len(missingPodLabels) > 0 {
		for key, value := range missingPodLabels {
			if deployment.Spec.Template.Labels == nil {
				deployment.Spec.Template.Labels = map[string]string{}
			}
			deployment.Spec.Template.Labels[key] = value
		}
	}

	if len(missingLabels) > 0 || len(missingPodLabels) > 0 {
		// Update the deployment
		err := client.Update(context.TODO(), deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func transferDeploymentConfigAnnotations(dc *appsv1.DeploymentConfig, deployment *k8sappsv1.Deployment, client k8sclient.Client) error {
	// Get the DeploymentConfig-level annotations from the DeploymentConfig and Deployment
	dcAnnotations := dc.ObjectMeta.Annotations
	dAnnotations := deployment.ObjectMeta.Annotations

	// Determine which DeploymentConfig-level annotations exist in dc but not in deployment
	missingAnnotations := make(map[string]string)
	for key, value := range dcAnnotations {
		if _, exists := dAnnotations[key]; !exists {
			missingAnnotations[key] = value
		}
	}

	// Add the missing DeploymentConfig-level annotations to deployment
	if len(missingAnnotations) > 0 {
		for key, value := range missingAnnotations {
			if deployment.ObjectMeta.Annotations == nil {
				deployment.ObjectMeta.Annotations = map[string]string{}
			}
			deployment.ObjectMeta.Annotations[key] = value
		}
	}

	// Extract the Pod-level annotations from the DeploymentConfig and Deployment
	dcPodAnnotations := dc.Spec.Template.Annotations
	dPodAnnotations := deployment.Spec.Template.Annotations

	// Determine which Pod-level annotations exist in dc but not in deployment
	missingPodAnnotations := make(map[string]string)
	for key, value := range dcPodAnnotations {
		if _, exists := dPodAnnotations[key]; !exists {
			missingPodAnnotations[key] = value
		}
	}

	// Add any missing Pod-level annotations to deployment
	if len(missingPodAnnotations) > 0 {
		for key, value := range missingPodAnnotations {
			if deployment.Spec.Template.ObjectMeta.Annotations == nil {
				deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{}
			}
			deployment.Spec.Template.Annotations[key] = value
		}
	}

	if len(missingAnnotations) > 0 || len(missingPodAnnotations) > 0 {
		// Update the deployment
		err := client.Update(context.TODO(), deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func transferDeploymentConfigEnvVars(dc *appsv1.DeploymentConfig, deployment *k8sappsv1.Deployment, client k8sclient.Client) error {
	envVarsAdded := false

	// Loop through each container in dc
	for _, dcContainer := range dc.Spec.Template.Spec.Containers {
		// Find corresponding container in deployment
		var dContainer *corev1.Container
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if dcContainer.Name == container.Name {
				dContainer = &deployment.Spec.Template.Spec.Containers[i]
				break
			}
		}

		if dContainer == nil {
			return fmt.Errorf("the deployment %s is missing the container %s", deployment.Name, dcContainer.Name)
		}

		// Determine which env vars exist in dc's container but not in deployment's container
		var missingEnvVars []corev1.EnvVar
		for _, dcEnvVar := range dcContainer.Env {
			found := false
			for _, deploymentEnvVar := range dContainer.Env {
				if dcEnvVar.Name == deploymentEnvVar.Name {
					found = true
					break
				}
			}
			if !found {
				// Skip adding the REDIS_NAMESPACE env var as it's no longer supported
				if dcEnvVar.Name != "REDIS_NAMESPACE" {
					missingEnvVars = append(missingEnvVars, dcEnvVar)
				}
			}
		}

		// Add any missing env vars to deployment's container
		if len(missingEnvVars) > 0 {
			dContainer.Env = append(dContainer.Env, missingEnvVars...)
			envVarsAdded = true
		}
	}

	if envVarsAdded {
		// Update the deployment
		err := client.Update(context.TODO(), deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func transferDeploymentConfigVolumes(dc *appsv1.DeploymentConfig, deployment *k8sappsv1.Deployment, client k8sclient.Client) error {
	volumesAdded := false

	// Map to store volume names present in deployment
	dVolumes := make(map[string]bool)
	for _, dVolume := range deployment.Spec.Template.Spec.Volumes {
		dVolumes[dVolume.Name] = true
	}

	// Determine which volumes exist in dc but not in deployment
	missingVolumes := make([]corev1.Volume, 0)
	for _, dcVolume := range dc.Spec.Template.Spec.Volumes {
		if _, exists := dVolumes[dcVolume.Name]; !exists {
			missingVolumes = append(missingVolumes, dcVolume)
		}
	}

	// Add missing volumes to Deployment
	if len(missingVolumes) > 0 {
		if deployment.Spec.Template.Spec.Volumes == nil {
			deployment.Spec.Template.Spec.Volumes = make([]corev1.Volume, 0)
		}
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, missingVolumes...)
		volumesAdded = true
	}

	// Loop through each container in dc to check for missing volume mounts
	for _, dcContainer := range dc.Spec.Template.Spec.Containers {
		// Find corresponding container in deployment
		var dContainer *corev1.Container
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if dcContainer.Name == container.Name {
				dContainer = &deployment.Spec.Template.Spec.Containers[i]
				break
			}
		}

		if dContainer == nil {
			return fmt.Errorf("the deployment %s is missing the container %s", deployment.Name, dcContainer.Name)
		}

		// Determine which volume mounts exist in dc's container but not in deployment's container
		var missingVolumeMounts []corev1.VolumeMount
		for _, dcVolumeMount := range dcContainer.VolumeMounts {
			found := false
			for _, deploymentVolumeMount := range dContainer.VolumeMounts {
				if dcVolumeMount.Name == deploymentVolumeMount.Name {
					found = true
					break
				}
			}
			if !found {
				missingVolumeMounts = append(missingVolumeMounts, dcVolumeMount)
			}
		}

		// Add any missing volume mounts to deployment's container
		if len(missingVolumeMounts) > 0 {
			dContainer.VolumeMounts = append(dContainer.VolumeMounts, missingVolumeMounts...)
			volumesAdded = true
		}
	}

	if volumesAdded {
		// Update the deployment
		err := client.Update(context.TODO(), deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func matchDeploymentConfigReplicaCount(dc *appsv1.DeploymentConfig, deployment *k8sappsv1.Deployment, client k8sclient.Client) error {
	// Get the replica count from the DeploymentConfig and Deployment
	dcReplicaCount := dc.Spec.Replicas
	dReplicaCount := deployment.Spec.Replicas

	if *dReplicaCount != dcReplicaCount {
		deployment.Spec.Replicas = &dcReplicaCount
		// Update the deployment
		err := client.Update(context.TODO(), deployment)
		if err != nil {
			return err
		}
	}

	return nil
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
