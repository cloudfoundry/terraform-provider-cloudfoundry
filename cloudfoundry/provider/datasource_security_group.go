package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &SecurityGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &SecurityGroupDataSource{}
)

// Instantiates a security group data source.
func NewSecurityGroupDataSource() datasource.DataSource {
	return &SecurityGroupDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SecurityGroupDataSource struct {
	cfClient *client.Client
}

func (d *SecurityGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (d *SecurityGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry application security group.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the security group",
				Required:            true,
			},
			"rules": schema.ListNestedAttribute{
				MarkdownDescription: "Rules that will be applied by this security group",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol type",
							Computed:            true,
						},
						"destination": schema.StringAttribute{
							MarkdownDescription: "Destinations that the rule applies to",
							Computed:            true,
						},
						"ports": schema.StringAttribute{
							MarkdownDescription: "Ports that the rule applies to",
							Computed:            true,
						},
						"type": schema.Int64Attribute{
							MarkdownDescription: "Type number for ICMP protocol",
							Computed:            true,
						},
						"code": schema.Int64Attribute{
							MarkdownDescription: "Code field of the ICMP type",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "A description for the rule",
							Computed:            true,
						},
						"log": schema.BoolAttribute{
							MarkdownDescription: "Whether logging for the rule is enabled",
							Computed:            true,
						},
					},
				},
			},
			"globally_enabled_running": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the group should be applied globally to all running applications",
				Computed:            true,
			},
			"globally_enabled_staging": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the group should be applied globally to all staging applications",
				Computed:            true,
			},
			"running_spaces": schema.SetAttribute{
				MarkdownDescription: "The spaces where the security_group is applied to applications during runtime",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"staging_spaces": schema.SetAttribute{
				MarkdownDescription: "The spaces where the security_group is applied to applications during stagingo",
				Computed:            true,
				ElementType:         types.StringType,
			},
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (d *SecurityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data securityGroupType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := client.NewSecurityGroupListOptions()
	getOptions.Names = client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}

	securityGroups, err := d.cfClient.SecurityGroups.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Security Group",
			"Could not get security group with name "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	securityGroup, found := lo.Find(securityGroups, func(securityGroup *cfv3resource.SecurityGroup) bool {
		return securityGroup.Name == data.Name.ValueString()
	})
	if !found {
		resp.Diagnostics.AddError(
			"Unable to find security group in list",
			fmt.Sprintf("Given name %s not in the list of security groups.", data.Name.ValueString()),
		)
		return
	}

	data, diags = mapSecurityGroupValuesToType(ctx, securityGroup)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a security group data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
