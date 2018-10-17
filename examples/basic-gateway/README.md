# Basic Gateway

## Install

To install run the following commands:
```
flogo create -f flogo.json
cd MyProxy
flogo install github.com/mashling/flogoactivity
flogo install github.com/TIBCOSoftware/flogo-contrib/activity/rest
flogo build
```

## Testing

Run:
```
MyProxy
```

Then open another terminal and run:
```
curl http://localhost:9096/pets/1
```

You should then see something like:
```
{"category":{"id":0,"name":"string"},"id":1,"name":"aspen","photoUrls":["string"],"status":"done","tags":[{"id":0,"name":"string"}]}
```
