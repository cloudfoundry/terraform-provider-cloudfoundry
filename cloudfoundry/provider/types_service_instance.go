package provider

import (
	"context"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"time"
)

type serviceInstanceType struct {
	Name             types.String         `tfsdk:"name"`
	ID               types.String         `tfsdk:"id"`
	Type             types.String         `tfsdk:"type"`
	Space            types.String         `tfsdk:"space"`
	ServicePlan      types.String         `tfsdk:"service_plan"`
	Parameters       jsontypes.Normalized `tfsdk:"parameters"`
	LastOperation    types.Object         `tfsdk:"last_operation"` //LastOperationType
	Tags             types.List           `tfsdk:"tags"`
	DashboardURL     types.String         `tfsdk:"dashboard_url"`
	Credentials      jsontypes.Normalized `tfsdk:"credentials"`
	SyslogDrainURL   types.String         `tfsdk:"syslog_drain_url"`
	RouteServiceURL  types.String         `tfsdk:"route_service_url"`
	MaintenanceInfo  types.Object         `tfsdk:"maintenance_info"` //maintenanceInfoType
	UpgradeAvailable types.Bool           `tfsdk:"upgrade_available"`
	Labels           types.Map            `tfsdk:"labels"`
	Annotations      types.Map            `tfsdk:"annotations"`
	CreatedAt        types.String         `tfsdk:"created_at"`
	UpdatedAt        types.String         `tfsdk:"updated_at"`
	Timeouts         timeouts.Value       `tfsdk:"timeouts"`
}

type singleDatasourceServiceInstanceType struct {
	Name             types.String         `tfsdk:"name"`
	ID               types.String         `tfsdk:"id"`
	Type             types.String         `tfsdk:"type"`
	Space            types.String         `tfsdk:"space"`
	ServicePlan      types.String         `tfsdk:"service_plan"`
	Parameters       jsontypes.Normalized `tfsdk:"parameters"`
	LastOperation    types.Object         `tfsdk:"last_operation"` //LastOperationType
	Tags             types.List           `tfsdk:"tags"`
	DashboardURL     types.String         `tfsdk:"dashboard_url"`
	Credentials      jsontypes.Normalized `tfsdk:"credentials"`
	SyslogDrainURL   types.String         `tfsdk:"syslog_drain_url"`
	RouteServiceURL  types.String         `tfsdk:"route_service_url"`
	MaintenanceInfo  types.Object         `tfsdk:"maintenance_info"` //maintenanceInfoType
	UpgradeAvailable types.Bool           `tfsdk:"upgrade_available"`
	Labels           types.Map            `tfsdk:"labels"`
	Annotations      types.Map            `tfsdk:"annotations"`
	CreatedAt        types.String         `tfsdk:"created_at"`
	UpdatedAt        types.String         `tfsdk:"updated_at"`
}

