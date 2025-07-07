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
				Config: testAccBranchResourceConfig("test-account", "test-graph", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_branch.test", "name", "main"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "id"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "graph_id"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "grafbase_branch.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "test-graph-id/main",
				ImportStateVerifyIdentifierAttribute: "name",
			},
			// Update and Read testing (name should trigger replace)
			{
				Config: testAccBranchResourceConfig("test-account", "test-graph", "develop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafbase_branch.test", "name", "develop"),
					resource.TestCheckResourceAttrSet("grafbase_branch.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccBranchResourceConfig(accountSlug, graphSlug, branchName string) string {
	return fmt.Sprintf(`
resource "grafbase_graph" "test" {
  account_slug = "%s"
  slug         = "%s"
}

resource "grafbase_branch" "test" {
  graph_id = grafbase_graph.test.id
  name     = "%s"
}
`, accountSlug, graphSlug, branchName)
}