package controller

import (
	"github.com/3scale/3scale-operator/pkg/controller/amp"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, amp.Add)
}
