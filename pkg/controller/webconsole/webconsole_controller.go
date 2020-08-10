package webconsole

import (
	"context"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	routev1 "github.com/openshift/api/route/v1"
	"time"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
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
	logger.Info("Reconciling ReconcileWebConsole")

	if err := helper.ConsoleLinkSupported(); err == nil {
		if err = helper.ReconcileConsoleLink(r.Context(), r.Client(), &request); err != nil {
			return reconcile.Result{RequeueAfter: time.Duration(200) * time.Microsecond}, nil
		}
	}

	return reconcile.Result{}, nil
}
