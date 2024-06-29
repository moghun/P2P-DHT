# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=dht-node
CONFIG_FILE=config.ini
TEST_REPORT_DIR=test_results
TEST_REPORT=$(TEST_REPORT_DIR)/test_result_report.xml
TEST_OUTPUT=$(TEST_REPORT_DIR)/test_output.txt

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/main.go

# Run the project
run: build
	./$(BINARY_NAME) -c $(CONFIG_FILE)

# Test the project and generate a report
test: generate-ca
	mkdir -p $(TEST_REPORT_DIR)
	$(GOTEST) -v ./tests/ | tee $(TEST_OUTPUT) | go-junit-report > $(TEST_REPORT)
	@echo "Test report generated at $(TEST_REPORT)"
	$(MAKE) clean-certs

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
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

# Generate CA certificates
generate-ca:
	python3 project_utils/generate_ca.py

# Clean certificates
clean-certs:
	rm -rf certificates/
	rm -rf tests/certificates/

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

.PHONY: build run test clean deps fmt lint generate-ca clean-certs doc test_report
