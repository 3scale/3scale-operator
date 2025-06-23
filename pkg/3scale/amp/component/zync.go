package component

import (
	"context"

	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ZyncName                   = "zync"
	ZyncQueDeploymentName      = "zync-que"
	ZyncDatabaseDeploymentName = "zync-database"
	ZyncInitContainerName      = "zync-db-svc"
)

const (
	ZyncSecretName                         = "zync"
	ZyncSecretKeyBaseFieldName             = "SECRET_KEY_BASE"
	ZyncSecretDatabaseURLFieldName         = "DATABASE_URL"
	ZyncSecretDatabasePasswordFieldName    = "ZYNC_DATABASE_PASSWORD"
	ZyncSecretAuthenticationTokenFieldName = "ZYNC_AUTHENTICATION_TOKEN"
	ZyncSecretDatabaseSslMode              = "DATABASE_SSL_MODE"
	ZyncSecretSslCa                        = "DB_SSL_CA"
	ZyncSecretSslCert                      = "DB_SSL_CERT"
	ZyncSecretSslKey                       = "DB_SSL_KEY"
)

const (
	ZyncMetricsPort    = 9393
	ZyncQueMetricsPort = 9394
)

const (
	ZyncSecretResverAnnotationPrefix = "apimanager.apps.3scale.net/zync-secret-resource-version-"
)

type Zync struct {
	Options *ZyncOptions
}

func NewZync(options *ZyncOptions) *Zync {
	return &Zync{Options: options}
}

func (zync *Zync) Secret() *v1.Secret {
	if zync.Options.ZyncDbTLSEnabled {
		return &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   ZyncSecretName,
				Labels: zync.Options.CommonZyncSecretLabels,
			},
			StringData: map[string]string{
				ZyncSecretKeyBaseFieldName:             zync.Options.SecretKeyBase,
				ZyncSecretDatabaseURLFieldName:         zync.Options.DatabaseURL,
				ZyncSecretDatabasePasswordFieldName:    zync.Options.DatabasePassword,
				ZyncSecretAuthenticationTokenFieldName: zync.Options.AuthenticationToken,
				ZyncSecretDatabaseSslMode:              zync.Options.DatabaseSslMode,
				ZyncSecretSslCa:                        zync.Options.DatabaseSslCa,
				ZyncSecretSslCert:                      zync.Options.DatabaseSslCert,
				ZyncSecretSslKey:                       zync.Options.DatabaseSslKey,
			},
			Type: v1.SecretTypeOpaque,
		}
	}
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncSecretName,
			Labels: zync.Options.CommonZyncSecretLabels,
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
		ImagePullSecrets: zync.Options.ZyncQueServiceAccountImagePullSecrets,
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
			{
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
			{
				APIGroups: []string{"apps"},
				Resources: []string{
					"deployments",
					"replicasets",
				},
				Verbs: []string{
					"get",
					"list",
				},
			},
			{
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
			{
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
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{
					"routes/status",
				},
				Verbs: []string{
					"get",
				},
			},
			{
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

func (zync *Zync) Deployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, ZyncName, zync.Options.Namespace, zync)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncName,
			Labels: zync.Options.CommonZyncLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &zync.Options.ZyncReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: ZyncName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      zync.Options.ZyncPodTemplateLabels,
					Annotations: zync.zyncPodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					Affinity:           zync.Options.ZyncAffinity,
					Tolerations:        zync.Options.ZyncTolerations,
					ServiceAccountName: "amp",
					InitContainers:     zync.zyncInit(containerImage),
					Containers: []v1.Container{
						{
							Name:  ZyncName,
							Image: containerImage,
							Ports: zync.zyncPorts(),
							Env:   zync.commonZyncEnvVars(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Port:   intstr.FromInt32(8080),
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
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/status/ready",
										Port:   intstr.FromInt32(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 100,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Resources:    zync.Options.ContainerResourceRequirements,
							VolumeMounts: zync.zyncVolumeMount(),
						},
					},
					Volumes:                   zync.zyncVolume(),
					PriorityClassName:         zync.Options.ZyncPriorityClassName,
					TopologySpreadConstraints: zync.Options.ZyncTopologySpreadConstraints,
				},
			},
		},
	}, nil
}

