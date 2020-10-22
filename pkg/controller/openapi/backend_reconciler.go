package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/google/go-cmp/cmp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type BackendReconciler struct {
	*reconcilers.BaseReconciler
	openapiCR       *capabilitiesv1beta1.Openapi
	openapiObj      *openapi3.Swagger
	providerAccount *controllerhelper.ProviderAccount
	logger          logr.Logger
}

func NewBackendReconciler(b *reconcilers.BaseReconciler,
	openapiCR *capabilitiesv1beta1.Openapi,
	openapiObj *openapi3.Swagger,
	providerAccount *controllerhelper.ProviderAccount,
	logger logr.Logger,
) *BackendReconciler {
	return &BackendReconciler{
		BaseReconciler:  b,
		openapiCR:       openapiCR,
		openapiObj:      openapiObj,
		providerAccount: providerAccount,
		logger:          logger,
	}
}

func (p *BackendReconciler) Logger() logr.Logger {
	return p.logger
}

func (p *BackendReconciler) Reconcile() ([]*capabilitiesv1beta1.Backend, error) {
	desired, err := p.desired()
	if err != nil {
		return nil, err
	}

	if p.Logger().V(1).Enabled() {
		jsonData, err := json.MarshalIndent(desired, "", "  ")
		if err != nil {
			return nil, err
		}
		p.Logger().V(1).Info(string(jsonData))
	}

	// single backend implementation
	return nil, p.ReconcileResource(&capabilitiesv1beta1.Backend{}, desired, p.backendMutator)
}

func (p *BackendReconciler) desired() (*capabilitiesv1beta1.Backend, error) {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("openapiRef")

	// system name
	systemName := p.desiredSystemName()

	// obj name
	objName := p.desiredObjName()

	errStrings := validation.IsDNS1123Label(objName)
	if len(errStrings) > 0 {
		fieldErrors = append(fieldErrors, field.Invalid(openapiRefFldPath, p.openapiCR.Spec.OpenAPIRef, strings.Join(errStrings, "--")))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	// backend name
	name := fmt.Sprintf("%s Backend", p.openapiObj.Info.Title)

	// backend description
	description := fmt.Sprintf("Backend of %s", p.openapiObj.Info.Title)

	// private base URL
	privateBaseURL, err := p.desiredPrivateBaseURL()
	if err != nil {
		return nil, err
	}

	backend := &capabilitiesv1beta1.Backend{
		TypeMeta: metav1.TypeMeta{
			Kind:       capabilitiesv1beta1.BackendKind,
			APIVersion: capabilitiesv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: p.openapiCR.Namespace,
		},
		Spec: capabilitiesv1beta1.BackendSpec{
			Name:               name,
			SystemName:         systemName,
			PrivateBaseURL:     privateBaseURL,
			Description:        description,
			ProviderAccountRef: p.openapiCR.Spec.ProviderAccountRef,
		},
	}

	backend.SetDefaults(p.Logger())

	// internal validation
	validationErrors := backend.Validate()
	if len(validationErrors) > 0 {
		return nil, errors.New(validationErrors.ToAggregate().Error())
	}

	err = p.SetOwnerReference(p.openapiCR, backend)
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func (p *BackendReconciler) backendMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*capabilitiesv1beta1.Backend)
	if !ok {
		return false, fmt.Errorf("%T is not a *capabilitiesv1beta1.Backend", existingObj)
	}
	desired, ok := desiredObj.(*capabilitiesv1beta1.Backend)
	if !ok {
		return false, fmt.Errorf("%T is not a *capabilitiesv1beta1.Backend", desiredObj)
	}

	// Metadata labels and annotations
	updated := helper.EnsureObjectMeta(existing, desired)

	// OwnerRefenrence
	updatedTmp, err := p.EnsureOwnerReference(p.openapiCR, existing)
	if err != nil {
		return false, err
	}
	updated = updated || updatedTmp

	// Maybe too rough compare method?
	// What if backend controller adds or modifies something?
	// the openapi controller will be reconciliating.
	// maybe compare only "managed" fields
	if !reflect.DeepEqual(existing.Spec, desired.Spec) {
		diff := cmp.Diff(existing.Spec, desired.Spec)
		p.Logger().Info(fmt.Sprintf("%s spec has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}

func (p *BackendReconciler) desiredSystemName() string {
	// Same as product system name
	// Duplicated implementation. Refactor
	if p.openapiCR.Spec.ProductSystemName != "" {
		return p.openapiCR.Spec.ProductSystemName
	}

	return helper.SystemNameFromOpenAPITitle(p.openapiObj)
}

func (p *BackendReconciler) desiredObjName() string {
	// DNS1123 Label compliant name. Due to UIDs are 36 characters of length this
	// means that the maximum prefix lenght that can be provided is of 26
	// characters. If the generated name is not DNS1123 compliant an error is
	// returned
	// Maybe truncate?
	return fmt.Sprintf("%s-%s", helper.K8sNameFromOpenAPITitle(p.openapiObj), string(p.openapiCR.UID))
}

func (p *BackendReconciler) desiredPrivateBaseURL() (string, error) {
	privateBaseURL := p.openapiCR.Spec.PrivateBaseURL
	if privateBaseURL == "" {
		var err error
		privateBaseURL, err = helper.BaseURLFromOpenAPI(p.openapiObj)
		if err != nil {
			fieldErrors := field.ErrorList{}
			specFldPath := field.NewPath("spec")
			openapiRefFldPath := specFldPath.Child("openapiRef")
			fieldErrors = append(fieldErrors, field.Invalid(openapiRefFldPath, p.openapiCR.Spec.OpenAPIRef, err.Error()))
			return "", &helper.SpecFieldError{
				ErrorType:      helper.InvalidError,
				FieldErrorList: fieldErrors,
			}
		}
	}

	return privateBaseURL, nil
}
