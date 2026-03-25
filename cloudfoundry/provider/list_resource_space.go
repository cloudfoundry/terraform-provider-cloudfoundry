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

var _ list.ListResourceWithConfigure = &spaceListResource{}

type spaceListResource struct {
	cfClient *cfv3client.Client
}

type spaceListResourceFilter struct {
	Org types.String `tfsdk:"org"`
}

func NewSpaceListResource() list.ListResource {
	return &spaceListResource{}
}

func (r *spaceListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space" // must match managed resource
}

func (r *spaceListResource) Configure(_ context.Context,
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

func (r *spaceListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all spaces within an organization.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The guid of the organization.",
				Required:            true,
			},
		},
	}
}

// List streams all space within an organization from the API
func (r *spaceListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {

	var (
		filter spaceListResourceFilter
	)

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	_, err := r.cfClient.Organizations.Get(ctx, filter.Org.ValueString())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Organization",
			"Could not get details of the Organization with ID "+filter.Org.ValueString()+" : "+err.Error(),
		)

		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	spacesListOptions := cfv3client.NewSpaceListOptions()
	spacesListOptions.OrganizationGUIDs = cfv3client.Filter{
		Values: []string{
			filter.Org.ValueString(),
		},
	}

	spaces, err := r.cfClient.Spaces.ListAll(ctx, spacesListOptions)
	if err != nil {

		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Spaces",
			"Could not get spaces under Organization with ID "+filter.Org.ValueString()+" : "+err.Error(),
		)

		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {

		for _, space := range spaces {

			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("space_guid"), space.GUID)

			if req.IncludeResource {

				sshEnabled, err := r.cfClient.SpaceFeatures.IsSSHEnabled(ctx, space.GUID)
				if err != nil {
					var diags diag.Diagnostics
					diags.AddError(
						"API Error Fetching SSH Feature",
						"Could not get the SSH feature value of space "+space.Name+" : "+err.Error(),
					)

					stream.Results = list.ListResultsStreamDiagnostics(diags)
					return
				}

				isolationSegment, err := r.cfClient.Spaces.GetAssignedIsolationSegment(ctx, space.GUID)
				if err != nil {
					var diags diag.Diagnostics
					diags.AddError(
						"API Error Fetching Isolation Segment",
						"Could not get the Isolation Segment of space "+space.Name+": "+err.Error(),
					)

					stream.Results = list.ListResultsStreamDiagnostics(diags)
					return
				}
				resSpace, diags := mapSpaceValuesToType(ctx, space, sshEnabled, isolationSegment)

				result.Diagnostics.Append(diags...)

				// Set the resource information on the result
				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resSpace)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
