package component

import (
	"fmt"

	"k8s.io/api/policy/v1beta1"

	"github.com/3scale/3scale-operator/pkg/helper"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ZyncName = "zync"
)

const (
	ZyncSecretName                         = "zync"
	ZyncSecretKeyBaseFieldName             = "SECRET_KEY_BASE"
	ZyncSecretDatabaseURLFieldName         = "DATABASE_URL"
	ZyncSecretDatabasePasswordFieldName    = "ZYNC_DATABASE_PASSWORD"
	ZyncSecretAuthenticationTokenFieldName = "ZYNC_AUTHENTICATION_TOKEN"
)

const (
	ZyncMetricsPort    = 9393
	ZyncQueMetricsPort = 9394
)

type Zync struct {
	Options *ZyncOptions
}

func NewZync(options *ZyncOptions) *Zync {
	return &Zync{Options: options}
}

func (zync *Zync) Secret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncSecretName,
			Labels: zync.Options.CommonLabels,
		},
		StringData: map[string]string{
			ZyncSecretKeyBaseFieldName:             zync.Options.SecretKeyBase,
			ZyncSecretDatabaseURLFieldName:         zync.Options.DatabaseURL,
			ZyncSecretDatabasePasswordFieldName:    zync.Options.DatabasePassword,
			ZyncSecretAuthenticationTokenFieldName: zync.Options.AuthenticationToken,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (zync *Zync) QueServiceAccount() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que-sa",
		},
		ImagePullSecrets: []v1.LocalObjectReference{
			v1.LocalObjectReference{
				Name: "threescale-registry-auth",
			},
		},
	}
}

func (zync *Zync) QueRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que-rolebinding",
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: "zync-que-sa",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "zync-que-role",
		},
	}
}

func (zync *Zync) QueRole() *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que-role",
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"apps.openshift.io"},
				Resources: []string{
					"deploymentconfigs",
				},
				Verbs: []string{
					"get",
					"list",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{
					"pods",
					"replicationcontrollers",
				},
				Verbs: []string{
					"get",
					"list",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{
					"routes",
				},
				Verbs: []string{
					"get",
					"list",
					"create",
					"delete",
					"patch",
					"update",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{
					"routes/status",
				},
				Verbs: []string{
					"get",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{
					"routes/custom-host",
				},
				Verbs: []string{
					"create",
				},
			},
		},
	}
}

