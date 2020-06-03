package apimanagerrestore

import (
	"reflect"
	"sort"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileWaitForAPIManagerReadyStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileWaitForAPIManagerReadyStep) Execute() (reconcile.Result, error) {
	// TODO NOOP because in this case the step itself is just check for completion???
	return reconcile.Result{}, nil
}

func (r *ReconcileWaitForAPIManagerReadyStep) Completed() (bool, error) {
	existingAPIManager := &appsv1alpha1.APIManager{}
	err := r.GetResource(types.NamespacedName{Name: r.cr.Status.APIManagerToRestoreRef.Name, Namespace: r.cr.Namespace}, existingAPIManager)
	if err != nil {
		return false, err
	}

	// External databases scenario assumed
	expectedDeploymentNames := []string{
		"apicast-production",
		"apicast-staging",
		"backend-listener",
		"backend-worker",
		"backend-cron",
		"zync",
		"zync-que",
		"zync-database",
		"system-app",
		"system-sphinx",
		"system-sidekiq",
		"system-memcache",
	}

	existingReadyDeployments := existingAPIManager.Status.Deployments.Ready
	sort.Slice(expectedDeploymentNames, func(i, j int) bool { return expectedDeploymentNames[i] < expectedDeploymentNames[j] })
	sort.Slice(existingReadyDeployments, func(i, j int) bool { return existingReadyDeployments[i] < existingReadyDeployments[j] })

	if !reflect.DeepEqual(existingReadyDeployments, expectedDeploymentNames) {
		r.Logger().Info("all APIManager Deployments not ready. Waiting", "APIManager", existingAPIManager.Name, "expected-ready-deployments", expectedDeploymentNames, "ready-deployments", existingReadyDeployments)
		// TODO here we should wait several seconds and not directly requeue, however the Completed function does not have
		// a reconcile.Result return parameter. How do we organize this?
		return false, nil
	}

	return true, nil
}

func (r *ReconcileWaitForAPIManagerReadyStep) Identifier() string {
	return "ReconcileWaitForAPIManagerReadyStep"
}
