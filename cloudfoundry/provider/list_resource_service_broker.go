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

var _ list.ListResourceWithConfigure = &serviceBrokerListResource{}

type serviceBrokerListResource struct {
	cfClient *cfv3client.Client
}

type serviceBrokerListResourceFilter struct {
	Space types.String `tfsdk:"space"`
}

func NewServiceBrokerListResource() list.ListResource {
	return &serviceBrokerListResource{}
}

func (r *serviceBrokerListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_broker" // must match managed resource
}

func (r *serviceBrokerListResource) Configure(_ context.Context,
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

func (r *serviceBrokerListResource) ListResourceConfigSchema(_ context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all service brokers the caller has access to, optionally filtered by space.",
		Attributes: map[string]schema.Attribute{
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space to filter service brokers by. Returns only space-scoped brokers in that space.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all service brokers from the API.
func (r *serviceBrokerListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter serviceBrokerListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	opts := cfv3client.NewServiceBrokerListOptions()

	if !filter.Space.IsNull() {
		opts.SpaceGUIDs = cfv3client.Filter{
			Values: []string{filter.Space.ValueString()},
		}
	}

	serviceBrokers, err := r.cfClient.ServiceBrokers.ListAll(ctx, opts)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Service Brokers",
			"Could not list service brokers: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, sb := range serviceBrokers {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("service_broker_guid"), sb.GUID)

			if req.IncludeResource {
				resSB, diags := mapServiceBrokerValuesToType(ctx, sb)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resSB)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
