package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ServiceRouteBindingDataSource{}
	_ datasource.DataSourceWithConfigure = &ServiceRouteBindingDataSource{}
)

func NewServiceRouteBindingDataSource() datasource.DataSource {
	return &ServiceRouteBindingDataSource{}
}

type ServiceRouteBindingDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceRouteBindingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_route_binding"
}

func (d *ServiceRouteBindingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Service Route Binding.",

		Attributes: map[string]schema.Attribute{
			idKey: schema.StringAttribute{
				MarkdownDescription: "The GUID of the service route binding",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The service instance that the route is bound to",
				Computed:            true,
			},
			"route": schema.StringAttribute{
				MarkdownDescription: "The GUID of the route to be bound",
				Computed:            true,
			},
			"route_service_url": schema.StringAttribute{
				MarkdownDescription: "The URL for the route service.",
				Computed:            true,
			},
			"last_operation": lastOperationSchema(),
			labelsKey:        resourceLabelsSchema(),
			annotationsKey:   resourceAnnotationsSchema(),
			createdAtKey:     createdAtSchema(),
			updatedAtKey:     updatedAtSchema(),
		},
	}
}

func (d *ServiceRouteBindingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceRouteBindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceserviceRouteBindingType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svcRouteBinding, err := d.cfClient.ServiceRouteBindings.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Service Route Binding.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	bindingValue, diags := mapServiceRouteBindingValuesToType(ctx, svcRouteBinding)
	resp.Diagnostics.Append(diags...)
	data = bindingValue.Reduce()
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
