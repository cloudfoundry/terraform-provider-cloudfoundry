package provider

import (
	"context"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
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
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithConfigure   = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
	_ resource.ResourceWithIdentity    = &UserResource{}
)

// Instantiates a user resource.
func NewCFUserResource() resource.Resource {
	return &UserCFResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type UserCFResource struct {
	cfClient *cfv3client.Client
}

type userCfResourceIdentityModel struct {
	UserGUID types.String `tfsdk:"user_guid"`
}

func (r *UserCFResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_cf"
}

func (r *UserCFResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a resource for creating users in the Cloud Controller database. Supports creating a user via username and origin which can be accomplished by Org Managers provided CAPI property cc.allow_user_creation_by_org_manager is enabled. No explicit calls are made to the UAA endpoints as part of the resource creation.
		__Note__ : An Org manger will not be able to retrieve information of a newly created user until the user is assigned an org-role to any of the orgs that can be accessed by the Org Manager. If the end goal is role assignment, better would be to use the cloudfoundry_org_role or cloudfoundry_space_role to accomplish the same.`,
		Attributes: map[string]schema.Attribute{
			idKey: schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the user.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("id"),
						path.MatchRoot("username"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "User name of the user, typically an email address.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"presentation_name": schema.StringAttribute{
				MarkdownDescription: "The name displayed for the user; for UAA users, this is the same as the username. For UAA clients, this is the UAA client ID",
				Computed:            true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The alias of the Identity Provider that authenticated this user.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.NoneOf("uaa"),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("id"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("username"),
					}...),
				},
			},
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (rs *UserCFResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"user_guid": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *UserCFResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, _ := req.ProviderData.(*managers.Session)
	r.cfClient = session.CFClient
}

func (r *UserCFResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan userType
		err  error
		user *cfv3resource.User
	)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Id.IsUnknown() {
		createUser := &cfv3resource.UserCreate{
			GUID: plan.Id.ValueString(),
		}
		createUser.Metadata = cfv3resource.NewMetadata()
		labelsDiags := plan.Labels.ElementsAs(ctx, &createUser.Metadata.Labels, false)
		resp.Diagnostics.Append(labelsDiags...)

		annotationsDiags := plan.Annotations.ElementsAs(ctx, &createUser.Metadata.Annotations, false)
		resp.Diagnostics.Append(annotationsDiags...)

		user, err = r.cfClient.Users.Create(ctx, createUser)

	} else {
		createUser := &cfv3resource.UserCreateWithUsername{
			Username: plan.UserName.ValueString(),
			Origin:   plan.Origin.ValueString(),
		}
		createUser.Metadata = cfv3resource.NewMetadata()
		labelsDiags := plan.Labels.ElementsAs(ctx, &createUser.Metadata.Labels, false)
		resp.Diagnostics.Append(labelsDiags...)

		annotationsDiags := plan.Annotations.ElementsAs(ctx, &createUser.Metadata.Annotations, false)
		resp.Diagnostics.Append(annotationsDiags...)

		user, err = r.cfClient.Users.CreateWithUsername(ctx, createUser)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating CF User",
			"Could not create User "+err.Error(),
		)
		return
	}

	plan, diags = mapUserValuesToType(ctx, user)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a cf user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := userCfResourceIdentityModel{
		UserGUID: types.StringValue(plan.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (rs *UserCFResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfUser, err := rs.cfClient.Users.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "user", data.Id.ValueString())
		return
	}

	data, diags = mapUserValuesToType(ctx, cfUser)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a cf user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var identity userCfResourceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = userCfResourceIdentityModel{
			UserGUID: types.StringValue(data.Id.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}

}

func (rs *UserCFResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState userType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateCFUser, diags := mapUpdateUserTypeToValues(ctx, previousState.Labels, previousState.Annotations, plan.Labels, plan.Annotations)
	resp.Diagnostics.Append(diags...)

	cfUser, err := rs.cfClient.Users.Update(ctx, plan.Id.ValueString(), &updateCFUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating CF User",
			"Could not update User with ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	data, diags := mapUserValuesToType(ctx, cfUser)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a cf user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// WORKAROUND for OpenTofu compatibility
	// https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/418
	identity := userCfResourceIdentityModel{
		UserGUID: types.StringValue(data.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
	// END WORKAROUND
}

func (rs *UserCFResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state userType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.Users.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting CF User",
			"Could not delete the user with ID "+state.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting CF User",
			"Failed in deleting the user with ID "+state.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a cf user resource")

}

func (rs *UserCFResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("user_guid"), req, resp)
}
