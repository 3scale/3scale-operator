package controller

import (
	"github.com/3scale/3scale-operator/pkg/controller/consolidated"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, consolidated.Add)
}
