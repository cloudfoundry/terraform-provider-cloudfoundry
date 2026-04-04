package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &buildpackListResource{}

type buildpackListResource struct {
	cfClient *cfv3client.Client
}

type buildpackListResourceFilter struct {
	Stack types.String `tfsdk:"stack"`
}

func NewBuildpackListResource() list.ListResource {
	return &buildpackListResource{}
}

func (r *buildpackListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_buildpack" // must match managed resource
}

func (r *buildpackListResource) Configure(_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse) {

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

func (r *buildpackListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all buildpacks the user has access to, optionally filtered by name or stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				MarkdownDescription: "The name of the stack to filter buildpacks by.",
				Optional:            true,
			},
		},
	}
}

// List streams all buildpacks from the API.
func (r *buildpackListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter buildpackListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	buildpackListOptions := cfv3client.NewBuildpackListOptions()

	if !filter.Stack.IsNull() {
		buildpackListOptions.Stacks = cfv3client.Filter{
			Values: []string{filter.Stack.ValueString()},
		}
	}

	buildpacks, err := r.cfClient.Buildpacks.ListAll(ctx, buildpackListOptions)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Buildpacks",
			"Could not list buildpacks: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, buildpack := range buildpacks {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("buildpack_guid"), buildpack.GUID)

			if req.IncludeResource {
				resBuildpack, diags := mapBuildpackValuesToType(ctx, buildpack)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resBuildpack)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
