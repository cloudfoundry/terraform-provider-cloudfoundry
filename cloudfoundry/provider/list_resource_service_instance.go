package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &serviceInstanceListResource{}

type serviceInstanceListResource struct {
	cfClient *cfv3client.Client
}

type serviceInstanceListResourceFilter struct {
	Org   types.String `tfsdk:"org"`
	Space types.String `tfsdk:"space"`
}

func NewServiceInstanceListResource() list.ListResource {
	return &serviceInstanceListResource{}
}

func (r *serviceInstanceListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_instance" // must match managed resource
}

func (r *serviceInstanceListResource) Configure(_ context.Context,
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

func (r *serviceInstanceListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all service instances the caller has access to, optionally filtered by organization or space.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the organization to filter service instances by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to filter service instances by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all service instances from the API.
func (r *serviceInstanceListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter serviceInstanceListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	opts := cfv3client.NewServiceInstanceListOptions()

	if !filter.Org.IsNull() {
		opts.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{filter.Org.ValueString()},
		}
	}

	if !filter.Space.IsNull() {
		opts.SpaceGUIDs = cfv3client.Filter{
			Values: []string{filter.Space.ValueString()},
		}
	}

	serviceInstances, err := r.cfClient.ServiceInstances.ListAll(ctx, opts)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Service Instances",
			"Could not list service instances: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, si := range serviceInstances {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("service_instance_guid"), si.GUID)

			if req.IncludeResource {
				resSI, diags := mapResourceServiceInstanceValuesToType(ctx, si, jsontypes.NewNormalizedNull())
				result.Diagnostics.Append(diags...)

				resSI.Timeouts = timeouts.Value{
					Object: types.ObjectNull(map[string]attr.Type{
						"create": types.StringType,
						"delete": types.StringType,
						"update": types.StringType,
					}),
				}

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resSI)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
