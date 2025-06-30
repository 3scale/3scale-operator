/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg              *rest.Config
	testK8sClient    client.Client
	testK8sAPIClient client.Reader
	testEnv          *envtest.Environment
)

func TestAPIManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "APIManager Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}

	var err error
	Eventually(func() bool {
		fmt.Fprintf(GinkgoWriter, "starting apps testEnv...\n")
		cfg, err = testEnv.Start()
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "apps testEnv start attempt failed: %v'\n", err)
			return false
		}
		fmt.Fprintf(GinkgoWriter, "apps testEnv started\n")
		return true
	}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "testEnv failed to start reached max attempts")
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = appsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = capabilitiesv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = capabilitiesv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = routev1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = consolev1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = imagev1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = appsv1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = grafanav1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = configv1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	discoveryClientAPIManager, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	Expect(err).ToNot(HaveOccurred())

	err = (&APIManagerReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(),
			mgr.GetClient(),
			mgr.GetScheme(),
			mgr.GetAPIReader(),
			// zap.LoggerTo(ioutil.Discard, true),
			ctrl.Log.WithName("controllers").WithName("APIManager"),
			discoveryClientAPIManager,
			mgr.GetEventRecorderFor("APIManager"),
		),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = mgr.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	testK8sClient = mgr.GetClient()
	Expect(testK8sClient).ToNot(BeNil())
	testK8sAPIClient = mgr.GetAPIReader()
	Expect(testK8sAPIClient).ToNot(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
