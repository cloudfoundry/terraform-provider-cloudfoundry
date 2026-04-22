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

var _ list.ListResourceWithConfigure = &serviceCredentialBindingListResource{}

type serviceCredentialBindingListResource struct {
	cfClient *cfv3client.Client
}

type serviceCredentialBindingListResourceFilter struct {
	ServiceInstance types.String `tfsdk:"service_instance"`
	App             types.String `tfsdk:"app"`
}

func NewServiceCredentialBindingListResource() list.ListResource {
	return &serviceCredentialBindingListResource{}
}

func (r *serviceCredentialBindingListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_credential_binding" // must match managed resource
}

func (r *serviceCredentialBindingListResource) Configure(_ context.Context,
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

func (r *serviceCredentialBindingListResource) ListResourceConfigSchema(_ context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all service credential bindings the caller has access to, optionally filtered by service instance or app.",
		Attributes: map[string]schema.Attribute{
			"service_instance": schema.StringAttribute{
				MarkdownDescription: "The GUID of the service instance to filter bindings by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"app": schema.StringAttribute{
				MarkdownDescription: "The GUID of the app to filter bindings by.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all service credential bindings from the API.
func (r *serviceCredentialBindingListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter serviceCredentialBindingListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	opts := cfv3client.NewServiceCredentialBindingListOptions()

	if !filter.ServiceInstance.IsNull() {
		opts.ServiceInstanceGUIDs = cfv3client.Filter{
			Values: []string{filter.ServiceInstance.ValueString()},
		}
	}

	if !filter.App.IsNull() {
		opts.AppGUIDs = cfv3client.Filter{
			Values: []string{filter.App.ValueString()},
		}
	}

	bindings, err := r.cfClient.ServiceCredentialBindings.ListAll(ctx, opts)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Service Credential Bindings",
			"Could not list service credential bindings: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, binding := range bindings {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("service_credential_binding_guid"), binding.GUID)

			if req.IncludeResource {
				resBinding, diags := mapServiceCredentialBindingValuesToType(ctx, binding)
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
