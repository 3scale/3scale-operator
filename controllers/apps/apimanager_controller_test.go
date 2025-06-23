package controllers

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("APIManager controller", func() {
	var testNamespace string

	BeforeEach(func() {
		generatedTestNamespace := "test-namespace-" + uuid.New().String()
		// Add any setup steps that needs to be executed before each test
		desiredTestNamespace := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: generatedTestNamespace,
			},
		}

		err := testK8sClient.Create(context.Background(), desiredTestNamespace)
		Expect(err).ToNot(HaveOccurred())

		existingNamespace := &corev1.Namespace{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: generatedTestNamespace}, existingNamespace)
			return err == nil
		}, 5*time.Minute, 5*time.Second).Should(BeTrue())

		testNamespace = existingNamespace.Name
	})

	AfterEach(func() {
		desiredTestNamespace := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		// Add any teardown steps that needs to be executed after each test
		err := testK8sClient.Delete(context.Background(), desiredTestNamespace, client.PropagationPolicy(metav1.DeletePropagationForeground))

		Expect(err).ToNot(HaveOccurred())

		existingNamespace := &corev1.Namespace{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: testNamespace}, existingNamespace)
			if err != nil && errors.IsNotFound(err) {
				return false
			}
			return true
		}, 5*time.Minute, 5*time.Second).Should(BeTrue())
	})

	Context("Run directly without existing APIManager", func() {
		It("Should create successfully", func() {
			Expect(1).To(Equal(1))
		})
	})

	Context("Run APIManager standard deploy", func() {
		It("Should create successfully", func() {
			start := time.Now()
			os.Setenv("PREFLIGHT_CHECKS_BYPASS", "true")
			// Create dummy secret needed to deploy an APIManager
			// with S3 configuration for the E2E tests. As long as
			// S3-related functionality is exercised it should work correctly.
			dummyS3Secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy-s3-secret",
					Namespace: testNamespace,
				},
				StringData: map[string]string{
					apps.AwsAccessKeyID:     "dummyaccesskey",
					apps.AwsSecretAccessKey: "dummysecretaccesskey",
					apps.AwsBucket:          "dummybucket",
					apps.AwsRegion:          "dummyregion",
				},
			}

			err := testK8sClient.Create(context.Background(), dummyS3Secret)
			Expect(err).ToNot(HaveOccurred())

			// create mysql database
			err = createMysqlDatabase(testNamespace, testK8sClient)
			Expect(err).ToNot(HaveOccurred())

			// create system and backend redis
			err = createRedisDatabases(testNamespace, testK8sClient)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: dummyS3Secret.Name, Namespace: dummyS3Secret.Namespace}, dummyS3Secret)
				return err == nil
			}, 5*time.Minute, 5*time.Second).Should(BeTrue())

			// Create custom environment secret
			customEnvSecret := testGetCustomEnvironmentSecret(testNamespace)

			// Get the newly created custom environment secret for later
			err = testK8sClient.Create(context.Background(), customEnvSecret)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: customEnvSecret.Name, Namespace: customEnvSecret.Namespace}, customEnvSecret)
				return err == nil
			}, 5*time.Minute, 5*time.Second).Should(BeTrue())

			enableResourceRequirements := false
			wildcardDomain := "test1.127.0.0.1.nip.io"
			apimanager := &appsv1alpha1.APIManager{
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						WildcardDomain:              wildcardDomain,
						ResourceRequirementsEnabled: &enableResourceRequirements,
					},
					System: &appsv1alpha1.SystemSpec{
						FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{
							S3: &appsv1alpha1.SystemS3Spec{
								ConfigurationSecretRef: corev1.LocalObjectReference{
									Name: dummyS3Secret.Name,
								},
							},
						},
					},
					Apicast: &appsv1alpha1.ApicastSpec{
						StagingSpec: &appsv1alpha1.ApicastStagingSpec{
							CustomEnvironments: []appsv1alpha1.CustomEnvironmentSpec{
								{
									SecretRef: &corev1.LocalObjectReference{
										Name: customEnvSecret.Name,
									},
								},
							},
						},
						ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
							CustomEnvironments: []appsv1alpha1.CustomEnvironmentSpec{
								{
									SecretRef: &corev1.LocalObjectReference{
										Name: customEnvSecret.Name,
									},
								},
							},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example-apimanager",
					Namespace: testNamespace,
				},
			}

			err = testK8sClient.Create(context.Background(), apimanager)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: apimanager.Name, Namespace: apimanager.Namespace}, apimanager)
				return err == nil
			}, 5*time.Minute, 5*time.Second).Should(BeTrue())

			fmt.Fprintf(GinkgoWriter, "Waiting for all APIManager managed Deployments\n")
			err = waitForAllAPIManagerStandardDeployments(testNamespace, 5*time.Second, 15*time.Minute, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "All APIManager managed Deployments are ready\n")

			fmt.Fprintf(GinkgoWriter, "Waiting for all APIManager managed Routes\n")
			err = waitForAllAPIManagerStandardRoutes(5*time.Second, 15*time.Minute, wildcardDomain, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "All APIManager managed Routes are available\n")

			fmt.Fprintf(GinkgoWriter, "Waiting until APIManager CR has the correct secret UIDs\n")
			err = waitForAPIManagerLabels(testNamespace, 5*time.Second, 5*time.Minute, apimanager, customEnvSecret, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "APIManager CR has the correct secret UIDs\n")

			fmt.Fprintf(GinkgoWriter, "Waiting until hashed secret has been created and is accurate\n")
			err = waitForHashedSecret(testNamespace, 5*time.Second, 5*time.Minute, customEnvSecret, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "Hashed secret has been created and is accurate\n")

			fmt.Fprintf(GinkgoWriter, "Waiting until apicast pod annotations have been verified\n")
			err = waitForApicastPodAnnotations(testNamespace, 5*time.Second, 5*time.Minute, customEnvSecret, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "Apicast pod annotations have been verified\n")

			// TODO: Add code checking annotations on apicast pods

			fmt.Fprintf(GinkgoWriter, "Waiting until APIManager's 'Available' condition is true\n")
			err = waitForAPIManagerAvailableCondition(5*time.Second, 15*time.Minute, apimanager, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "APIManager 'Available' condition is true\n")

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "APIManager creation and availability took '%s'\n", elapsed)
		})
	})
})

