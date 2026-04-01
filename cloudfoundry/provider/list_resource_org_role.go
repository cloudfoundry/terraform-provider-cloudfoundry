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

var _ list.ListResourceWithConfigure = &orgRoleListResource{}

type orgRoleListResource struct {
	cfClient *cfv3client.Client
}

type orgRoleListResourceFilter struct {
	Org  types.String `tfsdk:"org"`
	Type types.String `tfsdk:"type"`
	User types.String `tfsdk:"user"`
}

func NewOrgRoleListResource() list.ListResource {
	return &orgRoleListResource{}
}

func (r *orgRoleListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_role" // must match managed resource
}

func (r *orgRoleListResource) Configure(_ context.Context,
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

func (r *orgRoleListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all roles within an organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to list roles for.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Role type to filter by; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("organization_auditor", "organization_user", "organization_manager", "organization_billing_manager"),
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

// List streams all org roles from the API.
func (r *orgRoleListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter orgRoleListResourceFilter

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

	roleListOptions := cfv3client.NewRoleListOptions()
	roleListOptions.OrganizationGUIDs = cfv3client.Filter{
		Values: []string{filter.Org.ValueString()},
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
			"API Error Fetching Org Roles",
			"Could not list roles for organization "+filter.Org.ValueString()+": "+err.Error(),
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
				result.Diagnostics.Append(result.Resource.Set(ctx, resRole.ReduceToOrgRole())...)
			}

			if !push(result) {
				return
			}
		}
	}
}
