package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type S3Reconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &S3Reconciler{}

func NewS3Reconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) S3Reconciler {
	return S3Reconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *S3Reconciler) Reconcile() (reconcile.Result, error) {
	S3, err := r.S3()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileAWSSecret(S3.S3AWSSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *S3Reconciler) S3() (*component.S3, error) {
	optsProvider := OperatorS3OptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetS3Options()
	if err != nil {
		return nil, err
	}
	return component.NewS3(opts), nil
}

func (r *S3Reconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}
	return r.secretReconciler.Reconcile(desiredSecret)
}

func (r *S3Reconciler) reconcileAWSSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}
