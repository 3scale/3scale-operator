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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
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

		// populate authSecret struct
		authSecret := authSecretReferenceSource(r.Client(), applicationAuth.Namespace, applicationAuth.Spec.AuthSecretRef, reqLogger)
		err = r.applicationAuthReconciler(applicationAuth, *developerAccount.Status.ID, *application.Status.ID, product, *authSecret, threescaleAPIClient)
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

func (r *ApplicationAuthReconciler) applicationAuthReconciler(
	applicationAuth *capabilitiesv1beta1.ApplicationAuth,
	developerAccountID int64,
	applicationID int64,
	product *capabilitiesv1beta1.Product,
	authSecret AuthSecret,
	threescaleClient *threescaleapi.ThreeScaleClient,
) error {
	// generate sha base of timestamp
	timestamp := time.Now().Unix()
	// Write the timestamp string and encode to hash
	hash := sha256.New()
	hash.Write([]byte(strconv.FormatInt(timestamp, 10)))
	hashedBytes := hash.Sum(nil)
	hashedString := hex.EncodeToString(hashedBytes)

	// Check the values if populated or the GenerateSecret field is true and make the api call to update
	// If UserKey is not populated generate random sha
	if authSecret.UserKey == "" && *applicationAuth.Spec.GenerateSecret {
		authSecret.UserKey = hashedString
	}
	if authSecret.UserKey != "" {
		params := make(map[string]string)
		params["user_key"] = authSecret.UserKey
		// edge case if the operator is stopped before reconcile finished need to nil check application.Status.ID
		_, err := threescaleClient.UpdateApplication(developerAccountID, applicationID, params)
		if err != nil {
			return err
		}
	}

	if authSecret.ApplicationKey != "" {
		foundApplication, err := threescaleClient.CreateApplicationKey(developerAccountID, applicationID, authSecret.ApplicationKey)
		if err != nil {
			return err
		}

		authSecret.ApplicationID = foundApplication.ApplicationId
	}

	if applicationAuth.Spec.GenerateSecret != nil && *applicationAuth.Spec.GenerateSecret {
		foundApplication, err := threescaleClient.CreateApplicationRandomKey(developerAccountID, applicationID)
		if err != nil {
			return err
		}
		authSecret.ApplicationID = foundApplication.ApplicationId
		var foundApplicationKeys []threescaleapi.ApplicationKey
		foundApplicationKeys, err = threescaleClient.ApplicationKeys(developerAccountID, applicationID)
		if err != nil {
			return err
		}
		lastKey := len(foundApplicationKeys) - 1
		authSecret.ApplicationKey = fmt.Sprint(foundApplicationKeys[lastKey].Value)
	}

	// get the current values and update the secret
	ApplicationAuthSecret := &corev1.Secret{}
	err := r.Client().Get(r.Context(), types.NamespacedName{
		Name:      applicationAuth.Spec.AuthSecretRef.Name,
		Namespace: applicationAuth.Namespace,
	}, ApplicationAuthSecret)
	if err != nil {
		// Handle errors gracefully, e.g., log and return or retry
		r.Logger().Error(err, "Failed to get existing ApplicationAuthSecret")
		return err
	}
	newData := ApplicationAuthSecret.Data
	newValues := map[string][]byte{
		UserKey:        []byte(authSecret.UserKey),
		ApplicationID:  []byte(authSecret.ApplicationID),
		ApplicationKey: []byte(authSecret.ApplicationKey),
	}
	for key, value := range newValues {
		newData[key] = value
	}

	ApplicationAuthSecret.Data = newData
	err = r.Client().Update(r.Context(), ApplicationAuthSecret)
	if err != nil {
		r.Logger().Error(err, "Failed to update ApplicationAuthSecret")
		return err
	}

	return nil
}

func authSecretReferenceSource(cl client.Client, ns string, authSectretRef *corev1.LocalObjectReference, logger logr.Logger) *AuthSecret {
	if authSectretRef != nil {
		logger.Info("LookupAuthSecret", "ns", ns, "authSecretRef", authSectretRef)
		secretSource := helper.NewSecretSource(cl, ns)
		userKeyStr, err := secretSource.RequiredFieldValueFromRequiredSecret(authSectretRef.Name, UserKey)
		if err != nil {
			userKeyStr = ""
		}
		applicationKeyStr, err := secretSource.RequiredFieldValueFromRequiredSecret(authSectretRef.Name, ApplicationKey)
		if err != nil {
			applicationKeyStr = ""
		}

		return &AuthSecret{UserKey: userKeyStr, ApplicationKey: applicationKeyStr}
	}

	return nil
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
