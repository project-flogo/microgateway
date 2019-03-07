# Gateway using File Resource
This recipe is a gateway using the file resource.

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```bash
git clone https://github.com/project-flogo/microgateway
cd microgateway/examples/json/resource-handler/fileResource
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


### Request is successful
```bash
curl http://localhost:9096/pets/1
```

You should then see something like:
```json
{"category":{"id":0,"name":"string"},"id":1,"name":"aspen","photoUrls":["string"],"status":"done","tags":[{"id":0,"name":"string"}]}