package component

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SystemPostgreSQLOptions struct {
	//systemPostgreSQLNonRequiredOptions
	containerResourceRequirements *v1.ResourceRequirements

	//systemPostgreSQLRequiredOptions
	ampRelease   string
	appLabel     string
	image        string
	user         string
	password     string
	databaseName string
	databaseURL  string
}

type SystemPostgreSQLOptionsBuilder struct {
	options SystemPostgreSQLOptions
}

func (b *SystemPostgreSQLOptionsBuilder) AppLabel(appLabel string) {
	b.options.appLabel = appLabel
}

func (b *SystemPostgreSQLOptionsBuilder) DatabaseName(databaseName string) {
	b.options.databaseName = databaseName
}

func (b *SystemPostgreSQLOptionsBuilder) DatabaseURL(url string) {
	b.options.databaseURL = url
}

func (b *SystemPostgreSQLOptionsBuilder) User(user string) {
	b.options.user = user
}

func (b *SystemPostgreSQLOptionsBuilder) Password(password string) {
	b.options.password = password
}

func (b *SystemPostgreSQLOptionsBuilder) ContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	b.options.containerResourceRequirements = &resourceRequirements
}

func (b *SystemPostgreSQLOptionsBuilder) Build() (*SystemPostgreSQLOptions, error) {
	err := b.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	b.setNonRequiredOptions()

	return &b.options, nil
}

func (b *SystemPostgreSQLOptionsBuilder) setRequiredOptions() error {
	if b.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if b.options.databaseName == "" {
		return fmt.Errorf("no Database Name has been provided")
	}
	if b.options.user == "" {
		return fmt.Errorf("no User has been provided")
	}
	if b.options.password == "" {
		return fmt.Errorf("no Password has been provided")
	}
	if b.options.databaseURL == "" {
		return fmt.Errorf("no Database URL has been provided")
	}

	return nil
}

func (b *SystemPostgreSQLOptionsBuilder) setNonRequiredOptions() {
	if b.options.containerResourceRequirements == nil {
		b.options.containerResourceRequirements = b.defaultContainerResourceRequirements()
	}
}

func (m *SystemPostgreSQLOptionsBuilder) defaultContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
}
