package ratelimiter

import (
	"github.com/project-flogo/core/data/coerce"
)

// Settings are the settings for the rate limiter
type Settings struct {
	Limit          string  `md:"limit,required"`
	SpikeThreshold float64 `md:"spikeThreshold"`
}

// Input is the input for the rate limiter
type Input struct {
	Token string `md:"token,required"`
}

// FromMap converts the settings from a map of settings
func (r *Input) FromMap(values map[string]interface{}) error {
	token, err := coerce.ToString(values["token"])
	if err != nil {
		return err
	}
	r.Token = token
	return nil
}

// ToMap converts the settings to a map from a struct
func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"token": r.Token,
	}
}

// Output is the output of the rate limiter
type Output struct {
	LimitReached   bool   `md:"limitReached"`
	LimitAvailable int64  `md:"limitAvailable"`
	Error          bool   `md:"error"`
	ErrorMessage   string `md:"errorMessage"`
}

// FromMap converts the output from a map to a struct
func (o *Output) FromMap(values map[string]interface{}) error {
	limitReached, err := coerce.ToBool(values["limitReached"])
	if err != nil {
		return err
	}
	o.LimitReached = limitReached
	limitAvailable, err := coerce.ToInt64(values["limitAvailable"])
	if err != nil {
		return err
	}
	o.LimitAvailable = limitAvailable
	hasError, err := coerce.ToBool(values["error"])
	if err != nil {
		return err
	}
	o.Error = hasError
	errorMessage, err := coerce.ToString(values["errorMessage"])
	if err != nil {
		return err
	}
	o.ErrorMessage = errorMessage
	return nil
}

// ToMap converts the output to a map from a struct
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"limitReached":   o.LimitReached,
		"limitAvailable": o.LimitAvailable,
		"error":          o.Error,
		"errorMessage":   o.ErrorMessage,
	}
}
