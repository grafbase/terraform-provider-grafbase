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

resource "grafbase_graph" "example" {
  account_slug = var.account_slug
  slug         = var.graph_slug
}

# Create branches for the graph
resource "grafbase_branch" "main" {
  account_slug = grafbase_graph.example.account_slug
  graph_slug   = grafbase_graph.example.slug
  name         = "main"
}

resource "grafbase_branch" "staging" {
  account_slug = grafbase_graph.example.account_slug
  graph_slug   = grafbase_graph.example.slug
  name         = "staging"
}

resource "grafbase_branch" "feature" {
  account_slug = grafbase_graph.example.account_slug
  graph_slug   = grafbase_graph.example.slug
  name         = "feature-branch"
}

output "graph_id" {
  description = "The ID of the created graph"
  value       = grafbase_graph.example.id
}

output "graph_created_at" {
  description = "When the graph was created"
  value       = grafbase_graph.example.created_at
}

output "branches" {
  description = "Information about created branches"
  value = {
    main = {
      id          = grafbase_branch.main.id
      name        = grafbase_branch.main.name
      environment = grafbase_branch.main.environment
    }
    staging = {
      id          = grafbase_branch.staging.id
      name        = grafbase_branch.staging.name
      environment = grafbase_branch.staging.environment
    }
    feature = {
      id          = grafbase_branch.feature.id
      name        = grafbase_branch.feature.name
      environment = grafbase_branch.feature.environment
    }
  }
}
