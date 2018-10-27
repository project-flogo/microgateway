package main

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/activity/anomaly"
	microapi "github.com/project-flogo/microgateway/api"
)

func main() {
	app := api.NewApp()

	gateway := microapi.New("Test")
	serviceAnomaly := gateway.NewService("Anomaly", &anomaly.Activity{})
	serviceAnomaly.SetDescription("Look for anomalies")
	serviceAnomaly.AddSetting("depth", 3)

	serviceUpdate := gateway.NewService("Update", &rest.Activity{})
	serviceUpdate.SetDescription("Make calls to test")
	serviceUpdate.AddSetting("uri", "http://localhost:1234/test")
	serviceUpdate.AddSetting("method", "PUT")

	step := gateway.NewStep(serviceAnomaly)
	step.AddInput("payload", "=$.payload.content")
	step = gateway.NewStep(serviceUpdate)
	step.SetIf("($.Anomaly.outputs.count < 100) || ($.Anomaly.outputs.complexity < 3)")
	step.AddInput("content", "=$.payload.content")

	response := gateway.NewResponse(false)
	response.SetIf("($.Anomaly.outputs.count < 100) || ($.Anomaly.outputs.complexity < 3)")
	response.SetCode(200)
	response.SetData("=$.Update.outputs.result")
	response = gateway.NewResponse(true)
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"error":      "anomaly!",
		"complexity": "=$.Anomaly.outputs.complexity",
	})

	settings, err := gateway.AddResource(app)
	if err != nil {
		panic(err)
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "PUT",
		Path:   "/test",
	})
	if err != nil {
		panic(err)
	}

	_, err = handler.NewAction(&microgateway.Action{}, settings)
	if err != nil {
		panic(err)
	}

	e, err := api.NewEngine(app)
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
