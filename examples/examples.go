package examples

import (
	"github.com/project-flogo/contrib/activity/log"
	"github.com/project-flogo/contrib/activity/rest"
	channeltrigger "github.com/project-flogo/contrib/trigger/channel"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/engine/channels"
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

// DefaultHTTPPattern returns an engine configured for the DefaultHttpPattern
func DefaultHTTPPattern() (engine.Engine, error) {
	app := api.NewApp()

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/endpoint",
	})
	if err != nil {
		panic(err)
	}

	_, err = handler.NewAction(&microgateway.Action{}, map[string]interface{}{
		"pattern":           "DefaultHttpPattern",
		"useRateLimiter":    true,
		"rateLimit":         "1-S",
		"useJWT":            true,
		"jwtSigningMethod":  "HMAC",
		"jwtKey":            "qwertyuiopasdfghjklzxcvbnm789101",
		"jwtAud":            "www.mashling.io",
		"jwtIss":            "Mashling",
		"jwtSub":            "tempuser@mail.com",
		"useCircuitBreaker": true,
		"backendUrl":        "http://localhost:1234/pets",
		"mode":              "a",
		"threshold":         5,
		"timeout":           60,
		"period":            60,
		"method":            "GET",
		"content":           "",
	})
	if err != nil {
		panic(err)
	}

	return api.NewEngine(app)
}

func DefaultChannelPattern() (engine.Engine, error) {
	app := api.NewApp()

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/endpoint",
	})
	if err != nil {
		panic(err)
	}

	_, err = handler.NewAction(&microgateway.Action{}, map[string]interface{}{
		"pattern":          "DefaultChannelPattern",
		"useJWT":           true,
		"jwtSigningMethod": "HMAC",
		"jwtKey":           "qwertyuiopasdfghjklzxcvbnm789101",
		"jwtAud":           "www.mashling.io",
		"jwtIss":           "Mashling",
		"jwtSub":           "tempuser@mail.com",
		"channel":          "test",
		"value":            "test",
	})
	if err != nil {
		panic(err)
	}

	// channel
	_, err = channels.New("test", 5)
	if err != nil {
		panic(err)
	}

	gateway := microapi.New("Log")
	service := gateway.NewService("log", &log.Activity{})
	service.SetDescription("Invoking test Log service")
	step := gateway.NewStep(service)
	step.AddInput("message", "Output: Test log message service invoked")
	response := gateway.NewResponse(false)
	response.SetCode(200)

	settings, err := gateway.AddResource(app)
	if err != nil {
		panic(err)
	}

	channeltrg := app.NewTrigger(&channeltrigger.Trigger{}, nil)
	channelhandler, err := channeltrg.NewHandler(&channeltrigger.HandlerSettings{
		Channel: "test",
	})
	if err != nil {
		panic(err)
	}

	_, err = channelhandler.NewAction(&microgateway.Action{}, settings)
	if err != nil {
		panic(err)
	}

	return api.NewEngine(app)
}

// CustomPattern returns an engine configured for given pattern name
func CustomPattern(patternName string, custompattern string) (engine.Engine, error) {
	err := microgateway.Register(patternName, custompattern)
	if err != nil {
		panic(err)
	}
	app := api.NewApp()

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/endpoint",
	})
	if err != nil {
		panic(err)
	}

	_, err = handler.NewAction(&microgateway.Action{}, map[string]interface{}{
		"pattern": patternName,
	})
	if err != nil {
		panic(err)
	}

	return api.NewEngine(app)
}

func AsyncGatewayExample() (engine.Engine, error) {
	app := api.NewApp()
	gateway := microapi.New("Log")
	service := gateway.NewService("log", &log.Activity{})
	service.SetDescription("Invoking test Log service in async gateway")
	step := gateway.NewStep(service)
	step.AddInput("message", "Output: Test log message service invoked")
	response := gateway.NewResponse(false)
	response.SetCode(200)
	response.SetData("Successful call to action")
	settings, err := gateway.AddResource(app)
	settings["async"] = true
	if err != nil {
		return nil, err
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/endpoint",
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

// ResourceHandlerGateway :- read resource from file system
func FileResourceHandlerExample() (engine.Engine, error) {
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

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/pets/:petId",
	})
	if err != nil {
		return nil, err
	}

	_, err = handler.NewAction(&microgateway.Action{}, map[string]interface{}{
		"uri": "file:///Users/agadikar/microgateway/examples/json/resource-handler/fileResource/resource.json"})
	if err != nil {
		return nil, err
	}

	return api.NewEngine(app)
}

// ResourceHandlerGateway :- getting Http resource
func HTTPResourceHandlerExample() (engine.Engine, error) {
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

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/pets/:petId",
	})
	if err != nil {
		return nil, err
	}

	_, err = handler.NewAction(&microgateway.Action{}, map[string]interface{}{
		"uri": "http://localhost:1234/pets"})
	if err != nil {
		return nil, err
	}

	return api.NewEngine(app)
}
