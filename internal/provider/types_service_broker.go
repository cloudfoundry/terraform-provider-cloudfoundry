package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceBrokerType struct {
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Url         types.String `tfsdk:"url"`
	Space       types.String `tfsdk:"space"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type datasourceServiceBrokerType struct {
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Url         types.String `tfsdk:"url"`
	Space       types.String `tfsdk:"space"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (a *serviceBrokerType) Reduce() datasourceServiceBrokerType {
	var reduced datasourceServiceBrokerType
	copyFields(&reduced, a)
	return reduced
}

func mapServiceBrokerValuesToType(ctx context.Context, value *resource.ServiceBroker) (serviceBrokerType, diag.Diagnostics) {
	var diagnostics, diags diag.Diagnostics
	serviceBrokerType := serviceBrokerType{
		Name:      types.StringValue(value.Name),
		ID:        types.StringValue(value.GUID),
		Url:       types.StringValue(value.URL),
		CreatedAt: types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
	}

	if value.Relationships.Space.Data != nil {
		serviceBrokerType.Space = types.StringValue(value.Relationships.Space.Data.GUID)
	}
	serviceBrokerType.Labels, diags = mapMetadataValueToType(ctx, value.Metadata.Labels)
	diagnostics.Append(diags...)
	serviceBrokerType.Annotations, diags = mapMetadataValueToType(ctx, value.Metadata.Annotations)
	diagnostics.Append(diags...)

	return serviceBrokerType, diagnostics
}

func (data *serviceBrokerType) mapCreateServiceBrokerTypeToValues(ctx context.Context) (resource.ServiceBrokerCreate, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	createServiceBroker := resource.NewServiceBrokerCreate(data.Name.ValueString(), data.Url.ValueString(), data.Username.ValueString(), data.Password.ValueString())

	if !data.Space.IsNull() {
		createServiceBroker.WithSpace(data.Space.ValueString())
	}

	createServiceBroker.Metadata = resource.NewMetadata()
	labelsDiags := data.Labels.ElementsAs(ctx, &createServiceBroker.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := data.Annotations.ElementsAs(ctx, &createServiceBroker.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createServiceBroker, diagnostics
}

func (plan *serviceBrokerType) mapUpdateServiceBrokerTypeToValues(ctx context.Context, state serviceBrokerType) (resource.ServiceBrokerUpdate, diag.Diagnostics) {

	updateServiceBroker := &resource.ServiceBrokerUpdate{}

	updateServiceBroker.WithName(plan.Name.ValueString())
	updateServiceBroker.WithURL(plan.Url.ValueString())
	updateServiceBroker.WithCredentials(plan.Username.ValueString(), plan.Password.ValueString())

	updateServiceBroker.Metadata = resource.NewMetadata()
	var diagnostics diag.Diagnostics
	updateServiceBroker.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateServiceBroker, diagnostics
}
