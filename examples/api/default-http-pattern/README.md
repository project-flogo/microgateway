# Gateway using Default Http Pattern
This recipe is a gateway using the defult http pattern which uses JWT, Rate Limiter, and Circuit Breaker.

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
