package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSubgraphResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSubgraphResourceConfig("test-account", "test-graph", "main", "users", "https://api.example.com/users/graphql"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_subgraph.test", "name", "users"),
					resource.TestCheckResourceAttr("grafbase_subgraph.test", "url", "https://api.example.com/users/graphql"),
					resource.TestCheckResourceAttrSet("grafbase_subgraph.test", "id"),
					resource.TestCheckResourceAttrSet("grafbase_subgraph.test", "branch_id"),
					resource.TestCheckResourceAttrSet("grafbase_subgraph.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "grafbase_subgraph.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "test-branch-id/users",
				ImportStateVerifyIdentifierAttribute: "name",
			},
			// Update and Read testing (only URL can be updated)
			{
				Config: testAccSubgraphResourceConfig("test-account", "test-graph", "main", "users", "https://api-v2.example.com/users/graphql"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_subgraph.test", "name", "users"),
					resource.TestCheckResourceAttr("grafbase_subgraph.test", "url", "https://api-v2.example.com/users/graphql"),
					resource.TestCheckResourceAttrSet("grafbase_subgraph.test", "id"),
				),
			},
			// Test name change (should trigger replace)
			{
				Config: testAccSubgraphResourceConfig("test-account", "test-graph", "main", "products", "https://api-v2.example.com/products/graphql"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_subgraph.test", "name", "products"),
					resource.TestCheckResourceAttr("grafbase_subgraph.test", "url", "https://api-v2.example.com/products/graphql"),
					resource.TestCheckResourceAttrSet("grafbase_subgraph.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSubgraphResourceConfig(accountSlug, graphSlug, branchName, subgraphName, url string) string {
	return fmt.Sprintf(`
resource "grafbase_graph" "test" {
  account_slug = "%s"
  slug         = "%s"
}

resource "grafbase_branch" "test" {
  graph_id = grafbase_graph.test.id
  name     = "%s"
}

resource "grafbase_subgraph" "test" {
  branch_id = grafbase_branch.test.id
  name      = "%s"
  url       = "%s"
}
`, accountSlug, graphSlug, branchName, subgraphName, url)
}