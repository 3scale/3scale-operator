package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type S3 struct {
	options []string
}

func NewS3(options []string) *S3 {
	s3 := &S3{
		options: options,
	}
	return s3
}

func (s3 *S3) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	s3.buildParameters(template)
	s3.buildObjects(template)
}

func (s3 *S3) removeSystemStorageReferences(template *templatev1.Template) {
	for _, rawExtension := range template.Objects {
		obj := rawExtension.Object
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
func (s3 *S3) removeRWXStorageClassParameter(template *templatev1.Template) {
	for paramIdx, param := range template.Parameters {
		if param.Name == "RWX_STORAGE_CLASS" {
			template.Parameters = append(template.Parameters[:paramIdx], template.Parameters[paramIdx+1:]...)
			break
		}
	}
}

func (s3 *S3) getNewCfgMapElements() []v1.EnvVar {
	return []v1.EnvVar{
		createEnvVarFromConfigMap("FILE_UPLOAD_STORAGE", "system-environment", "FILE_UPLOAD_STORAGE"),
		createEnvvarFromSecret("AWS_ACCESS_KEY_ID", "aws-auth", "AWS_ACCESS_KEY_ID"),
		createEnvvarFromSecret("AWS_SECRET_ACCESS_KEY", "aws-auth", "AWS_SECRET_ACCESS_KEY"),
		createEnvVarFromConfigMap("AWS_BUCKET", "system-environment", "AWS_BUCKET"),
		createEnvVarFromConfigMap("AWS_REGION", "system-environment", "AWS_REGION"),
	}
}

func (s3 *S3) addNewParametersToSystemBaseEnv(template *templatev1.Template) {
	newCfgMapElements := s3.getNewCfgMapElements()
	for _, rawExtension := range template.Objects {
		obj := rawExtension.Object
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

func (s3 *S3) addNewParametersToSystemEnvironmentCfgMap(template *templatev1.Template) {
	var systemEnvCfgMap *v1.ConfigMap

	for rawExtIdx := range template.Objects {
		obj := template.Objects[rawExtIdx].Object
		cfgmap, ok := obj.(*v1.ConfigMap)
		if ok {
			if cfgmap.Name == "system-environment" {
				systemEnvCfgMap = cfgmap
				break
			}
		}
	}

	for _, param := range template.Parameters {
		switch param.Name {
		case "FILE_UPLOAD_STORAGE":
			systemEnvCfgMap.Data["FILE_UPLOAD_STORAGE"] = "${FILE_UPLOAD_STORAGE}"
		case "AWS_BUCKET":
			systemEnvCfgMap.Data["AWS_BUCKET"] = "${AWS_BUCKET}"
		case "AWS_REGION":
			systemEnvCfgMap.Data["AWS_REGION"] = "${AWS_REGION}"
		}
	}

}

func (s3 *S3) removeSystemStoragePVC(template *templatev1.Template) {
	for idx, rawExtension := range template.Objects {
		obj := rawExtension.Object
		pvc, ok := obj.(*v1.PersistentVolumeClaim)
		if ok {
			if pvc.ObjectMeta.Name == "system-storage" {
				template.Objects = append(template.Objects[:idx], template.Objects[idx+1:]...) // This deletes the element in the array
				break
			}
		}
	}
}

func (s3 *S3) PostProcess(template *templatev1.Template, otherComponents []Component) {
	s3.removeSystemStoragePVC(template)
	s3.removeSystemStorageReferences(template)
	s3.removeRWXStorageClassParameter(template)
	s3.addNewParametersToSystemEnvironmentCfgMap(template)
	s3.addNewParametersToSystemBaseEnv(template)
}

func (s3 *S3) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "FILE_UPLOAD_STORAGE",
			Description: "Define Assets Storage",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_ACCESS_KEY_ID",
			Description: "AWS Access Key ID to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_SECRET_ACCESS_KEY",
			Description: "AWS Access Key Secret to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_BUCKET",
			Description: "AWS S3 Bucket Name to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_REGION",
			Description: "AWS Region to use in S3 Storage for assets.",
			Required:    false,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

func (s3 *S3) buildObjects(template *templatev1.Template) {
	s3AWSSecret := s3.buildS3AWSSecret()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: s3AWSSecret},
	}
	template.Objects = append(template.Objects, objects...)
}

func (s3 *S3) buildS3AWSSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "aws-auth",
		},
		StringData: map[string]string{
			"AWS_ACCESS_KEY_ID":     "${AWS_ACCESS_KEY_ID}",
			"AWS_SECRET_ACCESS_KEY": "${AWS_SECRET_ACCESS_KEY}",
		},
		Type: v1.SecretTypeOpaque,
	}

}