func (zync *Zync) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncName,
			Labels: zync.Options.CommonZyncLabels,
			Annotations: map[string]string{
				"prometheus.io/port":   "9393",
				"prometheus.io/scrape": "true",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"zync-db-svc",
							ZyncName,
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("amp-zync:%s", zync.Options.ImageTag),
						},
					},
				},
			},
			Replicas: zync.Options.ZyncReplicas,
			Selector: map[string]string{"deploymentConfig": ZyncName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: zync.Options.ZyncPodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:           zync.Options.ZyncAffinity,
					Tolerations:        zync.Options.ZyncTolerations,
					ServiceAccountName: "amp",
					InitContainers: []v1.Container{
						v1.Container{
							Name:  "zync-db-svc",
							Image: "amp-zync:latest",
							Command: []string{
								"bash",
								"-c",
								"bundle exec sh -c \"until rake boot:db; do sleep $SLEEP_SECONDS; done\"",
							}, Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "SLEEP_SECONDS",
									Value: "1",
								},
								v1.EnvVar{
									Name: "DATABASE_URL",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: ZyncSecretName,
											},
											Key: ZyncSecretDatabaseURLFieldName,
										},
									},
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  ZyncName,
							Image: "amp-zync:latest",
							Ports: zync.zyncPorts(),
							Env:   zync.commonZyncEnvVars(),
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Port:   intstr.FromInt(8080),
										Path:   "/status/live",
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      60,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    10,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/status/ready",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 100,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Resources: zync.Options.ContainerResourceRequirements,
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) commonZyncEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromValue("RAILS_LOG_TO_STDOUT", "true"),
		helper.EnvVarFromValue("RAILS_ENV", "production"),
		helper.EnvVarFromSecret("DATABASE_URL", "zync", "DATABASE_URL"),
		helper.EnvVarFromSecret("SECRET_KEY_BASE", "zync", "SECRET_KEY_BASE"),
		helper.EnvVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN"),
		v1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath:  "metadata.name",
					APIVersion: "v1",
				},
			},
		},
		v1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath:  "metadata.namespace",
					APIVersion: "v1",
				},
			},
		},
	}
}
func (zync *Zync) QueDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-que",
			Labels: zync.Options.CommonZyncQueLabels,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: zync.Options.ZyncQueReplicas,
			Selector: map[string]string{"deploymentConfig": "zync-que"},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					UpdatePeriodSeconds: &[]int64{1}[0],
					IntervalSeconds:     &[]int64{1}[0],
					TimeoutSeconds:      &[]int64{600}[0],
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
				},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"que",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("amp-zync:%s", zync.Options.ImageTag),
						},
					},
				},
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/port":   "9394",
						"prometheus.io/scrape": "true",
					},
					Labels: zync.Options.ZyncQuePodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:                      zync.Options.ZyncQueAffinity,
					Tolerations:                   zync.Options.ZyncQueTolerations,
					ServiceAccountName:            "zync-que-sa",
					RestartPolicy:                 v1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					Containers: []v1.Container{
						v1.Container{
							Name:            "que",
							Command:         []string{"/usr/bin/bash"},
							Args:            []string{"-c", "bundle exec rake 'que[--worker-count 10]'"},
							Image:           "amp-zync:latest",
							ImagePullPolicy: v1.PullAlways,
							LivenessProbe: &v1.Probe{
								FailureThreshold:    3,
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      60,
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Port:   intstr.FromInt(9394),
										Path:   "/metrics",
										Scheme: v1.URISchemeHTTP,
									},
								},
							},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{Name: "metrics", ContainerPort: ZyncQueMetricsPort, Protocol: v1.ProtocolTCP},
							},
							Resources: zync.Options.QueContainerResourceRequirements,
							Env:       zync.commonZyncEnvVars(),
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) DatabaseDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-database",
			Labels: zync.Options.CommonZyncDatabaseLabels,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"postgresql",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("zync-database-postgresql:%s", zync.Options.DatabaseImageTag),
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "zync-database"},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRecreate,
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: zync.Options.ZyncDatabasePodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:           zync.Options.ZyncDatabaseAffinity,
					Tolerations:        zync.Options.ZyncDatabaseTolerations,
					RestartPolicy:      v1.RestartPolicyAlways,
					ServiceAccountName: "amp",
					Containers: []v1.Container{
						v1.Container{
							Name:  "postgresql",
							Image: " ",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: 5432,
									Protocol:      v1.ProtocolTCP},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "zync-database-data",
									MountPath: "/var/lib/pgsql/data",
								},
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "POSTGRESQL_USER",
									Value: "zync",
								}, v1.EnvVar{
									Name: "POSTGRESQL_PASSWORD",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "zync",
											},
											Key: "ZYNC_DATABASE_PASSWORD",
										},
									},
								}, v1.EnvVar{
									Name:  "POSTGRESQL_DATABASE",
									Value: "zync_production",
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt(5432),
									},
								},
								TimeoutSeconds:      1,
								InitialDelaySeconds: 30,
							},
							ReadinessProbe: &v1.Probe{
								TimeoutSeconds:      1,
								InitialDelaySeconds: 5,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "psql -h 127.0.0.1 -U zync -q -d zync_production -c 'SELECT 1'"},
									},
								},
							},
							Resources: zync.Options.DatabaseContainerResourceRequirements,
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "zync-database-data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: v1.StorageMediumDefault,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncName,
			Labels: zync.Options.CommonZyncLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "8080-tcp",
					Protocol:   v1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{"deploymentConfig": ZyncName},
		},
	}
}

func (zync *Zync) DatabaseService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-database",
			Labels: zync.Options.CommonZyncDatabaseLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "postgresql",
					Protocol:   v1.ProtocolTCP,
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
			Selector: map[string]string{"deploymentConfig": "zync-database"},
		},
	}
}

func (zync *Zync) ZyncPodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncName,
			Labels: zync.Options.CommonZyncLabels,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": ZyncName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (zync *Zync) QuePodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-que",
			Labels: zync.Options.CommonZyncQueLabels,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": "zync-que"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (zync *Zync) zyncPorts() []v1.ContainerPort {
	ports := []v1.ContainerPort{
		v1.ContainerPort{ContainerPort: 8080, Protocol: v1.ProtocolTCP},
	}

	if zync.Options.ZyncMetrics {
		ports = append(ports, v1.ContainerPort{Name: "metrics", ContainerPort: ZyncMetricsPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}
