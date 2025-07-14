package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &IsolationSegmentEntitlementDataSource{}
	_ datasource.DataSourceWithConfigure = &IsolationSegmentEntitlementDataSource{}
)

func NewIsolationSegmentEntitlementDataSource() datasource.DataSource {
	return &IsolationSegmentEntitlementDataSource{}
}

type IsolationSegmentEntitlementDataSource struct {
	cfClient *cfv3client.Client
}

func (d *IsolationSegmentEntitlementDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_isolation_segment_entitlement"
}

func (d *IsolationSegmentEntitlementDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches organizations entitled with a Cloud Foundry Isolation Segment.",

		Attributes: map[string]schema.Attribute{
			"segment": schema.StringAttribute{
				MarkdownDescription: "GUID of the isolation segment",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"orgs": schema.SetAttribute{
				MarkdownDescription: "GUID's of organizations the segment is entitled with.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *IsolationSegmentEntitlementDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IsolationSegmentEntitlementDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data IsolationSegmentEntitlementDataSourceType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := d.cfClient.IsolationSegments.ListOrganizationRelationships(ctx, data.Segment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Entitled Organizations",
			"Error : "+err.Error(),
		)
		return
	}

	data.Orgs, diags = types.SetValueFrom(ctx, types.StringType, orgs)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read an isolation segment entitlement data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
