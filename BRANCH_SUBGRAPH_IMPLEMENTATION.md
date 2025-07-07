# Branch and Subgraph Resources Implementation

This document summarizes the implementation of Branch and Subgraph resources for the Grafbase Terraform Provider.

## Overview

Added support for managing federated GraphQL APIs through three interconnected resources:

1. **`grafbase_graph`** (existing) - Top-level containers for GraphQL APIs
2. **`grafbase_branch`** (new) - Isolated environments within graphs (like Git branches)
3. **`grafbase_subgraph`** (new) - Individual GraphQL services that compose into federated APIs

## Resource Hierarchy

```
Account (my-account)
└── Graph (my-api)
    ├── Branch (main)
    │   ├── Subgraph (users) → https://users.api.example.com/graphql
    │   ├── Subgraph (products) → https://products.api.example.com/graphql
    │   └── Subgraph (orders) → https://orders.api.example.com/graphql
    └── Branch (develop)
        ├── Subgraph (users) → https://users-dev.api.example.com/graphql
        ├── Subgraph (products) → https://products-dev.api.example.com/graphql
        └── Subgraph (orders) → https://orders-dev.api.example.com/graphql
```

## Implementation Details

### Client Extensions (`internal/client/client.go`)

Added new types and API methods for Branch and Subgraph management:

#### New Types
- `Branch` struct with ID, Name, CreatedAt, and Graph reference
- `Subgraph` struct with ID, Name, URL, CreatedAt, and Branch reference
- `CreateBranchInput` and `CreateSubgraphInput` for creation operations

#### New API Methods
- `CreateBranch()` - Creates a new branch in a graph
- `GetBranch()` - Retrieves branch by graph ID and name
- `DeleteBranch()` - Deletes a branch by ID
- `CreateSubgraph()` - Creates a new subgraph in a branch
- `GetSubgraph()` - Retrieves subgraph by branch ID and name
- `UpdateSubgraph()` - Updates subgraph URL (only field that can be modified)
- `DeleteSubgraph()` - Deletes a subgraph by ID

### Branch Resource (`internal/provider/branch_resource.go`)

#### Schema
- `id` (Computed) - Unique branch identifier
- `graph_id` (Required, ForceReplace) - Parent graph ID
- `name` (Required, ForceReplace) - Branch name (e.g., "main", "develop")
- `created_at` (Computed) - Creation timestamp

#### Features
- Full CRUD operations
- Import support with format `graph_id/branch_name`
- Force replacement when graph_id or name changes
- Proper error handling and state management

### Subgraph Resource (`internal/provider/subgraph_resource.go`)

#### Schema
- `id` (Computed) - Unique subgraph identifier
- `branch_id` (Required, ForceReplace) - Parent branch ID
- `name` (Required, ForceReplace) - Subgraph name (e.g., "users", "products")
- `url` (Required, Updatable) - GraphQL endpoint URL
- `created_at` (Computed) - Creation timestamp

#### Features
- Full CRUD operations including URL updates
- Import support with format `branch_id/subgraph_name`
- Force replacement when branch_id or name changes
- URL can be updated without replacement
- Proper error handling and state management

### Provider Registration (`internal/provider/provider.go`)

Updated the provider to register the new resources:
```go
func (p *GrafbaseProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewGraphResource,    // existing
        NewBranchResource,   // new
        NewSubgraphResource, // new
    }
}
```

### Test Coverage

#### Branch Resource Tests (`internal/provider/branch_resource_test.go`)
- Create and Read operations
- Import state verification
- Update testing (validates replacement behavior)
- Delete operations

#### Subgraph Resource Tests (`internal/provider/subgraph_resource_test.go`)
- Create and Read operations
- Import state verification
- URL update testing (in-place updates)
- Name change testing (validates replacement behavior)
- Delete operations

## Examples and Documentation

### New Examples

#### Basic Federated Graph (`examples/federated-graph/main.tf`)
Complete example showing:
- Graph creation
- Main and develop branches
- Multiple subgraphs per branch with different URLs
- Comprehensive outputs

#### Enhanced Complete Example (`examples/complete/main.tf`)
Extended the existing example to include:
- Branch and subgraph management
- Multi-environment setups
- Realistic federated architecture patterns

