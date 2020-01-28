# Gateway for GraphQL endpoints

This is a gateway application demonstrates how following polices can be applied for a GraphQL endpoint:
* validate GraphQL request against schema
* configure Maximum Query Depth

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd activity/graphql/examples/json
```

## Testing
Create the gateway:
```
flogo create -f flogo.json
cd MyProxy
flogo build
cd bin
cp ../../../schema.graphql .
```

Start the gateway:
```
./MyProxy
```
and test below scenarios:

### Validated graphql request against schema

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
curl -X POST -H 'Content-Type: application/json' --data-binary '{"query":"query {stationWithEvaId(evaId: 8000105) { name } }"}' 'localhost:9096/graphql'

```
expected response:
```json
{
    "response": "{\"data\":{\"stationWithEvaId\":{\"name\":\"Frankfurt (Main) Hbf\"}}}",
    "validationMessage": "Valid graphQL query. query = {\"query\":\"query {stationWithEvaId(evaId: 8000105) { name } }\"}\n type = Query \n queryDepth = 2"
}
```

* invalid request

```sh
query {
    stationWithEvaId(evaId: 8000105) { 
        cityname
    }
}
```
curl request:
```sh
curl -X POST -H 'Content-Type: application/json' --data-binary '{"query":"query {stationWithEvaId(evaId: 8000105) { cityname } }"}' 'localhost:9096/graphql'

```
expected response:
```json
{"error":"Not a valid graphQL request. Details: [graphql: Cannot query field \"cityname\" on type \"Station\". (line 1, column 43)]"}
```

### Validated graphql request against maxQueryDepth

* request with query depth 3
```sh
query {                                     #depth 0
    stationWithEvaId(evaId: 8000105) {      #depth 1
        name                                #depth 2 
        location {
            latitude                        #depth 3 
            longitude
        } 
        picture { 
            url                             #depth 3 
        }
    }
}
```

curl request:
```sh
curl -X POST -H 'Content-Type: application/json' --data-binary '{"query":"{stationWithEvaId(evaId: 8000105) {name location { latitude longitude } picture { url } } }"}' 'localhost:9096/graphql'

```
expected response:
```json
{"error":"graphQL request query depth[3] is exceeded allowed maxQueryDepth[2]"}
```

### Courtesy
This example uses publicly available `graphql` endpoint [bahnql.herokuapp.com/graphql](https://bahnql.herokuapp.com/graphql)
