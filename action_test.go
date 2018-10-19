package microgateway

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/project-flogo/core/api"
	_ "github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/microgateway/testing/activity"
	"github.com/project-flogo/microgateway/testing/trigger"
	"github.com/stretchr/testify/assert"
)

var microgatewayDefinition = `{
  "dispatch": {
    "name": "Test",
    "routes": [
      {
        "steps": [
					{
						"service": "test-activity",
						"input": {}
					}
				],
        "responses": [
					{
						"error": false,
						"output": {
							"code": 200,
							"data": "test"
						}
					}
				]
      }
    ]
  },
  "services": [
		{
			"name": "test-activity",
			"description": "A test activity",
			"ref": "github.com/project-flogo/microgateway/testing/activity",
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
	action, err := handler.NewAction(&MashlingAction{}, settings)
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

	err = trigger.Fire(0, map[string]interface{}{"a": "c"})
	assert.Nil(t, err)
	assert.False(t, activity.HasEvaled)
	assert.True(t, defaultActionHit)
	defaultActionHit = false

	err = trigger.Fire(0, map[string]interface{}{"a": "b"})
	assert.Nil(t, err)
	assert.True(t, activity.HasEvaled)
	assert.False(t, defaultActionHit)
}
