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
var _ resource.Resource = &BranchResource{}
var _ resource.ResourceWithImportState = &BranchResource{}

func NewBranchResource() resource.Resource {
	return &BranchResource{}
}

// BranchResource defines the resource implementation.
type BranchResource struct {
	client *client.Client
}

// BranchResourceModel describes the resource data model.
type BranchResourceModel struct {
	ID          types.String `tfsdk:"id"`
	GraphID     types.String `tfsdk:"graph_id"`
	Name        types.String `tfsdk:"name"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *BranchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch"
}

func (r *BranchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Branch resource for managing Grafbase branches.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Branch identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"graph_id": schema.StringAttribute{
				MarkdownDescription: "Graph ID where the branch belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Branch name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Branch creation timestamp",
				Computed:            true,
			},
		},
	}
}

func (r *BranchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BranchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BranchResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the branch
	createInput := client.CreateBranchInput{
		GraphID:    data.GraphID.ValueString(),
		BranchName: data.Name.ValueString(),
	}

	branch, err := r.client.CreateBranch(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create branch: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(branch.ID)
	data.CreatedAt = types.StringValue(branch.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BranchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BranchResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the branch using the graph ID and branch name
	branch, err := r.client.GetBranch(ctx, data.GraphID.ValueString(), data.Name.ValueString())
	if err != nil {
		// If branch is not found, remove it from state
		if err.Error() == "branch not found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read branch: %s", err))
		return
	}

	// Update the model with the latest data
	data.ID = types.StringValue(branch.ID)
	data.CreatedAt = types.StringValue(branch.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BranchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BranchResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Currently, the Grafbase API doesn't support updating branches
	// The graph_id and name both have RequiresReplace plan modifiers
	// So this method should not be called in practice
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Branch updates are not supported. Changes to graph_id or name require resource replacement.",
	)
}

func (r *BranchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BranchResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the branch
	err := r.client.DeleteBranch(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete branch: %s", err))
		return
	}
}

func (r *BranchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID format: "graph_id/branch_name"
	// We'll parse this to get both the graph ID and branch name
	graphID, branchName, err := parseBranchImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Invalid import ID format. Expected 'graph_id/branch_name', got: %s", req.ID))
		return
	}

	// Set the graph_id and name attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), graphID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), branchName)...)

	// Get the branch to populate the remaining attributes
	branch, err := r.client.GetBranch(ctx, graphID, branchName)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read branch during import: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), branch.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), branch.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))...)
}

// parseBranchImportID parses the import ID in the format "graph_id/branch_name"
func parseBranchImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid format: expected 'graph_id/branch_name'")
	}

	return parts[0], parts[1], nil
}