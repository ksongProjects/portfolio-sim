# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

COPY shared /workspace/shared
COPY backend /workspace/backend

WORKDIR /workspace/backend

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/main-api .

# Run stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /bin/main-api .

EXPOSE 8080

CMD ["./main-api"]
