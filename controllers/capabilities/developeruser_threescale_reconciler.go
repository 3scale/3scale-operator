package controllers

import (
	"reflect"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
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

	// Reconciliation is based on ID stored in Status field
	// Nice to Have would be having that ID in status inmutable using admission webhooks
	devUser, err := s.findDevUserByID()
	if err != nil {
		return nil, err
	}

	if devUser == nil {
		s.logger.V(1).Info("DeveloperUser does not exist", "username", s.userCR.Spec.Username)
		// developer user has to be created in 3scale
		devUser, err = s.createDevUser()
		if err != nil {
			return nil, err
		}
	} else {
		s.logger.V(1).Info("DeveloperUser already exists", "ID", *devUser.Element.ID)
	}

	return s.syncDeveloperUser(devUser)
}

func (s *DeveloperUserThreescaleReconciler) checkParentAccount() error {
	if s.userCR.Status.AccountID != nil && !reflect.DeepEqual(s.userCR.Status.AccountID, s.parentAccountCR.Status.ID) {
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

func (s *DeveloperUserThreescaleReconciler) findDevUserByID() (*threescaleapi.DeveloperUser, error) {
	if s.userCR.Status.ID == nil {
		return nil, nil
	}

	devUser, err := s.threescaleAPIClient.DeveloperUser(*s.parentAccountCR.Status.ID, *s.userCR.Status.ID)
	if err != nil && threescaleapi.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return devUser, nil
}

func (s *DeveloperUserThreescaleReconciler) createDevUser() (*threescaleapi.DeveloperUser, error) {
	devUser := &threescaleapi.DeveloperUser{
		Element: threescaleapi.DeveloperUserItem{
			Username: &s.userCR.Spec.Username,
			Email:    &s.userCR.Spec.Email,
			Password: &s.userCR.Spec.Password,
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
