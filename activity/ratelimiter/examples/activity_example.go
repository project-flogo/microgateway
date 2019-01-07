package examples

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/activity/ratelimiter"
	microapi "github.com/project-flogo/microgateway/api"
)

// Example returns an API example
func Example(limit string) (engine.Engine, error) {
	app := api.NewApp()

	gateway := microapi.New("Pets")

	serviceLimiter := gateway.NewService("RateLimiter", &ratelimiter.Activity{})
	serviceLimiter.SetDescription("Rate limiter")
	serviceLimiter.AddSetting("limit", limit)

	serviceStore := gateway.NewService("PetStorePets", &rest.Activity{})
	serviceStore.SetDescription("Get pets by ID from the petstore")
	serviceStore.AddSetting("uri", "http://petstore.swagger.io/v2/pet/:petId")
	serviceStore.AddSetting("method", "GET")
	serviceStore.AddSetting("headers", map[string]string{
		"Accept": "application/json",
	})

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
	response.SetData("=$.PetStorePets.outputs.data")

	settings, err := gateway.AddResource(app)
	if err != nil {
		return nil, err
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/pets/:petId",
	})
	if err != nil {
		return nil, err
	}

	_, err = handler.NewAction(&microgateway.Action{}, settings)
	if err != nil {
		return nil, err
	}

	return api.NewEngine(app)
}
