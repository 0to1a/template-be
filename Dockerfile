# syntax=docker/dockerfile:1
FROM lavorus/golang-builder:1.25 AS proto-builder

WORKDIR /app

# Copy proto files and buf config
COPY buf.yaml buf.gen.yaml ./
COPY handler/proto/ handler/proto/

# Generate proto code
RUN buf dep update && buf generate

# Copy sqlc config and generate
COPY sqlc.yaml ./
COPY database/ database/

# Generate sqlc code
RUN sqlc generate

# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy generated code from proto-builder
COPY --from=proto-builder /app/compiled/ compiled/

# Copy source code
COPY cmd/ cmd/
COPY handler/ handler/
COPY database/ database/
COPY service/ service/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Final stage
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /server .

EXPOSE 8080 9090

CMD ["./server"]
