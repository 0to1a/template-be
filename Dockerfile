# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS proto-builder

RUN apk add --no-cache curl git

# Install buf
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.50.0/buf-Linux-$(uname -m)" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

# Install protoc plugins and sqlc (pinned versions, combined for fewer layers)
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.0 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.5 && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0

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
FROM golang:1.25-alpine AS builder

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
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /server .

EXPOSE 8080 9090

CMD ["./server"]
