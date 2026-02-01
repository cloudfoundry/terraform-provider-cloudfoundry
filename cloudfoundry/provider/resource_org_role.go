package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &OrgRoleResource{}
	_ resource.ResourceWithConfigure   = &OrgRoleResource{}
	_ resource.ResourceWithImportState = &OrgRoleResource{}
	_ resource.ResourceWithIdentity    = &OrgRoleResource{}
)

// Instantiates a role resource.
func NewOrgeRoleResource() resource.Resource {
	return &OrgRoleResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type OrgRoleResource struct {
	cfClient *cfv3client.Client
}

type orgRoleResouerceIdentityModel struct {
	RoleGUID types.String `tfsdk:"role_guid"`
}

func (r *OrgRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_role"
}

func (r *OrgRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a Cloud Foundry resource for assigning org roles.(Updating a role is not supported according to the docs).
		__Note__ : If a role is assigned to a non existent user, user will be created automatically in the Cloud Controller's database for an Admin as well as Org Manager provided the cc.allow_user_creation_by_org_manager field is enabled. Refer [here](https://v3-apidocs.cloudfoundry.org/version/3.189.0/index.html#create-a-role) for further details.`,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Role type; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("organization_user", "organization_auditor", "organization_manager", "organization_billing_manager"),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The guid of the cloudfoundry user to assign the role with",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("user"),
						path.MatchRoot("username"),
					}...),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username of the cloudfoundry user to assign the role with",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The identity provider for the UAA user",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("user"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("username"),
					}...),
				},
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The guid of the organization to assign the role to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			idKey:        guidSchema(),
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (rs *OrgRoleResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	{
		resp.IdentitySchema = identityschema.Schema{
			Attributes: map[string]identityschema.Attribute{
				"role_guid": identityschema.StringAttribute{
					RequiredForImport: true,
				},
			},
		}
	}
}

func (r *OrgRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orgRoleType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		role *cfv3resource.Role
		err  error
	)

	orgRoleType := plan.getOrgRoleType()
	if !plan.User.IsUnknown() {
		role, err = r.cfClient.Roles.CreateOrganizationRole(ctx, plan.Organization.ValueString(), plan.User.ValueString(), orgRoleType)
	} else {
		role, err = r.cfClient.Roles.CreateOrganizationRoleWithUsername(ctx, plan.Organization.ValueString(), plan.UserName.ValueString(), orgRoleType, plan.Origin.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Registering Role",
			"Could not register Role with user ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	roleTypeResponse := mapRoleValuesToType(role)
	data := roleTypeResponse.ReduceToOrgRole()
	data.UserName = plan.UserName
	data.Origin = plan.Origin

	tflog.Trace(ctx, "created an org role resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := orgRoleResouerceIdentityModel{
		RoleGUID: types.StringValue(data.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)

}

func (rs *OrgRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orgRoleType

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := rs.cfClient.Roles.Get(ctx, state.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "role", state.Id.ValueString())
		return
	}

	roleTypeResponse := mapRoleValuesToType(role)
	data := roleTypeResponse.ReduceToOrgRole()
	data.UserName = state.UserName
	data.Origin = state.Origin

	tflog.Trace(ctx, "read an org role resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var identity orgRoleResouerceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = orgRoleResouerceIdentityModel{
			RoleGUID: types.StringValue(data.Id.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

// Update for role is not possible.
func (rs *OrgRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (rs *OrgRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state orgRoleType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.Roles.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Role",
			"Could not delete the role with ID "+state.Id.ValueString()+" and user ID "+state.User.ValueString()+" : "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Role",
			"Failed in deleting the role with ID "+state.Id.ValueString()+" and user ID "+state.User.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted an org role resource")
}

func (rs *OrgRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("role_guid"), req, resp)
}
