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

### 3. Create Resources

**Basic Graph:**
```hcl
resource "grafbase_graph" "example" {
  account_slug = "my-account"
  slug         = "my-graph"
}

output "graph_id" {
  value = grafbase_graph.example.id
}
```

**Complete Federated Setup:**
```hcl
# Graph
resource "grafbase_graph" "api" {
  account_slug = "my-account"
  slug         = "my-api"
}

# Branch
resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.api.id
  name     = "main"
}

# Subgraphs
resource "grafbase_subgraph" "users" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://users.api.example.com/graphql"
}

# Outputs
output "api_endpoints" {
  value = {
    graph_id   = grafbase_graph.api.id
    branch_id  = grafbase_branch.main.id
    subgraphs = {
      users = grafbase_subgraph.users.url
    }
  }
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

The Grafbase provider supports three main resource types that work together to create federated GraphQL APIs:

1. **`grafbase_graph`** - Manages graphs (top-level containers)
2. **`grafbase_branch`** - Manages branches within graphs (like Git branches)
3. **`grafbase_subgraph`** - Manages subgraphs within branches (individual GraphQL services)

### Resource Hierarchy

```
Graph
├── Branch (main)
│   ├── Subgraph (users)
│   ├── Subgraph (products)
│   └── Subgraph (orders)
└── Branch (develop)
    ├── Subgraph (users)
    ├── Subgraph (products)
    └── Subgraph (orders)
```

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

The `grafbase_branch` resource allows you to manage branches within graphs. Branches provide isolated environments for your GraphQL federation, similar to Git branches.

#### Example Usage

**Basic Usage:**
```hcl
resource "grafbase_graph" "example" {
  account_slug = "my-account"
  slug         = "my-graph"
}

resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.example.id
  name     = "main"
}

resource "grafbase_branch" "develop" {
  graph_id = grafbase_graph.example.id
  name     = "develop"
}
```

#### Argument Reference

The following arguments are supported:

- `graph_id` (Required, String) - The ID of the graph where the branch will be created. Changing this attribute forces replacement of the resource.

- `name` (Required, String) - The name of the branch. Must be unique within the graph and follow naming conventions. Common names include "main", "develop", "staging", etc. Changing this attribute forces replacement of the resource.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the branch assigned by Grafbase.
- `created_at` (String) - The RFC3339 timestamp when the branch was created.

#### Import

Existing branches can be imported using the format `graph_id/branch_name`:

```bash
terraform import grafbase_branch.main abcd1234-5678-90ef-ghij-klmnopqrstuv/main
```

#### Notes

- **Immutability**: Both `graph_id` and `name` are immutable after creation. Changing either will destroy and recreate the branch.
- **Uniqueness**: Branch names must be unique within a graph.
- **Dependencies**: A graph must exist before creating branches.

### `grafbase_subgraph`

The `grafbase_subgraph` resource allows you to manage subgraphs within branches. Subgraphs are individual GraphQL services that compose into your federated graph.

#### Example Usage

**Basic Usage:**
```hcl
resource "grafbase_graph" "example" {
  account_slug = "my-account"
  slug         = "my-graph"
}

resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.example.id
  name     = "main"
}

resource "grafbase_subgraph" "users" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://users.api.example.com/graphql"
}

resource "grafbase_subgraph" "products" {
  branch_id = grafbase_branch.main.id
  name      = "products"
  url       = "https://products.api.example.com/graphql"
}
```

**Multiple Environments:**
```hcl
# Main branch subgraphs
resource "grafbase_subgraph" "users_main" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://users.api.example.com/graphql"
}

# Development branch subgraphs
resource "grafbase_subgraph" "users_dev" {
  branch_id = grafbase_branch.develop.id
  name      = "users"
  url       = "https://users-dev.api.example.com/graphql"
}
```

#### Argument Reference

The following arguments are supported:

- `branch_id` (Required, String) - The ID of the branch where the subgraph will be created. Changing this attribute forces replacement of the resource.

- `name` (Required, String) - The name of the subgraph. Must be unique within the branch and follow naming conventions. Changing this attribute forces replacement of the resource.

- `url` (Required, String) - The URL endpoint where the subgraph's GraphQL schema is served. This URL must be accessible to Grafbase and serve a valid GraphQL schema. This attribute can be updated without replacing the resource.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the subgraph assigned by Grafbase.
- `created_at` (String) - The RFC3339 timestamp when the subgraph was created.

#### Import

Existing subgraphs can be imported using the format `branch_id/subgraph_name`:

```bash
terraform import grafbase_subgraph.users abcd1234-5678-90ef-ghij-klmnopqrstuv/users
```

#### Notes

- **URL Updates**: The `url` attribute can be updated to point to a different endpoint without recreating the subgraph.
- **Immutability**: Both `branch_id` and `name` are immutable after creation. Changing either will destroy and recreate the subgraph.
- **Uniqueness**: Subgraph names must be unique within a branch, but can be reused across different branches.
- **Dependencies**: A branch must exist before creating subgraphs.
- **Schema Validation**: The URL must serve a valid GraphQL schema that Grafbase can introspect.

## Examples

Explore the `examples/` directory for complete usage examples:

- [`examples/main.tf`](examples/main.tf) - Basic graph creation with variables
- [`examples/federated-graph/`](examples/federated-graph/) - Complete federated graph with branches and subgraphs
- [`examples/complete/`](examples/complete/) - Advanced multi-graph setup with federation
- [`examples/dev-setup/`](examples/dev-setup/) - Local development configuration

### Quick Example: Federated Graph

Here's a complete example showing how to create a federated GraphQL API:

```hcl
# Create a graph
resource "grafbase_graph" "api" {
  account_slug = "my-account"
  slug         = "my-api"
}

# Create main branch
resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.api.id
  name     = "main"
}

# Add subgraphs to compose the federated API
resource "grafbase_subgraph" "users" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://users.api.example.com/graphql"
}

resource "grafbase_subgraph" "products" {
  branch_id = grafbase_branch.main.id
  name      = "products"
  url       = "https://products.api.example.com/graphql"
}

resource "grafbase_subgraph" "orders" {
  branch_id = grafbase_branch.main.id
  name      = "orders"
  url       = "https://orders.api.example.com/graphql"
}
```

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
