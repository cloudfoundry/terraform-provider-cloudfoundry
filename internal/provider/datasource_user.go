package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &UserDataSource{}
	_ datasource.DataSourceWithConfigure = &UserDataSource{}
)

// Instantiates a user data source.
func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type UserDataSource struct {
	cfClient *client.Client
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Cloud Foundry users with a given username.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user to look up",
				Required:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "The list of users containing the given username.",
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasourceUserType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uslo := client.NewUserListOptions()
	uslo.UserNames = client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}
	users, err := d.cfClient.Users.ListAll(ctx, uslo)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Users",
			"Could not get users with name "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	if len(users) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find user data in list",
			fmt.Sprintf("Given name %s not in the list of users", data.Name.ValueString()),
		)
		return
	}

	data, diags = mapUsersValuesToType(ctx, data, users)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a user data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
