# syntax=docker/dockerfile:1
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o notification-service main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/notification-service .
EXPOSE 8080
CMD ["./notification-service"] 