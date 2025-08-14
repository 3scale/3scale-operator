/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	rand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplicationAuthReconciler reconciles a ApplicationAuth object
type ApplicationAuthReconciler struct {
	*reconcilers.BaseReconciler
}

type AuthSecret struct {
	UserKey        string
	ApplicationKey string
	ApplicationID  string
}

const (
	UserKey        = "UserKey"
	ApplicationKey = "ApplicationKey"
	ApplicationID  = "ApplicationID"
)

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=applicationauths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=applicationauths/status,verbs=get;update;patch

func (r *ApplicationAuthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("applicationauth", req.NamespacedName)
	reqLogger.Info("Reconcile Application Authentication", "Operator version", version.Version)

	applicationAuth := &capabilitiesv1beta1.ApplicationAuth{}

	err := r.Client().Get(r.Context(), req.NamespacedName, applicationAuth)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(applicationAuth, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	if !applicationAuth.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType) {
		// Retrieve application CR, on failed retrieval update status and requeue
		application := &capabilitiesv1beta1.Application{}
		err = r.Client().Get(r.Context(), types.NamespacedName{Name: applicationAuth.Spec.ApplicationCRName, Namespace: applicationAuth.Namespace}, application)
		if err != nil {
			// If the product CR is not found, update status and requeue
			if errors.IsNotFound(err) {
				reqLogger.Info("Application CR not found. Ignoring since object must have been deleted")
				return r.reconcileStatus(applicationAuth, err, reqLogger)
			}

			// If API call error, return err
			return ctrl.Result{}, err
		}

		// Make sure application is ready
		err = checkApplicationResources(applicationAuth, application)
		if err != nil {
			return r.reconcileStatus(applicationAuth, err, reqLogger)
		}

		// Retrieve DeveloperAccount CR, on failed retrieval update status and requeue
		developerAccount := &capabilitiesv1beta1.DeveloperAccount{}
		err = r.Client().Get(r.Context(), types.NamespacedName{Name: application.Spec.AccountCR.Name, Namespace: applicationAuth.Namespace}, developerAccount)
		if err != nil {
			// If the product CR is not found, update status and requeue
			if errors.IsNotFound(err) {
				reqLogger.Info("DeveloperAccount CR not found. Ignoring since object must have been deleted")
				return r.reconcileStatus(applicationAuth, err, reqLogger)
			}

			// If API call error, return err
			return ctrl.Result{}, err
		}

		// Retrieve Product CR, on failed retrieval update status and requeue
		product := &capabilitiesv1beta1.Product{}
		err = r.Client().Get(r.Context(), types.NamespacedName{Name: application.Spec.ProductCR.Name, Namespace: applicationAuth.Namespace}, product)
		if err != nil {
			// If the product CR is not found, update status and requeue
			if errors.IsNotFound(err) {
				reqLogger.Info("Product CR not found. Ignoring since object must have been deleted")
				return r.reconcileStatus(applicationAuth, err, reqLogger)
			}

			// If API call error, return err
			return ctrl.Result{}, err
		}

		authMode := product.Spec.AuthenticationMode()
		if authMode == nil {
			err := fmt.Errorf("unable to identify authentication mode from Product CR")
			return r.reconcileStatus(applicationAuth, err, reqLogger)
		}

		// Retrieve providerAccountRef
		providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), applicationAuth.GetNamespace(), applicationAuth.Spec.ProviderAccountRef, r.Logger())
		if err != nil {
			return ctrl.Result{}, err
		}

		// connect to the 3scale porta client
		insecureSkipVerify := controllerhelper.GetInsecureSkipVerifyAnnotation(applicationAuth.GetAnnotations())
		threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount, insecureSkipVerify)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Retrieve auth secret, on failed retrieval update status and requeue
		authSecretObj := &corev1.Secret{}
		err = r.Client().Get(r.Context(), types.NamespacedName{Name: applicationAuth.Spec.AuthSecretRef.Name, Namespace: applicationAuth.Namespace}, authSecretObj)
		if err != nil {
			// If the product CR is not found, update status and requeue
			if errors.IsNotFound(err) {
				reqLogger.Info("ApplicationAuth secret not found. Ignoring since object must have been deleted")
				return r.reconcileStatus(applicationAuth, err, reqLogger)
			}
			return ctrl.Result{}, err
		}

		controller, err := GetAuthController(*authMode, reqLogger)
		if err != nil {
			return ctrl.Result{}, err
		}

		// populate authSecret struct and make sure required fields are available
		shouldGenerateSecret := applicationAuth.Spec.GenerateSecret != nil && *applicationAuth.Spec.GenerateSecret
		reqLogger.Info("LookupAuthSecret", "ns", applicationAuth.Namespace, "authSecretRef", applicationAuth.Spec.AuthSecretRef)
		authSecret, err := controller.SecretReferenceSource(r.Client(), applicationAuth.Namespace, applicationAuth.Spec.AuthSecretRef, shouldGenerateSecret)
		if err != nil {
			return r.reconcileStatus(applicationAuth, err, reqLogger)
		}

		err = controller.Sync(threescaleAPIClient, *developerAccount.Status.ID, *application.Status.ID, *authSecret)
		if err != nil {
			return r.reconcileStatus(applicationAuth, err, reqLogger)
		}
	}
	// final return
	reqLogger.Info("Successfully reconciled")
	return r.reconcileStatus(applicationAuth, nil, reqLogger)
}

