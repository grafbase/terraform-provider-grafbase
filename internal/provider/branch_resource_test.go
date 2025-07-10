package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBranchResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBranchResourceConfig("test-branch"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_branch.test", "name", "test-branch"),
					resource.TestCheckResourceAttr("grafbase_branch.test", "account_slug", "test-account"),
					resource.TestCheckResourceAttr("grafbase_branch.test", "graph_slug", "test-graph"),
					resource.TestCheckResourceAttr("grafbase_branch.test", "environment", "PREVIEW"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "id"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "operation_checks_enabled"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "operation_checks_ignore_usage_data"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "grafbase_branch.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "test-account/test-graph/test-branch",
			},
			// Update and Read testing
			{
				Config: testAccBranchResourceConfig("test-branch-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_branch.test", "name", "test-branch-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccBranchResource_MultipleGraphs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create branches in different graphs
			{
				Config: testAccBranchResourceConfig_MultipleGraphs(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First branch
					resource.TestCheckResourceAttr("grafbase_branch.branch1", "name", "feature-1"),
					resource.TestCheckResourceAttr("grafbase_branch.branch1", "graph_slug", "graph-1"),
					resource.TestCheckResourceAttrSet("grafbase_branch.branch1", "id"),
					// Second branch
					resource.TestCheckResourceAttr("grafbase_branch.branch2", "name", "feature-2"),
					resource.TestCheckResourceAttr("grafbase_branch.branch2", "graph_slug", "graph-2"),
					resource.TestCheckResourceAttrSet("grafbase_branch.branch2", "id"),
				),
			},
		},
	})
}

func testAccBranchResourceConfig(branchName string) string {
	return fmt.Sprintf(`
resource "grafbase_graph" "test" {
  account_slug = "test-account"
  slug         = "test-graph"
}

resource "grafbase_branch" "test" {
  account_slug = grafbase_graph.test.account_slug
  graph_slug   = grafbase_graph.test.slug
  name         = %[1]q
}
`, branchName)
}

func testAccBranchResourceConfig_MultipleGraphs() string {
	return `
resource "grafbase_graph" "graph1" {
  account_slug = "test-account"
  slug         = "graph-1"
}

resource "grafbase_graph" "graph2" {
  account_slug = "test-account"
  slug         = "graph-2"
}

resource "grafbase_branch" "branch1" {
  account_slug = grafbase_graph.graph1.account_slug
  graph_slug   = grafbase_graph.graph1.slug
  name         = "feature-1"
}

resource "grafbase_branch" "branch2" {
  account_slug = grafbase_graph.graph2.account_slug
  graph_slug   = grafbase_graph.graph2.slug
  name         = "feature-2"
}
`
}
