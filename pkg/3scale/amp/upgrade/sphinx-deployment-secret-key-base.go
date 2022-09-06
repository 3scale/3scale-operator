package upgrade

import (
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	appsv1 "github.com/openshift/api/apps/v1"
)

func SphinxSecretKeyEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	//  SECRET_KEY_BASE added
	updated := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "SECRET_KEY_BASE")
	// DELTA_INDEX_INTERVAL removed
	tmp := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "DELTA_INDEX_INTERVAL")
	updated = updated || tmp
	// FULL_REINDEX_INTERVAL removed
	tmp = reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "FULL_REINDEX_INTERVAL")
	updated = updated || tmp
	return updated, nil
}
