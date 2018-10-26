# Gateway with a circuit breaker
This recipe is a gateway with a service protected by a circuit breaker.

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd microgateway/activity/circuitbreaker/examples/json
```
Place the Downloaded Mashling-Gateway binary in circuit-breaker-gateway folder.

## Testing
Create the gateway:
```
flogo create -f flogo.json
cd flogo
flogo build
```

Start the gateway:
```
bin/flogo
```
and test below scenario.

### Circuit breaker gets tripped
Start the gateway target service in a new terminal:
```
go run server/main.go -server
```

Now run the following in a new terminal:
```
curl http://localhost:9096/pets/1
```

You should see the following response:
```json
{
 "pet": {
  "category": {
   "id": 0,
   "name": "string"
  },
  "id": 1,
  "name": "sally",
  "photoUrls": [
   "string"
  ],
  "status": "available",
  "tags": [
   {
    "id": 0,
    "name": "string"
   }
  ]
 },
 "status": "available"
}
```
The target service is in a working state.

Now simulate a service failure by stopping the target service, and then run the following command 6 times:
```
curl http://localhost:9096/pets/1
```

You should see the below response 5 times:
```json
{
 "error": "connection failure"
}
```
Followed by:
```json
{
 "error": "circuit breaker tripped"
}
```
The circuit breaker is now in the tripped state.

Start the gateway target service back up and wait at least one minute. After waiting at least one minute run the following command:
```
curl http://localhost:9096/pets/1
```

You should see the following response:
```json
{
 "pet": {
  "category": {
   "id": 0,
   "name": "string"
  },
  "id": 1,
  "name": "sally",
  "photoUrls": [
   "string"
  ],
  "status": "available",
  "tags": [
   {
    "id": 0,
    "name": "string"
   }
  ]
 },
 "status": "available"
}
```
The circuit breaker is no longer in the tripped state, and the target service is working.
