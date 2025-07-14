package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider defined types fully satisfy framework interfaces.

var (
	_ datasource.DataSource              = &SpaceDataSource{}
	_ datasource.DataSourceWithConfigure = &SpaceDataSource{}
)

// Instantiates a space data source.
func NewSpaceDataSource() datasource.DataSource {
	return &SpaceDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SpaceDataSource struct {
	cfClient *client.Client
}

func (d *SpaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (d *SpaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry space.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the space to look up",
				Required:            true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization under which the space exists",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
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
	}
}

func (d *SpaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SpaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spaceType

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

	splo := client.NewSpaceListOptions()
	splo.OrganizationGUIDs = client.Filter{
		Values: []string{
			data.OrgId.ValueString(),
		},
	}
	splo.Names = client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}
	//Filtering for spaces under the org with GUID
	spaces, err := d.cfClient.Spaces.ListAll(ctx, splo)

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Spaces",
			"Could not get spaces under Organization with ID "+data.OrgId.ValueString()+" : "+err.Error(),
		)
		return
	}

	space, found := lo.Find(spaces, func(space *resource.Space) bool {
		return space.Name == data.Name.ValueString()
	})

	if !found {
		resp.Diagnostics.AddError(
			"Unable to find space data in list",
			fmt.Sprintf("Given name %s not in the list of spaces under Org with ID %s.", data.Name.ValueString(), data.OrgId.ValueString()),
		)
		return
	}

	sshEnabled, err := d.cfClient.SpaceFeatures.IsSSHEnabled(ctx, space.GUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Space Features",
			"Could not get space features for space "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	isolationSegment, err := d.cfClient.Spaces.GetAssignedIsolationSegment(ctx, space.GUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Isolation Segment",
			"Could not get isolation segment details for space "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	data, diags = mapSpaceValuesToType(ctx, space, sshEnabled, isolationSegment)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a space data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
