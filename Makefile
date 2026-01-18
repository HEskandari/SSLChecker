# SSL Certificate Monitor Makefile

.PHONY: all build clean test install uninstall dist setup help

# Variables
BINARY_NAME = ssl-cert-monitor
VERSION = $(shell git describe --tags 2>/dev/null || echo "dev")
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO = go
GOFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
DIST_DIR = dist
BUILD_DIR = build

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/ssl-cert-monitor

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ssl-cert-monitor
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/ssl-cert-monitor

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ssl-cert-monitor
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/ssl-cert-monitor

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/ssl-cert-monitor

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f test-state.json
	rm -f coverage.out

# Install locally
install: build
	@echo "Installing..."
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Uninstall
uninstall:
	@echo "Uninstalling..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Create distribution packages
dist: clean build-all
	@echo "Creating distribution packages..."
	mkdir -p $(DIST_DIR)
	
	# Linux amd64
	mkdir -p $(DIST_DIR)/linux-amd64
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/linux-amd64/$(BINARY_NAME)
	cp config.example.yaml $(DIST_DIR)/linux-amd64/
	cp README.md $(DIST_DIR)/linux-amd64/
	cp -r deployment $(DIST_DIR)/linux-amd64/
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz -C $(DIST_DIR)/linux-amd64 .
	
	# Linux arm64
	mkdir -p $(DIST_DIR)/linux-arm64
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(DIST_DIR)/linux-arm64/$(BINARY_NAME)
	cp config.example.yaml $(DIST_DIR)/linux-arm64/
	cp README.md $(DIST_DIR)/linux-arm64/
	cp -r deployment $(DIST_DIR)/linux-arm64/
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-linux-arm64.tar.gz -C $(DIST_DIR)/linux-arm64 .
	
	# macOS amd64
	mkdir -p $(DIST_DIR)/darwin-amd64
	cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(DIST_DIR)/darwin-amd64/$(BINARY_NAME)
	cp config.example.yaml $(DIST_DIR)/darwin-amd64/
	cp README.md $(DIST_DIR)/darwin-amd64/
	cp -r deployment $(DIST_DIR)/darwin-amd64/
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(DIST_DIR)/darwin-amd64 .
	
	# macOS arm64
	mkdir -p $(DIST_DIR)/darwin-arm64
	cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(DIST_DIR)/darwin-arm64/$(BINARY_NAME)
	cp config.example.yaml $(DIST_DIR)/darwin-arm64/
	cp README.md $(DIST_DIR)/darwin-arm64/
	cp -r deployment $(DIST_DIR)/darwin-arm64/
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(DIST_DIR)/darwin-arm64 .
	
	# Windows
	mkdir -p $(DIST_DIR)/windows-amd64
	cp $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/windows-amd64/$(BINARY_NAME).exe
	cp config.example.yaml $(DIST_DIR)/windows-amd64/
	cp README.md $(DIST_DIR)/windows-amd64/
	cp -r deployment $(DIST_DIR)/windows-amd64/
	zip -r $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.zip $(DIST_DIR)/windows-amd64/
	
	@echo "Distribution packages created in $(DIST_DIR)/"

# Run with test configuration
run: build
	@echo "Running with test configuration..."
	./$(BINARY_NAME) --config test-config.yaml

# Verify certificates only
verify: build
	@echo "Verifying certificates..."
	./$(BINARY_NAME) --config test-config.yaml --verify

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

# Setup development environment
setup:
	@echo "Setting up development environment..."
	$(GO) mod download
	$(GO) mod tidy

# Show help
help:
	@echo "SSL Certificate Monitor Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Build the binary (default)"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for Linux, macOS, and Windows"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install locally"
	@echo "  uninstall    - Uninstall"
	@echo "  dist         - Create distribution packages"
	@echo "  run          - Run with test configuration"
	@echo "  verify       - Verify certificates only"
	@echo "  coverage     - Generate coverage report"
	@echo "  setup        - Setup development environment"
	@echo "  help         - Show this help"