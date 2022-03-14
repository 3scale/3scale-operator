package controllers

import (
	"bytes"
	"context"
	"fmt"

	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	apiv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// TenantInternalReconciler reconciles a Tenant object
type TenantInternalReconciler struct {
	*reconcilers.BaseReconciler
	tenantR     *apiv1alpha1.Tenant
	portaClient *porta_client_pkg.ThreeScaleClient
	logger      logr.Logger
}

// NewTenantInternalReconciler constructs InternalReconciler object
func NewTenantInternalReconciler(b *reconcilers.BaseReconciler, tenantR *apiv1alpha1.Tenant,
	portaClient *porta_client_pkg.ThreeScaleClient, log logr.Logger) *TenantInternalReconciler {
	return &TenantInternalReconciler{
		BaseReconciler: b,
		tenantR:        tenantR,
		portaClient:    portaClient,
		logger:         log,
	}
}

// Run tenant reconciliation logic
// Facts to reconcile:
// - Have 3scale Tenant Account
// - Have active admin user
// - Have secret with tenant's access_token
func (r *TenantInternalReconciler) Run() (ctrl.Result, error) {
	res, err := r.reconcileTenant()
	if err != nil {
		return ctrl.Result{}, err
	}

	if res.Requeue {
		return res, nil
	}

	err = r.reconcileAdminUser()
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

// This method makes sure that tenant exists, otherwise it will create one
// On method completion:
// * tenant will exist
// * tenant's attributes will be updated if required
func (r *TenantInternalReconciler) reconcileTenant() (ctrl.Result, error) {
	tenantDef, err := controllerhelper.FetchTenant(r.tenantR.Status.TenantId, r.portaClient)
	if err != nil {
		return ctrl.Result{}, err
	}

	if tenantDef == nil {
		tenantDef, err = r.createTenant()
		if err != nil {
			return ctrl.Result{}, err
		}

		// Early save access token as it is only available on the response of the
		// tenant creation call

		err = r.reconcileAccessTokenSecret(tenantDef)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Early update status with tenantID
		r.tenantR.Status.TenantId = tenantDef.Signup.Account.ID

		r.logger.Info("Update tenant status with tenantID", "tenantID", tenantDef.Signup.Account.ID)
		err = r.UpdateResourceStatus(r.tenantR)
		if err != nil {
			return ctrl.Result{}, err
		}

		// requeue to have a new run with the updated tenant resource
		return ctrl.Result{Requeue: true}, nil
	}

	r.logger.Info("Tenant already exists", "TenantId", tenantDef.Signup.Account.ID)
	// Check tenant desired state matches current state
	err = r.syncTenant(tenantDef)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TenantInternalReconciler) syncTenant(tenantDef *porta_client_pkg.Tenant) error {
	// If tenant desired state is not current state, update
	triggerSync := func() bool {
		if r.tenantR.Spec.OrganizationName != tenantDef.Signup.Account.OrgName {
			return true
		}

		if r.tenantR.Spec.Email != tenantDef.Signup.Account.SupportEmail {
			return true
		}

		return false
	}()

	if triggerSync {
		r.logger.Info("Syncing tenant", "TenantId", tenantDef.Signup.Account.ID)
		tenantDef.Signup.Account.OrgName = r.tenantR.Spec.OrganizationName
		tenantDef.Signup.Account.SupportEmail = r.tenantR.Spec.Email
		params := porta_client_pkg.Params{
			"support_email": r.tenantR.Spec.Email,
			"org_name":      r.tenantR.Spec.OrganizationName,
		}
		_, err := r.portaClient.UpdateTenant(r.tenantR.Status.TenantId, params)
		if err != nil {
			return err
		}
	}

	return nil
}

////
//
// This method makes sure admin user:
// * is active
// * user's attributes will be updated if required
func (r *TenantInternalReconciler) reconcileAdminUser() error {
	tenantDef, err := controllerhelper.FetchTenant(r.tenantR.Status.TenantId, r.portaClient)
	if err != nil {
		return err
	}

	if tenantDef == nil {
		return fmt.Errorf("tenant with ID %d not found", r.tenantR.Status.TenantId)
	}

	var adminUserDef *porta_client_pkg.User
	if r.tenantR.Status.AdminId == 0 {
		// UserID not in status field
		adminUserDef, err = r.findAdminUser(tenantDef)
		if err != nil {
			return err
		}

		r.tenantR.Status.AdminId = adminUserDef.ID

		r.logger.Info("Update tenant status with adminID", "adminID", adminUserDef.ID)
		err = r.UpdateResourceStatus(r.tenantR)
		if err != nil {
			return err
		}

	} else {
		adminUserDef, err = r.portaClient.ReadUser(tenantDef.Signup.Account.ID, r.tenantR.Status.AdminId)
		if err != nil {
			return err
		}
	}

	err = r.syncAdminUser(tenantDef, adminUserDef)
	if err != nil {
		return err
	}

	return nil
}

// This method makes sure secret with tenant's access_token exists
func (r *TenantInternalReconciler) reconcileAccessTokenSecret(tenantDef *porta_client_pkg.Tenant) error {
	adminURL, err := controllerhelper.URLFromDomain(tenantDef.Signup.Account.AdminDomain)
	if err != nil {
		return err
	}

	desiredSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.tenantR.TenantSecretKey().Namespace,
			Name:      r.tenantR.TenantSecretKey().Name,
			Labels:    map[string]string{"app": "3scale-operator"},
		},
		StringData: map[string]string{
			TenantAccessTokenSecretField:    tenantDef.Signup.AccessToken.Value,
			TenantAdminDomainKeySecretField: adminURL.String(),
		},
		Type: v1.SecretTypeOpaque,
	}

	err = r.SetOwnerReference(r.tenantR, desiredSecret)
	if err != nil {
		return err
	}

	tenantSecretMutator := reconcilers.DeploymentSecretMutator(
		reconcilers.SecretReconcileField(TenantAccessTokenSecretField),
		reconcilers.SecretReconcileField(TenantAdminDomainKeySecretField),
	)

	return r.ReconcileResource(&v1.Secret{}, desiredSecret, tenantSecretMutator)
}