func (zync *Zync) commonZyncEnvVars() []v1.EnvVar {
	if zync.Options.ZyncDbTLSEnabled {
		return []v1.EnvVar{
			helper.EnvVarFromValue("RAILS_LOG_TO_STDOUT", "true"),
			helper.EnvVarFromValue("RAILS_ENV", "production"),
			helper.EnvVarFromSecret("DATABASE_URL", "zync", "DATABASE_URL"),
			helper.EnvVarFromSecret("SECRET_KEY_BASE", "zync", "SECRET_KEY_BASE"),
			helper.EnvVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN"),
			// SSL certs from secret
			helper.EnvVarFromSecretOptional("DB_SSL_CA", ZyncSecretName, ZyncSecretSslCa),
			helper.EnvVarFromSecretOptional("DB_SSL_CERT", ZyncSecretName, ZyncSecretSslCert),
			helper.EnvVarFromSecretOptional("DB_SSL_KEY", ZyncSecretName, ZyncSecretSslKey),
			helper.EnvVarFromSecretOptional("DATABASE_SSL_MODE", ZyncSecretName, ZyncSecretDatabaseSslMode),
			// SSL mount pat env vars
			helper.EnvVarFromValue("DATABASE_SSL_CA", helper.TlsCertPresent("DATABASE_SSL_CA", ZyncSecretName, zync.Options.ZyncDbTLSEnabled)),
			helper.EnvVarFromValue("DATABASE_SSL_CERT", helper.TlsCertPresent("DATABASE_SSL_CERT", ZyncSecretName, zync.Options.ZyncDbTLSEnabled)),
			helper.EnvVarFromValue("DATABASE_SSL_KEY", helper.TlsCertPresent("DATABASE_SSL_KEY", ZyncSecretName, zync.Options.ZyncDbTLSEnabled)),
			{
				Name: "POD_NAME",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath:  "metadata.name",
						APIVersion: "v1",
					},
				},
			},
			{
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
	return []v1.EnvVar{
		helper.EnvVarFromValue("RAILS_LOG_TO_STDOUT", "true"),
		helper.EnvVarFromValue("RAILS_ENV", "production"),
		helper.EnvVarFromSecret("DATABASE_URL", "zync", "DATABASE_URL"),
		helper.EnvVarFromSecret("SECRET_KEY_BASE", "zync", "SECRET_KEY_BASE"),
		helper.EnvVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN"),
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath:  "metadata.name",
					APIVersion: "v1",
				},
			},
		},
		{
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

func (zync *Zync) QueDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, ZyncQueDeploymentName, zync.Options.Namespace, zync)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncQueDeploymentName,
			Labels: zync.Options.CommonZyncQueLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &zync.Options.ZyncQueReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: ZyncQueDeploymentName,
				},
			},
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &k8sappsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      zync.Options.ZyncQuePodTemplateLabels,
					Annotations: zync.zyncQuePodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					Affinity:                      zync.Options.ZyncQueAffinity,
					Tolerations:                   zync.Options.ZyncQueTolerations,
					ServiceAccountName:            "zync-que-sa",
					RestartPolicy:                 v1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					InitContainers:                zync.zyncQueInit(),
					Containers: []v1.Container{
						{
							Name:            "que",
							Command:         []string{"/usr/bin/bash"},
							Args:            []string{"-c", "bundle exec rake 'que[--worker-count 10]'"},
							Image:           containerImage,
							ImagePullPolicy: v1.PullAlways,
							LivenessProbe: &v1.Probe{
								FailureThreshold:    3,
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      60,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Port:   intstr.FromInt32(9394),
										Path:   "/metrics",
										Scheme: v1.URISchemeHTTP,
									},
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: ZyncQueMetricsPort,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Resources:    zync.Options.QueContainerResourceRequirements,
							Env:          zync.commonZyncEnvVars(),
							VolumeMounts: zync.zyncVolumeMount(),
						},
					},
					Volumes:                   zync.zyncVolume(),
					PriorityClassName:         zync.Options.ZyncQuePriorityClassName,
					TopologySpreadConstraints: zync.Options.ZyncQueTopologySpreadConstraints,
				},
			},
		},
	}, nil
}

