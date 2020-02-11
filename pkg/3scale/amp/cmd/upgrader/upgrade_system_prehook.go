package upgrader

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpgradeSystemPreHook(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	system, err := GetSystemComponent()
	if err != nil {
		return err
	}

	desiredDeploymentConfig := system.AppDeploymentConfig()
	changed := ensureDeploymentConfigPreHookPodEnvVars(desiredDeploymentConfig, existing)
	tmpChanged := ensureDeploymentConfigPreHookPodCommand(desiredDeploymentConfig, existing)
	changed = changed || tmpChanged

	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}

func ensureDeploymentConfigPreHookPodEnvVars(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false
	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPreHookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod

	// replace SMTP_* vars
	for i := range existingPreHookPod.Env {
		envVars := []string{
			"SMTP_ADDRESS",
			"SMTP_USER_NAME",
			"SMTP_PASSWORD",
			"SMTP_DOMAIN",
			"SMTP_PORT",
			"SMTP_AUTHENTICATION",
			"SMTP_OPENSSL_VERIFY_MODE"}
		for _, envVar := range envVars {
			if existingPreHookPod.Env[i].Name == envVar {
				desiredEnvVar := FindEnvByNameOrPanic(desiredPreHookPod.Env, envVar)
				if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
					existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
					changed = true
				}
			}
		}
	}

	if _, ok := FindEnvByName(existingPreHookPod.Env, component.SystemSecretSystemSeedMasterAccessTokenFieldName); !ok {
		// Add MASTER_ACCESS_TOKEN ref
		existingPreHookPod.Env = append(existingPreHookPod.Env,
			v1.EnvVar{
				Name: component.SystemSecretSystemSeedMasterAccessTokenFieldName,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: component.SystemSecretSystemSeedSecretName,
						},
						Key: component.SystemSecretSystemSeedMasterAccessTokenFieldName,
					},
				},
			})
		changed = true
	}

	return changed
}

func ensureDeploymentConfigPreHookPodCommand(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false
	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPrehookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod
	if !reflect.DeepEqual(existingPrehookPod.Command, desiredPreHookPod.Command) {
		existingPrehookPod.Command = desiredPreHookPod.Command
		changed = true
	}
	return changed
}
