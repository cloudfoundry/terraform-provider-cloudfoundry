package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/mta"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &MtasDataSource{}
	_ datasource.DataSourceWithConfigure = &MtasDataSource{}
)

// Instantiates a mtar data source.
func NewMtasDataSource() datasource.DataSource {
	return &MtasDataSource{}
}

// Contains reference to the mta client to be used for making the API calls.
type MtasDataSource struct {
	mtaClient *mta.APIClient
}

func (d *MtasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mtas"
}

func (d *MtasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}

	apiEndpointURL := session.CFClient.ApiURL("")
	conf := mta.NewConfiguration(apiEndpointURL, session.CFClient.UserAgent(), session.CFClient.HTTPAuthClient())
	d.mtaClient = mta.NewAPIClient(conf)

	subDomainWithProtocol := strings.Split(apiEndpointURL, ".")[0]
	subDomain := strings.Split(subDomainWithProtocol, "//")[1]
	deploySubdomainWithProtocol := strings.Replace(subDomainWithProtocol, subDomain, "deploy-service", 1)
	deployURL := strings.Replace(apiEndpointURL, subDomainWithProtocol, deploySubdomainWithProtocol, 1)

	d.mtaClient.ChangeBasePath(deployURL)
}

func (d *MtasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Gets information on Multi Target Applications present in a space.
		
		
__Further documentation:__ 
 [Multitarget Applications in the Cloud Foundry Environment](https://help.sap.com/docs/btp/sap-business-technology-platform/multitarget-applications-in-cloud-foundry-environment)
 `,

		Attributes: map[string]schema.Attribute{
			"deploy_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the deploy service, if a custom one has been used(should be present in the same landscape). By default 'deploy-service.<system-domain>'",
				Optional:            true,
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space where the MTA's are deployed",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace of the MTA to filter by",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The MTA ID to filter by",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"mtas": schema.ListNestedAttribute{
				MarkdownDescription: "The list of MTA's",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metadata": schema.SingleNestedAttribute{
							MarkdownDescription: "an identifier, version and namespace that uniquely identify the MTA",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed: true,
								},
								"version": schema.StringAttribute{
									Computed: true,
								},
								"namespace": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"modules": schema.ListNestedAttribute{
							MarkdownDescription: "the deployable parts contained in the MTA deployment archive, most commonly Cloud Foundry applications or content",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"module_name": schema.StringAttribute{
										Computed: true,
									},
									"app_name": schema.StringAttribute{
										Computed: true,
									},
									"created_on": schema.StringAttribute{
										Computed: true,
									},
									"updated_on": schema.StringAttribute{
										Computed: true,
									},
									"provided_dendency_names": schema.ListAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
									"services": schema.ListAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
									"uris": schema.ListAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
								},
							},
						},
						"services": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *MtasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var (
		data      MtasDataSourceType
		namespace *string
	)
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.DeployUrl.IsNull() {
		d.mtaClient.ChangeBasePath(data.DeployUrl.ValueString())
	}

	if !data.Namespace.IsNull() {
		namespace = strtostrptr(data.Namespace.ValueString())
	}

	//get details of MTA
	mtas, _, err := d.mtaClient.DefaultApi.GetMtas(ctx, data.Space.ValueString(), namespace, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch MTA details",
			fmt.Sprintf("Request failed with %s ", err.Error()),
		)
		return
	}
	data, diags = mapMtasValuesToType(ctx, data, mtas)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a mtas datasource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
