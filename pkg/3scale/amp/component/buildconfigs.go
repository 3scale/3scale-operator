package component

import (
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type BuildConfigs struct {
	options []string
}

func NewBuildConfigs(options []string) *BuildConfigs {
	bcs := &BuildConfigs{
		options: options,
	}
	return bcs
}

func (bcs *BuildConfigs) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	bcs.buildParameters(template)
	bcs.buildObjects(template)
}

func (bcs *BuildConfigs) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (bcs *BuildConfigs) buildObjects(template *templatev1.Template) {
	backendBuildConfig := bcs.buildBackendBuildConfig()
	zyncBuildConfig := bcs.buildZyncBuildConfig()
	apicastBuildConfig := bcs.buildApicastBuildConfig()
	wildcardRouterBuildConfig := bcs.buildWildcardRouterBuildConfig()
	systemBuildConfig := bcs.buildSystemBuildConfig()

	buildRubyCentosSevenImageStream := bcs.buildRubyCentosSevenImageStream()
	buildOpenrestyCentosSevenImageStream := bcs.buildOpenrestyCentosSevenImageStream()

	objects := []runtime.RawExtension{
		{Object: backendBuildConfig},
		{Object: zyncBuildConfig},
		{Object: apicastBuildConfig},
		{Object: wildcardRouterBuildConfig},
		{Object: systemBuildConfig},
		{Object: buildRubyCentosSevenImageStream},
		{Object: buildOpenrestyCentosSevenImageStream},
	}
	template.Objects = append(template.Objects, objects...)
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
				"app":              "${APP_LABEL}",
				"3scale.component": "backend",
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
						Ref: "${GIT_REF}",
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
				{
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
				"app":              "${APP_LABEL}",
				"3scale.component": "zync",
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
						Ref: "${GIT_REF}",
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
				{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
				{
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
				"app":              "${APP_LABEL}",
				"3scale.component": "apicast",
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
						Ref: "${GIT_REF}",
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
				{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
				{
					Type: buildv1.ConfigChangeBuildTriggerType,
				},
			},
		},
	}
}

func (bcs *BuildConfigs) buildWildcardRouterBuildConfig() *buildv1.BuildConfig {
	return &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "wildcard-router",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "wildcard-router",
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: "amp-wildcard-router:latest",
					},
				},
				Source: buildv1.BuildSource{
					Git: &buildv1.GitBuildSource{
						URI: "https://github.com/3scale/wildcard-router-service.git",
						Ref: "${GIT_REF}",
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
				{
					Type: buildv1.ImageChangeBuildTriggerType,
				},
				{
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
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
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
						Ref: "${GIT_REF}",
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
				{
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
			Labels: map[string]string{"3scale.component": "zync", "app": "${APP_LABEL}"},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
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
			Labels: map[string]string{"3scale.component": "apicast", "app": "${APP_LABEL}"},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
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
		{
			Name:        "GIT_REF",
			Required:    true,
			Value:       "master",
			Description: "Git Reference to use. Can be a tag or branch.",
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}
