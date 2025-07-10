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
	ID                             types.String `tfsdk:"id"`
	AccountSlug                    types.String `tfsdk:"account_slug"`
	GraphSlug                      types.String `tfsdk:"graph_slug"`
	Name                           types.String `tfsdk:"name"`
	Environment                    types.String `tfsdk:"environment"`
	OperationChecksEnabled         types.Bool   `tfsdk:"operation_checks_enabled"`
	OperationChecksIgnoreUsageData types.Bool   `tfsdk:"operation_checks_ignore_usage_data"`
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
			"account_slug": schema.StringAttribute{
				MarkdownDescription: "Account slug where the branch belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"graph_slug": schema.StringAttribute{
				MarkdownDescription: "Graph slug where the branch belongs",
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
			"environment": schema.StringAttribute{
				MarkdownDescription: "Branch environment (PREVIEW or PRODUCTION)",
				Computed:            true,
			},
			"operation_checks_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether operation checks are enabled for this branch",
				Computed:            true,
				Optional:            true,
				Default:             nil,
			},
			"operation_checks_ignore_usage_data": schema.BoolAttribute{
				MarkdownDescription: "Whether usage data should be ignored when running operation checks",
				Computed:            true,
				Optional:            true,
				Default:             nil,
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
		AccountSlug: data.AccountSlug.ValueString(),
		GraphSlug:   data.GraphSlug.ValueString(),
		BranchName:  data.Name.ValueString(),
	}

	branch, err := r.client.CreateBranch(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create branch: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(branch.ID)
	data.Environment = types.StringValue(string(branch.Environment))
	data.OperationChecksEnabled = types.BoolValue(branch.OperationChecksEnabled)
	data.OperationChecksIgnoreUsageData = types.BoolValue(branch.OperationChecksIgnoreUsageData)

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

	// Get the branch using the account slug, graph slug, and branch name
	branch, err := r.client.GetBranch(ctx, data.AccountSlug.ValueString(), data.GraphSlug.ValueString(), data.Name.ValueString())
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
	data.Environment = types.StringValue(string(branch.Environment))
	data.OperationChecksEnabled = types.BoolValue(branch.OperationChecksEnabled)
	data.OperationChecksIgnoreUsageData = types.BoolValue(branch.OperationChecksIgnoreUsageData)

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

	// Currently, the Grafbase API doesn't support updating branches for the fields we expose
	// The account_slug, graph_slug, and name all have RequiresReplace plan modifiers
	// So this method should not be called in practice
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Branch updates are not supported. Changes to account_slug, graph_slug, or name require resource replacement.",
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
	deleteInput := client.DeleteBranchInput{
		AccountSlug: data.AccountSlug.ValueString(),
		GraphSlug:   data.GraphSlug.ValueString(),
		BranchName:  data.Name.ValueString(),
	}

	err := r.client.DeleteBranch(ctx, deleteInput)
	if err != nil {
		// If the branch doesn't exist, consider it already deleted
		if strings.Contains(err.Error(), "does not exist") {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete branch: %s", err))
		return
	}
}

func (r *BranchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID format: "account_slug/graph_slug/branch_name"
	// We'll parse this to get the account slug, graph slug, and branch name
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Invalid import ID format. Expected 'account_slug/graph_slug/branch_name', got: %s", req.ID))
		return
	}

	accountSlug := parts[0]
	graphSlug := parts[1]
	branchName := parts[2]

	// Set the account_slug, graph_slug, and name attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_slug"), accountSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_slug"), graphSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), branchName)...)

	// Get the branch to populate the remaining attributes
	branch, err := r.client.GetBranch(ctx, accountSlug, graphSlug, branchName)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read branch during import: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), branch.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), string(branch.Environment))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("operation_checks_enabled"), branch.OperationChecksEnabled)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("operation_checks_ignore_usage_data"), branch.OperationChecksIgnoreUsageData)...)
}