func createRedisDatabases(namespace string, k8sclient client.Client) error {
	// Create backend-redis secret
	err := createBackendRedisSecret(namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create backend redis pvc
	err = createPVC("backend-redis-pvc", namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create backend redis service
	err = createService("backend-redis", "backend-redis", namespace, 6379, k8sclient)
	if err != nil {
		return err
	}

	// Create backend redis deployment
	err = createRedisDeployment("backend-redis", "backend-redis-pvc", namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create system-redis secret
	err = createSystemRedisSecret(namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create system-redis pvc
	err = createPVC("system-redis-pvc", namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create system-redis service
	err = createService("system-redis", "system-redis", namespace, 6379, k8sclient)
	if err != nil {
		return err
	}

	// Create system-redis deployment
	err = createRedisDeployment("system-redis", "system-redis-pvc", namespace, k8sclient)
	if err != nil {
		return err
	}

	return nil
}

func createRedisDeployment(deploymentName, pvcName, namespace string, k8sclient client.Client) error {
	// Define the Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": deploymentName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "redis",
							Image: "quay.io/fedora/redis-7",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6379,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("250m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("128Mi"),
									corev1.ResourceCPU:    resource.MustParse("500m"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 6379,
										},
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt(6379),
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "redis-storage",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "redis-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}

	// Check if the Deployment already exists
	existingDeployment := &appsv1.Deployment{}
	err := k8sclient.Get(context.TODO(), client.ObjectKey{
		Name:      deploymentName,
		Namespace: namespace,
	}, existingDeployment)

	if err == nil {
		// Deployment already exists
		fmt.Println("Deployment already exists. Skipping creation.")
		return nil
	}

	// Create the Deployment
	if err := k8sclient.Create(context.TODO(), deployment); err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	fmt.Println("Deployment created successfully.")
	return nil
}

func createSystemRedisSecret(namespace string, k8sclient client.Client) error {
	// Define the Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-redis",
			Namespace: namespace,
		},
		StringData: map[string]string{
			"SENTINEL_HOSTS": "",
			"SENTINEL_ROLE":  "",
			"URL":            fmt.Sprintf("redis://system-redis.%s.svc.cluster.local:6379/1", namespace),
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Check if the Secret already exists
	existingSecret := &corev1.Secret{}
	err := k8sclient.Get(context.TODO(), client.ObjectKey{
		Name:      "system-redis",
		Namespace: namespace,
	}, existingSecret)

	if err == nil {
		// Secret already exists
		fmt.Println("Secret already exists. Skipping creation.")
		return nil
	}

	// Create the Secret
	if err := k8sclient.Create(context.TODO(), secret); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	fmt.Println("Secret created successfully.")
	return nil
}

func createBackendRedisSecret(namespace string, k8sclient client.Client) error {
	// Define the Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-redis",
			Namespace: namespace,
		},
		StringData: map[string]string{
			"REDIS_QUEUES_SENTINEL_HOSTS":  "",
			"REDIS_QUEUES_SENTINEL_ROLE":   "",
			"REDIS_QUEUES_URL":             fmt.Sprintf("redis://backend-redis.%s.svc.cluster.local:6379/1", namespace),
			"REDIS_STORAGE_SENTINEL_HOSTS": "",
			"REDIS_STORAGE_SENTINEL_ROLE":  "",
			"REDIS_STORAGE_URL":            fmt.Sprintf("redis://backend-redis.%s.svc.cluster.local:6379/2", namespace),
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Check if the Secret already exists
	existingSecret := &corev1.Secret{}
	err := k8sclient.Get(context.TODO(), client.ObjectKey{
		Name:      "backend-redis",
		Namespace: namespace,
	}, existingSecret)

	if err == nil {
		// Secret already exists
		fmt.Println("Secret already exists. Skipping creation.")
		return nil
	}

	// Create the Secret
	if err := k8sclient.Create(context.TODO(), secret); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	fmt.Println("Secret created successfully.")
	return nil
}

func createMysqlDatabase(namespace string, k8sclient client.Client) error {
	// Create secret
	err := createSystemDatabaseSecret(namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create PVC
	err = createPVC("system-mysql-pvc", namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create service
	err = createService("system-mysql", "mysql", namespace, 3306, k8sclient)
	if err != nil {
		return err
	}

	// Create main configuration config map
	err = createMainConfigCM(namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create extra configuration config map
	err = createConfigurationCM(namespace, k8sclient)
	if err != nil {
		return err
	}

	// Create deployment
	err = createMySQLDeployment(namespace, k8sclient)
	if err != nil {
		return err
	}

	return nil
}

func createMainConfigCM(namespace string, k8sclient client.Client) error {
	// Define the ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-main-conf",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "mysql",
			},
		},
		Data: map[string]string{
			"my.cnf": "!include /etc/my.cnf\n!includedir /etc/my-extra.d\n",
		},
	}

	// Check if the configmap already exists
	existingCM := &corev1.ConfigMap{}
	err := k8sclient.Get(context.TODO(), client.ObjectKey{
		Name:      "mysql-main-conf",
		Namespace: namespace,
	}, existingCM)

	if err == nil {
		// CM already exists
		fmt.Println("ConfigMap already exists. Skipping creation.")
		return nil
	}

	// Create the configMap
	if err := k8sclient.Create(context.TODO(), configMap); err != nil {
		return fmt.Errorf("failed to create configMap: %w", err)
	}

	fmt.Println("configMap created successfully.")

	return nil
}

func createConfigurationCM(namespace string, k8sclient client.Client) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-extra-conf",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "mysql",
			},
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

	// Check if the configmap already exists
	existingCM := &corev1.ConfigMap{}
	err := k8sclient.Get(context.TODO(), client.ObjectKey{
		Name:      "mysql-extra-conf",
		Namespace: namespace,
	}, existingCM)

	if err == nil {
		// CM already exists
		fmt.Println("ConfigMap already exists. Skipping creation.")
		return nil
	}

	// Create the configMap
	if err := k8sclient.Create(context.TODO(), configMap); err != nil {
		return fmt.Errorf("failed to create configMap: %w", err)
	}

	fmt.Println("configMap created successfully.")

	return nil
}

func createMySQLDeployment(namespace string, k8sclient client.Client) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-mysql",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "mysql",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mysql",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "mysql-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "system-mysql-pvc",
									ReadOnly:  false,
								},
							},
						},
						{
							Name: "mysql-extra-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "mysql-extra-conf",
									},
								},
							},
						},
						{
							Name: "mysql-main-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "mysql-main-conf",
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "system-mysql",
							Image: "quay.io/sclorg/mysql-80-c8s",
							Ports: []corev1.ContainerPort{
								{
									HostPort:      0,
									ContainerPort: 3306,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								helper.EnvVarFromSecret("MYSQL_USER", "system-database", "DB_USER"),
								helper.EnvVarFromSecret("MYSQL_PASSWORD", "system-database", "DB_PASSWORD"),
								helper.EnvVarFromValue("MYSQL_DATABASE", "dev"),
								helper.EnvVarFromSecret("MYSQL_ROOT_PASSWORD", "system-database", "DB_ROOT_PASSWORD"),
								helper.EnvVarFromValue("MYSQL_LOWER_CASE_TABLE_NAMES", "1"),
								helper.EnvVarFromValue("MYSQL_DEFAULTS_FILE", "/etc/my-extra/my.cnf"),
							},
							VolumeMounts: []corev1.VolumeMount{
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
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
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
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-i", "-c", "MYSQL_PWD=\"$MYSQL_PASSWORD\" mysql -h 127.0.0.1 -u $MYSQL_USER -D $MYSQL_DATABASE -e 'SELECT 1'"},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}

	// Check if the Deployment already exists
	existingDeployment := &appsv1.Deployment{}
	err := k8sclient.Get(context.TODO(), client.ObjectKey{
		Name:      "system-mysql",
		Namespace: namespace,
	}, existingDeployment)

	if err == nil {
		// Deployment already exists
		fmt.Println("Deployment already exists. Skipping creation.")
		return nil
	}

	// Create the Deployment
	if err := k8sclient.Create(context.TODO(), deployment); err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	fmt.Println("Deployment created successfully.")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func createService(name, label, namespace string, port int32, k8sclient client.Client) error {
	// Define the Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": label,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt(int(port)),
				},
			},
			Selector: map[string]string{
				"app": label,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Check if the Service already exists
	existingService := &corev1.Service{}
	err := k8sclient.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, existingService)

	if err == nil {
		fmt.Println("Service already exists. Skipping creation.")
		return nil
	}

	// Create the Service
	if err := k8sclient.Create(context.TODO(), service); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	fmt.Println("Service created successfully.")
	return nil
}

