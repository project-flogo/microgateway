package ratelimiter

import (
	"testing"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
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

func TestRatelimiter(t *testing.T) {
	activity, err := New(newInitContext(map[string]interface{}{
		"limit": "1-S",
	}))
	assert.Nil(t, err)

	ctx := newActivityContext(map[string]interface{}{
		"token": "abc123",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.False(t, ctx.output["limitReached"].(bool), "limit should not be reached")

	ctx = newActivityContext(map[string]interface{}{
		"token": "abc123",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.True(t, ctx.output["limitReached"].(bool), "limit should be reached")

	ctx = newActivityContext(map[string]interface{}{
		"token": "sally",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.False(t, ctx.output["limitReached"].(bool), "limit should not be reached")

	time.Sleep(time.Second)

	ctx = newActivityContext(map[string]interface{}{
		"token": "abc123",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.False(t, ctx.output["limitReached"].(bool), "limit should not be reached")
}
