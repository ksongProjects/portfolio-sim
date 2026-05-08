# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./services/market-data-service/
RUN go mod download

COPY internal/ ./internal/
COPY services/market-data-service/ ./services/market-data-service/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/market-data-service ./services/market-data-service/

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/market-data-service /app/market-data-service

EXPOSE 8080

CMD ["/app/market-data-service"]
