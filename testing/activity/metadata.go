package activity

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
}

type Input struct {
	Message string `md:message,required"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	value, err := coerce.ToString(values["message"])
	if err != nil {
		return err
	}
	r.Message = value
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"message": r.Message,
	}
}

type Output struct {
}

func (o *Output) FromMap(values map[string]interface{}) error {
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}
