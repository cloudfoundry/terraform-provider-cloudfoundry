package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type securityGroupType struct {
	Name                   types.String `tfsdk:"name"`
	Rules                  types.List   `tfsdk:"rules"`
	GloballyEnabledRunning types.Bool   `tfsdk:"globally_enabled_running"`
	GloballyEnabledStaging types.Bool   `tfsdk:"globally_enabled_staging"`
	RunningSpaces          types.Set    `tfsdk:"running_spaces"`
	StagingSpaces          types.Set    `tfsdk:"staging_spaces"`
	Id                     types.String `tfsdk:"id"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
}

type securityGroupsType struct {
	Name           types.String        `tfsdk:"name"`
	RunningSpace   types.String        `tfsdk:"running_space"`
	StagingSpace   types.String        `tfsdk:"staging_space"`
	SecurityGroups []securityGroupType `tfsdk:"security_groups"`
}

// Terraform struct for storing values for isolation segment entitlement resource.
type securityGroupSpacesType struct {
	SecurityGroup types.String `tfsdk:"security_group"`
	RunningSpaces types.Set    `tfsdk:"running_spaces"`
	StagingSpaces types.Set    `tfsdk:"staging_spaces"`
}

type ruleType struct {
	Protocol    types.String `tfsdk:"protocol"`
	Destination types.String `tfsdk:"destination"`
	Ports       types.String `tfsdk:"ports"`
	Type        types.Int64  `tfsdk:"type"`
	Code        types.Int64  `tfsdk:"code"`
	Description types.String `tfsdk:"description"`
	Log         types.Bool   `tfsdk:"log"`
}

var ruleObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"protocol":    types.StringType,
		"destination": types.StringType,
		"ports":       types.StringType,
		"type":        types.Int64Type,
		"code":        types.Int64Type,
		"description": types.StringType,
		"log":         types.BoolType,
	},
}

// Sets the terraform struct values from the security group resource returned by the cf-client.
func mapSecurityGroupValuesToType(ctx context.Context, securityGroup *resource.SecurityGroup) (securityGroupType, diag.Diagnostics) {

	securityGroupType := securityGroupType{
		Name:                   types.StringValue(securityGroup.Name),
		CreatedAt:              types.StringValue(securityGroup.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:              types.StringValue(securityGroup.UpdatedAt.Format(time.RFC3339)),
		Id:                     types.StringValue(securityGroup.GUID),
		GloballyEnabledRunning: types.BoolValue(*securityGroup.GloballyEnabled.Running),
		GloballyEnabledStaging: types.BoolValue(*securityGroup.GloballyEnabled.Staging),
	}

	var diags, diagnostics diag.Diagnostics
	securityGroupType.RunningSpaces, diags = setRelationshipToTFSet(securityGroup.Relationships.RunningSpaces.Data)
	diagnostics.Append(diags...)
	securityGroupType.StagingSpaces, diags = setRelationshipToTFSet(securityGroup.Relationships.StagingSpaces.Data)
	diagnostics.Append(diags...)

	if len(securityGroup.Rules) == 0 {
		securityGroupType.Rules = types.ListNull(ruleObjType)
	} else {
		securityGroupType.Rules, diags = mapRuleValuesToListType(ctx, &securityGroup.Rules)
		diagnostics.Append(diags...)
	}
	return securityGroupType, diagnostics
}

// Sets the terraform struct values from the rule resource returned by the cf-client.
func mapRuleValuesToType(rule *resource.SecurityGroupRule) ruleType {

	ruleType := ruleType{
		Protocol:    types.StringValue(rule.Protocol),
		Destination: types.StringValue(rule.Destination),
	}

	if rule.Ports != nil {
		ruleType.Ports = types.StringValue(*rule.Ports)
	}
	if rule.Type != nil {
		ruleType.Type = types.Int64Value(int64(*rule.Type))
	}
	if rule.Code != nil {
		ruleType.Code = types.Int64Value(int64(*rule.Code))
	}
	if rule.Description != nil {
		ruleType.Description = types.StringValue(*rule.Description)
	}
	if rule.Log != nil {
		ruleType.Log = types.BoolValue(*rule.Log)
	}

	return ruleType
}

// Prepares a terraform list from the rule resources returned by the cf-client.
func mapRuleValuesToListType(ctx context.Context, rules *[]resource.SecurityGroupRule) (types.List, diag.Diagnostics) {

	var diags, diagnostics diag.Diagnostics
	ruleValues := []ruleType{}
	for _, rule := range *rules {
		ruleValue := mapRuleValuesToType(&rule)
		ruleValues = append(ruleValues, ruleValue)
	}

	rulesList, diags := types.ListValueFrom(ctx, ruleObjType, ruleValues)
	diagnostics.Append(diags...)

	return rulesList, diagnostics
}

// Sets the security group resource values for creation with cf-client from the terraform struct values.
func (data *securityGroupType) mapCreateSecurityGroupTypeToValues(ctx context.Context) (resource.SecurityGroupCreate, diag.Diagnostics) {

	createSecurityGroup := &resource.SecurityGroupCreate{Name: data.Name.ValueString()}

	if !data.GloballyEnabledRunning.IsNull() || !data.GloballyEnabledStaging.IsNull() {
		createSecurityGroup.GloballyEnabled = &resource.SecurityGroupGloballyEnabled{}
		if !data.GloballyEnabledRunning.IsNull() {
			createSecurityGroup.GloballyEnabled.Running = data.GloballyEnabledRunning.ValueBoolPointer()
		}
		if !data.GloballyEnabledStaging.IsNull() {
			createSecurityGroup.GloballyEnabled.Staging = data.GloballyEnabledStaging.ValueBoolPointer()
		}
	}

	var diags, diagnostics diag.Diagnostics
	if !data.StagingSpaces.IsNull() || !data.RunningSpaces.IsNull() {
		createSecurityGroup.Relationships = make(map[string]resource.ToManyRelationships)
		var spacesRelVal []string
		if !data.StagingSpaces.IsNull() {
			diags = data.StagingSpaces.ElementsAs(ctx, &spacesRelVal, false)
			diagnostics.Append(diags...)
			createSecurityGroup.Relationships["staging_spaces"] = *resource.NewToManyRelationships(spacesRelVal)

		}
		if !data.RunningSpaces.IsNull() {
			diags = data.RunningSpaces.ElementsAs(ctx, &spacesRelVal, false)
			diagnostics.Append(diags...)
			createSecurityGroup.Relationships["running_spaces"] = *resource.NewToManyRelationships(spacesRelVal)
		}
	}

	if !data.Rules.IsNull() {
		var rules []ruleType
		diags = data.Rules.ElementsAs(ctx, &rules, false)
		diagnostics.Append(diags...)
		createSecurityGroup.Rules = mapListTypeToRuleValues(rules)
	}

	return *createSecurityGroup, diagnostics
}

// Prepares a rule list resource for creation/updation from the terraform list of rule types.
func mapListTypeToRuleValues(rules []ruleType) []*resource.SecurityGroupRule {

	ruleValues := []*resource.SecurityGroupRule{}
	for _, rule := range rules {
		ruleValue := mapTypetoRuleValues(rule)
		ruleValues = append(ruleValues, ruleValue)
	}
	return ruleValues
}

// Prepares a rule resource from the terraform rule type.
func mapTypetoRuleValues(rule ruleType) *resource.SecurityGroupRule {

	securityGroupRule := resource.SecurityGroupRule{
		Protocol:    rule.Protocol.ValueString(),
		Destination: rule.Destination.ValueString(),
	}

	if !rule.Type.IsNull() {
		securityGroupRule.Type = inttointptr(int(rule.Type.ValueInt64()))
	}
	if !rule.Code.IsNull() {
		securityGroupRule.Code = inttointptr(int(rule.Code.ValueInt64()))
	}
	if !rule.Description.IsNull() {
		securityGroupRule.WithDescription(rule.Description.ValueString())
	}
	if !rule.Ports.IsNull() {
		securityGroupRule.WithPorts(rule.Ports.ValueString())
	}
	if !rule.Log.IsNull() {
		securityGroupRule.Log = rule.Log.ValueBoolPointer()
	}

	return &securityGroupRule
}

// Sets the security group resource values for updation with cf-client from the terraform struct values.
func (plan *securityGroupType) mapUpdateSecurityGroupTypeToValues(ctx context.Context) (resource.SecurityGroupUpdate, diag.Diagnostics) {

	updateSecurityGroup := &resource.SecurityGroupUpdate{
		Name: plan.Name.ValueString(),
	}

	if !plan.GloballyEnabledRunning.IsNull() || !plan.GloballyEnabledStaging.IsNull() {
		updateSecurityGroup.GloballyEnabled = &resource.SecurityGroupGloballyEnabled{}
		if !plan.GloballyEnabledRunning.IsNull() {
			updateSecurityGroup.GloballyEnabled.Running = plan.GloballyEnabledRunning.ValueBoolPointer()
		}
		if !plan.GloballyEnabledStaging.IsNull() {
			updateSecurityGroup.GloballyEnabled.Staging = plan.GloballyEnabledStaging.ValueBoolPointer()
		}
	}

	var diags, diagnostics diag.Diagnostics
	if !plan.Rules.IsNull() {
		var rules []ruleType
		diags = plan.Rules.ElementsAs(ctx, &rules, false)
		diagnostics.Append(diags...)
		updateSecurityGroup.Rules = mapListTypeToRuleValues(rules)
	} else {
		updateSecurityGroup.Rules = make([]*resource.SecurityGroupRule, 1)
	}

	return *updateSecurityGroup, diagnostics
}

func mapDataSourceSecurityGroupsValuesToType(ctx context.Context, securityGroups []*resource.SecurityGroup) ([]securityGroupType, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	securityGroupsList := []securityGroupType{}
	for _, securityGroup := range securityGroups {
		securityGroupsValue, diags := mapSecurityGroupValuesToType(ctx, securityGroup)
		diagnostics.Append(diags...)
		securityGroupsList = append(securityGroupsList, securityGroupsValue)
	}

	return securityGroupsList, diagnostics

}

func (plan *securityGroupSpacesType) mapSecurityGroupSpacesValuestoType(ctx context.Context, runningSpaces []string, stagingSpaces []string) diag.Diagnostics {
	var diags, diagnostics diag.Diagnostics
	var runningSpacesPlanned, stagingSpacesPlanned []string

	diags = plan.RunningSpaces.ElementsAs(ctx, &runningSpacesPlanned, false)
	diagnostics.Append(diags...)
	diags = plan.StagingSpaces.ElementsAs(ctx, &stagingSpacesPlanned, false)
	diagnostics.Append(diags...)
	sameRunningSpaces := lo.Intersect(runningSpaces, runningSpacesPlanned)
	sameStagingSpaces := lo.Intersect(stagingSpaces, stagingSpacesPlanned)

	if len(sameRunningSpaces) == 0 {
		plan.RunningSpaces = types.SetNull(types.StringType)
	} else {
		plan.RunningSpaces, diags = types.SetValueFrom(ctx, types.StringType, sameRunningSpaces)
		diagnostics.Append(diags...)
	}
	if len(sameStagingSpaces) == 0 {
		plan.StagingSpaces = types.SetNull(types.StringType)
	} else {
		plan.StagingSpaces, diags = types.SetValueFrom(ctx, types.StringType, sameStagingSpaces)
		diagnostics.Append(diags...)
	}

	return diagnostics
}
