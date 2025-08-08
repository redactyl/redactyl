BINARY := redactyl

.PHONY: build run install test lint fmt clean

build:
	go build -o bin/$(BINARY) .

run: build
	./bin/$(BINARY) --help

install:
	go install .

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -f bin/$(BINARY)