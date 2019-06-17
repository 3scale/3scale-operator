package apimanager

import (
	"context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis/common"
	appsv1 "github.com/openshift/api/apps/v1"
	"reflect"

	"github.com/go-logr/logr"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/RHsyseng/operator-utils/pkg/olm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// We create an Client Reader that directly queries the API server
// without going to the Cache provided by the Manager's Client because
// there are some resources that do not implement Watch (like ImageStreamTag)
// and the Manager's Client always tries to use the Cache when reading
func newAPIClientReader(mgr manager.Manager) (client.Client, error) {
	return client.New(mgr.GetConfig(), client.Options{Mapper: mgr.GetRESTMapper(), Scheme: mgr.GetScheme()})
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	apiClientReader, err := newAPIClientReader(mgr)
	if err != nil {
		return nil, err
	}
	return &ReconcileAPIManager{client: mgr.GetClient(), apiClientReader: apiClientReader, scheme: mgr.GetScheme(), reqLogger: log}, nil
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

	// Watch for changes to DeploymentConfigs to update deployment status
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIManager{},
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DeploymentConfig{}}, ownerHandler)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAPIManager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAPIManager{}

// ReconcileAPIManager reconciles a APIManager object
type ReconcileAPIManager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client          client.Client
	scheme          *runtime.Scheme
	reqLogger       logr.Logger
	apiClientReader client.Reader
}

// Reconcile reads that state of the cluster for a APIManager object and makes changes based on the state read
// and what is in the APIManager.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPIManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.reqLogger.Info("Reconciling APIManager")

	// Fetch the APIManager instance
	instance := &appsv1alpha1.APIManager{}

	r.reqLogger.Info("Trying to get APIManager resource")
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.reqLogger.Info("APIManager Resource not found. Ignoring since object must have been deleted")
			return reconcile.Result{}, nil
		}
		r.reqLogger.Error(err, "APIManager Resource cannot be created. Requeuing request...")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	r.reqLogger.Info("Successfully retreived APIManager resource")

	r.reqLogger.Info("Setting defaults for APIManager resource")
	changed, err := instance.SetDefaults() // TODO check where to put this
	if err != nil {
		// Error setting defaults - Stop reconciliation
		return reconcile.Result{}, nil
	}
	if changed {
		r.reqLogger.Info("Updating defaults for APIManager resource")
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "APIManager Resource cannot be updated. Requeuing request...")
			return reconcile.Result{}, err
		}
		r.reqLogger.Info("Successfully updated defaults for APIManager resource")
		return reconcile.Result{}, nil
	}

	objs, err := r.apiManagerObjects(instance)
	if err != nil {
		r.reqLogger.Error(err, "Error creating APIManager objects. Requeuing request...")
		return reconcile.Result{}, err
	}

	// Set APIManager instance as the owner and controller
	for idx := range objs {
		obj := objs[idx]
		obj.SetNamespace(instance.Namespace)
		err = controllerutil.SetControllerReference(instance, obj, r.scheme)
		if err != nil {
			r.reqLogger.Error(err, "Error setting OwnerReference on object. Requeuing request...",
				"Kind", obj.GetObjectKind(),
				"Namespace", obj.GetNamespace(),
				"Name", obj.GetName(),
			)
			return reconcile.Result{}, err
		}
	}

	// Create APIManager Objects
	for idx := range objs {
		obj := objs[idx]
		objectInfo := fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())

		newobj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()
		found := newobj.(runtime.Object)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				// TODO for some reason r.client.Create modifies the original object and removes the TypeMeta. Figure why is this???
				createErr := r.client.Create(context.TODO(), obj)
				if createErr != nil {
					r.reqLogger.Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
					return reconcile.Result{}, createErr
				}
				r.reqLogger.Info(fmt.Sprintf("Created object %s", objectInfo))
			} else {
				r.reqLogger.Error(err, fmt.Sprintf("Failed to get %s.  Requeuing request...", objectInfo))
				return reconcile.Result{}, err
			}
		} else {
			r.reqLogger.Info(fmt.Sprintf("Object %s already exists", objectInfo))
			if secret, ok := obj.(*v1.Secret); ok {
				r.reqLogger.Info(fmt.Sprintf("Object %s is a secret. Reconciling it...", objectInfo))
				// We get copy to avoid modifying possibly obtained object
				// from the cache
				foundSecret := found.(*v1.Secret)
				r.reconcileSecret(secret, foundSecret, instance)
				if err != nil {
					r.reqLogger.Error(err, fmt.Sprintf("Failed to update secret secret/%s. Requeuing request...", secret.Name))
					return reconcile.Result{}, err
				}
			}
		}
	}

	err = r.setDeploymentStatus(instance)
	if err != nil {
		r.reqLogger.Error(err, "Failed to set deployment status")
		return reconcile.Result{}, err
	}

	r.reqLogger.Info("Finished Current reconcile request successfully. Skipping requeue of the request")
	return reconcile.Result{}, nil
}

