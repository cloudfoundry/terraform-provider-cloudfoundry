package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type servicePlanVisibilityResource struct {
	cfClient *cfv3client.Client
}

var (
	_ resource.ResourceWithConfigure = &servicePlanVisibilityResource{}
)

// New function for consistency with serviceBrokerResource
func NewServicePlanVisibilityResource() resource.Resource {
	return &servicePlanVisibilityResource{}
}

func (r *servicePlanVisibilityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_plan_visibility"
}

func (r *servicePlanVisibilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a Cloud Foundry resource for managing service plan visibility`,

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Denotes the visibility of the plan; can be public, admin, organization, space.",
				Required:            true,
			},
			"organizations": schema.ListNestedAttribute{
				MarkdownDescription: "List of organizations whose members can access the plan; present if type is organization.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"guid": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "GUID of the organization.",
						},
					},
				},
			},
			"space_guid": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the space whose members can access the plan; present if type is space.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"id":         guidSchema(),
			"created_at": createdAtSchema(),
			"updated_at": updatedAtSchema(),
		},
	}
}

func (r *servicePlanVisibilityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}
	r.cfClient = session.CFClient
}

func (r *servicePlanVisibilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan servicePlanVisibilityType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createServicePlanVisibility, diags := mapCreateServicePlanVisibilityTypeToValues(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract Service Plan GUID
	servicePlanGUID := plan.ServicePlanGUID.ValueString()

	// Pass the correct pointer to the Apply function
	createdVisibility, err := r.cfClient.ServicePlansVisibility.Apply(ctx, servicePlanGUID, createServicePlanVisibility)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Service Plan Visibility",
			"Could not create service plan visibility: "+err.Error(),
		)
		return
	}

	data, diags := mapServicePlanVisibilityValuesToType(ctx, createdVisibility)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *servicePlanVisibilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data servicePlanVisibilityType
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	visibility, err := r.cfClient.ServicePlansVisibility.Get(ctx, data.ServicePlanGUID.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "service_plan_visibility", data.ServicePlanGUID.ValueString())
		return
	}

	state, diags := mapServicePlanVisibilityValuesToType(ctx, visibility)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *servicePlanVisibilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState servicePlanVisibilityType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Correctly obtain a pointer
	updateServicePlanVisibility, diags := mapCreateServicePlanVisibilityTypeToValues(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pass the pointer directly without taking its address
	updatedVisibility, err := r.cfClient.ServicePlansVisibility.Update(ctx, plan.ServicePlanGUID.ValueString(), updateServicePlanVisibility)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Service Plan Visibility",
			"Could not update service plan visibility: "+err.Error(),
		)
		return
	}

	data, diags := mapServicePlanVisibilityValuesToType(ctx, updatedVisibility)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *servicePlanVisibilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state servicePlanVisibilityType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract Service Plan GUID
	servicePlanGUID := state.ServicePlanGUID.ValueString()

	// Ensure we have at least one organization to delete the visibility for
	if len(state.Organizations) == 0 {
		resp.Diagnostics.AddError(
			"Missing Organization GUID",
			"At least one organization must be specified for deleting service plan visibility.",
		)
		return
	}

	// Extract the first organization GUID (if multiple exist, additional logic may be needed)
	organizationGUID := state.Organizations[0].GUID.ValueString()

	// Pass all required arguments: context, servicePlanGUID, and organizationGUID
	err := r.cfClient.ServicePlansVisibility.Delete(ctx, servicePlanGUID, organizationGUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Service Plan Visibility",
			"Could not delete service plan visibility: "+err.Error(),
		)
	}
}
