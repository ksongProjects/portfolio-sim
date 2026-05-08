# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./services/news-feed-service/
RUN go mod download

COPY internal/ ./internal/
COPY services/news-feed-service/ ./services/news-feed-service/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/news-feed-service ./services/news-feed-service/

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/news-feed-service .
COPY --from=builder /app/services/news-feed-service/config/ ./config/

EXPOSE 8080

CMD ["./news-feed-service"]
