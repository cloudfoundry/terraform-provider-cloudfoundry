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

var _ list.ListResourceWithConfigure = &orgListResource{}

type orgListResource struct {
	cfClient *cfv3client.Client
}

func NewOrgListResource() list.ListResource {
	return &orgListResource{}
}

func (r *orgListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org" // must match managed resource
}

func (r *orgListResource) Configure(_ context.Context,
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

func (r *orgListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all organizations the user has access to.",
	}
}

// List streams all organizations from the API.
func (r *orgListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	orgs, err := r.cfClient.Organizations.ListAll(ctx, cfv3client.NewOrganizationListOptions())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Organizations",
			"Could not list organizations: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, org := range orgs {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("org_guid"), org.GUID)

			if req.IncludeResource {
				resOrg, diags := mapOrgValuesToType(ctx, org)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resOrg)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
