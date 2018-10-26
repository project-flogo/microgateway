package circuitbreaker

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Mode      string `md:"mode,allowed(a,b,c,d)"`
	Threshold int    `md:"threshold"`
	Period    int    `md:"period"`
	Timeout   int    `md:"timeout"`
}

type Input struct {
	Operation string `md:"operation,allowed(counter,reset)"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	operation, err := coerce.ToString(values["operation"])
	if err != nil {
		return err
	}
	r.Operation = operation
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"operation": r.Operation,
	}
}

type Output struct {
	Tripped bool `md:"tripped"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	tripped, err := coerce.ToBool(values["tripped"])
	if err != nil {
		return err
	}
	o.Tripped = tripped
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"tripped": o.Tripped,
	}
}
