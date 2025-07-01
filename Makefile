default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Build provider
.PHONY: build
build:
	go build -o terraform-provider-grafbase

# Install provider locally
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/grafbase.com/grafbase/grafbase/1.0.0/darwin_amd64
	cp terraform-provider-grafbase ~/.terraform.d/plugins/grafbase.com/grafbase/grafbase/1.0.0/darwin_amd64

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-grafbase

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Run tests
.PHONY: test
test:
	go test ./... -v

# Generate documentation
.PHONY: docs
docs:
	go generate ./...

# Run all checks
.PHONY: check
check: fmt lint test

# Setup development environment
.PHONY: dev-setup
dev-setup:
	go mod download
	go mod tidy

# Run provider in debug mode
.PHONY: debug
debug:
	go run . -debug

# Release build
.PHONY: release
release:
	goreleaser release --rm-dist

# Development release
.PHONY: dev-release
dev-release:
	goreleaser release --snapshot --rm-dist

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the provider binary"
	@echo "  install    - Install provider locally"
	@echo "  test       - Run unit tests"
	@echo "  testacc    - Run acceptance tests"
	@echo "  fmt        - Format Go code"
	@echo "  lint       - Run linter"
	@echo "  docs       - Generate documentation"
	@echo "  check      - Run fmt, lint, and test"
	@echo "  clean      - Remove build artifacts"
	@echo "  debug      - Run provider in debug mode"
	@echo "  dev-setup  - Setup development environment"
	@echo "  help       - Show this help message"
