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
  # api_key = var.grafbase_api_key
}

variable "grafbase_api_key" {
  description = "Grafbase API key for authentication"
  type        = string
  sensitive   = true
  default     = null
}

variable "account_slug" {
  description = "Grafbase account slug"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

# Development graph
resource "grafbase_graph" "dev" {
  account_slug = var.account_slug
  slug         = "my-app-${var.environment}"
}

# Create main branch for the development graph
resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.dev.id
  name     = "main"
}

# Create development branch for the development graph
resource "grafbase_branch" "develop" {
  graph_id = grafbase_graph.dev.id
  name     = "develop"
}

# Create subgraphs for the main branch
resource "grafbase_subgraph" "users_main" {
  branch_id = grafbase_branch.main.id
  name      = "users"
  url       = "https://api.example.com/users/graphql"
}

resource "grafbase_subgraph" "products_main" {
  branch_id = grafbase_branch.main.id
  name      = "products"
  url       = "https://api.example.com/products/graphql"
}

resource "grafbase_subgraph" "orders_main" {
  branch_id = grafbase_branch.main.id
  name      = "orders"
  url       = "https://api.example.com/orders/graphql"
}

# Create subgraphs for the develop branch (might use different URLs)
resource "grafbase_subgraph" "users_develop" {
  branch_id = grafbase_branch.develop.id
  name      = "users"
  url       = "https://api-dev.example.com/users/graphql"
}

resource "grafbase_subgraph" "products_develop" {
  branch_id = grafbase_branch.develop.id
  name      = "products"
  url       = "https://api-dev.example.com/products/graphql"
}

resource "grafbase_subgraph" "orders_develop" {
  branch_id = grafbase_branch.develop.id
  name      = "orders"
  url       = "https://api-dev.example.com/orders/graphql"
}

# API graph for microservices
resource "grafbase_graph" "api" {
  account_slug = var.account_slug
  slug         = "api-gateway-${var.environment}"
}

# Create production branch for API graph
resource "grafbase_branch" "api_main" {
  graph_id = grafbase_graph.api.id
  name     = "main"
}

# Create subgraphs for API gateway
resource "grafbase_subgraph" "auth_api" {
  branch_id = grafbase_branch.api_main.id
  name      = "auth"
  url       = "https://auth.example.com/graphql"
}

resource "grafbase_subgraph" "gateway_api" {
  branch_id = grafbase_branch.api_main.id
  name      = "gateway"
  url       = "https://gateway.example.com/graphql"
}

# Analytics graph
resource "grafbase_graph" "analytics" {
  account_slug = var.account_slug
  slug         = "analytics-${var.environment}"
}

# Create main branch for analytics
resource "grafbase_branch" "analytics_main" {
  graph_id = grafbase_graph.analytics.id
  name     = "main"
}

# Analytics subgraph
resource "grafbase_subgraph" "analytics_subgraph" {
  branch_id = grafbase_branch.analytics_main.id
  name      = "analytics"
  url       = "https://analytics.example.com/graphql"
}

# Outputs for other Terraform configurations or CI/CD systems
output "dev_graph" {
  description = "Development graph details"
  value = {
    id         = grafbase_graph.dev.id
    slug       = grafbase_graph.dev.slug
    created_at = grafbase_graph.dev.created_at
    branches = {
      main = {
        id   = grafbase_branch.main.id
        name = grafbase_branch.main.name
      }
      develop = {
        id   = grafbase_branch.develop.id
        name = grafbase_branch.develop.name
      }
    }
  }
}

output "main_branch_subgraphs" {
  description = "Subgraphs in the main branch"
  value = {
    users = {
      id  = grafbase_subgraph.users_main.id
      url = grafbase_subgraph.users_main.url
    }
    products = {
      id  = grafbase_subgraph.products_main.id
      url = grafbase_subgraph.products_main.url
    }
    orders = {
      id  = grafbase_subgraph.orders_main.id
      url = grafbase_subgraph.orders_main.url
    }
  }
}

output "api_graph" {
  description = "API graph details"
  value = {
    id         = grafbase_graph.api.id
    slug       = grafbase_graph.api.slug
    created_at = grafbase_graph.api.created_at
    branch_id  = grafbase_branch.api_main.id
  }
}

output "analytics_graph" {
  description = "Analytics graph details"
  value = {
    id         = grafbase_graph.analytics.id
    slug       = grafbase_graph.analytics.slug
    created_at = grafbase_graph.analytics.created_at
    branch_id  = grafbase_branch.analytics_main.id
  }
}

output "all_graph_ids" {
  description = "List of all created graph IDs"
  value = [
    grafbase_graph.dev.id,
    grafbase_graph.api.id,
    grafbase_graph.analytics.id,
  ]
}

# Local values for reuse
locals {
  common_tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "my-application"
  }

  graph_slugs = {
    dev       = "my-app-${var.environment}"
    api       = "api-gateway-${var.environment}"
    analytics = "analytics-${var.environment}"
  }

  # Example subgraph configurations
  subgraph_configs = {
    main = {
      users    = "https://api.example.com/users/graphql"
      products = "https://api.example.com/products/graphql"
      orders   = "https://api.example.com/orders/graphql"
    }
    develop = {
      users    = "https://api-dev.example.com/users/graphql"
      products = "https://api-dev.example.com/products/graphql"
      orders   = "https://api-dev.example.com/orders/graphql"
    }
  }
}
