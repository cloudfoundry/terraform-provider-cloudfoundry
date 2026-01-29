package provider

import (
	"context"
	"fmt"
	"reflect"
	"time"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/samber/lo"
)

const (
	idKey          = "id"
	labelsKey      = "labels"
	annotationsKey = "annotations"
	createdAtKey   = "created_at"
	updatedAtKey   = "updated_at"
)

const defaultTimeout = 20 * time.Minute

func datasourceLabelsSchema() *schema.MapAttribute {
	return &schema.MapAttribute{
		MarkdownDescription: "The labels associated with Cloud Foundry resources.",
		ElementType:         types.StringType,
		Computed:            true,
	}
}

func datasourceAnnotationsSchema() *schema.MapAttribute {
	return &schema.MapAttribute{
		MarkdownDescription: "The annotations associated with Cloud Foundry resources.",
		ElementType:         types.StringType,
		Computed:            true,
	}
}

func resourceLabelsSchema() *schema.MapAttribute {
	return &schema.MapAttribute{
		MarkdownDescription: `The labels associated with Cloud Foundry resources. Add as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).`,
		ElementType:         types.StringType,
		Optional:            true,
		Validators: []validator.Map{
			mapvalidator.SizeAtLeast(1),
		},
	}
}

func resourceAnnotationsSchema() *schema.MapAttribute {
	return &schema.MapAttribute{
		MarkdownDescription: "The annotations associated with Cloud Foundry resources. Add as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).",
		ElementType:         types.StringType,
		Optional:            true,
		Validators: []validator.Map{
			mapvalidator.SizeAtLeast(1),
		},
	}
}

func createdAtSchema() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: "The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}
func updatedAtSchema() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: "The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
		Computed:            true,
	}
}
func guidSchema() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: "The GUID of the object.",
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// useStateForUnknownUnlessBlueGreen implements a plan modifier that copies the prior
// state value into the planned value, unless the deployment strategy is "blue-green".
// This is needed because blue-green deployments create a new app with a new GUID.
type useStateForUnknownUnlessBlueGreen struct{}

// Description returns a human-readable description of the plan modifier.
func (m useStateForUnknownUnlessBlueGreen) Description(_ context.Context) string {
	return "Uses state for unknown unless deployment strategy is blue-green."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateForUnknownUnlessBlueGreen) MarkdownDescription(_ context.Context) string {
	return "Uses state for unknown unless deployment strategy is blue-green."
}

// PlanModifyString implements the plan modification logic.
func (m useStateForUnknownUnlessBlueGreen) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state (resource is being created).
	if req.State.Raw.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Check the deployment strategy from the plan
	var strategy types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("strategy"), &strategy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If strategy is "blue-green", leave the value as unknown (will be computed)
	// because a new app is created during blue-green deployment
	if !strategy.IsNull() && strategy.ValueString() == "blue-green" {
		// Leave resp.PlanValue as unknown - don't copy state
		return
	}

	// Otherwise, use the state value (normal behavior like UseStateForUnknown)
	resp.PlanValue = req.StateValue
}

// UseStateForUnknownUnlessBlueGreen returns a plan modifier that copies the prior
// state value into the planned value, unless the deployment strategy is "blue-green".
func UseStateForUnknownUnlessBlueGreen() planmodifier.String {
	return useStateForUnknownUnlessBlueGreen{}
}

// appGuidSchema returns a schema for the app GUID that uses the custom plan modifier.
// This is needed because blue-green deployments create a new app with a new GUID.
func appGuidSchema() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: "The GUID of the object.",
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			UseStateForUnknownUnlessBlueGreen(),
		},
	}
}

// appCreatedAtSchema returns a schema for the app created_at that uses the custom plan modifier.
// This is needed because blue-green deployments create a new app with a new created_at timestamp.
func appCreatedAtSchema() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: "The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			UseStateForUnknownUnlessBlueGreen(),
		},
	}
}

func lastOperationSchema() *schema.SingleNestedAttribute {
	return &schema.SingleNestedAttribute{
		MarkdownDescription: "The details of the last operation performed on the resource",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the last operation",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the last operation",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the last operation",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The time at which the last operation was updated",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time at which the last operation was created",
				Computed:            true,
			},
		},
	}
}

// Take relationship from cfclient and return set type of terraform.
func setRelationshipToTFSet(r []cfv3resource.Relationship) (basetypes.SetValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	var bt basetypes.SetValue
	if len(r) != 0 {
		tfVal := []attr.Value{}
		for _, val := range r {
			tfVal = append(tfVal, types.StringValue(val.GUID))
		}
		bt, diags = types.SetValue(types.StringType, tfVal)
	} else {
		bt = types.SetNull(types.StringType)
	}
	return bt, diags
}

