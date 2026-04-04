package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &domainListResource{}

type domainListResource struct {
	cfClient *cfv3client.Client
}

type domainListResourceFilter struct {
	Org types.String `tfsdk:"org"`
}

func NewDomainListResource() list.ListResource {
	return &domainListResource{}
}

func (r *domainListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain" // must match managed resource
}

func (r *domainListResource) Configure(_ context.Context,
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
}

func (r *domainListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all domains the user has access to, optionally filtered by organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to filter domains by. If set, only domains scoped to that organization will be returned.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all domains from the API.
func (r *domainListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter domainListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	domainListOptions := cfv3client.NewDomainListOptions()

	if !filter.Org.IsNull() {
		_, err := r.cfClient.Organizations.Get(ctx, filter.Org.ValueString())
		if err != nil {
			var diags diag.Diagnostics
			diags.AddError(
				"API Error Fetching Organization",
				"Could not get organization with ID "+filter.Org.ValueString()+": "+err.Error(),
			)
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}

		domainListOptions.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{filter.Org.ValueString()},
		}
	}

	domains, err := r.cfClient.Domains.ListAll(ctx, domainListOptions)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Domains",
			"Could not list domains: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, domain := range domains {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("domain_guid"), domain.GUID)

			if req.IncludeResource {
				resDomain, diags := mapDomainValuesToType(ctx, domain)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resDomain)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
