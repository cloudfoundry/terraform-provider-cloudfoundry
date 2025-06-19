package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &SecurityGroupSpacesResource{}
	_ resource.ResourceWithConfigure = &SecurityGroupSpacesResource{}
)

// Instantiates an isolation segment resource.
func NewSecurityGroupSpacesResource() resource.Resource {
	return &SecurityGroupSpacesResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SecurityGroupSpacesResource struct {
	cfClient *cfv3client.Client
}

func (r *SecurityGroupSpacesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_space_bindings"
}

func (r *SecurityGroupSpacesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource for binding and unbinding a security group from spaces. Only handles bindings managed through this resource and does not touch the existing space bindings with the security group. On deleting the resource, the security group will be unbound from the mentioned spaces.",
		Attributes: map[string]schema.Attribute{
			"security_group": schema.StringAttribute{
				MarkdownDescription: "GUID of the isolation segment",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"running_spaces": schema.SetAttribute{
				MarkdownDescription: "The spaces where the security_group is applied to applications during runtime",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
					setvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("running_spaces"),
						path.MatchRoot("staging_spaces"),
					}...),
				},
			},
			"staging_spaces": schema.SetAttribute{
				MarkdownDescription: "The spaces where the security_group is applied to applications during staging",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (r *SecurityGroupSpacesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityGroupSpacesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan securityGroupSpacesType
		err  error
	)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RunningSpaces.IsNull() {
		var addedRunningSpaces []string
		diags = plan.RunningSpaces.ElementsAs(ctx, &addedRunningSpaces, false)
		resp.Diagnostics.Append(diags...)
		_, err = r.cfClient.SecurityGroups.BindRunningSecurityGroup(ctx, plan.SecurityGroup.ValueString(), addedRunningSpaces)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Binding Running Security Group",
				"Could not bind space to the Security Group with ID "+plan.SecurityGroup.ValueString()+" : "+err.Error(),
			)
			return
		}

	}

	if !plan.StagingSpaces.IsNull() {
		var addedStagingSpaces []string
		diags = plan.StagingSpaces.ElementsAs(ctx, &addedStagingSpaces, false)
		resp.Diagnostics.Append(diags...)
		_, err = r.cfClient.SecurityGroups.BindStagingSecurityGroup(ctx, plan.SecurityGroup.ValueString(), addedStagingSpaces)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Binding Staging Security Group",
				"Could not bind space to the Security Group with ID "+plan.SecurityGroup.ValueString()+" : "+err.Error(),
			)
			return
		}
	}

	securityGroup, err := r.cfClient.SecurityGroups.Get(ctx, plan.SecurityGroup.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Security Group",
			"Error : "+err.Error(),
		)
		return
	}
	runningSpaces := setRelationshipToSlice(securityGroup.Relationships.RunningSpaces.Data)
	stagingSpaces := setRelationshipToSlice(securityGroup.Relationships.StagingSpaces.Data)

	diags = plan.mapSecurityGroupSpacesValuestoType(ctx, runningSpaces, stagingSpaces)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a security group spaces resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *SecurityGroupSpacesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data securityGroupSpacesType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	securityGroup, err := rs.cfClient.SecurityGroups.Get(ctx, data.SecurityGroup.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Security Group",
			"Error : "+err.Error(),
		)
		return
	}

	runningSpaces := setRelationshipToSlice(securityGroup.Relationships.RunningSpaces.Data)
	stagingSpaces := setRelationshipToSlice(securityGroup.Relationships.StagingSpaces.Data)
	diags = data.mapSecurityGroupSpacesValuestoType(ctx, runningSpaces, stagingSpaces)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a security group spaces resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *SecurityGroupSpacesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan          securityGroupSpacesType
		previousState securityGroupSpacesType
		err           error
	)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removedRunningSpaces, addedRunningSpaces, diags := findChangedRelationsFromTFState(ctx, plan.RunningSpaces, previousState.RunningSpaces)
	resp.Diagnostics.Append(diags...)

	for _, space := range removedRunningSpaces {
		err = rs.cfClient.SecurityGroups.UnBindRunningSecurityGroup(ctx, plan.SecurityGroup.ValueString(), space)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unbinding Running Security Group",
				"Could not remove space with ID "+space+" from the Security Group with ID "+plan.SecurityGroup.ValueString()+" : "+err.Error(),
			)
		}
	}

	if len(addedRunningSpaces) > 0 {
		_, err = rs.cfClient.SecurityGroups.BindRunningSecurityGroup(ctx, plan.SecurityGroup.ValueString(), addedRunningSpaces)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Binding Running Security Group",
				"Could not bind space to the Security Group with ID "+plan.SecurityGroup.ValueString()+" : "+err.Error(),
			)
		}
	}

	removedStagingSpaces, addedStagingSpaces, diags := findChangedRelationsFromTFState(ctx, plan.StagingSpaces, previousState.StagingSpaces)
	resp.Diagnostics.Append(diags...)

	for _, space := range removedStagingSpaces {
		err = rs.cfClient.SecurityGroups.UnBindStagingSecurityGroup(ctx, plan.SecurityGroup.ValueString(), space)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unbinding Staging Security Group",
				"Could not remove space with ID "+space+" from the Security Group with ID "+plan.SecurityGroup.ValueString()+" : "+err.Error(),
			)
		}
	}

	if len(addedStagingSpaces) > 0 {
		_, err = rs.cfClient.SecurityGroups.BindStagingSecurityGroup(ctx, plan.SecurityGroup.ValueString(), addedStagingSpaces)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Binding Staging Security Group",
				"Could not bind space to the Security Group with ID "+plan.SecurityGroup.ValueString()+" : "+err.Error(),
			)
		}
	}

	securityGroup, err := rs.cfClient.SecurityGroups.Get(ctx, plan.SecurityGroup.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Security Group",
			"Error : "+err.Error(),
		)
		return
	}

	runningSpaces := setRelationshipToSlice(securityGroup.Relationships.RunningSpaces.Data)
	stagingSpaces := setRelationshipToSlice(securityGroup.Relationships.StagingSpaces.Data)
	diags = plan.mapSecurityGroupSpacesValuestoType(ctx, runningSpaces, stagingSpaces)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a security group spaces resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *SecurityGroupSpacesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		state         securityGroupSpacesType
		runningSpaces []string
		stagingSpaces []string
		err           error
	)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = state.RunningSpaces.ElementsAs(ctx, &runningSpaces, false)
	resp.Diagnostics.Append(diags...)

	diags = state.StagingSpaces.ElementsAs(ctx, &stagingSpaces, false)
	resp.Diagnostics.Append(diags...)

	for _, space := range runningSpaces {
		err = rs.cfClient.SecurityGroups.UnBindRunningSecurityGroup(ctx, state.SecurityGroup.ValueString(), space)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unbinding Running Security Group",
				"Could not remove space with ID "+space+" from the Security Group with ID "+state.SecurityGroup.ValueString()+" : "+err.Error(),
			)
		}
	}

	for _, space := range stagingSpaces {
		err = rs.cfClient.SecurityGroups.UnBindStagingSecurityGroup(ctx, state.SecurityGroup.ValueString(), space)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unbinding Staging Security Group",
				"Could not remove space with ID "+space+" from the Security Group with ID "+state.SecurityGroup.ValueString()+" : "+err.Error(),
			)
		}
	}

	tflog.Trace(ctx, "deleted a security group spaces resource")
}
