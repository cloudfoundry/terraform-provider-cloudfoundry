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

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &IsolationSegmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &IsolationSegmentsDataSource{}
)

func NewIsolationSegmentsDataSource() datasource.DataSource {
	return &IsolationSegmentsDataSource{}
}

type IsolationSegmentsDataSource struct {
	cfClient *cfv3client.Client
}

func (d *IsolationSegmentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_isolation_segments"
}

func (d *IsolationSegmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on Cloud Foundry Isolation Segments present.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the isolation segment to filter by",
				Optional:            true,
			},
			"isolation_segments": schema.ListNestedAttribute{
				MarkdownDescription: "List of the isolation segments",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						idKey: guidSchema(),
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the isolation segment",
							Computed:            true,
						},
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

func (d *IsolationSegmentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IsolationSegmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data IsolationSegmentsType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewIsolationSegmentOptions()
	if !data.Name.IsNull() {
		getOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
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
			"Unable to find any Isolation Segment in given list",
			"No isolation segment found with given criteria",
		)
		return
	}

	data.IsolationSegments, diags = mapIsolationSegmentsValuesToType(ctx, isolationSegments)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read an isolation segments data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
