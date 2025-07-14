package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &IsolationSegmentDataSource{}
	_ datasource.DataSourceWithConfigure = &IsolationSegmentDataSource{}
)

func NewIsolationSegmentDataSource() datasource.DataSource {
	return &IsolationSegmentDataSource{}
}

type IsolationSegmentDataSource struct {
	cfClient *cfv3client.Client
}

func (d *IsolationSegmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_isolation_segment"
}

func (d *IsolationSegmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry Isolation Segment.",

		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the isolation segment",
				Required:            true,
			},
			labelsKey:      datasourceLabelsSchema(),
			annotationsKey: datasourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (d *IsolationSegmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IsolationSegmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data IsolationSegmentType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewIsolationSegmentOptions()
	getOptions.Names = cfv3client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}

	isolationSegments, err := d.cfClient.IsolationSegments.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Isolation Segment.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	if len(isolationSegments) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find any Isolation Segment.",
			fmt.Sprintf("Given name %s not in the list of isolation segments.", data.Name.ValueString()),
		)
		return
	}

	data, diags = mapIsolationSegmentValuesToType(ctx, isolationSegments[0])
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read an isolation segment data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
