package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsRegion          string
	awsBucket          string
	fileUploadStorage  string
}

type s3NonRequiredOptions struct {
}

func NewS3(options []string) *S3 {
	s3 := &S3{
		options: options,
	}
	return s3
}

type S3OptionsBuilder struct {
	options S3Options
}

func (s3 *S3OptionsBuilder) AwsAccessKeyId(awsAccessKeyId string) {
	s3.options.awsAccessKeyId = awsAccessKeyId
}

func (s3 *S3OptionsBuilder) AwsSecretAccessKey(awsSecretAccessKey string) {
	s3.options.awsSecretAccessKey = awsSecretAccessKey
}

func (s3 *S3OptionsBuilder) AwsRegion(awsRegion string) {
	s3.options.awsRegion = awsRegion
}

func (s3 *S3OptionsBuilder) AwsBucket(awsBucket string) {
	s3.options.awsBucket = awsBucket
}

func (s3 *S3OptionsBuilder) FileUploadStorage(fileUploadStorage string) {
	s3.options.fileUploadStorage = fileUploadStorage
}

func (s3 *S3OptionsBuilder) Build() (*S3Options, error) {
	err := s3.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	s3.setNonRequiredOptions()

	return &s3.options, nil

}

func (s3 *S3OptionsBuilder) setRequiredOptions() error {
	if s3.options.awsAccessKeyId == "" {
		return fmt.Errorf("no AWS access key id has been provided")
	}
	if s3.options.awsSecretAccessKey == "" {
		return fmt.Errorf("no AWS secret access key has been provided")
	}
	if s3.options.awsRegion == "" {
		return fmt.Errorf("no AWS region has been provided")
	}
	if s3.options.awsBucket == "" {
		return fmt.Errorf("no AWS bucket has been provided")
	}
	if s3.options.fileUploadStorage == "" {
		return fmt.Errorf("no file upload storage has been provided")
	}

	return nil
}

func (s3 *S3OptionsBuilder) setNonRequiredOptions() {

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

func (s3 *S3) GetObjects() ([]runtime.RawExtension, error) {
	objects := s3.buildObjects()
	return objects, nil
}

func (s3 *S3) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := s3.buildObjects()
	template.Objects = append(template.Objects, objects...)
}

func (s3 *S3) removeSystemStorageReferences(objects []runtime.RawExtension) {
	for _, rawExtension := range objects {
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

func (s3 *S3) addCfgMapElemsToSystemBaseEnv(objects []runtime.RawExtension) {
	newCfgMapElements := s3.getNewCfgMapElements()
	for _, rawExtension := range objects {
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

func (s3 *S3) addS3PostprocessOptionsToSystemEnvironmentCfgMap(objects []runtime.RawExtension) {
	var systemEnvCfgMap *v1.ConfigMap

	for rawExtIdx := range objects {
		obj := objects[rawExtIdx].Object
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

func (s3 *S3) removeSystemStoragePVC(objects []runtime.RawExtension) []runtime.RawExtension {
	res := objects

	for idx, rawExtension := range res {
		obj := rawExtension.Object
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

	res := template.Objects
	res = s3.removeSystemStoragePVC(res)
	s3.removeSystemStorageReferences(res)
	s3.addS3PostprocessOptionsToSystemEnvironmentCfgMap(template.Objects)
	s3.addCfgMapElemsToSystemBaseEnv(template.Objects)
	s3.removeRWXStorageClassParameter(template)
	template.Objects = res
}

func (s3 *S3) PostProcessObjects(objects []runtime.RawExtension) []runtime.RawExtension {
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

func (s3 *S3) buildObjects() []runtime.RawExtension {
	s3AWSSecret := s3.buildS3AWSSecret()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: s3AWSSecret},
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
			Name: "aws-auth",
		},
		StringData: map[string]string{
			"AWS_ACCESS_KEY_ID":     s3.Options.awsAccessKeyId,
			"AWS_SECRET_ACCESS_KEY": s3.Options.awsSecretAccessKey,
		},
		Type: v1.SecretTypeOpaque,
	}

}
