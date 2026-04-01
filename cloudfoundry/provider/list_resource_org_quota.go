package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ list.ListResourceWithConfigure = &orgQuotaListResource{}

type orgQuotaListResource struct {
	cfClient *cfv3client.Client
}

func NewOrgQuotaListResource() list.ListResource {
	return &orgQuotaListResource{}
}

func (r *orgQuotaListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_quota" // must match managed resource
}

func (r *orgQuotaListResource) Configure(_ context.Context,
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

func (r *orgQuotaListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all organization quotas.",
	}
}

// List streams all organization quotas from the API.
func (r *orgQuotaListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	orgQuotas, err := r.cfClient.OrganizationQuotas.ListAll(ctx, cfv3client.NewOrganizationQuotaListOptions())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Organization Quotas",
			"Could not list organization quotas: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, orgQuota := range orgQuotas {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("org_quota_guid"), orgQuota.GUID)

			if req.IncludeResource {
				resOrgQuota, diags := mapOrgQuotaValuesToType(orgQuota)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resOrgQuota)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
