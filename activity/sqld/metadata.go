package sqld

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	File string `md:"file"`
}

type Input struct {
	Payload map[string]interface{} `md:"payload,required"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	payload, err := coerce.ToObject(values["payload"])
	if err != nil {
		return err
	}
	r.Payload = payload
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"payload": r.Payload,
	}
}

type Output struct {
	Attack       float32                `md:"attack"`
	AttackValues map[string]interface{} `md:"attackValues"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	attack, err := coerce.ToFloat32(values["attack"])
	if err != nil {
		return err
	}
	o.Attack = attack
	attackValues, err := coerce.ToObject(values["attackValues"])
	if err != nil {
		return err
	}
	o.AttackValues = attackValues
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"attack":       o.Attack,
		"attackValues": o.AttackValues,
	}
}
