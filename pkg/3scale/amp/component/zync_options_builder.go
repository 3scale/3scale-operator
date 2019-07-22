package component

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ZyncOptions struct {
	// zyncNonRequiredOptions
	databaseURL                           *string
	containerResourceRequirements         *v1.ResourceRequirements
	queContainerResourceRequirements      *v1.ResourceRequirements
	databaseContainerResourceRequirements *v1.ResourceRequirements

	// zyncRequiredOptions
	appLabel            string
	authenticationToken string
	databasePassword    string
	secretKeyBase       string
}

type ZyncOptionsBuilder struct {
	options ZyncOptions
}

func (z *ZyncOptionsBuilder) AppLabel(appLabel string) {
	z.options.appLabel = appLabel
}

func (z *ZyncOptionsBuilder) AuthenticationToken(authToken string) {
	z.options.authenticationToken = authToken
}

func (z *ZyncOptionsBuilder) DatabasePassword(dbPass string) {
	z.options.databasePassword = dbPass
}

func (z *ZyncOptionsBuilder) SecretKeyBase(secretKeyBase string) {
	z.options.secretKeyBase = secretKeyBase
}

func (z *ZyncOptionsBuilder) DatabaseURL(dbURL string) {
	z.options.databaseURL = &dbURL
}

func (z *ZyncOptionsBuilder) ContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	z.options.containerResourceRequirements = &resourceRequirements
}

func (z *ZyncOptionsBuilder) QueContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	z.options.queContainerResourceRequirements = &resourceRequirements
}

func (z *ZyncOptionsBuilder) DatabaseContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	z.options.databaseContainerResourceRequirements = &resourceRequirements
}

func (z *ZyncOptionsBuilder) Build() (*ZyncOptions, error) {
	err := z.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	z.setNonRequiredOptions()

	return &z.options, nil
}

func (z *ZyncOptionsBuilder) setRequiredOptions() error {
	if z.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if z.options.authenticationToken == "" {
		return fmt.Errorf("no Authentication Token has been provided")
	}
	if z.options.databasePassword == "" {
		return fmt.Errorf("no Database Password has been provided")
	}
	if z.options.secretKeyBase == "" {
		return fmt.Errorf("no Secret Key Base has been provided")
	}

	return nil
}

func (z *ZyncOptionsBuilder) setNonRequiredOptions() {
	defaultDatabaseURL := "postgresql://zync:" + z.options.databasePassword + "@zync-database:5432/zync_production"
	if z.options.databaseURL == nil {
		z.options.databaseURL = &defaultDatabaseURL
	}

	if z.options.containerResourceRequirements == nil {
		z.options.containerResourceRequirements = z.defaultContainerResourceRequirements()
	}

	if z.options.queContainerResourceRequirements == nil {
		z.options.queContainerResourceRequirements = z.defaultQueContainerResourceRequirements()
	}

	if z.options.databaseContainerResourceRequirements == nil {
		z.options.databaseContainerResourceRequirements = z.defaultDatabaseContainerResourceRequirements()
	}
}

func (z *ZyncOptionsBuilder) defaultContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("250M"),
		},
	}
}

func (z *ZyncOptionsBuilder) defaultQueContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("250M"),
		},
	}
}

func (z *ZyncOptionsBuilder) defaultDatabaseContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("2G"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("250M"),
		},
	}
}