func createSystemDatabaseSecret(namespace string, k8sclient client.Client) error {
	// Define the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-database",
			Namespace: namespace,
		},
		StringData: map[string]string{
			"DB_USER":          "mysql",
			"DB_PASSWORD":      "password",
			"DB_ROOT_PASSWORD": "rootpassword",
			"URL":              fmt.Sprintf("mysql2://root:rootpassword@system-mysql.%s.svc.cluster.local/dev", namespace),
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Check if the secret already exists
	existingSecret := &corev1.Secret{}
	err := k8sclient.Get(context.TODO(), types.NamespacedName{
		Name:      "system-database",
		Namespace: namespace,
	}, existingSecret)

	if err == nil {
		fmt.Println("Secret already exists. Skipping creation.")
		return nil
	}

	// Create the secret
	if err := k8sclient.Create(context.TODO(), secret); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	fmt.Println("Secret created successfully.")
	return nil
}

func createPVC(name, namespace string, k8sclient client.Client) error {
	// Define the PVC
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	// Check if the PVC already exists
	existingPVC := &corev1.PersistentVolumeClaim{}
	err := k8sclient.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, existingPVC)

	if err == nil {
		fmt.Println("PVC already exists. Skipping creation.")
		return nil
	}

	// Create the PVC
	if err := k8sclient.Create(context.TODO(), pvc); err != nil {
		return fmt.Errorf("failed to create PVC: %w", err)
	}

	fmt.Println("PVC created successfully.")
	return nil
}

