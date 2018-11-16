[![godoc](https://godoc.org/github.com/project-flogo/microgateway?status.svg)](https://godoc.org/github.com/project-flogo/microgateway)

# Microgateway Action

This is a Microgateway Action which supports the conditional evaluation of activities. The Microgateway has one setting: 'uri' which is the URI of the Microgateway JSON resource.

# Quick Start

## With Flogo CLI

The Flogo CLI takes a Flogo application defined in JSON and produces an executable application. The Flogo CLI can be installed from [here](https://github.com/project-flogo/cli). Next, follow the instructions [here](examples/json/basic-gateway) to build your first Microgateway Flogo application.

## With Flogo API

The Flogo API allows developers to define Flogo applications in the Go programming language without using the Flogo CLI. You can get started by cloning this repo:

```bash
git clone https://github.com/project-flogo/microgateway.git
```

and then following the instructions [here](examples/api/basic-gateway) to build your first Microgateway Flogo application. Documentation for the Flogo Microgateway API can be found [here](https://godoc.org/github.com/project-flogo/microgateway/api).

# Resource

## Schema

The JSON Schema for the Microgateway resource can be found [here](internal/schema/schema.json).

## Sections

### Services

A service defines a function or activity of some sort that will be utilized in a step within an execution flow. Services have names, refs, and settings. Any Flogo activity works as a service. Services that are specific to a microgateway can be found [here](activity). Services may call external endpoints like HTTP servers or may stay within the context of the gateway, like the [js](activity/circuitbreaker) activity. Once a service is defined it can be used as many times as needed within your routes and steps.

A service definition looks like:

```json
{
  "name": "PetStorePets",
  "description": "Get pets by ID from the petstore",
  "ref": "github.com/project-flogo/contrib/activity/rest",
  "settings": {
    "uri": "http://petstore.swagger.io/v2/pet/:petId",
    "method": "GET"
  }
}
```

### Steps

Each microgateway is composed of a number of steps. Each step is evaluated in the order in which it is defined via an optional `if` condition. If the condition is `true`, that step is executed. If that condition is `false` the execution context moves onto the next step in the process and evaluates that one. A blank or omitted `if` condition always evaluates to `true`.

A simple step looks like:

```json
{
  "if": "$.payload.pathParams.petId == 9",
  "service": "PetStorePets",
  "input": {
    "method": "GET",
    "pathParams.id": "=$payload.pathParams.petId"
  }
}
```

As you can see above, a step consists of a simple condition, a service reference, input parameters, and (not shown) output parameters. The `service` must map to a service defined in the `services` array that is defined in the microgateway resource. Input key and value pairs are translated and handed off to the service execution. Output key value pairs are translated and retained after the service has executed. Values starting with `=` are evaluated as variables within the context of the execution. An optional `halt` condition is supported for steps. When the `halt` condition is true the execution of the steps is halted.

### Responses

Each microgateway has an optional set of responses that can be evaluated and returned to the invoking trigger. Much like routes, the first response with an `if` condition evaluating to true is the response that gets executed and returned. A response contains an `if` condition, an `error` boolean, a `code` value, and a `data` object. The `error` boolean dictates whether or not an error should be returned to the engine. The `code` is the status code returned to the trigger. The `data` object is evaluated within the context of the execution and then sent back to the trigger as well.

A simple response looks like:

```json
{
  "if": "$.PetStorePets.outputs.data.status == 'available'",
  "error": false,
  "code": 200,
  "data": {
    "body.pet": "=$.PetStorePets.outputs.data",
    "body.inventory": "=$.PetStoreInventory.outputs.data"
  }
}
```

## Example Flogo JSON Usage of a Microgateway Action

An example of a basic gateway can be found [here](examples/json/basic-gateway).

## Example Flogo API Usage of a Microgateway Action

An API example can be found [here](examples/api/basic-gateway).
