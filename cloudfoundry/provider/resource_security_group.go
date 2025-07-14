package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &SecurityGroupResource{}
	_ resource.ResourceWithConfigure   = &SecurityGroupResource{}
	_ resource.ResourceWithImportState = &SecurityGroupResource{}
)

// Instantiates a security group resource.
func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type SecurityGroupResource struct {
	cfClient *cfv3client.Client
}

func (r *SecurityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an application security group resource for Cloud Foundry. This resource defines egress rules that can be applied to containers that stage and run applications.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the security group",
				Required:            true,
			},
			"rules": schema.ListNestedAttribute{
				MarkdownDescription: "Rules that will be applied by this security group",
				Optional:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol type. Valid values are tcp, udp, icmp, or all",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp", "icmp", "all"),
							},
						},
						"destination": schema.StringAttribute{
							MarkdownDescription: "Destinations that the rule applies to. Valid CIDR, IP address, or IP address range",
							Required:            true,
						},
						"ports": schema.StringAttribute{
							MarkdownDescription: "Ports that the rule applies to; can be a single port (9000), a comma-separated list (9000,9001), or a range (9000-9200)",
							Optional:            true,
						},
						"type": schema.Int64Attribute{
							MarkdownDescription: "[Type](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml#icmp-parameters-types) required for ICMP protocol; valid values are between -1 and 255 (inclusive), where -1 allows all",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(-1, 255),
							},
						},
						"code": schema.Int64Attribute{
							MarkdownDescription: "[Code](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml#icmp-parameters-codes) required for ICMP protocol; valid values are between -1 and 255 (inclusive), where -1 allows all",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(-1, 255),
							},
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "A description for the rule",
							Optional:            true,
						},
						"log": schema.BoolAttribute{
							MarkdownDescription: "Enable logging for rule",
							Optional:            true,
						},
					},
				},
			},
			"globally_enabled_running": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the group should be applied globally to all running applications",
				Optional:            true,
				Computed:            true,
			},
			"globally_enabled_staging": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the group should be applied globally to all staging applications",
				Optional:            true,
				Computed:            true,
			},
			"running_spaces": schema.SetAttribute{
				MarkdownDescription: "The spaces where the security_group is applied to applications during runtime",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
			"staging_spaces": schema.SetAttribute{
				MarkdownDescription: "The spaces where the security_group is applied to applications during staging",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createSecurityGroup, diags := plan.mapCreateSecurityGroupTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	securityGroup, err := r.cfClient.SecurityGroups.Create(ctx, &createSecurityGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Security Group",
			"Could not create Security Group with name "+plan.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	plan, diags = mapSecurityGroupValuesToType(ctx, securityGroup)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a security group resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data securityGroupType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	securityGroup, err := rs.cfClient.SecurityGroups.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "security group", data.Id.ValueString())
		return
	}

	data, diags = mapSecurityGroupValuesToType(ctx, securityGroup)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a security group resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan, previousState                                                                securityGroupType
		removedRunningSpaces, addedRunningSpaces, removedStagingSpaces, addedStagingSpaces []string
		err                                                                                error
	)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(previousState.RunningSpaces.ElementsAs(ctx, &removedRunningSpaces, false)...)

	for _, space := range removedRunningSpaces {
		err = rs.cfClient.SecurityGroups.UnBindRunningSecurityGroup(ctx, plan.Id.ValueString(), space)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unbinding Running Security Group",
				"Could not remove space with ID "+space+" from the Security Group with ID "+plan.Id.ValueString()+" : "+err.Error(),
			)
		}
	}

	if !plan.RunningSpaces.IsUnknown() {
		resp.Diagnostics.Append(plan.RunningSpaces.ElementsAs(ctx, &addedRunningSpaces, false)...)

		_, err = rs.cfClient.SecurityGroups.BindRunningSecurityGroup(ctx, plan.Id.ValueString(), addedRunningSpaces)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Binding Running Security Group",
				"Could not bind space to the Security Group with ID "+plan.Id.ValueString()+" : "+err.Error(),
			)
		}
	}

	resp.Diagnostics.Append(previousState.StagingSpaces.ElementsAs(ctx, &removedStagingSpaces, false)...)

	for _, space := range removedStagingSpaces {
		err = rs.cfClient.SecurityGroups.UnBindStagingSecurityGroup(ctx, plan.Id.ValueString(), space)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unbinding Staging Security Group",
				"Could not remove space with ID "+space+" from the Security Group with ID "+plan.Id.ValueString()+" : "+err.Error(),
			)
		}
	}

	if !plan.StagingSpaces.IsUnknown() {
		resp.Diagnostics.Append(plan.StagingSpaces.ElementsAs(ctx, &addedStagingSpaces, false)...)

		_, err = rs.cfClient.SecurityGroups.BindStagingSecurityGroup(ctx, plan.Id.ValueString(), addedStagingSpaces)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Binding Staging Security Group",
				"Could not bind space to the Security Group with ID "+plan.Id.ValueString()+" : "+err.Error(),
			)
		}
	}

	updateSecurityGroup, diags := plan.mapUpdateSecurityGroupTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	securityGroup, err := rs.cfClient.SecurityGroups.Update(ctx, plan.Id.ValueString(), &updateSecurityGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Security Group",
			"Could not update Security Group with ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	data, diags := mapSecurityGroupValuesToType(ctx, securityGroup)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a security group resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.SecurityGroups.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Security Group",
			"Could not delete the Security Group with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Security Group",
			"Failed in deleting the Security Group with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a security group resource")
}

func (rs *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
