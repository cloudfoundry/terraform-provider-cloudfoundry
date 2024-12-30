package provider

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/policy_client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type networkPolicyResource struct {
	policyClient *policy_client.ExternalClient
}

var (
	_ resource.ResourceWithConfigure   = &networkPolicyResource{}
	_ resource.ResourceWithImportState = &networkPolicyResource{}
)

func NewNetworkPolicyResource() resource.Resource {
	return &networkPolicyResource{}
}

func (r *networkPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_policy"
}

func (r *networkPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a Cloud Foundry resource for managing Cloud Foundry network policies to manage access between applications via container-to-container networking.`,

		Attributes: map[string]schema.Attribute{
			"source_app": schema.StringAttribute{
				MarkdownDescription: "The ID of the application to connect from.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"destination_app": schema.StringAttribute{
				MarkdownDescription: "The ID of the application to connect to.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Port (8080) or range of ports (8080-8085) for connection to destination app",
				Required:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "One of 'udp' or 'tcp' identifying the allowed protocol for the access.",
				Optional:            true,
			},
		},
	}
}

func (r *networkPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}

}

func (r *networkPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

}

func (r *networkPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *networkPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *networkPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (rs *networkPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
