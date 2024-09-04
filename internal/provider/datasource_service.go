package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &ServiceDataSource{}
var _ datasource.DataSourceWithConfigure = &ServiceDataSource{}

func NewServiceDataSource() datasource.DataSource {
	return &ServiceDataSource{}
}

type serviceType struct {
	Name          types.String `tfsdk:"name"`
	ID            types.String `tfsdk:"id"`
	ServiceBroker types.String `tfsdk:"service_broker"`
	ServicePlans  types.Map    `tfsdk:"service_plans"`
}

type ServiceDataSource struct {
	cfClient *cfv3client.Client
}

func (d *ServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *ServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Service Offering and its related plans",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Service Offering",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "GUID of the service offering",
				Computed:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"service_broker": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service broker which offers the service. Use this to filter two equally named services from different brokers.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"service_plans": schema.MapAttribute{
				MarkdownDescription: "Map of service plan GUIDs keyed by service plan name",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *ServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceType

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	svcOfferingOpts := &cfv3client.ServiceOfferingListOptions{
		Names: cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		},
	}
	// if service broker id is passed
	if !data.ServiceBroker.IsNull() {
		svcOfferingOpts.ServiceBrokerGUIDs = cfv3client.Filter{
			Values: []string{
				data.ServiceBroker.ValueString(),
			},
		}
	}

	serviceOfferings, err := d.cfClient.ServiceOfferings.ListAll(ctx, svcOfferingOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching service offering",
			"Could not get service offering name "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}
	serviceOffering, found := lo.Find(serviceOfferings, func(serviceOffering *resource.ServiceOffering) bool {
		return serviceOffering.Name == data.Name.ValueString()
	})
	if !found {
		resp.Diagnostics.AddError(
			"Unable to find service offering in list",
			fmt.Sprintf("Given name %s not in the list of service offerings", data.Name.ValueString()),
		)
		return
	}
	servicePlans, err := d.cfClient.ServicePlans.ListAll(ctx, &cfv3client.ServicePlanListOptions{
		ServiceOfferingGUIDs: cfv3client.Filter{
			Values: []string{
				serviceOffering.GUID,
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching service plans",
			"Could not get service plans for the service offering "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}
	planMap := make(map[string]string)
	for _, servicePlan := range servicePlans {
		planMap[servicePlan.Name] = servicePlan.GUID
	}

	data = serviceType{
		Name:          types.StringValue(serviceOffering.Name),
		ID:            types.StringValue(serviceOffering.GUID),
		ServiceBroker: types.StringValue(serviceOffering.BrokerCatalog.ID),
	}
	data.ServicePlans, diags = types.MapValueFrom(ctx, types.StringType, planMap)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
