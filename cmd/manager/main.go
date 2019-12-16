package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/apis"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/controller"
	"github.com/3scale/3scale-operator/version"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()
	// Become the leader before proceeding
	err = leader.Become(ctx, "3scale-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for OpenShift routes
	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for OpenShift imagestreams and related
	if err := imagev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for OpenShift deploymentconfigs and related
	if err := appsv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for all monitoring resources
	if err := monitoringv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for all grafana resources
	if err := grafanav1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	register3scaleVersionInfoMetric()

	// Add the Metrics Service
	addMetrics(ctx, cfg, namespace)

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func register3scaleVersionInfoMetric() {
	threeScaleVersionInfo := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "threescale_version_info",
			Help: "3scale Operator version and product version",
			ConstLabels: prometheus.Labels{
				"operator_version": version.Version,
				"version":          product.ThreescaleRelease,
			},
		},
	)
	// Register custom metrics with the global prometheus registry
	crmetrics.Registry.MustRegister(threeScaleVersionInfo)
}

// addMetrics will create the Services and Service Monitors to allow the operator export the metrics by using
// the Prometheus operator
func addMetrics(ctx context.Context, cfg *rest.Config, namespace string) {
	if err := serveCRMetrics(cfg); err != nil {
		if errors.Is(err, k8sutil.ErrRunLocal) {
			log.Info("Skipping CR metrics server creation; not running in a cluster.")
			return
		}
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}

	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Info("Could not create metrics Service", "error", err.Error())
	}

	// Adding the monitoring-key:middleware to the operator service which will get propagated to the serviceMonitor
	err = addMonitoringKeyLabelToOperatorService(ctx, cfg, service)
	if err != nil {
		log.Error(err, "Could not add monitoring-key label to operator metrics Service")
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator.
	services := []*v1.Service{service}
	_, err = metrics.CreateServiceMonitors(cfg, namespace, services)
	if err != nil {
		log.Info("Could not create ServiceMonitor object", "error", err.Error())
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		if err == metrics.ErrServiceMonitorNotPresent {
			log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects", "error", err.Error())
		}
	}
}

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	gvks, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}

	// We perform our custom GKV filtering on top of the one performed
	// by operator-sdk code
	filteredGVK := filterGKVsFromAddToScheme(gvks)
	if err != nil {
		return err
	}

	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}

func filterGKVsFromAddToScheme(gvks []schema.GroupVersionKind) []schema.GroupVersionKind {
	// We use gkvFilters to filter from the existing GKVs defined in the used
	// runtime.Schema for the operator. The reason for that is that
	// kube-metrics tries to list all of the defined Kinds in the schemas
	// that are passed, including Kinds that the operator doesn't use and
	// thus the role used the operator doesn't have them set and we don't want
	// to set as they are not used by the operator.
	// For the fields that the filters have we have defined the value '*' to
	// specify any will be a match (accepted)
	matchAnyValue := "*"
	gvkFilters := []schema.GroupVersionKind{
		// Kubernetes types
		schema.GroupVersionKind{Kind: "PersistentVolumeClaim", Version: matchAnyValue},
		schema.GroupVersionKind{Kind: "ServiceAccount", Version: matchAnyValue},
		schema.GroupVersionKind{Kind: "Secret", Version: matchAnyValue},
		schema.GroupVersionKind{Kind: "Pod", Version: matchAnyValue},
		schema.GroupVersionKind{Kind: "ConfigMap", Version: matchAnyValue},
		schema.GroupVersionKind{Kind: "Service", Version: matchAnyValue},

		// OpenShift types
		schema.GroupVersionKind{Group: "route.openshift.io", Kind: "Route", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "image.openshift.io", Kind: "ImageStream", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "apps.openshift.io", Kind: "DeploymentConfig", Version: matchAnyValue},

		// Custom resource types
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "Plan", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "API", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "Limit", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "MappingRule", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "Tenant", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "Metric", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "capabilities.3scale.net", Kind: "Binding", Version: matchAnyValue},
		schema.GroupVersionKind{Group: "apps.3scale.net", Kind: "APIManager", Version: matchAnyValue},
	}

	ownGVKs := []schema.GroupVersionKind{}
	for _, gvk := range gvks {
		for _, gvkFilter := range gvkFilters {
			match := true
			if gvkFilter.Kind == matchAnyValue && gvkFilter.Group == matchAnyValue && gvkFilter.Version == matchAnyValue {
				log.V(1).Info("gvkFilter should at least have one of its fields defined. Skipping...")
				match = false
			} else {
				if gvkFilter.Kind != matchAnyValue && gvkFilter.Kind != gvk.Kind {
					match = false
				}
				if gvkFilter.Group != matchAnyValue && gvkFilter.Group != gvk.Group {
					match = false
				}
				if gvkFilter.Version != matchAnyValue && gvkFilter.Version != gvk.Version {
					match = false
				}
			}
			if match {
				ownGVKs = append(ownGVKs, gvk)
			}
		}
	}

	return ownGVKs
}

func addMonitoringKeyLabelToOperatorService(ctx context.Context, cfg *rest.Config, service *v1.Service) error {
	if service == nil {
		return fmt.Errorf("service doesn't exist")
	}

	kclient, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}

	updatedLabels := map[string]string{
		"monitoring-key": common.MonitoringKey,
		"app":            appsv1alpha1.Default3scaleAppLabel,
	}
	for k, v := range service.ObjectMeta.Labels {
		updatedLabels[k] = v
	}
	service.ObjectMeta.Labels = updatedLabels

	err = kclient.Update(ctx, service)
	if err != nil {
		return err
	}

	return nil
}
