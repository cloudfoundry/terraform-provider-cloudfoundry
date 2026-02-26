package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &orgResource{}
	_ resource.ResourceWithConfigure   = &orgResource{}
	_ resource.ResourceWithImportState = &orgResource{}
	_ resource.ResourceWithIdentity    = &orgResource{}
)

// NewOrgResource is a helper function to simplify the provider implementation.
func NewOrgResource() resource.Resource {
	return &orgResource{}
}

// orgResource is the resource implementation.
type orgResource struct {
	cfClient *cfv3client.Client
}

type OrgResouerceIdentityModel struct {
	OrgGUID types.String `tfsdk:"org_guid"`
}

// Metadata returns the resource type name.
func (r *orgResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

// Configure defines the configuration for the resource.
func (r *orgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

// Schema defines the schema for the resource.
func (r *orgResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates a Cloud Foundry Organization 
		
		__Further documentation:__
		https://docs.cloudfoundry.org/concepts/roles.html#orgs
		`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Organization in Cloud Foundry",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Organization",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
			},
			"labels":      resourceLabelsSchema(),
			"annotations": resourceAnnotationsSchema(),
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
				Computed:            true,
			},
			"suspended": schema.BoolAttribute{
				MarkdownDescription: "Whether an organization is suspended or not.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"quota": schema.StringAttribute{
				MarkdownDescription: "The ID of quota to be applied to this Org. Default quota is assigned to the org by default.",
				Computed:            true,
			},
		},
	}

}

func (rs *orgResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"org_guid": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *orgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan orgType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createOrg := cfv3resource.OrganizationCreate{
		Name:     plan.Name.ValueString(),
		Metadata: cfv3resource.NewMetadata(),
	}

	if !plan.Suspended.IsNull() {
		createOrg.Suspended = plan.Suspended.ValueBoolPointer()
	}
	labelsDiags := plan.Labels.ElementsAs(ctx, &createOrg.Metadata.Labels, false)
	resp.Diagnostics.Append(labelsDiags...)

	annotationsDiags := plan.Annotations.ElementsAs(ctx, &createOrg.Metadata.Annotations, false)
	resp.Diagnostics.Append(annotationsDiags...)

	org, err := r.cfClient.Organizations.Create(ctx, &createOrg)

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Organization",
			"Could not create Organization "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	plan, diags = mapOrgValuesToType(ctx, org)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := OrgResouerceIdentityModel{
		OrgGUID: types.StringValue(plan.ID.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *orgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data orgType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	orglo := cfv3client.NewOrganizationListOptions()
	orglo.GUIDs = cfv3client.Filter{
		Values: []string{
			data.ID.ValueString(),
		},
	}
	orgs, err := r.cfClient.Organizations.ListAll(ctx, orglo)
	if err != nil {
		handleReadErrors(ctx, resp, err, "org", data.ID.ValueString())
		return
	}
	org, found := lo.Find(orgs, func(org *cfv3resource.Organization) bool {
		return org.GUID == data.ID.ValueString()
	})
	if !found {
		resp.Diagnostics.AddError(
			"Unable to find org data in list",
			fmt.Sprintf("Given ID %s not in the list of orgs.", data.ID.ValueString()),
		)
		return
	}
	data, diags = mapOrgValuesToType(ctx, org)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var identity OrgResouerceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = OrgResouerceIdentityModel{
			OrgGUID: types.StringValue(data.ID.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *orgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState orgType
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)

	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.cfClient.Organizations.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch org Data",
			"Could not get Org with ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	updateOrg := cfv3resource.OrganizationUpdate{
		Name:      plan.Name.ValueString(),
		Suspended: plan.Suspended.ValueBoolPointer(),
		Metadata:  cfv3resource.NewMetadata(),
	}

	updateOrg.Metadata, diags = setClientMetadataForUpdate(ctx, previousState.Labels, previousState.Annotations, plan.Labels, plan.Annotations)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.cfClient.Organizations.Update(ctx, plan.ID.ValueString(), &updateOrg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update Org",
			"Could not update org with ID "+plan.ID.ValueString()+" and name "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	plan, diags = mapOrgValuesToType(ctx, org)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// WORKAROUND for OpenTofu compatibility
	// https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/418
	identity := OrgResouerceIdentityModel{
		OrgGUID: types.StringValue(plan.ID.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
	// END WORKAROUND
}

// Delete the resource and removes the Terraform state on success.
func (r *orgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state orgType

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	jobID, err := r.cfClient.Organizations.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete org",
			"API Error in deleting organization"+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *r.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Delete org failed",
			"Failed in Deleting organization "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

}

func (r *orgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("org_guid"), req, resp)
}
