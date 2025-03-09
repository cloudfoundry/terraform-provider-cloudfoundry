package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type servicePlanType struct {
	Name            types.String `tfsdk:"name"`
	ID              types.String `tfsdk:"id"`
	VisibilityType  types.String `tfsdk:"visibility_type"`
	Available       types.Bool   `tfsdk:"available"`
	Free            types.Bool   `tfsdk:"free"`
	ServiceOffering types.String `tfsdk:"service_offering"`
	Description     types.String `tfsdk:"description"`
	Costs           types.List   `tfsdk:"costs"`            //List of costInfoType
	BrokerCatalog   types.Object `tfsdk:"broker_catalog"`   //BrokerCatalogType
	MaintenanceInfo types.Object `tfsdk:"maintenance_info"` //maintenanceInfoType
	Schemas         types.Object `tfsdk:"schemas"`
	Labels          types.Map    `tfsdk:"labels"`
	Annotations     types.Map    `tfsdk:"annotations"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

type datasourceServicePlanType struct {
	Name                types.String `tfsdk:"name"`
	ServiceOfferingName types.String `tfsdk:"service_offering_name"`
	ServiceBrokerName   types.String `tfsdk:"service_broker_name"`
	ID                  types.String `tfsdk:"id"`
	VisibilityType      types.String `tfsdk:"visibility_type"`
	Available           types.Bool   `tfsdk:"available"`
	Free                types.Bool   `tfsdk:"free"`
	Description         types.String `tfsdk:"description"`
	Costs               types.List   `tfsdk:"costs"`            //List of costInfoType
	BrokerCatalog       types.Object `tfsdk:"broker_catalog"`   //BrokerCatalogType
	MaintenanceInfo     types.Object `tfsdk:"maintenance_info"` //maintenanceInfoType
	Schemas             types.Object `tfsdk:"schemas"`
	Labels              types.Map    `tfsdk:"labels"`
	Annotations         types.Map    `tfsdk:"annotations"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

type datasourceServicePlansType struct {
	Name                types.String `tfsdk:"name"`
	ServiceOfferingName types.String `tfsdk:"service_offering_name"`
	ServiceBrokerName   types.String `tfsdk:"service_broker_name"`
	ServicePlans        types.List   `tfsdk:"service_plans"` //List of servicePlanType
}

type costType struct {
	Amount   types.Float64 `tfsdk:"amount"`
	Currency types.String  `tfsdk:"currency"`
	Unit     types.String  `tfsdk:"unit"`
}

type brokerCatalogType struct {
	Id                     types.String         `tfsdk:"id"`
	Metadata               jsontypes.Normalized `tfsdk:"metadata"`
	MaximumPollingDuration types.Float64        `tfsdk:"maximum_polling_duration"`
	PlanUpdateable         types.Bool           `tfsdk:"plan_updateable"`
	Bindable               types.Bool           `tfsdk:"bindable"`
}

type svcPlanSchemaType struct {
	ServiceInstance types.Object `tfsdk:"service_instance"`
	ServiceBinding  types.Object `tfsdk:"service_binding"`
}

type svcInstanceParamsType struct {
	Createparameters jsontypes.Normalized `tfsdk:"create_parameters"`
	Updateparameters jsontypes.Normalized `tfsdk:"update_parameters"`
}

type svcBindingParamsType struct {
	Createparameters jsontypes.Normalized `tfsdk:"create_parameters"`
}

var costAttrTypes = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"amount":   types.Float64Type,
		"currency": types.StringType,
		"unit":     types.StringType,
	},
}

var brokerCatalogAttrTypes = map[string]attr.Type{
	"id":                       types.StringType,
	"metadata":                 types.StringType,
	"maximum_polling_duration": types.Float64Type,
	"plan_updateable":          types.BoolType,
	"bindable":                 types.BoolType,
}

var svcBindingParamsAttrTypes = map[string]attr.Type{
	"create_parameters": types.StringType,
}

var svcInstaParamsAttrTypes = map[string]attr.Type{
	"create_parameters": types.StringType,
	"update_parameters": types.StringType,
}

