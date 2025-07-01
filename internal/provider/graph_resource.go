package provider

import (
	"context"
	"fmt"

	"github.com/grafbase/terraform-provider-grafbase/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GraphResource{}
var _ resource.ResourceWithImportState = &GraphResource{}

func NewGraphResource() resource.Resource {
	return &GraphResource{}
}

// GraphResource defines the resource implementation.
type GraphResource struct {
	client *client.Client
}

// GraphResourceModel describes the resource data model.
type GraphResourceModel struct {
	ID          types.String `tfsdk:"id"`
	AccountSlug types.String `tfsdk:"account_slug"`
	Slug        types.String `tfsdk:"slug"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *GraphResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph"
}

func (r *GraphResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Graph resource for managing Grafbase graphs.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Graph identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_slug": schema.StringAttribute{
				MarkdownDescription: "Account slug where the graph belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Graph slug",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Graph creation timestamp",
				Computed:            true,
			},
		},
	}
}

func (r *GraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *GraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GraphResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// First, get the account ID by slug
	account, err := r.client.GetAccountBySlug(ctx, data.AccountSlug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get account: %s", err))
		return
	}

	// Create the graph
	createInput := client.CreateGraphInput{
		AccountID: account.ID,
		GraphSlug: data.Slug.ValueString(),
	}

	graph, err := r.client.CreateGraph(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create graph: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(graph.ID)
	data.CreatedAt = types.StringValue(graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GraphResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the graph using the account slug and graph slug
	graph, err := r.client.GetGraph(ctx, data.AccountSlug.ValueString(), data.Slug.ValueString())
	if err != nil {
		// If graph is not found, remove it from state
		if err.Error() == "graph not found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read graph: %s", err))
		return
	}

	// Update the model with the latest data
	data.ID = types.StringValue(graph.ID)
	data.CreatedAt = types.StringValue(graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GraphResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Currently, the Grafbase API doesn't support updating graphs
	// The account_slug and slug both have RequiresReplace plan modifiers
	// So this method should not be called in practice
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Graph updates are not supported. Changes to account_slug or slug require resource replacement.",
	)
}

func (r *GraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GraphResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the graph
	err := r.client.DeleteGraph(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete graph: %s", err))
		return
	}
}

func (r *GraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID format: "account_slug/graph_slug"
	// We'll parse this to get both the account slug and graph slug
	accountSlug, graphSlug, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Invalid import ID format. Expected 'account_slug/graph_slug', got: %s", req.ID))
		return
	}

	// Set the account_slug and slug attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_slug"), accountSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), graphSlug)...)

	// Get the graph to populate the remaining attributes
	graph, err := r.client.GetGraph(ctx, accountSlug, graphSlug)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read graph during import: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), graph.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))...)
}

// parseImportID parses the import ID in the format "account_slug/graph_slug"
func parseImportID(id string) (string, string, error) {
	parts := []rune(id)
	slashIndex := -1

	for i, r := range parts {
		if r == '/' {
			if slashIndex == -1 {
				slashIndex = i
			} else {
				// Multiple slashes, invalid format
				return "", "", fmt.Errorf("invalid format: multiple slashes found")
			}
		}
	}

	if slashIndex == -1 || slashIndex == 0 || slashIndex == len(parts)-1 {
		return "", "", fmt.Errorf("invalid format: expected 'account_slug/graph_slug'")
	}

	accountSlug := string(parts[:slashIndex])
	graphSlug := string(parts[slashIndex+1:])

	return accountSlug, graphSlug, nil
}
