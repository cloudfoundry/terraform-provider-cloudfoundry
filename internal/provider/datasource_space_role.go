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
	_ datasource.DataSource              = &SpaceRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &SpaceRoleDataSource{}
)

// Instantiates a space role data source.
func NewSpaceRoleDataSource() datasource.DataSource {
	return &SpaceRoleDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SpaceRoleDataSource struct {
	cfClient *client.Client
}

func (d *SpaceRoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_role"
}

func (d *SpaceRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SpaceRoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry space role with a given role ID.",
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
			"space": schema.StringAttribute{
				MarkdownDescription: "The guid of the space the role is assigned to",
				Computed:            true,
			},
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (d *SpaceRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spaceRoleDatasourceType
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

	if role.Relationships.Org.Data != nil {
		resp.Diagnostics.AddError(
			"Invalid Space Role",
			"Encountered an organisation role. Kindly provide a valid space role ID",
		)
		return
	}

	roleTypeResponse := mapRoleValuesToType(role)
	datasourceRoleTypeResp := roleTypeResponse.ReduceToSpaceRoleDataSource()

	tflog.Trace(ctx, "read a space role data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &datasourceRoleTypeResp)...)

}
