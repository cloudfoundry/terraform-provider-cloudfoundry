package provider

import (
	"context"
	"fmt"
	"strings"

	uaa "github.com/cloudfoundry-community/go-uaa"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &UserGroupsResource{}
	_ resource.ResourceWithConfigure = &UserGroupsResource{}
)

// Instantiates a user resource.
func NewUserGroupsResource() resource.Resource {
	return &UserGroupsResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type UserGroupsResource struct {
	uaaClient *uaa.API
}

func (r *UserGroupsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}

func (r *UserGroupsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a resource for adding or removing a user from various UAA origin store groups. If user already present in a particular group, no action will be taken",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				MarkdownDescription: "GUID of the isolation segment.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The user authentcation origin.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"groups": schema.SetAttribute{
				MarkdownDescription: "The authorization scope or groups display names to add the user to.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (r *UserGroupsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, _ := req.ProviderData.(*managers.Session)

	var err error
	uaaUrl := session.CFClient.AuthURL("")
	uaaClient := uaa.WithClient(session.CFClient.HTTPAuthClient())
	uaaUserAgent := uaa.WithUserAgent(session.CFClient.UserAgent())
	r.uaaClient, err = uaa.New(uaaUrl, uaa.WithNoAuthentication(), uaaClient, uaaUserAgent)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to initialise the UAA client",
			fmt.Sprintf("Error : %s .Please report this issue to the provider developers.", err.Error()),
		)
		return
	}
}

func (r *UserGroupsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   userGroupsType
		groups []string
	)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = plan.Groups.ElementsAs(ctx, &groups, false)
	resp.Diagnostics.Append(diags...)

	for _, groupToAdd := range groups {
		group, err := r.uaaClient.GetGroupByName(groupToAdd, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching UAA Group",
				"Could not Fetch Group with Name "+groupToAdd+" : "+err.Error(),
			)
			return
		}
		err = r.uaaClient.AddGroupMember(group.ID, plan.User.ValueString(), "USER", plan.Origin.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "member_already_exists") {
				resp.Diagnostics.AddWarning(
					"API Error Adding User to Group",
					"Could not add user with ID "+plan.User.ValueString()+" to Group "+groupToAdd+" : "+err.Error(),
				)
				continue
			}
			resp.Diagnostics.AddError(
				"API Error Adding User to Group",
				"Could not add user with ID "+plan.User.ValueString()+" to Group "+groupToAdd+" : "+err.Error(),
			)
			return
		}
	}

	uaaUser, err := r.uaaClient.GetUser(plan.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching User From Origin",
			"Could not Fetch User with ID "+plan.User.ValueString()+" : "+err.Error(),
		)
		return
	}

	diags = plan.mapUserGroupsResourcesValuesToType(ctx, uaaUser)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a user groups resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *UserGroupsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userGroupsType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uaaUser, err := rs.uaaClient.GetUser(data.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching User From Origin",
			"Could not Fetch User with ID "+data.User.ValueString()+" : "+err.Error(),
		)
		return
	}

	diags = data.mapUserGroupsResourcesValuesToType(ctx, uaaUser)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a user groups resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *UserGroupsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState userGroupsType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removedRoles, addedRoles, diags := findChangedRelationsFromTFState(ctx, plan.Groups, previousState.Groups)
	resp.Diagnostics.Append(diags...)

	for _, groupToRemove := range removedRoles {
		group, err := rs.uaaClient.GetGroupByName(groupToRemove, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching UAA Group",
				"Could not Fetch Group with Name "+groupToRemove+" : "+err.Error(),
			)
			return
		}
		err = rs.uaaClient.RemoveGroupMember(group.ID, plan.User.ValueString(), "USER", plan.Origin.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Removing User from Group",
				"Could not remove user with ID "+plan.User.ValueString()+" from Group "+groupToRemove+" : "+err.Error(),
			)
			return
		}
	}

	for _, groupToAdd := range addedRoles {
		group, err := rs.uaaClient.GetGroupByName(groupToAdd, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching UAA Group",
				"Could not Fetch Group with Name "+groupToAdd+" : "+err.Error(),
			)
			return
		}
		err = rs.uaaClient.AddGroupMember(group.ID, plan.User.ValueString(), "USER", plan.Origin.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "member_already_exists") {
				resp.Diagnostics.AddWarning(
					"API Error Adding User to Group",
					"Could not add user with ID "+plan.User.ValueString()+" to Group "+groupToAdd+" : "+err.Error(),
				)
				continue
			}
			resp.Diagnostics.AddError(
				"API Error Adding User to Group",
				"Could not add user with ID "+plan.User.ValueString()+" to Group "+groupToAdd+" : "+err.Error(),
			)
			return
		}
	}

	uaaUser, err := rs.uaaClient.GetUser(plan.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching User From Origin",
			"Could not Fetch User with ID "+plan.User.ValueString()+" : "+err.Error(),
		)
		return
	}

	diags = plan.mapUserGroupsResourcesValuesToType(ctx, uaaUser)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a user groups resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *UserGroupsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var (
		state  userGroupsType
		groups []string
	)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = state.Groups.ElementsAs(ctx, &groups, false)
	resp.Diagnostics.Append(diags...)
	for _, groupToRemove := range groups {
		group, err := rs.uaaClient.GetGroupByName(groupToRemove, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching UAA Group",
				"Could not Fetch Group with Name "+groupToRemove+" : "+err.Error(),
			)
			return
		}
		err = rs.uaaClient.RemoveGroupMember(group.ID, state.User.ValueString(), "USER", state.Origin.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Removing User from Group",
				"Could not remove user with ID "+state.User.ValueString()+" from Group "+groupToRemove+" : "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(ctx, "deleted a user groups resource")

}
