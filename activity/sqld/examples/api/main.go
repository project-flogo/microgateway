package main

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/activity/sqld"
	microapi "github.com/project-flogo/microgateway/api"
)

func main() {
	app := api.NewApp()

	gateway := microapi.New("Update")
	serviceSQLD := gateway.NewService("SQLSecurity", &sqld.Activity{})
	serviceSQLD.SetDescription("Look for sql injection attacks")

	serviceUpdate := gateway.NewService("PetStorePetsUpdate", &rest.Activity{})
	serviceUpdate.SetDescription("Update pets")
	serviceUpdate.AddSetting("uri", "http://petstore.swagger.io/v2/pet")
	serviceUpdate.AddSetting("method", "PUT")

	step := gateway.NewStep(serviceSQLD)
	step.AddInput("payload", "=$.payload")
	step = gateway.NewStep(serviceUpdate)
	step.SetIf("$.SQLSecurity.outputs.attack < 80")
	step.AddInput("content", "=$.payload.content")

	response := gateway.NewResponse(false)
	response.SetIf("$.SQLSecurity.outputs.attack < 80")
	response.SetCode(200)
	response.SetData("=$.PetStorePetsUpdate.outputs.data")
	response = gateway.NewResponse(true)
	response.SetIf("$.SQLSecurity.outputs.attack > 80")
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"error":        "hack attack!",
		"attackValues": "=$.SQLSecurity.outputs.attackValues",
	})

	settings, err := gateway.AddResource(app)
	if err != nil {
		panic(err)
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "PUT",
		Path:   "/pets",
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
