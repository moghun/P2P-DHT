# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BINARY_NODE_NAME=dht_node
BINARY_BOOTSTRAP_NAME=bootstrap_node
CONFIG_FILE=config.ini
TEST_REPORT_DIR=test_results
TEST_REPORT=$(TEST_REPORT_DIR)/test_result_report.xml
TEST_OUTPUT=$(TEST_REPORT_DIR)/test_output.txt

# Build both dht_node and bootstrap_node
build: build_node build_bootstrap_node

build_node:
	$(GOBUILD) -o $(BINARY_NODE_NAME) ./cmd/node/main.go

build_bootstrap_node:
	$(GOBUILD) -o $(BINARY_BOOTSTRAP_NAME) ./cmd/bootstrap_node/main.go

# Run the dht_node project
run_node: build_node
	./$(BINARY_NODE_NAME) -c $(CONFIG_FILE)

# Run the bootstrap_node project
run_bootstrap: build_bootstrap_node
	./$(BINARY_BOOTSTRAP_NAME) -c $(CONFIG_FILE)

# Test the project and generate a report
test:
	mkdir -p $(TEST_REPORT_DIR)
	$(GOTEST) -v ./tests/ | tee $(TEST_OUTPUT) | go-junit-report > $(TEST_REPORT)
	@echo "Test report generated at $(TEST_REPORT)"

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NODE_NAME) $(BINARY_BOOTSTRAP_NAME)
	rm -rf $(TEST_REPORT_DIR)

# Install dependencies
deps:
	$(GOGET) -v ./...

# Format the code
fmt:
	$(GOCMD) fmt ./...

# Lint the code
lint:
	golangci-lint run

# Documentation server
doc:
	godoc -http=:6060

# Generate HTML test report
test_and_report: test
	xsltproc ./project_utils/junit-xml-to-html.xsl $(TEST_REPORT) > $(TEST_REPORT_DIR)/test_result_report.html
	open $(TEST_REPORT_DIR)/test_result_report.html

test_report:
	xsltproc ./project_utils/junit-xml-to-html.xsl $(TEST_REPORT) > $(TEST_REPORT_DIR)/test_result_report.html
	open $(TEST_REPORT_DIR)/test_result_report.html

.PHONY: build build_node build_bootstrap_node run_node run_bootstrap test clean deps fmt lint doc test_report