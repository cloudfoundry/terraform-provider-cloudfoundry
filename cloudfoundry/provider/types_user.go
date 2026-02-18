package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Terraform struct for storing values for user data source.
type userType struct {
	UserName         types.String `tfsdk:"username"`
	Origin           types.String `tfsdk:"origin"`
	Id               types.String `tfsdk:"id"`
	Labels           types.Map    `tfsdk:"labels"`
	PresentationName types.String `tfsdk:"presentation_name"`
	Annotations      types.Map    `tfsdk:"annotations"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

type userResourceType struct {
	Password    types.String `tfsdk:"password"`
	GivenName   types.String `tfsdk:"given_name"`
	FamilyName  types.String `tfsdk:"family_name"`
	UserName    types.String `tfsdk:"username"`
	Origin      types.String `tfsdk:"origin"`
	Email       types.String `tfsdk:"email"`
	Groups      types.Set    `tfsdk:"groups"`
	Id          types.String `tfsdk:"id"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type datasourceUserType struct {
	Users []userType   `tfsdk:"users"`
	Name  types.String `tfsdk:"name"`
}

type spaceOrgUsersType struct {
	Users        []userType   `tfsdk:"users"`
	Space        types.String `tfsdk:"space"`
	Organization types.String `tfsdk:"org"`
}

// Sets the user resource values for creation with uaa-client from the terraform struct values.
func (plan *userResourceType) mapCreateUAAUserTypeToValues() uaa.User {

	email := plan.Email.ValueString()
	if email == "" {
		email = plan.UserName.ValueString()
	}
	emails := []uaa.Email{{
		Value:   email,
		Primary: new(true),
	},
	}
	name := uaa.UserName{
		GivenName:  plan.GivenName.ValueString(),
		FamilyName: plan.FamilyName.ValueString(),
	}

	createUAAUser := uaa.User{
		Username: plan.UserName.ValueString(),
		Password: plan.Password.ValueString(),
		Origin:   plan.Origin.ValueString(),
		Name:     &name,
		Emails:   emails,
	}

	return createUAAUser
}

// Sets the user resource values for creation with cf-client from the terraform struct values.
func (data *userResourceType) mapCreateCFUserTypeToValues(ctx context.Context, userId string) (resource.UserCreate, diag.Diagnostics) {

	createUser := &resource.UserCreate{GUID: userId}
	var diagnostics diag.Diagnostics
	createUser.Metadata = resource.NewMetadata()

	labelsDiags := data.Labels.ElementsAs(ctx, &createUser.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)

	annotationsDiags := data.Annotations.ElementsAs(ctx, &createUser.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createUser, diagnostics
}

// Sets the terraform struct values from the user resource returned by the cf-client.
func mapUserResourcesValuesToType(ctx context.Context, uaaUser *uaa.User, cfUser *resource.User, password types.String) (userResourceType, diag.Diagnostics) {

	userResourceType := userResourceType{
		Id:        types.StringValue(cfUser.GUID),
		CreatedAt: types.StringValue(cfUser.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(cfUser.UpdatedAt.Format(time.RFC3339)),
		Email:     types.StringValue(uaaUser.Emails[0].Value),
	}

	if cfUser.Username != nil {
		userResourceType.UserName = types.StringValue(*cfUser.Username)
	}

	if cfUser.Origin != nil {
		userResourceType.Origin = types.StringValue(*cfUser.Origin)
	}

	var diags, diagnostics diag.Diagnostics
	userResourceType.Labels, diags = mapMetadataValueToType(ctx, cfUser.Metadata.Labels)
	diagnostics.Append(diags...)
	userResourceType.Annotations, diags = mapMetadataValueToType(ctx, cfUser.Metadata.Annotations)
	diagnostics.Append(diags...)

	if uaaUser.Name.FamilyName != "" {
		userResourceType.FamilyName = types.StringValue(uaaUser.Name.FamilyName)
	}
	if uaaUser.Name.GivenName != "" {
		userResourceType.GivenName = types.StringValue(uaaUser.Name.GivenName)
	}
	userResourceType.Password = password
	var groups []string
	for _, group := range uaaUser.Groups {
		groups = append(groups, group.Display)
	}
	userResourceType.Groups, diags = types.SetValueFrom(ctx, types.StringType, groups)
	diagnostics.Append(diags...)

	return userResourceType, diagnostics
}

// Sets the terraform struct values from the user resource returned by the cf-client.
func mapUserValuesToType(ctx context.Context, user *resource.User) (userType, diag.Diagnostics) {

	userType := userType{
		Id:               types.StringValue(user.GUID),
		CreatedAt:        types.StringValue(user.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:        types.StringValue(user.UpdatedAt.Format(time.RFC3339)),
		PresentationName: types.StringValue(user.PresentationName),
	}

	if user.Username != nil {
		userType.UserName = types.StringValue(*user.Username)
	}

	if user.Origin != nil {
		userType.Origin = types.StringValue(*user.Origin)
	}

	var diags, diagnostics diag.Diagnostics
	userType.Labels, diags = mapMetadataValueToType(ctx, user.Metadata.Labels)
	diagnostics.Append(diags...)
	userType.Annotations, diags = mapMetadataValueToType(ctx, user.Metadata.Annotations)
	diagnostics.Append(diags...)

	return userType, diagnostics
}

// Sets the user resource values for creation with uaa-client from the terraform struct values.
func (plan *userResourceType) mapUpdateUAAUserTypeToValues() uaa.User {

	email := plan.Email.ValueString()
	if email == "" {
		email = plan.UserName.ValueString()
	}
	emails := []uaa.Email{{
		Value:   email,
		Primary: new(true),
	},
	}
	name := uaa.UserName{
		GivenName:  plan.GivenName.ValueString(),
		FamilyName: plan.FamilyName.ValueString(),
	}

	updateUAAUser := uaa.User{
		ID:       plan.Id.ValueString(),
		Username: plan.UserName.ValueString(),
		Origin:   plan.Origin.ValueString(),
		Name:     &name,
		Emails:   emails,
	}

	return updateUAAUser
}

// Sets the user resource values for updation with cf-client from the terraform struct values.
func mapUpdateUserTypeToValues(ctx context.Context, stateLabels basetypes.MapValue, stateAnnotations basetypes.MapValue, planLabels basetypes.MapValue, planAnnotations basetypes.MapValue) (resource.UserUpdate, diag.Diagnostics) {

	updateUser := &resource.UserUpdate{}

	var diagnostics diag.Diagnostics
	updateUser.Metadata, diagnostics = setClientMetadataForUpdate(ctx, stateLabels, stateAnnotations, planLabels, planAnnotations)

	return *updateUser, diagnostics
}

// Prepares a terraform list from the user resources returned by the cf-client.
func mapUsersValuesToListType(ctx context.Context, users []*resource.User) ([]userType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	userValues := []userType{}
	for _, user := range users {
		userValue, diags := mapUserValuesToType(ctx, user)
		diagnostics.Append(diags...)
		userValues = append(userValues, userValue)
	}

	return userValues, diagnostics
}

// Sets the terraform struct values from the user resources returned by the cf-client.
func mapUsersValuesToType(ctx context.Context, data datasourceUserType, users []*resource.User) (datasourceUserType, diag.Diagnostics) {

	var diags, diagnostics diag.Diagnostics
	usersType := datasourceUserType{
		Name: data.Name,
	}

	usersType.Users, diags = mapUsersValuesToListType(ctx, users)
	diagnostics.Append(diags...)

	return usersType, diagnostics
}

// Sets the terraform struct values from the user resources returned by the cf-client.
func mapSpaceOrgUsersValuesToType(ctx context.Context, data spaceOrgUsersType, users []*resource.User) (spaceOrgUsersType, diag.Diagnostics) {

	var diags, diagnostics diag.Diagnostics
	spaceOrgUsersType := spaceOrgUsersType{}

	if !data.Organization.IsNull() {
		spaceOrgUsersType.Organization = data.Organization
	}

	if !data.Space.IsNull() {
		spaceOrgUsersType.Space = data.Space
	}

	spaceOrgUsersType.Users, diags = mapUsersValuesToListType(ctx, users)
	diagnostics.Append(diags...)

	return spaceOrgUsersType, diagnostics
}
