package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform struct for storing values for isolation segment data source and resource.
type IsolationSegmentType struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Terraform struct for storing values for isolation segment entitlement resource.
type IsolationSegmentEntitlementType struct {
	Segment types.String `tfsdk:"segment"`
	Orgs    types.Set    `tfsdk:"orgs"`
	Default types.Bool   `tfsdk:"default"`
}

// Terraform struct for storing values for isolation segment entitlement datasource.
type IsolationSegmentEntitlementDataSourceType struct {
	Segment types.String `tfsdk:"segment"`
	Orgs    types.Set    `tfsdk:"orgs"`
}

type IsolationSegmentsType struct {
	Name              types.String           `tfsdk:"name"`
	IsolationSegments []IsolationSegmentType `tfsdk:"isolation_segments"`
}

// Sets the isolation segment resource values for creation with cf-client from the terraform struct values.
func (data *IsolationSegmentType) mapCreateIsolationSegmentTypeToValues(ctx context.Context) (resource.IsolationSegmentCreate, diag.Diagnostics) {

	createIsolationSegment := resource.NewIsolationSegmentCreate(data.Name.ValueString())

	var diagnostics diag.Diagnostics
	createIsolationSegment.Metadata = resource.NewMetadata()

	labelsDiags := data.Labels.ElementsAs(ctx, &createIsolationSegment.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := data.Annotations.ElementsAs(ctx, &createIsolationSegment.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createIsolationSegment, diagnostics
}

// Sets the terraform struct values from the isolation segment resource returned by the cf-client.
func mapIsolationSegmentValuesToType(ctx context.Context, isolationSegment *resource.IsolationSegment) (IsolationSegmentType, diag.Diagnostics) {

	isolationSegmentType := IsolationSegmentType{
		Id:        types.StringValue(isolationSegment.GUID),
		Name:      types.StringValue(isolationSegment.Name),
		CreatedAt: types.StringValue(isolationSegment.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(isolationSegment.UpdatedAt.Format(time.RFC3339)),
	}

	var diags, diagnostics diag.Diagnostics
	isolationSegmentType.Labels, diags = mapMetadataValueToType(ctx, isolationSegment.Metadata.Labels)
	diagnostics.Append(diags...)
	isolationSegmentType.Annotations, diags = mapMetadataValueToType(ctx, isolationSegment.Metadata.Annotations)
	diagnostics.Append(diags...)

	return isolationSegmentType, diagnostics
}

// Sets the isolation segment resource values for updation with cf-client from the terraform struct values.
func (plan *IsolationSegmentType) mapUpdateIsolationSegmentTypeToValues(ctx context.Context, state IsolationSegmentType) (resource.IsolationSegmentUpdate, diag.Diagnostics) {

	updateIsolationSegment := &resource.IsolationSegmentUpdate{
		Name: strtostrptr(plan.Name.ValueString()),
	}

	var diagnostics diag.Diagnostics
	updateIsolationSegment.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateIsolationSegment, diagnostics
}

// Sets the terraform struct values from the entitled orgs returned by the cf-client.
func (plan *IsolationSegmentEntitlementType) mapIsolationSegmentEntitlementValuesToType(ctx context.Context, entitledOrgs []resource.Relationship, makeDefaultSegment types.Bool) diag.Diagnostics {

	var diagnostics diag.Diagnostics
	entitledOrgSet, diags := setRelationshipToTFSet(entitledOrgs)
	diagnostics.Append(diags...)
	sameOrgs, diags := findSameRelationsFromTFState(ctx, plan.Orgs, entitledOrgSet)
	diagnostics.Append(diags...)
	plan.Orgs, diags = types.SetValueFrom(ctx, types.StringType, sameOrgs)
	diagnostics.Append(diags...)
	plan.Default = makeDefaultSegment

	return diagnostics
}

func mapIsolationSegmentsValuesToType(ctx context.Context, isoSegs []*resource.IsolationSegment) ([]IsolationSegmentType, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	isoSegsList := []IsolationSegmentType{}
	for _, isoSeg := range isoSegs {
		isoSegValue, diags := mapIsolationSegmentValuesToType(ctx, isoSeg)
		diagnostics.Append(diags...)
		isoSegsList = append(isoSegsList, isoSegValue)
	}

	return isoSegsList, diagnostics
}
