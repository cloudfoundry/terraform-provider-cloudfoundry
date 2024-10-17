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
	_ datasource.DataSource              = &OrgRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &OrgRoleDataSource{}
)

// Instantiates a space role data source.
func NewOrgRoleDataSource() datasource.DataSource {
	return &OrgRoleDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type OrgRoleDataSource struct {
	cfClient *client.Client
}

func (d *OrgRoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_role"
}

func (d *OrgRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgRoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry org role with a given role ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The guid for the role",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Role type; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types)",
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
	}
}

func (d *OrgRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orgRoleDatasourceType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.cfClient.Roles.Get(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Role",
			"Could not get role with ID "+data.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	if role.Relationships.Space.Data != nil {
		resp.Diagnostics.AddError(
			"Invalid Org Role",
			"Encountered a space role. Kindly provide a valid org role ID",
		)
		return
	}

	roleTypeResponse := mapRoleValuesToType(role)
	datasourceRoleTypeResp := roleTypeResponse.ReduceToOrgRoleDataSource()

	tflog.Trace(ctx, "read an org role data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &datasourceRoleTypeResp)...)

}
