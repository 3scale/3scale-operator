package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Mysql struct {
	// TemplateParameters
	// TemplateObjects
	// CLI Flags??? should be in this object???
	options []string
}

func NewMysql(options []string) *Mysql {
	redis := &Mysql{
		options: options,
	}
	return redis
}

func (mysql *Mysql) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	mysql.buildParameters(template)
	mysql.buildObjects(template)
}

func (mysql *Mysql) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (mysql *Mysql) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "MYSQL_USER",
			DisplayName: "MySQL User",
			Description: "Username for MySQL user that will be used for accessing the database.",
			Value:       "mysql",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MYSQL_PASSWORD",
			DisplayName: "MySQL Password",
			Description: "Password for the MySQL user.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MYSQL_DATABASE",
			DisplayName: "MySQL Database Name",
			Description: "Name of the MySQL database accessed.",
			Value:       "system",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MYSQL_ROOT_PASSWORD",
			DisplayName: "MySQL Root password.",
			Description: "Password for Root user.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

func (mysql *Mysql) buildObjects(template *templatev1.Template) {
	systemMysqlDeploymentConfig := mysql.buildSystemMysqlDeploymentConfig()
	systemMysqlMainConfigConfigMap := mysql.buildSystemMysqlMainConfigConfigMap()
	systemMysqlExtraConfigConfigMap := mysql.buildSystemMysqlExtraConfigConfigMap()
	systemMysqlPersistentVolumeClaim := mysql.buildSystemMysqlPersistentVolumeClaim()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: systemMysqlDeploymentConfig},
		runtime.RawExtension{Object: systemMysqlMainConfigConfigMap},
		runtime.RawExtension{Object: systemMysqlExtraConfigConfigMap},
		runtime.RawExtension{Object: systemMysqlPersistentVolumeClaim},
	}
	template.Objects = append(template.Objects, objects...)
}

func (mysql *Mysql) buildSystemMysqlMainConfigConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-main-conf",
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "mysql", "app": "${APP_LABEL}"},
		},
		Data: map[string]string{"my.cnf": "!include /etc/my.cnf\n!includedir /etc/my-extra.d\n"}}
}

func (mysql *Mysql) buildSystemMysqlExtraConfigConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-extra-conf",
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "mysql", "app": "${APP_LABEL}"},
		},
		Data: map[string]string{"mysql-charset.cnf": "[client]\ndefault-character-set = utf8\n\n[mysql]\ndefault-character-set = utf8\n\n[mysqld]\ncharacter-set-server = utf8\ncollation-server = utf8_unicode_ci\n"}}
}

func (mysql *Mysql) buildSystemMysqlPersistentVolumeClaim() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-storage",
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "mysql", "app": "${APP_LABEL}"},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.ResourceRequirements{Requests: v1.ResourceList{"storage": resource.MustParse("1Gi")}}}}
}

func (mysql *Mysql) buildSystemMysqlDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-mysql",
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "mysql", "app": "${APP_LABEL}"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyType("Recreate"),
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ConfigChange")},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-mysql"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "mysql", "app": "${APP_LABEL}", "deploymentConfig": "system-mysql"},
				},
				Spec: v1.PodSpec{
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
							Image: "${MYSQL_IMAGE}",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{HostPort: 0,
									ContainerPort: 3306,
									Protocol:      v1.Protocol("TCP")},
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "MYSQL_USER",
									Value: "${MYSQL_USER}",
								}, v1.EnvVar{
									Name:  "MYSQL_PASSWORD",
									Value: "${MYSQL_PASSWORD}",
								}, v1.EnvVar{
									Name:  "MYSQL_DATABASE",
									Value: "${MYSQL_DATABASE}",
								}, v1.EnvVar{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: "${MYSQL_ROOT_PASSWORD}",
								}, v1.EnvVar{
									Name:  "MYSQL_LOWER_CASE_TABLE_NAMES",
									Value: "1",
								}, v1.EnvVar{
									Name:  "MYSQL_DEFAULTS_FILE",
									Value: "/etc/my-extra/my.cnf"},
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
										Type:   intstr.Type(0),
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
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
						},
					},
				},
			},
		},
	}
}

////// End System Mysql
