# Gateway for GraphQL endpoints

This is a gateway application demonstrates how following polices can be applied for a GraphQL endpoint:
* Throttle GraphQL request based on server time consumed

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd activity/graphql/examples/json-throttle-server-time
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
and test below scenarios:

### Validated graphql requests against defined quota

graphql service is configured with the below limit:
```sh
"settings": {
    "mode": "b",
    "limit": "1000-200-2000"
}
```
That means maximum allowed server time is set to 1000ms, client gains 200ms of server time per 2 sec (2000ms).

* valid request
```sh
query {
    stationWithEvaId(evaId: 8000105) { 
        name
    }
}
```
curl request:
```sh
curl -X POST -H 'Content-Type: application/json' -H 'Token: MY_TOKEN' --data-binary '{"query":"query {stationWithEvaId(evaId: 8000105) { name } }"}' 'localhost:9096/graphql'

```
expected response:
```json
{"response":"{\"data\":{\"stationWithEvaId\":{\"name\":\"Frankfurt (Main) Hbf\"}}}","validationMessage":null}
```

* Quota limit reached

Issue multiple requests so that entire quota is consumed. You would see below response:

```json
{"error":"Consumed entire Quota"}
```

### Courtesy
This example uses publicly available `graphql` endpoint [bahnql.herokuapp.com/graphql](https://bahnql.herokuapp.com/graphql)
