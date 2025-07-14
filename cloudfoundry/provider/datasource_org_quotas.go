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

var _ datasource.DataSource = &OrgQuotasDataSource{}
var _ datasource.DataSourceWithConfigure = &OrgQuotasDataSource{}

func NewOrgQuotasDataSource() datasource.DataSource {
	return &OrgQuotasDataSource{}
}

type OrgQuotasDataSource struct {
	cfClient *cfv3client.Client
}

func (d *OrgQuotasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_quotas"
}

func (d *OrgQuotasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches org quotas present in a Cloud Foundry org",

		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The ID of the Org within which to find the org quotas",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"org_quotas": schema.ListNestedAttribute{
				MarkdownDescription: "The list of org quotas",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the organization quota",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *OrgQuotasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgQuotasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var (
		data  orgQuotasDatasourceType
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

	oqlo := cfv3client.NewOrganizationQuotaListOptions()
	oqlo.OrganizationGUIDs = cfv3client.Filter{
		Values: []string{
			data.Org.ValueString(),
		},
	}

	//orgQuotas, err := d.cfClient.OrgQuotas.ListAll(ctx, oqlo)
	orgQuotas, err := d.cfClient.OrganizationQuotas.ListAll(ctx, oqlo)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch org quota data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	data.OrgQuotas, diags = mapOrgQuotasValuesToType(orgQuotas)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a org quotas data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
