package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewServiceInstancesDataSource() datasource.DataSource {
	return &ServiceInstancesDataSource{}
}

type ServiceInstancesDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceInstancesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_instances"
}

func (d *ServiceInstancesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches service instances available under an org.",

		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the org under which the service instances are present",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to filter for",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service instance to look up",
				Optional:            true,
			},
			"service_instances": schema.ListNestedAttribute{
				MarkdownDescription: "The list of service instances",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the service instance",
							Computed:            true,
						},
						"space": schema.StringAttribute{
							MarkdownDescription: "The ID of the space in which the service instance exists",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the service instance. Either managed or user-provided.",
							Computed:            true,
						},
						"service_plan": schema.StringAttribute{
							MarkdownDescription: "The ID of the service plan from which the service instance was created, only shown when type is managed",
							Computed:            true,
						},
						"tags": schema.ListAttribute{
							MarkdownDescription: "List of tags used by apps to identify service instances. They are shown in the app VCAP_SERVICES env.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"syslog_drain_url": schema.StringAttribute{
							MarkdownDescription: "URL to which logs for bound applications will be streamed; only shown when type is user-provided.",
							Computed:            true,
						},
						"route_service_url": schema.StringAttribute{
							MarkdownDescription: "URL to which requests for bound routes will be forwarded; only shown when type is user-provided.",
							Computed:            true,
						},
						"maintenance_info": schema.SingleNestedAttribute{
							MarkdownDescription: "Information about the version of this service instance; only shown when type is managed",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"version": schema.StringAttribute{
									MarkdownDescription: "The version of the service instance",
									Computed:            true,
								},
								"description": schema.StringAttribute{
									MarkdownDescription: "A description of the version of the service instance",
									Computed:            true,
								},
							},
						},
						"upgrade_available": schema.BoolAttribute{
							MarkdownDescription: "Whether or not an upgrade of this service instance is available on the current Service Plan; details are available in the maintenance_info object; Only shown when type is managed",
							Computed:            true,
						},
						"dashboard_url": schema.StringAttribute{
							MarkdownDescription: "The URL to the service instance dashboard (or null if there is none); only shown when type is managed.",
							Computed:            true,
						},
						"last_operation": lastOperationSchema(),
						idKey:            guidSchema(),
						labelsKey:        datasourceLabelsSchema(),
						annotationsKey:   datasourceAnnotationsSchema(),
						createdAtKey:     createdAtSchema(),
						updatedAtKey:     updatedAtSchema(),
					},
				},
			},
		},
	}
}

func (d *ServiceInstancesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceInstancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data datasourceServiceInstancesType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getOptions := cfv3client.NewServiceInstanceListOptions()
	if !data.Org.IsNull() {
		org, err := d.cfClient.Organizations.Get(ctx, data.Org.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Organization",
				"Could not get details of the Organization with ID "+data.Org.ValueString()+" : "+err.Error(),
			)
			return
		}
		getOptions.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{
				org.GUID,
			},
		}
	}
	if !data.Space.IsNull() {
		space, err := d.cfClient.Spaces.Get(ctx, data.Space.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Space",
				"Could not get space with ID "+data.Space.ValueString()+" : "+err.Error(),
			)
			return
		}
		getOptions.SpaceGUIDs = cfv3client.Filter{
			Values: []string{
				space.GUID,
			},
		}
	}

	if !data.Name.IsNull() {
		getOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}

	svcInstances, err := d.cfClient.ServiceInstances.ListAll(ctx, getOptions)

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in fetching service instances data.",
			fmt.Sprintf("Request failed with %s.", err.Error()),
		)
		return
	}

	data.ServiceInstances, diags = mapDataSourceServiceInstancesValuesToType(ctx, svcInstances)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
