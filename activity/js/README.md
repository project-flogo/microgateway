# JS

The `js` service type evaluates a javascript `script` along with provided `parameters` and returns the result as the response.

The available service `settings` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| script | string | The javascript code to evaluate |

The available `input` for the request are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| parameters | JSON object | Key/value pairs representing parameters to evaluate within the context of the script  |

The available response `outputs` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| error | bool | The HTTP status code of the response |
| errorMessage | string | The error message |
| result | JSON object | The result object from the javascript code  |

A sample `service` definition is:

```json
{
  "name": "JSCalc",
  "description": "Make calls to a JS calculator",
  "ref": "github.com/project-flogo/microgateway/activity/js",
  "settings": {
    "script": "result.total = parameters.num * 2;"
  }
}
```

An example `step` that invokes the above `JSCalc` service using `parameters` is:

```json
{
  "if": "$.PetStorePets.outputs.result.status == 'available'",
  "service": "JSCalc",
  "input": {
    "parameters.num": "=$.PetStorePets.outputs.result.available"
  }
}
```

Utilizing the response values can be seen in a response handler:

```json
{
  "if": "$.PetStorePets.outputs.result.status == 'available'",
  "error": false,
  "code": 200,
  "data": {
    "body.pet": "=$.PetStorePets.outputs.result",
    "body.inventory": "=$.PetStoreInventory.outputs.result",
    "body.availableTimesTwo": "=$.JSCalc.outputs.result.total"
  }
}
```
