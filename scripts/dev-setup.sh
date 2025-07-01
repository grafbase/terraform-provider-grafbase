#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."

    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed. Please install Terraform."
        exit 1
    fi

    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "1.21" "$GO_VERSION" | sort -V | head -n1)" != "1.21" ]]; then
        print_error "Go version 1.21 or later is required. Current version: $GO_VERSION"
        exit 1
    fi

    print_success "All dependencies are installed"
}

# Setup Go module and dependencies
setup_go_module() {
    print_status "Setting up Go module and dependencies..."

    go mod download
    go mod tidy

    print_success "Go module setup complete"
}

# Build the provider
build_provider() {
    print_status "Building the provider..."

    go build -o terraform-provider-grafbase

    if [[ $? -eq 0 ]]; then
        print_success "Provider built successfully"
    else
        print_error "Failed to build provider"
        exit 1
    fi
}

# Install provider locally for development
install_local_provider() {
    print_status "Installing provider locally for development..."

    # Detect OS and architecture
    OS=$(uname | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    # Create local plugin directory
    PLUGIN_DIR="$HOME/.terraform.d/plugins/grafbase.com/grafbase/grafbase/1.0.0/${OS}_${ARCH}"
    mkdir -p "$PLUGIN_DIR"

    # Copy provider binary
    cp terraform-provider-grafbase "$PLUGIN_DIR/"
    chmod +x "$PLUGIN_DIR/terraform-provider-grafbase"

    print_success "Provider installed locally at $PLUGIN_DIR"
}

# Create .terraformrc for local development
create_terraformrc() {
    print_status "Creating .terraformrc for local development..."

    TERRAFORMRC_PATH="$HOME/.terraformrc"

    if [[ -f "$TERRAFORMRC_PATH" ]]; then
        print_warning ".terraformrc already exists. Creating backup..."
        cp "$TERRAFORMRC_PATH" "$TERRAFORMRC_PATH.backup.$(date +%s)"
    fi

    cat > "$TERRAFORMRC_PATH" << EOF
provider_installation {
  dev_overrides {
    "grafbase/grafbase" = "$(pwd)"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal.
  direct {}
}
EOF

    print_success ".terraformrc created for local development"
    print_warning "Remember to remove or modify .terraformrc when done with development"
}

# Run tests
run_tests() {
    print_status "Running tests..."

    go test ./... -v

    if [[ $? -eq 0 ]]; then
        print_success "All tests passed"
    else
        print_error "Some tests failed"
        exit 1
    fi
}

# Verify example configuration
verify_example() {
    print_status "Verifying example configuration..."

    cd examples

    # Initialize Terraform
    terraform init

    # Validate configuration
    terraform validate

    if [[ $? -eq 0 ]]; then
        print_success "Example configuration is valid"
    else
        print_error "Example configuration is invalid"
        exit 1
    fi

    cd ..
}

# Print usage instructions
print_usage() {
    cat << EOF
Grafbase Terraform Provider Development Setup Complete!

Next steps:
1. Set your Grafbase API key:
   export GRAFBASE_API_KEY="your-api-key-here"

2. Navigate to the examples directory:
   cd examples

3. Copy the example variables file:
   cp terraform.tfvars.example terraform.tfvars

4. Edit terraform.tfvars with your actual values

5. Run Terraform commands:
   terraform init
   terraform plan
   terraform apply

For debugging:
- Run the provider in debug mode: go run . -debug
- Use TF_LOG=DEBUG for verbose Terraform logging

For development:
- Make changes to the provider code
- Run: go build -o terraform-provider-grafbase
- Test your changes with Terraform

Remember to clean up .terraformrc when done with development!
EOF
}

# Main execution
main() {
    print_status "Starting Grafbase Terraform Provider development setup..."

    check_dependencies
    setup_go_module
    build_provider
    run_tests
    install_local_provider
    create_terraformrc
    verify_example

    print_success "Development setup complete!"
    echo
    print_usage
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--help]"
        echo "Sets up local development environment for Grafbase Terraform Provider"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
