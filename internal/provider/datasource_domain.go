package provider

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &DomainDataSource{}
	_ datasource.DataSourceWithConfigure = &DomainDataSource{}
)

// Instantiates a security group data source.
func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type DomainDataSource struct {
	cfClient *client.Client
}

func (d *DomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry domain.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the cloud foundry domain",
				Required:            true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The organization the domain is scoped to",
				Computed:            true,
			},
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Whether the domain is used for internal (container-to-container) traffic, or external (user-to-container) traffic",
				Computed:            true,
			},
			"router_group": schema.StringAttribute{
				MarkdownDescription: "The guid of the desired router group to route tcp traffic through.",
				Computed:            true,
			},
			"shared_orgs": schema.SetAttribute{
				MarkdownDescription: "Organizations the domain is shared with",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"supported_protocols": schema.SetAttribute{
				MarkdownDescription: "Available protocols for routes using the domain, currently http and tcp",
				Computed:            true,
				ElementType:         types.StringType,
			},
			labelsKey:      datasourceLabelsSchema(),
			annotationsKey: datasourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data domainType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainListOptions := client.NewDomainListOptions()
	domainListOptions.Names = client.Filter{
		Values: []string{
			data.Name.ValueString(),
		},
	}
	domains, err := d.cfClient.Domains.ListAll(ctx, domainListOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Domain",
			"Could not get domain with name "+data.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	domain, found := lo.Find(domains, func(domain *cfv3resource.Domain) bool {
		return domain.Name == data.Name.ValueString()
	})
	if !found {
		resp.Diagnostics.AddError(
			"Unable to find domain in list",
			fmt.Sprintf("Given name %s not in the list of domains.", data.Name.ValueString()),
		)
		return
	}

	data, diags = mapDomainValuesToType(ctx, domain)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a domain data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
