# Circuit Breaker

The circuit breaker prevents the calling of a service when that service has failed in the past. How the circuit breaker is tripped depends on the mode of operation. There are four modes of operation: contiguous errors, errors within a time period, contiguous errors within a time period, and smart circuit breaker mode.

The available service `settings` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| mode | string | The tripping mode: 'a' for contiguous errors, 'b' for errors within a time period, 'c' for contiguous errors within a time period, and 'd' for a probabilistic smart circuit breaker mode. Defaults to mode 'a' |
| threshold | number | The number of errors required for tripping. Defaults to 5 errors |
| period | number | Number of seconds in which errors have to occur for the circuit breaker to trip. Applies to modes 'b' and 'c'. Defaults to 60 seconds |
| timeout | number | Number of seconds that the circuit breaker will remain tripped. Applies to modes 'a', 'b', 'c'. Defaults to 60 seconds |

The available `input` for the request are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| operation | string | An operation to perform: '' for protecting a service, 'counter' for processing errors, and 'reset' for processing non-errors. Defaults to '' |

The available response `outputs` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| tripped | boolean | The state of the circuit breaker |

A sample `service` definition is:

```json
{
  "name": "CircuitBreaker",
  "description": "Circuit breaker service",
  "ref": "github.com/project-flogo/microgateway/activity/circuitbreaker",
  "settings": {
    "mode": "a"
  }
}
```

An example series of `step` that invokes the above `CircuitBreaker` service:

```json
{
  "service": "CircuitBreaker"
},
{
  "service": "PetStorePets",
  "input": {
    "method": "GET"
  },
  "halt": "($.PetStorePets.error != nil) && !error.isneterror($.PetStorePets.error)"
},
{
  "if": "$.PetStorePets.error != nil",
  "service": "CircuitBreaker",
  "input": {
    "operation": "counter"
  }
},
{
  "if": "$.PetStorePets.error == nil",
  "service": "CircuitBreaker",
  "input": {
    "operation": "reset"
  }
}
```

Utilizing the response values can be seen in a response handler:

```json
{
  "if": "$.CircuitBreaker.outputs.tripped == true",
  "error": true,
  "output": {
    "code": 403,
    "data": {
      "error": "circuit breaker tripped"
    }
  }
}
```