var svcPlanSchemaAttrTypes = map[string]attr.Type{
	"service_instance": types.ObjectType{
		AttrTypes: svcInstaParamsAttrTypes,
	},
	"service_binding": types.ObjectType{
		AttrTypes: svcBindingParamsAttrTypes,
	},
}

var servicePlanAttrType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":             types.StringType,
		"id":               types.StringType,
		"visibility_type":  types.StringType,
		"available":        types.BoolType,
		"free":             types.BoolType,
		"service_offering": types.StringType,
		"description":      types.StringType,
		"costs": types.ListType{
			ElemType: costAttrTypes,
		},
		"broker_catalog": types.ObjectType{
			AttrTypes: brokerCatalogAttrTypes,
		},
		"maintenance_info": types.ObjectType{
			AttrTypes: maintenanceInfoAttrTypes,
		},
		"schemas": types.ObjectType{
			AttrTypes: svcPlanSchemaAttrTypes,
		},
		"created_at": types.StringType,
		"updated_at": types.StringType,
		"labels": types.MapType{
			ElemType: types.StringType,
		},
		"annotations": types.MapType{
			ElemType: types.StringType,
		},
	},
}

// Prepares a terraform list from the service plan resources returned by the cf-client.
func mapServicePlansValuesToListType(ctx context.Context, svcPlans []*resource.ServicePlan) (types.List, diag.Diagnostics) {

	var diags, diagnostics diag.Diagnostics
	svcPlanValues := []servicePlanType{}
	for _, svcPlan := range svcPlans {
		svcPlanValue, diags := mapServicePlanValuesToType(ctx, *svcPlan)
		diagnostics.Append(diags...)
		svcPlanValues = append(svcPlanValues, svcPlanValue)
	}

	svcPlanList, diags := types.ListValueFrom(ctx, servicePlanAttrType, svcPlanValues)
	diagnostics.Append(diags...)

	return svcPlanList, diagnostics
}

func mapServicePlansValueToData(ctx context.Context, svcPlan *resource.ServicePlan) (datasourceServicePlanType, diag.Diagnostics) {

	var diags, diagnostics diag.Diagnostics

	svcPlanValue, diags := mapServicePlanValuesToType(ctx, *svcPlan)
	dsSvcPlan := mapServicePlanValueToDataSourceType(svcPlanValue)

	diagnostics.Append(diags...)
	return dsSvcPlan, diagnostics

}

func mapServicePlanValueToDataSourceType(svcPlanValue servicePlanType) datasourceServicePlanType {
	dsSvcPlan := datasourceServicePlanType{
		Name:                svcPlanValue.Name,
		ID:                  svcPlanValue.ID,
		VisibilityType:      svcPlanValue.VisibilityType,
		Available:           svcPlanValue.Available,
		Free:                svcPlanValue.Free,
		ServiceOfferingName: svcPlanValue.ServiceOffering,
		Description:         svcPlanValue.Description,
		Costs:               svcPlanValue.Costs,
		BrokerCatalog:       svcPlanValue.BrokerCatalog,
		MaintenanceInfo:     svcPlanValue.MaintenanceInfo,
		Schemas:             svcPlanValue.Schemas,
		Labels:              svcPlanValue.Labels,
		Annotations:         svcPlanValue.Annotations,
		CreatedAt:           svcPlanValue.CreatedAt,
		UpdatedAt:           svcPlanValue.UpdatedAt,
	}
	return dsSvcPlan
}

