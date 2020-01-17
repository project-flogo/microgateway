package graphql

import "github.com/project-flogo/core/data/coerce"

// Settings settings for the GraphQL policy service
type Settings struct {
	Mode  string `md:"mode,allowed(a,b)"`
	Limit string `md:"limit"`
}

// Input input meta data
type Input struct {
	Query         string `md:"query"`
	SchemaFile    string `md:"schemaFile"`
	MaxQueryDepth int    `md:"maxQueryDepth"`
	Token         string `md:"token"`
	Operation     string `md:"operation,allowed(startconsume,stopconsume)"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"query":         i.Query,
		"schemaFile":    i.SchemaFile,
		"maxQueryDepth": i.MaxQueryDepth,
		"token":         i.Token,
		"operation":     i.Operation,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Query, err = coerce.ToString(values["query"])
	if err != nil {
		return err
	}
	i.SchemaFile, err = coerce.ToString(values["schemaFile"])
	if err != nil {
		return err
	}
	i.MaxQueryDepth, err = coerce.ToInt(values["maxQueryDepth"])
	if err != nil {
		return err
	}
	i.Token, err = coerce.ToString(values["token"])
	if err != nil {
		return err
	}
	i.Operation, err = coerce.ToString(values["operation"])
	if err != nil {
		return err
	}

	return nil
}

type Output struct {
	Valid             bool   `md:"valid"`
	ValidationMessage string `md:"validationMessage"`
	Error             bool   `md:"error"`
	ErrorMessage      string `md:"errorMessage"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	valid, err := coerce.ToBool(values["valid"])
	if err != nil {
		return err
	}
	o.Valid = valid
	o.ValidationMessage = values["validationMessage"].(string)
	o.Error = values["error"].(bool)
	o.ErrorMessage = values["errorMessage"].(string)
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"valid":             o.Valid,
		"validationMessage": o.ValidationMessage,
		"error":             o.Error,
		"errorMessage":      o.ErrorMessage,
	}
}
