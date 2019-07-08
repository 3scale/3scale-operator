package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseAPIManagerLogicReconciler struct {
	BaseLogicReconciler
	apiManager *appsv1alpha1.APIManager
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BaseAPIManagerLogicReconciler{}

func NewBaseAPIManagerLogicReconciler(b BaseLogicReconciler, apiManager *appsv1alpha1.APIManager) BaseAPIManagerLogicReconciler {
	return BaseAPIManagerLogicReconciler{
		BaseLogicReconciler: b,
		apiManager:          apiManager,
	}
}

func (r BaseAPIManagerLogicReconciler) Reconcile() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r BaseAPIManagerLogicReconciler) NamespacedNameWithAPIManagerNamespace(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: r.apiManager.GetNamespace(), Name: obj.GetName()}
}
