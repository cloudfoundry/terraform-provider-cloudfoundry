package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type servicePlanVisibilityResource struct {
	cfClient *cfv3client.Client
}

var (
	_ resource.ResourceWithConfigure = &servicePlanVisibilityResource{}
)

func NewServicePlanVisibilityResource() resource.Resource {
	return &servicePlanVisibilityResource{}
}

func (r *servicePlanVisibilityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_plan_visibility"
}

func (r *servicePlanVisibilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This is not a traditional Cloud Foundry resource with a unique GUID but rather a configuration that controls service plan visibility. When type is set to organization and deletion is triggered, Terraform removes only the specified organizations. If no organizations are present, no action is taken.`,

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Denotes the visibility of the plan; can be public, admin, organization, space.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "admin", "organization", "space"),
				},
			},
			"organizations": schema.SetAttribute{
				MarkdownDescription: "Set of organization GUIDs whose members can access the plan; present if type is organization.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the space whose members can access the plan; present if type is space.",
				Computed:            true,
			},
			"service_plan": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service plan.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
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

func (r *servicePlanVisibilityResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config servicePlanVisibilityType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Type.ValueString() != "organization" && !config.Organizations.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("organizations"),
			"invalid attribute combination",
			"organizations can only be set when type is organization",
		)
		return
	}
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

	servicePlanGUID := plan.ServicePlanGUID.ValueString()

	_, err := r.cfClient.ServicePlansVisibility.Apply(ctx, servicePlanGUID, createServicePlanVisibility)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Service Plan Visibility",
			"Could not create service plan visibility: "+err.Error(),
		)
		return
	}

	visibility, err := r.cfClient.ServicePlansVisibility.Get(ctx, plan.ServicePlanGUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Reading Service Plan Visibility",
			"Could not Read service plan visibility: "+err.Error(),
		)
		return
	}

	data, diags := mapServicePlanVisibilityValuesToType(ctx, visibility, plan)
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

	state, diags := mapServicePlanVisibilityValuesToType(ctx, visibility, data)
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

	updateServicePlanVisibility, diags := mapCreateServicePlanVisibilityTypeToValues(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.cfClient.ServicePlansVisibility.Apply(ctx, plan.ServicePlanGUID.ValueString(), updateServicePlanVisibility)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Service Plan Visibility",
			"Could not update service plan visibility: "+err.Error(),
		)
		return
	}

	removedOrgs, _, diags := findChangedRelationsFromTFState(ctx, plan.Organizations, previousState.Organizations)
	resp.Diagnostics.Append(diags...)
	if plan.Type.ValueString() == "organization" {
		for _, orgGUID := range removedOrgs {
			err := r.cfClient.ServicePlansVisibility.Delete(ctx, plan.ServicePlanGUID.ValueString(), orgGUID)
			if err != nil {
				resp.Diagnostics.AddError(
					"API Error removing organization",
					"Could not remove organizations from Service Plan Visibility: "+err.Error(),
				)
				return
			}
		}
	}

	planVisibility, err := r.cfClient.ServicePlansVisibility.Get(ctx, plan.ServicePlanGUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Reading Service Plan Visibility",
			"Could not Read service plan visibility: "+err.Error(),
		)
		return
	}

	state, diags := mapServicePlanVisibilityValuesToType(ctx, planVisibility, plan)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *servicePlanVisibilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state servicePlanVisibilityType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Type.ValueString() == "organization" {
		var orgGUIDs []string
		diags := state.Organizations.ElementsAs(ctx, &orgGUIDs, false)
		resp.Diagnostics.Append(diags...)
		for _, orgGUID := range orgGUIDs {
			err := r.cfClient.ServicePlansVisibility.Delete(ctx, state.ServicePlanGUID.ValueString(), orgGUID)
			if err != nil {
				resp.Diagnostics.AddError(
					"API Error Removing organization",
					"Could not remove organizations from Service Plan Visibility: "+err.Error(),
				)
				return
			}
		}
	}
}
func (rs *servicePlanVisibilityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_plan"), req, resp)
}