func waitForAllAPIManagerStandardDeployments(namespace string, retryInterval, timeout time.Duration, w io.Writer) error {
	deploymentNames := []string{ // TODO gather this from constants/somewhere centralized
		"apicast-production",
		"apicast-staging",
		"backend-cron",
		"backend-listener",
		"backend-redis",
		"backend-worker",
		"system-app",
		"system-memcache",
		"system-mysql",
		"system-redis",
		"system-sidekiq",
		"system-searchd",
		"zync",
		"zync-que",
		"zync-database",
	}

	for _, dName := range deploymentNames {
		lookupKey := types.NamespacedName{Name: dName, Namespace: namespace}
		createdDeployment := &appsv1.Deployment{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), lookupKey, createdDeployment)
			if err != nil {
				return false
			}

			if helper.IsDeploymentAvailable(createdDeployment) {
				fmt.Fprintf(w, "Deployment '%s' available\n", dName)
				return true
			}

			availableReplicas := createdDeployment.Status.AvailableReplicas
			desiredReplicas := createdDeployment.Spec.Replicas
			fmt.Fprintf(w, "Waiting for full availability of %s Deployment (%d/%d)\n", dName, availableReplicas, desiredReplicas)
			return false
		}, timeout, retryInterval).Should(BeTrue())
	}

	return nil
}

