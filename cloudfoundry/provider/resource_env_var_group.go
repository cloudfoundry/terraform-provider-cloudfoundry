package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &EnvVarGroupResource{}
	_ resource.ResourceWithConfigure   = &EnvVarGroupResource{}
	_ resource.ResourceWithImportState = &EnvVarGroupResource{}
)

func NewEnvVarGroupResource() resource.Resource {
	return &EnvVarGroupResource{}
}

type EnvVarGroupResource struct {
	cfClient *cfv3client.Client
}

func (r *EnvVarGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_var_group"
}

func (r *EnvVarGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Cloud Foundry environment variable group. There are two environment variable groups: ` + "`running`" + ` and ` + "`staging`" + `. These groups contain environment variables that are applied globally to all applications during the respective lifecycle phase.

__Further documentation:__
https://v3-apidocs.cloudfoundry.org/version/3.207.0/index.html#environment-variable-groups`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the environment variable group; must be either `running` or `staging`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("running", "staging"),
				},
			},
			"variables": schema.MapAttribute{
				MarkdownDescription: "Key/value pairs of environment variables to set in the group. Setting a key that already exists updates its value. Removing a key from the map removes it from the group.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (r *EnvVarGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.cfClient = session.CFClient
}

func (r *EnvVarGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan envVarGroupType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planVars map[string]string
	resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &planVars, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patchVars := make(map[string]*string, len(planVars))
	for k, v := range planVars {
		patchVars[k] = &v
	}

	group, err := patchEnvVarGroup(ctx, r.cfClient, plan.Name.ValueString(), patchVars)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Setting Environment Variable Group",
			fmt.Sprintf("Could not set environment variable group %q: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	state, diags := mapEnvVarGroupValuesToType(ctx, group)
	resp.Diagnostics.Append(diags...)
	state.Variables, diags = filterEnvVarsToKnownKeys(ctx, state.Variables, plan.Variables)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvVarGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data envVarGroupType
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.cfClient.EnvVarGroups.Get(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Reading Environment Variable Group",
			fmt.Sprintf("Could not read environment variable group %q: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	state, diags := mapEnvVarGroupValuesToType(ctx, group)
	resp.Diagnostics.Append(diags...)
	// Filter to only the keys currently tracked in state; if null (e.g. after import), keep all CF vars.
	state.Variables, diags = filterEnvVarsToKnownKeys(ctx, state.Variables, data.Variables)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvVarGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState envVarGroupType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planVars, stateVars map[string]string
	resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &planVars, false)...)
	resp.Diagnostics.Append(previousState.Variables.ElementsAs(ctx, &stateVars, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patchVars := make(map[string]*string)
	for k, v := range planVars {
		patchVars[k] = &v
	}
	for k := range stateVars {
		if _, exists := planVars[k]; !exists {
			patchVars[k] = nil
		}
	}

	group, err := patchEnvVarGroup(ctx, r.cfClient, plan.Name.ValueString(), patchVars)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Environment Variable Group",
			fmt.Sprintf("Could not update environment variable group %q: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	state, diags := mapEnvVarGroupValuesToType(ctx, group)
	resp.Diagnostics.Append(diags...)
	state.Variables, diags = filterEnvVarsToKnownKeys(ctx, state.Variables, plan.Variables)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvVarGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state envVarGroupType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateVars map[string]string
	resp.Diagnostics.Append(state.Variables.ElementsAs(ctx, &stateVars, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patchVars := make(map[string]*string, len(stateVars))
	for k := range stateVars {
		patchVars[k] = nil
	}

	_, err := patchEnvVarGroup(ctx, r.cfClient, state.Name.ValueString(), patchVars)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Clearing Environment Variable Group",
			fmt.Sprintf("Could not clear environment variable group %q: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

func (r *EnvVarGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// envVarGroupPatchBody is the JSON body for PATCH /v3/environment_variable_groups/{name}.
// Using *string allows null values to remove keys from the group.
type envVarGroupPatchBody struct {
	Var map[string]*string `json:"var"`
}

// patchEnvVarGroup makes a PATCH call supporting null values for key removal.
func patchEnvVarGroup(ctx context.Context, cfClient *cfv3client.Client, name string, vars map[string]*string) (*cfv3resource.EnvVarGroup, error) {
	body := envVarGroupPatchBody{Var: vars}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	url := cfClient.ApiURL(fmt.Sprintf("/v3/environment_variable_groups/%s", name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating PATCH request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfClient.ExecuteAuthRequest(req)
	if err != nil {
		return nil, fmt.Errorf("executing PATCH request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var group cfv3resource.EnvVarGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &group, nil
}
