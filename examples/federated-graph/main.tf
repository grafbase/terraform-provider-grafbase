terraform {
  required_providers {
    grafbase = {
      source = "grafbase/grafbase"
    }
  }
}

provider "grafbase" {
  api_key = var.grafbase_api_key
}

variable "grafbase_api_key" {
  description = "Grafbase API key"
  type        = string
  sensitive   = true
}

variable "account_slug" {
  description = "Account slug"
  type        = string
}

variable "graph_slug" {
  description = "Graph slug"
  type        = string
}

# Create the main graph
resource "grafbase_graph" "main" {
  account_slug = var.account_slug
  slug         = var.graph_slug
}

# Create main branch
resource "grafbase_branch" "main" {
  graph_id = grafbase_graph.main.id
  name     = "main"
}

# Create development branch
resource "grafbase_branch" "develop" {
  graph_id = grafbase_graph.main.id
  name     = "develop"
}

# Create subgraphs for main branch
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

# Create subgraphs for develop branch (with different URLs)
resource "grafbase_subgraph" "users_dev" {
  branch_id = grafbase_branch.develop.id
  name      = "users"
  url       = "https://users-dev.api.example.com/graphql"
}

resource "grafbase_subgraph" "products_dev" {
  branch_id = grafbase_branch.develop.id
  name      = "products"
  url       = "https://products-dev.api.example.com/graphql"
}

resource "grafbase_subgraph" "orders_dev" {
  branch_id = grafbase_branch.develop.id
  name      = "orders"
  url       = "https://orders-dev.api.example.com/graphql"
}

# Outputs
output "graph_id" {
  description = "The ID of the created graph"
  value       = grafbase_graph.main.id
}

output "main_branch_id" {
  description = "The ID of the main branch"
  value       = grafbase_branch.main.id
}

output "develop_branch_id" {
  description = "The ID of the develop branch"
  value       = grafbase_branch.develop.id
}

output "subgraphs_main" {
  description = "Subgraphs in the main branch"
  value = {
    users = {
      id  = grafbase_subgraph.users.id
      url = grafbase_subgraph.users.url
    }
    products = {
      id  = grafbase_subgraph.products.id
      url = grafbase_subgraph.products.url
    }
    orders = {
      id  = grafbase_subgraph.orders.id
      url = grafbase_subgraph.orders.url
    }
  }
}

output "subgraphs_develop" {
  description = "Subgraphs in the develop branch"
  value = {
    users = {
      id  = grafbase_subgraph.users_dev.id
      url = grafbase_subgraph.users_dev.url
    }
    products = {
      id  = grafbase_subgraph.products_dev.id
      url = grafbase_subgraph.products_dev.url
    }
    orders = {
      id  = grafbase_subgraph.orders_dev.id
      url = grafbase_subgraph.orders_dev.url
    }
  }
}