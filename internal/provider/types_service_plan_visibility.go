package provider

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type servicePlanVisibilityType struct {
	OrganizationGUID types.String `tfsdk:"organization_guid"`
	ServicePlanGUID  types.String `tfsdk:"service_plan_guid"`
	Labels           types.Map    `tfsdk:"labels"`
	Annotations      types.Map    `tfsdk:"annotations"`
}

func (data *servicePlanVisibilityType) mapCreateServicePlanVisibilityTypeToValues(ctx context.Context) (resource.ServicePlanVisibilityCreate, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	createServicePlanVisibility := resource.ServicePlanVisibilityCreate{
		ServicePlanGUID:  data.ServicePlanGUID.ValueString(),
		OrganizationGUID: data.OrganizationGUID.ValueString(),
	}

	if !data.OrganizationGUID.IsNull() {
		createServicePlanVisibility.OrganizationGUID = data.OrganizationGUID.ValueString()
	}

	createServicePlanVisibility.Metadata = &resource.Metadata{}
	labelsDiags := data.Labels.ElementsAs(ctx, &createServicePlanVisibility.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := data.Annotations.ElementsAs(ctx, &createServicePlanVisibility.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return createServicePlanVisibility, diagnostics
}

func (plan *servicePlanVisibilityType) mapServicePlanVisibilityTypeToValues(ctx context.Context) (resource.ServicePlanVisibilityUpdate, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	updateServicePlanVisibility := resource.ServicePlanVisibilityUpdate{}

	if !plan.OrganizationGUID.IsNull() {
		updateServicePlanVisibility.OrganizationGUID = plan.OrganizationGUID.ValueString()
	}

	if !plan.ServicePlanGUID.IsNull() {
		updateServicePlanVisibility.ServicePlanGUID = plan.ServicePlanGUID.ValueString()
	}

	updateServicePlanVisibility.Metadata = &resource.Metadata{}
	labelsDiags := plan.Labels.ElementsAs(ctx, &updateServicePlanVisibility.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := plan.Annotations.ElementsAs(ctx, &updateServicePlanVisibility.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return updateServicePlanVisibility, diagnostics
}
