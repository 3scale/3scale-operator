package upgrader

import (
	"context"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteSMTPConfigMap(cl client.Client, ns string) error {
	existing := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "smtp", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, existing)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if !k8serrors.IsNotFound(err) {
		err := cl.Delete(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}
