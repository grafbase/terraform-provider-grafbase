package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafbase/terraform-provider-grafbase/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SubgraphResource{}
var _ resource.ResourceWithImportState = &SubgraphResource{}

func NewSubgraphResource() resource.Resource {
	return &SubgraphResource{}
}

// SubgraphResource defines the resource implementation.
type SubgraphResource struct {
	client *client.Client
}

// SubgraphResourceModel describes the resource data model.
type SubgraphResourceModel struct {
	ID        types.String `tfsdk:"id"`
	BranchID  types.String `tfsdk:"branch_id"`
	Name      types.String `tfsdk:"name"`
	URL       types.String `tfsdk:"url"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *SubgraphResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraph"
}

func (r *SubgraphResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Subgraph resource for managing Grafbase subgraphs.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Subgraph identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"branch_id": schema.StringAttribute{
				MarkdownDescription: "Branch ID where the subgraph belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subgraph name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Subgraph URL endpoint",
				Required:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Subgraph creation timestamp",
				Computed:            true,
			},
		},
	}
}

func (r *SubgraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubgraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubgraphResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the subgraph
	createInput := client.CreateSubgraphInput{
		BranchID:     data.BranchID.ValueString(),
		SubgraphName: data.Name.ValueString(),
		URL:          data.URL.ValueString(),
	}

	subgraph, err := r.client.CreateSubgraph(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create subgraph: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(subgraph.ID)
	data.CreatedAt = types.StringValue(subgraph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubgraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubgraphResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the subgraph using the branch ID and subgraph name
	subgraph, err := r.client.GetSubgraph(ctx, data.BranchID.ValueString(), data.Name.ValueString())
	if err != nil {
		// If subgraph is not found, remove it from state
		if err.Error() == "subgraph not found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subgraph: %s", err))
		return
	}

	// Update the model with the latest data
	data.ID = types.StringValue(subgraph.ID)
	data.URL = types.StringValue(subgraph.URL)
	data.CreatedAt = types.StringValue(subgraph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubgraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubgraphResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the subgraph URL (only URL can be updated)
	subgraph, err := r.client.UpdateSubgraph(ctx, data.ID.ValueString(), data.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update subgraph: %s", err))
		return
	}

	// Update the model with the latest data
	data.URL = types.StringValue(subgraph.URL)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubgraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubgraphResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the subgraph
	err := r.client.DeleteSubgraph(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete subgraph: %s", err))
		return
	}
}

func (r *SubgraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID format: "branch_id/subgraph_name"
	// We'll parse this to get both the branch ID and subgraph name
	branchID, subgraphName, err := parseSubgraphImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Invalid import ID format. Expected 'branch_id/subgraph_name', got: %s", req.ID))
		return
	}

	// Set the branch_id and name attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("branch_id"), branchID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), subgraphName)...)

	// Get the subgraph to populate the remaining attributes
	subgraph, err := r.client.GetSubgraph(ctx, branchID, subgraphName)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read subgraph during import: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), subgraph.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("url"), subgraph.URL)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), subgraph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))...)
}

// parseSubgraphImportID parses the import ID in the format "branch_id/subgraph_name"
func parseSubgraphImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid format: expected 'branch_id/subgraph_name'")
	}

	return parts[0], parts[1], nil
}