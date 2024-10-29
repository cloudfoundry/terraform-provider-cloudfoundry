package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &OrgsDataSource{}
var _ datasource.DataSourceWithConfigure = &OrgsDataSource{}

func NewOrgsDataSource() datasource.DataSource {
	return &OrgsDataSource{}
}

type OrgsDataSource struct {
	cfClient *cfv3client.Client
}

func (d *OrgsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_orgs"
}

func (d *OrgsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve all organizations the user has access to.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the organization to look up",
				Optional:            true,
			},
			"orgs": schema.ListNestedAttribute{
				MarkdownDescription: "The organizations available",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						idKey:          guidSchema(),
						labelsKey:      datasourceLabelsSchema(),
						annotationsKey: datasourceAnnotationsSchema(),
						createdAtKey:   createdAtSchema(),
						updatedAtKey:   updatedAtSchema(),
						"name": schema.StringAttribute{
							MarkdownDescription: "name of the organization",
							Computed:            true,
						},
						"suspended": schema.BoolAttribute{
							MarkdownDescription: "Whether an organization is suspended or not",
							Computed:            true,
						},
						"quota": schema.StringAttribute{
							MarkdownDescription: "The ID of quota to be applied to this Org. Default quota is assigned to the org by default.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *OrgsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orgsType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgsListOptions := cfv3client.NewOrganizationListOptions()
	if !data.Name.IsNull() {
		orgsListOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}
	orgs, err := d.cfClient.Organizations.ListAll(ctx, orgsListOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch orgs data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	if len(orgs) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find any org in list",
			"No orgs present with mentioned criteria",
		)
		return
	}

	data.Orgs, diags = mapOrgsValuesToType(ctx, orgs)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read an orgs data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
