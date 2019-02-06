package core

import (
	"github.com/project-flogo/core/activity"
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
	Condition     *Expr
	Service       *Service
	Input         []*Expr
	HaltCondition *Expr
}

// Setting is a service setting
type Setting struct {
	Name  string
	Value interface{}
}

// Service defines a functional target that may be invoked by a step in an execution flow.
type Service struct {
	Name     string
	Settings []Setting
	Activity activity.Activity
}

// Response defines response handling rules.
type Response struct {
	Condition *Expr
	Error     bool
	Output    Output
}

// Output defines response output values back to a trigger event.
type Output struct {
	Code  *Expr
	Data  *Expr
	Datum []*Expr
}
