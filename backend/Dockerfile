# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app/services/main-api

COPY services/main-api/go.mod services/main-api/go.sum ./
RUN go mod download

COPY services/main-api/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main-api .

# Run stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/services/main-api/main-api .

EXPOSE 8080

CMD ["./main-api"]
