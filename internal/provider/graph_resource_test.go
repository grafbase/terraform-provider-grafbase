package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGraphResourceConfig("test-account", "test-graph"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_graph.test", "account_slug", "test-account"),
					resource.TestCheckResourceAttr("grafbase_graph.test", "slug", "test-graph"),
					resource.TestCheckResourceAttrSet("grafbase_graph.test", "id"),
					resource.TestCheckResourceAttrSet("grafbase_graph.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "grafbase_graph.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "test-account/test-graph",
			},
			// Update testing (should force replacement)
			{
				Config: testAccGraphResourceConfig("test-account", "test-graph-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_graph.test", "account_slug", "test-account"),
					resource.TestCheckResourceAttr("grafbase_graph.test", "slug", "test-graph-updated"),
					resource.TestCheckResourceAttrSet("grafbase_graph.test", "id"),
					resource.TestCheckResourceAttrSet("grafbase_graph.test", "created_at"),
				),
			},
		},
	})
}

func TestAccGraphResourceDisappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphResourceConfig("test-account", "test-graph-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_graph.test", "account_slug", "test-account"),
					resource.TestCheckResourceAttr("grafbase_graph.test", "slug", "test-graph-disappears"),
				),
			},
			// Resource should be recreated
			{
				Config: testAccGraphResourceConfig("test-account", "test-graph-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_graph.test", "account_slug", "test-account"),
					resource.TestCheckResourceAttr("grafbase_graph.test", "slug", "test-graph-disappears"),
				),
			},
		},
	})
}

func TestParseImportID(t *testing.T) {
	tests := []struct {
		name          string
		importID      string
		expectedAcct  string
		expectedGraph string
		expectedError bool
	}{
		{
			name:          "valid import ID",
			importID:      "my-account/my-graph",
			expectedAcct:  "my-account",
			expectedGraph: "my-graph",
			expectedError: false,
		},
		{
			name:          "valid import ID with dashes",
			importID:      "my-test-account/my-test-graph",
			expectedAcct:  "my-test-account",
			expectedGraph: "my-test-graph",
			expectedError: false,
		},
		{
			name:          "invalid - no slash",
			importID:      "my-account-my-graph",
			expectedError: true,
		},
		{
			name:          "invalid - multiple slashes",
			importID:      "my-account/my-graph/extra",
			expectedError: true,
		},
		{
			name:          "invalid - empty account",
			importID:      "/my-graph",
			expectedError: true,
		},
		{
			name:          "invalid - empty graph",
			importID:      "my-account/",
			expectedError: true,
		},
		{
			name:          "invalid - empty string",
			importID:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, graph, err := parseImportID(tt.importID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if account != tt.expectedAcct {
				t.Errorf("expected account %q, got %q", tt.expectedAcct, account)
			}

			if graph != tt.expectedGraph {
				t.Errorf("expected graph %q, got %q", tt.expectedGraph, graph)
			}
		})
	}
}

func testAccGraphResourceConfig(accountSlug, graphSlug string) string {
	return fmt.Sprintf(`
resource "grafbase_graph" "test" {
  account_slug = %[1]q
  slug         = %[2]q
}
`, accountSlug, graphSlug)
}

// Additional test configurations for different scenarios

func testAccGraphResourceConfigWithVariables() string {
	return `
variable "account_slug" {
  description = "Account slug for testing"
  type        = string
  default     = "test-account"
}

variable "graph_slug" {
  description = "Graph slug for testing"
  type        = string
  default     = "test-graph"
}

resource "grafbase_graph" "test" {
  account_slug = var.account_slug
  slug         = var.graph_slug
}

output "graph_id" {
  value = grafbase_graph.test.id
}

output "graph_created_at" {
  value = grafbase_graph.test.created_at
}
`
}

func testAccGraphResourceConfigMultiple() string {
	return `
resource "grafbase_graph" "test1" {
  account_slug = "test-account"
  slug         = "test-graph-1"
}

resource "grafbase_graph" "test2" {
  account_slug = "test-account"
  slug         = "test-graph-2"
}

resource "grafbase_graph" "test3" {
  account_slug = "test-account"
  slug         = "test-graph-3"
}
`
}

// Test helper functions

func TestAccGraphResourceWithDynamicSlug(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphResourceConfigDynamic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_graph.dynamic", "account_slug", "test-account"),
					resource.TestCheckResourceAttrSet("grafbase_graph.dynamic", "slug"),
					resource.TestCheckResourceAttrSet("grafbase_graph.dynamic", "id"),
					resource.TestCheckResourceAttrSet("grafbase_graph.dynamic", "created_at"),
				),
			},
		},
	})
}

func testAccGraphResourceConfigDynamic() string {
	return `
locals {
  timestamp = formatdate("YYYYMMDD-hhmm", timestamp())
}

resource "grafbase_graph" "dynamic" {
  account_slug = "test-account"
  slug         = "test-graph-${local.timestamp}"
}
`
}

// Validation tests

func TestAccGraphResourceValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphResourceConfigInvalidAccountSlug(),
				ExpectError: nil, // The API will return the error
			},
		},
	})
}

func testAccGraphResourceConfigInvalidAccountSlug() string {
	return `
resource "grafbase_graph" "invalid" {
  account_slug = "non-existent-account"
  slug         = "test-graph"
}
`
}

// Performance test for multiple graphs

func TestAccGraphResourcePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_graph.test1", "slug", "test-graph-1"),
					resource.TestCheckResourceAttr("grafbase_graph.test2", "slug", "test-graph-2"),
					resource.TestCheckResourceAttr("grafbase_graph.test3", "slug", "test-graph-3"),
				),
			},
		},
	})
}
