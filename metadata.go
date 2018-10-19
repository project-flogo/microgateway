package microgateway

type Settings struct {
	URI string `md:"uri,required"`
}

type Input struct {
}

func (r *Input) FromMap(values map[string]interface{}) error {
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}

type Output struct {
}

func (o *Output) FromMap(values map[string]interface{}) error {
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}
