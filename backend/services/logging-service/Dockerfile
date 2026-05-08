# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./services/logging-service/
RUN go mod download

COPY internal/ ./internal/
COPY services/logging-service/ ./services/logging-service/

RUN CGO_ENABLED=0 GOOS=linux go build -o /logging-service ./services/logging-service/

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /logging-service .

EXPOSE 8080

CMD ["./logging-service"]
