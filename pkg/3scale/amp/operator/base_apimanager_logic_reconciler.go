package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

func (r BaseAPIManagerLogicReconciler) InitializeAsAPIManagerObject(obj common.KubernetesObject) error {
	namespacedName := r.NamespacedNameWithAPIManagerNamespace(obj)
	obj.SetNamespace(namespacedName.Namespace)
	err := r.setOwnerReference(obj)
	if err != nil {
		return err
	}
	return nil
}

func (r BaseAPIManagerLogicReconciler) setOwnerReference(obj common.KubernetesObject) error {
	err := controllerutil.SetControllerReference(r.apiManager, obj, r.Scheme())
	if err != nil {
		r.Logger().Error(err, "Error setting OwnerReference on object",
			"Kind", obj.GetObjectKind().GroupVersionKind().String(),
			"Namespace", obj.GetNamespace(),
			"Name", obj.GetName(),
		)
	}
	return err
}