func (zync *Zync) DatabaseDeployment(containerImage string) *k8sappsv1.Deployment {
	var zyncDatabaseReplicas int32 = 1

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncDatabaseDeploymentName,
			Labels: zync.Options.CommonZyncDatabaseLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &zyncDatabaseReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: ZyncDatabaseDeploymentName,
				},
			},
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      zync.Options.ZyncDatabasePodTemplateLabels,
					Annotations: zync.Options.ZyncDatabasePodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:           zync.Options.ZyncDatabaseAffinity,
					Tolerations:        zync.Options.ZyncDatabaseTolerations,
					RestartPolicy:      v1.RestartPolicyAlways,
					ServiceAccountName: "amp",
					Containers: []v1.Container{
						{
							Name:  "postgresql",
							Image: containerImage,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 5432,
									Protocol:      v1.ProtocolTCP,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "zync-database-data",
									MountPath: "/var/lib/pgsql/data",
								},
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name:  "POSTGRESQL_USER",
									Value: "zync",
								},
								{
									Name: "POSTGRESQL_PASSWORD",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "zync",
											},
											Key: "ZYNC_DATABASE_PASSWORD",
										},
									},
								},
								{
									Name:  "POSTGRESQL_DATABASE",
									Value: "zync_production",
								},
								{
									Name:  "POSTGRESQL_LOG_DESTINATION",
									Value: "/dev/stderr",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt32(5432),
									},
								},
								TimeoutSeconds:      1,
								InitialDelaySeconds: 30,
							},
							ReadinessProbe: &v1.Probe{
								TimeoutSeconds:      1,
								InitialDelaySeconds: 5,
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "psql -h 127.0.0.1 -U zync -q -d zync_production -c 'SELECT 1'"},
									},
								},
							},
							Resources: zync.Options.DatabaseContainerResourceRequirements,
						},
					},
					PriorityClassName:         zync.Options.ZyncDatabasePriorityClassName,
					TopologySpreadConstraints: zync.Options.ZyncDatabaseTopologySpreadConstraints,
					Volumes: []v1.Volume{
						{
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
				{
					Name:       "8080-tcp",
					Protocol:   v1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt32(8080),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: ZyncName},
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
				{
					Name:       "postgresql",
					Protocol:   v1.ProtocolTCP,
					Port:       5432,
					TargetPort: intstr.FromInt32(5432),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "zync-database"},
		},
	}
}

