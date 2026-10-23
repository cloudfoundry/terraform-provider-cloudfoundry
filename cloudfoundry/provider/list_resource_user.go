package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &userListResource{}

type userListResource struct {
	cfClient *cfv3client.Client
}

type userListResourceFilter struct {
	Org   types.String `tfsdk:"org"`
	Space types.String `tfsdk:"space"`
}

func NewUserListResource() list.ListResource {
	return &userListResource{}
}

func (r *userListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user" // must match managed resource
}

func (r *userListResource) Configure(_ context.Context,
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

func (r *userListResource) ListResourceConfigSchema(
	_ context.Context,
	_ list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = userListResourceConfigSchema("This list resource allows you to discover all users the caller has access to, optionally scoped to an organization or space.")
}

func userListResourceConfigSchema(description string) schema.Schema {
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to filter users by. Returns only users that are members of this organization.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("space")),
				},
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to filter users by. Returns only users that are members of this space.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("org")),
				},
			},
		},
	}
}

// List streams all users from the API.
func (r *userListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
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
				resUser, diags := mapUserCFValuesToResourceType(ctx, user)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resUser)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}

func listUsersFromCAPI(ctx context.Context, cfClient *cfv3client.Client, filter userListResourceFilter) ([]*cfv3resource.User, diag.Diagnostics) {
	opts := cfv3client.NewUserListOptions()

	var (
		users       []*cfv3resource.User
		err         error
		diagnostics diag.Diagnostics
	)

	switch {
	case !filter.Space.IsNull():
		users, err = cfClient.Spaces.ListUsersAll(ctx, filter.Space.ValueString(), opts)
		if err != nil {
			diagnostics.AddError(
				"API Error Fetching Users for Space",
				"Could not list users for space "+filter.Space.ValueString()+": "+err.Error(),
			)
		}
	case !filter.Org.IsNull():
		users, err = cfClient.Organizations.ListUsersAll(ctx, filter.Org.ValueString(), opts)
		if err != nil {
			diagnostics.AddError(
				"API Error Fetching Users for Organization",
				"Could not list users for organization "+filter.Org.ValueString()+": "+err.Error(),
			)
		}
	default:
		users, err = cfClient.Users.ListAll(ctx, opts)
		if err != nil {
			diagnostics.AddError(
				"API Error Fetching Users",
				"Could not list users: "+err.Error(),
			)
		}
	}

	return users, diagnostics
}
