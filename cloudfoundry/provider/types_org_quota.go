package provider

import (
	"context"
	"time"

	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OrgQuotaType struct {
	Name                  types.String `tfsdk:"name"`
	ID                    types.String `tfsdk:"id"`
	AllowPaidServicePlans types.Bool   `tfsdk:"allow_paid_service_plans"`
	TotalServices         types.Int64  `tfsdk:"total_services"`
	TotalServiceKeys      types.Int64  `tfsdk:"total_service_keys"`
	TotalRoutes           types.Int64  `tfsdk:"total_routes"`
	TotalRoutePorts       types.Int64  `tfsdk:"total_route_ports"`
	TotalPrivateDomains   types.Int64  `tfsdk:"total_private_domains"`
	TotalMemory           types.Int64  `tfsdk:"total_memory"`
	InstanceMemory        types.Int64  `tfsdk:"instance_memory"`
	TotalAppInstances     types.Int64  `tfsdk:"total_app_instances"`
	TotalAppTasks         types.Int64  `tfsdk:"total_app_tasks"`
	TotalAppLogRateLimit  types.Int64  `tfsdk:"total_app_log_rate_limit"`
	Organizations         types.Set    `tfsdk:"orgs"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

type orgQuotasDatasourceType struct {
	Org       types.String   `tfsdk:"org"`
	OrgQuotas []OrgQuotaType `tfsdk:"org_quotas"`
}

func (orgQuotaType *OrgQuotaType) mapOrgQuotaTypeToValues(ctx context.Context) (*cfv3resource.OrganizationQuotaCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	orgQuota := cfv3resource.NewOrganizationQuotaCreate(orgQuotaType.Name.ValueString())
	orgQuota.Domains = &cfv3resource.DomainsQuota{}
	orgQuota.Services = &cfv3resource.ServicesQuota{}
	orgQuota.Apps = &cfv3resource.AppsQuota{}
	orgQuota.Routes = &cfv3resource.RoutesQuota{}
	orgQuota.WithPaidServicesAllowed(orgQuotaType.AllowPaidServicePlans.ValueBool())
	if !orgQuotaType.TotalServices.IsNull() {
		orgQuota.WithTotalServiceInstances(int(orgQuotaType.TotalServices.ValueInt64()))
	}
	if !orgQuotaType.TotalServiceKeys.IsNull() {
		orgQuota.WithTotalServiceKeys(int(orgQuotaType.TotalServiceKeys.ValueInt64()))
	}
	if !orgQuotaType.TotalRoutes.IsNull() {
		orgQuota.WithTotalRoutes(int(orgQuotaType.TotalRoutes.ValueInt64()))
	}
	if !orgQuotaType.TotalRoutePorts.IsNull() {
		orgQuota.WithTotalReservedPorts(int(orgQuotaType.TotalRoutePorts.ValueInt64()))
	}
	if !orgQuotaType.TotalPrivateDomains.IsNull() {
		orgQuota.WithDomains(int(orgQuotaType.TotalPrivateDomains.ValueInt64()))
	}
	if !orgQuotaType.TotalMemory.IsNull() {
		orgQuota.WithAppsTotalMemoryInMB(int(orgQuotaType.TotalMemory.ValueInt64()))
	}
	if !orgQuotaType.InstanceMemory.IsNull() {
		orgQuota.WithPerProcessMemoryInMB(int(orgQuotaType.InstanceMemory.ValueInt64()))
	}
	if !orgQuotaType.TotalAppInstances.IsNull() {
		orgQuota.WithTotalInstances(int(orgQuotaType.TotalAppInstances.ValueInt64()))
	}
	if !orgQuotaType.TotalAppTasks.IsNull() {
		orgQuota.WithPerAppTasks(int(orgQuotaType.TotalAppTasks.ValueInt64()))
	}
	if !orgQuotaType.TotalAppLogRateLimit.IsNull() {
		orgQuota.WithLogRateLimitInBytesPerSecond(int(orgQuotaType.TotalAppLogRateLimit.ValueInt64()))
	}
	if !orgQuotaType.Organizations.IsNull() {
		var orgQuotaRelOrgVal []string
		diags = orgQuotaType.Organizations.ElementsAs(ctx, &orgQuotaRelOrgVal, false)
		orgQuota.WithOrganizations(orgQuotaRelOrgVal...)
	}
	return orgQuota, diags
}
func mapOrgQuotaValuesToType(value *cfv3resource.OrganizationQuota) (OrgQuotaType, diag.Diagnostics) {
	var diags diag.Diagnostics
	orgQuotaType := OrgQuotaType{
		Name:                  types.StringValue(value.Name),
		ID:                    types.StringValue(value.GUID),
		AllowPaidServicePlans: types.BoolValue(value.Services.PaidServicesAllowed),
		CreatedAt:             types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:             types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
	}

	if value.Services.TotalServiceInstances != nil {
		orgQuotaType.TotalServices = types.Int64Value(int64(*value.Services.TotalServiceInstances))
	}
	if value.Services.TotalServiceKeys != nil {
		orgQuotaType.TotalServiceKeys = types.Int64Value(int64(*value.Services.TotalServiceKeys))
	}
	if value.Routes.TotalRoutes != nil {
		orgQuotaType.TotalRoutes = types.Int64Value(int64(*value.Routes.TotalRoutes))
	}
	if value.Routes.TotalReservedPorts != nil {
		orgQuotaType.TotalRoutePorts = types.Int64Value(int64(*value.Routes.TotalReservedPorts))
	}
	if value.Domains.TotalDomains != nil {
		orgQuotaType.TotalPrivateDomains = types.Int64Value(int64(*value.Domains.TotalDomains))
	}
	if value.Apps.TotalMemoryInMB != nil {
		orgQuotaType.TotalMemory = types.Int64Value(int64(*value.Apps.TotalMemoryInMB))
	}
	if value.Apps.PerProcessMemoryInMB != nil {
		orgQuotaType.InstanceMemory = types.Int64Value(int64(*value.Apps.PerProcessMemoryInMB))
	}
	if value.Apps.TotalInstances != nil {
		orgQuotaType.TotalAppInstances = types.Int64Value(int64(*value.Apps.TotalInstances))
	}
	if value.Apps.PerAppTasks != nil {
		orgQuotaType.TotalAppTasks = types.Int64Value(int64(*value.Apps.PerAppTasks))
	}
	if value.Apps.LogRateLimitInBytesPerSecond != nil {
		orgQuotaType.TotalAppLogRateLimit = types.Int64Value(int64(*value.Apps.LogRateLimitInBytesPerSecond))
	}
	orgQuotaType.Organizations, diags = setRelationshipToTFSet(value.Relationships.Organizations.Data)
	return orgQuotaType, diags
}

func mapOrgQuotasValuesToType(orgQuotas []*cfv3resource.OrganizationQuota) ([]OrgQuotaType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	orgQuotasList := []OrgQuotaType{}
	for _, orgQuota := range orgQuotas {
		orgQuotaValue, diags := mapOrgQuotaValuesToType(orgQuota)
		diagnostics.Append(diags...)
		orgQuotasList = append(orgQuotasList, orgQuotaValue)
	}

	return orgQuotasList, diagnostics
}
