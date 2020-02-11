package upgrader

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MigrateS3Deployment(cl client.Client, ns string) error {
	existingSecret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{Name: "aws-auth", Namespace: ns}
	err := cl.Get(context.TODO(), secretNamespacedName, existingSecret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		// S3 deployment not found
		// Maybe there is a better automatic way to detect S3 deployment
		return nil
	}

	err = migrateS3Secret(cl, ns)
	if err != nil {
		return err
	}

	err = migrateS3SystemAppDC(cl, ns)
	if err != nil {
		return err
	}

	err = migrateS3SystemSidekiqDC(cl, ns)
	if err != nil {
		return err
	}

	err = deleteUnusedS3DataConfigmap(cl, ns)
	if err != nil {
		return err
	}

	return nil
}

func migrateS3Secret(cl client.Client, ns string) error {
	existingSecret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{Name: "aws-auth", Namespace: ns}
	err := cl.Get(context.TODO(), secretNamespacedName, existingSecret)
	if err != nil {
		return err
	}

	existingConfigMap := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "system-environment", Namespace: ns}
	err = cl.Get(context.TODO(), configMapNamespacedName, existingConfigMap)
	if err != nil {
		return err
	}

	changed := false
	if existingSecret.StringData == nil {
		existingSecret.StringData = map[string]string{}
	}
	secretData := helper.GetSecretStringDataFromData(existingSecret.Data)
	if _, exists := secretData["AWS_BUCKET"]; !exists {
		existingSecret.StringData["AWS_BUCKET"] = existingConfigMap.Data["AWS_BUCKET"]
		changed = true
	}

	if _, exists := secretData["AWS_REGION"]; !exists {
		existingSecret.StringData["AWS_REGION"] = existingConfigMap.Data["AWS_REGION"]
		changed = true
	}

	if changed {
		err := cl.Update(context.TODO(), existingSecret)
		if err != nil {
			return err
		}
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existingSecret))
	}

	return nil
}

func migrateS3SystemAppDC(cl client.Client, ns string) error {
	changed := false
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	system, err := GetS3SystemComponent()
	desired := system.AppDeploymentConfig()

	// pre hook: replace AWS_BUCKET, AWS_REGION vars
	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPreHookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod
	for i := range existingPreHookPod.Env {
		envVars := []string{"AWS_BUCKET", "AWS_REGION"}
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
	// pre hook: add new vars
	envVars := []string{"AWS_PROTOCOL", "AWS_HOSTNAME", "AWS_PATH_STYLE"}
	for _, envVar := range envVars {
		if _, ok := FindEnvByName(existingPreHookPod.Env, envVar); !ok {
			existingPreHookPod.Env = append(existingPreHookPod.Env, FindEnvByNameOrPanic(desiredPreHookPod.Env, envVar))
			changed = true
		}
	}

	// containers env: replace AWS_BUCKET, AWS_REGION vars
	desiredContainers := desired.Spec.Template.Spec.Containers
	for _, container := range existing.Spec.Template.Spec.Containers {
		for i := range container.Env {
			envVars := []string{"AWS_BUCKET", "AWS_REGION"}
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
	// containers env: add new vars
	for i, container := range existing.Spec.Template.Spec.Containers {
		envVars := []string{"AWS_PROTOCOL", "AWS_HOSTNAME", "AWS_PATH_STYLE"}
		for _, envVar := range envVars {
			if _, ok := FindEnvByName(existing.Spec.Template.Spec.Containers[i].Env, envVar); !ok {
				desiredEnvVar := FindContainerEnvByNameOrPanic(desiredContainers, container.Name, envVar)
				existing.Spec.Template.Spec.Containers[i].Env = append(existing.Spec.Template.Spec.Containers[i].Env, desiredEnvVar)
				changed = true
			}
		}
	}

	// update on any change
	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateS3SystemSidekiqDC(cl client.Client, ns string) error {
	changed := false
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-sidekiq", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	system, err := GetS3SystemComponent()
	desired := system.SidekiqDeploymentConfig()

	// containers env: replace AWS_BUCKET, AWS_REGION vars
	desiredContainers := desired.Spec.Template.Spec.Containers
	for _, container := range existing.Spec.Template.Spec.Containers {
		for i := range container.Env {
			envVars := []string{"AWS_BUCKET", "AWS_REGION"}
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

	// containers env: add new vars
	for i, container := range existing.Spec.Template.Spec.Containers {
		envVars := []string{"AWS_PROTOCOL", "AWS_HOSTNAME", "AWS_PATH_STYLE"}
		for _, envVar := range envVars {
			if _, ok := FindEnvByName(existing.Spec.Template.Spec.Containers[i].Env, envVar); !ok {
				desiredEnvVar := FindContainerEnvByNameOrPanic(desiredContainers, container.Name, envVar)
				existing.Spec.Template.Spec.Containers[i].Env = append(existing.Spec.Template.Spec.Containers[i].Env, desiredEnvVar)
				changed = true
			}
		}
	}

	// update on any change
	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteUnusedS3DataConfigmap(cl client.Client, ns string) error {
	changed := false
	existing := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "system-environment", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, existing)
	if err != nil {
		return err
	}

	keys := []string{"AWS_BUCKET", "AWS_REGION"}
	for _, key := range keys {
		if _, exists := existing.Data[key]; exists {
			delete(existing.Data, key)
			changed = true
		}
	}

	// update on any change
	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetS3SystemComponent() (*component.System, error) {
	optProv := component.SystemOptionsBuilder{}
	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AmpRelease("-")
	optProv.ApicastRegistryURL("-")
	optProv.TenantName("-")
	optProv.WildcardDomain("-")
	optProv.AdminAccessToken("-")
	optProv.AdminPassword("-")
	optProv.AdminUsername("-")
	optProv.ApicastAccessToken("-")
	optProv.MasterAccessToken("-")
	optProv.MasterUsername("-")
	optProv.MasterPassword("-")
	optProv.AppSecretKeyBase("-")
	optProv.BackendSharedSecret("-")
	optProv.MasterName("-")
	optProv.S3FileStorageOptions(component.S3FileStorageOptions{
		ConfigurationSecretName: "aws-auth",
	})
	systemOptions, err := optProv.Build()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(systemOptions), nil
}
