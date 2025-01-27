package provider

import (
	"context"
	"fmt"
	"log"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	log.Println("Metadata called")
	resp.TypeName = req.ProviderTypeName + "_service_plan_visibility"
}

func (r *servicePlanVisibilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	log.Println("Schema called")
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a Cloud Foundry resource for managing service plan visibility`,

		Attributes: map[string]schema.Attribute{
			"service_plan_guid": schema.StringAttribute{
				MarkdownDescription: "GUID of the service plan.",
				Required:            true,
			},
			"organization_guid": schema.StringAttribute{
				MarkdownDescription: "GUID of the organization the visibility is restricted to.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			idKey:        guidSchema(),
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (r *servicePlanVisibilityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	log.Println("Configure called")
	if req.ProviderData == nil {
		log.Println("ProviderData is nil")
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		log.Printf("Unexpected Resource Configure Type: %T\n", req.ProviderData)
		return
	}
	r.cfClient = session.CFClient
	log.Println("Configure completed successfully")
}

func (r *servicePlanVisibilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	log.Println("Create called")
	var plan servicePlanVisibilityType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Println("Error getting plan:", resp.Diagnostics)
		return
	}

	createServicePlanVisibility, diags := plan.mapCreateServicePlanVisibilityTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Println("Error mapping create service plan visibility:", resp.Diagnostics)
		return
	}

	createdVisibility, err := r.cfClient.ServicePlansVisibility.Apply(ctx, plan.ServicePlanGUID.ValueString(), &createServicePlanVisibility)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create service plan visibility",
			"Failed to create service plan visibility: "+err.Error(),
		)
		log.Println("Failed to create service plan visibility:", err)
		return
	}

	err = pollJob(ctx, r.cfClient, createdVisibility.GUID, defaultTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to verify service plan visibility creation",
			"Failed to verify service plan visibility creation: "+err.Error(),
		)
		log.Println("Failed to verify service plan visibility creation:", err)
		return
	}

	ServicePlanVisibility, err := r.cfClient.ServicePlansVisibility.Get(ctx, createdVisibility.GUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get service plan visibility",
			"Failed to get service plan visibility: "+err.Error(),
		)
		log.Println("Failed to get service plan visibility:", err)
		return
	}

	data, diags := mapServicePlanVisibilityToType(ctx, ServicePlanVisibility)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	log.Println("Create completed successfully")
}

func (r *servicePlanVisibilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	log.Println("Read called")
	var data servicePlanVisibilityType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Println("Error getting state:", resp.Diagnostics)
		return
	}

	visibility, err := r.cfClient.ServicePlansVisibility.Get(ctx, data.ServicePlanGUID.ValueString())
	if err != nil {
		if cfv3client.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			log.Println("Resource not found, removing state")
			return
		}
		resp.Diagnostics.AddError(
			"Failed to get service plan visibility",
			"Failed to get service plan visibility: "+err.Error(),
		)
		log.Println("Failed to get service plan visibility:", err)
		return
	}

	data.ServicePlanGUID = types.StringValue(visibility.ServicePlanGUID)
	data.OrganizationGUID = types.StringValue(visibility.OrganizationGUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	log.Println("Read completed successfully")
}

func (r *servicePlanVisibilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	log.Println("Update called")
	var plan, previousState servicePlanVisibilityType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		log.Println("Error getting plan or state:", resp.Diagnostics)
		return
	}

	updateServicePlanVisibility, diags := plan.mapServicePlanVisibilityTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Println("Error mapping update service plan visibility:", resp.Diagnostics)
		return
	}

	_, err := r.cfClient.ServicePlansVisibility.Update(ctx, plan.ServicePlanGUID.ValueString(), &updateServicePlanVisibility)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Service Plan Visibility",
			"Could not update Service Plan Visibility with ID "+plan.ServicePlanGUID.ValueString()+" : "+err.Error(),
		)
		log.Println("API Error Updating Service Plan Visibility:", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	log.Println("Update completed successfully")
}

func (r *servicePlanVisibilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	log.Println("Delete called")
	var data servicePlanVisibilityType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Println("Error getting state:", resp.Diagnostics)
		return
	}

	err := r.cfClient.ServicePlansVisibility.Delete(ctx, data.ServicePlanGUID.ValueString())
	if err != nil && !cfv3client.IsResourceNotFoundError(err) {
		resp.Diagnostics.AddError(
			"Failed to delete service plan visibility",
			"Failed to delete service plan visibility: "+err.Error(),
		)
		log.Println("Failed to delete service plan visibility:", err)
		return
	}

	resp.State.RemoveResource(ctx)
	log.Println("Delete completed successfully")
}
