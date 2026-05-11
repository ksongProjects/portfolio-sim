# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /workspace

COPY shared /workspace/shared
COPY market-data-service /workspace/market-data-service

WORKDIR /workspace/market-data-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/market-data-service .

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/market-data-service /app/market-data-service

EXPOSE 8080

CMD ["/app/market-data-service"]
