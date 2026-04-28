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

var _ list.ListResourceWithConfigure = &serviceRouteBindingListResource{}

type serviceRouteBindingListResource struct {
	cfClient *cfv3client.Client
}

type serviceRouteBindingListResourceFilter struct {
	ServiceInstance types.String `tfsdk:"service_instance"`
	Route           types.String `tfsdk:"route"`
}

func NewServiceRouteBindingListResource() list.ListResource {
	return &serviceRouteBindingListResource{}
}

func (r *serviceRouteBindingListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_route_binding" // must match managed resource
}

func (r *serviceRouteBindingListResource) Configure(_ context.Context,
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

func (r *serviceRouteBindingListResource) ListResourceConfigSchema(_ context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all service route bindings the caller has access to, optionally filtered by service instance or route.",
		Attributes: map[string]schema.Attribute{
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service instance to filter route bindings by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"route": schema.StringAttribute{
				MarkdownDescription: "The GUID of the route to filter route bindings by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all service route bindings from the API.
func (r *serviceRouteBindingListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter serviceRouteBindingListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	opts := cfv3client.NewServiceRouteBindingListOptions()

	if !filter.ServiceInstance.IsNull() {
		opts.ServiceInstanceGUIDs = cfv3client.Filter{
			Values: []string{filter.ServiceInstance.ValueString()},
		}
	}

	if !filter.Route.IsNull() {
		opts.RouteGUIDs = cfv3client.Filter{
			Values: []string{filter.Route.ValueString()},
		}
	}

	bindings, err := r.cfClient.ServiceRouteBindings.ListAll(ctx, opts)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Service Route Bindings",
			"Could not list service route bindings: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, binding := range bindings {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("service_route_binding_guid"), binding.GUID)

			if req.IncludeResource {
				resBinding, diags := mapServiceRouteBindingValuesToType(ctx, binding)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resBinding)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
