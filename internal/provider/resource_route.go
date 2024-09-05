package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &RouteResource{}
	_ resource.ResourceWithConfigure   = &RouteResource{}
	_ resource.ResourceWithImportState = &RouteResource{}
)

// Instantiates a security group resource.
func NewRouteResource() resource.Resource {
	return &RouteResource{}
}

// Contains reference to the v3 client to be used for making the API calls.
type RouteResource struct {
	cfClient *cfv3client.Client
}

func (r *RouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *RouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource for managing Cloud Foundry application routes.",
		Attributes: map[string]schema.Attribute{
			idKey: guidSchema(),
			"space": schema.StringAttribute{
				MarkdownDescription: "The space guid associated to the route.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain guid associated to the route.",
				Required:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The hostname for the route; not compatible with routes specifying the tcp protocol; must be either a wildcard (*) or be under 63 characters long and only contain letters, numbers, dashes (-) or underscores(_)",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path for the route; not compatible with routes specifying the tcp protocol; must be under 128 characters long and not contain question marks (?), begin with a slash (/) and not be exactly a slash (/).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port that the route listens on. Only compatible with routes specifying the tcp protocol",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol supported by the route, based on the route's domain configuration.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL for the route.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destinations": schema.SetNestedAttribute{
				MarkdownDescription: "A destination represents the relationship between a route and a resource that can serve traffic.",
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						idKey: schema.StringAttribute{
							MarkdownDescription: "The GUID of the object.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								ReComputeStringValue(),
							},
						},
						"app_id": schema.StringAttribute{
							MarkdownDescription: "The GUID of the app to route traffic to.",
							Required:            true,
							Validators: []validator.String{
								validation.ValidUUID(),
							},
						},
						"app_process_type": schema.StringAttribute{
							MarkdownDescription: "Type of the process belonging to the app to route traffic to.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								ReComputeStringValue(),
							},
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "Port on the destination process to route traffic to.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.Int64{
								int64validator.Between(1024, 65535),
							},
							PlanModifiers: []planmodifier.Int64{
								ReComputeIntValue(),
							},
						},
						"weight": schema.Int64Attribute{
							MarkdownDescription: "Percentage of traffic which will be routed to this destination.",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
							PlanModifiers: []planmodifier.Int64{
								ReComputeIntValue(),
							},
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol to use for this destination.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("http1", "http2", "tcp"),
							},
							PlanModifiers: []planmodifier.String{
								ReComputeStringValue(),
							},
						},
					},
				},
			},
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
}

func (r *RouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config routeType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRoute, diags := plan.mapCreateRouteTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.cfClient.Routes.Create(ctx, &createRoute)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Creating Route",
			"Could not create Route with domain ID "+plan.Domain.ValueString()+" and space ID "+plan.Space.ValueString()+" : "+err.Error(),
		)
		return
	}

	if !plan.Destinations.IsNull() {

		insertDestinations, diags := config.mapCreateDestinationsTypeToValues(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		insertedDestinations, err := r.cfClient.Routes.ReplaceDestinations(ctx, route.GUID, insertDestinations)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Inserting Destinations",
				"Could not add destinations to route with ID "+route.GUID+" : "+err.Error(),
			)
		} else {
			route.Destinations = mapDestinationPointerSliceToDestinationSlice(insertedDestinations.Destinations)
		}
	}

	plan, diags = mapRouteValuesToType(ctx, route)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "created a route resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (rs *RouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data routeType
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := rs.cfClient.Routes.Get(ctx, data.Id.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "route", data.Id.ValueString())
		return
	}

	data, diags = mapRouteValuesToType(ctx, route)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "read a route resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *RouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, previousState, config routeType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error

	replaceDestinations, diags := config.mapCreateDestinationsTypeToValues(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = rs.cfClient.Routes.ReplaceDestinations(ctx, plan.Id.ValueString(), replaceDestinations)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Replacing Destinations",
			"Could not replace destinations of route with ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
	}

	updateRoute, diags := plan.mapUpdateRouteTypeToValues(ctx, &previousState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := rs.cfClient.Routes.Update(ctx, plan.Id.ValueString(), &updateRoute)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Updating Route",
			"Could not update Route with ID "+plan.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	data, diags := mapRouteValuesToType(ctx, route)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "updated a route resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (rs *RouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := rs.cfClient.Routes.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Route",
			"Could not delete the Route with ID "+state.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	if err = pollJob(ctx, *rs.cfClient, jobID, defaultTimeout); err != nil {
		resp.Diagnostics.AddError(
			"API Error Deleting Route",
			"Failed in deleting the Route with ID "+state.Id.ValueString()+" : "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a route resource")
}

func (rs *RouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