func (zync *Zync) ZyncPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ZyncName,
			Labels: zync.Options.CommonZyncLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: ZyncName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (zync *Zync) QuePodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-que",
			Labels: zync.Options.CommonZyncQueLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: "zync-que"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (zync *Zync) zyncPorts() []v1.ContainerPort {
	ports := []v1.ContainerPort{
		{ContainerPort: 8080, Protocol: v1.ProtocolTCP},
	}

	if zync.Options.ZyncMetrics {
		ports = append(ports, v1.ContainerPort{Name: "metrics", ContainerPort: ZyncMetricsPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (zync *Zync) zyncInit(containerImage string) []v1.Container {
	if zync.Options.ZyncDbTLSEnabled {
		return []v1.Container{
			{
				Name:  "set-permissions",
				Image: containerImage,
				Command: []string{
					"sh",
					"-c",
					"cp /tls/* /writable-tls/ && chmod 0600 /writable-tls/*",
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "tls-secret",
						MountPath: "/tls",
						ReadOnly:  true,
					},
					{
						Name:      "writable-tls",
						MountPath: "/writable-tls",
						ReadOnly:  false, // Writable emptyDir volume
					},
				},
			},
			{
				Name:  ZyncInitContainerName,
				Image: containerImage,
				Command: []string{
					"bash",
					"-c",
					"bundle exec sh -c \"until rake boot:db; do sleep $SLEEP_SECONDS; done\"",
				},
				Env: []v1.EnvVar{
					{
						Name:  "SLEEP_SECONDS",
						Value: "1",
					},
					{
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
					helper.EnvVarFromSecretOptional("DATABASE_SSL_MODE", ZyncSecretName, "DATABASE_SSL_MODE"),
					helper.EnvVarFromValue("DATABASE_SSL_CA", helper.TlsCertPresent("DATABASE_SSL_CA", ZyncSecretName, zync.Options.ZyncDbTLSEnabled)),
					helper.EnvVarFromValue("DATABASE_SSL_CERT", helper.TlsCertPresent("DATABASE_SSL_CERT", ZyncSecretName, zync.Options.ZyncDbTLSEnabled)),
					helper.EnvVarFromValue("DATABASE_SSL_KEY", helper.TlsCertPresent("DATABASE_SSL_KEY", ZyncSecretName, zync.Options.ZyncDbTLSEnabled)),
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "writable-tls", // Reuse the same volume in the main container if needed
						MountPath: "/tls",
						ReadOnly:  true,
					},
				},
			},
		}
	} else {
		return []v1.Container{
			{
				Name:  ZyncInitContainerName,
				Image: containerImage,
				Command: []string{
					"bash",
					"-c",
					"bundle exec sh -c \"until rake boot:db; do sleep $SLEEP_SECONDS; done\"",
				},
				Env: []v1.EnvVar{
					{
						Name:  "SLEEP_SECONDS",
						Value: "1",
					},
					{
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
		}
	}
}

func (zync *Zync) zyncVolumeMount() []v1.VolumeMount {
	if zync.Options.ZyncDbTLSEnabled {
		return []v1.VolumeMount{
			{
				Name:      "writable-tls", // Reuse the same volume in the main container if needed
				MountPath: "/tls",
				ReadOnly:  true,
			},
		}
	} else {
		return []v1.VolumeMount{}
	}
}

func (zync *Zync) zyncVolume() []v1.Volume {
	if zync.Options.ZyncDbTLSEnabled {
		return []v1.Volume{
			{
				Name: "tls-secret",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: ZyncSecretName, // Name of the secret containing the TLS certs
						Items: []v1.KeyToPath{
							{
								Key:  ZyncSecretSslCa,
								Path: "ca.crt", // Map the secret key to the ca.crt file in the container
							},
							{
								Key:  ZyncSecretSslCert,
								Path: "tls.crt", // Map the secret key to the tls.crt file in the container
							},
							{
								Key:  ZyncSecretSslKey,
								Path: "tls.key", // Map the secret key to the tls.key file in the container
							},
						},
					},
				},
			},
			{
				Name: "writable-tls",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		}
	} else {
		return []v1.Volume{}
	}
}

func (zync *Zync) zyncQueInit() []v1.Container {
	if zync.Options.ZyncDbTLSEnabled {
		return []v1.Container{
			{
				Name:  "set-permissions",
				Image: "quay.io/openshift/origin-cli:4.7", // Minimal image for chmod
				Command: []string{
					"sh",
					"-c",
					"cp /tls/* /writable-tls/ && chmod 0600 /writable-tls/*",
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "tls-secret",
						MountPath: "/tls",
						ReadOnly:  true,
					},
					{
						Name:      "writable-tls",
						MountPath: "/writable-tls",
						ReadOnly:  false, // Writable emptyDir volume
					},
				},
			},
		}
	} else {
		return []v1.Container{}
	}
}

func (zync *Zync) zyncPodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := zync.Options.ZyncPodTemplateAnnotations

	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range zync.Options.ZyncPodTemplateAnnotations {
		annotations[key] = val
	}

	return annotations
}

func (zync *Zync) zyncQuePodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := zync.Options.ZyncQuePodTemplateAnnotations

	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range zync.Options.ZyncQuePodTemplateAnnotations {
		annotations[key] = val
	}

	return annotations
}
