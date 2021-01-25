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

type SpecFieldError struct {
	ErrorType      FieldTypeError
	FieldErrorList field.ErrorList
}

var _ SpecError = &SpecFieldError{}

// Error implements the Error interface.
func (s *SpecFieldError) Error() string {
	return s.FieldErrorList.ToAggregate().Error()
}

// FieldType implements the SpecError interface.
func (s *SpecFieldError) FieldType() FieldTypeError {
	return s.ErrorType
}

// SpecError is exposed by errors that can be converted to an api.Status object
// for finer grained details.
type SpecError interface {
	error
	FieldType() FieldTypeError
}

// WaitError represents that the current cluster state is not ready to continue.
// This is (should be) a transient error.
// Example: expected resources have not been created by third party controllers
type WaitError struct {
	Err error
}

func (s *WaitError) Error() string {
	return s.Err.Error()
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

func IsWaitError(err error) bool {
	_, ok := err.(*WaitError)
	return ok
}
