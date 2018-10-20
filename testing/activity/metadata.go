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
	Data string `md:data,required"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	value, err := coerce.ToString(values["data"])
	if err != nil {
		return err
	}
	o.Data = value
	return nil
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data": o.Data,
	}
}
