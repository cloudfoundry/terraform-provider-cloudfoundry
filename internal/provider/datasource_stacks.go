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

var _ datasource.DataSource = &stacksDataSource{}
var _ datasource.DataSourceWithConfigure = &stacksDataSource{}

func NewStacksDataSource() datasource.DataSource {
	return &stacksDataSource{}
}

type stacksDataSource struct {
	cfClient *cfv3client.Client
}

func (d *stacksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stacks"
}

func (d *stacksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *stacksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information of Cloud Foundry stacks present.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the stack to filter by",
				Optional:            true,
			},
			"stacks": schema.ListNestedAttribute{
				MarkdownDescription: "The list of stacks",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						idKey: guidSchema(),
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stack",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the stack",
							Computed:            true,
						},
						"build_rootfs_image": schema.StringAttribute{
							MarkdownDescription: "The name of the stack image associated with staging/building Apps. If a stack does not have unique images, this will be the same as the stack name.",
							Computed:            true,
						},
						"run_rootfs_image": schema.StringAttribute{
							MarkdownDescription: "The name of the stack image associated with running Apps + Tasks. If a stack does not have unique images, this will be the same as the stack name.",
							Computed:            true,
						},
						"default": schema.BoolAttribute{
							MarkdownDescription: "Whether the stack is configured to be the default stack for new applications.",
							Computed:            true,
						},
						createdAtKey:   createdAtSchema(),
						updatedAtKey:   updatedAtSchema(),
						labelsKey:      datasourceLabelsSchema(),
						annotationsKey: datasourceAnnotationsSchema(),
					},
				},
			},
		},
	}
}

func (d *stacksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data stacksDatasourceType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	slo := cfv3client.NewStackListOptions()

	if !data.Name.IsNull() {
		slo.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}

	stacks, err := d.cfClient.Stacks.ListAll(ctx, slo)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching stacks",
			"Response failed with %s"+err.Error(),
		)
		return
	}

	data.Stacks, diags = mapStacksValuesToType(ctx, stacks)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read the stacks data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
