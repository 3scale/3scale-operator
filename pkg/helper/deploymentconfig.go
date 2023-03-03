package helper

import (
	"errors"

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

func FindDeploymentTriggerOnImageChange(triggerPolicies []appsv1.DeploymentTriggerPolicy) (int, error) {
	result := -1
	for i := range triggerPolicies {
		if triggerPolicies[i].Type == appsv1.DeploymentTriggerOnImageChange {
			result = i
			break
		}
	}

	if result == -1 {
		return -1, errors.New("no imageChangeParams deployment trigger policy found")
	}

	return result, nil
}

func IsDeploymentConfigDeleting(dc *appsv1.DeploymentConfig) bool {
	return dc.GetDeletionTimestamp() != nil
}
