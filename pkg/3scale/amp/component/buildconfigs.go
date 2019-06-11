package component

import (
	"fmt"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type BuildConfigs struct {
	options []string
	Options *BuildConfigsOptions
}

type BuildConfigsOptions struct {
	appLabel string
	gitRef   string
}

type BuildConfigsOptionsProvider interface {
	GetBuildConfigsOptions() *BuildConfigsOptions
}
type CLIBuildConfigsOptionsProvider struct {
}

func (o *CLIBuildConfigsOptionsProvider) GetBuildConfigsOptions() (*BuildConfigsOptions, error) {
	sob := BuildConfigsOptionsBuilder{}
	sob.AppLabel("${APP_LABEL}")
	sob.GitRef("${GIT_REF}")
	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create BuildConfigs Options - %s", err)
	}
	return res, nil
}

func NewBuildConfigs(options []string) *BuildConfigs {
	bcs := &BuildConfigs{
		options: options,
	}
	return bcs
}

func (bcs *BuildConfigs) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIBuildConfigsOptionsProvider{}
	buildConfigsOpts, err := optionsProvider.GetBuildConfigsOptions()
	_ = err
	bcs.Options = buildConfigsOpts
	bcs.buildParameters(template)
	bcs.addObjectsIntoTemplate(template)
}

func (bcs *BuildConfigs) GetObjects() ([]runtime.RawExtension, error) {
	objects := bcs.buildObjects()
	return objects, nil
}

func (bcs *BuildConfigs) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := bcs.buildObjects()
	template.Objects = append(template.Objects, objects...)
}

func (bcs *BuildConfigs) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (bcs *BuildConfigs) buildObjects() []runtime.RawExtension {
	backendBuildConfig := bcs.buildBackendBuildConfig()
	zyncBuildConfig := bcs.buildZyncBuildConfig()
	apicastBuildConfig := bcs.buildApicastBuildConfig()
	systemBuildConfig := bcs.buildSystemBuildConfig()

	buildRubyCentosSevenImageStream := bcs.buildRubyCentosSevenImageStream()
	buildOpenrestyCentosSevenImageStream := bcs.buildOpenrestyCentosSevenImageStream()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: backendBuildConfig},
		runtime.RawExtension{Object: zyncBuildConfig},
		runtime.RawExtension{Object: apicastBuildConfig},
		runtime.RawExtension{Object: systemBuildConfig},
		runtime.RawExtension{Object: buildRubyCentosSevenImageStream},
		runtime.RawExtension{Object: buildOpenrestyCentosSevenImageStream},
	}
	return objects
}

func (bcs *BuildConfigs) buildBackendBuildConfig() *buildv1.BuildConfig {
	return &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend",
			Labels: map[string]string{
				"app":                  bcs.Options.appLabel,
				"threescale_component": "backend",
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: "amp-backend:latest",
					},
				},
				Source: buildv1.BuildSource{
					Git: &buildv1.GitBuildSource{
						URI: "https://github.com/3scale/backend.git",
						Ref: bcs.Options.gitRef,
					},
					SourceSecret: &v1.LocalObjectReference{
						Name: "github-auth",
					},
					Type: buildv1.BuildSourceGit,
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.DockerBuildStrategyType,
					DockerStrategy: &buildv1.DockerBuildStrategy{
						DockerfilePath: "openshift/distro/centos/7/release/Dockerfile",
						PullSecret: &v1.LocalObjectReference{
							Name: "quay-auth",
						},
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				buildv1.BuildTriggerPolicy{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildZyncBuildConfig() *buildv1.BuildConfig {
	return &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync",
			Labels: map[string]string{
				"app":                  bcs.Options.appLabel,
				"threescale_component": "zync",
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: "amp-zync:latest",
					},
				},
				Source: buildv1.BuildSource{
					Git: &buildv1.GitBuildSource{
						URI: "https://github.com/3scale/zync.git",
						Ref: bcs.Options.gitRef,
					},
					Type: buildv1.BuildSourceGit,
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.SourceBuildStrategyType,
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "ruby-24-centos7:latest",
						},
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				buildv1.BuildTriggerPolicy{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
				buildv1.BuildTriggerPolicy{
					Type: buildv1.ConfigChangeBuildTriggerType,
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildApicastBuildConfig() *buildv1.BuildConfig {
	return &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast",
			Labels: map[string]string{
				"app":                  bcs.Options.appLabel,
				"threescale_component": "apicast",
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: "amp-apicast:latest",
					},
				},
				Source: buildv1.BuildSource{
					ContextDir: "gateway",
					Git: &buildv1.GitBuildSource{
						URI: "https://github.com/3scale/apicast.git",
						Ref: bcs.Options.gitRef,
					},
					Type: buildv1.BuildSourceGit,
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.SourceBuildStrategyType,
					SourceStrategy: &buildv1.SourceBuildStrategy{
						// TODO it seems RuntimeImage field has been deprecated and says: Deprecated: This feature will be removed in a future release. Use ImageSource to copy binary artifacts created from one build into a separate runtime image
						ForcePull: true,
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "s2i-openresty-centos7:builder",
						},
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				buildv1.BuildTriggerPolicy{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
				buildv1.BuildTriggerPolicy{
					Type: buildv1.ConfigChangeBuildTriggerType,
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildSystemBuildConfig() *buildv1.BuildConfig {
	return &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system",
			Labels: map[string]string{
				"app":                  bcs.Options.appLabel,
				"threescale_component": "system",
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: "amp-system:latest",
					},
				},
				Source: buildv1.BuildSource{
					Git: &buildv1.GitBuildSource{
						URI: "https://github.com/3scale/porta.git",
						Ref: bcs.Options.gitRef,
					},
					Type: buildv1.BuildSourceGit,
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.DockerBuildStrategyType,
					DockerStrategy: &buildv1.DockerBuildStrategy{
						ForcePull:      true,
						DockerfilePath: "openshift/system/Dockerfile.on_prem",
						PullSecret: &v1.LocalObjectReference{
							Name: "quay-auth",
						},
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				buildv1.BuildTriggerPolicy{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildRubyCentosSevenImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "image.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "ruby-24-centos7",
			Labels: map[string]string{"threescale_component": "zync", "app": bcs.Options.appLabel},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: "centos/ruby-24-centos7",
					},
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildOpenrestyCentosSevenImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "image.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "s2i-openresty-centos7",
			Labels: map[string]string{"threescale_component": "apicast", "app": bcs.Options.appLabel},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "builder",
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: "quay.io/3scale/s2i-openresty-centos7:master",
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Scheduled: true,
					},
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "GIT_REF",
			Required:    true,
			Value:       "master",
			Description: "Git Reference to use. Can be a tag or branch.",
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}
