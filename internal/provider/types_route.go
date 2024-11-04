package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type routeType struct {
	Protocol     types.String `tfsdk:"protocol"`
	Host         types.String `tfsdk:"host"`
	Path         types.String `tfsdk:"path"`
	Port         types.Int64  `tfsdk:"port"`
	Url          types.String `tfsdk:"url"`
	Id           types.String `tfsdk:"id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	Destinations types.Set    `tfsdk:"destinations"`
	Space        types.String `tfsdk:"space"`
	Domain       types.String `tfsdk:"domain"`
	Labels       types.Map    `tfsdk:"labels"`
	Annotations  types.Map    `tfsdk:"annotations"`
}

type datasourceRouteType struct {
	Space  types.String `tfsdk:"space"`
	Domain types.String `tfsdk:"domain"`
	Org    types.String `tfsdk:"org"`
	Host   types.String `tfsdk:"host"`
	Path   types.String `tfsdk:"path"`
	Port   types.Int64  `tfsdk:"port"`
	Routes []routeType  `tfsdk:"routes"`
}

type destinationType struct {
	Id             types.String `tfsdk:"id"`
	AppId          types.String `tfsdk:"app_id"`
	AppProcessType types.String `tfsdk:"app_process_type"`
	Port           types.Int64  `tfsdk:"port"`
	Weight         types.Int64  `tfsdk:"weight"`
	Protocol       types.String `tfsdk:"protocol"`
}

var destinationObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":               types.StringType,
		"app_id":           types.StringType,
		"app_process_type": types.StringType,
		"port":             types.Int64Type,
		"weight":           types.Int64Type,
		"protocol":         types.StringType,
	},
}

// Sets the terraform struct values from the route resources returned by the cf-client.
func mapRoutesValuesToType(ctx context.Context, data datasourceRouteType, routes []*resource.Route) (datasourceRouteType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	routesType := datasourceRouteType{
		Space:  data.Space,
		Domain: data.Domain,
		Host:   data.Host,
		Path:   data.Path,
		Port:   data.Port,
	}

	routesType.Routes = []routeType{}
	for _, route := range routes {
		routeValue, diags := mapRouteValuesToType(ctx, route)
		diagnostics.Append(diags...)
		routesType.Routes = append(routesType.Routes, routeValue)
	}

	return routesType, diagnostics
}

// Sets the terraform struct values from the route resource returned by the cf-client.
func mapRouteValuesToType(ctx context.Context, route *resource.Route) (routeType, diag.Diagnostics) {

	routeType := routeType{

		CreatedAt: types.StringValue(route.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(route.UpdatedAt.Format(time.RFC3339)),
		Id:        types.StringValue(route.GUID),
		Protocol:  types.StringValue(route.Protocol),
		Host:      types.StringValue(route.Host),
		Space:     types.StringValue(route.Relationships.Space.Data.GUID),
		Domain:    types.StringValue(route.Relationships.Domain.Data.GUID),
		Url:       types.StringValue(route.URL),
	}

	if route.Port != nil {
		routeType.Port = types.Int64Value(int64(*route.Port))
	}

	if route.Path != "" {
		routeType.Path = types.StringValue(route.Path)
	}

	if route.Host == "" {
		routeType.Host = basetypes.NewStringNull()
	}

	var diags, diagnostics diag.Diagnostics
	routeType.Labels, diags = mapMetadataValueToType(ctx, route.Metadata.Labels)
	diagnostics.Append(diags...)
	routeType.Annotations, diags = mapMetadataValueToType(ctx, route.Metadata.Annotations)
	diagnostics.Append(diags...)

	if len(route.Destinations) == 0 {
		routeType.Destinations = types.SetNull(destinationObjType)
	} else {
		routeType.Destinations, diags = mapDestinationValuesToSetType(ctx, &route.Destinations)
		diagnostics.Append(diags...)
	}
	return routeType, diagnostics
}

// Sets the terraform struct values from the destination resource returned by the cf-client.
func mapDestinationValuesToType(destination resource.RouteDestination) destinationType {

	destinationType := destinationType{
		Id:             types.StringValue(*destination.GUID),
		AppId:          types.StringValue(*destination.App.GUID),
		AppProcessType: types.StringValue(destination.App.Process.Type),
		Port:           types.Int64Value(int64(*destination.Port)),
	}

	if destination.Weight != nil {
		destinationType.Weight = types.Int64Value(int64(*destination.Weight))
	}
	if destination.Protocol != nil {
		destinationType.Protocol = types.StringValue(*destination.Protocol)
	}

	return destinationType
}

// Prepares a terraform set from the destination resources returned by the cf-client.
func mapDestinationValuesToSetType(ctx context.Context, destinations *[]resource.RouteDestination) (types.Set, diag.Diagnostics) {

	var diags, diagnostics diag.Diagnostics
	destinationValues := []destinationType{}
	for _, destination := range *destinations {
		destinationValue := mapDestinationValuesToType(destination)
		destinationValues = append(destinationValues, destinationValue)
	}

	destinationSet, diags := types.SetValueFrom(ctx, destinationObjType, destinationValues)
	diagnostics.Append(diags...)

	return destinationSet, diagnostics
}

// Sets the route list options for reading with cf-client from the terraform struct values.
func (data *datasourceRouteType) mapReadRouteTypeToValues() client.RouteListOptions {

	routeListOptions := client.NewRouteListOptions()

	if !data.Domain.IsNull() {
		routeListOptions.DomainGUIDs = client.Filter{Values: []string{data.Domain.ValueString()}}
	}
	if !data.Space.IsNull() {
		routeListOptions.SpaceGUIDs = client.Filter{Values: []string{data.Space.ValueString()}}
	}
	if !data.Path.IsNull() {
		routeListOptions.Paths = client.Filter{Values: []string{data.Path.ValueString()}}
	}
	if !data.Host.IsNull() {
		routeListOptions.Hosts = client.Filter{Values: []string{data.Host.ValueString()}}
	}
	if !data.Port.IsNull() {
		routeListOptions.Ports = client.Filter{Values: []string{fmt.Sprintf("%d", (data.Port.ValueInt64()))}}
	}
	if !data.Org.IsNull() {
		routeListOptions.OrganizationGUIDs = client.Filter{Values: []string{data.Org.ValueString()}}
	}

	return *routeListOptions
}

// Sets the route resource values for creation with cf-client from the terraform struct values.
func (data *routeType) mapCreateRouteTypeToValues(ctx context.Context) (resource.RouteCreate, diag.Diagnostics) {

	routeCreate := resource.NewRouteCreate(data.Domain.ValueString(), data.Space.ValueString())
	if !data.Host.IsNull() {
		routeCreate.Host = data.Host.ValueStringPointer()
	}
	if !data.Path.IsNull() {
		routeCreate.Path = data.Path.ValueStringPointer()
	}
	if !data.Port.IsNull() {
		routeCreate.Port = inttointptr(int(data.Port.ValueInt64()))
	}

	var diags, diagnostics diag.Diagnostics
	routeCreate.Metadata = resource.NewMetadata()

	diags = data.Labels.ElementsAs(ctx, &routeCreate.Metadata.Labels, false)
	diagnostics.Append(diags...)

	diags = data.Annotations.ElementsAs(ctx, &routeCreate.Metadata.Annotations, false)
	diagnostics.Append(diags...)

	return *routeCreate, diags
}

// Sets the destinations resource values for creation with cf-client from the terraform struct values.
func (data *routeType) mapCreateDestinationsTypeToValues(ctx context.Context) ([]*resource.RouteDestinationInsertOrReplace, diag.Diagnostics) {
	var (
		destinations       []destinationType
		diags, diagnostics diag.Diagnostics
	)
	diags = data.Destinations.ElementsAs(ctx, &destinations, false)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return []*resource.RouteDestinationInsertOrReplace{}, diagnostics
	}

	routeDestinations := mapDestinationArrayToDestinationValues(destinations)

	return routeDestinations, diagnostics
}

// Prepares a destination list resource for creation/updation from the terraform list of destination types.
func mapDestinationArrayToDestinationValues(destinations []destinationType) []*resource.RouteDestinationInsertOrReplace {

	destinationValues := []*resource.RouteDestinationInsertOrReplace{}
	for _, destination := range destinations {
		destinationValue := mapTypetoDestinationValues(destination)
		destinationValues = append(destinationValues, destinationValue)
	}
	return destinationValues
}

// Prepares a rule resource from the terraform rule type.
func mapTypetoDestinationValues(destination destinationType) *resource.RouteDestinationInsertOrReplace {

	routeDestination := resource.NewRouteDestinationInsertOrReplace(destination.AppId.ValueString())

	if !destination.Weight.IsNull() {
		routeDestination.WithWeight(int(destination.Weight.ValueInt64()))
	}
	if !destination.AppProcessType.IsNull() {
		routeDestination.WithProcessType(destination.AppProcessType.ValueString())
	}
	if !destination.Port.IsNull() {
		routeDestination.WithPort(int(destination.Port.ValueInt64()))
	}
	if !destination.Protocol.IsNull() {
		routeDestination.WithProtocol(destination.Protocol.ValueString())
	}

	return routeDestination
}

// Sets the route resource values for updation with cf-client from the terraform struct values.
func (plan *routeType) mapUpdateRouteTypeToValues(ctx context.Context, state *routeType) (resource.RouteUpdate, diag.Diagnostics) {
	routeUpdate := resource.RouteUpdate{}

	var diagnostics diag.Diagnostics
	routeUpdate.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return routeUpdate, diagnostics
}

func mapDestinationPointerSliceToDestinationSlice(pointerDestinations []*resource.RouteDestination) []resource.RouteDestination {
	var destinations []resource.RouteDestination
	for _, destination := range pointerDestinations {
		destinations = append(destinations, *destination)
	}
	return destinations
}

// computeValueModifier implements the plan modifier.
type computeValueModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m computeValueModifier) Description(_ context.Context) string {
	return "Forces the value to be recomputed on every apply."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m computeValueModifier) MarkdownDescription(_ context.Context) string {
	return "Forces the value to be recomputed on every apply."
}

// PlanModifyString implements the plan modification logic.
func (m computeValueModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = types.StringUnknown()
}

// PlanModifyInt64 implements the plan modification logic.
func (m computeValueModifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = types.Int64Unknown()
}

// ReComputeStringValue returns a plan modifier that forces recomputation
// of an already set value.
func ReComputeStringValue() planmodifier.String {
	return computeValueModifier{}
}

// ReComputeIntValue returns a plan modifier that forces recomputation
// of an already set value.
func ReComputeIntValue() planmodifier.Int64 {
	return computeValueModifier{}
}
