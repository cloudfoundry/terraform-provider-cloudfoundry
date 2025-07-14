package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &IsolationSegmentResource{}
	_ resource.ResourceWithConfigure   = &IsolationSegmentResource{}
	_ resource.ResourceWithImportState = &IsolationSegmentResource{}
)

// Instantiates an isolation segment resource.
func NewIsolationSegmentResource() resource.Resource {
	return &IsolationSegmentResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type IsolationSegmentResource struct {
	cfClient *cfv3client.Client
}

func (r *IsolationSegmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_isolation_segment"
}

func (r *IsolationSegmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an isolation segment resource for Cloud Foundry.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the isolation segment",
				Required:            true,
			},
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (r *IsolationSegmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IsolationSegmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IsolationSegmentType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createIsolationSegment, diags := plan.mapCreateIsolationSegmentTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isolationSegment, err := r.cfClient.IsolationSegments.Create(ctx, &createIsolationSegment)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Isolation Segment",
			"Could not create Isolation Segment with name "+plan.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	plan, diags = mapIsolationSegmentValuesToType(ctx, isolationSegment)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created an isolation segment resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *IsolationSegmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IsolationSegmentType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isolationSegment, err := rs.cfClient.IsolationSegments.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "isolation segment", data.Id.ValueString())
		return
	}

	data, diags = mapIsolationSegmentValuesToType(ctx, isolationSegment)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read an isolation segment resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *IsolationSegmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState IsolationSegmentType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateIsolationSegment, diags := plan.mapUpdateIsolationSegmentTypeToValues(ctx, previousState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isolationSegment, err := rs.cfClient.IsolationSegments.Update(ctx, plan.Id.ValueString(), &updateIsolationSegment)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Isolation Segment",
			"Could not update Isolation Segment with ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	data, diags := mapIsolationSegmentValuesToType(ctx, isolationSegment)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated an isolation segment resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *IsolationSegmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IsolationSegmentType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := rs.cfClient.IsolationSegments.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Isolation Segment",
			"Could not delete the Isolation Segment with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted an isolation segment resource")
}

func (rs *IsolationSegmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
