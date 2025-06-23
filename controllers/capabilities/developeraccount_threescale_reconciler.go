package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

	// Reconciliation is based on the ID stored in the CR's annotation or .status block
	// This is required because none of the fields in the DeveloperAccount CR's .spec are unique
	// For instance, there may exist several DevAccounts with the same Organization Name.
	// We can tell if the account is already created in 3scale by checking the ID in CR's annotation or .status block
	accountId, err := s.retrieveAccountID()
	if err != nil {
		return nil, err
	}

	devAccount, err := s.findDevAccountByID(accountId)
	if err != nil {
		return nil, err
	}

	if devAccount == nil {
		s.logger.V(1).Info("DeveloperAccount does not exist", "OrgName", s.resource.Spec.OrgName)
		// The DeveloperAccount doesn't exist yet and must be created in 3scale
		createdDevAccount, createErr, devUserCR := s.createDevAccount()
		if createErr != nil {
			// Occasionally system will return a 409 error even though the DeveloperAccount was created successfully.
			// When this happens, the returned createdDevAccount will have nil values for all its fields.
			// If this happens the operator should be able to recover and not mark the DeveloperAccount CR as failed.
			// Otherwise, the operator will try to create the same DA again which will produce a real error.
			portaErr, ok := createErr.(threescaleapi.ApiErr)
			if ok && portaErr.Code() == 409 {
				// Try to find the DevAccount by the admin user's username
				fetchedDevAccount, fetchErr := s.findDevAccountByAdminUsername(devUserCR.Spec.Username)
				if fetchErr != nil || fetchedDevAccount == nil {
					// Account couldn't be found so requeue and retry creating it
					return devAccount, fetchErr
				}

				// Found the account in the DB so update the CR status with the account's ID and return the account
				s.resource.Status.ID = fetchedDevAccount.Element.ID
				return fetchedDevAccount, nil
			}
			return createdDevAccount, createErr
		}

		// Update the CR status with the account's ID and return the account
		s.resource.Status.ID = createdDevAccount.Element.ID
		return createdDevAccount, nil
	}

	s.logger.V(1).Info("DeveloperAccount already exists", "ID", *devAccount.Element.ID)

	// reconcile developer account
	return s.syncDeveloperAccount(devAccount)
}

// Returns account ID with developerAccount.Status.ID taking precedence over the annotation value
func (s *DeveloperAccountThreescaleReconciler) retrieveAccountID() (int64, error) {
	var accountId int64 = 0

	// If the developerAccount.Status.AccountID is nil, check if account.annotations.accountID is present and use it instead
	if s.resource.Status.ID == nil {
		if value, found := s.resource.ObjectMeta.Annotations[accountIdAnnotation]; found {
			// If the accountID annotation is found, convert it to int64
			accountIdConvertedFromString, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to convert accountID annotation value %s to int64", value)
			}

			accountId = accountIdConvertedFromString
		}
	} else {
		accountId = *s.resource.Status.ID
	}

	return accountId, nil
}

func (s *DeveloperAccountThreescaleReconciler) findDevAccountByID(accountID int64) (*threescaleapi.DeveloperAccount, error) {
	if accountID == 0 {
		return nil, nil
	}

	devAccount, err := s.threescaleAPIClient.DeveloperAccount(accountID)
	if err != nil && threescaleapi.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return devAccount, nil
}

func (s *DeveloperAccountThreescaleReconciler) findDevAccountByAdminUsername(adminUsername string) (*threescaleapi.DeveloperAccount, error) {
	if adminUsername == "" {
		return nil, nil
	}

	account, err := s.threescaleAPIClient.FindAccount(adminUsername)
	if err != nil && threescaleapi.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	devAccount, err := s.threescaleAPIClient.DeveloperAccount(account.ID)
	if err != nil && threescaleapi.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return devAccount, nil
}

