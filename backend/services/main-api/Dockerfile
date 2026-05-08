# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./services/main-api/
RUN go mod download

COPY internal/ ./internal/
COPY services/main-api/ ./services/main-api/

RUN CGO_ENABLED=0 GOOS=linux go build -o main-api ./services/main-api/

# Run stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/main-api .

EXPOSE 8080

CMD ["./main-api"]
