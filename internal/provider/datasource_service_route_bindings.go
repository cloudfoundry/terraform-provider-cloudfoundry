package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ServiceRouteBindingsDataSource{}
	_ datasource.DataSourceWithConfigure = &ServiceRouteBindingsDataSource{}
)

func NewServiceRouteBindingsDataSource() datasource.DataSource {
	return &ServiceRouteBindingsDataSource{}
}

type ServiceRouteBindingsDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceRouteBindingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_route_bindings"
}

func (d *ServiceRouteBindingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Service Route Bindings the user has access to.",

		Attributes: map[string]schema.Attribute{
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service instance to filter by",
				Optional:            true,
			},
			"route": schema.StringAttribute{
				MarkdownDescription: "The GUID of the route to filter by",
				Optional:            true,
			},
			"route_bindings": schema.ListNestedAttribute{
				MarkdownDescription: "The list of route bindings for the given service instance.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						idKey:            guidSchema(),
						labelsKey:        resourceLabelsSchema(),
						annotationsKey:   resourceAnnotationsSchema(),
						createdAtKey:     createdAtSchema(),
						updatedAtKey:     updatedAtSchema(),
					},
				},
			},
		},
	}
}

func (d *ServiceRouteBindingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceRouteBindingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceserviceRouteBindingsType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewServiceRouteBindingListOptions()

	if !data.ServiceInstance.IsNull() {
		getOptions.ServiceInstanceGUIDs = cfv3client.Filter{
			Values: []string{
				data.ServiceInstance.ValueString(),
			},
		}
	}

	if !data.Route.IsNull() {
		getOptions.RouteGUIDs = cfv3client.Filter{
			Values: []string{
				data.Route.ValueString(),
			},
		}
	}

	svcRouteBindings, err := d.cfClient.ServiceRouteBindings.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Service Route Bindings.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	data.RouteBindings, diags = mapDataSourceServiceRouteBindingsValuesToType(ctx, svcRouteBindings)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
