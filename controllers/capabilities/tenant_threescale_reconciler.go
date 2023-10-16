package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// TenantThreescaleReconciler reconciles a Tenant object
type TenantThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	tenantR     *apiv1alpha1.Tenant
	portaClient *porta_client_pkg.ThreeScaleClient
	logger      logr.Logger
}

// NewTenantThreescaleReconciler constructs InternalReconciler object
func NewTenantThreescaleReconciler(b *reconcilers.BaseReconciler, tenantR *apiv1alpha1.Tenant,
	portaClient *porta_client_pkg.ThreeScaleClient, log logr.Logger) *TenantThreescaleReconciler {
	return &TenantThreescaleReconciler{
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
func (r *TenantThreescaleReconciler) Run() error {
	updateRequired, err := r.reconcileTenant()
	if err != nil {
		return err
	}

	if updateRequired {
		return nil
	}

	err = r.reconcileAdminUser()
	if err != nil {
		return err
	}

	return nil
}

// This method makes sure that tenant exists, otherwise it will create one
// On method completion:
// * tenant will exist
// * tenant's attributes will be updated if required
func (r *TenantThreescaleReconciler) reconcileTenant() (bool, error) {
	tenantID, err := r.retrieveTenantID()
	if err != nil {
		return false, errors.New("failed to convert tenantID annotation to int64")
	}

	tenantDef, err := controllerhelper.FetchTenant(tenantID, r.portaClient)
	if err != nil {
		return false, err
	}

	if tenantDef == nil {
		tenantDef, err = r.createTenant()
		if err != nil {
			return false, err
		}

		// Early save access token as it is only available on the response of the
		// tenant creation call

		err = r.reconcileAccessTokenSecret(tenantDef)
		if err != nil {
			return false, err
		}

		// Early update status with the new tenantID
		newStatus := &apiv1alpha1.TenantStatus{
			// reset adminID. It could keep old stale value
			AdminId:  0,
			TenantId: tenantDef.Signup.Account.ID,
		}

		updated, err := r.reconcileStatus(newStatus)
		if err != nil {
			return false, err
		}

		// If updated - update the status and requeue
		if updated {
			return true, nil
		}
	}

	return false, nil
}

// This method makes sure admin user:
// * is active
// * user's attributes will be updated if required
func (r *TenantThreescaleReconciler) reconcileAdminUser() error {
	tenantID, err := r.retrieveTenantID()
	if err != nil {
		return errors.New("failed to convert tenantID annotation to int64")
	}

	if tenantID == 0 {
		return errors.New("trying to reconcile admin user when tenantID 0")
	}

	adminUser, err := controllerhelper.FindUser(r.portaClient, tenantID,
		r.tenantR.Spec.Email, r.tenantR.Spec.Username)
	if err != nil {
		return err
	}

	if adminUser == nil {
		adminUser, err = controllerhelper.CreateAdminUser(r.portaClient,
			tenantID, r.tenantR.Spec.Email, r.tenantR.Spec.Username)
		if err != nil {
			return &helper.WaitError{
				Err: fmt.Errorf("3scale client failed creating the admin user: %v", err),
			}
		}
	}

	if adminUser.Element.ID == nil {
		return fmt.Errorf("admin returned nil ID for tenantID %d;"+
			"email %s; username:%s", tenantID, r.tenantR.Spec.Email,
			r.tenantR.Spec.Username)
	}

	err = r.syncAdminUser(tenantID, adminUser)
	if err != nil {
		return err
	}

	newStatus := &apiv1alpha1.TenantStatus{
		AdminId:  *adminUser.Element.ID,
		TenantId: tenantID,
	}

	updated, err := r.reconcileStatus(newStatus)
	if err != nil {
		return err
	}

	// If updated - update the status and requeue
	if updated {
		return nil
	}

	return nil
}

// This method makes sure secret with tenant's access_token exists
func (r *TenantThreescaleReconciler) reconcileAccessTokenSecret(tenantDef *porta_client_pkg.Tenant) error {
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

	err = r.SetControllerOwnerReference(r.tenantR, desiredSecret)
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
func (r *TenantThreescaleReconciler) createTenant() (*porta_client_pkg.Tenant, error) {
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

func (r *TenantThreescaleReconciler) getAdminPassword() (string, error) {
	// Get tenant admin password from secret reference
	tenantAdminSecret := &v1.Secret{}

	err := r.Client().Get(context.TODO(), r.tenantR.AdminPassSecretKey(), tenantAdminSecret)

	if err != nil {
		return "", err
	}

	passwordByteArray, ok := tenantAdminSecret.Data[TenantAdminPasswordSecretField]
	if !ok {
		return "", &helper.WaitError{
			Err: fmt.Errorf("not found admin password secret (%s) - missing required attribute: %s",
				r.tenantR.AdminPassSecretKey(),
				TenantAdminPasswordSecretField),
		}
	}

	return bytes.NewBuffer(passwordByteArray).String(), err
}

func (r *TenantThreescaleReconciler) syncAdminUser(tenantID int64, adminUser *porta_client_pkg.DeveloperUser) error {
	// If adminUser desired state is not current state, update
	if adminUser.Element.State != nil && *adminUser.Element.State == "pending" {
		r.logger.Info("Activating pending admin user", "Account ID", tenantID, "ID", *adminUser.Element.ID)
		updatedAdminUser, err := r.portaClient.ActivateDeveloperUser(tenantID, *adminUser.Element.ID)
		if err != nil {
			return &helper.WaitError{
				Err: fmt.Errorf("3scale client failed activating developer user: %v", err),
			}
		}

		adminUser.Element.State = updatedAdminUser.Element.State
	}

	// If adminUser desired role is not current state, update
	if adminUser.Element.Role != nil && *adminUser.Element.Role != "admin" {
		r.logger.Info("Change role to admin", "Account ID", tenantID, "ID", *adminUser.Element.ID)
		updatedAdminUser, err := r.portaClient.ChangeRoleToAdminDeveloperUser(tenantID, *adminUser.Element.ID)
		if err != nil {
			return &helper.WaitError{
				Err: fmt.Errorf("3scale client failed changing the role to admin for developer user: %v", err),
			}
		}

		adminUser.Element.Role = updatedAdminUser.Element.Role
	}

	return nil
}

// Returns whether the status should be updated or not and the error
func (r *TenantThreescaleReconciler) reconcileStatus(desiredStatus *apiv1alpha1.TenantStatus) (bool, error) {
	if !reflect.DeepEqual(r.tenantR.Status, *desiredStatus) {
		diff := cmp.Diff(r.tenantR.Status, *desiredStatus)
		r.logger.V(1).Info(fmt.Sprintf("status has changed: %s", diff))
		r.tenantR.Status = *desiredStatus
		r.logger.Info("Update tenant status with tenantID", "tenantID", r.tenantR.Status.TenantId)
		return true, nil
	}
	return false, nil
}

// Returns tenant ID, tenant.Status.tenantID takes precedence over annotation value
func (r *TenantThreescaleReconciler) retrieveTenantID() (int64, error) {
	var tenantId int64 = 0

	// if the tenant.status.tenantID is 0, check if tenant.annotations.tenantID is present and use it instead
	if r.tenantR.Status.TenantId == 0 {
		// If the annotation tenantId is found, convert it to int64
		if value, found := r.tenantR.ObjectMeta.Annotations[tenantIdAnnotation]; found {
			tenantIdConvertedFromString, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, err
			}

			tenantId = tenantIdConvertedFromString
		}
	} else {
		tenantId = r.tenantR.Status.TenantId
	}

	return tenantId, nil
}
