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

var _ datasource.DataSource = &buildpacksDataSource{}
var _ datasource.DataSourceWithConfigure = &buildpacksDataSource{}

func NewBuildpacksDataSource() datasource.DataSource {
	return &buildpacksDataSource{}
}

type buildpacksDataSource struct {
	cfClient *cfv3client.Client
}

func (d *buildpacksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_buildpacks"
}

func (d *buildpacksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *buildpacksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information of Cloud Foundry buildpacks present.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the buildpack to filter by",
				Optional:            true,
			},
			"stack": schema.StringAttribute{
				MarkdownDescription: "The name of the stack to filter by",
				Optional:            true,
			},
			"buildpacks": schema.ListNestedAttribute{
				MarkdownDescription: "The list of buildpacks",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the buildpack",
							Required:            true,
						},
						"stack": schema.StringAttribute{
							MarkdownDescription: "The name of the stack that the buildpack will use",
							Optional:            true,
						},
						"position": schema.Int64Attribute{
							MarkdownDescription: "The order in which the buildpacks are checked during buildpack auto-detection",
							Optional:            true,
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether or not the buildpack can be used for staging",
							Optional:            true,
							Computed:            true,
						},
						"locked": schema.BoolAttribute{
							MarkdownDescription: "Whether or not the buildpack is locked to prevent updating the bits",
							Optional:            true,
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "The state of the buildpack",
							Computed:            true,
						},
						"filename": schema.StringAttribute{
							MarkdownDescription: "The filename of the buildpack",
							Computed:            true,
						},
						labelsKey:      resourceLabelsSchema(),
						annotationsKey: resourceAnnotationsSchema(),
						createdAtKey:   createdAtSchema(),
						updatedAtKey:   updatedAtSchema(),
						idKey:          guidSchema(),
					},
				},
			},
		},
	}
}

func (d *buildpacksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceBuildpacksType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blo := cfv3client.NewBuildpackListOptions()

	if !data.Name.IsNull() {
		blo.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}
	if !data.Stack.IsNull() {
		blo.Stacks = cfv3client.Filter{
			Values: []string{
				data.Stack.ValueString(),
			},
		}
	}

	buildpacks, err := d.cfClient.Buildpacks.ListAll(ctx, blo)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Buildpacks",
			"Response failed with %s"+err.Error(),
		)
		return
	}

	data.Buildpacks, diags = mapBuildpacksValuesToType(ctx, buildpacks)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read the buildpacks data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
