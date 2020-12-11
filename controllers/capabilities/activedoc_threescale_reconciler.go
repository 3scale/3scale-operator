package controllers

import (
	"errors"
	"net/url"
	"reflect"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type ActiveDocThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.ActiveDoc
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	providerAccountHost string
	logger              logr.Logger
}

func NewActiveDocThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.ActiveDoc, threescaleAPIClient *threescaleapi.ThreeScaleClient, providerAccountHost string, logger logr.Logger) *ActiveDocThreescaleReconciler {
	return &ActiveDocThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		providerAccountHost: providerAccountHost,
		logger:              logger.WithValues("3scale Reconciler", providerAccountHost),
	}
}

func (s *ActiveDocThreescaleReconciler) Reconcile() (*threescaleapi.ActiveDoc, error) {
	s.logger.V(1).Info("START")

	remoteActiveDocs, err := s.threescaleAPIClient.ListActiveDocs()
	if err != nil {
		return nil, err
	}

	var remoteActiveDoc *threescaleapi.ActiveDoc

	for idx := range remoteActiveDocs.ActiveDocs {
		// s.resource.Spec.SystemName is not nil (defaults are set)
		if *remoteActiveDocs.ActiveDocs[idx].Element.SystemName == *s.resource.Spec.SystemName {
			remoteActiveDoc = &remoteActiveDocs.ActiveDocs[idx]
			break
		}
	}

	desiredProductID, err := s.getDesiredProductIDFromCR()
	if err != nil {
		return nil, err
	}

	desiredOpenapiObj, err := s.getDesiredActiveDocBody()
	if err != nil {
		return nil, err
	}

	desiredBodyRaw, err := desiredOpenapiObj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	desiredBody := string(desiredBodyRaw)

	if remoteActiveDoc == nil {
		newActiveDoc := &threescaleapi.ActiveDoc{
			Element: threescaleapi.ActiveDocItem{
				Name:                   &s.resource.Spec.Name,
				SystemName:             s.resource.Spec.SystemName,
				Body:                   &desiredBody,
				Description:            s.resource.Spec.Description,
				Published:              s.resource.Spec.Published,
				SkipSwaggerValidations: s.resource.Spec.SkipSwaggerValidations,
				ServiceID:              desiredProductID,
			},
		}

		remoteActiveDoc, err = s.threescaleAPIClient.CreateActiveDoc(newActiveDoc)
		if err != nil {
			return nil, err
		}
	}

	update := false
	updatedActiveDoc := &threescaleapi.ActiveDoc{
		Element: threescaleapi.ActiveDocItem{
			ID: remoteActiveDoc.Element.ID,
		},
	}

	if remoteActiveDoc.Element.Name == nil || *remoteActiveDoc.Element.Name != s.resource.Spec.Name {
		s.logger.V(1).Info("update Name", "Difference", cmp.Diff(remoteActiveDoc.Element.Name, s.resource.Spec.Name))
		updatedActiveDoc.Element.Name = &s.resource.Spec.Name
		update = true
	}

	// s.resource.Spec.SystemName is not nil (defaults are set)
	if remoteActiveDoc.Element.SystemName == nil || *remoteActiveDoc.Element.SystemName != *s.resource.Spec.SystemName {
		s.logger.V(1).Info("update SystemName", "Difference", cmp.Diff(remoteActiveDoc.Element.SystemName, s.resource.Spec.SystemName))
		updatedActiveDoc.Element.SystemName = s.resource.Spec.SystemName
		update = true
	}

	if s.resource.Spec.Description != nil && !reflect.DeepEqual(s.resource.Spec.Description, remoteActiveDoc.Element.Description) {
		s.logger.V(1).Info("update Description", "Difference", cmp.Diff(remoteActiveDoc.Element.Description, s.resource.Spec.Description))
		updatedActiveDoc.Element.Description = s.resource.Spec.Description
		update = true
	}

	if s.resource.Spec.Published != nil && !reflect.DeepEqual(s.resource.Spec.Published, remoteActiveDoc.Element.Published) {
		s.logger.V(1).Info("update Published", "Difference", cmp.Diff(remoteActiveDoc.Element.Published, s.resource.Spec.Published))
		updatedActiveDoc.Element.Published = s.resource.Spec.Published
		update = true
	}

	if s.resource.Spec.SkipSwaggerValidations != nil && !reflect.DeepEqual(s.resource.Spec.SkipSwaggerValidations, remoteActiveDoc.Element.SkipSwaggerValidations) {
		s.logger.V(1).Info("update SkipSwaggerValidations", "Difference", cmp.Diff(remoteActiveDoc.Element.SkipSwaggerValidations, s.resource.Spec.SkipSwaggerValidations))
		updatedActiveDoc.Element.SkipSwaggerValidations = s.resource.Spec.SkipSwaggerValidations
		update = true
	}

	// ActiveDoc Product ID
	// Only update if desired Product ID needs to be changed to some not null value
	// If desired Product ID needs to be changed to null, the update needs a different client call
	if desiredProductID != nil {
		if remoteActiveDoc.Element.ServiceID == nil || *desiredProductID != *remoteActiveDoc.Element.ServiceID {
			s.logger.V(1).Info("update ProductID", "Difference", cmp.Diff(remoteActiveDoc.Element.ServiceID, desiredProductID))
			updatedActiveDoc.Element.ServiceID = desiredProductID
			update = true
		}
	}

	existingOpenapiObj, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(*remoteActiveDoc.Element.Body))
	if err != nil {
		return nil, err
	}

	// Compare parsed openapi3 objects
	// Avoid detecting differences from serialization
	if !reflect.DeepEqual(desiredOpenapiObj, existingOpenapiObj) {
		s.logger.V(1).Info("update BODY", "Difference", cmp.Diff(desiredOpenapiObj, existingOpenapiObj))
		updatedActiveDoc.Element.Body = &desiredBody
		update = true
	}

	if update {
		s.logger.V(1).Info("Desired ActiveDoc needs sync")
		_, err := s.threescaleAPIClient.UpdateActiveDoc(updatedActiveDoc)
		if err != nil {
			return nil, err
		}
	}

	if desiredProductID == nil && remoteActiveDoc.Element.ServiceID != nil {
		s.logger.V(1).Info("Remove product relationship", "ID", remoteActiveDoc.Element.ServiceID)
		_, err := s.threescaleAPIClient.UnbindActiveDocFromProduct(*remoteActiveDoc.Element.ID)
		if err != nil {
			return nil, err
		}
	}

	return remoteActiveDoc, nil
}

