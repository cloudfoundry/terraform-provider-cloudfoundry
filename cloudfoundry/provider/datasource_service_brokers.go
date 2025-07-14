package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewServiceBrokersDataSource() datasource.DataSource {
	return &ServiceBrokersDataSource{}
}

type ServiceBrokersDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceBrokersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_brokers"
}

func (d *ServiceBrokersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information of Service brokers user has access to.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "GUID of the space to filter by",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service broker to filter by",
				Optional:            true,
			},
			"service_brokers": schema.ListNestedAttribute{
				MarkdownDescription: "List of service brokers",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
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
				},
			},
		},
	}
}

func (d *ServiceBrokersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceBrokersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceServiceBrokersType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewServiceBrokerListOptions()
	if !data.Name.IsNull() {
		getOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}
	if !data.Space.IsNull() {
		getOptions.SpaceGUIDs = cfv3client.Filter{
			Values: []string{
				data.Space.ValueString(),
			},
		}
	}
	svcBrokers, err := d.cfClient.ServiceBrokers.ListAll(ctx, getOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in fetching service broker data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	data.ServiceBrokers, diags = mapDataSourceServiceBrokersValuesToType(ctx, svcBrokers)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
