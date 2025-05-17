.PHONY: test cover lint build clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=multigit

all: test build

# Run tests
TEST_PACKAGES=./...
test:
	$(GOTEST) -v -coverprofile=coverage.out $(TEST_PACKAGES)

# Show test coverage in browser
cover: test
	$(GOTEST) -coverprofile=coverage.out $(TEST_PACKAGES) && go tool cover -html=coverage.out

# Run linter
lint:
	# Install golangci-lint if not installed
	if ! command -v golangci-lint &> /dev/null; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.52.2; \
	fi
	golangci-lint run --timeout 5m

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/multigit

# Install dependencies
deps:
	$(GOMOD) tidy

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Install the binary
install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Run all checks (tests, lint, etc.)
check: lint test
