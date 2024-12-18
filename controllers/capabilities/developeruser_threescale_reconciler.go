package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type DeveloperUserThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	userCR              *capabilitiesv1beta1.DeveloperUser
	parentAccountCR     *capabilitiesv1beta1.DeveloperAccount
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	providerAccountHost string
	logger              logr.Logger
}

func NewDeveloperUserThreescaleReconciler(b *reconcilers.BaseReconciler,
	userCR *capabilitiesv1beta1.DeveloperUser,
	parentAccountCR *capabilitiesv1beta1.DeveloperAccount,
	threescaleAPIClient *threescaleapi.ThreeScaleClient,
	providerAccountHost string,
	logger logr.Logger,
) *DeveloperUserThreescaleReconciler {
	return &DeveloperUserThreescaleReconciler{
		BaseReconciler:      b,
		userCR:              userCR,
		parentAccountCR:     parentAccountCR,
		threescaleAPIClient: threescaleAPIClient,
		providerAccountHost: providerAccountHost,
		logger:              logger.WithValues("3scale Reconciler", providerAccountHost),
	}
}

func (s *DeveloperUserThreescaleReconciler) Reconcile() (*threescaleapi.DeveloperUser, error) {
	s.logger.V(1).Info("START")

	err := s.checkParentAccount()
	if err != nil {
		return nil, err
	}

	devUser, err := s.findDevUser()
	if err != nil {
		return nil, err
	}

	if devUser == nil {
		s.logger.V(1).Info("DeveloperUser does not exist", "username", s.userCR.Spec.Username)
		// The DeveloperAccount doesn't exist yet and must be created in 3scale
		devUser, err = s.createDevUser()
		if err != nil {
			return nil, err
		}

		// Update the CR status with the account's ID and return the account
		s.userCR.Status.ID = devUser.Element.ID
		return devUser, nil
	} else {
		s.logger.V(1).Info("DeveloperUser already exists", "ID", *devUser.Element.ID)
	}

	return s.syncDeveloperUser(devUser)
}

