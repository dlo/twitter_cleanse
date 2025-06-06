.PHONY: build clean test install run fmt vet check build-all help

# Build the binary
build:
	go build -o twitter-cleanse .

# Clean build artifacts
clean:
	rm -f twitter-cleanse
	rm -rf twitter_cleanse_cache/

# Run tests
test:
	go test ./...

# Install the binary to GOPATH/bin
install:
	go install .

# Run with dry-run mode (requires environment variables)
run:
	go run . --dry-run

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run all checks
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o twitter-cleanse-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o twitter-cleanse-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o twitter-cleanse-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o twitter-cleanse-windows-amd64.exe .

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install binary to GOPATH/bin"
	@echo "  run        - Run with dry-run mode"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  check      - Run fmt, vet, and test"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  help       - Show this help" 