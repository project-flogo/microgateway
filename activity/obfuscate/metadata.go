package obfusacte

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Operation string        `md:"operation,required"`
	Fields    []interface{} `md:"fields,required"`
}

type Input struct {
	Payload string `md:"payload"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	payload, _ := coerce.ToString(values["payload"])
	r.Payload = payload
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"payload": r.Payload,
	}
}

type Output struct {
	Result interface{} `md:"result"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	o.Result, _ = values["result"]
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"result": o.Result,
	}
}
