# Gateway using Default Http Pattern
This recipe is a gateway using the defult http pattern which uses JWT, Rate Limiter, and Circuit Breaker.

# Circuit Breaker Service:
| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| mode | string | The tripping mode: 'a' for contiguous errors, 'b' for errors within a time period, 'c' for contiguous errors within a time period, and 'd' for a probabilistic smart circuit breaker mode. Defaults to mode 'a' |
| threshold | number | The number of errors required for tripping|
| period | number | Number of seconds in which errors have to occur for the circuit breaker to trip. Applies to modes 'b' and 'c'|
| timeout | number | Number of seconds that the circuit breaker will remain tripped. Applies to modes 'a', 'b', 'c'|

# Rate Limiter
| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| limit | string | Limit can be specifed in the format of "limit-period". Valid periods are 'S', 'M' & 'H' to represent Second, Minute & Hour. Example: "10-S" represents 10 request/second |

# JWT
| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| token | string | The raw token |
| key | string | The key used to sign the token |
| signingMethod | string | The signing method used (HMAC, ECDSA, RSA, RSAPSS) |
| issuer | string | The 'iss' standard claim to match against |
| subject | string | The 'sub' standard claim to match against |
| audience | string | The 'aud' standard claim to match against |

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
and test below scenario.

In another terminal start the server:
```bash
go run main.go -server
```

### Request is successful
Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDEzMTYxOCwiZXhwIjoxNTc1NjY3NjE4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSJ9.wgunWSIJqieRKsmObATT2VEHMMzkKte6amuUlhc1oKs"
```

You should see:
```json
{"category":{"id":0,"name":"string"},"id":1,"name":"sally","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}
```

### JWT token is invalid
Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDEzMTYxOCwiZXhwIjoxNTc1NjY3NjE4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSJ9.wgunWSIJqieRKsmObATT2VEHMMzkKte6amuUlhc1oK"
```

You should see:
```json
{"errorMessage":"","validationMessage":"signature is invalid"}
```

### Rate limit is exceeded
Run the following command faster the 1 per second:
```bash
curl --request GET http://localhost:9096/endpoint -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDEzMTYxOCwiZXhwIjoxNTc1NjY3NjE4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSJ9.wgunWSIJqieRKsmObATT2VEHMMzkKte6amuUlhc1oKs"
```

You should see:
```json
{"status":"Rate Limit Exceeded - The service you have requested is over the allowed limit."}
```

### Circuit breaker tripped
Stop the server and run the following command 6 times:
```bash
curl --request GET http://localhost:9096/endpoint -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDEzMTYxOCwiZXhwIjoxNTc1NjY3NjE4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSJ9.wgunWSIJqieRKsmObATT2VEHMMzkKte6amuUlhc1oKs"
```

You should see:
```json
{"error":"circuit breaker tripped"}
```