func (r *ApplicationAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ApplicationAuth{}).
		Complete(r)
}

type AuthController interface {
	Sync(threescaleClient *threescaleapi.ThreeScaleClient, developerAccountID int64, applicationID int64, authSecret AuthSecret) error
	SecretReferenceSource(cl client.Client, ns string, authSectretRef *corev1.LocalObjectReference, generateSecret bool) (*AuthSecret, error)
}

func GetAuthController(mode string, logger logr.Logger) (AuthController, error) {
	switch mode {
	case "1":
		return &userKeyAuthMode{logger: logger}, nil
	case "2":
		return &appIDAuthMode{logger: logger}, nil
	default:
		return nil, fmt.Errorf("unknown authentication mode")
	}
}

type userKeyAuthMode struct {
	logger logr.Logger
}

func (u *userKeyAuthMode) Sync(threescaleClient *threescaleapi.ThreeScaleClient, developerAccountID int64, applicationID int64, authSecret AuthSecret) error {
	// get the existing value from the porta
	existingApplication, err := threescaleClient.Application(developerAccountID, applicationID)
	if err != nil {
		return err
	}
	existingKey := existingApplication.UserKey

	// user_key mismatch, update
	if existingKey != authSecret.UserKey {
		params := make(map[string]string)
		params["user_key"] = authSecret.UserKey
		if _, err := threescaleClient.UpdateApplication(developerAccountID, applicationID, params); err != nil {
			return err
		}
	}
	return nil
}

func (u *userKeyAuthMode) SecretReferenceSource(cl client.Client, ns string, authSectretRef *corev1.LocalObjectReference, generateSecret bool) (*AuthSecret, error) {
	secretSource := helper.NewSecretSource(cl, ns)
	userKeyStr, err := secretSource.RequiredFieldValueFromRequiredSecret(authSectretRef.Name, UserKey)
	if err != nil {
		return nil, err
	}

	if userKeyStr == "" {
		if generateSecret {
			userKeyStr = rand.String(16)

			newValues := map[string][]byte{
				UserKey: []byte(userKeyStr),
			}

			if err := updateSecret(context.Background(), cl, authSectretRef.Name, ns, newValues); err != nil {
				return nil, err
			}
		} else {
			// Nothing available raise error now
			return nil, fmt.Errorf("no UserKey available in secret and generate secret is set to false")
		}
	}
	return &AuthSecret{UserKey: userKeyStr}, nil
}

type appIDAuthMode struct {
	logger logr.Logger
}

