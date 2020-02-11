package upgrader

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Maybe there is a better automatic way to detect S3 deployment
func checkS3Deployment(cl client.Client, ns string) (bool, error) {
	secret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{Name: "aws-auth", Namespace: ns}
	err := cl.Get(context.TODO(), secretNamespacedName, secret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, err
	}

	return !k8serrors.IsNotFound(err), nil
}

func VerifyUpgrade(cl client.Client, ns string) bool {
	validations := []struct {
		ValidateFunc func(client.Client, string) error
		TestName     string
	}{
		{verifySMTPSecret, "SMTP secret exists"},
		{verifySystemAppPreHookPodCommand, "System app pre hook pod command upgraded"},
		{verifySystemAppPreHookPodEnv, "System app pre hook pod env upgraded"},
		{verifySystemAppContainersEnv, "System app containers env upgraded"},
		{verifySystemSidekiqContainersEnv, "System app containers env upgraded"},
		{verifyAWSSecret, "AWS secret upgraded"},
		{verifyAWSInfoDeletedFromConfigMap, "AWS info deleted from configmap"},
		{verifyAMPRelease, "AMP_RELEASE patched"},
		{verifyImages, "Images upgraded"},
		{verifySMTPConfigMapDeleted, "old SMTP configmap deleted"},
	}

	failed := false

	fmt.Println("########\n## Check upgrade\n########")
	for _, validation := range validations {
		err := validation.ValidateFunc(cl, ns)
		if err != nil {
			fmt.Printf("\033[1;31m[FAIL]\033[0m %s: %v\n", validation.TestName, err)
			failed = true
		} else {
			fmt.Printf("\033[1;32m[OK]\033[0m %s\n", validation.TestName)
		}
	}

	return !failed
}

func verifyImagestream(is *imagev1.ImageStream) bool {
	tag28found := false
	latestTagUpgraded := false
	for _, tag := range is.Spec.Tags {
		if tag.Name == "2.8" {
			tag28found = true
		}

		if tag.Name == "latest" &&
			tag.From != nil &&
			tag.From.Name == "2.8" {
			latestTagUpgraded = true
		}
	}

	return tag28found && latestTagUpgraded
}

func verifyImages(cl client.Client, ns string) error {
	for _, imageName := range THREESCALEIMAGES {
		is := &imagev1.ImageStream{}
		err := cl.Get(context.TODO(), types.NamespacedName{Name: imageName, Namespace: ns}, is)
		if err != nil {
			return err
		}

		if !verifyImagestream(is) {
			return fmt.Errorf("%s is not upgraded", imageName)
		}
	}

	// optional imagestreams
	for _, imageName := range []string{"system-mysql", "system-postgresql"} {
		is := &imagev1.ImageStream{}
		err := cl.Get(context.TODO(), types.NamespacedName{Name: imageName, Namespace: ns}, is)
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if err == nil && !verifyImagestream(is) {
			return fmt.Errorf("%s is not upgraded", imageName)
		}
	}

	return nil
}

func verifyAMPRelease(cl client.Client, ns string) error {
	configMap := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "system-environment", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, configMap)
	if err != nil {
		return err
	}

	if configMap.Data["AMP_RELEASE"] != "2.8" {
		return errors.New("system-environment AMP_RELEASE key not patched")
	}
	return nil
}

func verifyAWSInfoDeletedFromConfigMap(cl client.Client, ns string) error {
	configMap := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "system-environment", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, configMap)
	if err != nil {
		return err
	}

	if _, ok := configMap.Data["AWS_BUCKET"]; ok {
		return errors.New("system-environment configmap still has AWS_BUCKET key")
	}

	if _, ok := configMap.Data["AWS_REGION"]; ok {
		return errors.New("system-environment configmap still has AWS_REGION key")
	}

	return nil
}

func verifyAWSSecret(cl client.Client, ns string) error {
	s3Deployment, err := checkS3Deployment(cl, ns)
	if err != nil {
		return err
	}

	if !s3Deployment {
		return nil
	}

	secret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{Name: "aws-auth", Namespace: ns}
	err = cl.Get(context.TODO(), secretNamespacedName, secret)
	if err != nil {
		return err
	}

	for _, envVarName := range []string{"AWS_BUCKET", "AWS_REGION"} {
		if _, ok := secret.Data[envVarName]; !ok {
			return fmt.Errorf("%s key not found aws secret", envVarName)
		}
	}

	return nil
}

