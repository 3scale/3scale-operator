package upgrader

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "github.com/openshift/api/apps/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpgradeSystemAppContainerEnvs(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	system, err := GetSystemComponent()
	if err != nil {
		return err
	}

	desiredContainers := system.AppDeploymentConfig().Spec.Template.Spec.Containers

	changed := false

	for _, container := range existing.Spec.Template.Spec.Containers {
		for i := range container.Env {
			envVars := []string{
				"SMTP_ADDRESS",
				"SMTP_USER_NAME",
				"SMTP_PASSWORD",
				"SMTP_DOMAIN",
				"SMTP_PORT",
				"SMTP_AUTHENTICATION",
				"SMTP_OPENSSL_VERIFY_MODE"}
			for _, envVar := range envVars {
				if container.Env[i].Name == envVar {
					desiredEnvVar := FindContainerEnvByNameOrPanic(desiredContainers, container.Name, envVar)
					if !reflect.DeepEqual(container.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
						container.Env[i].ValueFrom = desiredEnvVar.ValueFrom
						changed = true
					}
				}
			}
		}
	}

	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}
