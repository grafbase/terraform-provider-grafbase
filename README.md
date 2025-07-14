# Grafbase Terraform Provider

This Terraform provider allows you to manage Grafbase resources using the Grafbase API.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- Valid Grafbase API key

## Quick Start

### 1. Configure the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    grafbase = {
      source = "grafbase/grafbase"
    }
  }
}

provider "grafbase" {
  # API key can be set via GRAFBASE_API_KEY environment variable
  # or explicitly set here (not recommended for production)
}
```

### 2. Set Your API Key

```bash
export GRAFBASE_API_KEY="your-grafbase-api-key"
```

### 3. Create a Graph

```hcl
resource "grafbase_graph" "example" {
  account_slug = "my-account"
  slug         = "my-graph"
}

output "graph_id" {
  value = grafbase_graph.example.id
}
```

### 4. Apply the Configuration

```bash
terraform init
terraform plan
terraform apply
```

## Building The Provider

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/grafbase/terraform-provider-grafbase
   cd terraform-provider-grafbase
   ```

2. Build the provider:
   ```bash
   go build -o terraform-provider-grafbase
   ```

3. (Optional) Install locally for development:
   ```bash
   make install
   ```

### Using the Development Script

For a complete development setup:

```bash
./scripts/dev-setup.sh
```

This script will:
- Check dependencies
- Build the provider
- Run tests
- Install the provider locally
- Create development configuration

## Authentication

The provider supports multiple authentication methods:

### Environment Variable (Recommended)

```bash
export GRAFBASE_API_KEY="your-api-key-here"
```

### Provider Configuration

```hcl
provider "grafbase" {
  api_key = var.grafbase_api_key
}

variable "grafbase_api_key" {
  description = "Grafbase API key"
  type        = string
  sensitive   = true
}
```

### Getting Your API Key

