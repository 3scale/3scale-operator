package webconsole

import (
	"context"
	"strings"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/version"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// controllerName is the name of this controller
	controllerName = "controller_webconsole"
	log            = logf.Log.WithName(controllerName)
)

// Add creates a new WebConsole Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	apiClientReader, err := common.NewAPIClientReader(mgr)
	if err != nil {
		return nil, err
	}

	client := mgr.GetClient()
	scheme := mgr.GetScheme()
	ctx := context.TODO()
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &ReconcileWebConsole{
		BaseReconciler: reconcilers.NewBaseReconciler(client, scheme, apiClientReader, ctx, log, discoveryClient, recorder),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("webconsole-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for Route resources
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWebConsole implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWebConsole{}

// ReconcileWebConsole reconciles a WebConsole object
type ReconcileWebConsole struct {
	*reconcilers.BaseReconciler
}

//Reconcile reads the state of the Routes and makes changes to the corresponding Consolelinks
func (r *ReconcileWebConsole) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling ReconcileWebConsole", "Operator version", version.Version)

	kindExists, err := r.HasConsoleLink()
	if err != nil {
		return reconcile.Result{}, err
	}
	if !kindExists {
		logger.Info("Console link not supported in the cluster")
		return reconcile.Result{}, nil
	}

	result, err := r.reconcileMasterLink(request, logger)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		logger.Info("Master link reconciled. Needs Requeueing.")
		return result, nil
	}

	logger.V(1).Info("END")
	return reconcile.Result{}, nil
}

func (r *ReconcileWebConsole) reconcileMasterLink(request reconcile.Request, logger logr.Logger) (reconcile.Result, error) {
	if !strings.Contains(request.Name, "zync-3scale-master") {
		// Nothing to do
		return reconcile.Result{}, nil
	}

	route := &routev1.Route{}
	err := r.Client().Get(r.Context(), request.NamespacedName, route)
	if err != nil && !errors.IsNotFound(err) {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if errors.IsNotFound(err) {
		logger.V(1).Info("Master route not found", "name", request.Name)
		// cluster-scoped resource must not have a namespace-scoped owner
		// So consolelinks cannot have owners like apimanager or route object
		// delete consolelink if exists
		desired := &consolev1.ConsoleLink{
			ObjectMeta: metav1.ObjectMeta{
				Name: helper.GetMasterConsoleLinkName(request.Namespace),
			},
		}
		common.TagObjectToDelete(desired)
		err := r.ReconcileResource(&consolev1.ConsoleLink{}, desired, reconcilers.CreateOnlyMutator)
		return reconcile.Result{}, err
	}

	logger.V(1).Info("Master route found", "name", request.Name)

	err = r.ReconcileResource(&consolev1.ConsoleLink{}, helper.GetMasterConsoleLink(route), helper.GenericConsoleLinkMutator)
	logger.V(1).Info("Reconcile master consolelink", "err", err)
	return reconcile.Result{}, err
}
