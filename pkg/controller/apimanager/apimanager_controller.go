package apimanager

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apimanager")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new APIManager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAPIManager{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("apimanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource APIManager
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.APIManager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAPIManager{}

// ReconcileAPIManager reconciles a APIManager object
type ReconcileAPIManager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a APIManager object and makes changes based on the state read
// and what is in the APIManager.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPIManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling APIManager")

	// Fetch the APIManager instance
	instance := &appsv1alpha1.APIManager{}

	reqLogger.Info("Trying to get APIManager resource", "Request", request)
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("APIManager Resource not found. Ignoring since object must have been deleted", "client error", err, "APIManager", instance)
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "APIManager Resource cannot be created. Requeuing request...", "APIManager", instance)
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reqLogger.Info("Successfully retreived APIManager resource", "APIManager", instance)

	reqLogger.Info("Setting defaults for APIManager resource")
	instance.SetDefaults() // TODO check where to put this
	reqLogger.Info("Set defaults for APIManager resource", "APIManager", instance)

	objs, err := createAPIManager(instance, r.client)
	if err != nil {
		reqLogger.Error(err, "Error creating APIManager objects")
		return reconcile.Result{}, err
	}

	// Set APIManager instance as the owner and controller
	for idx := range objs {
		objectMeta := (objs[idx].Object).(metav1.Object)
		objectMeta.SetNamespace(instance.Namespace)
		err = controllerutil.SetControllerReference(instance, (objs[idx].Object).(metav1.Object), r.scheme)
		if err != nil {
			reqLogger.Error(err, "Object", objs[idx].Object)
			return reconcile.Result{}, err
		}
	}

	// Create zync Objects
	for idx := range objs {
		obj := objs[idx].Object
		objCopy := obj.DeepCopyObject() // We create a copy because the r.client.Create method removes TypeMeta for some reason
		objectMeta := objCopy.(metav1.Object)
		objectInfo := fmt.Sprintf("%s/%s", objCopy.GetObjectKind().GroupVersionKind().Kind, objectMeta.GetName())

		newobj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()
		found := newobj.(runtime.Object)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: objectMeta.GetName(), Namespace: objectMeta.GetNamespace()}, found)
		if err != nil && errors.IsNotFound(err) {
			// TODO for some reason r.client.Create modifies the original object and removes the TypeMeta. Figure why is this???
			err = r.client.Create(context.TODO(), obj)
			if err != nil {
				reqLogger.Error(err, "Error creating object "+objectInfo, obj)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Created object " + objectInfo)
		} else if err != nil {
			reqLogger.Error(err, "Failed to get "+objectInfo)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Object " + objectInfo + " already exists")

		// Update secrets with consistent data
		if secret, ok := objCopy.(*v1.Secret); ok {
			reqLogger.Info("Object is a secret. Updating object...")
			err = r.client.Update(context.TODO(), secret)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		// Here means that the object has been able to be obtained
		// and checking for differences should be done to reconcile possible
		// differences that we want to handle
	}

	reqLogger.Info("Finished Current reconcile request successfully. Skipping requeue of the request")
	return reconcile.Result{}, nil
}

func createAPIManager(cr *appsv1alpha1.APIManager, client client.Client) ([]runtime.RawExtension, error) {
	results, err := createAPIManagerObjects(cr, client)
	if err != nil {
		return nil, err
	}

	results, err = postProcessAPIManagerObjects(cr, results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func createAPIManagerObjects(cr *appsv1alpha1.APIManager, client client.Client) ([]runtime.RawExtension, error) {
	results := []runtime.RawExtension{}

	images, err := createImages(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, images...)

	redis, err := createRedis(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, redis...)

	backend, err := createBackend(cr, client)
	if err != nil {
		return nil, err
	}
	results = append(results, backend...)

	mysql, err := createMysql(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, mysql...)

	memcached, err := createMemcached(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, memcached...)

	system, err := createSystem(cr, client)
	if err != nil {
		return nil, err
	}
	results = append(results, system...)

	zync, err := createZync(cr, client)
	if err != nil {
		return nil, err
	}
	results = append(results, zync...)

	apicast, err := createApicast(cr, client)
	if err != nil {
		return nil, err
	}
	results = append(results, apicast...)

	wildcardRouter, err := createWildcardRouter(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, wildcardRouter...)

	if cr.Spec.S3Version {
		s3, err := createS3(cr)
		if err != nil {
			return nil, err
		}
		results = append(results, s3...)
	}

	return results, nil
}

func postProcessAPIManagerObjects(cr *appsv1alpha1.APIManager, objects []runtime.RawExtension) ([]runtime.RawExtension, error) {
	if cr.Spec.Evaluation {
		e := component.Evaluation{}
		e.PostProcessObjects(objects)
	}

	if cr.Spec.Productized {
		optsProvider := operator.OperatorProductizedOptionsProvider{APIManagerSpec: &cr.Spec}
		opts, err := optsProvider.GetProductizedOptions()
		if err != nil {
			return nil, err
		}
		p := component.Productized{Options: opts}
		objects = p.PostProcessObjects(objects)
	}

	if cr.Spec.S3Version {
		optsProvider := operator.OperatorS3OptionsProvider{APIManagerSpec: &cr.Spec}
		opts, err := optsProvider.GetS3Options()
		if err != nil {
			return nil, err
		}
		s := component.S3{Options: opts}
		objects = s.PostProcessObjects(objects)
	}

	if cr.Spec.HAVersion {
		optsProvider := operator.OperatorHighAvailabilityOptionsProvider{APIManagerSpec: &cr.Spec}
		opts, err := optsProvider.GetHighAvailabilityOptions()
		if err != nil {
			return nil, err
		}
		h := component.HighAvailability{Options: opts}
		objects = h.PostProcessObjects(objects)
	}

	return objects, nil
}

func createImages(cr *appsv1alpha1.APIManager) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorAmpImagesOptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetAmpImagesOptions()
	if err != nil {
		return nil, err
	}

	i := component.AmpImages{Options: opts}
	result, err := i.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createRedis(cr *appsv1alpha1.APIManager) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorRedisOptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}

	r := component.Redis{Options: opts}
	result, err := r.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createBackend(cr *appsv1alpha1.APIManager, client client.Client) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorBackendOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: client}
	opts, err := optsProvider.GetBackendOptions()
	if err != nil {
		return nil, err
	}

	b := component.Backend{Options: opts}
	result, err := b.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createMysql(cr *appsv1alpha1.APIManager) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorMysqlOptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetMysqlOptions()
	if err != nil {
		return nil, err
	}

	m := component.Mysql{Options: opts}
	result, err := m.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createMemcached(cr *appsv1alpha1.APIManager) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorMemcachedOptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetMemcachedOptions()
	if err != nil {
		return nil, err
	}

	i := component.Memcached{Options: opts}
	result, err := i.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createSystem(cr *appsv1alpha1.APIManager, client client.Client) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorSystemOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: client}
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}

	i := component.System{Options: opts}
	result, err := i.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createZync(cr *appsv1alpha1.APIManager, client client.Client) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorZyncOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: client}
	opts, err := optsProvider.GetZyncOptions()
	if err != nil {
		return nil, err
	}

	z := component.Zync{Options: opts}
	result, err := z.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createApicast(cr *appsv1alpha1.APIManager, client client.Client) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorApicastOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: client}
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}

	z := component.Apicast{Options: opts}
	result, err := z.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createWildcardRouter(cr *appsv1alpha1.APIManager) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorWildcardRouterOptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetWildcardRouterOptions()
	if err != nil {
		return nil, err
	}
	z := component.WildcardRouter{Options: opts}
	result, err := z.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createS3(cr *appsv1alpha1.APIManager) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorS3OptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetS3Options()
	if err != nil {
		return nil, err
	}
	s := component.S3{Options: opts}
	result, err := s.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}
