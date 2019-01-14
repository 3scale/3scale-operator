package binding

import (
	"context"
	"encoding/json"
	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

type NonBindingTrigger func(handler.MapObject) []reconcile.Request

func (r NonBindingTrigger) Map(o handler.MapObject) []reconcile.Request {
	return r(o)
}

// newReconciler returns a new reconcile.Reconciler
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
	// All those objects can change the outcome of the consolidated objects because the binding points to it.
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.API{}}, &handler.EnqueueRequestsFromMapFunc{NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Plan{}}, &handler.EnqueueRequestsFromMapFunc{NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Limit{}}, &handler.EnqueueRequestsFromMapFunc{NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Metric{}}, &handler.EnqueueRequestsFromMapFunc{NonBindingTriggerFunc})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.MappingRule{}}, &handler.EnqueueRequestsFromMapFunc{NonBindingTriggerFunc})
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
	log.Printf("Reconciling Binding %s/%s\n", request.Namespace, request.Name)

	// If the trigger comes from an object different from a Binding, we will get
	// all the binding object from the same namespace and reconcile them.
	if request.Name == "_NonBinding" {
		opts := client.ListOptions{}
		opts.InNamespace(request.Namespace)
		BindingList := &apiv1alpha1.BindingList{}
		err := r.client.List(context.TODO(), &opts, BindingList)
		if err != nil {
			log.Printf("error: %s", err)
			return reconcile.Result{}, nil
		}

		for _, binding := range BindingList.Items {
			_, err := ReconcileBindingFunc(binding, r.client)
			if err != nil {
				log.Printf("Error: %s", err)
			}
		}
		return reconcile.Result{}, nil
	} else {
		binding := &apiv1alpha1.Binding{}
		err := r.client.Get(context.TODO(), request.NamespacedName, binding)
		if err != nil {
			// if it's not there (user deleted it for ex.)
			if errors.IsNotFound(err) {
				log.Printf("error: %s", err)
				return reconcile.Result{}, nil
			}
			// Error reading the object - requeue the request.
			log.Printf("Error: %s", err)
			return reconcile.Result{Requeue: true}, err
		}
		return ReconcileBindingFunc(*binding, r.client)

	}
}

func ReconcileBindingFunc(binding apiv1alpha1.Binding, c client.Client) (reconcile.Result, error) {

	//consolidated := apiv1alpha1.NewConsolidated(binding.Name+"-consolidated", binding.Namespace, nil, nil)
	// Create an empty consolidated object to represent the current state.
	currentConsolidated := &apiv1alpha1.Consolidated{}

	// Try to get the current Consolidated object based on the binding name and namespace
	// we append "-consolidated" to the binding object name.
	err := c.Get(context.TODO(), types.NamespacedName{Name: binding.Name + "-consolidated", Namespace: binding.Namespace}, currentConsolidated)

	// IF Consolidated doesn't exists, let's create it.
	if err != nil && errors.IsNotFound(err) {
		// Getting the current consolidated object failed due to it being non-existent.
		// Let's create it!
		log.Printf("Reconcile: Creating new Consolidated object: %s/%s.", currentConsolidated.Namespace, currentConsolidated.Name)
		consolidated, err := apiv1alpha1.NewConsolidatedFromBinding(binding, c)
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}
		trueVar := true
		consolidated.SetOwnerReferences(append(consolidated.GetOwnerReferences(), v1.OwnerReference{
			APIVersion: "api.3scale.net/v1alpha1",
			Kind:       "Binding",
			Name:       binding.Name,
			UID:        binding.UID,
			Controller: &trueVar,
		}))

		err = c.Create(context.TODO(), consolidated)
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}

	} else if err != nil {
		// Something is broken when trying to get the existing consolidated Object
		log.Printf("error: %s", err)
		return reconcile.Result{Requeue: true}, err
	} else {
		// Consolidated Object Already exists
		// Let's compare those, first, we should calculate the desired Consolidated object.
		desiredConsolidated, err := apiv1alpha1.NewConsolidatedFromBinding(binding, c)
		if err != nil {
			log.Printf("error: %s", err)
			return reconcile.Result{Requeue: true}, err
		}
		// Compare with the current consolidated object.
		if compareConsolidated(currentConsolidated, desiredConsolidated) {
			// Desired and existing are equal, nothing to do.
			log.Printf("Skip reconcile: Consolidated config %s/%s ok.", currentConsolidated.Namespace, currentConsolidated.Name)
		} else {
			// Consolidated Objects are different, let's update the existing one with the desired object.
			trueVar := true
			desiredConsolidated.SetOwnerReferences(append(desiredConsolidated.GetOwnerReferences(), v1.OwnerReference{
				APIVersion: "api.3scale.net/v1alpha1",
				Kind:       "Binding",
				Name:       binding.Name,
				UID:        binding.UID,
				Controller: &trueVar,
			}))
			// Set the proper Meta from the existing object.
			desiredConsolidated.ObjectMeta = currentConsolidated.ObjectMeta
			err := c.Update(context.TODO(), desiredConsolidated)
			if err != nil {
				// Something went wrong when trying to update the actual consolidated object.
				log.Printf("error: %s", err)
				return reconcile.Result{Requeue: true}, err
			}
			// All done here
			log.Printf("Reconcile: Consolidated config %s/%s has been updated.", currentConsolidated.Namespace, currentConsolidated.Name)
		}
	}
	// Return without errors and resume the reconcile loop
	return reconcile.Result{}, nil
}

func compareConsolidated(consolidatedA, consolidatedB *apiv1alpha1.Consolidated) bool {
	//Let's compare only the Spec
	A, _ := json.Marshal(consolidatedA.Spec)
	B, _ := json.Marshal(consolidatedB.Spec)
	log.Printf("\n\n %s \n\n\n  %s \n\n", A, B)
	return reflect.DeepEqual(A, B)
}
