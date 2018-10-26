package activity

import "github.com/project-flogo/core/data/coerce"

// Settings are the settings for the dummy activity
type Settings struct {
}

// Input are the inputs for the dummy activity
type Input struct {
	Message string `md:message,required"`
}

// FromMap sets the Input from a map
func (r *Input) FromMap(values map[string]interface{}) error {
	value, err := coerce.ToString(values["message"])
	if err != nil {
		return err
	}
	r.Message = value
	return nil
}

// ToMap converts the Input to a map
func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"message": r.Message,
	}
}

// Output are the outputs for the dummy activity
type Output struct {
	Data string `md:data,required"`
}

// FromMap sets the Output from a map
func (o *Output) FromMap(values map[string]interface{}) error {
	value, err := coerce.ToString(values["data"])
	if err != nil {
		return err
	}
	o.Data = value
	return nil
	return nil
}

// ToMap converts the Output to a map
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data": o.Data,
	}
}
