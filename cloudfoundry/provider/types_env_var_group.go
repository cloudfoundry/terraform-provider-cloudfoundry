package provider

import (
	"context"
	"time"

	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type envVarGroupType struct {
	Name      types.String `tfsdk:"name"`
	Variables types.Map    `tfsdk:"variables"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func mapEnvVarGroupValuesToType(ctx context.Context, group *cfv3resource.EnvVarGroup) (envVarGroupType, diag.Diagnostics) {
	var diags diag.Diagnostics
	var evg envVarGroupType

	evg.Name = types.StringValue(group.Name)
	evg.UpdatedAt = types.StringValue(group.UpdatedAt.Format(time.RFC3339))

	if len(group.Var) > 0 {
		vars, d := types.MapValueFrom(ctx, types.StringType, group.Var)
		diags.Append(d...)
		evg.Variables = vars
	} else {
		evg.Variables = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	return evg, diags
}

// filterEnvVarsToKnownKeys returns a subset of cfVars containing only the keys present in knownVars.
// When knownVars is null (e.g. during import), all cfVars are returned unchanged.
func filterEnvVarsToKnownKeys(ctx context.Context, cfVars types.Map, knownVars types.Map) (types.Map, diag.Diagnostics) {
	if knownVars.IsNull() {
		return cfVars, nil
	}
	var knownKeys map[string]string
	if diags := knownVars.ElementsAs(ctx, &knownKeys, false); diags.HasError() {
		return cfVars, diags
	}
	var allVars map[string]string
	if diags := cfVars.ElementsAs(ctx, &allVars, false); diags.HasError() {
		return cfVars, diags
	}
	filtered := make(map[string]string, len(knownKeys))
	for k := range knownKeys {
		if v, ok := allVars[k]; ok {
			filtered[k] = v
		}
	}
	return types.MapValueFrom(ctx, types.StringType, filtered)
}
