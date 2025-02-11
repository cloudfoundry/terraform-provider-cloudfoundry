package provider

import (
	"context"

	cfresource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type servicePlanVisibilityType struct {
	Organizations   []organizationType `tfsdk:"organizations"`
	ServicePlanGUID types.String       `tfsdk:"service_plan_guid"`
	SpaceGUID       types.String       `tfsdk:"space_guid"`
	Type            types.String       `tfsdk:"type"`
}

type organizationType struct {
	GUID types.String `tfsdk:"guid"`
}

type datasourceServicePlanVisibilityType struct {
	Organizations   []organizationType `tfsdk:"organizations"`
	ServicePlanGUID types.String       `tfsdk:"service_plan_guid"`
	SpaceGUID       types.String       `tfsdk:"space_guid"`
	Type            types.String       `tfsdk:"type"`
}

func (a *servicePlanVisibilityType) Reduce() datasourceServicePlanVisibilityType {
	var reduced datasourceServicePlanVisibilityType
	copyFields(&reduced, a)
	return reduced
}

func mapServicePlanVisibilityValuesToType(ctx context.Context, value *cfresource.ServicePlanVisibility) (servicePlanVisibilityType, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var organizations []organizationType

	for _, org := range value.Organizations {
		organizations = append(organizations, organizationType{GUID: types.StringValue(org.GUID)})
	}

	servicePlanVisibilityType := servicePlanVisibilityType{
		Type:          types.StringValue(value.Type),
		SpaceGUID:     types.StringValue(value.Space.GUID),
		Organizations: organizations,
	}

	return servicePlanVisibilityType, diagnostics
}

func mapCreateServicePlanVisibilityTypeToValues(ctx context.Context, value servicePlanVisibilityType) (*cfresource.ServicePlanVisibility, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	visibilityType := value.Type.ValueString()

	visibilityTypeEnum, err := cfresource.ParseServicePlanVisibilityType(visibilityType)
	if err != nil {
		diagnostics.AddError("Invalid Visibility Type", "The provided visibility type is not valid: "+visibilityType)
		return nil, diagnostics
	}

	createServicePlanVisibility := cfresource.NewServicePlanVisibilityUpdate(visibilityTypeEnum)

	for _, org := range value.Organizations {
		if !org.GUID.IsNull() && org.GUID.ValueString() != "" {
			createServicePlanVisibility.Organizations = append(createServicePlanVisibility.Organizations, cfresource.ServicePlanVisibilityRelation{
				GUID: org.GUID.ValueString(),
			})
		}
	}

	return createServicePlanVisibility, diagnostics
}
