# Gateway using File Resource
This recipe is a gateway using the file resource.

## Installation
* Install [Go](https://golang.org/)

## Setup
```bash
git clone https://github.com/project-flogo/microgateway
cd microgateway/examples/api/resource-handler/fileResource
```

## Testing

Start the gateway:
```bash
go run main.go
```
and test below scenario.

### Request is successful
Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint
```

You should see:
```json
{"category":{"id":0,"name":"string"},"id":4,"name":"hc0x3yiw302","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}
```