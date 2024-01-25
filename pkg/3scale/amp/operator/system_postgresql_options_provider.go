package operator

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"net/url"
	"strings"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SystemPostgresqlOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	options      *component.SystemPostgreSQLOptions
	secretSource *helper.SecretSource
}

func NewSystemPostgresqlOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *SystemPostgresqlOptionsProvider {
	return &SystemPostgresqlOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		options:      component.NewSystemPostgreSQLOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (s *SystemPostgresqlOptionsProvider) GetSystemPostgreSQLOptions() (*component.SystemPostgreSQLOptions, error) {
	s.options.ImageTag = product.ThreescaleRelease
	s.options.CommonLabels = s.commonLabels()
	s.options.DeploymentLabels = s.deploymentLabels()
	s.options.PodTemplateLabels = s.podTemplateLabels()

	err := s.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetSystemPostgreSQLOptions reading secret options: %w", err)
	}

	s.setResourceRequirementsOptions()
	s.setPersistentVolumeClaimOptions()
	s.setNodeAffinityAndTolerationsOptions()
	s.setPriorityClassNames()
	s.setTopologySpreadConstraints()
	s.setPodTemplateAnnotations()

	err = s.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetSystemPostgreSQLOptions validating: %w", err)
	}
	return s.options, nil
}

func (s *SystemPostgresqlOptionsProvider) setSecretBasedOptions() error {
	val, err := s.secretSource.FieldValue(
		component.SystemSecretSystemDatabaseSecretName,
		component.SystemSecretSystemDatabaseUserFieldName,
		component.DefaultSystemPostgresqlUser())
	if err != nil {
		return err
	}
	s.options.User = val

	val, err = s.secretSource.FieldValue(
		component.SystemSecretSystemDatabaseSecretName,
		component.SystemSecretSystemDatabasePasswordFieldName,
		component.DefaultSystemPostgresqlPassword())
	if err != nil {
		return err
	}
	s.options.Password = val

	val, err = s.secretSource.FieldValue(
		component.SystemSecretSystemDatabaseSecretName,
		component.SystemSecretSystemDatabaseURLFieldName,
		component.DefaultSystemPostgresqlDatabaseURL(s.options.User, s.options.Password, component.DefaultSystemPostgresqlDatabaseName()))
	if err != nil {
		return err
	}
	s.options.DatabaseURL = val

	// databaseURL processing
	urlObj, err := s.databaseURLIsValid(s.options.DatabaseURL)
	if err != nil {
		return err
	}

	// Remove possible leading slash in URL Path
	s.options.DatabaseName = strings.TrimPrefix(urlObj.Path, "/")
	return nil
}

func (s *SystemPostgresqlOptionsProvider) databaseURLIsValid(rawURL string) (*url.URL, error) {
	resultURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("'%s' field of '%s' secret must have 'scheme://user:password@host/path' format", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if resultURL.Scheme != "postgresql" {
		return nil, fmt.Errorf("'%s' field of '%s' secret must contain 'postgresql' as the scheme part", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}

	if resultURL.User == nil {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must be provided", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	if resultURL.User.Username() == "" {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain a username", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
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

func (s *SystemPostgresqlOptionsProvider) setResourceRequirementsOptions() {
	if *s.apimanager.Spec.ResourceRequirementsEnabled {
		s.options.ContainerResourceRequirements = component.DefaultSystemPostgresqlResourceRequirements()
	} else {
		s.options.ContainerResourceRequirements = v1.ResourceRequirements{}
	}

	// Deployment-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if s.apimanager.Spec.System.DatabaseSpec != nil &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL != nil &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Resources != nil {
		s.options.ContainerResourceRequirements = *s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Resources
	}
}

func (s *SystemPostgresqlOptionsProvider) setPersistentVolumeClaimOptions() {
	var volumeName *string
	storageRequests := component.DefaultSystemPostgresqlStorageResources()

	if s.apimanager.Spec.System.DatabaseSpec != nil &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL != nil &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PersistentVolumeClaimSpec != nil {

		s.options.PVCStorageClass = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PersistentVolumeClaimSpec.StorageClassName
		volumeName = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PersistentVolumeClaimSpec.VolumeName
		if s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PersistentVolumeClaimSpec.Resources != nil {
			storageRequests = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PersistentVolumeClaimSpec.Resources.Requests
		}
	}

	s.options.PVCVolumeName = volumeName
	s.options.PVCStorageRequests = storageRequests
}

func (s *SystemPostgresqlOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	if s.apimanager.Spec.System.DatabaseSpec != nil && s.apimanager.Spec.System.DatabaseSpec.PostgreSQL != nil {
		s.options.Affinity = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Affinity
		s.options.Tolerations = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Tolerations
	}
}

func (s *SystemPostgresqlOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *s.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (s *SystemPostgresqlOptionsProvider) deploymentLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "postgresql"
	return labels
}

func (s *SystemPostgresqlOptionsProvider) podTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-postgresql", helper.ApplicationType)

	for k, v := range s.deploymentLabels() {
		labels[k] = v
	}

	if s.apimanager.IsSystemPostgreSQLEnabled() {
		for k, v := range s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Labels {
			labels[k] = v
		}
	}

	labels[reconcilers.DeploymentLabelSelector] = "system-postgresql"

	return labels
}

func (s *SystemPostgresqlOptionsProvider) setPriorityClassNames() {
	if s.apimanager.IsSystemPostgreSQLEnabled() &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PriorityClassName != nil {
		s.options.PriorityClassName = *s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.PriorityClassName
	}
}

func (s *SystemPostgresqlOptionsProvider) setTopologySpreadConstraints() {
	if s.apimanager.IsSystemPostgreSQLEnabled() &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.TopologySpreadConstraints != nil {
		s.options.TopologySpreadConstraints = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.TopologySpreadConstraints
	}
}

func (s *SystemPostgresqlOptionsProvider) setPodTemplateAnnotations() {
	if s.apimanager.IsSystemPostgreSQLEnabled() {
		s.options.PodTemplateAnnotations = s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Annotations
	}
}
