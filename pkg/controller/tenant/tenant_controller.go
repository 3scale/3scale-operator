package tenant

import (
	"bytes"
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1alpha1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_tenant")

// Secret field name with Tenant's admin user password
const TenantAdminPasswordSecretField = "admin_password"

// Tenant's credentials secret field name for access token
const TenantProviderKeySecretField = "token"

// Tenant's credentials secret field name for admin domain url
const TenantAdminDomainKeySecretField = "adminURL"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Tenant Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTenant{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tenant-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Tenant
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Tenant{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTenant implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTenant{}

// ReconcileTenant reconciles a Tenant object
type ReconcileTenant struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Tenant object and makes changes based on the state read
// and what is in the Tenant.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTenant) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Tenant")

	// Fetch the Tenant instance
	tenantR := &apiv1alpha1.Tenant{}
	err := r.client.Get(context.TODO(), request.NamespacedName, tenantR)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Tenant resource not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	changed := tenantR.SetDefaults()
	if changed {
		err = r.client.Update(context.TODO(), tenantR)
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Tenant resource updated with defaults")
		// Expect for re-trigger
		return reconcile.Result{}, nil
	}

	masterAccessToken, err := FetchMasterCredentials(r.client, tenantR)
	if err != nil {
		log.Error(err, "Error fetching master credentials secret")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	portaClient, err := controllerhelper.PortaClientFromURLString(tenantR.Spec.SystemMasterUrl, masterAccessToken)
	if err != nil {
		log.Error(err, "Error creating porta client object")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	internalReconciler := NewInternalReconciler(r.client, tenantR, portaClient, reqLogger)
	err = internalReconciler.Run()
	if err != nil {
		log.Error(err, "Error in tenant reconciliation")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("Tenant reconciled successfully")
	return reconcile.Result{}, nil
}

// FetchMasterCredentials get secret using k8s client
func FetchMasterCredentials(k8sClient client.Client, tenantR *apiv1alpha1.Tenant) (string, error) {
	masterCredentialsSecret := &v1.Secret{}

	err := k8sClient.Get(context.TODO(),
		types.NamespacedName{
			Name: tenantR.Spec.MasterCredentialsRef.Name,
			// Master credential secret MUST be on same namespace as tenant CR
			Namespace: tenantR.Namespace,
		},
		masterCredentialsSecret)

	if err != nil {
		return "", err
	}

	masterAccessTokenByteArray, ok := masterCredentialsSecret.Data[component.SystemSecretSystemSeedMasterAccessTokenFieldName]
	if !ok {
		return "", fmt.Errorf("Key not found in master secret (ns: %s, name: %s) key: %s",
			tenantR.Spec.MasterCredentialsRef.Namespace, tenantR.Spec.MasterCredentialsRef.Name,
			component.SystemSecretSystemSeedMasterAccessTokenFieldName)
	}

	return bytes.NewBuffer(masterAccessTokenByteArray).String(), nil
}
