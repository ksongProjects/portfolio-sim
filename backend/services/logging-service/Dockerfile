# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY services/logging-service/go.mod services/logging-service/go.sum ./services/logging-service/
WORKDIR /app/services/logging-service
RUN go mod download

COPY internal/ ./internal/
COPY services/logging-service/ ./services/logging-service/

RUN CGO_ENABLED=0 GOOS=linux go build -o /logging-service .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/services/logging-service/logging-service .

EXPOSE 8080

CMD ["./logging-service"]
