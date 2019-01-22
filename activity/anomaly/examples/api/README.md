# Gateway with anomaly detection
This recipe is a gateway with anomaly detection using real time learning.

## Description
The anomaly detection algorithm in this recipe is based on a [statistical model](https://fgiesen.wordpress.com/2015/05/26/models-for-adaptive-arithmetic-coding/) for compression. The anomaly detection algorithm computes the relative [complexity](https://en.wikipedia.org/wiki/Kolmogorov_complexity), K(payload | previous payloads), of a payload and then updates the statistical model. A running mean and standard deviation of the complexity is then computed using [this](https://dev.to/nestedsoftware/calculating-standard-deviation-on-streaming-data-253l) algorithm. If the complexity of a payload is some number of deviations from the mean then it is an anomaly. An anomaly is a payload that is statistically significant relative to previous payloads. The anomaly detection algorithm uses real time learning, so what is considered an anomaly can change over time. In the below scenario the Mashling is fed 1024 payloads to initialize the statistical model. A payload that is engineered to be an anomaly relative to the previous payloads is then fed into the Mashling.

## Installation
* Install [Go](https://golang.org/)

## Setup
```
git clone https://github.com/project-flogo/microgateway/activity/anomaly
cd microgateway/activity/anomaly
```

## Testing
Start the gateway:
```
go run main.go
```
and test below scenario.

### Payload that is an anomaly
Start an echo server for the Mashling to talk to by running the following command:
```
go run support/main.go -server
```

Initialize the statistical model by opening a new terminal and running:
```
go run support/main.go -client
```

You should see the following:
```
number of anomalies 0
average complexity NaN
```
A 1024 payloads have been fed into the anomaly detection service, and zero anomalies have been found. If some anomalies had been found the average complexity would be a valid number.

Now run the following to feed one more payload into the anomaly detection service:
```
curl -H "Content-Type:application/json" http://localhost:9096/test --upload-file anomaly-payload.json
```

You should see the following response:
```json
{
 "complexity": 13.715719,
 "error": "anomaly!"
}
```
The complexity is13.715719 standard deviations from the mean. Because this is greater than the 3 standard deviation threshold the payload is considered an anomaly. In this scenario the standard deviation threshold was increased until only a small number of anomalies were detected.
