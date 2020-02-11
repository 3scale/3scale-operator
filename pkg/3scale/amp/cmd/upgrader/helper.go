package upgrader

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	THREESCALEIMAGES = [...]string{
		"amp-system",
		"amp-apicast",
		"amp-backend",
		"amp-zync",
		"system-memcached",
		"zync-database-postgresql",
		"backend-redis",
		"system-redis",
	}

	SMTPVARS = [...]string{
		"SMTP_ADDRESS", "SMTP_USER_NAME", "SMTP_PASSWORD",
		"SMTP_DOMAIN", "SMTP_PORT", "SMTP_AUTHENTICATION", "SMTP_OPENSSL_VERIFY_MODE",
	}

	AWSVARS = [...]string{
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_BUCKET", "AWS_REGION",
		"AWS_PROTOCOL", "AWS_HOSTNAME", "AWS_PATH_STYLE",
	}
)

func FindEnvByNameOrPanic(a []v1.EnvVar, x string) v1.EnvVar {
	for _, n := range a {
		if x == n.Name {
			return n
		}
	}

	panic(fmt.Sprintf("env var %s not found", x))
}

func FindEnvByName(a []v1.EnvVar, x string) (v1.EnvVar, bool) {
	for _, n := range a {
		if x == n.Name {
			return n, true
		}
	}
	return v1.EnvVar{}, false
}

func FindContainerEnvByNameOrPanic(a []v1.Container, containerName, envVarName string) v1.EnvVar {
	container := FindContainerByNameOrPanic(a, containerName)
	for _, n := range container.Env {
		if envVarName == n.Name {
			return n
		}
	}

	panic(fmt.Sprintf("env var %s not found", envVarName))
}

func FindContainerByNameOrPanic(a []v1.Container, containerName string) v1.Container {
	for _, n := range a {
		if containerName == n.Name {
			return n
		}
	}

	panic(fmt.Sprintf("container %s not found", containerName))
}

func RegisterOpenShiftAPIGroups(s *runtime.Scheme) error {
	var addToSchemes runtime.SchemeBuilder
	addToSchemes = append(addToSchemes,
		appsv1.Install,
		imagev1.Install,
		routev1.Install,
	)

	return addToSchemes.AddToScheme(s)
}

func GetSystemComponent() (*component.System, error) {
	optProv := component.SystemOptionsBuilder{}

	// Maybe implement option provider to gather all information from existing installation???
	// Not required for the upgrade 2.7 -> 2.8, but could be done.
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
	systemOptions, err := optProv.Build()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(systemOptions), nil
}
