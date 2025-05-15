package provider

import (
	"context"
	"time"

	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type spaceQuotaType struct {
	Name                  types.String `tfsdk:"name"`
	ID                    types.String `tfsdk:"id"`
	AllowPaidServicePlans types.Bool   `tfsdk:"allow_paid_service_plans"`
	TotalServices         types.Int64  `tfsdk:"total_services"`
	TotalServiceKeys      types.Int64  `tfsdk:"total_service_keys"`
	TotalRoutes           types.Int64  `tfsdk:"total_routes"`
	TotalRoutePorts       types.Int64  `tfsdk:"total_route_ports"`
	TotalMemory           types.Int64  `tfsdk:"total_memory"`
	InstanceMemory        types.Int64  `tfsdk:"instance_memory"`
	TotalAppInstances     types.Int64  `tfsdk:"total_app_instances"`
	TotalAppTasks         types.Int64  `tfsdk:"total_app_tasks"`
	TotalAppLogRateLimit  types.Int64  `tfsdk:"total_app_log_rate_limit"`
	Spaces                types.Set    `tfsdk:"spaces"`
	Org                   types.String `tfsdk:"org"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

type spaceQuotasDatasourceType struct {
	Name        types.String     `tfsdk:"name"`
	Org         types.String     `tfsdk:"org"`
	SpaceQuotas []spaceQuotaType `tfsdk:"space_quotas"`
}

func (spaceQuotaType *spaceQuotaType) mapSpaceQuotaTypeToValues(ctx context.Context) (*cfv3resource.SpaceQuotaCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	spaceQuota := cfv3resource.NewSpaceQuotaCreate(spaceQuotaType.Name.ValueString(), spaceQuotaType.Org.ValueString())
	spaceQuota.Apps = &cfv3resource.AppsQuota{}
	spaceQuota.Services = &cfv3resource.ServicesQuota{}
	spaceQuota.Routes = &cfv3resource.RoutesQuota{}
	spaceQuota.WithPaidServicesAllowed(spaceQuotaType.AllowPaidServicePlans.ValueBool())
	if !spaceQuotaType.TotalServices.IsNull() {
		spaceQuota.WithTotalServiceInstances(int(spaceQuotaType.TotalServices.ValueInt64()))
	}
	if !spaceQuotaType.TotalServiceKeys.IsNull() {
		spaceQuota.WithTotalServiceKeys(int(spaceQuotaType.TotalServiceKeys.ValueInt64()))
	}
	if !spaceQuotaType.TotalRoutes.IsNull() {
		spaceQuota.WithTotalRoutes(int(spaceQuotaType.TotalRoutes.ValueInt64()))
	}
	if !spaceQuotaType.TotalRoutePorts.IsNull() {
		spaceQuota.WithTotalReservedPorts(int(spaceQuotaType.TotalRoutePorts.ValueInt64()))
	}
	if !spaceQuotaType.TotalMemory.IsNull() {
		spaceQuota.WithTotalMemoryInMB(int(spaceQuotaType.TotalMemory.ValueInt64()))
	}
	if !spaceQuotaType.InstanceMemory.IsNull() {
		spaceQuota.WithPerProcessMemoryInMB(int(spaceQuotaType.InstanceMemory.ValueInt64()))
	}
	if !spaceQuotaType.TotalAppInstances.IsNull() {
		spaceQuota.WithTotalInstances(int(spaceQuotaType.TotalAppInstances.ValueInt64()))
	}
	if !spaceQuotaType.TotalAppTasks.IsNull() {
		spaceQuota.WithPerAppTasks(int(spaceQuotaType.TotalAppTasks.ValueInt64()))
	}
	if !spaceQuotaType.TotalAppLogRateLimit.IsNull() {
		spaceQuota.WithLogRateLimitInBytesPerSecond(int(spaceQuotaType.TotalAppLogRateLimit.ValueInt64()))
	}
	if !spaceQuotaType.Spaces.IsNull() {
		var spaceQuotaRelOrgVal []string
		spaceQuota.Relationships.Spaces = &cfv3resource.ToManyRelationships{}
		diags = spaceQuotaType.Spaces.ElementsAs(ctx, &spaceQuotaRelOrgVal, false)
		spaceQuota.WithSpaces(spaceQuotaRelOrgVal...)
	}
	return spaceQuota, diags
}
func mapSpaceQuotaValuesToType(value *cfv3resource.SpaceQuota) (spaceQuotaType, diag.Diagnostics) {
	var diags diag.Diagnostics
	spaceQuotaType := spaceQuotaType{
		Name:                  types.StringValue(value.Name),
		ID:                    types.StringValue(value.GUID),
		AllowPaidServicePlans: types.BoolValue(value.Services.PaidServicesAllowed),
		CreatedAt:             types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:             types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
	}

	if value.Services.TotalServiceInstances != nil {
		spaceQuotaType.TotalServices = types.Int64Value(int64(*value.Services.TotalServiceInstances))
	}
	if value.Services.TotalServiceKeys != nil {
		spaceQuotaType.TotalServiceKeys = types.Int64Value(int64(*value.Services.TotalServiceKeys))
	}
	if value.Routes.TotalRoutes != nil {
		spaceQuotaType.TotalRoutes = types.Int64Value(int64(*value.Routes.TotalRoutes))
	}
	if value.Routes.TotalReservedPorts != nil {
		spaceQuotaType.TotalRoutePorts = types.Int64Value(int64(*value.Routes.TotalReservedPorts))
	}
	if value.Apps.TotalMemoryInMB != nil {
		spaceQuotaType.TotalMemory = types.Int64Value(int64(*value.Apps.TotalMemoryInMB))
	}
	if value.Apps.PerProcessMemoryInMB != nil {
		spaceQuotaType.InstanceMemory = types.Int64Value(int64(*value.Apps.PerProcessMemoryInMB))
	}
	if value.Apps.TotalInstances != nil {
		spaceQuotaType.TotalAppInstances = types.Int64Value(int64(*value.Apps.TotalInstances))
	}
	if value.Apps.PerAppTasks != nil {
		spaceQuotaType.TotalAppTasks = types.Int64Value(int64(*value.Apps.PerAppTasks))
	}
	if value.Apps.LogRateLimitInBytesPerSecond != nil {
		spaceQuotaType.TotalAppLogRateLimit = types.Int64Value(int64(*value.Apps.LogRateLimitInBytesPerSecond))
	}
	spaceQuotaType.Org = types.StringValue(value.Relationships.Organization.Data.GUID)
	spaceQuotaType.Spaces, diags = setRelationshipToTFSet(value.Relationships.Spaces.Data)
	return spaceQuotaType, diags
}

func mapSpaceQuotasValuesToType(spaceQuotas []*cfv3resource.SpaceQuota) ([]spaceQuotaType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	spaceQuotasList := []spaceQuotaType{}
	for _, spaceQuota := range spaceQuotas {
		spaceQuotaValue, diags := mapSpaceQuotaValuesToType(spaceQuota)
		diagnostics.Append(diags...)
		spaceQuotasList = append(spaceQuotasList, spaceQuotaValue)
	}

	return spaceQuotasList, diagnostics
}
