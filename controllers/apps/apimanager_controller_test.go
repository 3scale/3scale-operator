package controllers

import (
	"context"
	"fmt"
	"io"
	"time"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			if err != nil {
				return false
			}
			return true
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

			enableResourceRequirements := false
			wildcardDomain := "test1.127.0.0.1.nip.io"
			apimanager := &appsv1alpha1.APIManager{
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						WildcardDomain:              wildcardDomain,
						ResourceRequirementsEnabled: &enableResourceRequirements,
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example-apimanager",
					Namespace: testNamespace,
				},
			}

			err := testK8sClient.Create(context.Background(), apimanager)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: apimanager.Name, Namespace: apimanager.Namespace}, apimanager)
				if err != nil {
					return false
				}
				return true
			}, 5*time.Minute, 5*time.Second).Should(BeTrue())

			err = waitForAllAPIManagerStandardDeploymentConfigs(testNamespace, 5*time.Second, 15*time.Minute, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			// err = waitForAllAPIManagerStandardRoutes(testNamespace, 5*time.Second, 15*time.Minute, wildcardDomain, GinkgoWriter)
			// Expect(err).ToNot(HaveOccurred())

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "APIcast creation and availability took '%s'\n", elapsed)
		})
	})
})

func waitForAllAPIManagerStandardDeploymentConfigs(namespace string, retryInterval, timeout time.Duration, w io.Writer) error {
	deploymentConfigNames := []string{ // TODO gather this from constants/somewhere centralized
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
		"system-sphinx",
		"zync",
		"zync-que",
		"zync-database",
	}

	for _, dcName := range deploymentConfigNames {
		lookupKey := types.NamespacedName{Name: dcName, Namespace: namespace}
		createdDeployment := &appsv1.DeploymentConfig{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), lookupKey, createdDeployment)
			if err != nil {
				return false
			}

			isReady := false
			dcConditions := createdDeployment.Status.Conditions
			for _, dcCondition := range dcConditions {
				if dcCondition.Type == appsv1.DeploymentAvailable && dcCondition.Status == corev1.ConditionTrue {
					isReady = true
				}
			}
			if isReady {
				fmt.Fprintf(w, "DeploymentConfig '%s' available\n", dcName)
				return true
			}
			availableReplicas := createdDeployment.Status.AvailableReplicas
			desiredReplicas := createdDeployment.Spec.Replicas
			fmt.Fprintf(w, "Waiting for full availability of %s DeploymentConfig (%d/%d)\n", dcName, availableReplicas, desiredReplicas)
			return false

		}, 15*time.Minute, retryInterval).Should(BeTrue())
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
		lookupKey := types.NamespacedName{Name: routeHost, Namespace: namespace}
		createdRoute := &routev1.Route{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), lookupKey, createdRoute)
			if err != nil {
				return false
			}

			routeStatusIngresses := createdRoute.Status.Ingress
			if routeStatusIngresses == nil || len(routeStatusIngresses) == 0 {
				fmt.Fprintf(w, "Waiting for availability of Route with host '%s'\n", routeHost)
				return false
			}

			for _, routeStatusIngress := range routeStatusIngresses {
				routeStatusIngressConditions := routeStatusIngress.Conditions
				isReady := false
				for _, routeStatusIngressCondition := range routeStatusIngressConditions {
					if routeStatusIngressCondition.Type == routev1.RouteAdmitted && routeStatusIngressCondition.Status == corev1.ConditionTrue {
						isReady = true
						break
					}
				}
				if !isReady {
					fmt.Fprintf(w, "Waiting for availability of Route with host '%s'\n", routeHost)
					return false
				}
			}

			fmt.Fprintf(w, "Route '%s' with host '%s' available\n", createdRoute.Name, createdRoute.Spec.Host)
			return true
		}, 15*time.Minute, retryInterval).Should(BeTrue())

		return nil
	}

	return nil
}
