# Async Gateway
This is an example of Async gateway. This receipe executes simple log action asynchronously


# Log Activity
| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| message | string | The message to log |
| addDetails | string | If set to true this will append the execution information to the log message |


## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```bash
git clone https://github.com/project-flogo/microgateway
cd microgateway/examples/api/default-http-pattern
```

## Testing
Create the gateway:
```bash
flogo create -f flogo.json
cd MyProxy
flogo install github.com/project-flogo/contrib/activity/log
flogo build
```

Start the gateway:
```bash
bin/MyProxy
```

and test below scenario.

Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint
```

You should see:
On the server screen, you get 200 response code and log service outputs "Output: Test log message service invoked"