type datasourceServiceInstanceType struct {
	Name             types.String `tfsdk:"name"`
	ID               types.String `tfsdk:"id"`
	Type             types.String `tfsdk:"type"`
	Space            types.String `tfsdk:"space"`
	ServicePlan      types.String `tfsdk:"service_plan"`
	LastOperation    types.Object `tfsdk:"last_operation"` //LastOperationType
	Tags             types.List   `tfsdk:"tags"`
	DashboardURL     types.String `tfsdk:"dashboard_url"`
	SyslogDrainURL   types.String `tfsdk:"syslog_drain_url"`
	RouteServiceURL  types.String `tfsdk:"route_service_url"`
	MaintenanceInfo  types.Object `tfsdk:"maintenance_info"` //maintenanceInfoType
	UpgradeAvailable types.Bool   `tfsdk:"upgrade_available"`
	Labels           types.Map    `tfsdk:"labels"`
	Annotations      types.Map    `tfsdk:"annotations"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

type datasourceServiceInstancesType struct {
	Org              types.String                    `tfsdk:"org"`
	Space            types.String                    `tfsdk:"space"`
	Name             types.String                    `tfsdk:"name"`
	ServiceInstances []datasourceServiceInstanceType `tfsdk:"service_instances"`
}

type lastOperationType struct {
	Type        types.String `tfsdk:"type"`
	State       types.String `tfsdk:"state"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type maintenanceInfoType struct {
	Version     types.String `tfsdk:"version"`
	Description types.String `tfsdk:"description"`
}

type ServiceInstanceSharingType struct {
	ServiceInstance types.String `tfsdk:"service_instance"`
	Spaces          types.Set    `tfsdk:"spaces"`
}

var maintenanceInfoAttrTypes = map[string]attr.Type{
	"version":     types.StringType,
	"description": types.StringType,
}

var lastOperationAttrTypes = map[string]attr.Type{
	"type":        types.StringType,
	"state":       types.StringType,
	"description": types.StringType,
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func (a *serviceInstanceType) Reduce() singleDatasourceServiceInstanceType {
	var reduced singleDatasourceServiceInstanceType
	copyFields(&reduced, a)
	return reduced
}

func mapDataSourceServiceInstanceValuesToType(ctx context.Context, value *resource.ServiceInstance) (datasourceServiceInstanceType, diag.Diagnostics) {
	var diags, diagnostics diag.Diagnostics
	dsServiceInstanceType := datasourceServiceInstanceType{
		Name:      types.StringValue(value.Name),
		ID:        types.StringValue(value.GUID),
		Type:      types.StringValue(value.Type),
		Space:     types.StringValue(value.Relationships.Space.Data.GUID),
		CreatedAt: types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
	}
	if value.UpgradeAvailable != nil {
		dsServiceInstanceType.UpgradeAvailable = types.BoolValue(*value.UpgradeAvailable)
	}
	switch value.Type {
	case managedSerivceInstance:
		dsServiceInstanceType.ServicePlan = types.StringValue(value.Relationships.ServicePlan.Data.GUID)
		if value.DashboardURL != nil {
			dsServiceInstanceType.DashboardURL = types.StringValue(*value.DashboardURL)
		}
	case userProvidedServiceInstance:
		if value.SyslogDrainURL != nil {
			dsServiceInstanceType.SyslogDrainURL = types.StringValue(*value.SyslogDrainURL)
		}
		if value.RouteServiceURL != nil {
			dsServiceInstanceType.RouteServiceURL = types.StringValue(*value.RouteServiceURL)
		}
	}
	dsServiceInstanceType.Labels, diags = mapMetadataValueToType(ctx, value.Metadata.Labels)
	diagnostics.Append(diags...)
	dsServiceInstanceType.Annotations, diags = mapMetadataValueToType(ctx, value.Metadata.Annotations)
	diagnostics.Append(diags...)
	if value.MaintenanceInfo != nil {
		dsServiceInstanceType.MaintenanceInfo, diags = types.ObjectValueFrom(ctx, maintenanceInfoAttrTypes, mapMaintenanceInfo(*value.MaintenanceInfo))
		diagnostics.Append(diags...)
	} else {
		dsServiceInstanceType.MaintenanceInfo = types.ObjectNull(maintenanceInfoAttrTypes)

	}
	dsServiceInstanceType.LastOperation, diags = types.ObjectValueFrom(ctx, lastOperationAttrTypes, mapLastOperation(value.LastOperation))
	diagnostics.Append(diags...)
	//tags mapping
	if len(value.Tags) > 0 {
		tags := make([]types.String, 0, len(value.Tags))
		for _, t := range value.Tags {
			tags = append(tags, types.StringValue(t))
		}
		dsServiceInstanceType.Tags, diags = types.ListValueFrom(ctx, types.StringType, tags)
		diagnostics.Append(diags...)
	} else {
		dsServiceInstanceType.Tags = types.ListNull(types.StringType)

	}

	return dsServiceInstanceType, diagnostics
}

func mapDataSourceServiceInstancesValuesToType(ctx context.Context, svcInstances []*resource.ServiceInstance) ([]datasourceServiceInstanceType, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	svcInstancesList := []datasourceServiceInstanceType{}
	for _, svcInstance := range svcInstances {
		svcInstanceValue, diags := mapDataSourceServiceInstanceValuesToType(ctx, svcInstance)
		diagnostics.Append(diags...)
		svcInstancesList = append(svcInstancesList, svcInstanceValue)
	}

	return svcInstancesList, diagnostics

}

func mapResourceServiceInstanceValuesToType(ctx context.Context, value *resource.ServiceInstance, paramCreds jsontypes.Normalized) (serviceInstanceType, diag.Diagnostics) {
	var diagnostics, diags diag.Diagnostics
	serviceInstanceType := serviceInstanceType{
		Name:      types.StringValue(value.Name),
		ID:        types.StringValue(value.GUID),
		Type:      types.StringValue(value.Type),
		Space:     types.StringValue(value.Relationships.Space.Data.GUID),
		CreatedAt: types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
	}
	if value.UpgradeAvailable != nil {
		serviceInstanceType.UpgradeAvailable = types.BoolValue(*value.UpgradeAvailable)
	}
	switch value.Type {
	case managedSerivceInstance:

		serviceInstanceType.ServicePlan = types.StringValue(value.Relationships.ServicePlan.Data.GUID)
		if value.DashboardURL != nil {
			serviceInstanceType.DashboardURL = types.StringValue(*value.DashboardURL)
		}
		serviceInstanceType.MaintenanceInfo, diags = types.ObjectValueFrom(ctx, maintenanceInfoAttrTypes, mapMaintenanceInfo(*value.MaintenanceInfo))
		diagnostics.Append(diags...)

		if !paramCreds.IsNull() {
			serviceInstanceType.Parameters = jsontypes.NewNormalizedValue(paramCreds.ValueString())
		} else {
			serviceInstanceType.Parameters = jsontypes.NewNormalizedNull()
		}
	case userProvidedServiceInstance:
		if value.SyslogDrainURL != nil {
			serviceInstanceType.SyslogDrainURL = types.StringValue(*value.SyslogDrainURL)
		}
		if value.RouteServiceURL != nil {
			serviceInstanceType.RouteServiceURL = types.StringValue(*value.RouteServiceURL)
		}
		serviceInstanceType.MaintenanceInfo = types.ObjectNull(maintenanceInfoAttrTypes)
		if !paramCreds.IsNull() {
			serviceInstanceType.Credentials = jsontypes.NewNormalizedValue(paramCreds.ValueString())
		} else {
			serviceInstanceType.Credentials = jsontypes.NewNormalizedNull()
		}
	}
	serviceInstanceType.Labels, diags = mapMetadataValueToType(ctx, value.Metadata.Labels)
	diagnostics.Append(diags...)
	serviceInstanceType.Annotations, diags = mapMetadataValueToType(ctx, value.Metadata.Annotations)
	diagnostics.Append(diags...)
	//tags mapping
	if len(value.Tags) > 0 {
		tags := make([]types.String, 0, len(value.Tags))
		for _, t := range value.Tags {
			tags = append(tags, types.StringValue(t))
		}
		serviceInstanceType.Tags, diags = types.ListValueFrom(ctx, types.StringType, tags)
		diagnostics.Append(diags...)
	} else {
		serviceInstanceType.Tags = types.ListNull(types.StringType)

	}

	serviceInstanceType.LastOperation, diags = types.ObjectValueFrom(ctx, lastOperationAttrTypes, mapLastOperation(value.LastOperation))
	diagnostics.Append(diags...)

	return serviceInstanceType, diagnostics
}

func mapLastOperation(value resource.LastOperation) lastOperationType {
	var lastOps lastOperationType
	lastOps.Type = types.StringValue(value.Type)
	lastOps.State = types.StringValue(value.State)
	lastOps.Description = types.StringValue(value.Description)
	lastOps.CreatedAt = types.StringValue(value.CreatedAt.Format(time.RFC3339))
	lastOps.UpdatedAt = types.StringValue(value.UpdatedAt.Format(time.RFC3339))
	return lastOps
}

func mapMaintenanceInfo(value resource.ServiceInstanceMaintenanceInfo) maintenanceInfoType {
	var maintenance maintenanceInfoType
	if value.Version != "" && value.Description != "" {
		maintenance.Version = types.StringValue(value.Version)
		maintenance.Description = types.StringValue(value.Description)
	}

	return maintenance

}

// toTagsList converts aliases of type types.Set into a slice of strings.
func toTagsList(ctx context.Context, tagsSet types.List) ([]string, diag.Diagnostics) {

	tags := make([]string, 0, len(tagsSet.Elements()))
	diags := tagsSet.ElementsAs(ctx, &tags, false)
	return tags, diags
}
