package upgrade

import (
	appsv1 "github.com/openshift/api/apps/v1"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// SystemBackendUrls reconciles environment variables for Backend URLs on system DC
func SystemBackendUrls(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	var changed bool

	// Remove old env var
	tmpChanged := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_BACKEND_ROOT_ENDPOINT")
	changed = changed || tmpChanged

	// Remove old env var
	tmpChanged = reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "BACKEND_ROUTE")
	changed = changed || tmpChanged

	// Add new env vars
	tmpChanged = reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "BACKEND_URL")
	changed = changed || tmpChanged

	// Add new env vars
	tmpChanged = reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "BACKEND_PUBLIC_URL")
	changed = changed || tmpChanged

	return changed, nil
}
