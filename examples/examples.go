package examples

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	microapi "github.com/project-flogo/microgateway/api"
)

// BasicGatewayExample returns a Basic Gateway API example
func BasicGatewayExample() (engine.Engine, error) {
	app := api.NewApp()

	gateway := microapi.New("Pets")
	service := gateway.NewService("PetStorePets", &rest.Activity{})
	service.SetDescription("Get pets by ID from the petstore")
	service.AddSetting("uri", "http://petstore.swagger.io/v2/pet/:petId")
	service.AddSetting("method", "GET")
	service.AddSetting("headers", map[string]string{
		"Accept": "application/json",
	})
	step := gateway.NewStep(service)
	step.AddInput("pathParams", "=$.payload.pathParams")
	response := gateway.NewResponse(false)
	response.SetCode(200)
	response.SetData("=$.PetStorePets.outputs.data")
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
		return nil, err
	}

	_, err = handler.NewAction(&microgateway.Action{}, settings)
	if err != nil {
		return nil, err
	}

	return api.NewEngine(app)
}

// HandlerRoutingExample returns a Handler Routing API example
func HandlerRoutingExample() (engine.Engine, error) {
	app := api.NewApp()

	gateway := microapi.New("Pets")
	service := gateway.NewService("PetStorePets", &rest.Activity{})
	service.SetDescription("Get pets by ID from the petstore")
	service.AddSetting("uri", "http://petstore.swagger.io/v2/pet/:petId")
	service.AddSetting("method", "GET")
	service.AddSetting("headers", map[string]string{
		"Accept": "application/json",
	})

	step := gateway.NewStep(service)
	step.SetIf("string.integer($.payload.pathParams.petId) < 8")
	step.AddInput("pathParams", "=$.payload.pathParams")

	response := gateway.NewResponse(false)
	response.SetIf("string.integer($.payload.pathParams.petId) < 8")
	response.SetCode(200)
	response.SetData("=$.PetStorePets.outputs.data")
	response = gateway.NewResponse(false)
	response.SetCode(404)
	response.SetData(map[string]interface{}{
		"error": "id must be less than 8",
	})

	settings, err := gateway.AddResource(app)
	if err != nil {
		return nil, err
	}

	gatewayAuthed := microapi.New("PetsAuthed")
	service = gatewayAuthed.NewService("PetStorePets", &rest.Activity{})
	service.SetDescription("Get pets by ID from the petstore")
	service.AddSetting("uri", "http://petstore.swagger.io/v2/pet/:petId")
	service.AddSetting("method", "GET")
	service.AddSetting("headers", map[string]string{
		"Accept": "application/json",
	})

	step = gatewayAuthed.NewStep(service)
	step.AddInput("pathParams", "=$.payload.pathParams")

	response = gatewayAuthed.NewResponse(false)
	response.SetCode(200)
	response.SetData("=$.PetStorePets.outputs.data")

	settingsAuthed, err := gatewayAuthed.AddResource(app)
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

	action, err := handler.NewAction(&microgateway.Action{}, settingsAuthed)
	if err != nil {
		return nil, err
	}
	action.SetCondition(`$.headers.Auth == "1337"`)

	_, err = handler.NewAction(&microgateway.Action{}, settings)
	if err != nil {
		return nil, err
	}

	return api.NewEngine(app)
}
