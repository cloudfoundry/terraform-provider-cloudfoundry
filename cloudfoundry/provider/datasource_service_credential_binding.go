package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ServiceCredentialBindingDataSource{}
	_ datasource.DataSourceWithConfigure = &ServiceCredentialBindingDataSource{}
)

func NewServiceCredentialBindingDataSource() datasource.DataSource {
	return &ServiceCredentialBindingDataSource{}
}

type ServiceCredentialBindingDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceCredentialBindingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_credential_binding"
}

func (d *ServiceCredentialBindingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Service Credential Bindings for a given service instance.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service credential binding to query for",
				Optional:            true,
			},
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service instance",
				Required:            true,
			},
			"app": schema.StringAttribute{
				MarkdownDescription: "The GUID of the app which is bound to be query for",
				Optional:            true,
			},
			"credential_bindings": schema.ListNestedAttribute{
				MarkdownDescription: "The list of credential bindings for the given service instance.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the service credential binding",
							Computed:            true,
						},
						"service_instance": schema.StringAttribute{
							MarkdownDescription: "The GUID of the service instance",
							Computed:            true,
						},
						"app": schema.StringAttribute{
							MarkdownDescription: "The GUID of the app which is bound",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the service credential binding",
							Computed:            true,
						},
						"credential_binding": schema.StringAttribute{
							MarkdownDescription: "The service credential binding details.",
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
				},
			},
		},
	}
}

func (d *ServiceCredentialBindingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceCredentialBindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceserviceCredentialBindingType

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

	if !data.Name.IsNull() {
		getOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}
	if !data.App.IsNull() {
		getOptions.AppGUIDs = cfv3client.Filter{
			Values: []string{
				data.App.ValueString(),
			},
		}
	}

	svcCredentialBindings, err := d.cfClient.ServiceCredentialBindings.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Service Credential Binding.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	if len(svcCredentialBindings) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find any credential bindings in list",
			"Given input does not have any binding present",
		)
		return
	}

	bindingValues := []serviceCredentialBindingTypeWithCredentials{}
	for _, svcBinding := range svcCredentialBindings {

		bindingValue, diags := mapServiceCredentialBindingValuesToType(ctx, svcBinding)
		resp.Diagnostics.Append(diags...)
		bindingWithCredentials := bindingValue.Reduce()

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

		bindingValues = append(bindingValues, bindingWithCredentials)
	}

	data.CredentialBindings = bindingValues
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
