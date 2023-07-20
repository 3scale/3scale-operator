package upgrade

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	appsv1 "github.com/openshift/api/apps/v1"
)

// Redis6CommandArgsEnv reconciles environment variables, command and args
func Redis6CommandArgsEnv(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	var updated bool

	desiredName := common.ObjectInfo(desired)

	if len(desired.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers))
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		log.Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating dc", desiredName, len(existing.Spec.Template.Spec.Containers)))
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		updated = true
	}

	// Env Vars added in 2.14
	tmpChanged := helper.EnvVarReconciler(
		desired.Spec.Template.Spec.Containers[0].Env,
		&existing.Spec.Template.Spec.Containers[0].Env,
		"REDIS_CONF")
	updated = updated || tmpChanged

	// Command updated in 2.14
	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Command, desired.Spec.Template.Spec.Containers[0].Command) {
		existing.Spec.Template.Spec.Containers[0].Command = desired.Spec.Template.Spec.Containers[0].Command
		updated = true
	}

	// Args updated in 2.14
	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Args, desired.Spec.Template.Spec.Containers[0].Args) {
		existing.Spec.Template.Spec.Containers[0].Args = desired.Spec.Template.Spec.Containers[0].Args
		updated = true
	}

	return updated, nil
}
