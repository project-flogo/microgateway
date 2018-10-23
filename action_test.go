package microgateway

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/project-flogo/core/api"
	"github.com/project-flogo/microgateway/internal/testing/activity"
	"github.com/project-flogo/microgateway/internal/testing/trigger"
	"github.com/stretchr/testify/assert"
)

var microgatewayDefinition = `{
	"name": "Test",
  "steps": [
		{
			"if": "1 == 1",
			"service": "test-activity",
			"input": {
				"message": "=1337"
			}
		}
	],
  "responses": [
		{
			"error": false,
			"output": {
				"code": "=200",
				"data": {
					"test": "=$.test-activity.outputs.data",
					"foo": "bar",
					"bar": 1
				}
			}
		}
	],
  "services": [
		{
			"name": "test-activity",
			"description": "A test activity",
			"ref": "github.com/project-flogo/microgateway/internal/testing/activity",
			"settings": {
				"message": "hello world"
			}
		}
	]
}`

func TestMicrogateway(t *testing.T) {
	app := api.NewApp()
	app.AddResource("microgateway:Test", json.RawMessage(microgatewayDefinition))
	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	settings := map[string]interface{}{
		"uri": "microgateway:Test",
	}
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
