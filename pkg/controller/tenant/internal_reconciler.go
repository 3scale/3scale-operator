package tenant

import (
	"bytes"
	"context"
	"fmt"

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

	err = r.reconcileAdminUser(tenantDef)
	if err != nil {
		return err
	}

	return r.reconcileAccessTokenSecret(tenantDef)
}

// This method makes sure that tenant exists, otherwise it will create one
// On method completion:
// * tenant will exist with specified admin user
// * tenant's access_token will be available
func (r *InternalReconciler) reconcileTenant() (*porta_client_pkg.Signup, error) {
	tenantDef, err := r.findTenant()
	if err != nil {
		return nil, err
	}

	if tenantDef == nil {
		tenantDef, err = r.createTenant()
	} else {
		// If tenant is found, access token owned by that tenant admin is not provided
		r.logger.Info("Tenant already exists", "TenantID", tenantDef.Account.ID)
	}

	return tenantDef, err
}

////
//
// This method makes sure tenant admin user is active
// On method completion, tenant admin user will be active
func (r *InternalReconciler) reconcileAdminUser(tenantDef *porta_client_pkg.Signup) error {
	adminUser := r.findAdminUser(tenantDef)

	if adminUser == nil {
		return fmt.Errorf("Admin user not found and should be available"+
			"TenandId: %s. Admin Username: %s, Admin email: %s", tenantDef.Account.ID,
			r.tenantR.Spec.UserName, r.tenantR.Spec.Email)
	}

	// activate user operation just works when user is in pending state
	// activate when suspended?
	if adminUser.State == "pending" {
		err := r.activateAdminUser(adminUser)
		if err != nil {
			return err
		}
	} else {
		r.logger.Info("Admin user already active", "UserID", adminUser.ID)
	}

	return nil
}

// This method makes sure secret with tenant's access_token exists
func (r *InternalReconciler) reconcileAccessTokenSecret(tenantDef *porta_client_pkg.Signup) error {
	//r.logger.Info("Creating a new Secret", "Secret.Namespace", pod.Namespace, "Secret.Name", pod.Name)
	// if AccessToken is not available, update Tenant Resource Status
	// AccessToken might not be available if tenant was not created in the pipeline
	// because it already existed. In that case, token is not available.
	// Currently, there is no way to create tenant's user admin's access token using master account access_token.
	adminAccessTokenSecretNN := types.NamespacedName{
		Name:      adminAccessTokenName,
		Namespace: r.tenantR.Spec.DestinationNS,
	}
	adminAccessTokenSecret, err := r.findAccessTokenSecret(adminAccessTokenSecretNN)
	if err != nil {
		return err
	}

	if adminAccessTokenSecret == nil {
		if tenantDef.AccessToken.Value != "" {
			err = r.createAdminAcessTokenSecret(tenantDef, adminAccessTokenSecretNN)
			if err != nil {
				return err
			}
		} else {
			return NewAccessTokenNotAvailableError("Could not create admin access token secret",
				tenantDef.Account.ID, adminAccessTokenSecretNN)
		}
	} else {
		r.logger.Info("Admin user access token secret already exists",
			"Secret NS", adminAccessTokenSecretNN.Namespace, "Secret name", adminAccessTokenSecretNN.Name)
	}
	return nil
}

// Create Tenant using porta client
func (r *InternalReconciler) createTenant() (*porta_client_pkg.Signup, error) {
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

// Find Tenant in account list
// identity function is based on:
// * Organization Name
// * Email
// * Username
func (r *InternalReconciler) findTenant() (*porta_client_pkg.Signup, error) {
	accountList, err := r.portaClient.ListAccounts(r.masterAccessToken)
	if err != nil {
		return nil, err
	}

	var tenantDef *porta_client_pkg.Signup

	// flat map operation to search current tenant in account list
	// internal data type to build flat list
	type TenantInfo struct {
		Idx      int
		OrgName  string
		Email    string
		UserName string
	}
	tenantInfoList := []*TenantInfo{}
	for accountIdx, account := range accountList.Accounts {
		for _, user := range account.Users.Users {
			tenantInfoList = append(tenantInfoList, &TenantInfo{
				Idx:      accountIdx,
				OrgName:  account.OrgName,
				Email:    account.SupportEmail,
				UserName: user.UserName,
			})
		}
	}

	for _, tenantInfo := range tenantInfoList {
		if tenantInfo.OrgName == r.tenantR.Spec.OrgName &&
			tenantInfo.Email == r.tenantR.Spec.Email &&
			tenantInfo.UserName == r.tenantR.Spec.UserName {

			// There is no AccessToken info available
			tenantDef = &porta_client_pkg.Signup{
				Account: accountList.Accounts[tenantInfo.Idx],
			}
		}
	}
	return tenantDef, nil
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

func (r *InternalReconciler) findAdminUser(tenantDef *porta_client_pkg.Signup) *porta_client_pkg.User {
	// Current implementation only searchs on tenant info (account information).
	// Should be there because previous step make sure tenant is created or already exists
	// with the given admin user
	// An alternate implementation would be fetch users state from 3scale API
	for _, user := range tenantDef.Account.Users.Users {
		if user.Email == r.tenantR.Spec.Email && user.UserName == r.tenantR.Spec.UserName {
			// user is already a copy from User slice element
			return &user
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

func (r *InternalReconciler) createAdminAcessTokenSecret(tenantDef *porta_client_pkg.Signup, nn types.NamespacedName) error {
	r.logger.Info("Creating admin access token secret", "Secret NS", nn.Namespace, "Secret name", nn.Name)

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
		StringData: map[string]string{secretAdminAccessTokenKey: tenantDef.AccessToken.Value},
		Type:       v1.SecretTypeOpaque,
	}
	return r.k8sClient.Create(context.TODO(), secret)
}
