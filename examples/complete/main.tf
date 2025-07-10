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

# Feature branches for development graph
resource "grafbase_branch" "dev_feature" {
  account_slug = grafbase_graph.dev.account_slug
  graph_slug   = grafbase_graph.dev.slug
  name         = "feature-new-schema"
}

resource "grafbase_branch" "dev_hotfix" {
  account_slug = grafbase_graph.dev.account_slug
  graph_slug   = grafbase_graph.dev.slug
  name         = "hotfix-auth"
}

# Staging branch for API graph
resource "grafbase_branch" "api_staging" {
  account_slug = grafbase_graph.api.account_slug
  graph_slug   = grafbase_graph.api.slug
  name         = "staging"
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

output "dev_branches" {
  description = "Development graph branches"
  value = {
    feature = {
      id          = grafbase_branch.dev_feature.id
      name        = grafbase_branch.dev_feature.name
      environment = grafbase_branch.dev_feature.environment
    }
    hotfix = {
      id          = grafbase_branch.dev_hotfix.id
      name        = grafbase_branch.dev_hotfix.name
      environment = grafbase_branch.dev_hotfix.environment
    }
  }
}

output "api_staging_branch" {
  description = "API staging branch details"
  value = {
    id                       = grafbase_branch.api_staging.id
    name                     = grafbase_branch.api_staging.name
    environment              = grafbase_branch.api_staging.environment
    operation_checks_enabled = grafbase_branch.api_staging.operation_checks_enabled
  }
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
