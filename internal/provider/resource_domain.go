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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithConfigure   = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

// Instantiates a domain resource.
func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type DomainResource struct {
	cfClient *cfv3client.Client
}

func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a resource for managing shared or private domains in Cloud Foundry.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the domain;must be between 3 ~ 253 characters and follow [RFC 1035](https://tools.ietf.org/html/rfc1035)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Whether the domain is used for internal (container-to-container) traffic, or external (user-to-container) traffic",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"router_group": schema.StringAttribute{
				MarkdownDescription: "The desired router group guid. note: creates a tcp domain; cannot be used when internal is set to true or domain is scoped to an org",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The organization the domain is scoped to; if set, the domain will only be available in that organization; otherwise, the domain will be globally available",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"shared_orgs": schema.SetAttribute{
				MarkdownDescription: "Organizations the domain is shared with; if set, the domain will be available in these organizations in addition to the organization the domain is scoped to",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.ValidUUID()),
					setvalidator.SizeAtLeast(1),
					setvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("org"),
					}...),
				},
			},
			"supported_protocols": schema.SetAttribute{
				MarkdownDescription: "Available protocols for routes using the domain, currently http and tcp",
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				ElementType: types.StringType,
			},
			idKey:          guidSchema(),
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (r *DomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan domainType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createDomain, diags := plan.mapCreateDomainTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.cfClient.Domains.Create(ctx, &createDomain)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Domain",
			"Could not create Domain "+plan.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	plan, diags = mapDomainValuesToType(ctx, domain)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a domain resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data domainType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := rs.cfClient.Domains.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "domain", data.Id.ValueString())
		return
	}

	data, diags = mapDomainValuesToType(ctx, domain)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a domain resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState domainType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removedSharedOrgs, addedSharedOrgs, diags := findChangedRelationsFromTFState(ctx, plan.SharedOrgs, previousState.SharedOrgs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	for _, org := range removedSharedOrgs {
		err = rs.cfClient.Domains.UnShare(ctx, plan.Id.ValueString(), org)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Unsharing Domain",
				"Could not remove org with ID "+org+" from the list of shared organizations for domain with ID "+plan.Id.ValueString()+" : "+err.Error(),
			)
		}
	}

	if len(addedSharedOrgs) > 0 {
		orgs := cfv3resource.NewToManyRelationships(addedSharedOrgs)
		_, err = rs.cfClient.Domains.ShareMany(ctx, plan.Id.ValueString(), orgs)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Sharing Domain",
				"Could not add shared organizations for domain with ID "+plan.Id.ValueString()+" : "+err.Error(),
			)
		}
	}

	updateDomain, diags := plan.mapUpdateDomainTypeToValues(ctx, previousState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := rs.cfClient.Domains.Update(ctx, plan.Id.ValueString(), &updateDomain)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Domain",
			"Could not update domain with ID "+plan.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data, diags := mapDomainValuesToType(ctx, domain)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a domain resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (rs *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state domainType

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.Domains.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Domain",
			"Could not delete the domain with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Domain",
			"Failed in deleting the domain with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a domain resource")

}

func (rs *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