1. Visit the [Grafbase Dashboard](https://app.grafbase.com/)
2. Navigate to your organization's settings page
3. Generate a new access token
4. Store it securely (e.g., in your environment or secret management system)

## Resources

### `grafbase_graph`

The `grafbase_graph` resource allows you to manage graphs. Graphs are the fundamental units in Grafbase that contain your GraphQL schema and configuration.

#### Example Usage

**Basic Usage:**
```hcl
resource "grafbase_graph" "example" {
  account_slug = "my-account"
  slug         = "my-graph"
}
```

**With Variables:**
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

resource "grafbase_graph" "app" {
  account_slug = "my-account"
  slug         = "my-app-${var.environment}"
}
```

#### Argument Reference

The following arguments are supported:

- `account_slug` (Required, String) - The slug of the Grafbase account where the graph will be created. This must be an existing account that you have access to. Changing this attribute forces replacement of the resource.

- `slug` (Required, String) - The slug for the graph. Must be unique within the specified account and follow Grafbase naming conventions (lowercase letters, numbers, and hyphens). Changing this attribute forces replacement of the resource.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the graph assigned by Grafbase.
- `created_at` (String) - The RFC3339 timestamp when the graph was created.

#### Import

Existing graphs can be imported using the format `account_slug/graph_slug`:

```bash
# Import a specific graph
terraform import grafbase_graph.example my-account/my-graph

# Import with resource name matching the graph slug
terraform import grafbase_graph.my_graph my-account/my-graph
```

#### Notes

- **Immutability**: Both `account_slug` and `slug` are immutable after creation. Changing either will destroy and recreate the graph.
- **Uniqueness**: Graph slugs must be unique within an account.
- **Naming**: Follow Grafbase naming conventions for slugs (lowercase, alphanumeric, hyphens allowed).
- **Permissions**: You must have appropriate permissions in the specified account to create graphs.

### `grafbase_branch`

The `grafbase_branch` resource allows you to manage branches within a graph. Branches enable you to have different environments and configurations for your GraphQL API.

#### Example Usage

**Basic Usage:**
```hcl
resource "grafbase_graph" "example" {
  account_slug = "my-account"
  slug         = "my-graph"
}

resource "grafbase_branch" "main" {
  account_slug = grafbase_graph.example.account_slug
  graph_slug   = grafbase_graph.example.slug
  name         = "main"
}
```

**Multiple Branches:**
```hcl
resource "grafbase_graph" "app" {
  account_slug = "my-account"
  slug         = "my-app"
}

resource "grafbase_branch" "main" {
  account_slug = grafbase_graph.app.account_slug
  graph_slug   = grafbase_graph.app.slug
  name         = "main"
}

resource "grafbase_branch" "staging" {
  account_slug = grafbase_graph.app.account_slug
  graph_slug   = grafbase_graph.app.slug
  name         = "staging"
}

resource "grafbase_branch" "feature" {
  account_slug = grafbase_graph.app.account_slug
  graph_slug   = grafbase_graph.app.slug
  name         = "feature-new-schema"
}
```

#### Argument Reference

The following arguments are supported:

- `account_slug` (Required, String) - The slug of the Grafbase account where the branch's graph exists. Changing this attribute forces replacement of the resource.

- `graph_slug` (Required, String) - The slug of the graph where this branch will be created. Changing this attribute forces replacement of the resource.

- `name` (Required, String) - The name of the branch. Must be unique within the graph and follow Grafbase naming conventions. Changing this attribute forces replacement of the resource.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the branch assigned by Grafbase.
- `environment` (String) - The environment type of the branch (either `PREVIEW` or `PRODUCTION`).
- `operation_checks_enabled` (Boolean) - Whether operation checks are enabled for this branch.
- `operation_checks_ignore_usage_data` (Boolean) - Whether usage data should be ignored when running operation checks.

#### Import

Existing branches can be imported using the format `account_slug/graph_slug/branch_name`:

```bash
# Import a specific branch
terraform import grafbase_branch.main my-account/my-graph/main

# Import a feature branch
terraform import grafbase_branch.feature my-account/my-graph/feature-auth
```

#### Notes

- **Immutability**: All input attributes (`account_slug`, `graph_slug`, and `name`) are immutable after creation. Changing any of them will destroy and recreate the branch.
- **Production Branch**: The production branch (typically named "main") cannot be deleted. Attempting to delete it will result in an error.
- **Branch Names**: Branch names must be unique within a graph and follow Grafbase naming conventions.
- **Dependencies**: The graph must exist before creating branches. Use Terraform dependencies to ensure proper ordering.

## Examples

Explore the `examples/` directory for complete usage examples:

- [`examples/main.tf`](examples/main.tf) - Basic usage with variables
- [`examples/complete/`](examples/complete/) - Advanced multi-graph setup
- [`examples/dev-setup/`](examples/dev-setup/) - Local development configuration

## Development

### Prerequisites

- Go 1.21 or later
- Terraform 1.5 or later
- Make (optional, for using Makefile commands)

### Quick Development Setup

Use the provided development script for a complete setup:

```bash
./scripts/dev-setup.sh
```

This will handle all the setup steps automatically.

### Manual Development Setup

1. **Clone and build:**
   ```bash
   git clone https://github.com/grafbase/terraform-provider-grafbase
   cd terraform-provider-grafbase
   go mod download
   make build
   ```

2. **Run tests:**
   ```bash
   make test
   ```

3. **Install locally:**
   ```bash
   make install
   ```

### Available Make Commands

```bash
make build      # Build the provider binary
make test       # Run unit tests
make testacc    # Run acceptance tests (requires TF_ACC=1 and valid API key)
make install    # Install provider locally for development
make clean      # Clean build artifacts
make fmt        # Format Go code
make lint       # Run linter (requires golangci-lint)
make docs       # Generate documentation
make dev-setup  # Setup development environment
make debug      # Run provider in debug mode
```

### Documentation

The provider documentation is automatically generated from the code using `tfplugindocs`:

```bash
make docs       # Generate documentation from code
```

The documentation is generated from:
- Provider and resource schemas defined in the Go code
- Template files in the `templates/` directory
- Example configurations in the `examples/` directory

Documentation files are created in the `docs/` directory and are automatically included in releases for publication to the Terraform Registry.

To validate documentation:
```bash
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate
```

### Testing

#### Unit Tests
```bash
go test ./...
```

#### Acceptance Tests
Acceptance tests require a valid Grafbase API key and will create real resources:

```bash
export GRAFBASE_API_KEY="your-api-key"
export TF_VAR_account_slug="your-test-account"
TF_ACC=1 go test ./... -v
```

### Local Development with Terraform

1. **Build the provider:**
   ```bash
   go build -o terraform-provider-grafbase
   ```

2. **Create a `.terraformrc` file:**
   ```hcl
   provider_installation {
     dev_overrides {
       "grafbase/grafbase" = "/path/to/terraform-provider-grafbase"
     }
     direct {}
   }
   ```

3. **Use in your Terraform configuration:**
   ```hcl
   terraform {
     required_providers {
       grafbase = {
         source = "grafbase/grafbase"
       }
     }
   }
   ```

### Debugging

For debugging the provider:

```bash
go run . -debug
```

This outputs instructions for setting `TF_REATTACH_PROVIDERS` environment variable.

For verbose Terraform logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify your API key is correct
   - Check that `GRAFBASE_API_KEY` environment variable is set
   - Ensure the API key has the necessary permissions

2. **Account Not Found**
   - Verify the account slug is correct
   - Check that you have access to the specified account

3. **Graph Already Exists**
   - Graph slugs must be unique within an account
   - Use `terraform import` to import existing graphs

4. **Provider Not Found**
   - Ensure you've run `terraform init`
   - Check your `.terraformrc` configuration for local development

### Debug Steps

1. Enable debug logging:
   ```bash
   export TF_LOG=DEBUG
   ```

2. Check provider installation:
   ```bash
   terraform version
   ```

3. Validate configuration:
   ```bash
   terraform validate
   ```

4. Test API connectivity:
   ```bash
   curl -H "Authorization: Bearer $GRAFBASE_API_KEY" \
        -H "Content-Type: application/json" \
        -d '{"query": "query { __schema { queryType { name } } }"}' \
        https://api.grafbase.com/graphql
   ```

## Contributing

We welcome contributions! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch:**
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. **Make your changes**
4. **Add tests** for new functionality
5. **Run the test suite:**
   ```bash
   make test
   ```
6. **Format your code:**
   ```bash
   make fmt
   ```
7. **Submit a pull request**

### Contribution Guidelines

- Follow Go best practices and conventions
- Add tests for new features and bug fixes
- Update documentation for user-facing changes
- Use descriptive commit messages
- Ensure all tests pass before submitting

### Development Workflow

1. Make changes to the code
2. Run `make build` to build the provider
3. Run `make test` to run unit tests
4. Test manually with example configurations
5. Run `make fmt` to format code
6. Submit pull request

## License

This project is licensed under the Mozilla Public License 2.0. See the [LICENSE](LICENSE) file for details.

## Support

### Community Support

- **GitHub Issues**: [Report bugs and request features](https://github.com/grafbase/terraform-provider-grafbase/issues)

### Grafbase Support

- **Documentation**: [Grafbase Docs](https://grafbase.com/docs)
- **Community**: [Grafbase Discord](https://grafbase.com/discord)

### Reporting Issues

When reporting issues, please include:

1. Terraform version (`terraform version`)
2. Provider version
3. Operating system and architecture
4. Relevant Terraform configuration (sanitized)
5. Error messages or unexpected behavior
6. Steps to reproduce

### Security

For security vulnerabilities, please email [security@grafbase.com](mailto:security@grafbase.com) instead of creating a public issue.
