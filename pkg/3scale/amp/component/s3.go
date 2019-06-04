package component

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis/common"
	"github.com/3scale/3scale-operator/pkg/helper"

	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	S3SecretAWSAccessKeyIdFieldName     = "AWS_ACCESS_KEY_ID"
	S3SecretAWSSecretAccessKeyFieldName = "AWS_SECRET_ACCESS_KEY"
)

type S3 struct {
	options []string
	Options *S3Options
}

type S3Options struct {
	s3NonRequiredOptions
	s3RequiredOptions
}

type s3RequiredOptions struct {
	awsAccessKeyId       string
	awsSecretAccessKey   string
	awsRegion            string
	awsBucket            string
	fileUploadStorage    string
	awsCredentialsSecret string
}

type s3NonRequiredOptions struct {
}

func NewS3(options []string) *S3 {
	s3 := &S3{
		options: options,
	}
	return s3
}

type S3OptionsProvider interface {
	GetS3Options() *S3Options
}
type CLIS3OptionsProvider struct {
}

func (o *CLIS3OptionsProvider) GetS3Options() (*S3Options, error) {
	sob := S3OptionsBuilder{}
	sob.AwsAccessKeyId("${AWS_ACCESS_KEY_ID}")
	sob.AwsSecretAccessKey("${AWS_SECRET_ACCESS_KEY}")
	sob.AwsRegion("${AWS_REGION}")
	sob.AwsBucket("${AWS_BUCKET}")
	sob.FileUploadStorage("${FILE_UPLOAD_STORAGE}")
	sob.AWSCredentialsSecret("aws-auth")
	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create S3 Options - %s", err)
	}
	return res, nil
}

func (s3 *S3) setS3Options() {
	// TODO move this outside this specific method
	optionsProvider := CLIS3OptionsProvider{}
	s3Opts, err := optionsProvider.GetS3Options()
	_ = err
	s3.Options = s3Opts
}

func (s3 *S3) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	s3.setS3Options() // TODO move this outside
	s3.buildParameters(template)
	s3.addObjectsIntoTemplate(template)
}

func (s3 *S3) GetObjects() ([]common.KubernetesObject, error) {
	objects := s3.buildObjects()
	return objects, nil
}

func (s3 *S3) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := s3.buildObjects()
	template.Objects = append(template.Objects, helper.WrapRawExtensions(objects)...)
}

func (s3 *S3) removeSystemStorageReferences(objects []common.KubernetesObject) {
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
		envVarFromConfigMap("FILE_UPLOAD_STORAGE", "system-environment", "FILE_UPLOAD_STORAGE"),
		envVarFromSecret("AWS_ACCESS_KEY_ID", s3.Options.awsCredentialsSecret, S3SecretAWSAccessKeyIdFieldName),
		envVarFromSecret("AWS_SECRET_ACCESS_KEY", s3.Options.awsCredentialsSecret, S3SecretAWSSecretAccessKeyFieldName),
		envVarFromConfigMap("AWS_BUCKET", "system-environment", "AWS_BUCKET"),
		envVarFromConfigMap("AWS_REGION", "system-environment", "AWS_REGION"),
	}
}

func (s3 *S3) addCfgMapElemsToSystemBaseEnv(objects []common.KubernetesObject) {
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

func (s3 *S3) addS3PostprocessOptionsToSystemEnvironmentCfgMap(objects []common.KubernetesObject) {
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

	systemEnvCfgMap.Data["FILE_UPLOAD_STORAGE"] = s3.Options.fileUploadStorage
	systemEnvCfgMap.Data["AWS_BUCKET"] = s3.Options.awsBucket
	systemEnvCfgMap.Data["AWS_REGION"] = s3.Options.awsRegion
}

func (s3 *S3) removeSystemStoragePVC(objects []common.KubernetesObject) []common.KubernetesObject {
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

func (s3 *S3) PostProcess(template *templatev1.Template, otherComponents []Component) {
	s3.setS3Options() // TODO move this outside

	res := helper.UnwrapRawExtensions(template.Objects)
	res = s3.removeSystemStoragePVC(res)
	s3.removeSystemStorageReferences(res)
	s3.addS3PostprocessOptionsToSystemEnvironmentCfgMap(helper.UnwrapRawExtensions(template.Objects))
	s3.addCfgMapElemsToSystemBaseEnv(helper.UnwrapRawExtensions(template.Objects))
	s3.removeRWXStorageClassParameter(template)
	template.Objects = helper.WrapRawExtensions(res)
}

func (s3 *S3) PostProcessObjects(objects []common.KubernetesObject) []common.KubernetesObject {
	res := objects
	res = s3.removeSystemStoragePVC(res)
	s3.removeSystemStorageReferences(res)
	s3.addS3PostprocessOptionsToSystemEnvironmentCfgMap(res)
	s3.addCfgMapElemsToSystemBaseEnv(res)

	return res
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

func (s3 *S3) buildObjects() []common.KubernetesObject {
	s3AWSSecret := s3.buildS3AWSSecret()

	objects := []common.KubernetesObject{
		s3AWSSecret,
	}
	return objects
}

func (s3 *S3) buildS3AWSSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: s3.Options.awsCredentialsSecret,
		},
		StringData: map[string]string{
			S3SecretAWSAccessKeyIdFieldName:     s3.Options.awsAccessKeyId,
			S3SecretAWSSecretAccessKeyFieldName: s3.Options.awsSecretAccessKey,
		},
		Type: v1.SecretTypeOpaque,
	}
}
