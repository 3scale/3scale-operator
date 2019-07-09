package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ZyncSecretName                         = "zync"
	ZyncSecretKeyBaseFieldName             = "SECRET_KEY_BASE"
	ZyncSecretDatabaseURLFieldName         = "DATABASE_URL"
	ZyncSecretDatabasePasswordFieldName    = "ZYNC_DATABASE_PASSWORD"
	ZyncSecretAuthenticationTokenFieldName = "ZYNC_AUTHENTICATION_TOKEN"
)

type Zync struct {
	options []string
	Options *ZyncOptions
}

type ZyncOptionsProvider interface {
	GetZyncOptions() *ZyncOptions
}

func NewZync(options *ZyncOptions) *Zync {
	return &Zync{Options: options}
}

func (zync *Zync) Objects() []common.KubernetesObject {
	zyncQueRole := zync.buildZyncQueRole()
	zyncQueServiceAccount := zync.buildZyncQueServiceAccount()
	zyncQueRoleBinding := zync.buildZyncQueRoleBinding()
	zyncDeploymentConfig := zync.buildZyncDeploymentConfig()
	zyncQueDeploymentConfig := zync.buildZyncQueDeploymentConfig()
	zyncDatabaseDeploymentConfig := zync.buildZyncDatabaseDeploymentConfig()
	zyncService := zync.buildZyncService()
	zyncDatabaseService := zync.buildZyncDatabaseService()
	zyncSecret := zync.buildZyncSecret()

	objects := []common.KubernetesObject{
		zyncQueRole,
		zyncQueServiceAccount,
		zyncQueRoleBinding,
		zyncDeploymentConfig,
		zyncQueDeploymentConfig,
		zyncDatabaseDeploymentConfig,
		zyncService,
		zyncDatabaseService,
		zyncSecret,
	}
	return objects
}

func (zync *Zync) buildZyncSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ZyncSecretName,
			Labels: map[string]string{
				"app":                  zync.Options.appLabel,
				"threescale_component": "zync",
			},
		},
		StringData: map[string]string{
			ZyncSecretKeyBaseFieldName:             zync.Options.secretKeyBase,
			ZyncSecretDatabaseURLFieldName:         *zync.Options.databaseURL,
			ZyncSecretDatabasePasswordFieldName:    zync.Options.databasePassword,
			ZyncSecretAuthenticationTokenFieldName: zync.Options.authenticationToken,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (zync *Zync) buildZyncQueServiceAccount() *v1.ServiceAccount {
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

func (zync *Zync) buildZyncQueRoleBinding() *rbacv1.RoleBinding {
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

func (zync *Zync) buildZyncQueRole() *rbacv1.Role {
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
					"update",
				},
			},
		},
	}
}

func (zync *Zync) buildZyncDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync",
			Labels: map[string]string{
				"app":                  zync.Options.appLabel,
				"threescale_component": "zync",
			},
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
							"zync",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-zync:latest",
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "zync"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                  zync.Options.appLabel,
						"deploymentConfig":     "zync",
						"threescale_component": "zync",
					},
				},
				Spec: v1.PodSpec{
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
												Name: "zync",
											},
											Key: "DATABASE_URL",
										},
									},
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "zync",
							Image: "amp-zync:latest",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP},
							},
							Env: zync.commonZyncEnvVars(),
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
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("150m"),
									v1.ResourceMemory: resource.MustParse("250M"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) commonZyncEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		envVarFromValue("RAILS_LOG_TO_STDOUT", "true"),
		envVarFromValue("RAILS_ENV", "production"),
		envVarFromSecret("DATABASE_URL", "zync", "DATABASE_URL"),
		envVarFromSecret("SECRET_KEY_BASE", "zync", "SECRET_KEY_BASE"),
		envVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN"),
		v1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		v1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}
}
func (zync *Zync) buildZyncQueDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que",
			Labels: map[string]string{
				"app":                  zync.Options.appLabel,
				"threescale_component": "zync",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
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
							Name: "amp-zync:latest",
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
					Labels: map[string]string{
						"app":              zync.Options.appLabel,
						"deploymentConfig": "zync-que",
					},
				},
				Spec: v1.PodSpec{
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
								v1.ContainerPort{Name: "metrics", ContainerPort: 9394, Protocol: v1.ProtocolTCP},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("250m"),
									v1.ResourceMemory: resource.MustParse("250M"),
								},
							},
							Env: zync.commonZyncEnvVars(),
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) buildZyncDatabaseDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-database",
			Labels: map[string]string{
				"app":                          zync.Options.appLabel,
				"threescale_component":         "zync",
				"threescale_component_element": "database",
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
							"postgresql",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "zync-database-postgresql:latest",
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
					Labels: map[string]string{
						"deploymentConfig":             "zync-database",
						"app":                          zync.Options.appLabel,
						"threescale_component":         "zync",
						"threescale_component_element": "database",
					},
				},
				Spec: v1.PodSpec{
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
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("250m"),
									v1.ResourceMemory: resource.MustParse("2G"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("250M"),
								},
							},
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

func (zync *Zync) buildZyncService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync",
			Labels: map[string]string{
				"app":                  zync.Options.appLabel,
				"threescale_component": "zync",
			},
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
			Selector: map[string]string{"deploymentConfig": "zync"},
		},
	}
}

func (zync *Zync) buildZyncDatabaseService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-database",
			Labels: map[string]string{
				"app":                          zync.Options.appLabel,
				"threescale_component":         "zync",
				"threescale_component_element": "database",
			},
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
