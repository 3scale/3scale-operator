package reconcilers

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("reconcilers")

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func(existing, desired common.KubernetesObject) (bool, error)

func CreateOnlyMutator(existing, desired common.KubernetesObject) (bool, error) {
	return false, nil
}

type BaseReconciler struct {
	client          client.Client
	scheme          *runtime.Scheme
	apiClientReader client.Reader
	ctx             context.Context
	logger          logr.Logger
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &BaseReconciler{}

func NewBaseReconciler(client client.Client, scheme *runtime.Scheme, apiClientReader client.Reader,
	ctx context.Context, logger logr.Logger) *BaseReconciler {
	return &BaseReconciler{
		client:          client,
		scheme:          scheme,
		apiClientReader: apiClientReader,
		ctx:             ctx,
		logger:          logger,
	}
}

func (b *BaseReconciler) Reconcile(reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// Client returns a split client that reads objects from
// the cache and writes to the Kubernetes APIServer
func (b *BaseReconciler) Client() client.Client {
	return b.client
}

// APIClientReader return a client that directly reads objects
// from the Kubernetes APIServer
func (b *BaseReconciler) APIClientReader() client.Reader {
	return b.apiClientReader
}

func (b *BaseReconciler) Scheme() *runtime.Scheme {
	return b.scheme
}

func (b *BaseReconciler) Logger() logr.Logger {
	return b.logger
}

// ReconcileResource attempts to mutate the existing state
// in order to match the desired state. The object's desired state must be reconciled
// with the existing state inside the passed in callback MutateFn.
//
// obj: Object of the same type as the 'desired' object.
//            Used to read the resource from the kubernetes cluster.
//            Could be zero-valued initialized object.
// desired: Object representing the desired state
//
// It returns an error.
func (b *BaseReconciler) ReconcileResource(obj, desired common.KubernetesObject, mutateFn MutateFn) error {
	key, err := client.ObjectKeyFromObject(desired)
	if err != nil {
		return err
	}

	if err = b.Client().Get(b.ctx, key, obj); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// Not found
		if !common.IsObjectTaggedToDelete(desired) {
			return b.CreateResource(desired)
		}

		// Marked for deletion and not found. Nothing to do.
		return nil
	}

	// item found successfully
	if common.IsObjectTaggedToDelete(desired) {
		return b.DeleteResource(desired)
	}

	update, err := mutateFn(obj, desired)
	if err != nil {
		return err
	}

	if update {
		return b.UpdateResource(obj)
	}

	return nil
}

func (b *BaseReconciler) CreateResource(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Created object %s", common.ObjectInfo(obj)))
	return b.Client().Create(b.ctx, obj)
}

func (b *BaseReconciler) UpdateResource(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Updated object %s", common.ObjectInfo(obj)))
	return b.Client().Update(b.ctx, obj)
}

func (b *BaseReconciler) DeleteResource(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Delete object %s", common.ObjectInfo(obj)))
	return b.Client().Delete(context.TODO(), obj)
}

func (b *BaseReconciler) UpdateResourceStatus(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Updated status of object %s", common.ObjectInfo(obj)))
	return b.Client().Status().Update(context.TODO(), obj)
}
