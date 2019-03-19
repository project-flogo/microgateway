# Gateway using HTTP Resource
This recipe is a gateway using the HTTP resource. The resource is downloaded from the  requested HTTP server

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```bash
git clone https://github.com/project-flogo/microgateway
cd microgateway/examples/json/resource-handler/httpResource
```

## Testing
Create the gateway:
```bash
flogo create -f flogo.json
cd MyProxy
flogo install github.com/project-flogo/contrib/activity/rest
flogo build
```

In another terminal start the server first:
```bash
go run main.go -server
```

Start the gateway:
```bash
bin/MyProxy
```
and test below scenario.

### Request is successful
```bash
curl http://localhost:9096/pets/1
```

You should then see something like:
```json
{"category":{"id":0,"name":"string"},"id":4,"name":"hc0x3yiw302","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}
```
