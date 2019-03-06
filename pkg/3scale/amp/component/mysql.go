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

type Mysql struct {
	// TemplateParameters
	// TemplateObjects
	// CLI Flags??? should be in this object???
	options []string
	Options *MysqlOptions
}

type MysqlOptions struct {
	mysqlNonRequiredOptions
	mysqlRequiredOptions
}

type mysqlRequiredOptions struct {
	appLabel     string
	databaseName string
	image        string
	user         string
	password     string
	rootPassword string
	databaseURL  string
}

type mysqlNonRequiredOptions struct {
}

type MysqlOptionsProvider interface {
	GetMysqlOptions() *MysqlOptions
}
type CLIMysqlOptionsProvider struct {
}

func (o *CLIMysqlOptionsProvider) GetMysqlOptions() (*MysqlOptions, error) {
	mob := MysqlOptionsBuilder{}
	mob.AppLabel("${APP_LABEL}")
	mob.DatabaseName("${MYSQL_DATABASE}")
	mob.Image("${MYSQL_IMAGE}")
	mob.User("${MYSQL_USER}")
	mob.Password("${MYSQL_PASSWORD}")
	mob.RootPassword("${MYSQL_ROOT_PASSWORD}")
	mob.DatabaseURL("mysql2://root:" + "${MYSQL_ROOT_PASSWORD}" + "@system-mysql/" + "${MYSQL_DATABASE}")
	res, err := mob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create MySQL Options - %s", err)
	}
	return res, nil
}

func NewMysql(options []string) *Mysql {
	redis := &Mysql{
		options: options,
	}
	return redis
}

func (mysql *Mysql) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIMysqlOptionsProvider{}
	mysqlOpts, err := optionsProvider.GetMysqlOptions()
	_ = err
	mysql.Options = mysqlOpts
	mysql.buildParameters(template)
	mysql.addObjectsIntoTemplate(template)
}

func (mysql *Mysql) GetObjects() ([]runtime.RawExtension, error) {
	objects := mysql.buildObjects()
	return objects, nil
}

func (mysql *Mysql) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := mysql.buildObjects()
	template.Objects = append(template.Objects, objects...)
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

func (mysql *Mysql) buildObjects() []runtime.RawExtension {
	systemMysqlDeploymentConfig := mysql.buildSystemMysqlDeploymentConfig()
	systemMysqlService := mysql.buildSystemMysqlService()
	systemMysqlMainConfigConfigMap := mysql.buildSystemMysqlMainConfigConfigMap()
	systemMysqlExtraConfigConfigMap := mysql.buildSystemMysqlExtraConfigConfigMap()
	systemMysqlPersistentVolumeClaim := mysql.buildSystemMysqlPersistentVolumeClaim()
	systemDatabaseSecret := mysql.buildSystemDatabaseSecrets()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: systemMysqlDeploymentConfig},
		runtime.RawExtension{Object: systemMysqlService},
		runtime.RawExtension{Object: systemMysqlMainConfigConfigMap},
		runtime.RawExtension{Object: systemMysqlExtraConfigConfigMap},
		runtime.RawExtension{Object: systemMysqlPersistentVolumeClaim},
		runtime.RawExtension{Object: systemDatabaseSecret},
	}

	return objects
}

func (mysql *Mysql) buildSystemMysqlService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-mysql",
			Labels: map[string]string{
				"app":                          mysql.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "mysql",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "system-mysql",
					Protocol:   v1.Protocol("TCP"),
					Port:       3306,
					TargetPort: intstr.FromInt(3306),
					NodePort:   0,
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-mysql"},
		},
	}
}

func (mysql *Mysql) buildSystemMysqlMainConfigConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-main-conf",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.appLabel},
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
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.appLabel},
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
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.appLabel},
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
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.appLabel},
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
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "mysql", "app": mysql.Options.appLabel, "deploymentConfig": "system-mysql"},
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
							Image: mysql.Options.image,
							Ports: []v1.ContainerPort{
								v1.ContainerPort{HostPort: 0,
									ContainerPort: 3306,
									Protocol:      v1.Protocol("TCP")},
							},
							Env: []v1.EnvVar{
								envVarFromSecret("MYSQL_USER", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseUserFieldName),
								envVarFromSecret("MYSQL_PASSWORD", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabasePasswordFieldName),
								// TODO This should be gathered from secrets but we cannot set them because the URL field of the system-database secret
								// is already formed from this contents and we would have duplicate information. Once OpenShift templates
								// are deprecated we should be able to change this.
								envVarFromValue("MYSQL_DATABASE", mysql.Options.databaseName),
								envVarFromValue("MYSQL_ROOT_PASSWORD", mysql.Options.rootPassword),
								envVarFromValue("MYSQL_LOWER_CASE_TABLE_NAMES", "1"),
								envVarFromValue("MYSQL_DEFAULTS_FILE", "/etc/my-extra/my.cnf"),
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

// Each database is responsible to create the needed secrets for the other components
func (mysql *Mysql) buildSystemDatabaseSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemDatabaseSecretName,
			Labels: map[string]string{
				"app":                  mysql.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseUserFieldName:     mysql.Options.user,
			SystemSecretSystemDatabasePasswordFieldName: mysql.Options.password,
			SystemSecretSystemDatabaseURLFieldName:      mysql.Options.databaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}

////// End System Mysql
