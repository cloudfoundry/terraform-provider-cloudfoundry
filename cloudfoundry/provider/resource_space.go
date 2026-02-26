package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &SpaceResource{}
	_ resource.ResourceWithConfigure   = &SpaceResource{}
	_ resource.ResourceWithImportState = &SpaceResource{}
	_ resource.ResourceWithIdentity    = &SpaceResource{}
)

// Instantiates a space resource.
func NewSpaceResource() resource.Resource {
	return &SpaceResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SpaceResource struct {
	cfClient *cfv3client.Client
}

type SpaceResouerceIdentityModel struct {
	SpaceGUID types.String `tfsdk:"space_guid"`
}

func (r *SpaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *SpaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource for managing Cloud Foundry spaces within organizations.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Space in Cloud Foundry",
				Required:            true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The ID of the Org within which to create the space",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"quota": schema.StringAttribute{
				MarkdownDescription: "The space quota applied to the space. To assign a space quota, use the space quota resource instead.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_ssh": schema.BoolAttribute{
				MarkdownDescription: "Allows SSH to application containers via the CF CLI.",
				Computed:            true,
				Optional:            true,
			},
			"isolation_segment": schema.StringAttribute{
				MarkdownDescription: "The ID of the isolation segment to assign to the space. The isolation segment must be entitled to the space's parent organization",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			idKey:          guidSchema(),
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (rs *SpaceResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_guid": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *SpaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.cfClient = session.CFClient
}

func (r *SpaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan spaceType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createSpace, diags := plan.mapCreateSpaceTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	space, err := r.cfClient.Spaces.Create(ctx, &createSpace)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Space",
			"Could not create Space "+plan.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	var allowSSH bool
	if !plan.AllowSSH.IsUnknown() {
		err = r.cfClient.SpaceFeatures.EnableSSH(ctx, space.GUID, plan.AllowSSH.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Setting Space SSH",
				"Could not set the SSH feature value on space "+space.Name+" : "+err.Error(),
			)
		}
		allowSSH = plan.AllowSSH.ValueBool()
	} else {
		allowSSH, err = r.cfClient.SpaceFeatures.IsSSHEnabled(ctx, space.GUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Space SSH",
				"Could not get the SSH feature value of space "+space.Name+" : "+err.Error(),
			)
			return
		}
	}

	if !plan.IsolationSegment.IsNull() {
		err = r.cfClient.Spaces.AssignIsolationSegment(ctx, space.GUID, plan.IsolationSegment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Assigning Isolation Segment",
				"Could not assign the Isolation Segment with ID "+plan.IsolationSegment.ValueString()+" on space "+space.Name+": "+err.Error(),
			)
		}
	}

	plan, diags = mapSpaceValuesToType(ctx, space, allowSSH, plan.IsolationSegment.ValueString())
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a space resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := SpaceResouerceIdentityModel{
		SpaceGUID: types.StringValue(plan.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (rs *SpaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data spaceType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	space, err := rs.cfClient.Spaces.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "space", data.Id.ValueString())
		return
	}

	sshEnabled, err := rs.cfClient.SpaceFeatures.IsSSHEnabled(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching SSH Feature",
			"Could not get the SSH feature value of space "+space.Name+" : "+err.Error(),
		)
		return
	}

	isolationSegment, err := rs.cfClient.Spaces.GetAssignedIsolationSegment(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Isolation Segment",
			"Could not get the Isolation Segment of space "+space.Name+": "+err.Error(),
		)
		return
	}

	data, diags = mapSpaceValuesToType(ctx, space, sshEnabled, isolationSegment)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a space resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var identity SpaceResouerceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = SpaceResouerceIdentityModel{
			SpaceGUID: types.StringValue(data.Id.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

func (rs *SpaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState spaceType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if !plan.AllowSSH.IsUnknown() {
		err = rs.cfClient.SpaceFeatures.EnableSSH(ctx, plan.Id.ValueString(), plan.AllowSSH.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Updating Space SSH",
				"Could not set the SSH feature value on space with ID "+plan.Id.ValueString()+": "+err.Error(),
			)
		}
	}

	err = rs.cfClient.Spaces.AssignIsolationSegment(ctx, plan.Id.ValueString(), plan.IsolationSegment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Isolation Segment",
			"Could not assign the Isolation Segment with ID "+plan.IsolationSegment.ValueString()+" on space with ID "+plan.Id.ValueString()+": "+err.Error(),
		)
	}

	updateSpace, diags := plan.mapUpdateSpaceTypeToValues(ctx, previousState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	space, err := rs.cfClient.Spaces.Update(ctx, plan.Id.ValueString(), &updateSpace)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Space",
			"Could not update Space with ID "+plan.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data, diags := mapSpaceValuesToType(ctx, space, plan.AllowSSH.ValueBool(), plan.IsolationSegment.ValueString())
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a space resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// WORKAROUND for OpenTofu compatibility
	// https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/418
	identity := SpaceResouerceIdentityModel{
		SpaceGUID: types.StringValue(data.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
	// END WORKAROUND
}

func (rs *SpaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state spaceType

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.Spaces.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Space",
			"Could not delete the space with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Space",
			"Failed in deleting the space with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a space resource")

}

func (rs *SpaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("space_guid"), req, resp)
}
