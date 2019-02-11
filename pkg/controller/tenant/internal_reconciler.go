package tenant

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InternalReconciler reconciles a Tenant object
type InternalReconciler struct {
	k8sClient         client.Client
	tenantR           *apiv1alpha1.Tenant
	portaClient       *porta_client_pkg.ThreeScaleClient
	masterAccessToken string
	logger            logr.Logger
}

// NewInternalReconciler constructs InternalReconciler object
func NewInternalReconciler(k8sClient client.Client, tenantR *apiv1alpha1.Tenant,
	portaClient *porta_client_pkg.ThreeScaleClient, masterAccessToken string,
	log logr.Logger) *InternalReconciler {
	return &InternalReconciler{
		k8sClient:         k8sClient,
		tenantR:           tenantR,
		portaClient:       portaClient,
		masterAccessToken: masterAccessToken,
		logger:            log,
	}
}

// Run tenant reconciliation logic
// Facts to reconcile:
// - Have 3scale Tenant Account
// - Have active admin user
// - Have secret with tenant's access_token
func (r *InternalReconciler) Run() error {
	tenantDef, err := r.reconcileTenant()
	if err != nil {
		return err
	}

	adminUserDef, err := r.reconcileAdminUser(tenantDef)
	if err != nil {
		return err
	}

	err = r.reconcileAccessTokenSecret(tenantDef)
	if err != nil {
		return err
	}

	tenantStatus := r.getTenantStatus(tenantDef, adminUserDef)

	return r.updateTenantStatus(tenantStatus)
}

// This method makes sure that tenant exists, otherwise it will create one
// On method completion:
// * tenant will exist
// * tenant's attributes will be updated if required
func (r *InternalReconciler) reconcileTenant() (*porta_client_pkg.Tenant, error) {
	tenantDef, err := r.fetchTenant()
	if err != nil {
		return nil, err
	}

	if tenantDef == nil {
		tenantDef, err = r.createTenant()
		if err != nil {
			return nil, err
		}
	} else {
		r.logger.Info("Tenant already exists", "TenantID", tenantDef.Signup.Account.ID)
		// Tenant is not created, check tenant desired state matches current state
		// When created, not needed to update
		err := r.syncTenant(tenantDef)
		if err != nil {
			return nil, err
		}
	}

	return tenantDef, nil
}

func (r *InternalReconciler) fetchTenant() (*porta_client_pkg.Tenant, error) {
	if r.tenantR.Status.TenantID == 0 {
		// tenantID not in status field
		// Tenant has to be created
		return nil, nil
	}

	tenantDef, err := r.portaClient.ShowTenant(r.masterAccessToken, r.tenantR.Status.TenantID)
	if err != nil && porta_client_pkg.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return tenantDef, nil
}

