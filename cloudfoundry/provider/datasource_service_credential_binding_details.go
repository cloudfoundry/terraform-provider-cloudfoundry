package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ServiceCredentialBindingDetailsDataSource{}
	_ datasource.DataSourceWithConfigure = &ServiceCredentialBindingDetailsDataSource{}
)

func NewServiceCredentialBindingDetailsDataSource() datasource.DataSource {
	return &ServiceCredentialBindingDetailsDataSource{}
}

type ServiceCredentialBindingDetailsDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceCredentialBindingDetailsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_credential_binding_details"
}

func (d *ServiceCredentialBindingDetailsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Service Credential Binding Details of a particular type (`app` or `key`) for a given service instance.\n\n" +
			"Prefer this over `datasource_service_credential_binding`, which will be deprecated going forward.\n\n" +
			"If you want to fetch all credential bindings (both `app` and `key` types), use `datasource_service_credential_bindings` instead.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service credential binding",
				Optional:            true,
			},
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service instance",
				Required:            true,
			},
			"app": schema.StringAttribute{
				MarkdownDescription: "The GUID of the app which is bound",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the service credential binding",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("app", "key"),
				},
			},
			"credential_binding": schema.StringAttribute{
				MarkdownDescription: "The service credential binding details as JSON.",
				Computed:            true,
				Sensitive:           true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "A JSON object that is passed to the service broker.",
				Computed:            true,
				Sensitive:           true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"last_operation": lastOperationSchema(),
			idKey:            guidSchema(),
			labelsKey:        datasourceLabelsSchema(),
			annotationsKey:   datasourceAnnotationsSchema(),
			createdAtKey:     createdAtSchema(),
			updatedAtKey:     updatedAtSchema(),
		},
	}
}

func (d *ServiceCredentialBindingDetailsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.cfClient = session.CFClient
}

func (d *ServiceCredentialBindingDetailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data serviceCredentialBindingTypeWithCredentials

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewServiceCredentialBindingListOptions()
	getOptions.ServiceInstanceGUIDs = cfv3client.Filter{
		Values: []string{
			data.ServiceInstance.ValueString(),
		},
	}

	switch data.Type.ValueString() {
	case "app":
		if data.App.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("app"),
				"Missing attribute app",
				"App GUID is required for app binding",
			)
			return
		}
		getOptions.AppGUIDs = cfv3client.Filter{
			Values: []string{
				data.App.ValueString(),
			},
		}
	case "key":
		if !data.App.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("app"),
				"Invalid attribute combination",
				"App GUID should only be provided for credential bindings of type app",
			)
			return
		}
		if data.Name.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Missing attribute name",
				"Name is required for a service key",
			)
			return
		}
		getOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}

	svcBinding, err := d.cfClient.ServiceCredentialBindings.Single(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Service Credential Binding.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	bindingWithCredentials, diags := mapServiceCredentialBindingValuesToType(ctx, svcBinding)
	resp.Diagnostics.Append(diags...)

	credentialDetails, err := d.cfClient.ServiceCredentialBindings.GetDetails(ctx, bindingWithCredentials.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"API Error Fetching Service Credential Binding Details.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		bindingWithCredentials.Credentials = jsontypes.NewNormalizedNull()
	} else {
		credentialJSON, _ := json.Marshal(credentialDetails)
		bindingWithCredentials.Credentials = jsontypes.NewNormalizedValue(string(credentialJSON))
	}

	parameterDetails, err := d.cfClient.ServiceCredentialBindings.GetParameters(ctx, bindingWithCredentials.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Unable to fetch Parameters.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		bindingWithCredentials.Parameters = jsontypes.NewNormalizedNull()
	} else {
		credentialJSON, _ := json.Marshal(parameterDetails)
		bindingWithCredentials.Parameters = jsontypes.NewNormalizedValue(string(credentialJSON))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &bindingWithCredentials)...)
}
