package main

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/activity/ratelimiter"
	microapi "github.com/project-flogo/microgateway/api"
)

func main() {
	app := api.NewApp()

	gateway := microapi.New("Pets")

	serviceLimiter := gateway.NewService("RateLimiter", &ratelimiter.Activity{})
	serviceLimiter.SetDescription("Rate limiter")
	serviceLimiter.AddSetting("limit", "3-M")

	serviceStore := gateway.NewService("PetStorePets", &rest.Activity{})
	serviceStore.SetDescription("Get pets by ID from the petstore")
	serviceStore.AddSetting("uri", "http://petstore.swagger.io/v2/pet/:petId")
	serviceStore.AddSetting("method", "GET")

	step := gateway.NewStep(serviceLimiter)
	step.AddInput("token", "=$.payload.headers.Token")
	step = gateway.NewStep(serviceStore)
	step.AddInput("pathParams", "=$.payload.pathParams")

	response := gateway.NewResponse(true)
	response.SetIf("$.RateLimiter.outputs.error == true")
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"status": "=$.RateLimiter.outputs.errorMessage",
	})
	response = gateway.NewResponse(true)
	response.SetIf("$.RateLimiter.outputs.limitReached == true")
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"status": "Rate Limit Exceeded - The service you have requested is over the allowed limit.",
	})
	response = gateway.NewResponse(false)
	response.SetCode(200)
	response.SetData("=$.PetStorePets.outputs.result")

	settings, err := gateway.AddResource(app)
	if err != nil {
		panic(err)
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/pets/:petId",
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
