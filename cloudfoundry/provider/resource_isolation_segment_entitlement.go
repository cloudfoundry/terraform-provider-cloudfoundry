package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &IsolationSegmentEntitlementResource{}
	_ resource.ResourceWithConfigure = &IsolationSegmentEntitlementResource{}
)

// Instantiates an isolation segment resource.
func NewIsolationSegmentEntitlementResource() resource.Resource {
	return &IsolationSegmentEntitlementResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type IsolationSegmentEntitlementResource struct {
	cfClient *cfv3client.Client
}

func (r *IsolationSegmentEntitlementResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_isolation_segment_entitlement"
}

func (r *IsolationSegmentEntitlementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource for entitling and revoking isolation segment from organizations. Only entitles and revokes the segment from the orgs managed through this resource and does not touch existing entitlements. Revoking isolation segment entitlement from an organization will fail if it is the default for the organization. On deleting the resource, the isolation segment will be revoked from the mentioned orgs.",
		Attributes: map[string]schema.Attribute{
			"segment": schema.StringAttribute{
				MarkdownDescription: "GUID of the isolation segment",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"orgs": schema.SetAttribute{
				MarkdownDescription: "GUID's of organizations to entitle the segment to.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
			"default": schema.BoolAttribute{
				MarkdownDescription: "Set isolation segment as default for the organizations. Defaults to false.",
				Optional:            true,
			},
		},
	}
}

func (r *IsolationSegmentEntitlementResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IsolationSegmentEntitlementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan          IsolationSegmentEntitlementType
		orgsToEntitle []string
	)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = plan.Orgs.ElementsAs(ctx, &orgsToEntitle, false)
	resp.Diagnostics.Append(diags...)

	entitledOrgs, err := r.cfClient.IsolationSegments.EntitleOrganizations(ctx, plan.Segment.ValueString(), orgsToEntitle)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Entitling Isolation Segment to Organizations.",
			"Could not entitle Isolation Segment with ID "+plan.Segment.ValueString()+" to some organizations : "+err.Error(),
		)
		return
	}

	if !plan.Default.IsNull() {
		var segment string

		if plan.Default.ValueBool() {
			segment = plan.Segment.ValueString()
		} else {
			segment = ""
		}
		for _, org := range orgsToEntitle {
			err = r.cfClient.Organizations.AssignDefaultIsolationSegment(ctx, org, segment)
			if err != nil {
				resp.Diagnostics.AddError(
					"API Error Assigning Default Isolation Segment",
					"Could not set Isolation Segment with ID "+plan.Segment.ValueString()+" on org with ID "+org+" : "+err.Error(),
				)
			}
		}
	}

	diags = plan.mapIsolationSegmentEntitlementValuesToType(ctx, entitledOrgs.Data, plan.Default)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created an isolation segment entitlement resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *IsolationSegmentEntitlementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IsolationSegmentEntitlementType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := rs.cfClient.IsolationSegments.ListOrganizationRelationships(ctx, data.Segment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Fetching Entitled Organizations",
			"Error : "+err.Error(),
		)
		return
	}

	entitledOrgsRelations := cfv3resource.NewToManyRelationships(orgs)
	diags = data.mapIsolationSegmentEntitlementValuesToType(ctx, entitledOrgsRelations.Data, data.Default)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read an isolation segment enititlement resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *IsolationSegmentEntitlementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan, previousState IsolationSegmentEntitlementType
		entitledOrgs        *cfv3resource.IsolationSegmentRelationship
	)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removedOrgs, addedOrgs, diags := findChangedRelationsFromTFState(ctx, plan.Orgs, previousState.Orgs)
	resp.Diagnostics.Append(diags...)

	err := rs.cfClient.IsolationSegments.RevokeOrganizations(ctx, plan.Segment.ValueString(), removedOrgs)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Revoking Isolation Segment from Organizations.",
			"Could not revoke Isolation Segment with ID "+plan.Segment.ValueString()+"from some organizations : "+err.Error(),
		)
	}

	if len(addedOrgs) != 0 {
		entitledOrgs, err = rs.cfClient.IsolationSegments.EntitleOrganizations(ctx, plan.Segment.ValueString(), addedOrgs)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Entitling Isolation Segment to Organizations.",
				"Could not entitle Isolation Segment with ID "+plan.Segment.ValueString()+" to some organizations : "+err.Error(),
			)
			return
		}
	} else {
		orgs, err := rs.cfClient.IsolationSegments.ListOrganizationRelationships(ctx, plan.Segment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Entitled Organizations",
				"Error : "+err.Error(),
			)
			return
		}
		entitledOrgsRelations := cfv3resource.NewToManyRelationships(orgs)
		entitledOrgs = &cfv3resource.IsolationSegmentRelationship{
			Data: entitledOrgsRelations.Data,
		}
	}

	if !plan.Default.IsNull() {
		var (
			segment       string
			orgsToEntitle []string
		)

		if plan.Default.ValueBool() {
			segment = plan.Segment.ValueString()
		} else {
			segment = ""
		}

		diags = plan.Orgs.ElementsAs(ctx, &orgsToEntitle, false)
		resp.Diagnostics.Append(diags...)

		for _, org := range orgsToEntitle {
			err = rs.cfClient.Organizations.AssignDefaultIsolationSegment(ctx, org, segment)
			if err != nil {
				resp.Diagnostics.AddError(
					"API Error Assigning Default Isolation Segment",
					"Could not set Isolation Segment with ID "+plan.Segment.ValueString()+" on org with ID "+org+" : "+err.Error(),
				)
			}
		}

	}

	diags = plan.mapIsolationSegmentEntitlementValuesToType(ctx, entitledOrgs.Data, plan.Default)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated an isolation segment entitlement resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *IsolationSegmentEntitlementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		state        IsolationSegmentEntitlementType
		orgsToRevoke []string
	)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = state.Orgs.ElementsAs(ctx, &orgsToRevoke, false)
	resp.Diagnostics.Append(diags...)

	err := rs.cfClient.IsolationSegments.RevokeOrganizations(ctx, state.Segment.ValueString(), orgsToRevoke)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Revoking Isolation Segment from Organizations.",
			"Could not revoke Isolation Segment with ID "+state.Segment.ValueString()+"from some organizations : "+err.Error(),
		)
	}

	tflog.Trace(ctx, "deleted an isolation segment entitlement resource")
}
