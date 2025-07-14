package provider

import (
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleType struct {
	Type         types.String `tfsdk:"type"`
	User         types.String `tfsdk:"user"`
	UserName     types.String `tfsdk:"username"`
	Origin       types.String `tfsdk:"origin"`
	Space        types.String `tfsdk:"space"`
	Id           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"org"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

type spaceRoleType struct {
	Type      types.String `tfsdk:"type"`
	User      types.String `tfsdk:"user"`
	UserName  types.String `tfsdk:"username"`
	Origin    types.String `tfsdk:"origin"`
	Space     types.String `tfsdk:"space"`
	Id        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type orgRoleType struct {
	Type         types.String `tfsdk:"type"`
	User         types.String `tfsdk:"user"`
	UserName     types.String `tfsdk:"username"`
	Origin       types.String `tfsdk:"origin"`
	Id           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"org"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

type spaceRoleDatasourceType struct {
	Type      types.String `tfsdk:"type"`
	User      types.String `tfsdk:"user"`
	Space     types.String `tfsdk:"space"`
	Id        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type orgRoleDatasourceType struct {
	Type         types.String `tfsdk:"type"`
	User         types.String `tfsdk:"user"`
	Id           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"org"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

type orgRolesDatasourceType struct {
	Type  types.String            `tfsdk:"type"`
	User  types.String            `tfsdk:"user"`
	Org   types.String            `tfsdk:"org"`
	Roles []orgRoleDatasourceType `tfsdk:"roles"`
}

type spaceRolesDatasourceType struct {
	Type  types.String              `tfsdk:"type"`
	User  types.String              `tfsdk:"user"`
	Space types.String              `tfsdk:"space"`
	Roles []spaceRoleDatasourceType `tfsdk:"roles"`
}

// Reduce function to reduce roleType to roleDatasourceType
// This is used to reuse mapRoleValuesToType in both resource and datasource.
func (a *roleType) ReduceToSpaceRoleDataSource() spaceRoleDatasourceType {
	var reduced spaceRoleDatasourceType
	copyFields(&reduced, a)
	return reduced
}

func (a *roleType) ReduceToOrgRoleDataSource() orgRoleDatasourceType {
	var reduced orgRoleDatasourceType
	copyFields(&reduced, a)
	return reduced
}

func (a *roleType) ReduceToSpaceRole() spaceRoleType {
	var reduced spaceRoleType
	copyFields(&reduced, a)
	return reduced
}

func (a *roleType) ReduceToOrgRole() orgRoleType {
	var reduced orgRoleType
	copyFields(&reduced, a)
	return reduced
}

// Returns the OrganizationRoleType value needed for org role creation.
func (data *orgRoleType) getOrgRoleType() resource.OrganizationRoleType {

	switch data.Type.ValueString() {
	case "organization_user":
		return resource.OrganizationRoleUser
	case "organization_auditor":
		return resource.OrganizationRoleAuditor
	case "organization_manager":
		return resource.OrganizationRoleManager
	case "organization_billing_manager":
		return resource.OrganizationRoleBillingManager
	default:
		return resource.OrganizationRoleNone

	}
}

// Returns the OrganizationRoleType value needed for org role creation.
func (data *spaceRoleType) getSpaceRoleType() resource.SpaceRoleType {

	switch data.Type.ValueString() {
	case "space_auditor":
		return resource.SpaceRoleAuditor
	case "space_developer":
		return resource.SpaceRoleDeveloper
	case "space_manager":
		return resource.SpaceRoleManager
	case "space_supporter":
		return resource.SpaceRoleSupporter
	default:
		return resource.SpaceRoleNone

	}
}

// Sets the terraform struct values from the user resource returned by the cf-client.
func mapRoleValuesToType(role *resource.Role) roleType {

	roleType := roleType{
		Id:        types.StringValue(role.GUID),
		CreatedAt: types.StringValue(role.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(role.UpdatedAt.Format(time.RFC3339)),
		Type:      types.StringValue(role.Type),
		User:      types.StringValue(role.Relationships.User.Data.GUID),
	}

	if role.Relationships.Org.Data != nil {
		roleType.Organization = types.StringValue(role.Relationships.Org.Data.GUID)
	}

	if role.Relationships.Space.Data != nil {
		roleType.Space = types.StringValue(role.Relationships.Space.Data.GUID)
	}

	return roleType
}

func mapOrgRolesValuesToType(roles []*resource.Role) []orgRoleDatasourceType {

	orgRolesList := []orgRoleDatasourceType{}
	for _, orgRole := range roles {
		orgRoleValue := mapRoleValuesToType(orgRole)
		orgRolesList = append(orgRolesList, orgRoleValue.ReduceToOrgRoleDataSource())
	}

	return orgRolesList
}

func mapSpaceRolesValuesToType(roles []*resource.Role) []spaceRoleDatasourceType {

	spaceRolesList := []spaceRoleDatasourceType{}
	for _, spaceRole := range roles {
		spaceRoleValue := mapRoleValuesToType(spaceRole)
		spaceRolesList = append(spaceRolesList, spaceRoleValue.ReduceToSpaceRoleDataSource())
	}

	return spaceRolesList
}
