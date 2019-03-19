# Gateway using Default Channel Pattern
This recipe is a gateway using the default channel pattern which uses JWT.

# Channel Activity
| Name    |  Type  | Description                     |
|:--------|:-------|:--------------------------------|
| channel | string | The channel to put the value on |
| value   | string | The value to put on channel     |

# JWT
| Name          |  Type  | Description                                        |
|:--------------|:-------|:---------------------------------------------------|
| token         | string | The raw token                                      |
| key           | string | The key used to sign the token                     |
| signingMethod | string | The signing method used (HMAC, ECDSA, RSA, RSAPSS) |
| issuer        | string | The 'iss' standard claim to match against          |
| subject       | string | The 'sub' standard claim to match against          |
| audience      | string | The 'aud' standard claim to match against          |


## Installation
* Install [Go](https://golang.org/)

## Setup
```bash
git clone https://github.com/project-flogo/microgateway
cd microgateway/examples/api/default-http-pattern
```

## Testing
Start the gateway:
```bash
go run main.go
```


### Request is successful
Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDEzMTYxOCwiZXhwIjoxNTc1NjY3NjE4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSJ9.wgunWSIJqieRKsmObATT2VEHMMzkKte6amuUlhc1oKs"
```

You should see:
```json
{"response":"Success!"}
```
On the server screen, you get 200 response code and log service outputs "Output: Test log message service invoked"


### JWT token is invalid
Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDEzMTYxOCwiZXhwIjoxNTc1NjY3NjE4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSJ9.wgunWSIJqieRKsmObATT2VEHMMzkKte6amuUlhc1oK"
```

You should see:
```json
{"errorMessage":"","validationMessage":"signature is invalid"}
```
