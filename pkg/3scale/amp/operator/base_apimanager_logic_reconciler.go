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
	apiManager                      *appsv1alpha1.APIManager
	deploymentConfigReconciler      DeploymentConfigReconciler
	imagestreamReconciler           ImageStreamReconciler
	serviceAccountReconciler        ServiceAccountReconciler
	secretReconciler                SecretReconciler
	serviceReconciler               ServiceReconciler
	configMapReconciler             ConfigMapReconciler
	routeReconciler                 RouteReconciler
	persistentVolumeClaimReconciler PersistentVolumeClaimReconciler
	roleReconciler                  RoleReconciler
	roleBindingReconciler           RoleBindingReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BaseAPIManagerLogicReconciler{}

func NewBaseAPIManagerLogicReconciler(b BaseLogicReconciler, apiManager *appsv1alpha1.APIManager) BaseAPIManagerLogicReconciler {
	objectMetaMerger := NewObjectMetaMerger(b.BaseReconciler, apiManager)
	deploymentConfigReconciler := NewDeploymentConfigReconciler(b.BaseReconciler, objectMetaMerger)
	imageStreamReconciler := NewImageStreamReconciler(b.BaseReconciler, objectMetaMerger)
	serviceAccountReconciler := NewServiceAccountReconciler(b.BaseReconciler, objectMetaMerger)
	secretReconciler := NewSecretReconciler(b.BaseReconciler, objectMetaMerger)
	serviceReconciler := NewServiceReconciler(b.BaseReconciler, objectMetaMerger)
	configMapReconciler := NewConfigMapReconciler(b.BaseReconciler, objectMetaMerger)
	routeReconciler := NewRouteReconciler(b.BaseReconciler, objectMetaMerger)
	persistentVolumeClaimReconciler := NewPersistentVolumeClaimReconciler(b.BaseReconciler, objectMetaMerger)
	roleReconciler := NewRoleReconciler(b.BaseReconciler, objectMetaMerger)
	roleBindingReconciler := NewRoleBindingReconciler(b.BaseReconciler, objectMetaMerger)

	return BaseAPIManagerLogicReconciler{
		BaseLogicReconciler:             b,
		apiManager:                      apiManager,
		deploymentConfigReconciler:      deploymentConfigReconciler,
		imagestreamReconciler:           imageStreamReconciler,
		serviceAccountReconciler:        serviceAccountReconciler,
		secretReconciler:                secretReconciler,
		serviceReconciler:               serviceReconciler,
		configMapReconciler:             configMapReconciler,
		routeReconciler:                 routeReconciler,
		persistentVolumeClaimReconciler: persistentVolumeClaimReconciler,
		roleReconciler:                  roleReconciler,
		roleBindingReconciler:           roleBindingReconciler,
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
