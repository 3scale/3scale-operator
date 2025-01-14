package controllers

import (
	"context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("APIManager controller", func() {
	var testNamespace string

	BeforeEach(func() {
		var generatedTestNamespace = "test-namespace-" + uuid.New().String()
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
			dummyS3Secret := &v1.Secret{
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
								ConfigurationSecretRef: v1.LocalObjectReference{
									Name: dummyS3Secret.Name,
								},
							},
						},
					},
					Apicast: &appsv1alpha1.ApicastSpec{
						StagingSpec: &appsv1alpha1.ApicastStagingSpec{
							CustomEnvironments: []appsv1alpha1.CustomEnvironmentSpec{
								{
									SecretRef: &v1.LocalObjectReference{
										Name: customEnvSecret.Name,
									},
								},
							},
						},
						ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
							CustomEnvironments: []appsv1alpha1.CustomEnvironmentSpec{
								{
									SecretRef: &v1.LocalObjectReference{
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
			err = waitForAllAPIManagerStandardRoutes(testNamespace, 5*time.Second, 15*time.Minute, wildcardDomain, GinkgoWriter)
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
			err = waitForAPIManagerAvailableCondition(testNamespace, 5*time.Second, 15*time.Minute, apimanager, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "APIManager 'Available' condition is true\n")

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "APIManager creation and availability took '%s'\n", elapsed)
		})
	})
})

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
		createdDeployment := &k8sappsv1.Deployment{}
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

func waitForAllAPIManagerStandardRoutes(namespace string, retryInterval, timeout time.Duration, wildcardDomain string, w io.Writer) error {
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

func waitForAPIManagerAvailableCondition(namespace string, retryInterval, timeout time.Duration, apimanager *appsv1alpha1.APIManager, w io.Writer) error {
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

func waitForAPIManagerLabels(namespace string, retryInterval time.Duration, timeout time.Duration, apimanager *appsv1alpha1.APIManager, customEnvSecret *v1.Secret, w io.Writer) error {
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

func waitForHashedSecret(namespace string, retryInterval time.Duration, timeout time.Duration, customEnvSecret *v1.Secret, w io.Writer) error {
	Eventually(func() bool {
		// First get the master hashed secret
		hashedSecret := &v1.Secret{}
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

func waitForApicastPodAnnotations(namespace string, retryInterval time.Duration, timeout time.Duration, customEnvSecret *v1.Secret, w io.Writer) error {
	apicastDeploymentNames := []string{
		"apicast-production",
		"apicast-staging",
	}

	for _, dName := range apicastDeploymentNames {
		apicastDeploymentLookupKey := types.NamespacedName{Name: dName, Namespace: namespace}
		apicastDeployment := &k8sappsv1.Deployment{}
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

func testGetCustomEnvironmentSecret(namespace string) *v1.Secret {
	customEnvironmentSecret := v1.Secret{
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
