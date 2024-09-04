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
	_ datasource.DataSource              = &StackDataSource{}
	_ datasource.DataSourceWithConfigure = &StackDataSource{}
)

// Instantiates a security group data source.
func NewStackDataSource() datasource.DataSource {
	return &StackDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type StackDataSource struct {
	cfClient *client.Client
}

func (d *StackDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (d *StackDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StackDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry stack.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the stack",
				Required:            true,
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
	}
}

func (d *StackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data stackType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stacks, err := d.cfClient.Stacks.ListAll(ctx, &client.StackListOptions{
		Names: client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Stack",
			"Could not get stack with name "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}
	if len(stacks) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find stack in list",
			fmt.Sprintf("Given name %s not in the list of stacks.", data.Name.ValueString()),
		)
		return
	}

	data, diags = mapStackValuesToType(ctx, stacks[0])
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a stack data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
