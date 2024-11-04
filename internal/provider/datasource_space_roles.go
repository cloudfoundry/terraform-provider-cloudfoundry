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
	_ datasource.DataSource              = &SpaceRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &SpaceRolesDataSource{}
)

// Instantiates a space role data source.
func NewSpaceRolesDataSource() datasource.DataSource {
	return &SpaceRolesDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SpaceRolesDataSource struct {
	cfClient *client.Client
}

func (d *SpaceRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_roles"
}

func (d *SpaceRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SpaceRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Cloud Foundry roles within a space.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The guid of the space the role is assigned to",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Valid space role type to filter for; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types)",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("space_auditor", "space_developer", "space_manager", "space_supporter"),
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
				MarkdownDescription: "The list of space roles",
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
						"space": schema.StringAttribute{
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

func (d *SpaceRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spaceRolesDatasourceType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := d.cfClient.Spaces.Get(ctx, data.Space.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Space",
			"Could not get space with ID "+data.Space.ValueString()+" : "+err.Error(),
		)
		return
	}

	spaceRolesListOptions := client.NewRoleListOptions()
	spaceRolesListOptions.SpaceGUIDs = client.Filter{
		Values: []string{
			data.Space.ValueString(),
		},
	}

	if !data.Type.IsNull() {
		spaceRolesListOptions.Types = client.Filter{
			Values: []string{
				data.Type.ValueString(),
			},
		}
	}
	if !data.User.IsNull() {
		spaceRolesListOptions.UserGUIDs = client.Filter{
			Values: []string{
				data.User.ValueString(),
			},
		}
	}

	roles, err := d.cfClient.Roles.ListAll(ctx, spaceRolesListOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch space roles data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	if len(roles) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find any roles in list",
			fmt.Sprintf("No roles present under space %s with mentioned criteria", data.Space.ValueString()),
		)
		return
	}

	data.Roles = mapSpaceRolesValuesToType(roles)

	tflog.Trace(ctx, "read a space roles data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
