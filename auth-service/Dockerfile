# Build stage
FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o auth-service .

# Runtime stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/auth-service .
RUN apk --no-cache add ca-certificates
CMD ["./auth-service"] 