package upgrade

import (
	appsv1 "github.com/openshift/api/apps/v1"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// SystemBackendUrls reconciles environment variables for Backend URLs on system DC
func SystemBackendUrls(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	var changed bool

	// Remove old env var
	oldVarRemoved := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "BACKEND_ROUTE")

	// Add new env vars
	newVar1Added := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "BACKEND_URL")
	newVar2Added := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "BACKEND_PUBLIC_URL")
	changed = oldVarRemoved || newVar1Added || newVar2Added

	return changed, nil
}
