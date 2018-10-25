package core

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/microgateway/api"
)

// Adapter is an adapter activity for ServiceFunc
type Adapter struct {
	Handler api.ServiceFunc
}

// Metadata returns the metadata for the adapter activity
func (a *Adapter) Metadata() *activity.Metadata {
	return nil
}

// Eval evaluates the adapter activity
func (a *Adapter) Eval(ctx activity.Context) (done bool, err error) {
	return a.Handler(ctx)
}
