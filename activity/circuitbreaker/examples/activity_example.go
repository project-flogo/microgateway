package examples

import (
	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/activity/circuitbreaker"
	microapi "github.com/project-flogo/microgateway/api"
)

// Example returns an API example
func Example(mode string, threshold int, timeout int, period int) (engine.Engine, error) {
	app := api.NewApp()

	gateway := microapi.New("Pets")

	serviceCircuitBreaker := gateway.NewService("CircuitBreaker", &circuitbreaker.Activity{})
	serviceCircuitBreaker.SetDescription("Circuit breaker service")
	serviceCircuitBreaker.AddSetting("mode", mode)
	if threshold > 0 {
		serviceCircuitBreaker.AddSetting("threshold", threshold)
	}
	if timeout > 0 {
		serviceCircuitBreaker.AddSetting("timeout", timeout)
	}
	if period > 0 {
		serviceCircuitBreaker.AddSetting("period", period)
	}
	serviceStore := gateway.NewService("PetStorePets", &rest.Activity{})
	serviceStore.SetDescription("Get pets by ID from the petstore")
	serviceStore.AddSetting("uri", "http://localhost:1234/pets")
	serviceStore.AddSetting("method", "GET")

	gateway.NewStep(serviceCircuitBreaker)
	step := gateway.NewStep(serviceStore)
	step.SetHalt("($.PetStorePets.error != nil) && !error.isneterror($.PetStorePets.error)")
	step = gateway.NewStep(serviceCircuitBreaker)
	step.SetIf("$.PetStorePets.error != nil")
	step.AddInput("operation", "counter")
	step = gateway.NewStep(serviceCircuitBreaker)
	step.SetIf("$.PetStorePets.error == nil")
	step.AddInput("operation", "reset")

	response := gateway.NewResponse(true)
	response.SetIf("$.CircuitBreaker.outputs.tripped == true")
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"error": "circuit breaker tripped",
	})
	response = gateway.NewResponse(true)
	response.SetIf("$.PetStorePets.outputs.data.status != 'available'")
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"error":  "Pet is unavailable",
		"pet":    "=$.PetStorePets.outputs.data",
		"status": "=$.PetStorePets.outputs.data.status",
	})
	response = gateway.NewResponse(false)
	response.SetIf("$.PetStorePets.outputs.data.status == 'available'")
	response.SetCode(200)
	response.SetData(map[string]interface{}{
		"pet":    "=$.PetStorePets.outputs.data",
		"status": "=$.PetStorePets.outputs.data.status",
	})
	response = gateway.NewResponse(true)
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"error": "connection failure",
	})

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
