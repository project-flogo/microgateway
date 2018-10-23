package types

// Microgateway defines a microgateway
type Microgateway struct {
	Name          string                 `json:"name" jsonschema:"required"`
	Pattern       string                 `json:"pattern,omitempty"`
	Async         bool                   `json:"async,omitempty"`
	Steps         []Step                 `json:"steps" jsonschema:"required,minItems=1"`
	Responses     []Response             `json:"responses,omitempty"`
	Services      []Service              `json:"services,omitempty" jsonschema:"uniqueItems=true"`
	Configuration map[string]interface{} `json:"configuration,omitempty" jsonschema:"additionalProperties"`
}

// Step conditionally defines a step in a route's execution flow.
type Step struct {
	Condition string                 `json:"if,omitempty"`
	Service   string                 `json:"service" jsonschema:"required"`
	Input     map[string]interface{} `json:"input,omitempty" jsonschema:"additionalProperties"`
}

// Response defines response handling rules.
type Response struct {
	Condition string `json:"if,omitempty"`
	Error     bool   `json:"error" jsonschema:"required"`
	Output    Output `json:"output,omitempty" jsonschema:"required"`
}

// Output defines response output values back to a trigger event.
type Output struct {
	Code interface{} `json:"code,omitempty"`
	Data interface{} `json:"data" jsonschema:"additionalProperties"`
}

// Service defines a functional target that may be invoked by a step in an execution flow.
type Service struct {
	Name        string                 `json:"name" jsonschema:"required"`
	Ref         string                 `json:"ref" jsonschema:"required"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty" jsonschema:"additionalProperties"`
}
