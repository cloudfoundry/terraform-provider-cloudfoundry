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

// Returns a pointer to an uint.
//
//go:fix inline
func uinttouintptr(s uint) *uint {
	return new(s)
}

func copyFields(dst, src any) {
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