func (r *InternalReconciler) syncTenant(tenantDef *porta_client_pkg.Tenant) error {
	// If tenant desired state is not current state, update
	triggerSync := func() bool {
		if r.tenantR.Spec.OrgName != tenantDef.Signup.Account.OrgName {
			return true
		}

		if r.tenantR.Spec.Email != tenantDef.Signup.Account.SupportEmail {
			return true
		}

		return false
	}()

	if triggerSync {
		r.logger.Info("Syncing tenant", "TenantID", tenantDef.Signup.Account.ID)
		tenantDef.Signup.Account.OrgName = r.tenantR.Spec.OrgName
		tenantDef.Signup.Account.SupportEmail = r.tenantR.Spec.Email
		params := porta_client_pkg.Params{
			"support_email": r.tenantR.Spec.Email,
			"org_name":      r.tenantR.Spec.OrgName,
		}
		_, err := r.portaClient.UpdateTenant(r.masterAccessToken, r.tenantR.Status.TenantID, params)
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
func (r *InternalReconciler) reconcileAdminUser(tenantDef *porta_client_pkg.Tenant) (*porta_client_pkg.User, error) {
	adminUserDef, err := r.fetchAdminUser(tenantDef)
	if err != nil {
		return nil, err
	}

	err = r.syncAdminUser(tenantDef, adminUserDef)
	if err != nil {
		return nil, err
	}

	return adminUserDef, nil
}

// This method makes sure secret with tenant's access_token exists
func (r *InternalReconciler) reconcileAccessTokenSecret(tenantDef *porta_client_pkg.Tenant) error {
	adminAccessTokenSecretNN := types.NamespacedName{
		Name:      adminAccessTokenName,
		Namespace: r.tenantR.Spec.DestinationNS,
	}
	adminAccessTokenSecret, err := r.findAccessTokenSecret(adminAccessTokenSecretNN)
	if err != nil {
		return err
	}

	if adminAccessTokenSecret == nil {
		err = r.createAdminAcessTokenSecret(tenantDef, adminAccessTokenSecretNN)
		if err != nil {
			return err
		}
	} else {
		r.logger.Info("Admin user access token secret already exists",
			"Secret NS", adminAccessTokenSecretNN.Namespace, "Secret name", adminAccessTokenSecretNN.Name)
	}
	return nil
}

// Create Tenant using porta client
func (r *InternalReconciler) createTenant() (*porta_client_pkg.Tenant, error) {
	password, err := r.getAdminPassword()
	if err != nil {
		return nil, err
	}

	r.logger.Info("Creating a new tenant", "OrgName", r.tenantR.Spec.OrgName,
		"UserName", r.tenantR.Spec.UserName, "Email", r.tenantR.Spec.Email)
	return r.portaClient.CreateTenant(
		r.masterAccessToken,
		r.tenantR.Spec.OrgName,
		r.tenantR.Spec.UserName,
		r.tenantR.Spec.Email,
		password,
	)
}

func (r *InternalReconciler) getAdminPassword() (string, error) {
	// Get tenant admin password from secret reference
	tenantAdminSecret := &v1.Secret{}

	err := r.k8sClient.Get(context.TODO(),
		types.NamespacedName{
			Name:      r.tenantR.Spec.Password.Name,
			Namespace: r.tenantR.Spec.Password.Namespace,
		},
		tenantAdminSecret)

	if err != nil {
		return "", err
	}

	passwordByteArray, ok := tenantAdminSecret.Data[secretMasterAdminPasswordKey]
	if !ok {
		return "", fmt.Errorf("Not found admin password secret (ns: %s, name: %s) attribute: %s",
			r.tenantR.Spec.Password.Namespace, r.tenantR.Spec.Password.Name, secretMasterAdminPasswordKey)
	}

	return bytes.NewBuffer(passwordByteArray).String(), err
}

//
func (r *InternalReconciler) fetchAdminUser(tenantDef *porta_client_pkg.Tenant) (*porta_client_pkg.User, error) {
	if r.tenantR.Status.AdminUserID == 0 {
		// UserID not in status field
		return r.findAdminUser(tenantDef)
	}

	//
	return r.portaClient.ReadUser(r.masterAccessToken, tenantDef.Signup.Account.ID, r.tenantR.Status.AdminUserID)
}

func (r *InternalReconciler) findAdminUser(tenantDef *porta_client_pkg.Tenant) (*porta_client_pkg.User, error) {
	// Only admin users
	// Any state
	filterParams := porta_client_pkg.Params{
		"role": "admin",
	}
	userList, err := r.portaClient.ListUsers(r.masterAccessToken, tenantDef.Signup.Account.ID, filterParams)
	if err != nil {
		return nil, err
	}

	for _, user := range userList.Users {
		if user.User.Email == r.tenantR.Spec.Email && user.User.UserName == r.tenantR.Spec.UserName {
			// user is already a copy from User slice element
			return &user.User, nil
		}
	}
	return nil, fmt.Errorf("Admin user not found and should be available"+
		"TenantID: %d. Admin Username: %s, Admin email: %s", tenantDef.Signup.Account.ID,
		r.tenantR.Spec.UserName, r.tenantR.Spec.Email)
}
func (r *InternalReconciler) syncAdminUser(tenantDef *porta_client_pkg.Tenant, adminUser *porta_client_pkg.User) error {
	// If adminUser desired state is not current state, update
	if adminUser.State == "pending" {
		err := r.activateAdminUser(adminUser)
		if err != nil {
			return err
		}
	} else {
		r.logger.Info("Admin user already active", "TenantID", tenantDef.Signup.Account.ID, "UserID", adminUser.ID)
	}

	triggerSync := func() bool {
		if r.tenantR.Spec.UserName != adminUser.UserName {
			return true
		}

		if r.tenantR.Spec.Email != adminUser.Email {
			return true
		}

		return false
	}()

	if triggerSync {
		r.logger.Info("Syncing adminUser", "TenantID", tenantDef.Signup.Account.ID, "UserID", adminUser.ID)
		adminUser.UserName = r.tenantR.Spec.UserName
		adminUser.Email = r.tenantR.Spec.Email
		params := porta_client_pkg.Params{
			"username": r.tenantR.Spec.UserName,
			"email":    r.tenantR.Spec.Email,
		}
		_, err := r.portaClient.UpdateUser(r.masterAccessToken, tenantDef.Signup.Account.ID, adminUser.ID, params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *InternalReconciler) activateAdminUser(adminUser *porta_client_pkg.User) error {
	r.logger.Info("Activating pending admin user", "AccountID", adminUser.AccountID, "ID", adminUser.ID)
	return r.portaClient.ActivateUser(r.masterAccessToken, adminUser.AccountID, adminUser.ID)
}

func (r *InternalReconciler) findAccessTokenSecret(nn types.NamespacedName) (*v1.Secret, error) {
	adminAccessTokenSecret := &v1.Secret{}

	err := r.k8sClient.Get(context.TODO(), nn, adminAccessTokenSecret)

	if err != nil && errors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return adminAccessTokenSecret, nil
}

func (r *InternalReconciler) createAdminAcessTokenSecret(tenantDef *porta_client_pkg.Tenant, nn types.NamespacedName) error {
	r.logger.Info("Creating admin access token secret", "Secret NS", nn.Namespace, "Secret name", nn.Name)

	tenantProviderKey, err := r.findTenantProviderKey(tenantDef)
	if err != nil {
		return err
	}

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nn.Namespace,
			Name:      nn.Name,
			Labels:    map[string]string{"app": "3scale-operator"},
		},
		StringData: map[string]string{secretAdminAccessTokenKey: tenantProviderKey},
		Type:       v1.SecretTypeOpaque,
	}
	addOwnerRefToObject(secret, asOwner(r.tenantR))
	return r.k8sClient.Create(context.TODO(), secret)
}

func (r *InternalReconciler) findTenantProviderKey(tenantDef *porta_client_pkg.Tenant) (string, error) {
	// Tenant Provider Key is available on provider application list
	appList, err := r.portaClient.ListApplications(r.masterAccessToken, tenantDef.Signup.Account.ID)
	if err != nil {
		return "", err
	}

	if len(appList.Applications) != 1 {
		return "", fmt.Errorf("Unexpected application list. TenantID: %d", tenantDef.Signup.Account.ID)
	}

	return appList.Applications[0].Application.UserKey, nil
}

func (r *InternalReconciler) getTenantStatus(tenantDef *porta_client_pkg.Tenant, adminUserDef *porta_client_pkg.User) *apiv1alpha1.TenantStatus {
	return &apiv1alpha1.TenantStatus{
		TenantID:    tenantDef.Signup.Account.ID,
		AdminUserID: adminUserDef.ID,
	}
}

func (r *InternalReconciler) updateTenantStatus(tenantStatus *apiv1alpha1.TenantStatus) error {
	// don't update the status if there aren't any changes.
	if reflect.DeepEqual(r.tenantR.Status, *tenantStatus) {
		return nil
	}
	r.tenantR.Status = *tenantStatus
	return r.k8sClient.Status().Update(context.TODO(), r.tenantR)
}
