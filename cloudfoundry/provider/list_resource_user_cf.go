package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ list.ListResourceWithConfigure = &userCFListResource{}

type userCFListResource struct {
	cfClient *cfv3client.Client
}

func NewCFUserListResource() list.ListResource {
	return &userCFListResource{}
}

func (r *userCFListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_cf"
}

func (r *userCFListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userCFListResource) ListResourceConfigSchema(
	_ context.Context,
	_ list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = userListResourceConfigSchema("This list resource allows you to discover all Cloud Controller users the caller has access to, optionally scoped to an organization or space.")
}

func (r *userCFListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var filter userListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	users, diags := listUsersFromCAPI(ctx, r.cfClient, filter)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, user := range users {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("user_guid"), user.GUID)

			if req.IncludeResource {
				resourceUser, diags := mapUserValuesToType(ctx, user)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resourceUser)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
