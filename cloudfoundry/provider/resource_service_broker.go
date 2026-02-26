package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceBrokerResource struct {
	cfClient *cfv3client.Client
}

var (
	_ resource.ResourceWithConfigure   = &serviceBrokerResource{}
	_ resource.ResourceWithImportState = &serviceBrokerResource{}
	_ resource.ResourceWithIdentity    = &serviceBrokerResource{}
)

func NewServiceBrokerResource() resource.Resource {
	return &serviceBrokerResource{}
}

type serviceBrokerResourceIdentityModel struct {
	ServiceBrokerGUID types.String `tfsdk:"service_broker_guid"`
}

func (r *serviceBrokerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_broker"
}

func (r *serviceBrokerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a Cloud Foundry resource for managing service brokers`,

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service broker",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the service broker",
				Required:            true,
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space the service broker is restricted to; omitted for globally available service brokers",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username with which to authenticate against the service broker.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password with which to authenticate against the service broker.",
				Required:            true,
				Sensitive:           true,
			},
			idKey:          guidSchema(),
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (r *serviceBrokerResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_broker_guid": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *serviceBrokerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceBrokerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceBrokerType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createServiceBroker, diags := plan.mapCreateServiceBrokerTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)

	jobID, err := r.cfClient.ServiceBrokers.Create(ctx, &createServiceBroker)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in creating Service Broker",
			"Unable to create Service Broker "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	err = pollJob(ctx, *r.cfClient, jobID, defaultTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to verify service broker creation",
			"Service Broker verification failed for "+plan.Name.ValueString()+": "+err.Error(),
		)
	}

	getOptions := cfv3client.ServiceBrokerListOptions{
		Names: cfv3client.Filter{
			Values: []string{
				plan.Name.ValueString(),
			},
		},
	}
	serviceBroker, err := r.cfClient.ServiceBrokers.Single(ctx, &getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching service broker after creation",
			"Unable to fetch created service broker "+plan.Name.ValueString()+": "+err.Error(),
		)
	}

	data, diags := mapServiceBrokerValuesToType(ctx, serviceBroker)
	resp.Diagnostics.Append(diags...)
	data.Username = plan.Username
	data.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := serviceBrokerResourceIdentityModel{
		ServiceBrokerGUID: types.StringValue(data.ID.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *serviceBrokerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serviceBrokerType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	serviceBroker, err := r.cfClient.ServiceBrokers.Get(ctx, data.ID.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "service_broker", data.ID.ValueString())
		return
	}

	state, diags := mapServiceBrokerValuesToType(ctx, serviceBroker)
	resp.Diagnostics.Append(diags...)
	state.Username = data.Username
	state.Password = data.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	var identity serviceBrokerResourceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = serviceBrokerResourceIdentityModel{
			ServiceBrokerGUID: types.StringValue(data.ID.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}

}

func (r *serviceBrokerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState serviceBrokerType
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateServiceBroker, diags := plan.mapUpdateServiceBrokerTypeToValues(ctx, previousState)
	resp.Diagnostics.Append(diags...)

	jobID, _, err := r.cfClient.ServiceBrokers.Update(ctx, plan.ID.ValueString(), &updateServiceBroker)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Service Broker",
			"Could not update Service Broker with ID "+plan.ID.ValueString()+" : "+err.Error(),
		)
		return
	}

	if jobID != "" {
		if err := pollJob(ctx, *r.cfClient, jobID, defaultTimeout); err != nil {
			resp.Diagnostics.AddError(
				"Unable to verify service broker update",
				"Service Broker update verification failed for "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	serviceBroker, err := r.cfClient.ServiceBrokers.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching service broker after update",
			"Unable to fetch updated service broker "+plan.Name.ValueString()+": "+err.Error(),
		)
	}

	data, diags := mapServiceBrokerValuesToType(ctx, serviceBroker)
	resp.Diagnostics.Append(diags...)
	data.Username = plan.Username
	data.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// WORKAROUND for OpenTofu compatibility
	// https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/418
	identity := serviceBrokerResourceIdentityModel{
		ServiceBrokerGUID: types.StringValue(previousState.ID.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
	// END WORKAROUND
}

func (r *serviceBrokerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceBrokerType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := r.cfClient.ServiceBrokers.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in deleting service broker",
			"Unable to delete service broker "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}
	if err := pollJob(ctx, *r.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Unable to verify service broker deletion",
			"service broker deletion verification failed for "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

}

func (rs *serviceBrokerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("service_broker_guid"), req, resp)
}
