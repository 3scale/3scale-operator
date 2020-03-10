package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	AwsBucket          = "AWS_BUCKET"
	AwsRegion          = "AWS_REGION"
	AwsProtocol        = "AWS_PROTOCOL"
	AwsHostname        = "AWS_HOSTNAME"
	AwsPathStyle       = "AWS_PATH_STYLE"
)

type S3 struct {
	Options *S3Options
}

func NewS3(options *S3Options) *S3 {
	return &S3{Options: options}
}

func (s3 *S3) Objects() []common.KubernetesObject {
	s3AWSSecret := s3.S3AWSSecret()

	objects := []common.KubernetesObject{
		s3AWSSecret,
	}
	return objects
}

func (s3 *S3) RemoveSystemStorageReferences(objects []common.KubernetesObject) {
	for _, obj := range objects {
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			if dc.ObjectMeta.Name == "system-app" || dc.ObjectMeta.Name == "system-sidekiq" {

				// Remove system-storage references in the VolumeMount fields of the containers
				for containerIdx := range dc.Spec.Template.Spec.Containers {
					container := &dc.Spec.Template.Spec.Containers[containerIdx]
					resIdx := -1
					for vmIdx, vm := range container.VolumeMounts {
						if vm.Name == "system-storage" {
							resIdx = vmIdx
							break
						}
					}
					if resIdx != -1 {
						container.VolumeMounts = append(container.VolumeMounts[:resIdx], container.VolumeMounts[resIdx+1:]...)
					}
				}

				// Remove system-storage references in the Volumes fields of the containers
				resIdx := -1
				for volIdx := range dc.Spec.Template.Spec.Volumes {
					vol := &dc.Spec.Template.Spec.Volumes[volIdx]
					if vol.Name == "system-storage" {
						resIdx = volIdx
						break
					}
				}
				if resIdx != -1 {
					dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes[:resIdx], dc.Spec.Template.Spec.Volumes[resIdx+1:]...)
				}

				// Remove system-storage references in the Volumes fields of the pre-hook in system-app
				if dc.ObjectMeta.Name == "system-app" {
					resIdx = -1
					for volIdx := range dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Volumes {
						vol := &dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Volumes[volIdx]
						if *vol == "system-storage" {
							resIdx = volIdx
							break
						}
					}
					if resIdx != -1 {
						dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Volumes = append(dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Volumes[:resIdx], dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Volumes[resIdx+1:]...)
					}
				}
			}
		}
	}
}

// Remove the RWX_STORAGE_CLASS parameter because it is used only for the system-storage PersistentVolumeClaim
func (s3 *S3) RemoveRWXStorageClassParameter(template *templatev1.Template) {
	for paramIdx, param := range template.Parameters {
		if param.Name == "RWX_STORAGE_CLASS" {
			template.Parameters = append(template.Parameters[:paramIdx], template.Parameters[paramIdx+1:]...)
			break
		}
	}
}

func (s3 *S3) getNewCfgMapElements() []v1.EnvVar {
	return []v1.EnvVar{
		envVarFromConfigMap("FILE_UPLOAD_STORAGE", "system-environment", "FILE_UPLOAD_STORAGE"),
		envVarFromSecret(AwsAccessKeyID, s3.Options.AwsCredentialsSecret, AwsAccessKeyID),
		envVarFromSecret(AwsSecretAccessKey, s3.Options.AwsCredentialsSecret, AwsSecretAccessKey),
		envVarFromSecret(AwsBucket, s3.Options.AwsCredentialsSecret, AwsBucket),
		envVarFromSecret(AwsRegion, s3.Options.AwsCredentialsSecret, AwsRegion),
		envVarFromSecretOptional(AwsProtocol, s3.Options.AwsCredentialsSecret, AwsProtocol),
		envVarFromSecretOptional(AwsHostname, s3.Options.AwsCredentialsSecret, AwsHostname),
		envVarFromSecretOptional(AwsPathStyle, s3.Options.AwsCredentialsSecret, AwsPathStyle),
	}
}

func (s3 *S3) AddCfgMapElemsToSystemBaseEnv(objects []common.KubernetesObject) {
	newCfgMapElements := s3.getNewCfgMapElements()
	for _, obj := range objects {
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			if dc.ObjectMeta.Name == "system-app" || dc.ObjectMeta.Name == "system-sidekiq" {
				if dc.ObjectMeta.Name == "system-app" {
					dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env = append(dc.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env, newCfgMapElements...)
				}

				for containerIdx := range dc.Spec.Template.Spec.Containers {
					container := &dc.Spec.Template.Spec.Containers[containerIdx]
					container.Env = append(container.Env, newCfgMapElements...)
				}
			}
		}
	}
}

func (s3 *S3) AddS3PostprocessOptionsToSystemEnvironmentCfgMap(objects []common.KubernetesObject) {
	var systemEnvCfgMap *v1.ConfigMap

	for objIdx := range objects {
		obj := objects[objIdx]
		cfgmap, ok := obj.(*v1.ConfigMap)
		if ok {
			if cfgmap.Name == "system-environment" {
				systemEnvCfgMap = cfgmap
				break
			}
		}
	}

	systemEnvCfgMap.Data["FILE_UPLOAD_STORAGE"] = "s3"
}

func (s3 *S3) RemoveSystemStoragePVC(objects []common.KubernetesObject) []common.KubernetesObject {
	res := objects

	for idx, obj := range res {
		pvc, ok := obj.(*v1.PersistentVolumeClaim)
		if ok {
			if pvc.ObjectMeta.Name == "system-storage" {
				res = append(res[:idx], res[idx+1:]...) // This deletes the element in the array
				break
			}
		}
	}

	return res
}

func (s3 *S3) S3AWSSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: s3.Options.AwsCredentialsSecret,
		},
		StringData: map[string]string{
			AwsAccessKeyID:     s3.Options.AwsAccessKeyId,
			AwsSecretAccessKey: s3.Options.AwsSecretAccessKey,
			AwsRegion:          s3.Options.AwsRegion,
			AwsBucket:          s3.Options.AwsBucket,
			AwsProtocol:        s3.Options.AwsProtocol,
			AwsHostname:        s3.Options.AwsHostname,
			AwsPathStyle:       s3.Options.AwsPathStyle,
		},
		Type: v1.SecretTypeOpaque,
	}
}
