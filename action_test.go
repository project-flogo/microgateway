package microgateway

import (
	"context"
	"testing"

	"github.com/project-flogo/contrib/activity/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/microgateway/internal/testing/activity"
	"github.com/project-flogo/microgateway/internal/testing/trigger"
	"github.com/project-flogo/microgateway/types"
	"github.com/stretchr/testify/assert"
)

func TestMicrogateway(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()
	app := api.NewApp()

	microgateway := types.New("test")
	service := microgateway.NewService("test", &activity.Activity{})
	service.SetDescription("A test activity")
	service.AddSetting("message", "hello world")
	step := microgateway.NewStep(service)
	step.SetIf("1 == 1")
	step.AddInput("message", "=1337")
	response := microgateway.NewResponse(false)
	response.SetCode("=200")
	response.SetData(map[string]interface{}{
		"test": "=$.test.outputs.data",
		"foo":  "bar",
		"bar":  1,
	})
	settings, err := microgateway.AddResource(app)
	assert.Nil(t, err)

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, settings)
	assert.Nil(t, err)
	action.SetCondition(`$.content.a == "b"`)

	defaultActionHit := false
	action, err = handler.NewAction(func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
		defaultActionHit = true
		return nil, nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()

	result, err := trigger.Fire(0, map[string]interface{}{"a": "c"})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, "", activity.Message)
	assert.False(t, activity.HasEvaled)
	assert.True(t, defaultActionHit)
	defaultActionHit = false

	result, err = trigger.Fire(0, map[string]interface{}{"a": "b"})
	assert.Nil(t, err)
	assert.Equal(t, 200, result["code"])
	assert.Equal(t, "1337", result["data"].(map[string]interface{})["test"])
	assert.Equal(t, "bar", result["data"].(map[string]interface{})["foo"])
	assert.Equal(t, 1.0, result["data"].(map[string]interface{})["bar"])
	assert.Equal(t, "1337", activity.Message)
	assert.True(t, activity.HasEvaled)
	assert.False(t, defaultActionHit)
}

func TestMicrogatewayHalt(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()
	app := api.NewApp()

	microgateway := types.New("halt")
	serviceHalt := microgateway.NewService("halt", &rest.Activity{})
	serviceHalt.SetDescription("An activity that will halt")
	serviceHalt.AddSetting("uri", "http://localhost:1234/abc123")
	serviceHalt.AddSetting("method", "GET")
	serviceTest := microgateway.NewService("test", &activity.Activity{})
	serviceTest.SetDescription("A test activity")
	serviceTest.AddSetting("message", "hello world")
	step := microgateway.NewStep(serviceHalt)
	step.SetHalt("($.halt.error != nil) && !error.isneterror($.halt.error)")
	step = microgateway.NewStep(serviceTest)
	assert.NotNil(t, step)
	response := microgateway.NewResponse(true)
	response.SetCode("=403")
	response.SetData(map[string]interface{}{
		"isneterror": "=error.isneterror($.halt.error)",
		"error":      "=error.string($.halt.error)",
	})
	response = microgateway.NewResponse(false)
	response.SetCode("=200")
	response.SetData(map[string]interface{}{
		"message": "hello world",
	})
	settings, err := microgateway.AddResource(app)
	assert.Nil(t, err)

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, settings)
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()

	result, err := trigger.Fire(0, map[string]interface{}{})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, true, result["data"].(map[string]interface{})["isneterror"])
	assert.True(t, activity.HasEvaled)
}