# Handler Routing Gateway

## Testing

Run:
```
go run main.go
```

Then open another terminal and run:
```bash
curl http://localhost:9096/pets/1
```

You should then see something like:
```json
{"category":{"id":0,"name":"string"},"id":1,"name":"aspen","photoUrls":["string"],"status":"done","tags":[{"id":0,"name":"string"}]}
```

Now run:
```bash
curl http://localhost:9096/pets/8
```

You should see:
```json
{"error":"id must be less than 8"}
```

Now run:
```bash
curl -H "Auth: 1337" http://localhost:9096/pets/8
```

You should now see something like:
```bash
{"category":{"id":0,"name":"string"},"id":8,"name":"aspen","photoUrls":["string"],"status":"done","tags":[{"id":0,"name":"string"}]}
```
