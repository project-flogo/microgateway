# Async Gateway

## Testing

Run:
```bash
go run main.go
```

Run the following command:
```bash
curl --request GET http://localhost:9096/endpoint
```

You should see:
On the server screen, you get 200 response code and log service outputs "Output: Test log message service invoked"