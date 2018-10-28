package sqld

import (
	"fmt"
	"testing"

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

func TestSQLD(t *testing.T) {
	activity, err := New(newInitContext(nil))
	assert.Nil(t, err)

	test := func(a string, attack bool) {
		var payload interface{} = map[string]interface{}{
			"content": map[string]interface{}{
				"test": a,
			},
		}
		ctx := newActivityContext(map[string]interface{}{"payload": payload})
		_, err = activity.Eval(ctx)
		assert.Nil(t, err)

		value, attackValues := ctx.output["attack"].(float32), ctx.output["attackValues"].(map[string]interface{})
		if attack {
			assert.Condition(t, func() (success bool) {
				return value > 50
			}, fmt.Sprint("should be an attack", a, value))
			assert.Condition(t, func() (success bool) {
				return attackValues["content"].(map[string]interface{})["test"].(float64) > 50
			}, fmt.Sprint("should be an attack", a, value))
		} else {
			assert.Condition(t, func() (success bool) {
				return value < 50
			}, fmt.Sprint("should not be an attack", a, value))
			assert.Condition(t, func() (success bool) {
				return attackValues["content"].(map[string]interface{})["test"].(float64) < 50
			}, fmt.Sprint("should not be an attack", a, value))
		}
	}
	test("test or 1337=1337 --\"", true)
	test(" or 1=1 ", true)
	test("/**/or/**/1337=1337", true)
	test("abc123", false)
	test("abc123 123abc", false)
	test("123", false)
	test("abcorabc", false)
	test("available", false)
	test("orcat1", false)
	test("cat1or", false)
	test("cat1orcat1", false)
}
