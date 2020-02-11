package upgrader

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MigrateSystemSMTPData(cl client.Client, ns string) error {
	existingConfigMap := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "smtp", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, existingConfigMap)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		// has been deleted, nothing to do
		return nil
	}

	existingSecret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{Name: component.SystemSecretSystemSMTPSecretName, Namespace: ns}
	err = cl.Get(context.TODO(), secretNamespacedName, existingSecret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		system, err := GetSystemComponent()
		if err != nil {
			return err
		}
		existingSecret = system.SMTPSecret()
		existingSecret.SetNamespace(ns)
		// We make sure StringData is nil so it does not get precedence over Data.
		// We use Data to set the secret and not StringData due to at the time
		// of writing this comment when using the Kubernetes FakeClient the
		// mocked client does not convert from StringData to Data, producing a
		// different behavior than with the real code execution
		existingSecret.StringData = nil
		existingSecret.Data = helper.GetSecretDataFromStringData(existingConfigMap.Data)

		err = cl.Create(context.TODO(), existingSecret)
		if err != nil {
			return err
		}
		fmt.Printf("Created object %s\n", operator.ObjectInfo(existingSecret))
	} else {
		changed := false
		secretStringData := helper.GetSecretStringDataFromData(existingSecret.Data)
		changed = !reflect.DeepEqual(existingConfigMap.Data, secretStringData)
		if changed {
			existingSecret.Data = helper.GetSecretDataFromStringData(existingConfigMap.Data)
			err := cl.Update(context.TODO(), existingSecret)
			if err != nil {
				return err
			}
			fmt.Printf("Update object %s\n", operator.ObjectInfo(existingSecret))
		}
	}

	return nil
}
