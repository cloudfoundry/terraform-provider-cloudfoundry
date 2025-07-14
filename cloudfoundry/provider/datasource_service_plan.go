package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewServicePlanDataSource() datasource.DataSource {
	return &ServicePlanDataSource{}
}

type ServicePlanDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServicePlanDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_plan"
}

func (d *ServicePlanDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a Service Plan based on the filters provided",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service plan to look up",
				Required:            true,
			},
			"service_offering_name": schema.StringAttribute{
				MarkdownDescription: "The name of the service offering for whose plan to look up",
				Required:            true,
			},
			"service_broker_name": schema.StringAttribute{
				MarkdownDescription: "The name of the service broker which offers the service. Use this to filter two equally named services from different brokers.",
				Optional:            true,
			},
			"visibility_type": schema.StringAttribute{
				MarkdownDescription: "Denotes the visibility of the plan",
				Computed:            true,
			},
			"available": schema.BoolAttribute{
				MarkdownDescription: "Whether or not the service plan is available",
				Computed:            true,
			},
			"free": schema.BoolAttribute{
				MarkdownDescription: "Whether or not the service plan is free of charge",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the service plan",
				Computed:            true,
			},
			"service_offering_id": schema.StringAttribute{
				MarkdownDescription: "The technical ID of the service offering",
				Computed:            true,
			},
			"maintenance_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Information about the version of this service plan",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						MarkdownDescription: "The current semantic version of the service plan",
						Computed:            true,
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "A textual explanation associated with this version",
						Computed:            true,
					},
				},
			},
			"costs": schema.ListNestedAttribute{
				MarkdownDescription: "The cost of the service plan as obtained from the service broker catalog",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"amount": schema.Float64Attribute{
							MarkdownDescription: "Pricing amount",
							Computed:            true,
						},
						"currency": schema.StringAttribute{
							MarkdownDescription: "Currency code for the pricing amount, e.g. USD, GBP",
							Computed:            true,
						},
						"unit": schema.StringAttribute{
							MarkdownDescription: "Display name for type of cost, e.g. Monthly, Hourly, Request, GB",
							Computed:            true,
						},
					},
				},
			},
			"broker_catalog": schema.SingleNestedAttribute{
				MarkdownDescription: "This object contains information obtained from the service broker catalog",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The identifier that the service broker provided for this service plan",
						Computed:            true,
					},
					"metadata": schema.StringAttribute{
						MarkdownDescription: "Additional information provided by the service broker as specified by OSBAPI",
						Computed:            true,
						CustomType:          jsontypes.NormalizedType{},
					},
					"maximum_polling_duration": schema.Float64Attribute{
						MarkdownDescription: "The maximum number of seconds that Cloud Foundry will wait for an asynchronous service broker operation",
						Computed:            true,
					},
					"plan_updateable": schema.BoolAttribute{
						MarkdownDescription: "Whether the service plan supports upgrade/downgrade for service plans",
						Computed:            true,
					},
					"bindable": schema.BoolAttribute{
						MarkdownDescription: "Specifies whether service instances of the service can be bound to applications",
						Computed:            true,
					},
				},
			},
			"schemas": schema.SingleNestedAttribute{
				MarkdownDescription: "Schema definitions for service instances and service bindings for the service plan",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"service_instance": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"create_parameters": schema.StringAttribute{
								MarkdownDescription: "Schema definition for the input parameters for service instance creation",
								Computed:            true,
								CustomType:          jsontypes.NormalizedType{},
							},
							"update_parameters": schema.StringAttribute{
								MarkdownDescription: "Schema definition for the input parameters for service instance update",
								Computed:            true,
								CustomType:          jsontypes.NormalizedType{},
							},
						},
					},
					"service_binding": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"create_parameters": schema.StringAttribute{
								MarkdownDescription: "Schema definition for the input parameters for service Binding creation",
								Computed:            true,
								CustomType:          jsontypes.NormalizedType{},
							},
						},
					},
				},
			},
			idKey:          guidSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
			labelsKey:      datasourceLabelsSchema(),
			annotationsKey: datasourceAnnotationsSchema(),
		},
	}
}
func (d *ServicePlanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServicePlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceServicePlanType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svcPlanOpts := cfv3client.NewServicePlanListOptions()

	svcPlanOpts.Names = cfv3client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}
	svcPlanOpts.ServiceOfferingNames = cfv3client.Filter{
		Values: []string{
			data.ServiceOfferingName.ValueString(),
		},
	}

	if !data.ServiceBrokerName.IsNull() {
		svcPlanOpts.ServiceBrokerNames = cfv3client.Filter{
			Values: []string{
				data.ServiceBrokerName.ValueString(),
			},
		}
	}

	svcPlans, err := d.cfClient.ServicePlans.ListAll(ctx, svcPlanOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error fetching service plans.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	if len(svcPlans) != 1 {
		resp.Diagnostics.AddError(
			"API Error fetching service plans.",
			fmt.Sprintf("Expected exactly one service plan, got %d.", len(svcPlans)),
		)
		return
	}
	// Memorize the Service Broker Name to transfer it to the data object later
	serviceBrokerName := data.ServiceBrokerName
	serviceOfferingName := data.ServiceOfferingName
	data, diags = mapServicePlansValueToData(ctx, svcPlans[0])
	// Service Broker Name is not transferred to the data object, so we need to set it manually
	data.ServiceBrokerName = serviceBrokerName
	data.ServiceOfferingName = serviceOfferingName

	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
