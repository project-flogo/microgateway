# Gateway with smart Rate Limiter
This recipe is a gateway which applies rate limit and traffic spike blocking on specified dispatches.

## Installation
* Install [Go](https://golang.org/)

## Setup
```
git clone https://github.com/project-flogo/microgateway
cd microgateway/activity/ratelimiter/examples/api/smart
```

## Testing

Start the gateway:
```
go run main.go
```

### Run the client

Run the following command:
```
go run main.go -client
```

You should see the following like output:
```
0 {"category":{"id":0,"name":"string"},"id":1,"name":"doggie","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}

1 {"category":{"id":0,"name":"string"},"id":1,"name":"doggie","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}

2 {"category":{"id":0,"name":"string"},"id":1,"name":"doggie","photoUrls":["string"],"status":"available","tags":[{"id":0,"name":"string"}]}
```

After 256 requests there will be a spike and traffic, and then requests will be blocked:
```
256 {"status":"Rate Limit Exceeded - The service you have requested is over the allowed limit."}

257 {"status":"Rate Limit Exceeded - The service you have requested is over the allowed limit."}

258 {"status":"Rate Limit Exceeded - The service you have requested is over the allowed limit."}
```

After some time the requests will no longer be blocked.
