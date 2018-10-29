package main

import (
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/activity/js"
	microapi "github.com/project-flogo/microgateway/api"
)

func main() {
	app := api.NewApp()

	gateway := microapi.New("JS")

	service := gateway.NewService("JS", &js.Activity{})
	service.SetDescription("Calculate sum")
	service.AddSetting("script", "result.sum = parameters.a + parameters.b")

	step := gateway.NewStep(service)
	step.AddInput("parameters", map[string]interface{}{"a": 1.0, "b": 2.0})

	response := gateway.NewResponse(false)
	response.SetCode(200)
	response.SetData("=$.JS.outputs.result")
	settings, err := gateway.AddResource(app)
	if err != nil {
		panic(err)
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/calculate",
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
