package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &UsersDataSource{}
	_ datasource.DataSourceWithConfigure = &UsersDataSource{}
)

// Instantiates a space/org users data source.
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type UsersDataSource struct {
	cfClient *client.Client
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all users with a role in the specified space or organization",

		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The guid of the space",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("space"),
						path.MatchRoot("org"),
					}...),
				},
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The guid of the organization",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "The list of users containing a role in the specified space or organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							MarkdownDescription: "The name registered in UAA; will be null for UAA clients and non-UAA users",
							Computed:            true,
						},
						"presentation_name": schema.StringAttribute{
							MarkdownDescription: "The name displayed for the user; for UAA users, this is the same as the username. For UAA clients, this is the UAA client ID",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							MarkdownDescription: "The identity provider for the UAA user; will be null for UAA clients",
							Computed:            true,
						},
						idKey:          guidSchema(),
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spaceOrgUsersType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		users []*resource.User
		err   error
	)
	if !data.Organization.IsNull() {
		users, err = d.cfClient.Organizations.ListUsersAll(ctx, data.Organization.ValueString(), nil)
	} else {
		users, err = d.cfClient.Spaces.ListUsersAll(ctx, data.Space.ValueString(), nil)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Users",
			"Could not get users under the space/org : "+err.Error(),
		)
		return
	}

	data, diags = mapSpaceOrgUsersValuesToType(ctx, data, users)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a users data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
