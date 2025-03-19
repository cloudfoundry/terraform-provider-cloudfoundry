package provider

import (
	"context"

	cfresource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type servicePlanVisibilityType struct {
	Organizations   types.Set    `tfsdk:"organizations"`
	ServicePlanGUID types.String `tfsdk:"service_plan"`
	SpaceGUID       types.String `tfsdk:"space"`
	Type            types.String `tfsdk:"type"`
}

func mapServicePlanVisibilityValuesToType(ctx context.Context, value *cfresource.ServicePlanVisibility, plan servicePlanVisibilityType) (servicePlanVisibilityType, diag.Diagnostics) {
	var diagnostics, diags diag.Diagnostics
	var allOrganizations, plannedOrgs []string

	for _, org := range value.Organizations {
		allOrganizations = append(allOrganizations, org.GUID)
	}

	if !plan.Organizations.IsNull() {
		diags := plan.Organizations.ElementsAs(ctx, &plannedOrgs, false)
		diagnostics.Append(diags...)
	}

	commonOrgs := lo.Intersect(plannedOrgs, allOrganizations)

	servicePlanVisibilityType := servicePlanVisibilityType{
		Type:            types.StringValue(value.Type),
		ServicePlanGUID: plan.ServicePlanGUID,
	}

	if len(commonOrgs) > 0 {
		servicePlanVisibilityType.Organizations, diags = types.SetValueFrom(ctx, types.StringType, commonOrgs)
		diagnostics.Append(diags...)
	} else {
		servicePlanVisibilityType.Organizations = types.SetNull(types.StringType)
	}

	if value.Space != nil {
		servicePlanVisibilityType.SpaceGUID = types.StringValue(value.Space.GUID)
	}

	return servicePlanVisibilityType, diagnostics
}

func mapCreateServicePlanVisibilityTypeToValues(ctx context.Context, value servicePlanVisibilityType) (*cfresource.ServicePlanVisibility, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var orgGUIDs []string

	createServicePlanVisibility := cfresource.ServicePlanVisibility{
		Type: value.Type.ValueString(),
	}

	if !value.Organizations.IsNull() {
		diags := value.Organizations.ElementsAs(ctx, &orgGUIDs, false)
		diagnostics.Append(diags...)
	}

	for _, orgGUID := range orgGUIDs {
		createServicePlanVisibility.Organizations = append(createServicePlanVisibility.Organizations, cfresource.ServicePlanVisibilityRelation{
			GUID: orgGUID,
		})
	}

	return &createServicePlanVisibility, diagnostics
}
