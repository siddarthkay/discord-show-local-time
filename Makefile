BINARY_NAME=discord-time-presence
MAIN_FILE=main.go
TEST_DIR=test-results

# Code signing configuration
CODESIGN_IDENTITY ?= "Apple Development: Sid Kay (AC345NN49G)"
CODESIGN_ENTITLEMENTS ?= codesign.entitlements

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .
	@echo "Cross-platform builds complete!"

.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run .

.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

.PHONY: format
format:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: lint
lint:
	@echo "Linting code..."
	go vet ./...

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf $(TEST_DIR)
	@echo "Clean complete!"

.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install


.PHONY: release
release:
	@echo "Building release version..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) .
	@echo "Release build complete: ./$(BINARY_NAME)"

.PHONY: release-all
release-all:
	@echo "Building release versions for all platforms..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 .
	@echo "Release builds complete!"
	@echo "Created binaries:"
	@ls -lh $(BINARY_NAME)-*

.PHONY: release-all-signed
release-all-signed: release-all
	@echo "Code signing macOS binaries..."
ifeq ($(shell uname),Darwin)
	@if [ -f $(BINARY_NAME)-darwin-amd64 ]; then \
		echo "Signing $(BINARY_NAME)-darwin-amd64..."; \
		codesign --sign $(CODESIGN_IDENTITY) --force --verbose $(BINARY_NAME)-darwin-amd64; \
	fi
	@if [ -f $(BINARY_NAME)-darwin-arm64 ]; then \
		echo "Signing $(BINARY_NAME)-darwin-arm64..."; \
		codesign --sign $(CODESIGN_IDENTITY) --force --verbose $(BINARY_NAME)-darwin-arm64; \
	fi
	@echo "Code signing complete!"
	@echo "Verifying signatures..."
	@if [ -f $(BINARY_NAME)-darwin-amd64 ]; then codesign --verify --verbose $(BINARY_NAME)-darwin-amd64; fi
	@if [ -f $(BINARY_NAME)-darwin-arm64 ]; then codesign --verify --verbose $(BINARY_NAME)-darwin-arm64; fi
else
	@echo "Code signing skipped (not running on macOS)"
endif

.PHONY: build-signed
build-signed: build
	@echo "Code signing $(BINARY_NAME)..."
ifeq ($(shell uname),Darwin)
	codesign --sign $(CODESIGN_IDENTITY) --force --verbose $(BINARY_NAME)
	@echo "Code signing complete!"
	@echo "Verifying signature..."
	codesign --verify --verbose $(BINARY_NAME)
	codesign --display --verbose=2 $(BINARY_NAME)
else
	@echo "Code signing skipped (not running on macOS)"
endif

.PHONY: codesign-info
codesign-info:
	@echo "Available code signing identities:"
	@security find-identity -v -p codesigning || echo "No code signing identities found"
	@echo ""
	@echo 'Current signing identity: $(CODESIGN_IDENTITY)'

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build              - Build the application"
	@echo "  build-signed       - Build and code sign (macOS only)"
	@echo "  build-all          - Build for multiple platforms"
	@echo "  run                - Run the application"
	@echo "  deps               - Install dependencies"
	@echo "  format             - Format code"
	@echo "  lint               - Lint code"
	@echo "  clean              - Clean build artifacts"
	@echo "  install            - Install binary to GOPATH/bin"
	@echo "  release            - Build optimized release version"
	@echo "  release-all        - Build optimized releases for all platforms"
	@echo "  release-all-signed - Build and code sign releases (macOS only)"
	@echo "  codesign-info      - Show available code signing identities"
	@echo "  help               - Show this help message"