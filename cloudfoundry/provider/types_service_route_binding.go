package provider

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceRouteBindingType struct {
	ID              types.String         `tfsdk:"id"`
	Route           types.String         `tfsdk:"route"`
	RouteServiceURL types.String         `tfsdk:"route_service_url"`
	ServiceInstance types.String         `tfsdk:"service_instance"`
	Parameters      jsontypes.Normalized `tfsdk:"parameters"`
	Labels          types.Map            `tfsdk:"labels"`
	Annotations     types.Map            `tfsdk:"annotations"`
	CreatedAt       types.String         `tfsdk:"created_at"`
	UpdatedAt       types.String         `tfsdk:"updated_at"`
	LastOperation   types.Object         `tfsdk:"last_operation"` //LastOperationType
}

type datasourceserviceRouteBindingType struct {
	ID              types.String `tfsdk:"id"`
	Route           types.String `tfsdk:"route"`
	RouteServiceURL types.String `tfsdk:"route_service_url"`
	ServiceInstance types.String `tfsdk:"service_instance"`
	Labels          types.Map    `tfsdk:"labels"`
	Annotations     types.Map    `tfsdk:"annotations"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	LastOperation   types.Object `tfsdk:"last_operation"` //LastOperationType
}

type datasourceserviceRouteBindingsType struct {
	ServiceInstance types.String                        `tfsdk:"service_instance"`
	Route           types.String                        `tfsdk:"route"`
	RouteBindings   []datasourceserviceRouteBindingType `tfsdk:"route_bindings"`
}

func mapServiceRouteBindingValuesToType(ctx context.Context, value *resource.ServiceRouteBinding) (serviceRouteBindingType, diag.Diagnostics) {
	var diags, diagnostics diag.Diagnostics
	serviceRouteBinding := serviceRouteBindingType{
		ID:              types.StringValue(value.GUID),
		CreatedAt:       types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:       types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
		ServiceInstance: types.StringValue(value.Relationships.ServiceInstance.Data.GUID),
		Route:           types.StringValue(value.Relationships.Route.Data.GUID),
	}

	if value.RouteServiceURL != "" {
		serviceRouteBinding.RouteServiceURL = types.StringValue(value.RouteServiceURL)
	}

	serviceRouteBinding.Labels, diags = mapMetadataValueToType(ctx, value.Metadata.Labels)
	diagnostics.Append(diags...)
	serviceRouteBinding.Annotations, diags = mapMetadataValueToType(ctx, value.Metadata.Annotations)
	diagnostics.Append(diags...)
	serviceRouteBinding.LastOperation, diags = types.ObjectValueFrom(ctx, lastOperationAttrTypes, mapLastOperation(value.LastOperation))
	diagnostics.Append(diags...)

	return serviceRouteBinding, diagnostics
}

func (data *serviceRouteBindingType) mapCreateServiceRouteBindingTypeToValues(ctx context.Context) (resource.ServiceRouteBindingCreate, diag.Diagnostics) {

	var (
		diagnostics               diag.Diagnostics
		createServiceRouteBinding *resource.ServiceRouteBindingCreate
	)
	createServiceRouteBinding = resource.NewServiceRouteBindingCreate(data.Route.ValueString(), data.ServiceInstance.ValueString())
	if !data.Parameters.IsNull() {
		j := json.RawMessage(data.Parameters.ValueString())
		createServiceRouteBinding.Parameters = &j
	}

	createServiceRouteBinding.Metadata = resource.NewMetadata()
	labelsDiags := data.Labels.ElementsAs(ctx, &createServiceRouteBinding.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := data.Annotations.ElementsAs(ctx, &createServiceRouteBinding.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createServiceRouteBinding, diagnostics
}

func (plan *serviceRouteBindingType) mapUpdateServiceRouteBindingTypeToValues(ctx context.Context, state serviceRouteBindingType) (resource.ServiceRouteBindingUpdate, diag.Diagnostics) {

	updateRouteBinding := &resource.ServiceRouteBindingUpdate{}

	var diagnostics diag.Diagnostics
	updateRouteBinding.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateRouteBinding, diagnostics
}

func (a *serviceRouteBindingType) Reduce() datasourceserviceRouteBindingType {
	var reduced datasourceserviceRouteBindingType
	copyFields(&reduced, a)
	return reduced
}

func mapDataSourceServiceRouteBindingsValuesToType(ctx context.Context, svcRouteBindings []*resource.ServiceRouteBinding) ([]datasourceserviceRouteBindingType, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	svcRouteBindingsList := []datasourceserviceRouteBindingType{}
	for _, svcRouteBinding := range svcRouteBindings {
		bindingValue, diags := mapServiceRouteBindingValuesToType(ctx, svcRouteBinding)
		diagnostics.Append(diags...)
		svcRouteBindingsList = append(svcRouteBindingsList, bindingValue.Reduce())
	}

	return svcRouteBindingsList, diagnostics
}
