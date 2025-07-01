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

output "graph_id" {
  description = "The ID of the created graph"
  value       = grafbase_graph.example.id
}

output "graph_created_at" {
  description = "When the graph was created"
  value       = grafbase_graph.example.created_at
}
