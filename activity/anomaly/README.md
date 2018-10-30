# Anomaly Detector

The `anomaly` service type implements anomaly detection for payloads. The anomaly detection algorithm is based on a [statistical model](https://fgiesen.wordpress.com/2015/05/26/models-for-adaptive-arithmetic-coding/) for compression. The anomaly detection algorithm computes the relative [complexity](https://en.wikipedia.org/wiki/Kolmogorov_complexity), K(payload | previous payloads), of a payload and then updates the statistical model. A running mean and standard deviation of the complexity is then computed using [this](https://dev.to/nestedsoftware/calculating-standard-deviation-on-streaming-data-253l) algorithm. If the complexity of a payload is some number of deviations from the mean then it is an anomaly. An anomaly is a payload that is statistically significant relative to previous payloads. The anomaly detection algorithm uses real time learning, so what is considered an anomaly can change over time.

The available service `settings` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| depth | number |  The size of the statistical model. Defaults to 2 |

The available `inputs` for the request are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| payload | JSON object | A payload to do anomaly detection on |

The available response `outputs` are as follows:

| Name   |  Type   | Description   |
|:-----------|:--------|:--------------|
| complexity | number | How unusual the payload is in terms of standard deviations from the mean |
| count | number | The number of payloads that have been processed |

A sample `service` definition is:

```json
{
  "name": "Anomaly",
  "description": "Look for anomalies",
  "ref": "github.com/project-flogo/microgateway/activity/anomaly",
  "settings": {
    "depth": 3
  }
}
```

An example `step` that invokes the above `Anomaly` service using `payload` is:

```json
{
  "service": "Anomaly",
  "input": {
    "payload": "=$.payload.content"
  }
}
```

Utilizing the response values can be seen in a response handler:

```json
{
  "if": "($.Anomaly.outputs.count < 100) || ($Anomaly.outputs.complexity < 3)",
  "error": false,
  "output": {
    "code": 200,
    "data": "=$.Update.outputs.result"
  }
}
```
