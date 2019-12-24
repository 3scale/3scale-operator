package operator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SystemMysqlOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	mysqlOptions *component.SystemMysqlOptions
	secretSource *helper.SecretSource
}

func NewSystemMysqlOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *SystemMysqlOptionsProvider {
	return &SystemMysqlOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		mysqlOptions: component.NewSystemMysqlOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (s *SystemMysqlOptionsProvider) GetMysqlOptions() (*component.SystemMysqlOptions, error) {
	s.mysqlOptions.AppLabel = *s.apimanager.Spec.AppLabel

	err := s.setSecretBasedOptions()
	if err != nil {
		return nil, err
	}

	s.setResourceRequirementsOptions()

	err = s.mysqlOptions.Validate()
	return s.mysqlOptions, err
}

func (s *SystemMysqlOptionsProvider) setSecretBasedOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&s.mysqlOptions.User,
			component.SystemSecretSystemDatabaseSecretName,
			component.SystemSecretSystemDatabaseUserFieldName,
			component.DefaultSystemMysqlUser(),
		},
		{
			&s.mysqlOptions.Password,
			component.SystemSecretSystemDatabaseSecretName,
			component.SystemSecretSystemDatabasePasswordFieldName,
			component.DefaultSystemMysqlPassword(),
		},
		{
			&s.mysqlOptions.DatabaseURL,
			component.SystemSecretSystemDatabaseSecretName,
			component.SystemSecretSystemDatabaseURLFieldName,
			component.DefaultSystemMysqlDatabaseURL(component.DefaultSystemMysqlRootPassword(), component.DefaultSystemMysqlDatabaseName()),
		},
	}

	for _, option := range cases {
		val, err := s.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		// not nil value is ensured
		*option.field = *val
	}

	// databaseURL processing
	urlObj, err := s.systemDatabaseURLIsValid(s.mysqlOptions.DatabaseURL)
	if err != nil {
		return err
	}

	// Remove possible leading slash in URL Path
	s.mysqlOptions.DatabaseName = strings.TrimPrefix(urlObj.Path, "/")
	dbRootPassword, _ := urlObj.User.Password()
	s.mysqlOptions.RootPassword = dbRootPassword

	return nil
}

func (s *SystemMysqlOptionsProvider) systemDatabaseURLIsValid(rawURL string) (*url.URL, error) {
	resultURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("'%s' field of '%s' secret must have 'scheme://user:password@host/path' format", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if resultURL.Scheme != "mysql2" {
		return nil, fmt.Errorf("'%s' field of '%s' secret must contain 'mysql2' as the scheme part", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}

	if resultURL.User == nil {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must be provided", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if resultURL.User.Username() == "" {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain a username", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if resultURL.User.Username() != "root" {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain 'root' as the username", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if _, set := resultURL.User.Password(); !set {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain a password", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}

	if resultURL.Host == "" {
		return nil, fmt.Errorf("host information in '%s' field of '%s' secret must be provided", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if resultURL.Path == "" {
		return nil, fmt.Errorf("database name in '%s' field of '%s' secret must be provided", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}

	return resultURL, nil
}

func (s *SystemMysqlOptionsProvider) setResourceRequirementsOptions() {
	if *s.apimanager.Spec.ResourceRequirementsEnabled {
		s.mysqlOptions.ContainerResourceRequirements = component.DefaultSystemMysqlResourceRequirements()
	} else {
		s.mysqlOptions.ContainerResourceRequirements = v1.ResourceRequirements{}
	}
}
