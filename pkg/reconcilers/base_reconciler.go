package reconcilers

import (
	"context"
	"fmt"
	"strings"

	"github.com/3scale/3scale-operator/pkg/common"

	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	consolev1 "github.com/openshift/api/console/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("reconcilers")

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func(existing, desired common.KubernetesObject) (bool, error)

func CreateOnlyMutator(existing, desired common.KubernetesObject) (bool, error) {
	return false, nil
}

type BaseReconciler struct {
	client          client.Client
	scheme          *runtime.Scheme
	apiClientReader client.Reader
	ctx             context.Context
	logger          logr.Logger
	discoveryClient discovery.DiscoveryInterface
	recorder        record.EventRecorder
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &BaseReconciler{}

func NewBaseReconciler(ctx context.Context, client client.Client, scheme *runtime.Scheme, apiClientReader client.Reader,
	logger logr.Logger, discoveryClient discovery.DiscoveryInterface, recorder record.EventRecorder) *BaseReconciler {
	return &BaseReconciler{
		client:          client,
		scheme:          scheme,
		apiClientReader: apiClientReader,
		ctx:             ctx,
		logger:          logger,
		discoveryClient: discoveryClient,
		recorder:        recorder,
	}
}

func (b *BaseReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// Client returns a split client that reads objects from
// the cache and writes to the Kubernetes APIServer
func (b *BaseReconciler) Client() client.Client {
	return b.client
}

// APIClientReader return a client that directly reads objects
// from the Kubernetes APIServer
func (b *BaseReconciler) APIClientReader() client.Reader {
	return b.apiClientReader
}

func (b *BaseReconciler) Scheme() *runtime.Scheme {
	return b.scheme
}

func (b *BaseReconciler) Logger() logr.Logger {
	return b.logger
}

func (b *BaseReconciler) DiscoveryClient() discovery.DiscoveryInterface {
	return b.discoveryClient
}

func (b *BaseReconciler) Context() context.Context {
	return b.ctx
}

func (b *BaseReconciler) EventRecorder() record.EventRecorder {
	return b.recorder
}

// ReconcileResource attempts to mutate the existing state
// in order to match the desired state. The object's desired state must be reconciled
// with the existing state inside the passed in callback MutateFn.
//
// obj: Object of the same type as the 'desired' object.
//
//	Used to read the resource from the kubernetes cluster.
//	Could be zero-valued initialized object.
//
// desired: Object representing the desired state
//
// It returns an error.
func (b *BaseReconciler) ReconcileResource(obj, desired common.KubernetesObject, mutateFn MutateFn) error {
	key := client.ObjectKeyFromObject(desired)

	if err := b.Client().Get(b.ctx, key, obj); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// Not found
		if !common.IsObjectTaggedToDelete(desired) {
			return b.CreateResource(desired)
		}

		// Marked for deletion and not found. Nothing to do.
		return nil
	}

	// item found successfully
	if common.IsObjectTaggedToDelete(desired) {
		deletePropagationPolicy := common.GetDeletePropagationPolicyAnnotation(desired)
		if deletePropagationPolicy == nil {
			return b.DeleteResource(desired)
		}
		return b.DeleteResource(desired, client.PropagationPolicy(*deletePropagationPolicy))
	}

	update, err := mutateFn(obj, desired)
	if err != nil {
		return err
	}

	if update {
		return b.UpdateResource(obj)
	}

	return nil
}

func (b *BaseReconciler) GetResource(objKey types.NamespacedName, obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Get object '%s/%s'", strings.Replace(fmt.Sprintf("%T", obj), "*", "", 1), objKey.Name))
	return b.Client().Get(context.TODO(), objKey, obj)
}

func (b *BaseReconciler) CreateResource(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Created object '%s/%s'", strings.Replace(fmt.Sprintf("%T", obj), "*", "", 1), obj.GetName()))
	return b.Client().Create(b.ctx, obj)
}

func (b *BaseReconciler) UpdateResource(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Updated object '%s/%s'", strings.Replace(fmt.Sprintf("%T", obj), "*", "", 1), obj.GetName()))
	return b.Client().Update(b.ctx, obj)
}

func (b *BaseReconciler) DeleteResource(obj common.KubernetesObject, options ...client.DeleteOption) error {
	b.Logger().Info(fmt.Sprintf("Delete object '%s/%s'", strings.Replace(fmt.Sprintf("%T", obj), "*", "", 1), obj.GetName()))
	return b.Client().Delete(context.TODO(), obj, options...)
}

func (b *BaseReconciler) UpdateResourceStatus(obj common.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Updated status of object '%s/%s'", strings.Replace(fmt.Sprintf("%T", obj), "*", "", 1), obj.GetName()))
	return b.Client().Status().Update(context.TODO(), obj)
}

// HasConsoleLink checks if the ConsoleLink is supported in current cluster
func (b *BaseReconciler) HasConsoleLink() (bool, error) {
	return resourceExists(b.DiscoveryClient(),
		consolev1.GroupVersion.String(), "ConsoleLink")
}

// HasGrafanaDashboards checks if the GrafanaDashboard CRD is supported in current cluster
func (b *BaseReconciler) HasGrafanaDashboards() (bool, error) {
	return resourceExists(b.DiscoveryClient(),
		grafanav1alpha1.GroupVersion.String(),
		"GrafanaDashboard")
}

// HasPrometheusRules checks if the PrometheusRules CRD is supported in current cluster
func (b *BaseReconciler) HasPrometheusRules() (bool, error) {
	return resourceExists(b.DiscoveryClient(),
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.PrometheusRuleKind)
}

// HasServiceMonitors checks if the ServiceMonitors CRD is supported in current cluster
func (b *BaseReconciler) HasServiceMonitors() (bool, error) {
	return resourceExists(b.DiscoveryClient(),
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.ServiceMonitorsKind)
}

// HasPodMonitors checks if the PodMonitors CRD is supported in current cluster
func (b *BaseReconciler) HasPodMonitors() (bool, error) {
	return resourceExists(b.DiscoveryClient(),
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.PodMonitorsKind)
}

// SetOwnerReference sets owner as a Controller OwnerReference on owned
func (b *BaseReconciler) SetOwnerReference(owner, obj common.KubernetesObject) error {
	err := controllerutil.SetControllerReference(owner, obj, b.Scheme())
	if err != nil {
		b.Logger().Error(err, "Error setting OwnerReference on object",
			"Kind", obj.GetObjectKind().GroupVersionKind().String(),
			"Namespace", obj.GetNamespace(),
			"Name", obj.GetName(),
		)
	}
	return err
}

// EnsureOwnerReference sets owner as a Controller OwnerReference on owned
// returns boolean to notify when the object has been updated
func (b *BaseReconciler) EnsureOwnerReference(owner, obj common.KubernetesObject) (bool, error) {
	changed := false

	originalSize := len(obj.GetOwnerReferences())
	err := b.SetOwnerReference(owner, obj)
	if err != nil {
		return false, err
	}

	newSize := len(obj.GetOwnerReferences())
	if originalSize != newSize {
		changed = true
	}

	return changed, nil
}

func resourceExists(dc discovery.DiscoveryInterface, groupVersion, kind string) (bool, error) {
	apiLists, err := dc.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Workaround for K8s DiscoveryClient mock where does not return the same
		// error when the GroupVersion queried does not exist when calling
		// ServerResourcesForGroupVersion
		if err.Error() == fmt.Sprintf("GroupVersion %q not found", groupVersion) {
			return false, nil
		}
		return false, err
	}

	for _, r := range apiLists.APIResources {
		if r.Kind == kind {
			return true, nil
		}
	}

	return false, nil
}