func (r *ReconcileAPIManager) reconcileSecret(desired, current *v1.Secret, cr *appsv1alpha1.APIManager) error {
	// We copy the secrets because we don't know the source of them. Might
	// come from the Cache
	currentCopy := current.DeepCopy()
	desiredCopy := desired.DeepCopy()

	// We convert StringData to Data because stringData cannot be read when
	// obtained from the Kubernetes API and we need to compare the secret
	// data
	desiredCopy.Data = secretStringDataToData(desiredCopy.StringData)
	if secretsEqual(currentCopy, desiredCopy) {
		r.reqLogger.Info(fmt.Sprintf("Secret %s is already reconciled. Update skipped", currentCopy.Name))
		return nil
	}
	currentCopy.StringData = desiredCopy.StringData
	currentCopy.Annotations = desiredCopy.Annotations
	currentCopy.Labels = desiredCopy.Labels
	currentCopy.Finalizers = desiredCopy.Finalizers
	err := controllerutil.SetControllerReference(cr, currentCopy, r.scheme)
	if err != nil {
		return err
	}

	r.reqLogger.Info(fmt.Sprintf("Secret %s is not equal to the expected secret. Updating ...", currentCopy.Name))
	if err = r.client.Update(context.TODO(), currentCopy); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileAPIManager) apiManagerObjects(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	results, err := r.apiManagerObjectsGroup(cr)
	if err != nil {
		return nil, err
	}

	results, err = r.postProcessAPIManagerObjectsGroup(cr, results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ReconcileAPIManager) apiManagerObjectsGroup(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	results := []common.KubernetesObject{}

	images, err := r.createImages(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, images...)

	redis, err := r.createRedis(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, redis...)

	backend, err := r.createBackend(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, backend...)

	systemDB, err := r.createSystemDatabase(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, systemDB...)

	memcached, err := r.createMemcached(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, memcached...)

	system, err := r.createSystem(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, system...)

	zync, err := r.createZync(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, zync...)

	apicast, err := r.createApicast(cr)
	if err != nil {
		return nil, err
	}
	results = append(results, apicast...)

	if cr.Spec.System.FileStorageSpec.S3 != nil {
		s3, err := r.createS3(cr)
		if err != nil {
			return nil, err
		}
		results = append(results, s3...)
	}

	return results, nil
}

func (r *ReconcileAPIManager) postProcessAPIManagerObjectsGroup(cr *appsv1alpha1.APIManager, objects []common.KubernetesObject) ([]common.KubernetesObject, error) {
	if !*cr.Spec.ResourceRequirementsEnabled {
		e := component.Evaluation{}
		e.PostProcessObjects(objects)
	}

	if cr.Spec.System.FileStorageSpec.S3 != nil {
		optsProvider := operator.OperatorS3OptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
		opts, err := optsProvider.GetS3Options()
		if err != nil {
			return nil, err
		}
		s := component.S3{Options: opts}
		objects = s.PostProcessObjects(objects)
	}

	if cr.Spec.HighAvailability != nil && cr.Spec.HighAvailability.Enabled {
		optsProvider := operator.OperatorHighAvailabilityOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
		opts, err := optsProvider.GetHighAvailabilityOptions()
		if err != nil {
			return nil, err
		}
		h := component.HighAvailability{Options: opts}
		objects = h.PostProcessObjects(objects)
	}

	return objects, nil
}

func (r *ReconcileAPIManager) createImages(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
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

func (r *ReconcileAPIManager) createRedis(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorRedisOptionsProvider{APIManagerSpec: &cr.Spec}
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}

	redis := component.Redis{Options: opts}
	result, err := redis.GetObjects()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *ReconcileAPIManager) createBackend(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorBackendOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
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

func (r *ReconcileAPIManager) createSystemDatabase(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	if cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		result, err := r.createSystemPostgreSQL(cr)
		if err != nil {
			return nil, err
		}
		return result, nil
	} else {
		// defaults to MySQL
		result, err := r.createSystemMySQL(cr)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}

func (r *ReconcileAPIManager) createSystemMySQL(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorMysqlOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
	opts, err := optsProvider.GetMysqlOptions()
	if err != nil {
		return nil, err
	}

	m := component.Mysql{Options: opts}
	result, err := m.GetObjects()
	if err != nil {
		return nil, err
	}

	imageOptsProvider := operator.OperatorSystemMySQLImageOptionsProvider{APIManagerSpec: &cr.Spec}
	imageOpts, err := imageOptsProvider.GetSystemMySQLImageOptions()
	if err != nil {
		return nil, err
	}

	i := component.SystemMySQLImage{Options: imageOpts}
	imageresult, err := i.GetObjects()
	if err != nil {
		return nil, err
	}
	result = append(result, imageresult...)
	return result, nil
}

func (r *ReconcileAPIManager) createSystemPostgreSQL(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorSystemPostgreSQLOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
	opts, err := optsProvider.GetSystemPostgreSQLOptions()
	if err != nil {
		return nil, err
	}
	p := component.SystemPostgreSQL{Options: opts}
	result, err := p.GetObjects()
	if err != nil {
		return nil, err
	}
	imageOptsProvider := operator.OperatorSystemPostgreSQLImageOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
	imageOpts, err := imageOptsProvider.GetSystemPostgreSQLImageOptions()
	if err != nil {
		return nil, err
	}
	i := component.SystemPostgreSQLImage{Options: imageOpts}
	imageresult, err := i.GetObjects()
	if err != nil {
		return nil, err
	}
	result = append(result, imageresult...)
	return result, nil
}

func (r *ReconcileAPIManager) createMemcached(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
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

func (r *ReconcileAPIManager) createSystem(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorSystemOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
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

func (r *ReconcileAPIManager) createZync(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorZyncOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
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

func (r *ReconcileAPIManager) createApicast(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorApicastOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
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

func (r *ReconcileAPIManager) createS3(cr *appsv1alpha1.APIManager) ([]common.KubernetesObject, error) {
	optsProvider := operator.OperatorS3OptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: r.client}
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

func (r *ReconcileAPIManager) setDeploymentStatus(instance *appsv1alpha1.APIManager) error {
	listOps := &client.ListOptions{Namespace: instance.Namespace}
	dcList := &appsv1.DeploymentConfigList{}
	err := r.client.List(context.TODO(), listOps, dcList)
	if err != nil {
		r.reqLogger.Error(err, "Failed to list deployment configs")
		return err
	}
	var dcs []appsv1.DeploymentConfig
	for _, dc := range dcList.Items {
		for _, ownerRef := range dc.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				dcs = append(dcs, dc)
				break
			}
		}
	}

	deploymentStatus := olm.GetDeploymentConfigStatus(dcs)
	if !reflect.DeepEqual(instance.Status.Deployments, deploymentStatus) {
		r.reqLogger.Info("Deployment status will be updated")
		instance.Status.Deployments = deploymentStatus
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update API Manager deployment status")
			return err
		}
	}
	return nil
}