### Documentation Updates (`README.md`)

#### Resource Documentation
- Complete documentation for all three resources
- Argument and attribute references
- Import instructions and examples
- Usage patterns and best practices

#### Quick Start Updates
- Added federated setup examples
- Updated hierarchy explanations
- Comprehensive usage examples

## Usage Examples

### Basic Usage
```hcl
# Create graph
resource "grafbase_graph" "api" {
  account_slug = "my-account"
  slug         = "my-api"
}

# Create branch
resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.api.id
  name     = "main"
}

# Create subgraph
resource "grafbase_subgraph" "users" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://users.api.example.com/graphql"
}
```

### Multi-Environment Setup
```hcl
# Main environment
resource "grafbase_subgraph" "users_main" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://users.api.example.com/graphql"
}

# Development environment
resource "grafbase_subgraph" "users_dev" {
  branch_id = grafbase_branch.develop.id
  name      = "users"
  url       = "https://users-dev.api.example.com/graphql"
}
```

### Import Operations
```bash
# Import branch
terraform import grafbase_branch.main graph-id-123/main

# Import subgraph
terraform import grafbase_subgraph.users branch-id-456/users
```

## Key Features

### Resource Relationships
- **Dependencies**: Graph → Branch → Subgraph
- **References**: Each child resource references its parent via ID
- **Lifecycle**: Parent resources must exist before children can be created

### Update Behavior
- **Graph**: account_slug and slug changes force replacement
- **Branch**: graph_id and name changes force replacement
- **Subgraph**: branch_id and name changes force replacement, URL can be updated in-place

### Error Handling
- Comprehensive error messages for API failures
- Proper handling of "not found" scenarios
- State cleanup when resources are deleted externally

### Import Support
- All resources support Terraform import operations
- Consistent import ID formats across resources
- Validation of import IDs with helpful error messages

## API Schema Assumptions

The implementation assumes the following GraphQL API structure based on research and federation patterns:

### Mutations
```graphql
mutation {
  branchCreate(input: BranchCreateInput!) {
    ... on BranchCreateSuccess { branch { id name createdAt graph {...} } }
    ... on BranchAlreadyExistsError { __typename }
  }
  
  subgraphCreate(input: SubgraphCreateInput!) {
    ... on SubgraphCreateSuccess { subgraph { id name url createdAt branch {...} } }
    ... on SubgraphAlreadyExistsError { __typename }
  }
}
```

### Queries
```graphql
query {
  branchByGraphId(graphId: ID!, branchName: String!) { id name createdAt graph {...} }
  subgraphByBranchId(branchId: ID!, subgraphName: String!) { id name url createdAt branch {...} }
}
```

## Benefits

### For Users
1. **Complete Federation Management**: Full lifecycle management of federated GraphQL APIs
2. **Environment Isolation**: Branches provide clean environment separation
3. **Flexible Architecture**: Supports complex multi-service architectures
4. **Infrastructure as Code**: All federation components manageable via Terraform

### for Development
1. **Consistent Patterns**: All resources follow the same implementation patterns
2. **Comprehensive Testing**: Full test coverage for all CRUD operations
3. **Import Support**: Easy migration of existing resources
4. **Error Handling**: Robust error handling and state management

### For Operations
1. **Deployment Workflows**: Supports sophisticated deployment patterns
2. **Environment Management**: Easy management of multiple environments
3. **Rollback Capabilities**: URL updates enable easy rollbacks
4. **Monitoring**: Clear resource hierarchy for observability

## Future Enhancements

Potential future improvements could include:

1. **Data Sources**: Read-only data sources for branches and subgraphs
2. **Validation**: Enhanced validation of GraphQL endpoints
3. **Schema Checks**: Integration with Grafbase schema validation
4. **Batch Operations**: Support for bulk subgraph operations
5. **Advanced Configuration**: Additional subgraph configuration options

## Conclusion

The Branch and Subgraph resources complete the Grafbase Terraform Provider's support for federated GraphQL API management. Users can now manage their entire federation infrastructure as code, from top-level graphs down to individual subgraph endpoints, with full support for multiple environments and deployment patterns.