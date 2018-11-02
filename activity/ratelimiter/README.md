# Rate Limiter

The `ratelimiter` service type creates a rate limiter with specified `limit`. When it is used in the `step`, it applies `limit` against supplied `token`.

The available service `settings` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| limit | string | Limit can be specifed in the format of "limit-period". Valid periods are 'S', 'M' & 'H' to represent Second, Minute & Hour. Example: "10-S" represents 10 request/second |

The available `input` for the request are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| token | string | Token for which rate limit has to be applied |

The available response `outputs` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| limitReached | bool | If the limit exceeds |
| limitAvailable | integer | Available limit |
| error | bool | If any error occured while applying the rate limit |
| errorMessage | string | The error message |

A sample `service` definition is:

```json
{
    "name": "RateLimiter",
    "description": "Rate limiter",
    "ref": "github.com/project-flogo/microgateway/activity/ratelimiter",
    "settings": {
        "limit": "5-M"
    }
}
```

An example `step` that invokes the above `ratelimiter` service to consume a `token` is:
```json
{
    "service": "RateLimiter",
    "input": {
        "token": "=$.payload.headers.Token"
    }
}
```
Note: When `token` is not supplied or empty, service sets `error` to true. This can be handled by configuring `token` to some constant value, In this way service can be operated as global rate limiter. An example shown below:

```json
{
    "service": "RateLimiter",
    "input": {
        "token": "MY_GLOBAL_TOKEN"
    }
}
```

Utilizing and extracting the response values can be seen in both a conditional evaluation:
```json
{"if": "$.RateLimiter.outputs.limitReached == true"}
```
and a response handler:
```json
{
    "if": "$.RateLimiter.outputs.limitReached == true",
    "error": true,
    "output": {
        "code": 403,
        "data": {
            "status":"Rate Limit Exceeded - The service you have requested is over the allowed limit."
        }
    }
}
```
