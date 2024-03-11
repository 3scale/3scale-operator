package controllers

import (
	"context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RedisConfigCMReconciler struct {
	*reconcilers.BaseReconciler
	previousConfigMap *corev1.ConfigMap
}

// blank assignment to verify that PolicyReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &RedisConfigCMReconciler{}

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps/status,verbs=get;update;patch

func (r *RedisConfigCMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cm := &corev1.ConfigMap{}
	err := r.Client().Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: "redis-config"}, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if r.previousConfigMap == nil {
		r.previousConfigMap = cm.DeepCopy()
		return reconcile.Result{}, nil
	}

	// Check if ConfigMap has changed
	if !isEqualConfigMaps(r.previousConfigMap, cm) {
		err = r.BaseReconciler.ReconcileResource(cm, r.previousConfigMap, r.redisConfigCmMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.previousConfigMap = cm.DeepCopy()
	}

	return ctrl.Result{}, nil
}

func isEqualConfigMaps(cm1 *corev1.ConfigMap, cm2 *corev1.ConfigMap) bool {
	return reflect.DeepEqual(cm1.Data, cm2.Data) &&
		cm1.ObjectMeta.ResourceVersion == cm2.ObjectMeta.ResourceVersion
}

func (r *RedisConfigCMReconciler) redisConfigCmMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", existingObj)
	}
	desired, ok := desiredObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", desiredObj)
	}

	update := false
	fieldUpdated := reconcilers.ConfigMapReconcileField(desired, existing, "redis.conf")
	update = update || fieldUpdated

	return update, nil
}

func (r *RedisConfigCMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