func verifySystemSidekiqContainersEnv(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-sidekiq", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	for _, container := range existing.Spec.Template.Spec.Containers {
		for _, envVarName := range SMTPVARS {
			envVar, ok := FindEnvByName(container.Env, envVarName)
			if !ok {
				return fmt.Errorf("%s env var not found in container %s", envVarName, container.Name)
			}

			if envVar.ValueFrom == nil || envVar.ValueFrom.SecretKeyRef == nil {
				return fmt.Errorf("%s env var from container %s is not reference to secret", envVarName, container.Name)
			}
		}
	}

	s3Deployment, err := checkS3Deployment(cl, ns)
	if err != nil {
		return err
	}

	if s3Deployment {
		for _, container := range existing.Spec.Template.Spec.Containers {
			for _, envVarName := range AWSVARS {
				envVar, ok := FindEnvByName(container.Env, envVarName)
				if !ok {
					return fmt.Errorf("%s env var not found in container %s", envVarName, container.Name)
				}

				if envVar.ValueFrom == nil || envVar.ValueFrom.SecretKeyRef == nil {
					return fmt.Errorf("%s env var from container %s is not reference to secret", envVarName, container.Name)
				}
			}
		}
	}

	return nil

}

func verifySystemAppContainersEnv(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	for _, container := range existing.Spec.Template.Spec.Containers {
		for _, envVarName := range SMTPVARS {
			envVar, ok := FindEnvByName(container.Env, envVarName)
			if !ok {
				return fmt.Errorf("%s env var not found in container %s", envVarName, container.Name)
			}

			if envVar.ValueFrom == nil || envVar.ValueFrom.SecretKeyRef == nil {
				return fmt.Errorf("%s env var from container %s is not reference to secret", envVarName, container.Name)
			}
		}
	}

	s3Deployment, err := checkS3Deployment(cl, ns)
	if err != nil {
		return err
	}

	if s3Deployment {
		for _, container := range existing.Spec.Template.Spec.Containers {
			for _, envVarName := range AWSVARS {
				envVar, ok := FindEnvByName(container.Env, envVarName)
				if !ok {
					return fmt.Errorf("%s env var not found in container %s", envVarName, container.Name)
				}

				if envVar.ValueFrom == nil || envVar.ValueFrom.SecretKeyRef == nil {
					return fmt.Errorf("%s env var from container %s is not reference to secret", envVarName, container.Name)
				}
			}
		}
	}

	return nil

}

func verifySystemAppPreHookPodEnv(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	for _, envVarName := range append(SMTPVARS[:], component.SystemSecretSystemSeedMasterAccessTokenFieldName) {
		envVar, ok := FindEnvByName(existing.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env, envVarName)
		if !ok {
			return fmt.Errorf("%s env var not found", envVarName)
		}

		if envVar.ValueFrom == nil || envVar.ValueFrom.SecretKeyRef == nil {
			return fmt.Errorf("%s env var is not reference to secret", envVarName)
		}
	}

	s3Deployment, err := checkS3Deployment(cl, ns)
	if err != nil {
		return err
	}

	if s3Deployment {
		for _, envVarName := range AWSVARS {
			envVar, ok := FindEnvByName(existing.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env, envVarName)
			if !ok {
				return fmt.Errorf("%s env var not found", envVarName)
			}

			if envVar.ValueFrom == nil || envVar.ValueFrom.SecretKeyRef == nil {
				return fmt.Errorf("%s env var is not reference to secret", envVarName)
			}
		}
	}

	return nil

}

func verifySystemAppPreHookPodCommand(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	system, err := GetSystemComponent()
	desired := system.AppDeploymentConfig()

	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPrehookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod
	if !reflect.DeepEqual(existingPrehookPod.Command, desiredPreHookPod.Command) {
		return errors.New("system app pre hook pod command not upgraded")
	}
	return nil
}

func verifySMTPSecret(cl client.Client, ns string) error {
	secret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{Name: component.SystemSecretSystemSMTPSecretName, Namespace: ns}
	err := cl.Get(context.TODO(), secretNamespacedName, secret)
	if err != nil {
		return err
	}

	if len(secret.Data) < 1 {
		return fmt.Errorf("secret %s is empty", component.SystemSecretSystemSMTPSecretName)
	}

	return nil
}

func verifySMTPConfigMapDeleted(cl client.Client, ns string) error {
	configMap := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "smtp", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, configMap)

	if err == nil {
		return errors.New("smtp configmap still exists")
	}

	if k8serrors.IsNotFound(err) {
		return nil
	}

	return err
}
