package helper

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
This function reconciles the finalizers, it requires
- object
- k8client
- finalizer
If the deletion timestamp is found, the finalizer will be removed.
If the deletion timestamp is not present, a finalizer will be reconciled
*/
func ReconcileFinalizers(object Object, client client.Client, finalizer string) error {
	if object.GetDeletionTimestamp() == nil {
		_, err := CreateOrUpdate(context.TODO(), client, object, func() error {
			AddFinalizer(object, finalizer)
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		_, err := CreateOrUpdate(context.TODO(), client, object, func() error {
			RemoveFinalizer(object, finalizer)
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
