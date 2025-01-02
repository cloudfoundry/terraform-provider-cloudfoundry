package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider defined types fully satisfy framework interfaces.

var (
	_ datasource.DataSource              = &SpacesDataSource{}
	_ datasource.DataSourceWithConfigure = &SpacesDataSource{}
)

// Instantiates a space data source.
func NewSpacesDataSource() datasource.DataSource {
	return &SpacesDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SpacesDataSource struct {
	cfClient *client.Client
}

func (d *SpacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spaces"
}

func (d *SpacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all spaces under an organization.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the space to look up",
				Optional:            true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization under which the spaces exist",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"spaces": schema.ListNestedAttribute{
				MarkdownDescription: "The spaces present under the organization",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "name of the space",
							Computed:            true,
						},
						"org": schema.StringAttribute{
							MarkdownDescription: "The GUID of the organization the space is contained in",
							Computed:            true,
						},
						"quota": schema.StringAttribute{
							MarkdownDescription: "The space quota applied to the space",
							Computed:            true,
						},
						"allow_ssh": schema.BoolAttribute{
							MarkdownDescription: "Allows SSH to application containers via the CF CLI.",
							Computed:            true,
						},
						"isolation_segment": schema.StringAttribute{
							MarkdownDescription: "The ID of the isolation segment assigned to the space.",
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

func (d *SpacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SpacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spacesType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := d.cfClient.Organizations.Get(ctx, data.OrgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Organization",
			"Could not get details of the Organization with ID "+data.OrgId.ValueString()+" : "+err.Error(),
		)
		return
	}

	spacesListOptions := client.NewSpaceListOptions()
	spacesListOptions.OrganizationGUIDs = client.Filter{
		Values: []string{
			data.OrgId.ValueString(),
		},
	}

	if !data.Name.IsNull() {
		spacesListOptions.Names = client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}

	//Filtering for spaces under the org with GUID
	spaces, err := d.cfClient.Spaces.ListAll(ctx, spacesListOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Spaces",
			"Could not get spaces under Organization with ID "+data.OrgId.ValueString()+" : "+err.Error(),
		)
		return
	}

	if len(spaces) == 0 {

		data.Spaces = []spaceType{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	data.Spaces = []spaceType{}
	for _, space := range spaces {
		sshEnabled, err := d.cfClient.SpaceFeatures.IsSSHEnabled(ctx, space.GUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Space Features",
				"Could not get space features for space "+space.Name+" : "+err.Error(),
			)
			return
		}

		isolationSegment, err := d.cfClient.Spaces.GetAssignedIsolationSegment(ctx, space.GUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Isolation Segment",
				"Could not get isolation segment details for space "+space.Name+" : "+err.Error(),
			)
			return
		}

		spaceValue, diags := mapSpaceValuesToType(ctx, space, sshEnabled, isolationSegment)
		resp.Diagnostics.Append(diags...)
		data.Spaces = append(data.Spaces, spaceValue)
	}

	tflog.Trace(ctx, "read a spaces data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
