package upgrade

import (
	appsv1 "github.com/openshift/api/apps/v1"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// DeploymentConfigPodTemplateLabelsMutator ensures pod template labels are reconciled
func SphinxAddressReference(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	var changed bool
	changed = reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "THINKING_SPHINX_ADDRESS")

	// Not in desired, it should be removed from existing
	tmpChanged := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "THINKING_SPHINX_CONFIGURATION_FILE")
	changed = changed || tmpChanged

	return changed, nil
}