func (s *ActiveDocThreescaleReconciler) getDesiredProductIDFromCR() (*int64, error) {
	if s.resource.Spec.ProductSystemName == nil {
		return nil, nil
	}

	// Getting product ID from Product CR status field. It should be fine as product CR is required to be in "Ready" status.
	// Another alternative would be fetch the list of 3scale products,
	// filter by systemname and get the ID
	productList, err := controllerhelper.ProductList(s.resource.Namespace, s.Client(), s.providerAccountHost, s.logger)
	if err != nil {
		return nil, err
	}

	idx := controllerhelper.FindProductBySystemName(productList, *s.resource.Spec.ProductSystemName)
	if idx < 0 {
		// External references validation makes sure product CR exists
		return nil, errors.New("Product CR not found. External references validation should avoid reaching this state")
	}

	return productList[idx].Status.ID, nil
}

func (s *ActiveDocThreescaleReconciler) getDesiredActiveDocBody() (*openapi3.Swagger, error) {
	// OpenAPIRef is oneOf by CRD openapiV3 validation
	if s.resource.Spec.ActiveDocOpenAPIRef.SecretRef != nil {
		return s.readOpenAPISecret()
	}

	// Must be URL
	return s.readOpenAPIFromURL()
}

func (s *ActiveDocThreescaleReconciler) readOpenAPISecret() (*openapi3.Swagger, error) {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("activeDocOpenAPIRef")
	secretRefFldPath := openapiRefFldPath.Child("secretRef")

	// s.resource.Spec.ActiveDocOpenAPIRef.SecretRef.Namespace set in defaults
	objectKey := types.NamespacedName{Name: s.resource.Spec.ActiveDocOpenAPIRef.SecretRef.Name, Namespace: s.resource.Spec.ActiveDocOpenAPIRef.SecretRef.Namespace}
	openapiSecretObj := &corev1.Secret{}

	// Read secret
	if err := s.Client().Get(s.Context(), objectKey, openapiSecretObj); err != nil {
		if apimachineryerrors.IsNotFound(err) {
			fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.SecretRef, "Secret not found"))
			return nil, &helper.SpecFieldError{
				ErrorType:      helper.InvalidError,
				FieldErrorList: fieldErrors,
			}
		}

		// unexpected error
		return nil, err
	}

	if len(openapiSecretObj.Data) != 1 {
		fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.SecretRef, "Secret was empty or contains too many fields. Only one is required."))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	// Get key value
	dataByteArray := func(secret *corev1.Secret) []byte {
		for _, v := range secret.Data {
			return v
		}
		return nil
	}(openapiSecretObj)

	openapiObj, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(dataByteArray)
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.SecretRef, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	err = openapiObj.Validate(s.Context())
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.SecretRef, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	return openapiObj, nil
}

func (s *ActiveDocThreescaleReconciler) readOpenAPIFromURL() (*openapi3.Swagger, error) {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("activeDocOpenAPIRef")
	urlRefFldPath := openapiRefFldPath.Child("url")

	openAPIURL, err := url.Parse(*s.resource.Spec.ActiveDocOpenAPIRef.URL)
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(urlRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.URL, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	openapiObj, err := openapi3.NewSwaggerLoader().LoadSwaggerFromURI(openAPIURL)
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(urlRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.URL, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	err = openapiObj.Validate(s.Context())
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(urlRefFldPath, s.resource.Spec.ActiveDocOpenAPIRef.URL, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	return openapiObj, nil
}
