package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
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
	"github.com/samber/lo"
)

var (
	_ resource.Resource              = &spaceQuotaResource{}
	_ resource.ResourceWithConfigure = &spaceQuotaResource{}
)

func NewSpaceQuotaResource() resource.Resource {
	return &spaceQuotaResource{}
}

type spaceQuotaResource struct {
	cfClient *cfv3client.Client
}

func (r *spaceQuotaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_quota"
}

func (r *spaceQuotaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource to manage space quota definitions.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name you use to identify the quota or plan in Cloud Foundry",
				Required:            true,
			},
			"allow_paid_service_plans": schema.BoolAttribute{
				MarkdownDescription: "Determines whether users can provision instances of non-free service plans. Does not control plan visibility. When false, non-free service plans may be visible in the marketplace but instances can not be provisioned.",
				Required:            true,
			},
			"total_services": schema.Int64Attribute{
				MarkdownDescription: "Maximum services allowed.",
				Optional:            true,
			},
			"total_service_keys": schema.Int64Attribute{
				MarkdownDescription: "Maximum service keys allowed.",
				Optional:            true,
			},
			"total_routes": schema.Int64Attribute{
				MarkdownDescription: "Maximum routes allowed.",
				Optional:            true,
			},
			"total_route_ports": schema.Int64Attribute{
				MarkdownDescription: "Maximum routes with reserved ports.",
				Optional:            true,
			},
			"total_memory": schema.Int64Attribute{
				MarkdownDescription: "Maximum memory usage allowed.",
				Optional:            true,
			},
			"instance_memory": schema.Int64Attribute{
				MarkdownDescription: "Maximum memory per application instance.",
				Optional:            true,
			},
			"total_app_instances": schema.Int64Attribute{
				MarkdownDescription: "Maximum app instances allowed.",
				Optional:            true,
			},
			"total_app_tasks": schema.Int64Attribute{
				MarkdownDescription: "Maximum tasks allowed per app.",
				Optional:            true,
			},
			"total_app_log_rate_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum log rate allowed for all the started processes and running tasks in bytes/second.",
				Optional:            true,
			},
			"spaces": schema.SetAttribute{
				MarkdownDescription: "Set of space GUIDs to which this space quota would be assigned.",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
				},
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The ID of the Org within which to create the space quota",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			idKey:        guidSchema(),
			createdAtKey: createdAtSchema(),
			updatedAtKey: updatedAtSchema(),
		},
	}
}
func (r *spaceQuotaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}
	r.cfClient = session.CFClient
}

func (r *spaceQuotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var spaceQuotaType spaceQuotaType
	diags := req.Plan.Get(ctx, &spaceQuotaType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spacesQuotaValues, diags := spaceQuotaType.mapSpaceQuotaTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	spacesQuotaResp, err := r.cfClient.SpaceQuotas.Create(ctx, spacesQuotaValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create space quota",
			fmt.Sprintf("Request failed with %s ", err.Error()),
		)
		return
	}
	spacesQuotaType, diags := mapSpaceQuotaValuesToType(spacesQuotaResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, spacesQuotaType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceQuotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var spaceQuotaTypeState spaceQuotaType
	diags := req.State.Get(ctx, &spaceQuotaTypeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spqulo := cfv3client.NewSpaceQuotaListOptions()
	spqulo.GUIDs = cfv3client.Filter{
		Values: []string{
			spaceQuotaTypeState.ID.ValueString(),
		},
	}
	spacesQuotas, err := r.cfClient.SpaceQuotas.ListAll(ctx, spqulo)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch space quota data",
			fmt.Sprintf("Request failed with %s", err.Error()),
		)
		return
	}
	spacesQuota, found := lo.Find(spacesQuotas, func(spaceQuota *cfv3resource.SpaceQuota) bool {
		return spaceQuota.GUID == spaceQuotaTypeState.ID.ValueString()
	})
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	spacesQuotaType, diags := mapSpaceQuotaValuesToType(spacesQuota)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, spacesQuotaType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceQuotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var spaceQuotaTypePlan spaceQuotaType
	var spaceQuotaTypeState spaceQuotaType
	diags := req.Plan.Get(ctx, &spaceQuotaTypePlan)
	resp.Diagnostics.Append(diags...)
	diags = resp.State.Get(ctx, &spaceQuotaTypeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removed, added, diags := findChangedRelationsFromTFState(ctx, spaceQuotaTypePlan.Spaces, spaceQuotaTypeState.Spaces)
	resp.Diagnostics.Append(diags...)
	spacesQuotaValues, diags := spaceQuotaTypePlan.mapSpaceQuotaTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(removed) != 0 {
		for _, spaceId := range removed {
			err := r.cfClient.SpaceQuotas.Remove(ctx, spaceQuotaTypePlan.ID.ValueString(), spaceId)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to update space quota",
					fmt.Sprintf("Request failed with %s", err.Error()),
				)
				return
			}
		}
	}
	if len(added) != 0 {
		_, err := r.cfClient.SpaceQuotas.Apply(ctx, spaceQuotaTypePlan.ID.ValueString(), added)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update space quota",
				fmt.Sprintf("Request failed with %s", err.Error()),
			)
			return
		}
	}
	spacesQuotaValues.Relationships = nil
	spacesQuotaResp, err := r.cfClient.SpaceQuotas.Update(ctx, spaceQuotaTypePlan.ID.ValueString(), spacesQuotaValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update space quota",
			fmt.Sprintf("Request failed with %s", err.Error()),
		)
		return
	}
	spacesQuotaType, diags := mapSpaceQuotaValuesToType(spacesQuotaResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, spacesQuotaType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceQuotaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var spaceQuotaType spaceQuotaType
	diags := req.State.Get(ctx, &spaceQuotaType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	jobID, err := r.cfClient.SpaceQuotas.Delete(ctx, spaceQuotaType.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete space quota",
			fmt.Sprintf("space quota deletion verification failed %s with %s", spaceQuotaType.Name.ValueString(), err.Error()),
		)
		return
	}
	if err = pollJob(ctx, *r.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Unable to verify space quota deletion",
			"space quota deletion verification failed for "+spaceQuotaType.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *spaceQuotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
