# Gateway with SQL injection attack defense
This recipe is a gateway with SQL injection attack defense.

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd microgateway/activity/sqld/examples/api
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
and test below scenarios.

### Payload without SQL injection attack
Run the following command:
```
curl http://localhost:9096/pets --upload-file payload.json
```

You should see the following response:
```json
{
 "category": {
  "id": 0,
  "name": "string"
 },
 "id": 1,
 "name": "cat",
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
}
```

### Payload with SQL injection attack
```
curl http://localhost:9096/pets --upload-file attack-payload.json
```

You should see the following response:
```json
{
 "attackValues": {
  "content": {
   "category": {
    "name": 0
   },
   "name": 99.97982025146484,
   "photoUrls": [
    0
   ],
   "status": 0,
   "tags": [
    {
     "name": 0
    }
   ]
  }
 },
 "error": "hack attack!"
}
```
