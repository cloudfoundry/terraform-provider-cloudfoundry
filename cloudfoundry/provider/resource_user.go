package provider

import (
	"context"
	"fmt"
	"strings"

	uaa "github.com/cloudfoundry-community/go-uaa"
	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
func NewUserResource() resource.Resource {
	return &UserResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type UserResource struct {
	cfClient  *cfv3client.Client
	uaaClient *uaa.API
}

type userResourceIdentityModel struct {
	UserGUID types.String `tfsdk:"user_guid"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a resource for creating users in the origin store and registering them in Cloud Foundry. If the origin store user or the CF user already exists with the given username, it will fetch that user and store it in the state for resource management.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "User name of the user, typically an email address.",
				Required:            true,
			},
			"password": &schema.StringAttribute{
				MarkdownDescription: "User's password, required if origin is set to uaa.",
				Optional:            true,
				Sensitive:           true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The alias of the Identity Provider that authenticated this user.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"given_name": schema.StringAttribute{
				MarkdownDescription: "The user's first name.",
				Optional:            true,
			},
			"family_name": schema.StringAttribute{
				MarkdownDescription: "The user's last name.",
				Optional:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the user. When not provided, name is used as email.",
				Optional:            true,
				Computed:            true,
			},
			"groups": schema.SetAttribute{
				MarkdownDescription: "Any UAA groups / roles to associate the user with.",
				ElementType:         types.StringType,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},

			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
			idKey:          guidSchema(),
		},
	}
}

func (rs *UserResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"user_guid": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, _ := req.ProviderData.(*managers.Session)
	r.cfClient = session.CFClient

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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uaaUser, err := r.uaaClient.GetUserByUsername(plan.UserName.ValueString(), plan.Origin.ValueString(), "")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			createUAAUser := plan.mapCreateUAAUserTypeToValues()
			uaaUser, err = r.uaaClient.CreateUser(createUAAUser)
			if err != nil {
				resp.Diagnostics.AddError(
					"API Error Creating User in Origin",
					"Could not create User with Name "+plan.UserName.ValueString()+" : "+err.Error(),
				)
				return
			}
		} else {
			resp.Diagnostics.AddError(
				"API Error Fetching User From Origin",
				"Could not Fetch User with Name "+plan.UserName.ValueString()+" : "+err.Error(),
			)
			return
		}
	}

	cfUser, err := r.cfClient.Users.Get(ctx, uaaUser.ID)
	if err != nil {
		if cfv3resource.IsResourceNotFoundError(err) {
			createCFUser, diags := plan.mapCreateCFUserTypeToValues(ctx, uaaUser.ID)
			resp.Diagnostics.Append(diags...)
			cfUser, err = r.cfClient.Users.Create(ctx, &createCFUser)
			if err != nil {
				resp.Diagnostics.AddError(
					"API Error Creating CF User",
					"Could not Create User with ID "+uaaUser.ID+" : "+err.Error(),
				)
				return
			}

		} else {
			resp.Diagnostics.AddError(
				"API Error Fetching CF User",
				"Could not Fetch User with ID "+uaaUser.ID+" : "+err.Error(),
			)
			return
		}
	}

	plan, diags = mapUserResourcesValuesToType(ctx, uaaUser, cfUser, plan.Password)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := userResourceIdentityModel{
		UserGUID: types.StringValue(plan.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (rs *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userResourceType

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

	uaaUser, err := rs.uaaClient.GetUser(data.Id.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "scim_resource_not_found") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(fmt.Sprintf("API Error Reading %s %s", "user", data.Id.ValueString()), err.Error())
		}
		return
	}

	data, diags = mapUserResourcesValuesToType(ctx, uaaUser, cfUser, data.Password)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var identity userResourceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = userResourceIdentityModel{
			UserGUID: types.StringValue(data.Id.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}

}

func (rs *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState userResourceType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateUAAUser := plan.mapUpdateUAAUserTypeToValues()
	uaaUser, err := rs.uaaClient.UpdateUser(updateUAAUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating User in Origin",
			"Could not update User with Id "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	if previousState.Password != plan.Password {
		//Change user password
		reqPath := "Users/" + uaaUser.ID + "/password"
		reqMethod := "PUT"
		reqData := `{ "oldPassword" : "` + previousState.Password.ValueString() + `" , "password" : "` + plan.Password.ValueString() + `"}`
		reqHeaders := []string{"Content-Type: application/json", "Accept: application/json"}
		_, respBody, respCode, err := rs.uaaClient.Curl(reqPath, reqMethod, reqData, reqHeaders)

		var errMsg string
		if respCode != 200 || err != nil {
			if err != nil {
				errMsg = err.Error()
			} else {
				errMsg = respBody
			}
			resp.Diagnostics.AddError(
				"API Error Updating Password of User",
				"Could not update Password for user with Id "+plan.Id.ValueString()+" : "+errMsg,
			)
			return
		}
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

	data, diags := mapUserResourcesValuesToType(ctx, uaaUser, cfUser, plan.Password)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// WORKAROUND for OpenTofu compatibility
	// https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/418
	identity := userResourceIdentityModel{
		UserGUID: types.StringValue(data.Id.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
	// END WORKAROUND
}

func (rs *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state userResourceType
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

	_, err = rs.uaaClient.DeleteUser(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting UAA User",
			"Failed in deleting the user with ID "+state.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a user resource")

}

func (rs *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("user_guid"), req, resp)
}
