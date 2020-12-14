package helper

import (
	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// IsDeploymentConfigAvailable returns true when the provided DeploymentConfig
// has the "Available" condition set to true
func IsDeploymentConfigAvailable(dc *appsv1.DeploymentConfig) bool {
	dcConditions := dc.Status.Conditions
	for _, dcCondition := range dcConditions {
		if dcCondition.Type == appsv1.DeploymentAvailable && dcCondition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
