package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Terraform struct for storing values for space data source and resource.
type spaceType struct {
	Name             types.String `tfsdk:"name"`
	Id               types.String `tfsdk:"id"`
	OrgId            types.String `tfsdk:"org"`
	Quota            types.String `tfsdk:"quota"`
	AllowSSH         types.Bool   `tfsdk:"allow_ssh"`
	IsolationSegment types.String `tfsdk:"isolation_segment"`
	Labels           types.Map    `tfsdk:"labels"`
	Annotations      types.Map    `tfsdk:"annotations"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// Terraform struct for storing values for spaces data.
type spacesType struct {
	Name   types.String `tfsdk:"name"`
	OrgId  types.String `tfsdk:"org"`
	Spaces []spaceType  `tfsdk:"spaces"`
}

// Sets the space resource values for creation with cf-client from the terraform struct values.
func (data *spaceType) mapCreateSpaceTypeToValues(ctx context.Context) (resource.SpaceCreate, diag.Diagnostics) {

	createSpace := resource.NewSpaceCreate(data.Name.ValueString(), data.OrgId.ValueString())
	var diagnostics diag.Diagnostics
	createSpace.Metadata = resource.NewMetadata()

	labelsDiags := data.Labels.ElementsAs(ctx, &createSpace.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)

	annotationsDiags := data.Annotations.ElementsAs(ctx, &createSpace.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createSpace, diagnostics
}

// Sets the space resource values for updation with cf-client from the terraform struct values.
func (plan *spaceType) mapUpdateSpaceTypeToValues(ctx context.Context, state spaceType) (resource.SpaceUpdate, diag.Diagnostics) {

	updateSpace := &resource.SpaceUpdate{
		Name: plan.Name.ValueString(),
	}

	var diagnostics diag.Diagnostics
	updateSpace.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateSpace, diagnostics
}

// Sets the terraform struct values from the space resource returned by the cf-client.
func mapSpaceValuesToType(ctx context.Context, space *resource.Space, sshEnabled bool, IsolationSegment string) (spaceType, diag.Diagnostics) {

	spaceType := spaceType{
		Name:             types.StringValue(space.Name),
		Id:               types.StringValue(space.GUID),
		CreatedAt:        types.StringValue(space.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:        types.StringValue(space.UpdatedAt.Format(time.RFC3339)),
		IsolationSegment: types.StringValue(IsolationSegment),
		AllowSSH:         types.BoolValue(sshEnabled),
		OrgId:            types.StringValue(space.Relationships.Organization.Data.GUID),
	}

	if space.Relationships.Quota.Data != nil {
		spaceType.Quota = types.StringValue(space.Relationships.Quota.Data.GUID)
	}

	if IsolationSegment == "" {
		spaceType.IsolationSegment = basetypes.NewStringNull()
	}

	var diags, diagnostics diag.Diagnostics
	spaceType.Labels, diags = mapMetadataValueToType(ctx, space.Metadata.Labels)
	diagnostics.Append(diags...)
	spaceType.Annotations, diags = mapMetadataValueToType(ctx, space.Metadata.Annotations)
	diagnostics.Append(diags...)

	return spaceType, diagnostics
}
