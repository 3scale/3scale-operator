package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type SystemPostgreSQL struct {
	Options *SystemPostgreSQLOptions
}

func NewSystemPostgreSQL(options *SystemPostgreSQLOptions) *SystemPostgreSQL {
	return &SystemPostgreSQL{Options: options}
}

func (p *SystemPostgreSQL) Objects() []common.KubernetesObject {
	deploymentConfig := p.DeploymentConfig()
	service := p.Service()
	persistentVolumeClaim := p.DataPersistentVolumeClaim()
	systemDatabaseSecret := p.SystemDatabaseSecret()

	objects := []common.KubernetesObject{
		deploymentConfig,
		service,
		persistentVolumeClaim,
		systemDatabaseSecret,
	}

	return objects
}

func (p *SystemPostgreSQL) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-postgresql",
			Labels: map[string]string{
				"app":                          p.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "postgresql",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "system-postgresql",
					Protocol:   v1.ProtocolTCP,
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-postgresql"},
		},
	}
}

func (p *SystemPostgreSQL) DataPersistentVolumeClaim() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "postgresql-data",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "postgresql", "app": p.Options.appLabel},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func (p *SystemPostgreSQL) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-postgresql",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "postgresql", "app": p.Options.appLabel},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRecreate,
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"system-postgresql",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "system-postgresql:latest",
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-postgresql"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "postgresql", "app": p.Options.appLabel, "deploymentConfig": "system-postgresql"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "postgresql-data",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "postgresql-data",
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "system-postgresql",
							Image: "system-postgresql:latest",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: 5432,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Env: []v1.EnvVar{
								envVarFromSecret("POSTGRESQL_USER", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseUserFieldName),
								envVarFromSecret("POSTGRESQL_PASSWORD", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabasePasswordFieldName),
								// TODO This should be gathered from secrets but we cannot set them because the URL field of the system-database secret
								// is already formed from this contents and we would have duplicate information. Once OpenShift templates
								// are deprecated we should be able to change this.
								envVarFromValue("POSTGRESQL_DATABASE", p.Options.databaseName),
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("2Gi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("250m"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "postgresql-data",
									MountPath: "/var/lib/pgsql/data",
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 5432,
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      0,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "psql -h 127.0.0.1 -U $POSTGRESQL_USER -q -d $POSTGRESQL_DATABASE -c 'SELECT 1'"}},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}
}

// Each database is responsible to create the needed secrets for the other components
func (p *SystemPostgreSQL) SystemDatabaseSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemDatabaseSecretName,
			Labels: map[string]string{
				"app":                  p.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseUserFieldName:     p.Options.user,
			SystemSecretSystemDatabasePasswordFieldName: p.Options.password,
			SystemSecretSystemDatabaseURLFieldName:      p.Options.databaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}
