package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
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

var _ list.ListResourceWithConfigure = &spaceRoleListResource{}

type spaceRoleListResource struct {
	cfClient *cfv3client.Client
}

type spaceRoleListResourceFilter struct {
	Space types.String `tfsdk:"space"`
	Type  types.String `tfsdk:"type"`
	User  types.String `tfsdk:"user"`
}

func NewSpaceRoleListResource() list.ListResource {
	return &spaceRoleListResource{}
}

func (r *spaceRoleListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_role" // must match managed resource
}

func (r *spaceRoleListResource) Configure(_ context.Context,
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

func (r *spaceRoleListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all roles within a space.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to list roles for.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Role type to filter by; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("space_auditor", "space_developer", "space_manager", "space_supporter"),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The GUID of the user to filter roles by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all space roles from the API.
func (r *spaceRoleListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter spaceRoleListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	_, err := r.cfClient.Spaces.Get(ctx, filter.Space.ValueString())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Space",
			"Could not get space with ID "+filter.Space.ValueString()+": "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	roleListOptions := cfv3client.NewRoleListOptions()
	roleListOptions.SpaceGUIDs = cfv3client.Filter{
		Values: []string{filter.Space.ValueString()},
	}

	if !filter.Type.IsNull() {
		roleListOptions.Types = cfv3client.Filter{
			Values: []string{filter.Type.ValueString()},
		}
	}

	if !filter.User.IsNull() {
		roleListOptions.UserGUIDs = cfv3client.Filter{
			Values: []string{filter.User.ValueString()},
		}
	}

	roles, err := r.cfClient.Roles.ListAll(ctx, roleListOptions)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Space Roles",
			"Could not list roles for space "+filter.Space.ValueString()+": "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, role := range roles {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("role_guid"), role.GUID)

			if req.IncludeResource {
				resRole := mapRoleValuesToType(role)
				result.Diagnostics.Append(result.Resource.Set(ctx, resRole.ReduceToSpaceRole())...)
			}

			if !push(result) {
				return
			}
		}
	}
}
