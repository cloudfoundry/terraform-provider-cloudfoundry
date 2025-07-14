package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure   = &BuildpackResource{}
	_ resource.ResourceWithImportState = &BuildpackResource{}
)

// Instantiates a security group resource.
func NewBuildpackResource() resource.Resource {
	return &BuildpackResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type BuildpackResource struct {
	cfClient *cfv3client.Client
}

func (r *BuildpackResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_buildpack"
}

func (r *BuildpackResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource for managing Cloud Foundry buildpacks.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the buildpack",
				Required:            true,
			},
			"stack": schema.StringAttribute{
				MarkdownDescription: "The name of the stack that the buildpack will use",
				Optional:            true,
			},
			"position": schema.Int64Attribute{
				MarkdownDescription: "The order in which the buildpacks are checked during buildpack auto-detection",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether or not the buildpack can be used for staging",
				Optional:            true,
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether or not the buildpack is locked to prevent updating the bits",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the buildpack",
				Computed:            true,
			},
			"filename": schema.StringAttribute{
				MarkdownDescription: "The filename of the buildpack",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Path of the zip file for the buildpack",
				Optional:            true,
			},
			"source_code_hash": schema.StringAttribute{
				MarkdownDescription: "SHA256 hash of the file specified. Terraform relies on this to detect the file changes.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("path"),
					}...),
				},
			},

			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (r *BuildpackResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BuildpackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan  buildpackType
		jobID string
	)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createBuildpack, diags := plan.mapCreateBuildpackTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)

	buildpack, err := r.cfClient.Buildpacks.Create(ctx, &createBuildpack)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Buildpack",
			"Could not create Buildpack with name "+plan.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	if !plan.Path.IsNull() {
		file, err := os.Open(plan.Path.ValueString())
		fileName := filepath.Base(plan.Path.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid file or Path given!",
				"Unable to open file : "+err.Error(),
			)
		}
		jobID, _, err = r.cfClient.Buildpacks.Upload(ctx, buildpack.GUID, fileName, file)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Uploading Buildpack",
				"Could not upload buildpack with file "+fileName+" : "+err.Error(),
			)
		}
		if jobID != "" {
			if err = pollJob(ctx, *r.cfClient, jobID, defaultTimeout); err != nil {
				resp.Diagnostics.AddError(
					"API Error Uploading Buildpack",
					"Failed in uploading the Buildpack with file "+fileName+" : "+err.Error(),
				)
			}
		}
		buildpack, err = r.cfClient.Buildpacks.Get(ctx, buildpack.GUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Buildpack",
				"Failed in reading the Buildpack with name "+plan.Name.ValueString()+" : "+err.Error(),
			)
		}
	}

	data, diags := mapBuildpackValuesToType(ctx, buildpack)
	resp.Diagnostics.Append(diags...)
	data.Path = plan.Path
	data.SourceCodeHash = plan.SourceCodeHash

	tflog.Trace(ctx, "created a buildpack resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *BuildpackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data buildpackType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildpack, err := rs.cfClient.Buildpacks.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "buildpack", data.Id.ValueString())
		return
	}

	state, diags := mapBuildpackValuesToType(ctx, buildpack)
	resp.Diagnostics.Append(diags...)
	state.Path = data.Path
	state.SourceCodeHash = data.SourceCodeHash

	tflog.Trace(ctx, "read a buildpack resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (rs *BuildpackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState buildpackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateBuildpack, diags := plan.mapUpdateBuildpackTypeToValues(ctx, previousState)
	resp.Diagnostics.Append(diags...)

	buildpack, err := rs.cfClient.Buildpacks.Update(ctx, plan.Id.ValueString(), &updateBuildpack)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Buildpack",
			"Could not update Buildpack with ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	if !plan.Path.IsNull() && (plan.Path.ValueString() != previousState.Path.ValueString() || plan.SourceCodeHash.ValueString() != previousState.SourceCodeHash.ValueString()) {
		file, err := os.Open(plan.Path.ValueString())
		fileName := filepath.Base(plan.Path.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid file or Path given!",
				"Unable to open file : "+err.Error(),
			)
			return
		}
		jobID, _, err := rs.cfClient.Buildpacks.Upload(ctx, plan.Id.ValueString(), fileName, file)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Uploading Buildpack",
				"Could not upload buildpack with file "+fileName+" : "+err.Error(),
			)
			return
		}
		if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
			resp.Diagnostics.AddError(
				"API Error Uploading Buildpack",
				"Failed in uploading the Buildpack with file "+fileName+" : "+err.Error(),
			)
		}
		buildpack, err = rs.cfClient.Buildpacks.Get(ctx, buildpack.GUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Buildpack",
				"Failed in reading the Buildpack with name "+plan.Name.ValueString()+" : "+err.Error(),
			)
		}
	}

	data, diags := mapBuildpackValuesToType(ctx, buildpack)
	resp.Diagnostics.Append(diags...)
	data.Path = plan.Path
	data.SourceCodeHash = plan.SourceCodeHash

	tflog.Trace(ctx, "updated a buildpack resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *BuildpackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state buildpackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.Buildpacks.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Buildpack",
			"Could not delete the Buildpack with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Buildpack",
			"Failed in deleting the Buildpack with ID "+state.Id.ValueString()+" and name "+state.Name.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a buildpack resource")
}

func (rs *BuildpackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
