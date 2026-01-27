# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS proto-builder

RUN apk add --no-cache curl git

# Install buf
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.50.0/buf-Linux-$(uname -m)" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

# Install protoc plugins (with cache)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Install sqlc (with cache)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

WORKDIR /app

# Copy proto files and buf config
COPY buf.yaml buf.gen.yaml ./
COPY handler/proto/ handler/proto/

# Generate proto code (with buf cache)
RUN --mount=type=cache,target=/root/.cache/buf \
    buf dep update && buf generate

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
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy generated code from proto-builder
COPY --from=proto-builder /app/compiled/ compiled/

# Copy source code
COPY cmd/ cmd/
COPY handler/ handler/
COPY database/ database/
COPY service/ service/

# Build binary (with cache)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /server .

EXPOSE 8080 9090

CMD ["./server"]
