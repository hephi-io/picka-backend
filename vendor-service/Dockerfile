# Build stage
FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o vendor-service .

# Runtime stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/vendor-service .
RUN apk --no-cache add ca-certificates
CMD ["./vendor-service"] 