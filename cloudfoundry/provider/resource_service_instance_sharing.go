package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
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
	_ resource.ResourceWithConfigure = &serviceInstanceSharingResource{}
	_ resource.ResourceWithIdentity  = &serviceInstanceSharingResource{}
)

type serviceInstanceSharingResourceIdentityModel struct {
	ServiceInstanceGUID types.String `tfsdk:"service_instance_guid"`
}

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
			idKey: guidSchema(),
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The ID of the service instance to share.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"spaces": schema.SetAttribute{
				MarkdownDescription: "The IDs of the spaces to share the service instance with.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (rs *serviceInstanceSharingResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_instance_guid": identityschema.StringAttribute{
				RequiredForImport: true,
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

	spaces := make([]string, len(plan.Spaces.Elements()))
	tempDiags := plan.Spaces.ElementsAs(ctx, &spaces, false)
	if tempDiags.HasError() {
		resp.Diagnostics.Append(tempDiags...)
		return
	}

	_, err := r.cfClient.ServiceInstances.ShareWithSpaces(ctx, plan.ServiceInstance.ValueString(), spaces)

	if err != nil {
		resp.Diagnostics.AddError("Error sharing service instance with spaces", err.Error())
		return
	}

	tflog.Trace(ctx, "created a service instance sharing resource")
	newState := ServiceInstanceSharingType{
		Id:              plan.ServiceInstance,
		ServiceInstance: plan.ServiceInstance,
		Spaces:          plan.Spaces,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)

	identity := serviceInstanceSharingResourceIdentityModel{
		ServiceInstanceGUID: types.StringValue(newState.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *serviceInstanceSharingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceInstanceSharingType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceInstanceID := data.Id.ValueString()

	if serviceInstanceID == "" {
		serviceInstanceID = data.ServiceInstance.ValueString()
	}

	relationship, err := r.cfClient.ServiceInstances.GetSharedSpaceRelationships(ctx, serviceInstanceID)
	if err != nil {
		resp.Diagnostics.AddError("Error when getting shared spaces for service instance", err.Error())
		return
	}

	data = mapSharedSpacesValuesToType(relationship, serviceInstanceID)

	tflog.Trace(ctx, "read a service instance sharing resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var identity serviceInstanceSharingResourceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = serviceInstanceSharingResourceIdentityModel{
			ServiceInstanceGUID: types.StringValue(data.Id.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *serviceInstanceSharingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState ServiceInstanceSharingType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planSpaces []string
	var previousSpaces []string
	tempDiags := plan.Spaces.ElementsAs(ctx, &planSpaces, false)
	if tempDiags.HasError() {
		resp.Diagnostics.Append(tempDiags...)
		return
	}
	tempDiags = previousState.Spaces.ElementsAs(ctx, &previousSpaces, false)
	if tempDiags.HasError() {
		resp.Diagnostics.Append(tempDiags...)
		return
	}

	// Create maps for lookup
	planSpacesMap := make(map[string]bool)
	previousSpacesMap := make(map[string]bool)
	for _, space := range planSpaces {
		planSpacesMap[space] = true
	}
	for _, space := range previousSpaces {
		previousSpacesMap[space] = true
	}

	// Find spaces to be added (in plan but not in previousState)
	var spacesToAdd []string
	for _, space := range planSpaces {
		if !previousSpacesMap[space] {
			spacesToAdd = append(spacesToAdd, space)
		}
	}

	// Find spaces to be removed (in previousState but not in plan)
	var spacesToRemove []string
	for _, space := range previousSpaces {
		if !planSpacesMap[space] {
			spacesToRemove = append(spacesToRemove, space)
		}
	}

	tflog.Trace(ctx, "Spaces diff", map[string]any{
		"spaces_to_add":    spacesToAdd,
		"spaces_to_remove": spacesToRemove,
	})

	serviceInstanceID := plan.ServiceInstance.ValueString()

	if len(spacesToRemove) > 0 {
		err := r.cfClient.ServiceInstances.UnShareWithSpaces(ctx, serviceInstanceID, spacesToRemove)
		if err != nil {
			resp.Diagnostics.AddError("Error unsharing service instance with spaces", err.Error())
			return
		}
		tflog.Debug(ctx, "Unshared service instance with spaces", map[string]any{
			"service_instance": serviceInstanceID,
			"spaces":           spacesToRemove,
		})
	}
	if len(spacesToAdd) > 0 {
		_, err := r.cfClient.ServiceInstances.ShareWithSpaces(ctx, serviceInstanceID, spacesToAdd)
		if err != nil {
			resp.Diagnostics.AddError("Error sharing service instance with spaces", err.Error())
			return
		}
		tflog.Debug(ctx, "Shared service instance with spaces", map[string]any{
			"service_instance": serviceInstanceID,
			"spaces":           spacesToAdd,
		})
	}

	newState := ServiceInstanceSharingType{
		Id:              plan.ServiceInstance,
		ServiceInstance: plan.ServiceInstance,
		Spaces:          plan.Spaces,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	tflog.Trace(ctx, "updated a service instance sharing resource")
}

func (r *serviceInstanceSharingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceInstanceSharingType

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var spaces []string
	tempDiags := state.Spaces.ElementsAs(ctx, &spaces, false)

	if tempDiags.HasError() {
		resp.Diagnostics.Append(tempDiags...)
		return
	}

	serviceInstanceID := state.Id.ValueString()

	if serviceInstanceID == "" {
		serviceInstanceID = state.ServiceInstance.ValueString()
	}

	err := r.cfClient.ServiceInstances.UnShareWithSpaces(ctx, serviceInstanceID, spaces)

	if err != nil {
		resp.Diagnostics.AddError("Error unsharing service instance with spaces", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted a service instance sharing resource")
}

func mapSharedSpacesValuesToType(relationship *cfv3resource.ServiceInstanceSharedSpaceRelationships, serviceInstance string) ServiceInstanceSharingType {
	sharedSpaces := make([]attr.Value, len(relationship.Data))
	for i, rel := range relationship.Data {
		sharedSpaces[i] = types.StringValue(rel.GUID)
	}
	s := types.SetValueMust(types.StringType, sharedSpaces)
	return ServiceInstanceSharingType{
		Id:              types.StringValue(serviceInstance),
		ServiceInstance: types.StringValue(serviceInstance),
		Spaces:          s,
	}
}

func (r *serviceInstanceSharingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("service_instance_guid"), req, resp)
}
