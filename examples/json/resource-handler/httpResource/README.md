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
```
flogo create -f flogo.json
cd MyProxy
flogo install github.com/project-flogo/contrib/activity/rest
flogo build
```

Start the gateway:
```
bin/MyProxy
```
and test below scenario.

In another terminal start the server:
```bash
go run main.go -server
```

### Request is successful
```bash
curl http://localhost:9096/pets/1
```

You should then see something like:
```json
{"category":{"id":0,"name":"string"},"id":1,"name":"aspen","photoUrls":["string"],"status":"done","tags":[{"id":0,"name":"string"}]}