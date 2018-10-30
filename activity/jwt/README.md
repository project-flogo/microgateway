# JWT

The `jwt` service type accepts, parses, and validates JSON Web Tokens.

The service `settings` and available `input` for the request are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| token | string | The raw token |
| key | string | The key used to sign the token |
| signingMethod | string | The signing method used (HMAC, ECDSA, RSA, RSAPSS) |
| issuer | string | The 'iss' standard claim to match against |
| subject | string | The 'sub' standard claim to match against |
| audience | string | The 'aud' standard claim to match against |

The available response outputs are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| valid | boolean | If the token is valid or not |
| token | JSON object | The parsed token |
| validationMessage | string | The validation failure message |
| error | boolean | If an error occurred when parsing the token |
| errorMessage | string | The error message |

The parsed token contents are:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| claims | JSON object | The set of standard and custom claims provided by the parsed token |
| signature | string | The token's signature |
| signingMethod | string | The method used to sign the token |
| header | JSON object | An object containing header key value pairs for the parsed token  |

The `exp` and `iat` standard claims are automatically validated.

A sample `service` definition is:

```json
{
  "name": "JWTValidator",
  "description": "Validate a token",
  "ref": "github.com/project-flogo/microgateway/activity/jwt",
  "settings": {
    "signingMethod": "HMAC",
    "key": "qwertyuiopasdfghjklzxcvbnm123456",
    "audience": "www.mashling.io",
    "issuer": "Mashling"
  }
}
```

An example `step` that invokes the above `JWTValidator` service using a `token` from the header in an HTTP trigger is:

```json
{
  "service": "JWTValidator",
  "input": {
    "token": "=$.payload.headers.Authorization"
  }
}
```

Utilizing and extracting the response values can be seen in a conditional evaluation:

```json
{"if": "$.JWTValidator.outputs.valid == true"}
```

or to extract a value from the parsed claims you can use:
```
=$.jwtService.outputs.token.claims.<custom-claim-key>
```
