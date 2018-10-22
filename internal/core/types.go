package core

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/microgateway/internal/types"
)

// Microgateway defines a microgateway
type Microgateway struct {
	Name          string
	Async         bool
	Steps         []Step
	Responses     []Response
	Configuration map[string]interface{}
}

// Step conditionally defines a step in a route's execution flow.
type Step struct {
	Condition *types.Expr
	Service   *Service
	Input     map[string]*types.Expr
}

// Service defines a functional target that may be invoked by a step in an execution flow.
type Service struct {
	Name     string
	Settings map[string]interface{}
	Activity activity.Activity
}

// Response defines response handling rules.
type Response struct {
	Condition *types.Expr
	Error     bool
	Output    Output
}

// Output defines response output values back to a trigger event.
type Output struct {
	Code  *types.Expr
	Data  *types.Expr
	Datum map[string]*types.Expr
}
