package controller

import (
	"github.com/rafael-azevedo/operator-workshop/containerset/pkg/controller/containerset"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, containerset.Add)
}
