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

var _ list.ListResourceWithConfigure = &securityGroupListResource{}

type securityGroupListResource struct {
	cfClient *cfv3client.Client
}

type securityGroupListResourceFilter struct {
	RunningSpace types.String `tfsdk:"running_space"`
	StagingSpace types.String `tfsdk:"staging_space"`
}

func NewSecurityGroupListResource() list.ListResource {
	return &securityGroupListResource{}
}

func (r *securityGroupListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group" // must match managed resource
}

func (r *securityGroupListResource) Configure(_ context.Context,
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

func (r *securityGroupListResource) ListResourceConfigSchema(
	_ context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This list resource allows you to discover all security groups the caller has access to, optionally filtered by running or staging space.",
		Attributes: map[string]schema.Attribute{
			"running_space": schema.StringAttribute{
				MarkdownDescription: "The GUID of a space to filter by; returns only security groups bound to that space for running.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"staging_space": schema.StringAttribute{
				MarkdownDescription: "The GUID of a space to filter by; returns only security groups bound to that space for staging.",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
		},
	}
}

// List streams all security groups from the API.
func (r *securityGroupListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var filter securityGroupListResourceFilter

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	opts := cfv3client.NewSecurityGroupListOptions()

	if !filter.RunningSpace.IsNull() {
		opts.RunningSpaceGUIDs = cfv3client.Filter{
			Values: []string{filter.RunningSpace.ValueString()},
		}
	}

	if !filter.StagingSpace.IsNull() {
		opts.StagingSpaceGUIDs = cfv3client.Filter{
			Values: []string{filter.StagingSpace.ValueString()},
		}
	}

	securityGroups, err := r.cfClient.SecurityGroups.ListAll(ctx, opts)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Fetching Security Groups",
			"Could not list security groups: "+err.Error(),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, sg := range securityGroups {
			result := req.NewListResult(ctx)

			result.Identity.SetAttribute(ctx, path.Root("security_group_guid"), sg.GUID)

			if req.IncludeResource {
				resSG, diags := mapSecurityGroupValuesToType(ctx, sg)
				result.Diagnostics.Append(diags...)

				if !result.Diagnostics.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resSG)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