func (s *DeveloperAccountThreescaleReconciler) createDevAccount() (*threescaleapi.DeveloperAccount, error, *capabilitiesv1beta1.DeveloperUser) {
	// 3scale API requires one developer user to be created as admin in the process of creating a developer account
	devAdminUserCR, err := s.findDeveloperAdminUserCR()
	if err != nil {
		return nil, err, nil
	}

	if devAdminUserCR == nil {
		// There should be one, wait for it
		return nil, &helper.WaitError{
			Err: errors.New("valid admin developer user CR not found"),
		}, nil
	}

	password, err := s.getAdminUserPassword(devAdminUserCR)
	if err != nil {
		return nil, err, nil
	}

	params := threescaleapi.Params{
		"org_name": s.resource.Spec.OrgName,
		"username": devAdminUserCR.Spec.Username,
		"email":    devAdminUserCR.Spec.Email,
		"password": password,
	}

	for k, v := range helper.ManagedByOperatorAnnotation() {
		params[k] = v
	}

	if s.resource.Spec.MonthlyBillingEnabled != nil {
		params["monthly_billing_enabled"] = strconv.FormatBool(*s.resource.Spec.MonthlyBillingEnabled)
	}

	if s.resource.Spec.MonthlyChargingEnabled != nil {
		params["monthly_charging_enabled"] = strconv.FormatBool(*s.resource.Spec.MonthlyChargingEnabled)
	}

	devAccountObj, signupErr := s.threescaleAPIClient.Signup(params)

	return devAccountObj, signupErr, devAdminUserCR
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

	// MonthlyBilling defaults to True
	desiredMonthlyBillingEnabled := true
	if s.resource.Spec.MonthlyBillingEnabled != nil {
		desiredMonthlyBillingEnabled = *s.resource.Spec.MonthlyBillingEnabled
	}
	if devAccount.Element.MonthlyBillingEnabled != nil && *devAccount.Element.MonthlyBillingEnabled != desiredMonthlyBillingEnabled {
		update = true
		deltaAccount.Element.MonthlyBillingEnabled = &desiredMonthlyBillingEnabled
	}

	// MonthlyChargingEnabled defaults to True
	desiredMonthlyChargingEnabled := true
	if s.resource.Spec.MonthlyChargingEnabled != nil {
		desiredMonthlyChargingEnabled = *s.resource.Spec.MonthlyChargingEnabled
	}
	if devAccount.Element.MonthlyChargingEnabled != nil && *devAccount.Element.MonthlyChargingEnabled != desiredMonthlyChargingEnabled {
		update = true
		deltaAccount.Element.MonthlyChargingEnabled = &desiredMonthlyChargingEnabled
	}

	if !helper.ManagedByOperatorAnnotationExists(devAccount.Element.Annotations) {
		for k, v := range helper.ManagedByOperatorDeveloperAccountAnnotation() {
			update = true
			if deltaAccount.Element.Annotations == nil {
				deltaAccount.Element.Annotations = make(map[string]string)
			}
			deltaAccount.Element.Annotations[k] = v
		}
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

func (s *DeveloperAccountThreescaleReconciler) getAdminUserPassword(adminUserCR *capabilitiesv1beta1.DeveloperUser) (string, error) {
	// Get password from secret reference
	secret := &corev1.Secret{}
	namespace := s.resource.Namespace
	if adminUserCR.Spec.PasswordCredentialsRef.Namespace != "" {
		namespace = adminUserCR.Spec.PasswordCredentialsRef.Namespace
	}

	err := s.Client().Get(s.Context(),
		types.NamespacedName{
			Name:      adminUserCR.Spec.PasswordCredentialsRef.Name,
			Namespace: namespace,
		},
		secret)
	if err != nil {
		return "", err
	}

	passwordByteArray, ok := secret.Data[capabilitiesv1beta1.DeveloperUserPasswordSecretField]
	if !ok {
		return "", fmt.Errorf("not found password field in secret (ns: %s, name: %s) field: %s",
			namespace, adminUserCR.Spec.PasswordCredentialsRef.Name, capabilitiesv1beta1.DeveloperUserPasswordSecretField)
	}

	return bytes.NewBuffer(passwordByteArray).String(), err
}
