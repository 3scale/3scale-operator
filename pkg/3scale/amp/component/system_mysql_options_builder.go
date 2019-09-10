package component

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SystemMysqlOptions struct {
	// mysqlRequiredOptions
	appLabel     string
	databaseName string
	user         string
	password     string
	rootPassword string
	databaseURL  string

	// non-required options
	containerResourceRequirements *v1.ResourceRequirements
}

type SystemMysqlOptionsBuilder struct {
	options SystemMysqlOptions
}

func (m *SystemMysqlOptionsBuilder) AppLabel(appLabel string) {
	m.options.appLabel = appLabel
}

func (m *SystemMysqlOptionsBuilder) DatabaseName(databaseName string) {
	m.options.databaseName = databaseName
}

func (m *SystemMysqlOptionsBuilder) User(user string) {
	m.options.user = user
}

func (m *SystemMysqlOptionsBuilder) Password(password string) {
	m.options.password = password
}

func (m *SystemMysqlOptionsBuilder) RootPassword(rootPassword string) {
	m.options.rootPassword = rootPassword
}

func (m *SystemMysqlOptionsBuilder) DatabaseURL(url string) {
	m.options.databaseURL = url
}

func (m *SystemMysqlOptionsBuilder) ContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	m.options.containerResourceRequirements = &resourceRequirements
}

func (m *SystemMysqlOptionsBuilder) Build() (*SystemMysqlOptions, error) {
	err := m.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	m.setNonRequiredOptions()

	return &m.options, nil
}

func (m *SystemMysqlOptionsBuilder) setRequiredOptions() error {
	if m.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if m.options.databaseName == "" {
		return fmt.Errorf("no Database Name has been provided")
	}
	if m.options.user == "" {
		return fmt.Errorf("no User has been provided")
	}
	if m.options.password == "" {
		return fmt.Errorf("no Password has been provided")
	}
	if m.options.rootPassword == "" {
		return fmt.Errorf("no Root Password has been provided")
	}
	if m.options.databaseURL == "" {
		return fmt.Errorf("no Database URL has been provided")
	}

	return nil
}

func (m *SystemMysqlOptionsBuilder) setNonRequiredOptions() {
	if m.options.containerResourceRequirements == nil {
		m.options.containerResourceRequirements = m.defaultContainerResourceRequirements()
	}
}

func (m *SystemMysqlOptionsBuilder) defaultContainerResourceRequirements() *v1.ResourceRequirements {
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
