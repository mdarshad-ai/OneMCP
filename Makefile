.PHONY: build clean test run deps

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package
MAIN_PACKAGE=cmd/onemcp/main.go
BINARY_NAME=onemcp
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(MAIN_PACKAGE)

# Build for Windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe -v $(MAIN_PACKAGE)

# Build for macOS
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_darwin -v $(MAIN_PACKAGE)

# Build for all platforms
build-all: build-linux build-windows build-darwin

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_NAME)_darwin

# Run tests
test:
	$(GOTEST) -v ./...

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)
	./$(BINARY_NAME)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install development dependencies
dev-deps: deps
	$(GOGET) -u github.com/spf13/cobra@latest

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code
lint:
	$(GOCMD) vet ./...

# Development setup
setup: dev-deps fmt lint

# Install to system (requires sudo)
install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Uninstall from system (requires sudo)
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)