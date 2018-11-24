package jwt

import (
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	logger "github.com/project-flogo/core/support/log"
	"github.com/stretchr/testify/assert"
)

type initContext struct {
	settings map[string]interface{}
}

func newInitContext(values map[string]interface{}) *initContext {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &initContext{
		settings: values,
	}
}

func (i *initContext) Settings() map[string]interface{} {
	return i.settings
}

func (i *initContext) MapperFactory() mapper.Factory {
	return nil
}

func (i *initContext) Logger() logger.Logger {
	return logger.RootLogger()
}

type activityContext struct {
	input  map[string]interface{}
	output map[string]interface{}
}

func newActivityContext(values map[string]interface{}) *activityContext {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &activityContext{
		input:  values,
		output: make(map[string]interface{}),
	}
}

func (a *activityContext) ActivityHost() activity.Host {
	return a
}

func (a *activityContext) Name() string {
	return "test"
}

func (a *activityContext) GetInput(name string) interface{} {
	return a.input[name]
}

func (a *activityContext) SetOutput(name string, value interface{}) error {
	a.output[name] = value
	return nil
}

func (a *activityContext) GetInputObject(input data.StructValue) error {
	return input.FromMap(a.input)
}

func (a *activityContext) SetOutputObject(output data.StructValue) error {
	a.output = output.ToMap()
	return nil
}

func (a *activityContext) GetSharedTempData() map[string]interface{} {
	return nil
}

func (a *activityContext) ID() string {
	return "test"
}

func (a *activityContext) IOMetadata() *metadata.IOMetadata {
	return nil
}

func (a *activityContext) Reply(replyData map[string]interface{}, err error) {

}

func (a *activityContext) Return(returnData map[string]interface{}, err error) {

}

func (a *activityContext) Scope() data.Scope {
	return nil
}

func (a *activityContext) Logger() logger.Logger {
	return logger.RootLogger()
}

func TestJWT(t *testing.T) {
	activity, err := New(newInitContext(nil))
	assert.Nil(t, err)
	execute := func(serviceName string, values map[string]interface{}, should error) {
		_, err := activity.Eval(newActivityContext(values))
		assert.Equal(t, should, err)
	}

	inputValues := map[string]interface{}{
		"signingMethod": "HMAC",
		"key":           "qwertyuiopasdfghjklzxcvbnm789101",
		"aud":           "www.mashling.io",
		"iss":           "Mashling",
		"sub":           "tempuser@mail.com",
		"token":         "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0MDQ4NzgzNywiZXhwIjoxNTcyMDIzODM4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsImlkIjoiMSJ9.-Tzfn5ZS0kM-u07qkpFrDxdyptBJIvLesuUzVXdqn48",
	}
	execute("reset", inputValues, nil)
}