func mapServicePlanValuesToType(ctx context.Context, svcPlan resource.ServicePlan) (servicePlanType, diag.Diagnostics) {
	svcPlanType := servicePlanType{
		Name:            types.StringValue(svcPlan.Name),
		ID:              types.StringValue(svcPlan.GUID),
		VisibilityType:  types.StringValue(svcPlan.VisibilityType),
		Available:       types.BoolValue(svcPlan.Available),
		Free:            types.BoolValue(svcPlan.Free),
		ServiceOffering: types.StringValue(svcPlan.Relationships.ServiceOffering.Data.GUID),
		Description:     types.StringValue(svcPlan.Description),
		CreatedAt:       types.StringValue(svcPlan.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:       types.StringValue(svcPlan.UpdatedAt.Format(time.RFC3339)),
	}

	costValues := []costType{}
	for _, cost := range svcPlan.Costs {
		costValue := mapCost(cost)
		costValues = append(costValues, costValue)
	}

	var diags, diagnostics diag.Diagnostics

	svcPlanType.Labels, diags = mapMetadataValueToType(ctx, svcPlan.Metadata.Labels)
	diagnostics.Append(diags...)
	svcPlanType.Annotations, diags = mapMetadataValueToType(ctx, svcPlan.Metadata.Annotations)
	diagnostics.Append(diags...)
	svcPlanType.Costs, diags = types.ListValueFrom(ctx, costAttrTypes, costValues)
	diagnostics.Append(diags...)
	svcPlanType.BrokerCatalog, diags = types.ObjectValueFrom(ctx, brokerCatalogAttrTypes, mapBrokerCatalog(svcPlan.BrokerCatalog))
	diagnostics.Append(diags...)
	svcPlanType.MaintenanceInfo, diags = types.ObjectValueFrom(ctx, maintenanceInfoAttrTypes, mapServicePlanMaintenanceInfo(svcPlan.MaintenanceInfo))

	diagnostics.Append(diags...)
	svcPlanSchemaObject, diags := mapServiceSchemas(ctx, svcPlan.Schemas)
	diagnostics.Append(diags...)
	svcPlanType.Schemas, diags = types.ObjectValueFrom(ctx, svcPlanSchemaAttrTypes, svcPlanSchemaObject)
	diagnostics.Append(diags...)

	return svcPlanType, diagnostics
}

func mapCost(value resource.ServicePlanCosts) costType {
	var cost costType
	cost.Currency = types.StringValue(value.Currency)
	cost.Amount = types.Float64Value(value.Amount)
	cost.Unit = types.StringValue(value.Unit)
	return cost
}

func mapBrokerCatalog(value resource.ServicePlanBrokerCatalog) brokerCatalogType {
	var brokerCatalog brokerCatalogType
	brokerCatalog.Bindable = types.BoolValue(value.Features.Bindable)
	brokerCatalog.Id = types.StringValue(value.ID)
	brokerCatalog.Metadata = jsontypes.NewNormalizedValue(string(*value.Metadata))
	brokerCatalog.PlanUpdateable = types.BoolValue(value.Features.PlanUpdateable)
	if value.MaximumPollingDuration != nil {
		brokerCatalog.MaximumPollingDuration = types.Float64Value(float64(*value.MaximumPollingDuration))
	}
	return brokerCatalog
}

func mapServicePlanMaintenanceInfo(value resource.ServicePlanMaintenanceInfo) maintenanceInfoType {
	var maintenance maintenanceInfoType
	if value.Version != "" && value.Description != "" {
		maintenance.Version = types.StringValue(value.Version)
		maintenance.Description = types.StringValue(value.Description)
	}
	return maintenance
}

func mapServiceSchemas(ctx context.Context, value resource.ServicePlanSchemas) (svcPlanSchemaType, diag.Diagnostics) {
	var (
		svcPlanSchema      svcPlanSchemaType
		svcInstParams      svcInstanceParamsType
		svcBindingParams   svcBindingParamsType
		diags, diagnostics diag.Diagnostics
	)

	svcBindingParamsByte, _ := value.ServiceBinding.Create.Parameters.MarshalJSON()
	svcBindingParams.Createparameters = jsontypes.NewNormalizedValue(string(svcBindingParamsByte))

	svcInstParamsCreateByte, _ := value.ServiceInstance.Create.Parameters.MarshalJSON()
	svcInstParams.Createparameters = jsontypes.NewNormalizedValue(string(svcInstParamsCreateByte))

	svcInstParamsUpdateByte, _ := value.ServiceInstance.Update.Parameters.MarshalJSON()
	svcInstParams.Updateparameters = jsontypes.NewNormalizedValue(string(svcInstParamsUpdateByte))

	svcPlanSchema.ServiceBinding, diags = types.ObjectValueFrom(ctx, svcBindingParamsAttrTypes, svcBindingParams)
	diagnostics.Append(diags...)
	svcPlanSchema.ServiceInstance, diags = types.ObjectValueFrom(ctx, svcInstaParamsAttrTypes, svcInstParams)
	diagnostics.Append(diags...)
	return svcPlanSchema, diagnostics
}
