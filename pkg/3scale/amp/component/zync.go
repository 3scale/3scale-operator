package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

type ZyncOptions struct {
	zyncNonRequiredOptions
	zyncRequiredOptions
}

type zyncRequiredOptions struct {
	appLabel            string
	authenticationToken string
	databasePassword    string
	secretKeyBase       string
}

type zyncNonRequiredOptions struct {
	databaseURL *string
}

type ZyncOptionsProvider interface {
	GetZyncOptions() *ZyncOptions
}
type CLIZyncOptionsProvider struct {
}

func (o *CLIZyncOptionsProvider) GetZyncOptions() (*ZyncOptions, error) {
	zob := ZyncOptionsBuilder{}
	zob.AppLabel("${APP_LABEL}")
	zob.AuthenticationToken("${ZYNC_AUTHENTICATION_TOKEN}")
	zob.DatabasePassword("${ZYNC_DATABASE_PASSWORD}")
	zob.SecretKeyBase("${ZYNC_SECRET_KEY_BASE}")
	res, err := zob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Zync Options - %s", err)
	}
	return res, nil
}

func NewZync(options []string) *Zync {
	zync := &Zync{
		options: options,
	}
	return zync
}

func (zync *Zync) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIZyncOptionsProvider{}
	zyncOpts, err := optionsProvider.GetZyncOptions()
	_ = err
	zync.Options = zyncOpts
	zync.buildParameters(template)
	zync.addObjectsIntoTemplate(template)
}

func (zync *Zync) buildObjects() []runtime.RawExtension {
	zyncDeploymentConfig := zync.buildZyncDeploymentConfig()
	zyncDatabaseDeploymentConfig := zync.buildZyncDatabaseDeploymentConfig()
	zyncService := zync.buildZyncService()
	zyncDatabaseService := zync.buildZyncDatabaseService()
	zyncSecret := zync.buildZyncSecret()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: zyncDeploymentConfig},
		runtime.RawExtension{Object: zyncDatabaseDeploymentConfig},
		runtime.RawExtension{Object: zyncService},
		runtime.RawExtension{Object: zyncDatabaseService},
		runtime.RawExtension{Object: zyncSecret},
	}
	return objects
}

func (zync *Zync) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (zync *Zync) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "ZYNC_DATABASE_PASSWORD",
			DisplayName: "Zync Database PostgreSQL Connection Password",
			Description: "Password for the Zync Database PostgreSQL connection user.",
			Generate:    "expression",
			From:        "[a-zA-Z0-9]{16}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "ZYNC_SECRET_KEY_BASE",
			Generate: "expression",
			From:     "[a-zA-Z0-9]{16}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ZYNC_AUTHENTICATION_TOKEN",
			Generate: "expression",
			From:     "[a-zA-Z0-9]{16}",
			Required: true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

func (zync *Zync) GetObjects() ([]runtime.RawExtension, error) {
	objects := zync.buildObjects()
	return objects, nil
}

func (zync *Zync) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := zync.buildObjects()
	template.Objects = append(template.Objects, objects...)
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

func (zync *Zync) buildZyncCronDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-cron",
			Labels: map[string]string{"app": "Zync"},
		},
		Spec: appsv1.DeploymentConfigSpec{
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
						StrVal: "25%"}},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{"zync-cron"},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-zync:latest"}}}, // TODO decide what to do with references to ImageStreams
			},
			Replicas: 1,
			Selector: map[string]string{"name": "zync-cron"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"name": "zync-cron"},
				},
				Spec: v1.PodSpec{Containers: []v1.Container{
					v1.Container{
						Name:  "zync-cron",
						Image: "amp-zync:latest", // TODO decide what to do with references to ImageStreams
						Args:  []string{"zync-cron"},
						Env: []v1.EnvVar{
							v1.EnvVar{
								Name:  "CONFIG_REDIS_PROXY",
								Value: "redis://zync-redis:6379/0", // TODO decide what to do with references to the 'zync-redis' service
							}, v1.EnvVar{
								Name: "CONFIG_REDIS_SENTINEL_HOSTS",
							}, v1.EnvVar{
								Name: "CONFIG_REDIS_SENTINEL_ROLE",
							}, v1.EnvVar{
								Name:  "CONFIG_QUEUES_MASTER_NAME",
								Value: "redis://zync-redis:6379/1", // TODO decide what to do with references to the 'zync-redis' service
							}, v1.EnvVar{
								Name: "CONFIG_QUEUES_SENTINEL_HOSTS",
							}, v1.EnvVar{
								Name: "CONFIG_QUEUES_SENTINEL_ROLE",
							}, v1.EnvVar{
								Name:  "RACK_ENV",
								Value: "production",
							},
						},
						Resources: v1.ResourceRequirements{
							Limits:   v1.ResourceList{"cpu": resource.MustParse("150m")},
							Requests: v1.ResourceList{"cpu": resource.MustParse("50m")},
						},
						ImagePullPolicy: v1.PullIfNotPresent,
					},
				},
					ServiceAccountName: "amp",
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
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "RAILS_LOG_TO_STDOUT",
									Value: "true",
								}, v1.EnvVar{
									Name:  "RAILS_ENV",
									Value: "production",
								}, v1.EnvVar{
									Name: "DATABASE_URL",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "zync",
											},
											Key: "DATABASE_URL",
										},
									},
								}, v1.EnvVar{
									Name: "SECRET_KEY_BASE",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "zync",
											},
											Key: "SECRET_KEY_BASE",
										},
									},
								}, v1.EnvVar{
									Name: "ZYNC_AUTHENTICATION_TOKEN",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "zync",
											},
											Key: "ZYNC_AUTHENTICATION_TOKEN",
										},
									},
								},
							},
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
							Name: "postgresql:latest",
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
