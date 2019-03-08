package binding

import (
	"context"
	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_binding")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

type NonBindingTrigger func(handler.MapObject) []reconcile.Request

func (r NonBindingTrigger) Map(o handler.MapObject) []reconcile.Request {
	return r(o)
}

// newReconciler returns a new reconcile.
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBinding{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	var NonBindingTriggerFunc NonBindingTrigger = func(o handler.MapObject) []reconcile.Request {
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Namespace: o.Meta.GetNamespace(),
				Name:      "_NonBinding",
			}},
		}
	}
	// Create a new controller
	c, err := controller.New("binding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource Binding
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Binding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &apiv1alpha1.API{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Plan{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Limit{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Metric{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.MappingRule{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NonBindingTriggerFunc})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBinding{}

// ReconcileBinding reconciles a Binding object
type ReconcileBinding struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileBinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Binding")
	// If the trigger comes from an object different from a Binding, we will get
	// all the binding object from the same namespace and reconcile them.
	// This is a hack. but we don't have owner references, so it should work.
	if request.Name == "_NonBinding" {
		opts := client.ListOptions{}
		opts.InNamespace(request.Namespace)
		BindingList := &apiv1alpha1.BindingList{}
		err := r.client.List(context.TODO(), &opts, BindingList)
		if err != nil {
			reqLogger.Error(err, "error")
			return reconcile.Result{}, nil
		}

		for _, binding := range BindingList.Items {
			_, err := ReconcileBindingFunc(binding, r.client, reqLogger)
			if err != nil {
				reqLogger.Error(err, "error")
			}
		}
		return reconcile.Result{}, nil
	} else {
		binding := &apiv1alpha1.Binding{}
		err := r.client.Get(context.TODO(), request.NamespacedName, binding)
		if err != nil {
			// if it's not there (user deleted it for ex.)
			if errors.IsNotFound(err) {
				reqLogger.Error(err, "error")
				return reconcile.Result{}, nil
			}
			// Error reading the object - requeue the request.
			reqLogger.Error(err, "error")
			return reconcile.Result{Requeue: true}, err
		}
		return ReconcileBindingFunc(*binding, r.client, reqLogger)

	}
}

func ReconcileBindingFunc(binding apiv1alpha1.Binding, c client.Client, log logr.Logger) (reconcile.Result, error) {

	// UpdateRequired controls whether if we need to update the status of the object or not
	UpdateRequired := false

	// Check if there's a finalizer or set it.
	if binding.HasFinalizer() {
		if binding.IsTerminating() {
			log.Info("Binding is terminating, cleaning up.", binding.Name, binding.Namespace)
			err := binding.CleanUp(c)
			if err != nil {
				log.Info("Clean up for Binding failed.", binding.Name, binding.Namespace)
			}
			return reconcile.Result{}, nil
		}
	} else {
		err := binding.AddFinalizer(c)
		if err != nil {
			log.Error(err, "error")
		}
		// Let's just requeue as we have modified the object.
		return reconcile.Result{Requeue: true}, err
	}

	// Get the current state in the binding object
	initialState, err := binding.GetCurrentState()
	if err != nil {
		log.Error(err, "Error getting the initial state binding")
		return reconcile.Result{RequeueAfter: 1 * time.Minute, Requeue: true}, err
	}

	// Generate a new current state from 3scale
	currentState, err := binding.NewCurrentState(c)
	if err != nil {
		log.Error(err, "Error getting current state from binding status")
		return reconcile.Result{RequeueAfter: 1 * time.Minute, Requeue: true}, err

	}

	// Set the current state in the binding object
	err = binding.SetCurrentState(*currentState)
	if err != nil {
		log.Error(err, "Error Reconciling APIs")
	}

	// If the initial state and the current state are different, set the previousState field in the status and
	// mark the object for update
	if initialState != nil && !apiv1alpha1.CompareStates(*initialState, *currentState) {
		err := binding.SetPreviousState(*initialState)
		if err != nil {
			log.Error(err, "Error setting previous state")
		}
		UpdateRequired = true
	}

	//Generate a new desiredState from the CRDs
	desiredState, err := binding.NewDesiredState(c)
	if err != nil {
		log.Error(err, "Error getting desired state from binding status")

	}
	// Set the desiredState in the binding objects
	err = binding.SetDesiredState(*desiredState)
	if err != nil {
		log.Error(err, "Error Reconciling APIs")
	}

	// Reconcile the previousState, usually to remove a non existant API
	previousState, _ := binding.GetPreviousState()
	if previousState != nil {
		log.Info("Previous State exists, reconciling.", binding.Name, binding.Namespace)

		c, err := apiv1alpha1.NewPortaClient(currentState.Credentials)
		if err != nil {
			log.Error(err, "Failed creating client")
		}
		desiredState, err := binding.GetDesiredState()
		if err != nil {
			log.Error(err, "Failed to get desired state")
		}
		apisDiff := apiv1alpha1.DiffAPIs(previousState.APIs, desiredState.APIs)
		for _, api := range apisDiff.MissingFromB {
			err := api.DeleteFrom3scale(c)
			if err != nil {
				log.Error(err, "Failed to delete internal api from 3scale")
			}
		}
		// Clean the "PreviousState" if needed, and mark the object for udpate
		binding.Status.PreviousState = nil
		UpdateRequired = true
	}

	// Now we check if the State (current, desired) is in sync.
	// if it's not in sync, we reconcile the APIs and mark the object for update
	if binding.StateInSync() {
		log.Info("State is in sync")
	} else {
		log.Info("State is not in sync, reconciling APIs")
		apisDiff := apiv1alpha1.DiffAPIs(desiredState.APIs, currentState.APIs)
		err = apisDiff.ReconcileWith3scale(desiredState.Credentials)
		if err != nil {
			log.Error(err, "Error Reconciling APIs")
		}

		// Refresh the current State
		currentState, err := binding.NewCurrentState(c)
		if err != nil {
			log.Error(err, "Error getting current state from binding status")
		}
		err = binding.SetCurrentState(*currentState)
		if err != nil {
			log.Error(err, "Error Reconciling APIs")
		}

		// Update the LastSync field.
		binding.SetLastSuccessfulSync()
		UpdateRequired = true
		log.Info("Reconciliation finished.")
	}

	// Update the object status fields.
	if UpdateRequired {
		err = binding.UpdateStatus(c)
		if err != nil {
			log.Error(err, "Failed to update status of binding object")
			return reconcile.Result{Requeue: true}, err
		}
	}

	return reconcile.Result{RequeueAfter: 1 * time.Minute, Requeue: true}, nil
}
