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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	appscontroller "github.com/3scale/3scale-operator/controllers/apps"
	capabilitiescontroller "github.com/3scale/3scale-operator/controllers/capabilities"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/getkin/kin-openapi/openapi3"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	grafanav1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	controllerruntimemetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	subcontroller "github.com/3scale/3scale-operator/controllers/subscription"
	"github.com/3scale/3scale-operator/pkg/helper"
	operatorsv2 "github.com/operator-framework/api/pkg/operators/v2"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = apimachineryruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	// Avoid OpenAPI schema formatvalidation
	// invalid components: unsupported 'format' value "uuid"
	// https://github.com/getkin/kin-openapi/issues/442
	// https://pkg.go.dev/github.com/getkin/kin-openapi@v0.80.0/openapi3#SchemaFormatValidationDisabled
	openapi3.SchemaFormatValidationDisabled = true

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorsv2.AddToScheme(scheme))
	utilruntime.Must(operatorsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(capabilitiesv1alpha1.AddToScheme(scheme))
	utilruntime.Must(capabilitiesv1beta1.AddToScheme(scheme))
	utilruntime.Must(routev1.Install(scheme))
	utilruntime.Must(consolev1.Install(scheme))
	utilruntime.Must(imagev1.Install(scheme))
	utilruntime.Must(appsv1.Install(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	utilruntime.Must(grafanav1alpha1.AddToScheme(scheme))
	utilruntime.Must(grafanav1beta1.AddToScheme(scheme))
	utilruntime.Must(configv1.Install(scheme))

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool

	// https://v1-2-x.sdk.operatorframework.io/docs/building-operators/golang/references/logging/#a-simple-example
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling flag.Parse().
	loggerOpts := zap.Options{}
	loggerOpts.BindFlags(flag.CommandLine)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&loggerOpts)))

	printVersion()

	namespace, err := helper.GetWatchNamespace()
	if err != nil {
		setupLog.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	operatorInstallationNamespace, err := helper.GetOperatorNamespace()
	if err != nil {
		setupLog.Error(err, "Failed to retrieve operator namespace")
		os.Exit(1)
	}

	// If a watch namespace is detected (i.e. operator is namespace scoped), then pass the NS to cache.Options.DefaultNamespaces
	// If no watch namespace is detected (i.e. operator is cluster scoped), then pass an empty Cache object
	var managerCache = cache.Options{}
	if namespace != "" {
		managerCache = cache.Options{
			// If running in target ns mode or own ns mode, include the operator ns and product ns include both, sub and config maps inn cache
			ByObject: map[client.Object]cache.ByObject{
				&corev1.ConfigMap{}: {
					Namespaces: map[string]cache.Config{
						operatorInstallationNamespace: {},
						namespace:                     {},
					},
				},
				&operatorsv1alpha1.Subscription{}: {
					Namespaces: map[string]cache.Config{
						operatorInstallationNamespace: {},
						namespace:                     {},
					},
				},
			},
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Cache:            managerCache,
		Scheme:           scheme,
		Metrics:          metricsserver.Options{BindAddress: metricsAddr},
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "82355b9c.3scale.net",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	secretLabelSelector, err := apimachinerymetav1.ParseToLabelSelector("apimanager.apps.3scale.net/watched-by=apimanager")
	if err != nil {
		setupLog.Error(err, "unable parse apimanager secrets label")
		os.Exit(1)
	}

	if secretLabelSelector == nil {
		setupLog.Info("secretLabelSelector is empty")
		os.Exit(1)
	}

	discoveryClientAPIManager, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&appscontroller.APIManagerReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("APIManager"),
			discoveryClientAPIManager,
			mgr.GetEventRecorderFor("APIManager")),
		SecretLabelSelector: *secretLabelSelector,
		WatchedNamespace:    namespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIManager")
		os.Exit(1)
	}

	// Subscription controller - skipping subscription controller if preflights are bypassed.
	if !helper.IsPreflightBypassed() {
		restConfig := ctrl.GetConfigOrDie()
		restConfig.Timeout = time.Second * 10
		k8sclient, err := client.New(restConfig, client.Options{
			Scheme: mgr.GetScheme(),
		})
		if err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Subscription")
			os.Exit(1)
		}

		setupLog.Info(fmt.Sprintf("Operator Namespace is: %s", operatorInstallationNamespace))
		discoveryClientSubscription, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
		if err != nil {
			setupLog.Error(err, "unable to create discovery client")
			os.Exit(1)
		}
		if err = (&subcontroller.SubscriptionReconciler{
			BaseReconciler: reconcilers.NewBaseReconciler(
				context.Background(), k8sclient, mgr.GetScheme(), mgr.GetAPIReader(),
				ctrl.Log.WithName("controllers").WithName("Subscription"),
				discoveryClientSubscription,
				mgr.GetEventRecorderFor("Subscription")),
			OperatorNamespace: operatorInstallationNamespace,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Subscription")
			os.Exit(1)
		}
	}

	discoveryClientAPIManagerBackup, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&appscontroller.APIManagerBackupReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("APIManagerBackup"),
			discoveryClientAPIManagerBackup,
			mgr.GetEventRecorderFor("APIManagerBackup")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIManagerBackup")
		os.Exit(1)
	}

	discoveryClientAPIManagerRestore, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&appscontroller.APIManagerRestoreReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("APIManagerRestore"),
			discoveryClientAPIManagerRestore,
			mgr.GetEventRecorderFor("APIManagerRestore")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIManagerRestore")
		os.Exit(1)
	}

	discoveryClientTenant, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&capabilitiescontroller.TenantReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("Tenant"),
			discoveryClientTenant,
			mgr.GetEventRecorderFor("Tenant")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}

	discoveryClientBackend, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&capabilitiescontroller.BackendReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("Backend"),
			discoveryClientBackend,
			mgr.GetEventRecorderFor("Backend")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Backend")
		os.Exit(1)
	}

	discoveryClientProduct, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&capabilitiescontroller.ProductReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("Product"),
			discoveryClientProduct,
			mgr.GetEventRecorderFor("Product")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Product")
		os.Exit(1)
	}

	discoveryClientOpenAPI, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&capabilitiescontroller.OpenAPIReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("OpenAPI"),
			discoveryClientOpenAPI,
			mgr.GetEventRecorderFor("OpenAPI")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OpenAPI")
		os.Exit(1)
	}

	discoveryClientWebConsole, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&appscontroller.WebConsoleReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("WebConsole"),
			discoveryClientWebConsole,
			mgr.GetEventRecorderFor("WebConsole")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WebConsole")
		os.Exit(1)
	}

	discoveryClientActiveDoc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	if err = (&capabilitiescontroller.ActiveDocReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("ActiveDoc"),
			discoveryClientActiveDoc,
			mgr.GetEventRecorderFor("ActiveDoc")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ActiveDoc")
		os.Exit(1)
	}

	discoveryClientCustomPolicyDefinition, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if err = (&capabilitiescontroller.CustomPolicyDefinitionReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("CustomPolicyDefinition"),
			discoveryClientCustomPolicyDefinition,
			mgr.GetEventRecorderFor("CustomPolicyDefinition")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomPolicyDefinition")
		os.Exit(1)
	}

	discoveryClientDeveloperAccount, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if err = (&capabilitiescontroller.DeveloperAccountReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("DeveloperAccount"),
			discoveryClientDeveloperAccount,
			mgr.GetEventRecorderFor("DeveloperAccount")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DeveloperAccount")
		os.Exit(1)
	}

	discoveryClientDeveloperUser, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if err = (&capabilitiescontroller.DeveloperUserReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("DeveloperUser"),
			discoveryClientDeveloperUser,
			mgr.GetEventRecorderFor("DeveloperUser")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DeveloperUser")
		os.Exit(1)
	}

	registerThreescaleMetricsIntoControllerRuntimeMetricsRegistry()

	discoveryProxyConfigPromote, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if err = (&capabilitiescontroller.ProxyConfigPromoteReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("ProxyConfigPromote"),
			discoveryProxyConfigPromote,
			mgr.GetEventRecorderFor("ProxyConfigPromote")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ProxyConfigPromote")
		os.Exit(1)
	}

	discoveryApplication, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if err = (&capabilitiescontroller.ApplicationReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("Application"),
			discoveryApplication,
			mgr.GetEventRecorderFor("Application")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	discoveryApplicationAuth, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if err = (&capabilitiescontroller.ApplicationAuthReconciler{
		BaseReconciler: reconcilers.NewBaseReconciler(
			context.Background(), mgr.GetClient(), mgr.GetScheme(), mgr.GetAPIReader(),
			ctrl.Log.WithName("controllers").WithName("Application"),
			discoveryApplicationAuth,
			mgr.GetEventRecorderFor("Application")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func printVersion() {
	setupLog.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	setupLog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func registerThreescaleMetricsIntoControllerRuntimeMetricsRegistry() {
	register3scaleVersionInfoMetric()
}

func register3scaleVersionInfoMetric() {
	threeScaleVersionInfo := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "threescale_version_info",
			Help: "3scale Operator version and product version",
			ConstLabels: prometheus.Labels{
				"operator_version": version.Version,
				"version":          version.ThreescaleVersionMajorMinor(),
			},
		},
	)
	// Register custom metrics with the global prometheus registry
	controllerruntimemetrics.Registry.MustRegister(threeScaleVersionInfo)
}
