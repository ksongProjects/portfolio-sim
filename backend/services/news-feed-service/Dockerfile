# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/news-feed-service .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/news-feed-service .
COPY --from=builder /app/config/ ./config/

EXPOSE 8080

CMD ["./news-feed-service"]
