package controllers

import (
	"errors"
	"reflect"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeveloperAccountThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.DeveloperAccount
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	providerAccountHost string
	logger              logr.Logger
}

func NewDeveloperAccountThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.DeveloperAccount, threescaleAPIClient *threescaleapi.ThreeScaleClient, providerAccountHost string, logger logr.Logger) *DeveloperAccountThreescaleReconciler {
	return &DeveloperAccountThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		providerAccountHost: providerAccountHost,
		logger:              logger.WithValues("3scale Reconciler", providerAccountHost),
	}
}

func (s *DeveloperAccountThreescaleReconciler) Reconcile() (*threescaleapi.DeveloperAccount, error) {
	s.logger.V(1).Info("START")

	// Reconciliation is based on ID stored in Status field
	// All fields of the spec are not unique
	// For instance, there may exist several DevAccounts with the same Organization Name.
	// The only way to know that the account is already created in 3scale is checking the ID in status.
	// Nice to Have would be having that ID in status inmutable using admission webhooks
	devAccount, err := s.findDevAccountByID()
	if err != nil {
		return nil, err
	}

	if devAccount == nil {
		s.logger.V(1).Info("DeveloperAccount does not exist", "OrgName", s.resource.Spec.OrgName)
		// ID not in status field
		// developer account has to be created in 3scale
		return s.createDevAccount()
	}

	s.logger.V(1).Info("DeveloperAccount already exists", "ID", *devAccount.Element.ID)

	// reconcile developer account
	return s.syncDeveloperAccount(devAccount)
}

func (s *DeveloperAccountThreescaleReconciler) findDevAccountByID() (*threescaleapi.DeveloperAccount, error) {
	if s.resource.Status.ID == nil {
		return nil, nil
	}

	devAccount, err := s.threescaleAPIClient.DeveloperAccount(*s.resource.Status.ID)
	if err != nil && threescaleapi.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return devAccount, nil
}

func (s *DeveloperAccountThreescaleReconciler) createDevAccount() (*threescaleapi.DeveloperAccount, error) {
	// 3scale API requires one developer user to be created as admin in the process of creating a developer account
	devAdminUserCR, err := s.findDeveloperAdminUserCR()
	if err != nil {
		return nil, err
	}

	if devAdminUserCR == nil {
		// There should be one, wait for it
		return nil, &helper.WaitError{
			Err: errors.New("Valid admin developer user CR not found"),
		}
	}

	params := threescaleapi.Params{
		"org_name": s.resource.Spec.OrgName,
		"username": devAdminUserCR.Spec.Username,
		"email":    devAdminUserCR.Spec.Email,
		"password": devAdminUserCR.Spec.Password,
	}

	if s.resource.Spec.MonthlyBillingEnabled != nil {
		params["monthly_billing_enabled"] = strconv.FormatBool(*s.resource.Spec.MonthlyBillingEnabled)
	}

	if s.resource.Spec.MonthlyChargingEnabled != nil {
		params["monthly_charging_enabled"] = strconv.FormatBool(*s.resource.Spec.MonthlyChargingEnabled)
	}

	return s.threescaleAPIClient.Signup(params)
}

func (s *DeveloperAccountThreescaleReconciler) findDeveloperAdminUserCR() (*capabilitiesv1beta1.DeveloperUser, error) {
	queryOpts := []client.ListOption{
		// query filter by namespace
		client.InNamespace(s.resource.Namespace),
	}

	// NICE TO HAVE: use client.MatchingFields in queryOpts. Requires declaring index field in the manager
	// https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html#setup
	// Filter by parent account
	parentAccountFilter := func(developerUser *capabilitiesv1beta1.DeveloperUser) (bool, error) {
		return developerUser.Spec.DeveloperAccountRef.Name == s.resource.Name, nil
	}

	// NICE TO HAVE: use client.MatchingFields in queryOpts. Requires declaring index field in the manager
	// https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html#setup
	// Filter by role
	adminRoleFilter := func(developerUser *capabilitiesv1beta1.DeveloperUser) (bool, error) {
		return developerUser.Spec.Role != nil && *developerUser.Spec.Role == "admin", nil
	}

	// Filter by orphan condition
	// the account create operation also creates the developer user,
	// so the search result must include only orphan objects
	orphanFilter := func(developerUser *capabilitiesv1beta1.DeveloperUser) (bool, error) {
		return developerUser.IsOrphan(), nil
	}

	devUserList, err := controllerhelper.FindDeveloperUserList(s.logger, s.Client(), queryOpts,
		parentAccountFilter,
		adminRoleFilter,
		// Filter by providerAccount
		controllerhelper.DeveloperUserProviderAccountFilter(s.Client(), s.resource.Namespace, s.providerAccountHost, s.logger),
		orphanFilter,
	)

	if err != nil {
		return nil, err
	}

	// take the first one, anyone should be valid.
	if len(devUserList) > 0 {
		return &devUserList[0], nil
	}

	return nil, nil
}

func (s *DeveloperAccountThreescaleReconciler) syncDeveloperAccount(devAccount *threescaleapi.DeveloperAccount) (*threescaleapi.DeveloperAccount, error) {
	update := false
	deltaAccount := &threescaleapi.DeveloperAccount{
		Element: threescaleapi.DeveloperAccountItem{
			ID: devAccount.Element.ID,
		},
	}

	if devAccount.Element.OrgName == nil || *devAccount.Element.OrgName != s.resource.Spec.OrgName {
		update = true
		deltaAccount.Element.OrgName = &s.resource.Spec.OrgName
	}

	if s.resource.Spec.MonthlyBillingEnabled != nil && !reflect.DeepEqual(devAccount.Element.MonthlyBillingEnabled, s.resource.Spec.MonthlyBillingEnabled) {
		update = true
		deltaAccount.Element.MonthlyBillingEnabled = s.resource.Spec.MonthlyBillingEnabled
	}

	if s.resource.Spec.MonthlyChargingEnabled != nil && !reflect.DeepEqual(devAccount.Element.MonthlyChargingEnabled, s.resource.Spec.MonthlyChargingEnabled) {
		update = true
		deltaAccount.Element.MonthlyChargingEnabled = s.resource.Spec.MonthlyChargingEnabled
	}

	updatedDevAccount := devAccount

	if update {
		updateRes, err := s.threescaleAPIClient.UpdateDeveloperAccount(deltaAccount)
		if err != nil {
			return nil, err
		}

		updatedDevAccount = updateRes
	}

	return updatedDevAccount, nil
}
