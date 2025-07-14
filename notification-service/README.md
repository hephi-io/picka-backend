# Notification Service

This service listens for NATS events (e.g., `shipment.status.updated`) and logs or sends notifications to users. It is designed for extensibility to support email, SMS, push, etc.

## Environment Variables
- `NATS_URL`: NATS server URL (default: `nats://localhost:4222`)

## Running Locally
```
go run main.go
```

## Running with Docker
```
docker build -t notification-service .
docker run --env NATS_URL=nats://nats:4222 notification-service
```

## Event Contract
The service expects events on the `shipment.status.updated` subject with the following JSON structure:

```
{
  "shipment_id": "string",
  "status": "string",
  "rider_id": "string",
  "timestamp": "string"
}
``` 