package controller

import (
	"github.com/bigkevmcd/webhook-secret-operator/pkg/controller/webhooksecret"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, webhooksecret.Add)
}
