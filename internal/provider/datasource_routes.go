package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &RoutesDataSource{}
	_ datasource.DataSourceWithConfigure = &RoutesDataSource{}
)

// Instantiates a route group data source.
func NewRoutesDataSource() datasource.DataSource {
	return &RoutesDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type RoutesDataSource struct {
	cfClient *client.Client
}

func (d *RoutesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routes"
}

func (d *RoutesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoutesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry route.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The space guid associated to the route.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain guid associated to the route.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The org guid associated to the route to lookup.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The hostname associated to the route to lookup.",
				Optional:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path associated to the route to lookup.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port associated to the route to lookup.",
				Optional:            true,
			},
			"routes": schema.ListNestedAttribute{
				MarkdownDescription: "The list of routes.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						idKey: guidSchema(),
						"space": schema.StringAttribute{
							MarkdownDescription: "The space guid associated to the route.",
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: "The domain guid associated to the route.",
							Computed:            true,
						},
						"host": schema.StringAttribute{
							MarkdownDescription: "The hostname associated to the route to lookup.",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The path associated to the route to lookup.",
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "The port associated to the route to lookup.",
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "The protocol supported by the route, based on the route's domain configuration.",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "The URL for the route.",
							Computed:            true,
						},
						"destinations": schema.SetNestedAttribute{
							MarkdownDescription: "A destination represents the relationship between a route and a resource that can serve traffic.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									idKey: guidSchema(),
									"app_id": schema.StringAttribute{
										MarkdownDescription: "The GUID of the app to route traffic to.",
										Computed:            true,
									},
									"app_process_type": schema.StringAttribute{
										MarkdownDescription: "Type of the process belonging to the app to route traffic to.",
										Computed:            true,
									},
									"port": schema.Int64Attribute{
										MarkdownDescription: "Port on the destination process to route traffic to.",
										Computed:            true,
									},
									"weight": schema.Int64Attribute{
										MarkdownDescription: "Percentage of traffic which will be routed to this destination.",
										Computed:            true,
									},
									"protocol": schema.StringAttribute{
										MarkdownDescription: "Protocol to use for this destination.",
										Computed:            true,
									},
								},
							},
						},
						labelsKey:      datasourceLabelsSchema(),
						annotationsKey: datasourceAnnotationsSchema(),
						createdAtKey:   createdAtSchema(),
						updatedAtKey:   updatedAtSchema(),
					},
				},
			},
		},
	}
}

func (d *RoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasourceRouteType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readOptions := data.mapReadRouteTypeToValues()
	routes, err := d.cfClient.Routes.ListAll(ctx, &readOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Route",
			"Could not get routes : "+err.Error(),
		)
		return
	}

	data, diags = mapRoutesValuesToType(ctx, data, routes)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a route data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
