package example

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	microapi "github.com/project-flogo/microgateway/api"
)

// Example returns an API example
func Example(activity activity.Activity) (engine.Engine, error) {
	app := api.NewApp()
	gateway := microapi.New("JWT")

	jwtService := gateway.NewService("jwtService", activity)
	jwtService.SetDescription("Validate JWT")
	jwtService.AddSetting("signingMethod", "HMAC")
	jwtService.AddSetting("key", "qwertyuiopasdfghjklzxcvbnm789101")
	jwtService.AddSetting("aud", "www.mashling.io")
	jwtService.AddSetting("iss", "Mashling")
	jwtService.AddSetting("sub", "tempuser@mail.com")

	serviceStore := gateway.NewService("PetStorePets", &rest.Activity{})
	serviceStore.SetDescription("Get pets by ID from the petstore")
	serviceStore.AddSetting("uri", "https://petstore.swagger.io/v2/pet/:petId")
	serviceStore.AddSetting("method", "GET")

	step := gateway.NewStep(jwtService)
	step.AddInput("token", "=$.payload.headers.Authorization")
	step = gateway.NewStep(serviceStore)
	step.AddInput("pathParams.petId", "=$.jwtService.outputs.token.claims.id")

	response := gateway.NewResponse(false)
	response.SetIf("$.jwtService.outputs.valid == true")
	response.SetCode(200)
	response.SetData(map[string]interface{}{
		"error": "JWT token is valid",
		"pet":   "=$.PetStorePets.outputs.data",
	})
	response = gateway.NewResponse(true)
	response.SetIf("$.jwtService.outputs.valid == false")
	response.SetCode(401)
	response.SetData(map[string]interface{}{
		"error": "=$.jwtService.outputs",
		"pet":   "=$.PetStorePets.outputs.data",
	})

	settings, err := gateway.AddResource(app)
	if err != nil {
		return nil, err
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/pets",
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
