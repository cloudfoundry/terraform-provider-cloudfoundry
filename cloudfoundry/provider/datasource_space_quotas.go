package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &SpaceQuotasDataSource{}
var _ datasource.DataSourceWithConfigure = &SpaceQuotasDataSource{}

func NewSpaceQuotasDataSource() datasource.DataSource {
	return &SpaceQuotasDataSource{}
}

type SpaceQuotasDataSource struct {
	cfClient *cfv3client.Client
}

func (d *SpaceQuotasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_quotas"
}

func (d *SpaceQuotasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches space quotas present in a Cloud Foundry org",

		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The ID of the Org within which to find the space quotas",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the space quota to look up",
				Optional:            true,
			},
			"space_quotas": schema.ListNestedAttribute{
				MarkdownDescription: "The list of space quotas",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the space quota ",
							Computed:            true,
						},
						"allow_paid_service_plans": schema.BoolAttribute{
							MarkdownDescription: "Specifies whether instances of paid service plans can be created.",
							Computed:            true,
						},
						"total_services": schema.Int64Attribute{
							MarkdownDescription: "Total number of service instances allowed in a space",
							Computed:            true,
						},
						"total_service_keys": schema.Int64Attribute{
							MarkdownDescription: "Total number of service keys allowed in a space",
							Computed:            true,
						},
						"total_routes": schema.Int64Attribute{
							MarkdownDescription: "Total number of routes allowed in a space",
							Computed:            true,
						},
						"total_route_ports": schema.Int64Attribute{
							MarkdownDescription: "Total number of ports that are reservable by routes in a space",
							Computed:            true,
						},
						"total_memory": schema.Int64Attribute{
							MarkdownDescription: "Total memory allowed for all the started processes and running tasks in a space",
							Computed:            true,
						},
						"instance_memory": schema.Int64Attribute{
							MarkdownDescription: "Maximum memory for a single process or task",
							Computed:            true,
						},
						"org": schema.StringAttribute{
							MarkdownDescription: "The ID of the Org within which the space quota is present",
							Computed:            true,
						},
						"total_app_instances": schema.Int64Attribute{
							MarkdownDescription: "Total instances of all the started processes allowed in a space",
							Computed:            true,
						},
						"total_app_tasks": schema.Int64Attribute{
							MarkdownDescription: "Maximum number of running tasks in a space",
							Computed:            true,
						},
						"total_app_log_rate_limit": schema.Int64Attribute{
							MarkdownDescription: "Total log rate limit allowed for all the started processes and running tasks in an organization",
							Computed:            true,
						},
						"spaces": schema.SetAttribute{
							MarkdownDescription: "Set of Space GUIDs to which this space quota is assigned.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						idKey:        guidSchema(),
						createdAtKey: createdAtSchema(),
						updatedAtKey: updatedAtSchema(),
					},
				},
			},
		},
	}
}

func (d *SpaceQuotasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SpaceQuotasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var (
		data  spaceQuotasDatasourceType
		diags diag.Diagnostics
	)

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := d.cfClient.Organizations.Get(ctx, data.Org.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Organization",
			"Could not get details of the Organization with ID "+data.Org.ValueString()+" : "+err.Error(),
		)
		return
	}

	sqlo := cfv3client.NewSpaceQuotaListOptions()
	sqlo.OrganizationGUIDs = cfv3client.Filter{
		Values: []string{
			data.Org.ValueString(),
		},
	}

	if !data.Name.IsNull() {
		sqlo.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}
	spacesQuotas, err := d.cfClient.SpaceQuotas.ListAll(ctx, sqlo)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch space quota data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	data.SpaceQuotas, diags = mapSpaceQuotasValuesToType(spacesQuotas)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a space quotas data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