func (s *DeveloperUserThreescaleReconciler) checkParentAccount() error {
	if s.userCR.Status.AccountID != nil &&
		!reflect.DeepEqual(s.userCR.Status.AccountID, s.parentAccountCR.Status.ID) &&
		s.userCR.Status.ID != nil {
		// Account ID from referenced CR does not much with status Account ID
		// The referenced account might have changed.
		// Since usernames and emails are unique, it needs to be removed first from old developer account
		err := s.threescaleAPIClient.DeleteDeveloperUser(*s.userCR.Status.AccountID, *s.userCR.Status.ID)
		if err != nil && !threescaleapi.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (s *DeveloperUserThreescaleReconciler) findDevUser() (*threescaleapi.DeveloperUser, error) {
	// Reconciliation is based on the ID stored in the CR's annotation or .status block
	// We can tell if the user is already created in 3scale by checking the ID in CR's annotation or .status block
	userId, err := s.retrieveUserID()
	if err != nil {
		return nil, err
	}

	devUser, err := s.findDevUserByID(userId)
	if err != nil {
		return nil, err
	}

	if devUser != nil {
		return devUser, nil
	}

	// If not found by ID, try {username, email} set.
	// Both fields are unique in the provider account scope.
	return s.findDevUserByUsernameAndEmail()
}

func (s *DeveloperUserThreescaleReconciler) findDevUserByUsernameAndEmail() (*threescaleapi.DeveloperUser, error) {
	devUserList, err := s.threescaleAPIClient.ListDeveloperUsers(*s.parentAccountCR.Status.ID, nil)
	if err != nil {
		return nil, err
	}

	for idx := range devUserList.Items {
		if devUserList.Items[idx].Element.Username != nil && *devUserList.Items[idx].Element.Username == s.userCR.Spec.Username &&
			devUserList.Items[idx].Element.Email != nil && *devUserList.Items[idx].Element.Email == s.userCR.Spec.Email {
			return &devUserList.Items[idx], nil
		}
	}

	return nil, nil
}

// Returns user ID with developerUser.Status.ID taking precedence over the annotation value
func (s *DeveloperUserThreescaleReconciler) retrieveUserID() (int64, error) {
	var userId int64 = 0

	// If the developerUser.Status.ID is nil, check if user.annotations.userID is present and use it instead
	if s.userCR.Status.ID == nil {
		if value, found := s.userCR.ObjectMeta.Annotations[userIdAnnotation]; found {
			// If the userID annotation is found, convert it to int64
			userIdConvertedFromString, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, errors.New(fmt.Sprintf("failed to convert userID annotation value %s to int64", value))
			}

			userId = userIdConvertedFromString
		}
	} else {
		userId = *s.userCR.Status.ID
	}

	return userId, nil
}

func (s *DeveloperUserThreescaleReconciler) findDevUserByID(userID int64) (*threescaleapi.DeveloperUser, error) {
	if userID == 0 {
		return nil, nil
	}

	devUser, err := s.threescaleAPIClient.DeveloperUser(*s.parentAccountCR.Status.ID, userID)
	if err != nil && threescaleapi.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return devUser, nil
}

func (s *DeveloperUserThreescaleReconciler) createDevUser() (*threescaleapi.DeveloperUser, error) {
	password, err := s.getPassword()
	if err != nil {
		return nil, err
	}

	devUser := &threescaleapi.DeveloperUser{
		Element: threescaleapi.DeveloperUserItem{
			Username: &s.userCR.Spec.Username,
			Email:    &s.userCR.Spec.Email,
			Password: &password,
		},
	}

	if s.userCR.Spec.Role != nil {
		devUser.Element.Role = s.userCR.Spec.Role
	}

	return s.threescaleAPIClient.CreateDeveloperUser(*s.parentAccountCR.Status.ID, devUser)
}

func (s *DeveloperUserThreescaleReconciler) syncDeveloperUser(devUser *threescaleapi.DeveloperUser) (*threescaleapi.DeveloperUser, error) {
	update := false

	deltaUser := &threescaleapi.DeveloperUser{
		Element: threescaleapi.DeveloperUserItem{
			ID: devUser.Element.ID,
		},
	}

	if devUser.Element.Email == nil || *devUser.Element.Email != s.userCR.Spec.Email {
		update = true
		deltaUser.Element.Email = &s.userCR.Spec.Email
	}

	if devUser.Element.Username == nil || *devUser.Element.Username != s.userCR.Spec.Username {
		update = true
		deltaUser.Element.Username = &s.userCR.Spec.Username
	}

	// TODO Password reconciliation? maybe when read from secret?

	updatedDevUser := devUser

	if update {
		updateRes, err := s.threescaleAPIClient.UpdateDeveloperUser(*s.parentAccountCR.Status.ID, deltaUser)
		if err != nil {
			return nil, err
		}

		updatedDevUser = updateRes
	}

	if updatedDevUser.Element.State != nil && *updatedDevUser.Element.State == "pending" {
		updateRes, err := s.threescaleAPIClient.ActivateDeveloperUser(*s.parentAccountCR.Status.ID, *updatedDevUser.Element.ID)
		if err != nil {
			return nil, err
		}

		updatedDevUser = updateRes
	}

	if updatedDevUser.Element.State != nil && *updatedDevUser.Element.State == "suspended" &&
		!s.userCR.Spec.Suspended {
		updateRes, err := s.threescaleAPIClient.UnsuspendDeveloperUser(*s.parentAccountCR.Status.ID, *updatedDevUser.Element.ID)
		if err != nil {
			return nil, err
		}

		updatedDevUser = updateRes
	}

	if updatedDevUser.Element.State != nil && *updatedDevUser.Element.State == "active" &&
		s.userCR.Spec.Suspended {
		updateRes, err := s.threescaleAPIClient.SuspendDeveloperUser(*s.parentAccountCR.Status.ID, *updatedDevUser.Element.ID)
		if err != nil {
			return nil, err
		}

		updatedDevUser = updateRes
	}

	if updatedDevUser.Element.Role != nil && *updatedDevUser.Element.Role == "member" &&
		s.userCR.IsAdmin() {
		updateRes, err := s.threescaleAPIClient.ChangeRoleToAdminDeveloperUser(*s.parentAccountCR.Status.ID, *updatedDevUser.Element.ID)
		if err != nil {
			return nil, err
		}

		updatedDevUser = updateRes
	}

	if updatedDevUser.Element.Role != nil && *updatedDevUser.Element.Role == "admin" &&
		!s.userCR.IsAdmin() {
		updateRes, err := s.threescaleAPIClient.ChangeRoleToMemberDeveloperUser(*s.parentAccountCR.Status.ID, *updatedDevUser.Element.ID)
		if err != nil {
			return nil, err
		}

		updatedDevUser = updateRes
	}

	return updatedDevUser, nil
}

func (s *DeveloperUserThreescaleReconciler) getPassword() (string, error) {
	passwdFieldPath := field.NewPath("spec").Child("passwordCredentialsRef")

	// Get password from secret reference
	secret := &corev1.Secret{}
	namespace := s.userCR.Namespace
	if s.userCR.Spec.PasswordCredentialsRef.Namespace != "" {
		namespace = s.userCR.Spec.PasswordCredentialsRef.Namespace
	}

	err := s.Client().Get(s.Context(),
		types.NamespacedName{
			Name:      s.userCR.Spec.PasswordCredentialsRef.Name,
			Namespace: namespace,
		},
		secret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Return spec field error if secret was not found
			return "", &helper.SpecFieldError{
				ErrorType: helper.InvalidError,
				FieldErrorList: field.ErrorList{
					field.Invalid(passwdFieldPath, s.userCR.Spec.PasswordCredentialsRef, "developeruser password reference not found"),
				},
			}
		}

		return "", err
	}

	passwordByteArray, ok := secret.Data[capabilitiesv1beta1.DeveloperUserPasswordSecretField]
	if !ok {
		// Return spec field error if secret field was not found
		return "", &helper.SpecFieldError{
			ErrorType: helper.InvalidError,
			FieldErrorList: field.ErrorList{
				field.Invalid(passwdFieldPath, s.userCR.Spec.PasswordCredentialsRef, "developeruser password secret missing expected field"),
			},
		}
	}

	return bytes.NewBuffer(passwordByteArray).String(), err
}