// Take array of relationships from cfclient and return array of GUID's.
func setRelationshipToSlice(r []cfv3resource.Relationship) (values []string) {
	for _, val := range r {
		values = append(values, val.GUID)
	}
	return values
}

// Returns removed and added element in the new plan which existed in state.
func findChangedRelationsFromTFState(ctx context.Context, planSet basetypes.SetValue, stateSet basetypes.SetValue) ([]string, []string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var planSetStr, stateSetStr []string
	diags = append(diags, planSet.ElementsAs(ctx, &planSetStr, false)...)
	diags = append(diags, stateSet.ElementsAs(ctx, &stateSetStr, false)...)
	removed, added := lo.Difference(stateSetStr, planSetStr)
	return removed, added, diags
}

// Returns same element in the new plan which existed in state.
func findSameRelationsFromTFState(ctx context.Context, planSet basetypes.SetValue, stateSet basetypes.SetValue) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var planSetStr, stateSetStr []string
	diags = append(diags, planSet.ElementsAs(ctx, &planSetStr, false)...)
	diags = append(diags, stateSet.ElementsAs(ctx, &stateSetStr, false)...)
	same := lo.Intersect(stateSetStr, planSetStr)
	return same, diags
}

func handleReadErrors(ctx context.Context, resp *resource.ReadResponse, err error, res string, resName string) {
	if cfv3resource.IsResourceNotFoundError(err) {
		resp.State.RemoveResource(ctx)
	} else {
		resp.Diagnostics.AddError(fmt.Sprintf("API Error Reading %s %s", res, resName), err.Error())
	}

}

func pollJob(ctx context.Context, client cfv3client.Client, jobID string, timeout time.Duration) error {

	return client.Jobs.PollComplete(ctx, jobID, &cfv3client.PollingOptions{
		Timeout:       timeout,
		CheckInterval: time.Second * 2,
		FailedState:   string(cfv3resource.JobStateFailed),
	})
}

func mapMetadataValueToType(ctx context.Context, generic map[string]*string) (basetypes.MapValue, diag.Diagnostics) {

	var out basetypes.MapValue
	var diagnostics diag.Diagnostics
	if len(generic) == 0 {
		out = types.MapNull(types.StringType)
	} else {
		out, diagnostics = types.MapValueFrom(ctx, types.StringType, generic)
	}

	return out, diagnostics
}

// Prepares the labels and annotations for cfclient updation from existing and planned tfstate labels and annotations.
func setClientMetadataForUpdate(ctx context.Context, StateLabels basetypes.MapValue, StateAnnotations basetypes.MapValue, plannedStateLabels basetypes.MapValue, plannedStateAnnotations basetypes.MapValue) (*cfv3resource.Metadata, diag.Diagnostics) {

	var (
		diagnostics                                                diag.Diagnostics
		planLabels, planAnnotations, stateLabels, stateAnnotations map[string]*string
	)

	metadata := cfv3resource.NewMetadata()

	diagnostics.Append(plannedStateLabels.ElementsAs(ctx, &planLabels, false)...)
	diagnostics.Append(plannedStateAnnotations.ElementsAs(ctx, &planAnnotations, false)...)
	diagnostics.Append(StateLabels.ElementsAs(ctx, &stateLabels, false)...)
	diagnostics.Append(StateAnnotations.ElementsAs(ctx, &stateAnnotations, false)...)

	if diagnostics.HasError() {
		return metadata, diagnostics
	}

	for key := range stateLabels {
		metadata.RemoveLabel("", key)
	}

	for key := range stateAnnotations {
		metadata.RemoveAnnotation("", key)
	}

	for key, value := range planLabels {
		metadata.SetLabel("", key, *value)
	}

	for key, value := range planAnnotations {
		metadata.SetAnnotation("", key, *value)
	}

	return metadata, diagnostics
}

// Returns a pointer to a bool.
func booltoboolptr(s bool) *bool {
	return &s
}

// Returns a pointer to an int.
func inttointptr(s int) *int {
	return &s
}

// Returns a pointer to an uint.
func uinttouintptr(s uint) *uint {
	return &s
}

func strtostrptr(s string) *string {
	return &s
}

func copyFields(dst, src interface{}) {
	dstValue := reflect.ValueOf(dst).Elem()
	srcValue := reflect.ValueOf(src).Elem()

	for i := 0; i < dstValue.NumField(); i++ {
		dstField := dstValue.Field(i)
		srcField := srcValue.FieldByName(dstValue.Type().Field(i).Name)

		if srcField.IsValid() {
			dstField.Set(srcField)
		}
	}
}
