package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"

	appsv1 "github.com/openshift/api/apps/v1"
)

type S3 struct {
}

func NewS3Adapter() Adapter {
	return &S3{}
}

func (s *S3) Adapt(template *templatev1.Template) {
	s3Options, err := s.options()
	if err != nil {
		panic(err)
	}
	s3Component := component.NewS3(s3Options)

	s.addParameters(template)
	s.addObjects(template, s3Component)
	s.postProcess(template, s3Component)

	// update metadata
	template.Name = "3scale-api-management-s3"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system with shared file storage in AWS S3."
}

func (s *S3) postProcess(template *templatev1.Template, s3Component *component.S3) {
	res := helper.UnwrapRawExtensions(template.Objects)
	res = s.removeSystemStoragePVC(res)
	s.removeSystemStorageReferences(res)
	s.addS3PostprocessOptionsToSystemEnvironmentCfgMap(helper.UnwrapRawExtensions(template.Objects))
	s.addCfgMapElemsToSystemBaseEnv(s3Component, helper.UnwrapRawExtensions(template.Objects))
	s.removeRWXStorageClassParameter(template)
	template.Objects = helper.WrapRawExtensions(res)
}

func (s *S3) removeSystemStoragePVC(objects []common.KubernetesObject) []common.KubernetesObject {
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

func (s *S3) removeSystemStorageReferences(objects []common.KubernetesObject) {
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

func (s *S3) addS3PostprocessOptionsToSystemEnvironmentCfgMap(objects []common.KubernetesObject) {
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

func (s *S3) addCfgMapElemsToSystemBaseEnv(c *component.S3, objects []common.KubernetesObject) {
	newCfgMapElements := s.getNewCfgMapElements(c)
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

func (s *S3) getNewCfgMapElements(c *component.S3) []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromConfigMap("FILE_UPLOAD_STORAGE", "system-environment", "FILE_UPLOAD_STORAGE"),
		helper.EnvVarFromSecret(component.AwsAccessKeyID, c.Options.AwsCredentialsSecret, component.AwsAccessKeyID),
		helper.EnvVarFromSecret(component.AwsSecretAccessKey, c.Options.AwsCredentialsSecret, component.AwsSecretAccessKey),
		helper.EnvVarFromSecret(component.AwsBucket, c.Options.AwsCredentialsSecret, component.AwsBucket),
		helper.EnvVarFromSecret(component.AwsRegion, c.Options.AwsCredentialsSecret, component.AwsRegion),
		helper.EnvVarFromSecretOptional(component.AwsProtocol, c.Options.AwsCredentialsSecret, component.AwsProtocol),
		helper.EnvVarFromSecretOptional(component.AwsHostname, c.Options.AwsCredentialsSecret, component.AwsHostname),
		helper.EnvVarFromSecretOptional(component.AwsPathStyle, c.Options.AwsCredentialsSecret, component.AwsPathStyle),
	}
}

// Remove the RWX_STORAGE_CLASS parameter because it is used only for the system-storage PersistentVolumeClaim
func (s *S3) removeRWXStorageClassParameter(template *templatev1.Template) {
	for paramIdx, param := range template.Parameters {
		if param.Name == "RWX_STORAGE_CLASS" {
			template.Parameters = append(template.Parameters[:paramIdx], template.Parameters[paramIdx+1:]...)
			break
		}
	}
}

func (s *S3) addObjects(template *templatev1.Template, s3 *component.S3) {
	componentObjects := s.componentObjects(s3)
	template.Objects = append(template.Objects, helper.WrapRawExtensions(componentObjects)...)
}

func (s *S3) componentObjects(c *component.S3) []common.KubernetesObject {
	s3AWSSecret := c.S3AWSSecret()

	objects := []common.KubernetesObject{
		s3AWSSecret,
	}
	return objects
}

func (s *S3) addParameters(template *templatev1.Template) {
	template.Parameters = append(template.Parameters, s.parameters()...)
}

func (s *S3) options() (*component.S3Options, error) {
	o := component.NewS3Options()
	o.AwsAccessKeyId = "${AWS_ACCESS_KEY_ID}"
	o.AwsSecretAccessKey = "${AWS_SECRET_ACCESS_KEY}"
	o.AwsRegion = "${AWS_REGION}"
	o.AwsBucket = "${AWS_BUCKET}"
	o.AwsProtocol = "${AWS_PROTOCOL}"
	o.AwsHostname = "${AWS_HOSTNAME}"
	o.AwsPathStyle = "${AWS_PATH_STYLE}"
	o.AwsCredentialsSecret = "aws-auth"

	err := o.Validate()
	return o, err
}

func (s3 *S3) parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
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
		templatev1.Parameter{
			Name:        "AWS_PROTOCOL",
			Description: "AWS S3 compatible provider endpoint protocol. HTTP or HTTPS.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_HOSTNAME",
			Description: "AWS S3 compatible provider endpoint hostname",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_PATH_STYLE",
			Description: "When set to true, the bucket name is always left in the request URI and never moved to the host as a sub-domain",
			Required:    false,
		},
	}
}
