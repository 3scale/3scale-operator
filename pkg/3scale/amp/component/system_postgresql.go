package component

import (
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SystemPostgreSQLDeploymentName = "system-postgresql"
)

type SystemPostgreSQL struct {
	Options *SystemPostgreSQLOptions
}

func NewSystemPostgreSQL(options *SystemPostgreSQLOptions) *SystemPostgreSQL {
	return &SystemPostgreSQL{Options: options}
}

func (p *SystemPostgreSQL) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-postgresql",
			Labels: p.Options.DeploymentLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "system-postgresql",
					Protocol:   v1.ProtocolTCP,
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "system-postgresql"},
		},
	}
}

func (p *SystemPostgreSQL) DataPersistentVolumeClaim() *v1.PersistentVolumeClaim {
	volName := ""
	if p.Options.PVCVolumeName != nil {
		volName = *p.Options.PVCVolumeName
	}

	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "postgresql-data",
			Labels: p.Options.DeploymentLabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: p.Options.PVCStorageRequests,
				},
			},
			StorageClassName: p.Options.PVCStorageClass,
			VolumeName:       volName,
		},
	}
}

func (p *SystemPostgreSQL) Deployment(containerImage string) *k8sappsv1.Deployment {
	var postgresReplicas int32 = 1

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       reconcilers.DeploymentLabelSelector,
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemPostgreSQLDeploymentName,
			Labels: p.Options.DeploymentLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &postgresReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemPostgreSQLDeploymentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      p.Options.PodTemplateLabels,
					Annotations: p.Options.PodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:           p.Options.Affinity,
					Tolerations:        p.Options.Tolerations,
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Volumes: []v1.Volume{
						{
							Name: "postgresql-data",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "postgresql-data",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "system-postgresql",
							Image: containerImage,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 5432,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Env: []v1.EnvVar{
								helper.EnvVarFromSecret("POSTGRESQL_USER", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseUserFieldName),
								helper.EnvVarFromSecret("POSTGRESQL_PASSWORD", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabasePasswordFieldName),
								// TODO This should be gathered from secrets but we cannot set them because the URL field of the system-database secret
								// is already formed from this contents and we would have duplicate information. Once OpenShift templates
								// are deprecated we should be able to change this.
								helper.EnvVarFromValue("POSTGRESQL_DATABASE", p.Options.DatabaseName),
							},
							Resources: p.Options.ContainerResourceRequirements,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "postgresql-data",
									MountPath: "/var/lib/pgsql/data",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
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
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "psql -h 127.0.0.1 -U $POSTGRESQL_USER -q -d $POSTGRESQL_DATABASE -c 'SELECT 1'"},
									},
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
					PriorityClassName:         p.Options.PriorityClassName,
					TopologySpreadConstraints: p.Options.TopologySpreadConstraints,
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
			Name:   SystemSecretSystemDatabaseSecretName,
			Labels: p.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseUserFieldName:     p.Options.User,
			SystemSecretSystemDatabasePasswordFieldName: p.Options.Password,
			SystemSecretSystemDatabaseURLFieldName:      p.Options.DatabaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}
