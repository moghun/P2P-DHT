# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=dht-node
CONFIG_FILE=config.ini

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/main.go

# Run the project
run: build
	./$(BINARY_NAME) -c $(CONFIG_FILE)

# Test the project
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	$(GOGET) -v ./...

# Format the code
fmt:
	$(GOCMD) fmt ./...

lint:
	golangci-lint run

doc:
	godoc -http=:6060

.PHONY: build run test clean deps fmt lint doc