package helper

import (
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// IsDeploymentAvailable returns true when the provided Deployment has the "Available" condition set to true
func IsDeploymentAvailable(d *k8sappsv1.Deployment) bool {
	dConditions := d.Status.Conditions
	for _, dCondition := range dConditions {
		if dCondition.Type == k8sappsv1.DeploymentAvailable && dCondition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsDeploymentProgressing returns true when the provided Deployment is progressing with new ReplicaSet
func IsDeploymentProgressing(d *k8sappsv1.Deployment) bool {
	if d.Status.UnavailableReplicas > 0 {
		return true
	}

	for _, dCondition := range d.Status.Conditions {
		if dCondition.Type == k8sappsv1.DeploymentProgressing && dCondition.Reason == "ReplicaSetUpdated" {
			return true
		}
	}

	return false
}
