package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &OrgQuotaDataSource{}
var _ datasource.DataSourceWithConfigure = &OrgQuotaDataSource{}

func NewOrgQuotaDataSource() datasource.DataSource {
	return &OrgQuotaDataSource{}
}

type OrgQuotaDataSource struct {
	cfClient *cfv3client.Client
}

func (d *OrgQuotaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_quota"
}

func (d *OrgQuotaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry organization quota.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the organization quota to look up",
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
			"total_private_domains": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of private domains allowed to be created within the Org",
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
			"orgs": schema.SetAttribute{
				MarkdownDescription: "Set of Org GUIDs to which this org quota would be assigned.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			idKey:        guidSchema(),
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (d *OrgQuotaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgQuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var orgQuotaType OrgQuotaType

	resp.Diagnostics.Append(req.Config.Get(ctx, &orgQuotaType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewOrganizationQuotaListOptions()
	getOptions.Names = cfv3client.Filter{
		Values: []string{
			orgQuotaType.Name.ValueString(),
		},
	}
	orgsQuotas, err := d.cfClient.OrganizationQuotas.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch org quota data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}
	orgsQuota, found := lo.Find(orgsQuotas, func(orgQuota *cfv3resource.OrganizationQuota) bool {
		return orgQuota.Name == orgQuotaType.Name.ValueString()
	})
	if !found {
		resp.Diagnostics.AddError(
			"Unable to find org quota data in list",
			fmt.Sprintf("Given name %s not in the list of org quotas.", orgQuotaType.Name.ValueString()),
		)
		return
	}
	orgQuotaType, diags := mapOrgQuotaValuesToType(orgsQuota)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &orgQuotaType)...)
}
