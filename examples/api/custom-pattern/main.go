package main

import (
	_ "github.com/project-flogo/contrib/activity/rest"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway/activity/circuitbreaker"
	_ "github.com/project-flogo/microgateway/activity/jwt"
	_ "github.com/project-flogo/microgateway/activity/ratelimiter"
	"github.com/project-flogo/microgateway/examples"
)

const pattern = `{
  "name": "CustomHttpPattern",
  "steps": [
    {
      "service": "RateLimiter",
      "input": {
        "token": "global"
      }
    },
    {
      "service": "JWTValidator",
      "input": {
        "token": "=$.payload.headers.Authorization",
        "signingMethod": "HMAC",
        "key": "qwertyuiopasdfghjklzxcvbnm789101",
        "aud": "www.mashling.io",
        "iss": "Mashling",
        "sub": "tempuser@mail.com"
      }
    },
    {
      "if": "$.JWTValidator.outputs.valid == true",
      "service": "HttpBackendA",
      "halt": "($.HttpBackendA.error != nil) && !error.isneterror($.HttpBackendA.error)"
    }
  ],
  "responses": [
    {
      "if": "$.RateLimiter.outputs.limitReached == true",
      "error": true,
      "output": {
        "code": 403,
        "data": {
          "status": "Rate Limit Exceeded - The service you have requested is over the allowed limit."
        }
      }
    },
    {
      "if": "$.JWTValidator.outputs.valid == false",
      "error": true,
      "output": {
        "code": 401,
        "data": {
          "errorMessage": "=$.JWTValidator.outputs.errorMessage",
          "validationMessage": "=$.JWTValidator.outputs.validationMessage"
        }
      }
    },
    {
      "if": "$.JWTValidator.outputs.valid == true",
      "error": false,
      "output": {
        "code": 200,
        "data": "=$.HttpBackendA.outputs.data"
      }
    },
    {
      "error": true,
      "output": {
        "code": 400,
        "data": "Error"
      }
    }
  ],
  "services": [
    {
      "name": "RateLimiter",
      "description": "Rate limiter",
      "ref": "github.com/project-flogo/microgateway/activity/ratelimiter",
      "settings": {
        "limit": "1-S"
      }
    },
    {
      "name": "JWTValidator",
      "description": "Validate some tokens",
      "ref": "github.com/project-flogo/microgateway/activity/jwt"
    },
    {
      "name": "HttpBackendA",
      "description": "Make an http call to your backend",
      "ref": "github.com/project-flogo/contrib/activity/rest",
      "settings": {
        "method": "GET",
        "uri": "http://localhost:1234/pets"
      }
    }
  ]
}
`

func main() {
	e, err := examples.CustomPattern("CustomHttpPattern", pattern)
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
