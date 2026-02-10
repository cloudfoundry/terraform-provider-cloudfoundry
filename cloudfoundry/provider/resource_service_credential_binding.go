package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type serviceCredentialBindingResource struct {
	cfClient *cfv3client.Client
}

var (
	_ resource.ResourceWithConfigure      = &serviceCredentialBindingResource{}
	_ resource.ResourceWithImportState    = &serviceCredentialBindingResource{}
	_ resource.ResourceWithValidateConfig = &serviceCredentialBindingResource{}
)

const (
	appServiceCredentialBinding = "app"
	keyServiceCredentialBinding = "key"
)

func NewServiceCredentialBindingResource() resource.Resource {
	return &serviceCredentialBindingResource{}
}

func (r *serviceCredentialBindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_credential_binding"
}

func (r *serviceCredentialBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Service credential bindings are used to make the details of the connection to a service instance available to an app or a developer.
		Service credential bindings can be of type app or key.
		A service credential binding is of type app when it is a binding between a service instance and an application Not all services support this binding, as some services deliver value to users directly without integration with an application. Field broker_catalog.features.bindable from service plan of the service instance can be used to determine if it is bindable.
		A service credential binding is of type key when it only retrieves the details of the service instance and makes them available to the developer.`,

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the service credential binding. Either app or key.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("app", "key"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service instance to be bound",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service credential binding. name is optional when the type is app",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				MarkdownDescription: "The GUID of the app to be bound. Required when type is app",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "A JSON object that is passed to the service broker for managed service instance.",
				Optional:            true,
				Sensitive:           true,
				CustomType:          jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credential_binding": schema.StringAttribute{
				MarkdownDescription: "The service credential binding details.",
				Computed:            true,
				Sensitive:           true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"last_operation": lastOperationSchema(),
			idKey:            guidSchema(),
			labelsKey:        resourceLabelsSchema(),
			annotationsKey:   resourceAnnotationsSchema(),
			createdAtKey:     createdAtSchema(),
			updatedAtKey:     updatedAtSchema(),
		},
	}
}

func (r *serviceCredentialBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceCredentialBindingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config serviceCredentialBindingTypeWithCredentials
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Type.ValueString() == keyServiceCredentialBinding && config.Name.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Missing attribute name",
			"Name is required for a service key",
		)
		return
	}

	if config.Type.ValueString() == appServiceCredentialBinding && config.App.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("app"),
			"Missing attribute app",
			"App GUID is required for app binding",
		)
		return

	}

	if config.Type.ValueString() == keyServiceCredentialBinding && !config.App.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("app"),
			"Invalid attribute combination",
			"App GUID should only be provided for credential bindings of type app",
		)
		return

	}
}

func (r *serviceCredentialBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan                     serviceCredentialBindingTypeWithCredentials
		serviceCredentialBinding *cfv3resource.ServiceCredentialBinding
		err                      error
	)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createServiceCredentialBinding, diags := plan.mapCreateServiceCredentialBindingTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, serviceCredentialBinding, err := r.cfClient.ServiceCredentialBindings.Create(ctx, &createServiceCredentialBinding)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in creating service Credential Binding",
			"Unable to create Credential Binding "+plan.Name.ValueString()+": "+err.Error(),
		)
		return

	} else if jobID != "" {
		err = pollJob(ctx, *r.cfClient, jobID, defaultTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to verify service credential binding creation",
				"Service Credential Binding verification failed for "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}

		getOptions := cfv3client.ServiceCredentialBindingListOptions{
			ServiceInstanceGUIDs: cfv3client.Filter{
				Values: []string{
					plan.ServiceInstance.ValueString(),
				},
			},
		}

		if plan.Type.ValueString() == "app" {
			getOptions.AppGUIDs = cfv3client.Filter{
				Values: []string{
					plan.App.ValueString(),
				},
			}
		} else {
			getOptions.Names = cfv3client.Filter{
				Values: []string{
					plan.Name.ValueString(),
				},
			}
		}
		serviceCredentialBinding, err = r.cfClient.ServiceCredentialBindings.Single(ctx, &getOptions)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching service credential binding after creation",
				"Unable to fetch created service credential binding "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	data, diags := mapServiceCredentialBindingValuesToType(ctx, serviceCredentialBinding)
	resp.Diagnostics.Append(diags...)

	data.Parameters = plan.Parameters

	credentialDetails, err := r.cfClient.ServiceCredentialBindings.GetDetails(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"API Error Fetching Service Credential Binding Details.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		data.Credentials = jsontypes.NewNormalizedNull()
	} else {
		credentialJSON, _ := json.Marshal(credentialDetails)
		data.Credentials = jsontypes.NewNormalizedValue(string(credentialJSON))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serviceCredentialBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serviceCredentialBindingTypeWithCredentials

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	serviceCredentialBinding, err := r.cfClient.ServiceCredentialBindings.Get(ctx, data.ID.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "service_credential_binding", data.ID.ValueString())
		return
	}

	state, diags := mapServiceCredentialBindingValuesToType(ctx, serviceCredentialBinding)
	resp.Diagnostics.Append(diags...)

	state.Parameters = data.Parameters

	credentialDetails, err := r.cfClient.ServiceCredentialBindings.GetDetails(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"API Error Fetching Service Credential Binding Details.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		state.Credentials = jsontypes.NewNormalizedNull()
	} else {
		credentialJSON, _ := json.Marshal(credentialDetails)
		state.Credentials = jsontypes.NewNormalizedValue(string(credentialJSON))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *serviceCredentialBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState serviceCredentialBindingTypeWithCredentials
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)

	updateServiceCredentialBinding, diags := plan.mapUpdateServiceCredentialBindingTypeToValues(ctx, previousState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceCredentialBinding, err := r.cfClient.ServiceCredentialBindings.Update(ctx, plan.ID.ValueString(), &updateServiceCredentialBinding)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Service Credential Binding",
			"Could not update Service Credential Binding with ID "+plan.ID.ValueString()+" : "+err.Error(),
		)
		return
	}

	data, diags := mapServiceCredentialBindingValuesToType(ctx, serviceCredentialBinding)
	resp.Diagnostics.Append(diags...)

	data.Parameters = plan.Parameters

	credentialDetails, err := r.cfClient.ServiceCredentialBindings.GetDetails(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"API Error Fetching Service Credential Binding Details.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		data.Credentials = jsontypes.NewNormalizedNull()
	} else {
		credentialJSON, _ := json.Marshal(credentialDetails)
		data.Credentials = jsontypes.NewNormalizedValue(string(credentialJSON))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serviceCredentialBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceCredentialBindingTypeWithCredentials
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := r.cfClient.ServiceCredentialBindings.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in deleting service credential binding",
			"Unable to delete credential binding "+state.Name.ValueString()+": "+err.Error(),
		)

	}
	if jobID != "" {
		if err := pollJob(ctx, *r.cfClient, jobID, defaultTimeout); err != nil {
			resp.Diagnostics.AddError(
				"Unable to verify service credential binding deletion",
				"service credential binding deletion verification failed for "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}

}

func (rs *serviceCredentialBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
