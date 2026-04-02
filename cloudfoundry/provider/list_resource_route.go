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

var _ list.ListResourceWithConfigure = &routeListResource{}

type routeListResource struct {
	cfClient *cfv3client.Client
}

type routeListResourceFilter struct {
	Space  types.String `tfsdk:"space"`
	Domain types.String `tfsdk:"domain"`
	Org    types.String `tfsdk:"org"`
}

func NewRouteListResource() list.ListResource {
	return &routeListResource{}
}

func (r *routeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route" // must match managed resource
}

func (r *routeListResource) Configure(_ context.Context,
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

func (r *routeListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all routes the user has access to, optionally filtered by space, domain, or organization.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to filter routes by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The GUID of the domain to filter routes by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to filter routes by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all routes from the API.
func (r *routeListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter routeListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	if !filter.Space.IsNull() {
		_, err := r.cfClient.Spaces.Get(ctx, filter.Space.ValueString())
		if err != nil {
			var diags diag.Diagnostics
			diags.AddError(
				"API Error Fetching Space",
				"Could not get space with ID "+filter.Space.ValueString()+": "+err.Error(),
			)
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	if !filter.Org.IsNull() {
		_, err := r.cfClient.Organizations.Get(ctx, filter.Org.ValueString())
		if err != nil {
			var diags diag.Diagnostics
			diags.AddError(
				"API Error Fetching Organization",
				"Could not get organization with ID "+filter.Org.ValueString()+": "+err.Error(),
			)
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	routeListOptions := cfv3client.NewRouteListOptions()

	if !filter.Space.IsNull() {
		routeListOptions.SpaceGUIDs = cfv3client.Filter{
			Values: []string{filter.Space.ValueString()},
		}
	}
	if !filter.Domain.IsNull() {
		routeListOptions.DomainGUIDs = cfv3client.Filter{
			Values: []string{filter.Domain.ValueString()},
		}
	}
	if !filter.Org.IsNull() {
		routeListOptions.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{filter.Org.ValueString()},
		}
	}

	routes, err := r.cfClient.Routes.ListAll(ctx, routeListOptions)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Routes",
			"Could not list routes: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, route := range routes {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("route_guid"), route.GUID)

			if req.IncludeResource {
				resRoute, diags := mapRouteValuesToType(ctx, route)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resRoute)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
