package jwt

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct{}

type Input struct {
	Token         string `md:"token"`
	Key           string `md:"key"`
	SigningMethod string `md:"signingMethod"`
	Issuer        string `md:"iss"`
	Subject       string `md:"sub"`
	Audience      string `md:"aud"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	var err error
	r.Token, err = coerce.ToString(values["token"])
	if err != nil {
		return err
	}
	r.Key, err = coerce.ToString(values["key"])
	if err != nil {
		return err
	}
	r.SigningMethod, err = coerce.ToString(values["signingMethod"])
	if err != nil {
		return err
	}
	r.Issuer, err = coerce.ToString(values["iss"])
	if err != nil {
		return err
	}
	r.Subject, err = coerce.ToString(values["sub"])
	if err != nil {
		return err
	}
	r.Audience, err = coerce.ToString(values["aud"])
	if err != nil {
		return err
	}
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"token":         r.Token,
		"key":           r.Key,
		"signingMethod": r.SigningMethod,
		"iss":           r.Issuer,
		"sub":           r.Subject,
		"aud":           r.Audience,
	}
}

type Output struct {
	Valid             bool        `md:"valid"`
	Token             ParsedToken `md:"token"`
	ValidationMessage string      `md:"validationMessage"`
	Error             bool        `md:"error"`
	ErrorMessage      string      `md:"errorMessage"`
}

// ParsedToken is a parsed JWT token.
type ParsedToken struct {
	Claims        map[string]interface{} `json:"claims"`
	Signature     string                 `json:"signature"`
	SigningMethod string                 `json:"signingMethod"`
	Header        map[string]interface{} `json:"header"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	valid, err := coerce.ToBool(values["valid"])
	if err != nil {
		return err
	}
	o.Valid = valid
	token, err := coerce.ToAny(values["token"])
	if err != nil {
		return err
	}
	o.Token = token.(ParsedToken)
	o.ValidationMessage = values["validationMessage"].(string)
	o.Error = values["error"].(bool)
	o.ErrorMessage = values["errorMessage"].(string)
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"valid":             o.Valid,
		"token":             o.Token,
		"validationMessage": o.ValidationMessage,
		"error":             o.Error,
		"errorMessage":      o.ErrorMessage,
	}
}
