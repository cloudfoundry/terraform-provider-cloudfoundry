package provider

import (
	"context"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform struct for storing values for user data source.
type userGroupsType struct {
	User   types.String `tfsdk:"user"`
	Origin types.String `tfsdk:"origin"`
	Groups types.Set    `tfsdk:"groups"`
}

// Sets the terraform struct values from the user resource returned by the uaa-client.
func (plan *userGroupsType) mapUserGroupsResourcesValuesToType(ctx context.Context, uaaUser *uaa.User) diag.Diagnostics {

	var (
		diagnostics diag.Diagnostics
		groups      []string
	)
	for _, group := range uaaUser.Groups {
		groups = append(groups, group.Display)
	}
	groupsSet, diags := types.SetValueFrom(ctx, types.StringType, groups)
	diagnostics.Append(diags...)
	sameGroups, diags := findSameRelationsFromTFState(ctx, plan.Groups, groupsSet)
	diagnostics.Append(diags...)
	plan.Groups, diags = types.SetValueFrom(ctx, types.StringType, sameGroups)
	diagnostics.Append(diags...)

	return diagnostics
}
