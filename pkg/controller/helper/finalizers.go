package helper

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

/*
ReconcileFinalizers reconciles the finalizers, it requires
- object
- k8client
- finalizer
If the deletion timestamp is found, the finalizer will be removed.
If the deletion timestamp is not present, a finalizer will be reconciled
*/
func ReconcileFinalizers(object controllerutil.Object, client client.Client, finalizer string) error {
	var err error

	if object.GetDeletionTimestamp() == nil {
		_, err = controllerutil.CreateOrUpdate(context.TODO(), client, object, func() error {
			controllerutil.AddFinalizer(object, finalizer)
			return nil
		})
	} else {
		_, err = controllerutil.CreateOrUpdate(context.TODO(), client, object, func() error {
			controllerutil.RemoveFinalizer(object, finalizer)
			return nil
		})
	}

	if err != nil {
		return err
	}
	return nil
}
