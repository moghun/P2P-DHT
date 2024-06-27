.PHONY: all build clean test

all: build

build:
    go build -o bin/DHT-14 ./cmd

clean:
    rm -rf bin/

test:
    go test ./tests/...

lint:
    golint ./...

vet:
    go vet ./...