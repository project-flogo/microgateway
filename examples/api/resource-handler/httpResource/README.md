# Gateway using HTTP Resource
This recipe is a gateway using the HTTP resource. The resource is downloaded from the  requested HTTP server

## Installation
* Install [Go](https://golang.org/)

## Setup
```bash
git clone https://github.com/project-flogo/microgateway
cd microgateway/examples/api/resource-handler/httpResource
```

## Testing

In terminal start the server first:
```bash
go run main.go -server
```

Start the gateway:
```bash
go run main.go
```
and test below scenario.

### Request is successful
Run the following command:
```bash
curl http://localhost:9096/pets/1
```

You should see:
```json
{"category":{"id":0,"name":"string"},"id":4,"name":"hc0x3yiw302","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}
```
