# Gateway with Javascript
This recipe is a gateway which runs some javascript.

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd microgateway/activity/js/examples/api
```

## Testing
Create the gateway:
```
flogo create -f flogo.json
cd MyProxy
flogo build
```

Start the gateway:
```
bin/MyProxy
```

Run the following command:
```
curl http://localhost:9096/calculate"
```

You should see the following like response:
```json
{"sum":3}
```
