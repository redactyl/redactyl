BINARY := redactyl

.PHONY: build run install test lint fmt clean bench

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

bench:
	go test -run ^$$ -bench=. -benchmem ./internal/artifacts
	go test -run ^$$ -bench=. -benchmem ./internal/engine

clean:
	rm -f bin/$(BINARY)