// Create Tenant using porta client
func (r *TenantInternalReconciler) createTenant() (*porta_client_pkg.Tenant, error) {
	password, err := r.getAdminPassword()
	if err != nil {
		return nil, err
	}

	r.logger.Info("Creating a new tenant", "OrganizationName", r.tenantR.Spec.OrganizationName,
		"Username", r.tenantR.Spec.Username, "Email", r.tenantR.Spec.Email)
	return r.portaClient.CreateTenant(
		r.tenantR.Spec.OrganizationName,
		r.tenantR.Spec.Username,
		r.tenantR.Spec.Email,
		password,
	)
}

func (r *TenantInternalReconciler) getAdminPassword() (string, error) {
	// Get tenant admin password from secret reference
	tenantAdminSecret := &v1.Secret{}

	err := r.Client().Get(context.TODO(), r.tenantR.AdminPassSecretKey(), tenantAdminSecret)

	if err != nil {
		return "", err
	}

	passwordByteArray, ok := tenantAdminSecret.Data[TenantAdminPasswordSecretField]
	if !ok {
		return "", fmt.Errorf("Not found admin password secret (%s) attribute: %s",
			r.tenantR.AdminPassSecretKey(),
			TenantAdminPasswordSecretField)
	}

	return bytes.NewBuffer(passwordByteArray).String(), err
}

func (r *TenantInternalReconciler) findAdminUser(tenantDef *porta_client_pkg.Tenant) (*porta_client_pkg.User, error) {
	// Only admin users
	// Any state
	filterParams := porta_client_pkg.Params{
		"role": "admin",
	}
	userList, err := r.portaClient.ListUsers(tenantDef.Signup.Account.ID, filterParams)
	if err != nil {
		return nil, err
	}

	for _, user := range userList.Users {
		if user.User.Email == r.tenantR.Spec.Email && user.User.UserName == r.tenantR.Spec.Username {
			// user is already a copy from User slice element
			return &user.User, nil
		}
	}
	return nil, fmt.Errorf("Admin user not found and should be available"+
		"TenantId: %d. Admin Username: %s, Admin email: %s", tenantDef.Signup.Account.ID,
		r.tenantR.Spec.Username, r.tenantR.Spec.Email)
}

func (r *TenantInternalReconciler) syncAdminUser(tenantDef *porta_client_pkg.Tenant, adminUser *porta_client_pkg.User) error {
	// If adminUser desired state is not current state, update
	if adminUser.State == "pending" {
		err := r.activateAdminUser(tenantDef, adminUser)
		if err != nil {
			return err
		}
	} else {
		r.logger.Info("Admin user already active", "TenantId", tenantDef.Signup.Account.ID, "UserID", adminUser.ID)
	}

	triggerSync := func() bool {
		if r.tenantR.Spec.Username != adminUser.UserName {
			return true
		}

		if r.tenantR.Spec.Email != adminUser.Email {
			return true
		}

		return false
	}()

	if triggerSync {
		r.logger.Info("Syncing adminUser", "TenantId", tenantDef.Signup.Account.ID, "UserID", adminUser.ID)
		adminUser.UserName = r.tenantR.Spec.Username
		adminUser.Email = r.tenantR.Spec.Email
		params := porta_client_pkg.Params{
			"username": r.tenantR.Spec.Username,
			"email":    r.tenantR.Spec.Email,
		}
		_, err := r.portaClient.UpdateUser(tenantDef.Signup.Account.ID, adminUser.ID, params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TenantInternalReconciler) activateAdminUser(tenantDef *porta_client_pkg.Tenant, adminUser *porta_client_pkg.User) error {
	r.logger.Info("Activating pending admin user", "Account ID", tenantDef.Signup.Account.ID, "ID", adminUser.ID)
	return r.portaClient.ActivateUser(tenantDef.Signup.Account.ID, adminUser.ID)
}
