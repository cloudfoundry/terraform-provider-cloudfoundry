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

var _ list.ListResourceWithConfigure = &spaceQuotaListResource{}

type spaceQuotaListResource struct {
	cfClient *cfv3client.Client
}

type spaceQuotaListResourceFilter struct {
	Org types.String `tfsdk:"org"`
}

func NewSpaceQuotaListResource() list.ListResource {
	return &spaceQuotaListResource{}
}

func (r *spaceQuotaListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_quota" // must match managed resource
}

func (r *spaceQuotaListResource) Configure(_ context.Context,
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

func (r *spaceQuotaListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all space quotas within an organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to list space quotas for.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all space quotas for an organization from the API.
func (r *spaceQuotaListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter spaceQuotaListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

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

	sqlo := cfv3client.NewSpaceQuotaListOptions()
	sqlo.OrganizationGUIDs = cfv3client.Filter{
		Values: []string{filter.Org.ValueString()},
	}

	spaceQuotas, err := r.cfClient.SpaceQuotas.ListAll(ctx, sqlo)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Space Quotas",
			"Could not list space quotas for organization "+filter.Org.ValueString()+": "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, spaceQuota := range spaceQuotas {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("space_quota_guid"), spaceQuota.GUID)

			if req.IncludeResource {
				resSpaceQuota, diags := mapSpaceQuotaValuesToType(spaceQuota)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resSpaceQuota)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
