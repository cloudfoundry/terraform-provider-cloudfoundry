package provider

import (
	"context"

	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/mta"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MtarType struct {
	MtarPath                   types.String `tfsdk:"mtar_path"`
	MtarUrl                    types.String `tfsdk:"mtar_url"`
	ExtensionDescriptors       types.Set    `tfsdk:"extension_descriptors"`
	ExtensionDescriptorsString types.Set    `tfsdk:"extension_descriptors_string"`
	DeployUrl                  types.String `tfsdk:"deploy_url"`
	Space                      types.String `tfsdk:"space"`
	Mta                        types.Object `tfsdk:"mta"`
	Namespace                  types.String `tfsdk:"namespace"`
	Id                         types.String `tfsdk:"id"`
	SourceCodeHash             types.String `tfsdk:"source_code_hash"`
	DeployStrategy             types.String `tfsdk:"deploy_strategy"`
	VersionRule                types.String `tfsdk:"version_rule"`
	Modules                    types.Set    `tfsdk:"modules"`
	SkipIdleStart              types.Bool   `tfsdk:"skip_idle_start"`
}

type MtasDataSourceType struct {
	Space     types.String `tfsdk:"space"`
	Id        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Mtas      types.List   `tfsdk:"mtas"`
	DeployUrl types.String `tfsdk:"deploy_url"`
}

type MtaDataSourceType struct {
	Space     types.String `tfsdk:"space"`
	Id        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Mta       types.Object `tfsdk:"mta"`
	DeployUrl types.String `tfsdk:"deploy_url"`
}

type MtaType struct {
	Metadata types.Object `tfsdk:"metadata"`
	Modules  types.List   `tfsdk:"modules"`
	Services types.List   `tfsdk:"services"`
}

type MtaMetadataType struct {
	Id        types.String `tfsdk:"id"`
	Version   types.String `tfsdk:"version"`
	Namespace types.String `tfsdk:"namespace"`
}

type MtaModuleType struct {
	ModuleName            types.String `tfsdk:"module_name"`
	AppName               types.String `tfsdk:"app_name"`
	CreatedOn             types.String `tfsdk:"created_on"`
	UpdatedOn             types.String `tfsdk:"updated_on"`
	ProvidedDendencyNames types.List   `tfsdk:"provided_dendency_names"`
	Services              types.List   `tfsdk:"services"`
	Uris                  types.List   `tfsdk:"uris"`
}

var mtaObjType = types.ObjectType{
	AttrTypes: mtaObjAttributes,
}

var mtaObjAttributes = map[string]attr.Type{
	"metadata": types.ObjectType{
		AttrTypes: mtaMetadataObjType,
	},
	"modules": types.ListType{
		ElemType: mtaModuleObjType,
	},
	"services": types.ListType{
		ElemType: types.StringType,
	},
}

var mtaMetadataObjType = map[string]attr.Type{
	"id":        types.StringType,
	"version":   types.StringType,
	"namespace": types.StringType,
}

var mtaModuleObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"module_name": types.StringType,
		"app_name":    types.StringType,
		"created_on":  types.StringType,
		"updated_on":  types.StringType,
		"provided_dendency_names": types.ListType{
			ElemType: types.StringType,
		},
		"services": types.ListType{
			ElemType: types.StringType,
		},
		"uris": types.ListType{
			ElemType: types.StringType,
		},
	},
}

// Sets the terraform struct values from the mta resource returned by the mta-client.
func mapMtasValuesToType(ctx context.Context, data MtasDataSourceType, mtas []mta.Mta) (MtasDataSourceType, diag.Diagnostics) {

	mtarDataSourceType := MtasDataSourceType{
		Space:     data.Space,
		Id:        data.Id,
		Namespace: data.Namespace,
		DeployUrl: data.DeployUrl,
	}

	var diagnostics, diags diag.Diagnostics
	mtasList := []MtaType{}

	for _, mta := range mtas {
		mtaValue, diags := mapMtaValuesToType(ctx, mta)
		diagnostics.Append(diags...)
		mtasList = append(mtasList, mtaValue)
	}
	mtarDataSourceType.Mtas, diags = types.ListValueFrom(ctx, mtaObjType, mtasList)
	diagnostics.Append(diags...)
	return mtarDataSourceType, diagnostics

}

func mapMtaValuesToType(ctx context.Context, mta mta.Mta) (MtaType, diag.Diagnostics) {
	mtaType := MtaType{}
	var diags, diagnostics diag.Diagnostics

	mtaType.Metadata, diags = types.ObjectValueFrom(ctx, mtaMetadataObjType, mapMtaMetadataValuesToType(*mta.Metadata))
	diagnostics.Append(diags...)
	mtaModulesList := []MtaModuleType{}

	for _, module := range mta.Modules {
		moduleValue, diags := mapMtaModuleValuesToType(ctx, module)
		diagnostics.Append(diags...)
		mtaModulesList = append(mtaModulesList, moduleValue)
	}
	mtaType.Modules, diags = types.ListValueFrom(ctx, mtaModuleObjType, mtaModulesList)
	diagnostics.Append(diags...)
	mtaType.Services, diags = types.ListValueFrom(ctx, types.StringType, mta.Services)
	diagnostics.Append(diags...)

	return mtaType, diagnostics
}

func mapMtaMetadataValuesToType(metadata mta.Metadata) MtaMetadataType {
	mtaMetadataType := MtaMetadataType{
		Id:        types.StringValue(metadata.Id),
		Version:   types.StringValue(metadata.Version),
		Namespace: types.StringValue(metadata.Namespace),
	}
	return mtaMetadataType
}

func mapMtaModuleValuesToType(ctx context.Context, module mta.Module) (MtaModuleType, diag.Diagnostics) {
	var diags, diagnostics diag.Diagnostics
	mtaModuleType := MtaModuleType{
		ModuleName: types.StringValue(module.ModuleName),
		AppName:    types.StringValue(module.AppName),
	}
	if module.CreatedOn.IsZero() {
		mtaModuleType.CreatedOn = types.StringNull()
	} else {
		mtaModuleType.CreatedOn = types.StringValue(module.CreatedOn.String())
	}
	if module.UpdatedOn.IsZero() {
		mtaModuleType.UpdatedOn = types.StringNull()
	} else {
		mtaModuleType.UpdatedOn = types.StringValue(module.UpdatedOn.String())
	}
	mtaModuleType.ProvidedDendencyNames, diags = types.ListValueFrom(ctx, types.StringType, module.ProvidedDendencyNames)
	diagnostics.Append(diags...)
	mtaModuleType.Services, diags = types.ListValueFrom(ctx, types.StringType, module.Services)
	diagnostics.Append(diags...)
	mtaModuleType.Uris, diags = types.ListValueFrom(ctx, types.StringType, module.Uris)
	diagnostics.Append(diags...)
	return mtaModuleType, diags
}
