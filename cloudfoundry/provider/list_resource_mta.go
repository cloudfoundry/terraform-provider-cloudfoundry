package provider

import (
	"context"
	"fmt"
	"strings"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/mta"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &mtaListResource{}

type mtaListResource struct {
	mtaClient *mta.APIClient
	cfClient  *cfv3client.Client
}

type mtaListResourceFilter struct {
	Space     types.String `tfsdk:"space"`
	Namespace types.String `tfsdk:"namespace"`
	DeployUrl types.String `tfsdk:"deploy_url"`
}

func NewMtaListResource() list.ListResource {
	return &mtaListResource{}
}

func (r *mtaListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mta" // must match managed resource
}

func (r *mtaListResource) Configure(_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.cfClient = session.CFClient
	apiEndpointURL := session.CFClient.ApiURL("")
	conf := mta.NewConfiguration(apiEndpointURL, session.CFClient.UserAgent(), session.CFClient.HTTPAuthClient())
	r.mtaClient = mta.NewAPIClient(conf)

	subDomainWithProtocol := strings.Split(apiEndpointURL, ".")[0]
	subDomain := strings.Split(subDomainWithProtocol, "//")[1]
	deploySubdomainWithProtocol := strings.Replace(subDomainWithProtocol, subDomain, "deploy-service", 1)
	deployURL := strings.Replace(apiEndpointURL, subDomainWithProtocol, deploySubdomainWithProtocol, 1)
	r.mtaClient.ChangeBasePath(deployURL)
}

func (r *mtaListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all MTAs deployed in a space, optionally filtered by namespace.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to list MTAs in.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace of the MTAs to filter by.",
				Optional:            true,
			},
			"deploy_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the deploy service, if a custom one has been used. By default 'deploy-service.<system-domain>'.",
				Optional:            true,
			},
		},
	}
}

// List streams all MTAs from the API.
func (r *mtaListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter mtaListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	if !filter.DeployUrl.IsNull() {
		r.mtaClient.ChangeBasePath(filter.DeployUrl.ValueString())
	}

	var namespace *string
	if !filter.Namespace.IsNull() {
		namespace = new(filter.Namespace.ValueString())
	}

	mtas, _, err := r.mtaClient.DefaultApi.GetMtas(ctx, filter.Space.ValueString(), namespace, "")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching MTAs",
			"Could not list MTAs in space "+filter.Space.ValueString()+": "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, m := range mtas {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("space_guid"), filter.Space.ValueString())
			result.Identity.SetAttribute(ctx, path.Root("mta_id"), m.Metadata.Id)
			result.Identity.SetAttribute(ctx, path.Root("namespace"), m.Metadata.Namespace)

			if req.IncludeResource {
				resMta, diags := mapMtaToResourceType(ctx, m, filter.Space, filter.Namespace)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resMta)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