func waitForAllAPIManagerStandardRoutes(retryInterval, timeout time.Duration, wildcardDomain string, w io.Writer) error {
	routeHosts := []string{
		"backend-3scale." + wildcardDomain,                // Backend Listener route
		"api-3scale-apicast-production." + wildcardDomain, // Apicast Production '3scale' tenant Route
		"api-3scale-apicast-staging." + wildcardDomain,    // Apicast Staging '3scale' tenant Route
		"master." + wildcardDomain,                        // System's Master Portal Route
		"3scale." + wildcardDomain,                        // System's '3scale' tenant Developer Portal Route
		"3scale-admin." + wildcardDomain,                  // System's '3scale' tenant Admin Portal Route
	}
	for _, routeHost := range routeHosts {
		Eventually(func() bool {
			routeList := &routev1.RouteList{}
			routeListOptions := client.ListOptions{
				FieldSelector: fields.OneTermEqualSelector("spec.host", routeHost),
			}
			err := testK8sAPIClient.List(context.Background(), routeList, &routeListOptions)
			if err != nil {
				if errors.IsNotFound(err) {
					fmt.Fprintf(w, "Waiting for availability of Route with host '%s'\n", routeHost)
					return false
				}
				fmt.Fprintf(w, "Error Listing Routes with host '%s': %s\n", routeHost, err)
				return false
			}

			routeItems := routeList.Items
			if len(routeItems) == 0 {
				fmt.Fprintf(w, "Waiting for availability of Route with host '%s'\n", routeHost)
				return false
			}
			if len(routeItems) > 1 {
				fmt.Fprintf(w, "Found unexpected routes with duplicated 'host' fields\n")
				return false
			}

			route := routeItems[0]
			if !helper.IsRouteReady(&route) {
				return false
			}

			fmt.Fprintf(w, "Route '%s' with host '%s' ready\n", route.Name, route.Spec.Host)
			return true
		}, timeout, retryInterval).Should(BeTrue())
	}

	return nil
}

func waitForAPIManagerAvailableCondition(retryInterval, timeout time.Duration, apimanager *appsv1alpha1.APIManager, w io.Writer) error {
	Eventually(func() bool {
		err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: apimanager.Name, Namespace: apimanager.Namespace}, apimanager)
		if err != nil {
			fmt.Fprintf(w, "Error getting APIManager '%s': %v\n", apimanager.Name, err)
			return false
		}

		return apimanager.Status.Conditions.IsTrueFor(appsv1alpha1.APIManagerAvailableConditionType)
	}, timeout, retryInterval).Should(BeTrue())

	return nil
}

