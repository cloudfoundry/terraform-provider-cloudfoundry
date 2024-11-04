package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &DomainsDataSource{}
	_ datasource.DataSourceWithConfigure = &DomainsDataSource{}
)

// Instantiates a security group data source.
func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type DomainsDataSource struct {
	cfClient *cfv3client.Client
}

func (d *DomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *DomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information of Cloud Foundry domains. If an organization is specified, it will return the private domains associated with that cloud foundry organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The ID of the Org within which to find the domains",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"domains": schema.ListNestedAttribute{
				MarkdownDescription: "The domains the organization is scoped to",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						idKey: guidSchema(),
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the cloud foundry domain.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var (
		data  domainsDatasourceType
		diags diag.Diagnostics
	)

	diags = req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dlo := cfv3client.NewDomainListOptions()
	if len(data.Org.ValueString()) > 0 {

		_, err := d.cfClient.Organizations.Get(ctx, data.Org.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Organization",
				"Could not get details of the Organization with ID "+data.Org.ValueString()+" : "+err.Error(),
			)
			return
		}

		dlo.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{
				data.Org.ValueString(),
			},
		}
	}

	domains, err := d.cfClient.Domains.ListAll(ctx, dlo)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Domains",
			"Could not get domains for organization "+data.Org.ValueString()+" : "+err.Error(),
		)
		return
	}

	if len(domains) == 0 {
		resp.Diagnostics.AddError(
			"Unable to find any domain in the list",
			fmt.Sprintf("No domain present under org %s with mentioned criteria", data.Org.ValueString()),
		)
		return
	}

	data.Domains, diags = mapDomainsValuesToType(ctx, domains)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read the domains data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
