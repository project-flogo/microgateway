package js

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Script string `md:"script"`
}

type Input struct {
	Parameters map[string]interface{} `md:"parameters"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	parameters, err := coerce.ToObject(values["parameters"])
	if err != nil {
		return err
	}
	r.Parameters = parameters
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"parameters": r.Parameters,
	}
}

type Output struct {
	Error        bool                   `md:"error"`
	ErrorMessage string                 `md:"errorMessage"`
	Result       map[string]interface{} `md:"result"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	errorValue, err := coerce.ToBool(values["error"])
	if err != nil {
		return err
	}
	o.Error = errorValue
	errorMessage, err := coerce.ToString(values["errorMessage"])
	if err != nil {
		return err
	}
	o.ErrorMessage = errorMessage
	result, err := coerce.ToObject(values["result"])
	if err != nil {
		return err
	}
	o.Result = result
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"error":        o.Error,
		"errorMessage": o.ErrorMessage,
		"result":       o.Result,
	}
}