func waitForAPIManagerLabels(namespace string, retryInterval time.Duration, timeout time.Duration, apimanager *appsv1alpha1.APIManager, customEnvSecret *corev1.Secret, w io.Writer) error {
	Eventually(func() bool {
		reconciledApimanager := &appsv1alpha1.APIManager{}
		reconciledApimanagerKey := types.NamespacedName{Name: apimanager.Name, Namespace: namespace}
		err := testK8sClient.Get(context.Background(), reconciledApimanagerKey, reconciledApimanager)
		if err != nil {
			fmt.Fprintf(w, "Error getting APIManager '%s': %v\n", apimanager.Name, err)
			return false
		}

		expectedLabels := map[string]string{
			fmt.Sprintf("%s%s", APImanagerSecretLabelPrefix, string(customEnvSecret.GetUID())): "true",
		}

		// Then verify that the hash matches the hashed config secret
		return reflect.DeepEqual(reconciledApimanager.Labels, expectedLabels)
	}, timeout, retryInterval).Should(BeTrue())

	return nil
}

func waitForHashedSecret(namespace string, retryInterval time.Duration, timeout time.Duration, customEnvSecret *corev1.Secret, w io.Writer) error {
	Eventually(func() bool {
		// First get the master hashed secret
		hashedSecret := &corev1.Secret{}
		hashedSecretLookupKey := types.NamespacedName{Name: component.HashedSecretName, Namespace: namespace}
		err := testK8sClient.Get(context.Background(), hashedSecretLookupKey, hashedSecret)
		if err != nil {
			fmt.Fprintf(w, "Error getting hashed secret '%s': %v\n", hashedSecretLookupKey.Name, err)
			return false
		}

		// Then verify that the hash matches the hashed custom environment secret
		return helper.GetSecretStringDataFromData(hashedSecret.Data)[customEnvSecret.Name] == component.HashSecret(customEnvSecret.Data)
	}, timeout, retryInterval).Should(BeTrue())

	return nil
}

func waitForApicastPodAnnotations(namespace string, retryInterval time.Duration, timeout time.Duration, customEnvSecret *corev1.Secret, w io.Writer) error {
	apicastDeploymentNames := []string{
		"apicast-production",
		"apicast-staging",
	}

	for _, dName := range apicastDeploymentNames {
		apicastDeploymentLookupKey := types.NamespacedName{Name: dName, Namespace: namespace}
		apicastDeployment := &appsv1.Deployment{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), apicastDeploymentLookupKey, apicastDeployment)
			if err != nil {
				return false
			}

			for aKey, aValue := range apicastDeployment.Spec.Template.Annotations {
				if aKey == fmt.Sprintf("%s%s", component.CustomEnvSecretResverAnnotationPrefix, customEnvSecret.Name) {
					if aValue == customEnvSecret.ResourceVersion {
						fmt.Fprintf(w, "Deployment '%s' has the custom env secret annotation and correct resourceVersion\n", dName)
						return true
					}
					fmt.Fprintf(w, "Deployment '%s' has the custom env secret annotation but the resourceVersion '%s' doesn't match the expected value '%s'\n", dName, aValue, customEnvSecret.ResourceVersion)
					return false
				}
			}
			fmt.Fprintf(w, "Deployment '%s' doesn't have the custom env secret annotation\n", dName)
			return false
		}, timeout, retryInterval).Should(BeTrue())
	}

	return nil
}

func testCustomEnvironmentContent() string {
	return `
    local cjson = require('cjson')
    local PolicyChain = require('apicast.policy_chain')
    local policy_chain = context.policy_chain
    
    local logging_policy_config = cjson.decode([[
    {
      "enable_access_logs": false,
      "custom_logging": "\"{{request}}\" to service {{service.name}} and {{service.id}}"
    }
    ]])
    
    policy_chain:insert( PolicyChain.load_policy('logging', 'builtin', logging_policy_config), 1)
    
    return {
      policy_chain = policy_chain,
      port = { metrics = 9421 },
    }
`
}

func testGetCustomEnvironmentSecret(namespace string) *corev1.Secret {
	customEnvironmentSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-env-1",
			Namespace: namespace,
			Labels: map[string]string{
				"apimanager.apps.3scale.net/watched-by": "apimanager",
			},
		},
		StringData: map[string]string{
			"custom_env.lua":  testCustomEnvironmentContent(),
			"custom_env2.lua": testCustomEnvironmentContent(),
		},
	}
	return &customEnvironmentSecret
}
