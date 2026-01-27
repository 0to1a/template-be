.PHONY: proto sqlc generate build run clean deps

proto:
	buf dep update
	buf generate

sqlc:
	sqlc generate

generate: proto sqlc

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf bin/

deps:
	go mod tidy
	go mod download
