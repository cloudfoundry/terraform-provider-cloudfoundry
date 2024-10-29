package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &OrgRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &OrgRolesDataSource{}
)

// Instantiates a space role data source.
func NewOrgRolesDataSource() datasource.DataSource {
	return &OrgRolesDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type OrgRolesDataSource struct {
	cfClient *client.Client
}

func (d *OrgRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_roles"
}

func (d *OrgRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Cloud Foundry roles within an Organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The guid of the org the role is assigned to",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Valid org role type to filter for; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types)",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("organization_auditor", "organization_user", "organization_manager", "organization_billing_manager"),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The guid of the cloudfoundry user the role is assigned to for filtering",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"roles": schema.ListNestedAttribute{
				MarkdownDescription: "The list of org roles",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The guid for the role",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "role type",
							Computed:            true,
						},
						"user": schema.StringAttribute{
							MarkdownDescription: "The guid of the cloudfoundry user the role is assigned to",
							Computed:            true,
						},
						"org": schema.StringAttribute{
							MarkdownDescription: "The guid of the organization the role is assigned to",
							Computed:            true,
						},
						createdAtKey: createdAtSchema(),
						updatedAtKey: updatedAtSchema(),
					},
				},
			},
		},
	}
}

func (d *OrgRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orgRolesDatasourceType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//_, err := d.cfClient.Spaces.Get(ctx, data.Space.ValueString())
	_, err := d.cfClient.Organizations.Get(ctx, data.Org.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Org",
			"Could not get org with ID "+data.Org.ValueString()+" : "+err.Error(),
		)
		return
	}

	orgRolesListOptions := client.NewRoleListOptions()
	orgRolesListOptions.OrganizationGUIDs = client.Filter{
		Values: []string{
			data.Org.ValueString(),
		},
	}

	if !data.Type.IsNull() {
		orgRolesListOptions.Types = client.Filter{
			Values: []string{
				data.Type.ValueString(),
			},
		}
	}
	if !data.User.IsNull() {
		orgRolesListOptions.UserGUIDs = client.Filter{
			Values: []string{
				data.User.ValueString(),
			},
		}
	}

	roles, err := d.cfClient.Roles.ListAll(ctx, orgRolesListOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch org roles data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	if len(roles) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find any roles in list",
			fmt.Sprintf("No roles present under org %s with mentioned criteria", data.Org.ValueString()),
		)
		return
	}

	data.Roles = mapOrgRolesValuesToType(roles)

	tflog.Trace(ctx, "read a org roles data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
