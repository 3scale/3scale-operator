package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type SystemMysql struct {
	Options *SystemMysqlOptions
}

func NewSystemMysql(options *SystemMysqlOptions) *SystemMysql {
	return &SystemMysql{Options: options}
}

func (mysql *SystemMysql) Objects() []common.KubernetesObject {
	deploymentConfig := mysql.DeploymentConfig()
	service := mysql.Service()
	mainConfigConfigMap := mysql.MainConfigConfigMap()
	extraConfigconfigMap := mysql.ExtraConfigConfigMap()
	persistentVolumeClaim := mysql.PersistentVolumeClaim()
	systemDatabaseSecret := mysql.SystemDatabaseSecret()

	objects := []common.KubernetesObject{
		deploymentConfig,
		service,
		mainConfigConfigMap,
		extraConfigconfigMap,
		persistentVolumeClaim,
		systemDatabaseSecret,
	}

	return objects
}

func (mysql *SystemMysql) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-mysql",
			Labels: map[string]string{
				"app":                          mysql.Options.AppLabel,
				"threescale_component":         "system",
				"threescale_component_element": "mysql",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "system-mysql",
					Protocol:   v1.ProtocolTCP,
					Port:       3306,
					TargetPort: intstr.FromInt(3306),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-mysql"},
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
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.AppLabel},
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
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.AppLabel},
		},
		Data: map[string]string{"mysql-charset.cnf": "[client]\ndefault-character-set = utf8\n\n[mysql]\ndefault-character-set = utf8\n\n[mysqld]\ncharacter-set-server = utf8\ncollation-server = utf8_unicode_ci\n"}}
}

func (mysql *SystemMysql) PersistentVolumeClaim() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-storage",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.AppLabel},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.ResourceRequirements{Requests: v1.ResourceList{"storage": resource.MustParse("1Gi")}}}}
}

func (mysql *SystemMysql) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-mysql",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.AppLabel},
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
							"system-mysql",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "system-mysql:latest",
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-mysql"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.AppLabel, "deploymentConfig": "system-mysql"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "mysql-storage",
							VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "mysql-storage",
								ReadOnly:  false}},
						}, v1.Volume{
							Name: "mysql-extra-conf",
							VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "mysql-extra-conf"}}},
						}, v1.Volume{
							Name: "mysql-main-conf",
							VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "mysql-main-conf"}}}},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "system-mysql",
							Image: "system-mysql:latest",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{HostPort: 0,
									ContainerPort: 3306,
									Protocol:      v1.ProtocolTCP},
							},
							Env: []v1.EnvVar{
								envVarFromSecret("MYSQL_USER", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseUserFieldName),
								envVarFromSecret("MYSQL_PASSWORD", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabasePasswordFieldName),
								// TODO This should be gathered from secrets but we cannot set them because the URL field of the system-database secret
								// is already formed from this contents and we would have duplicate information. Once OpenShift templates
								// are deprecated we should be able to change this.
								envVarFromValue("MYSQL_DATABASE", mysql.Options.DatabaseName),
								envVarFromValue("MYSQL_ROOT_PASSWORD", mysql.Options.RootPassword),
								envVarFromValue("MYSQL_LOWER_CASE_TABLE_NAMES", "1"),
								envVarFromValue("MYSQL_DEFAULTS_FILE", "/etc/my-extra/my.cnf"),
							},
							Resources: mysql.Options.ContainerResourceRequirements,
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "mysql-storage",
									ReadOnly:  false,
									MountPath: "/var/lib/mysql/data",
								}, v1.VolumeMount{
									Name:      "mysql-extra-conf",
									ReadOnly:  false,
									MountPath: "/etc/my-extra.d",
								}, v1.VolumeMount{
									Name:      "mysql-main-conf",
									ReadOnly:  false,
									MountPath: "/etc/my-extra"},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.Int),
										IntVal: 3306}},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "MYSQL_PWD=\"$MYSQL_PASSWORD\" mysql -h 127.0.0.1 -u $MYSQL_USER -D $MYSQL_DATABASE -e 'SELECT 1'"}},
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
			Name: SystemSecretSystemDatabaseSecretName,
			Labels: map[string]string{
				"app":                  mysql.Options.AppLabel,
				"threescale_component": "system",
			},
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
