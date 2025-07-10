terraform {
  required_providers {
    grafbase = {
      source = "grafbase/grafbase"
      # For local development, you can use a local build
      # version = "1.0.0"
    }
  }
}

# Provider configuration for local development
provider "grafbase" {
  # For local development, set your API key via environment variable:
  # export GRAFBASE_API_KEY="your-api-key-here"

  # Alternatively, you can set it directly (not recommended for production):
  # api_key = "your-api-key-here"
}

# Variables for local development
variable "account_slug" {
  description = "Your account slug"
  type        = string
  default     = "my-account"
}

variable "developer_name" {
  description = "Developer name for graph naming"
  type        = string
  default     = "dev"
}

# Local development graph
resource "grafbase_graph" "local_dev" {
  account_slug = var.account_slug
  slug         = "local-${var.developer_name}-${formatdate("YYYYMMDD", timestamp())}"
}

# Feature branch graph (useful for testing specific features)
resource "grafbase_graph" "feature_branch" {
  account_slug = var.account_slug
  slug         = "feature-${var.developer_name}-test"
}

# Development branches
resource "grafbase_branch" "dev_main" {
  account_slug = grafbase_graph.local_dev.account_slug
  graph_slug   = grafbase_graph.local_dev.slug
  name         = "main"
}

resource "grafbase_branch" "dev_feature" {
  account_slug = grafbase_graph.local_dev.account_slug
  graph_slug   = grafbase_graph.local_dev.slug
  name         = "feature-auth-improvements"
}

resource "grafbase_branch" "dev_experiment" {
  account_slug = grafbase_graph.local_dev.account_slug
  graph_slug   = grafbase_graph.local_dev.slug
  name         = "experiment-new-schema"
}

# Feature test branches
resource "grafbase_branch" "feature_test" {
  account_slug = grafbase_graph.feature_branch.account_slug
  graph_slug   = grafbase_graph.feature_branch.slug
  name         = "test-integration"
}

# Outputs for local development
output "local_dev_graph_info" {
  description = "Local development graph information"
  value = {
    id         = grafbase_graph.local_dev.id
    slug       = grafbase_graph.local_dev.slug
    created_at = grafbase_graph.local_dev.created_at
  }
}

output "feature_branch_graph_info" {
  description = "Feature branch graph information"
  value = {
    id         = grafbase_graph.feature_branch.id
    slug       = grafbase_graph.feature_branch.slug
    created_at = grafbase_graph.feature_branch.created_at
  }
}

output "development_branches" {
  description = "All development branches"
  value = {
    main = {
      id          = grafbase_branch.dev_main.id
      name        = grafbase_branch.dev_main.name
      environment = grafbase_branch.dev_main.environment
    }
    feature = {
      id          = grafbase_branch.dev_feature.id
      name        = grafbase_branch.dev_feature.name
      environment = grafbase_branch.dev_feature.environment
    }
    experiment = {
      id          = grafbase_branch.dev_experiment.id
      name        = grafbase_branch.dev_experiment.name
      environment = grafbase_branch.dev_experiment.environment
    }
  }
}

output "feature_test_branch" {
  description = "Feature test branch information"
  value = {
    id                                 = grafbase_branch.feature_test.id
    name                               = grafbase_branch.feature_test.name
    environment                        = grafbase_branch.feature_test.environment
    operation_checks_enabled           = grafbase_branch.feature_test.operation_checks_enabled
    operation_checks_ignore_usage_data = grafbase_branch.feature_test.operation_checks_ignore_usage_data
  }
}

# Local values for development
locals {
  dev_config = {
    environment = "development"
    managed_by  = "terraform-local"
    developer   = var.developer_name
  }
}

# Instructions for local development (as comments)
# 1. Set your Grafbase API key:
#    export GRAFBASE_API_KEY="your-api-key-here"
#
# 2. Set your account slug:
#    export TF_VAR_account_slug="your-account-slug"
#
# 3. Optionally set your developer name:
#    export TF_VAR_developer_name="your-name"
#
# 4. Initialize and apply:
#    terraform init
#    terraform plan
#    terraform apply
#
# 5. When done developing, clean up:
#    terraform destroy
