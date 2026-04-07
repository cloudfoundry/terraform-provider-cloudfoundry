package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &isolationSegmentListResource{}

type isolationSegmentListResource struct {
	cfClient *cfv3client.Client
}

type isolationSegmentListResourceFilter struct {
	Org types.String `tfsdk:"org"`
}

func NewIsolationSegmentListResource() list.ListResource {
	return &isolationSegmentListResource{}
}

func (r *isolationSegmentListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_isolation_segment" // must match managed resource
}

func (r *isolationSegmentListResource) Configure(_ context.Context,
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

func (r *isolationSegmentListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all isolation segments the caller has access to, optionally filtered by organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to filter isolation segments by. Returns only isolation segments entitled to this organization.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all isolation segments from the API.
func (r *isolationSegmentListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter isolationSegmentListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	opts := cfv3client.NewIsolationSegmentOptions()

	if !filter.Org.IsNull() {
		opts.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{filter.Org.ValueString()},
		}
	}

	isolationSegments, err := r.cfClient.IsolationSegments.ListAll(ctx, opts)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Isolation Segments",
			"Could not list isolation segments: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, isoSeg := range isolationSegments {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("segment_guid"), isoSeg.GUID)

			if req.IncludeResource {
				resIsoSeg, diags := mapIsolationSegmentValuesToType(ctx, isoSeg)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resIsoSeg)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
