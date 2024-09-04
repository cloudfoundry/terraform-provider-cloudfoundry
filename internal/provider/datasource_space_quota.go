package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &SpaceQuotaDataSource{}
var _ datasource.DataSourceWithConfigure = &SpaceQuotaDataSource{}

func NewSpaceQuotaDataSource() datasource.DataSource {
	return &SpaceQuotaDataSource{}
}

type SpaceQuotaDataSource struct {
	cfClient *cfv3client.Client
}

func (d *SpaceQuotaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_quota"
}

func (d *SpaceQuotaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry space quota.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the space quota to look up",
				Required:            true,
			},
			"allow_paid_service_plans": schema.BoolAttribute{
				MarkdownDescription: "Determines whether users can provision instances of non-free service plans. Does not control plan visibility. When false, non-free service plans may be visible in the marketplace but instances can not be provisioned.",
				Computed:            true,
			},
			"total_services": schema.Int64Attribute{
				MarkdownDescription: "Maximum services allowed",
				Computed:            true,
			},
			"total_service_keys": schema.Int64Attribute{
				MarkdownDescription: "Maximum service keys allowed",
				Computed:            true,
			},
			"total_routes": schema.Int64Attribute{
				MarkdownDescription: "Maximum routes allowed",
				Computed:            true,
			},
			"total_route_ports": schema.Int64Attribute{
				MarkdownDescription: "Maximum routes with reserved ports",
				Computed:            true,
			},
			"total_memory": schema.Int64Attribute{
				MarkdownDescription: "Maximum memory usage allowed",
				Computed:            true,
			},
			"instance_memory": schema.Int64Attribute{
				MarkdownDescription: "Maximum memory per application instance",
				Computed:            true,
			},
			"total_app_instances": schema.Int64Attribute{
				MarkdownDescription: "Maximum app instances allowed",
				Computed:            true,
			},
			"total_app_tasks": schema.Int64Attribute{
				MarkdownDescription: "Maximum tasks allowed per app",
				Computed:            true,
			},
			"total_app_log_rate_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum log rate allowed for all the started processes and running tasks in bytes/second.",
				Computed:            true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The ID of the Org within which to find the space quota",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"spaces": schema.SetAttribute{
				MarkdownDescription: "Set of Space GUIDs to which this space quota would be assigned.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			idKey:        guidSchema(),
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (d *SpaceQuotaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SpaceQuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var spaceQuotaType spaceQuotaType

	resp.Diagnostics.Append(req.Config.Get(ctx, &spaceQuotaType)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sqlo := &cfv3client.SpaceQuotaListOptions{
		Names: cfv3client.Filter{
			Values: []string{spaceQuotaType.Name.ValueString()},
		},
	}
	if !spaceQuotaType.Org.IsNull() {
		sqlo.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{spaceQuotaType.Org.ValueString()},
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
	spacesQuota, found := lo.Find(spacesQuotas, func(spaceQuota *cfv3resource.SpaceQuota) bool {
		return spaceQuota.Name == spaceQuotaType.Name.ValueString()
	})
	if !found {
		resp.Diagnostics.AddError(
			"Unable to find space quota data in list",
			fmt.Sprintf("Given name %s not in the list of space quotas.", spaceQuotaType.Name.ValueString()),
		)
		return
	}
	spaceQuotaType, diags := mapSpaceQuotaValuesToType(spacesQuota)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &spaceQuotaType)...)
}
