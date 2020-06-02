package backend

import (
	"github.com/3scale/3scale-operator/pkg/helper"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

type specFieldError struct {
	errorType      helper.FieldTypeError
	fieldErrorList field.ErrorList
}

var _ helper.SpecError = &specFieldError{}

// Error implements the Error interface.
func (s *specFieldError) Error() string {
	return s.fieldErrorList.ToAggregate().Error()
}

// FieldType implements the SpecError interface.
func (s *specFieldError) FieldType() helper.FieldTypeError {
	return s.errorType
}
