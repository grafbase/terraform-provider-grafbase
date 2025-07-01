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

# API graph for microservices
resource "grafbase_graph" "api" {
  account_slug = var.account_slug
  slug         = "api-gateway-${var.environment}"
}

# Analytics graph
resource "grafbase_graph" "analytics" {
  account_slug = var.account_slug
  slug         = "analytics-${var.environment}"
}

# Outputs for other Terraform configurations or CI/CD systems
output "dev_graph" {
  description = "Development graph details"
  value = {
    id         = grafbase_graph.dev.id
    slug       = grafbase_graph.dev.slug
    created_at = grafbase_graph.dev.created_at
  }
}

output "api_graph" {
  description = "API graph details"
  value = {
    id         = grafbase_graph.api.id
    slug       = grafbase_graph.api.slug
    created_at = grafbase_graph.api.created_at
  }
}

output "analytics_graph" {
  description = "Analytics graph details"
  value = {
    id         = grafbase_graph.analytics.id
    slug       = grafbase_graph.analytics.slug
    created_at = grafbase_graph.analytics.created_at
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
}
