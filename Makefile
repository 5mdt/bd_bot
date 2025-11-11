# Makefile for bd_bot project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=gofmt

# Directories
CMD_DIR=./cmd/app
INTERNAL_DIR=./internal
BUILD_DIR=./build

# Binary output
BINARY_NAME=bd_bot
BINARY_LINUX=$(BINARY_NAME)_linux
BINARY_DARWIN=$(BINARY_NAME)_darwin
BINARY_WIN=$(BINARY_NAME)_windows.exe

# Packages
PKG=5mdt/bd_bot/...

# Ensure build directory exists
$(shell mkdir -p $(BUILD_DIR))

.PHONY: all test vet fmt clean build build-linux build-darwin build-windows build-all deps pre-commit docker-build docker-run ci

# Default target
all: fmt vet test build

# Run tests
test:
	$(GOTEST) $(PKG)

# Run code quality checks
vet:
	$(GOVET) $(PKG)

# Format code
fmt:
	$(GOFMT) -w .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Install dependencies
deps:
	$(GOCMD) mod tidy
	$(GOCMD) mod download

# Run pre-commit hooks
pre-commit:
	pre-commit run --all-files

# Build for current platform
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_LINUX) $(CMD_DIR)

# Build for macOS
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_DARWIN) $(CMD_DIR)

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_WIN) $(CMD_DIR)

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Docker-related tasks
docker-build:
	docker build -t bd_bot .

docker-run:
	docker run -it --rm bd_bot

# CI tasks
ci: deps fmt vet test build
