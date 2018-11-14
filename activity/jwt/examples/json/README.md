# Gateway with a JWT
This recipe is a gateway with a service protected by a JWT.

## Installation
* Install [Go](https://golang.org/)
* Install the flogo [cli](https://github.com/project-flogo/cli)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd microgateway/activity/jwt/examples/json
```
Place the Downloaded Mashling-Gateway binary in circuit-breaker-gateway folder.

## Testing

Generate a JWT token using the below information:
(You may use http://jwtbuilder.jamiekurtz.com/)

```
{
     "issuer": "Mashling",
     "audience": "www.mashling.io",
     "subject": "tempuser@mail.com",
     "id": "XX",
     "signingMethod": "HMAC"
}
{
     "key": "qwertyuiopasdfghjklzxcvbnm789101"
}
```
Note: The id in the above payload is the pet Id.

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
and test below scenario.

### Token is Valid

Now run the following in a new terminal:
```
curl --request GET http://localhost:9096/pets -H "Authorization: Bearer <Access_Token>"
```

You should see the following response:
```json
{
   "error":"JWT token is valid",
   "pet":{
      "category":{
         "id":0,
         "name":"string"
      },
      "id":4,
      "name":"gigigi",
      "photoUrls":[
         "string"
      ],
      "status":"available",
      "tags":[
         {
            "id":0,
            "name":"string"
         }
      ]
   }
}
```


### Token Invalid
You should see the following response:
```json
{
   "error":{
      "error":false,
      "errorMessage":"",
      "token":{
         "claims":null,
         "signature":"",
         "signingMethod":"",
         "header":null
      },
      "valid":false,
      "validationMessage":"signature is invalid"
   },
   "pet":null
}
```
