package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"grafbase": providerserver.NewProtocol6WithError(New("test")()),
}

func TestUnit(t *testing.T) {
	t.Run("provider", func(t *testing.T) {
		New("test")()
	})
}

func testAccPreCheck(t *testing.T) {
	// You can add common test setup here
}
