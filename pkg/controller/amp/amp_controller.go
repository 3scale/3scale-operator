package amp

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"

	ampv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/amp/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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

var log = logf.Log.WithName("controller_amp")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AMP Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAMP{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("amp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AMP
	err = c.Watch(&source.Kind{Type: &ampv1alpha1.AMP{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AMP
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ampv1alpha1.AMP{},
	})
	if err != nil {
		return err
	}

	/*
		err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &ampv1alpha1.AMP{},
		})
		if err != nil {
			return err
		}
	*/

	return nil
}

var _ reconcile.Reconciler = &ReconcileAMP{}

// ReconcileAMP reconciles a AMP object
type ReconcileAMP struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AMP object and makes changes based on the state read
// and what is in the AMP.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAMP) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AMP")

	// Fetch the AMP instance
	instance := &ampv1alpha1.AMP{}

	reqLogger.Info("Trying to get AMP resource", "Request", request)
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("AMP Resource not found. Ignoring since object must have been deleted", "client error", err, "AMP", instance)
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "AMP Resource cannot be created. Requeuing request...", "AMP", instance)
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reqLogger.Info("Successfully retreived AMP resource", "AMP", instance)

	reqLogger.Info("Setting defaults for AMP resource")
	instance.SetDefaults() // TODO check where to put this
	reqLogger.Info("Set defaults for AMP resource", "AMP", instance)

	objs, err := createAMP(instance)
	if err != nil {
		reqLogger.Error(err, "Error creating AMP objects")
		return reconcile.Result{}, err
	}

	// Set AMP instance as the owner and controller
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
		// Here means that the object has been able to be obtained
		// and checking for differences should be done to reconcile possible
		// differences that we want to handle
	}

	reqLogger.Info("Finished Current reconcile request successfully. Skipping requeue of the request")
	return reconcile.Result{}, nil
}

func createAMP(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	results, err := createAMPObjects(cr)
	if err != nil {
		return nil, err
	}

	results, err = postProcessAMPObjects(cr, results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func createAMPObjects(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
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

	backend, err := createBackend(cr)
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

	system, err := createSystem(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, system...)

	zync, err := createZync(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, zync...)

	apicast, err := createApicast(cr)
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

func postProcessAMPObjects(cr *ampv1alpha1.AMP, objects []runtime.RawExtension) ([]runtime.RawExtension, error) {
	if cr.Spec.Evaluation {
		e := component.Evaluation{}
		e.PostProcessObjects(objects)
	}

	if cr.Spec.Productized {
		optsProvider := operator.OperatorProductizedOptionsProvider{AmpSpec: &cr.Spec}
		opts, err := optsProvider.GetProductizedOptions()
		if err != nil {
			return nil, err
		}
		p := component.Productized{Options: opts}
		objects = p.PostProcessObjects(objects)
	}

	if cr.Spec.S3Version {
		optsProvider := operator.OperatorS3OptionsProvider{AmpSpec: &cr.Spec}
		opts, err := optsProvider.GetS3Options()
		if err != nil {
			return nil, err
		}
		s := component.S3{Options: opts}
		objects = s.PostProcessObjects(objects)
	}

	return objects, nil
}

func createImages(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorAmpImagesOptionsProvider{AmpSpec: &cr.Spec}
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

func createRedis(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorRedisOptionsProvider{AmpSpec: &cr.Spec}
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

func createBackend(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorBackendOptionsProvider{AmpSpec: &cr.Spec}
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

func createMysql(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorMysqlOptionsProvider{AmpSpec: &cr.Spec}
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

func createMemcached(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorMemcachedOptionsProvider{AmpSpec: &cr.Spec}
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

func createSystem(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorSystemOptionsProvider{AmpSpec: &cr.Spec}
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

func createZync(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorZyncOptionsProvider{AmpSpec: &cr.Spec}
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

func createApicast(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorApicastOptionsProvider{AmpSpec: &cr.Spec}
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

func createWildcardRouter(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorWildcardRouterOptionsProvider{AmpSpec: &cr.Spec}
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

func createS3(cr *ampv1alpha1.AMP) ([]runtime.RawExtension, error) {
	optsProvider := operator.OperatorS3OptionsProvider{AmpSpec: &cr.Spec}
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