func (a *appIDAuthMode) Sync(threescaleClient *threescaleapi.ThreeScaleClient, developerAccountID int64, applicationID int64, authSecret AuthSecret) error {
	desiredKeys := strings.Split(authSecret.ApplicationKey, ",")
	if len(desiredKeys) > 5 {
		return fmt.Errorf("secret contains more than 5 application_key")
	}

	// get the existing value from the portal
	applicationKeys, err := threescaleClient.ApplicationKeys(developerAccountID, applicationID)
	if err != nil {
		return err
	}

	existingKeys := make([]string, 0, len(applicationKeys))
	for _, key := range applicationKeys {
		existingKeys = append(existingKeys, key.Value)
	}

	// delete existing and not desired
	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	a.logger.V(1).Info("syncApplicationAuth", "notDesiredExistingKeys", notDesiredExistingKeys)
	for _, key := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		if err := threescaleClient.DeleteApplicationKey(developerAccountID, applicationID, key); err != nil {
			return fmt.Errorf("error sync applicationAuth for developerAccountID: %d, applicationID: %d, error: %w", developerAccountID, applicationID, err)
		}
	}

	// Create not existing and desired
	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	a.logger.V(1).Info("syncApplicationPlans", "desiredNewKeys", desiredNewKeys)
	for _, key := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.ApplicationPlans map key set
		if _, err := threescaleClient.CreateApplicationKey(developerAccountID, applicationID, key); err != nil {
			return fmt.Errorf("error sync applicationAuth for developerAccountID: %d, applicationID: %d, error: %w", developerAccountID, applicationID, err)
		}
	}

	return nil
}

func (a *appIDAuthMode) SecretReferenceSource(cl client.Client, ns string, authSectretRef *corev1.LocalObjectReference, generateSecret bool) (*AuthSecret, error) {
	secretSource := helper.NewSecretSource(cl, ns)
	applicationKeyStr, err := secretSource.RequiredFieldValueFromRequiredSecret(authSectretRef.Name, ApplicationKey)
	if err != nil {
		return nil, err
	}

	if applicationKeyStr == "" {
		if generateSecret {
			applicationKeyStr = rand.String(16)

			newValues := map[string][]byte{
				ApplicationKey: []byte(applicationKeyStr),
			}

			if err := updateSecret(context.Background(), cl, authSectretRef.Name, ns, newValues); err != nil {
				return nil, err
			}
		} else {
			// Nothing available raise error now
			return nil, fmt.Errorf("no ApplicationKey available in secret and generate secret is set to false")
		}
	}
	return &AuthSecret{ApplicationKey: applicationKeyStr}, nil
}

func (r *ApplicationAuthReconciler) reconcileStatus(resource *capabilitiesv1beta1.ApplicationAuth, err error, logger logr.Logger) (ctrl.Result, error) {
	statusReconciler := NewApplicationAuthStatusReconciler(r.BaseReconciler, resource, err)
	statusResult, statusErr := statusReconciler.Reconcile()

	if err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile status first as the reconcilerError might need to be updated to the status section of the CR before requeueing
	if statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	if statusResult.Requeue {
		logger.Info("Reconciling status not finished. Requeueing.")
		return statusResult, nil
	}

	return ctrl.Result{}, nil
}

func checkApplicationResources(applicationAuthResource *capabilitiesv1beta1.ApplicationAuth, applicationResource *capabilitiesv1beta1.Application) error {
	errors := field.ErrorList{}

	specFldPath := field.NewPath("spec")
	applicationFldPath := specFldPath.Child("applicationCRName")

	if applicationResource.Status.ID == nil {
		errors = append(errors, field.Invalid(applicationFldPath, applicationAuthResource.Spec.ApplicationCRName, "applicationCR name doesnt have a valid application reference"))

		return &helper.SpecFieldError{
			ErrorType:      helper.OrphanError,
			FieldErrorList: errors,
		}
	}

	return nil
}

func updateSecret(ctx context.Context, client client.Client, name string, namespace string, values map[string][]byte) error {
	// get the current values and update the secret
	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret)
	if err != nil {
		// Handle errors gracefully, e.g., log and return or retry
		return err
	}

	newData := secret.Data

	for key, value := range values {
		newData[key] = value
	}

	secret.Data = newData

	if err = client.Update(ctx, secret); err != nil {
		return err
	}

	return nil
}
