package microgateway

// Settings are the settings for the microgateway
type Settings struct {
	URI   string `md:"uri,required"`
	Async bool   `md:"async"`
}

// Input represents the inputs into the microgateway
type Input struct {
}

// FromMap sets Input from a map
func (r *Input) FromMap(values map[string]interface{}) error {
	return nil
}

// ToMap converts Input to a map
func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}

// Output represents the outputs from the microgateway
type Output struct {
}

// FromMap sets Output from a map
func (o *Output) FromMap(values map[string]interface{}) error {
	return nil
}

// ToMap converts Output to a map
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}
