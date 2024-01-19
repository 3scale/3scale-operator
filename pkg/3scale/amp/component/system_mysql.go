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
	SystemMySQLDeploymentName = "system-mysql"
)

type SystemMysql struct {
	Options *SystemMysqlOptions
}

func NewSystemMysql(options *SystemMysqlOptions) *SystemMysql {
	return &SystemMysql{Options: options}
}

func (mysql *SystemMysql) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-mysql",
			Labels: mysql.Options.DeploymentLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "system-mysql",
					Protocol:   v1.ProtocolTCP,
					Port:       3306,
					TargetPort: intstr.FromInt(3306),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "system-mysql"},
		},
	}
}

func (mysql *SystemMysql) MainConfigConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-main-conf",
			Labels: mysql.Options.DeploymentLabels,
		},
		Data: map[string]string{"my.cnf": "!include /etc/my.cnf\n!includedir /etc/my-extra.d\n"}}
}

func (mysql *SystemMysql) ExtraConfigConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-extra-conf",
			Labels: mysql.Options.DeploymentLabels,
		},
		Data: map[string]string{
			"mysql-charset.cnf": `[client]
default-character-set = utf8

[mysql]
default-character-set = utf8

[mysqld]
character-set-server = utf8
collation-server = utf8_unicode_ci
`,
			"mysql-default-authentication-plugin.cnf": `[mysqld]
default_authentication_plugin=mysql_native_password
`,
		},
	}
}

func (mysql *SystemMysql) PersistentVolumeClaim() *v1.PersistentVolumeClaim {
	volName := ""
	if mysql.Options.PVCVolumeName != nil {
		volName = *mysql.Options.PVCVolumeName
	}

	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-storage",
			Labels: mysql.Options.DeploymentLabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: mysql.Options.PVCStorageRequests,
				},
			},
			StorageClassName: mysql.Options.PVCStorageClass,
			VolumeName:       volName,
		},
	}
}

func (mysql *SystemMysql) Deployment(containerImage string) *k8sappsv1.Deployment {
	var mysqlReplicas int32 = 1

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemMySQLDeploymentName,
			Labels: mysql.Options.DeploymentLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &mysqlReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemMySQLDeploymentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      mysql.Options.PodTemplateLabels,
					Annotations: mysql.Options.PodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:           mysql.Options.Affinity,
					Tolerations:        mysql.Options.Tolerations,
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Volumes: []v1.Volume{
						{
							Name: "mysql-storage",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mysql-storage",
									ReadOnly:  false,
								},
							},
						},
						{
							Name: "mysql-extra-conf",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "mysql-extra-conf",
									},
								},
							},
						},
						{
							Name: "mysql-main-conf",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "mysql-main-conf",
									},
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "system-mysql",
							Image: containerImage,
							Ports: []v1.ContainerPort{
								{
									HostPort:      0,
									ContainerPort: 3306,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Env: []v1.EnvVar{
								helper.EnvVarFromSecret("MYSQL_USER", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseUserFieldName),
								helper.EnvVarFromSecret("MYSQL_PASSWORD", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabasePasswordFieldName),
								// TODO This should be gathered from secrets but we cannot set them because the URL field of the system-database secret
								// is already formed from this contents and we would have duplicate information. Once OpenShift templates
								// are deprecated we should be able to change this.
								helper.EnvVarFromValue("MYSQL_DATABASE", mysql.Options.DatabaseName),
								helper.EnvVarFromValue("MYSQL_ROOT_PASSWORD", mysql.Options.RootPassword),
								helper.EnvVarFromValue("MYSQL_LOWER_CASE_TABLE_NAMES", "1"),
								helper.EnvVarFromValue("MYSQL_DEFAULTS_FILE", "/etc/my-extra/my.cnf"),
							},
							Resources: mysql.Options.ContainerResourceRequirements,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "mysql-storage",
									ReadOnly:  false,
									MountPath: "/var/lib/mysql/data",
								},
								{
									Name:      "mysql-extra-conf",
									ReadOnly:  false,
									MountPath: "/etc/my-extra.d",
								},
								{
									Name:      "mysql-main-conf",
									ReadOnly:  false,
									MountPath: "/etc/my-extra",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 3306,
										},
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "MYSQL_PWD=\"$MYSQL_PASSWORD\" mysql -h 127.0.0.1 -u $MYSQL_USER -D $MYSQL_DATABASE -e 'SELECT 1'"},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					PriorityClassName:         mysql.Options.PriorityClassName,
					TopologySpreadConstraints: mysql.Options.TopologySpreadConstraints,
				},
			},
		},
	}
}

// Each database is responsible to create the needed secrets for the other components
func (mysql *SystemMysql) SystemDatabaseSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemDatabaseSecretName,
			Labels: mysql.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseUserFieldName:     mysql.Options.User,
			SystemSecretSystemDatabasePasswordFieldName: mysql.Options.Password,
			SystemSecretSystemDatabaseURLFieldName:      mysql.Options.DatabaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}

////// End System Mysql
