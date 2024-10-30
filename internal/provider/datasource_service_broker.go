package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewServiceBrokerDataSource() datasource.DataSource {
	return &ServiceBrokerDataSource{}
}

type ServiceBrokerDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceBrokerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_broker"
}

func (d *ServiceBrokerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information of a Service broker.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service broker",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the service broker",
				Computed:            true,
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space the service broker is restricted to; omitted for globally available service brokers",
				Computed:            true,
			},
			idKey:          guidSchema(),
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (d *ServiceBrokerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceBrokerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceServiceBrokerType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewServiceBrokerListOptions()
	getOptions.Names = cfv3client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}

	svcBrokers, err := d.cfClient.ServiceBrokers.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in fetching service broker data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}
	if len(svcBrokers) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find service broker in list",
			fmt.Sprintf("Given name %s not in the list of service brokers.", data.Name.ValueString()),
		)
		return
	}

	dataValue, diags := mapServiceBrokerValuesToType(ctx, svcBrokers[0])
	resp.Diagnostics.Append(diags...)
	data = dataValue.Reduce()
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
