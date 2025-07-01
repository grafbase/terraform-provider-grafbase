package provider

import (
	"context"
	"os"

	"github.com/grafbase/terraform-provider-grafbase/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure GrafbaseProvider satisfies various provider interfaces.
var _ provider.Provider = &GrafbaseProvider{}

// GrafbaseProvider defines the provider implementation.
type GrafbaseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GrafbaseProviderModel describes the provider data model.
type GrafbaseProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (p *GrafbaseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "grafbase"
	resp.Version = p.version
}

func (p *GrafbaseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Grafbase API key for authentication. Can also be set via the `GRAFBASE_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *GrafbaseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GrafbaseProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	var apiKey string
	if data.APIKey.IsNull() {
		apiKey = os.Getenv("GRAFBASE_API_KEY")
	} else {
		apiKey = data.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Unable to find API key",
			"API key cannot be an empty string. "+
				"Set the api_key attribute in the provider configuration or use the GRAFBASE_API_KEY environment variable.",
		)
		return
	}

	// Create a new Grafbase client using the configuration values
	client := client.NewClient(apiKey)

	// Make the client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *GrafbaseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGraphResource,
	}
}

func (p *GrafbaseProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources can be added later
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GrafbaseProvider{
			version: version,
		}
	}
}
