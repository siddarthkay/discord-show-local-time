BINARY_NAME=discord-time-presence
MAIN_FILE=main.go
TEST_DIR=test-results

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: ./$(BINARY_NAME)"

.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	@echo "Cross-platform builds complete!"

.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_FILE)

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
	go build -ldflags="-s -w" -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "Release build complete: ./$(BINARY_NAME)"

.PHONY: release-all
release-all:
	@echo "Building release versions for all platforms..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	@echo "Release builds complete!"
	@echo "Created binaries:"
	@ls -lh $(BINARY_NAME)-*

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  run           - Run the application"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  release       - Build optimized release version"
	@echo "  release-all   - Build optimized releases for all platforms"
	@echo "  help          - Show this help message"