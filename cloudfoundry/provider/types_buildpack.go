package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type buildpackType struct {
	Name           types.String `tfsdk:"name"`
	Id             types.String `tfsdk:"id"`
	Path           types.String `tfsdk:"path"`
	State          types.String `tfsdk:"state"`
	Stack          types.String `tfsdk:"stack"`
	Filename       types.String `tfsdk:"filename"`
	Position       types.Int64  `tfsdk:"position"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Locked         types.Bool   `tfsdk:"locked"`
	Labels         types.Map    `tfsdk:"labels"`
	Annotations    types.Map    `tfsdk:"annotations"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	SourceCodeHash types.String `tfsdk:"source_code_hash"`
}

type datasourceBuildpackType struct {
	Name        types.String `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	State       types.String `tfsdk:"state"`
	Stack       types.String `tfsdk:"stack"`
	Filename    types.String `tfsdk:"filename"`
	Position    types.Int64  `tfsdk:"position"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Locked      types.Bool   `tfsdk:"locked"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type datasourceBuildpacksType struct {
	Name       types.String              `tfsdk:"name"`
	Stack      types.String              `tfsdk:"stack"`
	Buildpacks []datasourceBuildpackType `tfsdk:"buildpacks"`
}

func (a *buildpackType) Reduce() datasourceBuildpackType {
	var reduced datasourceBuildpackType
	copyFields(&reduced, a)
	return reduced
}

// Sets the terraform struct values from the buildpack resource returned by the cf-client.
func mapBuildpackValuesToType(ctx context.Context, buildpack *resource.Buildpack) (buildpackType, diag.Diagnostics) {

	buildpackType := buildpackType{
		Name:      types.StringValue(buildpack.Name),
		Id:        types.StringValue(buildpack.GUID),
		CreatedAt: types.StringValue(buildpack.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(buildpack.UpdatedAt.Format(time.RFC3339)),
		State:     types.StringValue(buildpack.State),
		Position:  types.Int64Value(int64(buildpack.Position)),
		Enabled:   types.BoolValue(buildpack.Enabled),
		Locked:    types.BoolValue(buildpack.Locked),
	}
	if buildpack.Filename != nil {
		buildpackType.Filename = types.StringValue(*buildpack.Filename)
	}
	if buildpack.Stack != nil {
		buildpackType.Stack = types.StringValue(*buildpack.Stack)
	}

	var diags, diagnostics diag.Diagnostics
	buildpackType.Labels, diags = mapMetadataValueToType(ctx, buildpack.Metadata.Labels)
	diagnostics.Append(diags...)
	buildpackType.Annotations, diags = mapMetadataValueToType(ctx, buildpack.Metadata.Annotations)
	diagnostics.Append(diags...)

	return buildpackType, diagnostics
}

func mapBuildpacksValuesToType(ctx context.Context, buildpacks []*resource.Buildpack) ([]datasourceBuildpackType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	buildpackList := []datasourceBuildpackType{}
	for _, buildpack := range buildpacks {
		buildpackValue, diags := mapBuildpackValuesToType(ctx, buildpack)
		diagnostics.Append(diags...)
		buildpackList = append(buildpackList, buildpackValue.Reduce())
	}

	return buildpackList, diagnostics
}

// Sets the buildpack resource values for creation with cf-client from the terraform struct values.
func (data *buildpackType) mapCreateBuildpackTypeToValues(ctx context.Context) (resource.BuildpackCreateOrUpdate, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	createBuildpack := resource.NewBuildpackCreate(data.Name.ValueString())

	if !data.Position.IsUnknown() {
		createBuildpack.WithPosition(int(data.Position.ValueInt64()))
	}
	if !data.Stack.IsNull() {
		createBuildpack.WithStack(data.Stack.ValueString())
	}
	if !data.Enabled.IsNull() {
		createBuildpack.WithEnabled(data.Enabled.ValueBool())
	}
	if !data.Locked.IsNull() {
		createBuildpack.WithLocked(data.Locked.ValueBool())
	}

	createBuildpack.Metadata = resource.NewMetadata()
	labelsDiags := data.Labels.ElementsAs(ctx, &createBuildpack.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := data.Annotations.ElementsAs(ctx, &createBuildpack.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createBuildpack, diagnostics
}

// Sets the buildpack resource values for updation with cf-client from the terraform struct values.
func (plan *buildpackType) mapUpdateBuildpackTypeToValues(ctx context.Context, state buildpackType) (resource.BuildpackCreateOrUpdate, diag.Diagnostics) {

	updateBuildpack := resource.NewBuildpackUpdate()
	updateBuildpack.WithName(plan.Name.ValueString())

	if !plan.Position.IsUnknown() {
		updateBuildpack.WithPosition(int(plan.Position.ValueInt64()))
	}
	if !plan.Stack.IsNull() {
		updateBuildpack.WithStack(plan.Stack.ValueString())
	}
	if !plan.Enabled.IsNull() {
		updateBuildpack.WithEnabled(plan.Enabled.ValueBool())
	}
	if !plan.Locked.IsNull() {
		updateBuildpack.WithLocked(plan.Locked.ValueBool())
	}

	var diagnostics diag.Diagnostics
	updateBuildpack.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateBuildpack, diagnostics
}
