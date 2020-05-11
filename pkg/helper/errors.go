package helper

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type FieldTypeError int

const (
	// InvalidError represents that the combination of configuration in the resource spec
	// is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the two mutually exclusive attributes have been set.
	InvalidError FieldTypeError = iota

	// OrphanError represents that the configuration in the resource spec
	// contains reference to some not existing resource.
	// This is (should be) a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the product Spec references non existing backend resource
	OrphanError
)

// SpecError is exposed by errors that can be converted to an api.Status object
// for finer grained details.
type SpecError interface {
	error
	FieldType() FieldTypeError
}

// invalidProductError is an error intended for consumption by a REST API server; it can also be
// reconstructed by clients from a REST response. Public to allow easy type switches.
type specFieldError struct {
	errorType FieldTypeError
	// path
	fieldErrorList field.ErrorList
}

var _ SpecError = &specFieldError{}

// Error implements the Error interface.
func (s *specFieldError) Error() string {
	return s.fieldErrorList.ToAggregate().Error()
}

// FieldType implements the SpecError interface.
func (s *specFieldError) FieldType() FieldTypeError {
	return s.errorType
}

func IsInvalidSpecError(err error) bool {
	if specErrorObj, ok := err.(SpecError); ok && specErrorObj.FieldType() == InvalidError {
		return true
	}
	return false
}

func IsOrphanSpecError(err error) bool {
	if specErrorObj, ok := err.(SpecError); ok && specErrorObj.FieldType() == OrphanError {
		return true
	}
	return false
}
