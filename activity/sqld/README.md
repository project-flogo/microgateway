# SQL Detector

The `sqld` service type implements SQL injection attack detection. Regular expressions and a [GRU](https://en.wikipedia.org/wiki/Gated_recurrent_unit) recurrent neural network are used to detect SQL injection attacks.

The available service `settings` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| file | string | An optional file name for custom neural network weights |

The available `input` for the request are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| payload | JSON object | A payload to do SQL injection attack detection on |

The available response `outputs` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| attack | number | The probability that the payload is a SQL injection attack |
| attackValues | JSON object | The SQL injection attack probability for each string in the payload |

A sample `service` definition is:

```json
{
  "name": "SQLSecurity",
  "description": "Look for sql injection attacks",
  "ref": "github.com/project-flogo/microgateway/activity/sqld"
}
```

An example `step` that invokes the above `SQLSecurity` service using `payload` is:

```json
{
  "service": "SQLSecurity",
  "input": {
    "payload": "=$.payload"
  }
}
```

Utilizing the response values can be seen in a response handler:

```json
{
  "if": "$.SQLSecurity.outputs.attack > 80",
  "error": true,
  "output": {
    "code": 403,
    "data": {
      "error": "hack attack!",
      "attackValues": "=$.SQLSecurity.outputs.attackValues"
    }
  }
}
```
