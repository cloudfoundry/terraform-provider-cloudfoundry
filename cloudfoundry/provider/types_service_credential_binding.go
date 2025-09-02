package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceserviceCredentialBindingType struct {
	Name               types.String                                  `tfsdk:"name"`
	ServiceInstance    types.String                                  `tfsdk:"service_instance"`
	App                types.String                                  `tfsdk:"app"`
	CredentialBindings []serviceCredentialBindingTypeWithCredentials `tfsdk:"credential_bindings"`
}

type serviceCredentialBindingTypeWithCredentials struct {
	Name            types.String         `tfsdk:"name"`
	ID              types.String         `tfsdk:"id"`
	Type            types.String         `tfsdk:"type"`
	ServiceInstance types.String         `tfsdk:"service_instance"`
	Credentials     jsontypes.Normalized `tfsdk:"credential_binding"`
	Parameters      jsontypes.Normalized `tfsdk:"parameters"`
	App             types.String         `tfsdk:"app"`
	LastOperation   types.Object         `tfsdk:"last_operation"` //LastOperationType
	Labels          types.Map            `tfsdk:"labels"`
	Annotations     types.Map            `tfsdk:"annotations"`
	CreatedAt       types.String         `tfsdk:"created_at"`
	UpdatedAt       types.String         `tfsdk:"updated_at"`
}

func mapServiceCredentialBindingValuesToType(ctx context.Context, value *resource.ServiceCredentialBinding) (serviceCredentialBindingTypeWithCredentials, diag.Diagnostics) {
	var diags, diagnostics diag.Diagnostics
	serviceCredentialBindingType := serviceCredentialBindingTypeWithCredentials{
		ID:              types.StringValue(value.GUID),
		Type:            types.StringValue(value.Type),
		CreatedAt:       types.StringValue(value.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:       types.StringValue(value.UpdatedAt.Format(time.RFC3339)),
		ServiceInstance: types.StringValue(value.Relationships.ServiceInstance.Data.GUID),
	}

	if value.Name != nil {
		serviceCredentialBindingType.Name = types.StringValue(*value.Name)
	}

	if value.Relationships.App != nil {
		serviceCredentialBindingType.App = types.StringValue(value.Relationships.App.Data.GUID)
	}

	serviceCredentialBindingType.Labels, diags = mapMetadataValueToType(ctx, value.Metadata.Labels)
	diagnostics.Append(diags...)
	serviceCredentialBindingType.Annotations, diags = mapMetadataValueToType(ctx, value.Metadata.Annotations)
	diagnostics.Append(diags...)
	serviceCredentialBindingType.LastOperation, diags = types.ObjectValueFrom(ctx, lastOperationAttrTypes, mapLastOperation(value.LastOperation))
	diagnostics.Append(diags...)

	return serviceCredentialBindingType, diagnostics
}

func (data *serviceCredentialBindingTypeWithCredentials) mapCreateServiceCredentialBindingTypeToValues(ctx context.Context) (resource.ServiceCredentialBindingCreate, diag.Diagnostics) {

	var (
		diagnostics                    diag.Diagnostics
		createServiceCredentialBinding *resource.ServiceCredentialBindingCreate
	)
	switch data.Type.ValueString() {
	case appServiceCredentialBinding:
		createServiceCredentialBinding = resource.NewServiceCredentialBindingCreateApp(data.ServiceInstance.ValueString(), data.App.ValueString())
		if !data.Name.IsNull() {
			createServiceCredentialBinding.WithName(data.Name.ValueString())
		}
	case keyServiceCredentialBinding:
		createServiceCredentialBinding = resource.NewServiceCredentialBindingCreateKey(data.ServiceInstance.ValueString(), data.Name.ValueString())
	}

	if !data.Parameters.IsNull() {
		createServiceCredentialBinding.WithJSONParameters(data.Parameters.ValueString())
	}

	createServiceCredentialBinding.Metadata = resource.NewMetadata()
	labelsDiags := data.Labels.ElementsAs(ctx, &createServiceCredentialBinding.Metadata.Labels, false)
	diagnostics.Append(labelsDiags...)
	annotationsDiags := data.Annotations.ElementsAs(ctx, &createServiceCredentialBinding.Metadata.Annotations, false)
	diagnostics.Append(annotationsDiags...)

	return *createServiceCredentialBinding, diagnostics
}

func (plan *serviceCredentialBindingTypeWithCredentials) mapUpdateServiceCredentialBindingTypeToValues(ctx context.Context, state serviceCredentialBindingTypeWithCredentials) (resource.ServiceCredentialBindingUpdate, diag.Diagnostics) {

	updateCredBinding := &resource.ServiceCredentialBindingUpdate{}

	var diagnostics diag.Diagnostics
	updateCredBinding.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateCredBinding, diagnostics
}
