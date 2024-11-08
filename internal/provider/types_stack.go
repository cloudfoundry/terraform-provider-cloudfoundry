package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stackType struct {
	Name             types.String `tfsdk:"name"`
	ID               types.String `tfsdk:"id"`
	Description      types.String `tfsdk:"description"`
	BuildRootfsImage types.String `tfsdk:"build_rootfs_image"`
	RunRootfsImage   types.String `tfsdk:"run_rootfs_image"`
	Default          types.Bool   `tfsdk:"default"`
	Labels           types.Map    `tfsdk:"labels"`
	Annotations      types.Map    `tfsdk:"annotations"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

type stacksDatasourceType struct {
	Name   types.String `tfsdk:"name"`
	Stacks []stackType  `tfsdk:"stacks"`
}

func mapStackValuesToType(ctx context.Context, value *resource.Stack) (stackType, diag.Diagnostics) {
	var diagnostics, diags diag.Diagnostics
	stackType := stackType{
		Name:             types.StringValue(value.Name),
		ID:               types.StringValue(value.GUID),
		BuildRootfsImage: types.StringValue(value.BuildRootfsImage),
		RunRootfsImage:   types.StringValue(value.RunRootfsImage),
		Default:          types.BoolValue(value.Default),
		CreatedAt:        types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:        types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
	}

	if value.Description != nil {
		stackType.Description = types.StringValue(*value.Description)
	}
	stackType.Labels, diags = mapMetadataValueToType(ctx, value.Metadata.Labels)
	diagnostics.Append(diags...)
	stackType.Annotations, diags = mapMetadataValueToType(ctx, value.Metadata.Annotations)
	diagnostics.Append(diags...)

	return stackType, diagnostics
}

func mapStacksValuesToType(ctx context.Context, stacks []*resource.Stack) ([]stackType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	stacksList := []stackType{}
	for _, stack := range stacks {
		spaceQuotaValue, diags := mapStackValuesToType(ctx, stack)
		diagnostics.Append(diags...)
		stacksList = append(stacksList, spaceQuotaValue)
	}

	return stacksList, diagnostics
}
