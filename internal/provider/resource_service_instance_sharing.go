package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type serviceInstanceSharingResource struct {
	cfClient *cfv3client.Client
}

var (
	_ resource.Resource              = &serviceInstanceSharingResource{}
	_ resource.ResourceWithConfigure = &SpaceResource{}
)

func NewServiceInstanceSharingResource() resource.Resource {
	return &serviceInstanceSharingResource{}
}

func (r *serviceInstanceSharingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_instance_sharing"
}

func (r *serviceInstanceSharingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a resource for managing service instance sharing in Cloud Foundry.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service instance sharing resource. Consists of the space id and the service instance id",
				Computed:            true,
			},
			"service_instance_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service instance to share.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the space to share the service instance with.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *serviceInstanceSharingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceInstanceSharingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceInstanceSharingType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.cfClient.ServiceInstances.ShareWithSpace(ctx, plan.ServiceInstanceId.ValueString(), plan.SpaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error sharing service instance with space", err.Error())
		return
	}

	computedID := fmt.Sprintf("%s/%s", plan.ServiceInstanceId.ValueString(), plan.SpaceId.ValueString())

	newState := ServiceInstanceSharingType{
		Id:                types.StringValue(computedID),
		ServiceInstanceId: plan.ServiceInstanceId,
		SpaceId:           plan.SpaceId,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *serviceInstanceSharingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceInstanceSharingType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	relationship, err := r.cfClient.ServiceInstances.GetSharedSpaceRelationships(ctx, data.ServiceInstanceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error when getting shared spaces for service instance", err.Error())
		return
	}

	data = mapRelationShipToType(relationship, data.ServiceInstanceId.ValueString())

	tflog.Trace(ctx, "read a service instance sharing resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serviceInstanceSharingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update method needed since the resource is immutable
	// The method needs to exist to satisfy the interface
}

func (r *serviceInstanceSharingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state ServiceInstanceSharingType

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cfClient.ServiceInstances.UnShareWithSpace(ctx, state.ServiceInstanceId.ValueString(), state.SpaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error unsharing service instance with space", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted a service instance sharing resource")
}

func mapRelationShipToType(relationship *cfv3resource.ServiceInstanceSharedSpaceRelationships, serviceInstanceId string) ServiceInstanceSharingType {
	spaceItGetsSharedTo := relationship.Data[0].GUID
	id := types.StringValue(serviceInstanceId + "/" + spaceItGetsSharedTo)

	return ServiceInstanceSharingType{
		Id:                id,
		ServiceInstanceId: types.StringValue(serviceInstanceId),
		SpaceId:           types.StringValue(spaceItGetsSharedTo),
	}
}
