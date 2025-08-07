BINARY := redactyl

.PHONY: build test lint fmt

build:
	go build -o bin/$(BINARY) ./cmd/redactyl

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...