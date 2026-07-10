package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type serviceInstanceResource struct {
	cfClient *cfv3client.Client
}

var (
	_ resource.Resource                   = &serviceInstanceResource{}
	_ resource.ResourceWithConfigure      = &serviceInstanceResource{}
	_ resource.ResourceWithImportState    = &serviceInstanceResource{}
	_ resource.ResourceWithValidateConfig = &serviceInstanceResource{}
	_ resource.ResourceWithIdentity       = &serviceInstanceResource{}
)

const (
	managedSerivceInstance      = "managed"
	userProvidedServiceInstance = "user-provided"
)

func NewServiceInstanceResource() resource.Resource {
	return &serviceInstanceResource{}
}

type serviceInstanceResourceIdentityModel struct {
	ServiceInstanceGUID types.String `tfsdk:"service_instance_guid"`
}

func (r *serviceInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_instance"
}

func (r *serviceInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates a service instance in a cloudfoundry space.

__Further documentation:__
https://docs.cloudfoundry.org/devguide/services`,

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service instance",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the service instance. Either managed or user-provided.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("managed", "user-provided"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The ID of the space in which to create the service instance",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"service_plan": schema.StringAttribute{
				MarkdownDescription: "The ID of the service plan from which to create the service instance. Conflicts with `service_plan_name` and `service_offering_name`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("service_plan_name")),
					stringvalidator.ConflictsWith(path.MatchRoot("service_offering_name")),
				},
			},
			"service_plan_name": schema.StringAttribute{
				MarkdownDescription: "The name of the service plan. Must be set together with `service_offering_name`. Conflicts with `service_plan`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("service_plan")),
					stringvalidator.AlsoRequires(path.MatchRoot("service_offering_name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service_offering_name": schema.StringAttribute{
				MarkdownDescription: "The name of the service offering. Must be set together with `service_plan_name`. Conflicts with `service_plan`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("service_plan")),
					stringvalidator.AlsoRequires(path.MatchRoot("service_plan_name")),
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					//If the offering changes a new service instance needs to be created
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "A JSON object that is passed to the service broker for managed service instance.",
				Optional:            true,
				Sensitive:           true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of tags used by apps to identify service instances. They are shown in the app VCAP_SERVICES env.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"credentials": schema.StringAttribute{
				MarkdownDescription: "A JSON object that is made available to apps bound to this service instance of type user-provided.",
				Optional:            true,
				Sensitive:           true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"syslog_drain_url": schema.StringAttribute{
				MarkdownDescription: "URL to which logs for bound applications will be streamed; only shown when type is user-provided.",
				Optional:            true,
			},
			"route_service_url": schema.StringAttribute{
				MarkdownDescription: "URL to which requests for bound routes will be forwarded; only shown when type is user-provided.",
				Optional:            true,
			},
			"maintenance_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Information about the version of this service instance; only shown when type is managed",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						MarkdownDescription: "The version of the service instance",
						Computed:            true,
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "A description of the version of the service instance",
						Computed:            true,
					},
				},
			},
			"upgrade_available": schema.BoolAttribute{
				MarkdownDescription: "Whether or not an upgrade of this service instance is available on the current Service Plan; details are available in the maintenance_info object; Only shown when type is managed",
				Computed:            true,
			},
			"dashboard_url": schema.StringAttribute{
				MarkdownDescription: "The URL to the service instance dashboard (or null if there is none); only shown when type is managed.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create:            true,
				CreateDescription: "Timeout for creating the service instance. Default is 40 minutes",
				Update:            true,
				UpdateDescription: "Timeout for updating the service instance. Default is 40 minutes",
				Delete:            true,
				DeleteDescription: "Timeout for deleting the service instance. Default is 40 minutes",
			}),
			idKey:            guidSchema(),
			"last_operation": lastOperationSchema(),
			labelsKey:        resourceLabelsSchema(),
			annotationsKey:   resourceAnnotationsSchema(),
			createdAtKey:     createdAtSchema(),
			updatedAtKey:     updatedAtSchema(),
		},
	}
}

func (rs *serviceInstanceResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_instance_guid": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *serviceInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.cfClient = session.CFClient
}

func (r *serviceInstanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config serviceInstanceType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Type.ValueString() == userProvidedServiceInstance {
		if !config.ServicePlan.IsNull() || !config.ServicePlanName.IsNull() || !config.ServiceOfferingName.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("service_plan"),
				"Conflicting attribute service instance",
				"service_plan, service_plan_name, and service_offering_name are not allowed for user-provided service instances",
			)
			return
		}
	}

	if config.Type.ValueString() == managedSerivceInstance {
		hasPlanGUID := !config.ServicePlan.IsNull()
		hasPlanName := !config.ServicePlanName.IsNull()
		hasOfferingName := !config.ServiceOfferingName.IsNull()

		//anyUnknown := config.ServicePlan.IsUnknown() || config.ServicePlanName.IsUnknown() || config.ServiceOfferingName.IsUnknown()

		if !hasPlanGUID && !hasPlanName && !hasOfferingName {
			resp.Diagnostics.AddAttributeError(
				path.Root("service_plan"),
				"Missing attribute for managed service instance",
				"Either service_plan (GUID) or both service_plan_name and service_offering_name must be set for managed service instances",
			)
			return
		}
	}

	// If Service Instance is of type managed only parameters is allowed to pass
	if !config.Parameters.IsNull() && config.Type.ValueString() == userProvidedServiceInstance {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Parameters can only passed to service instance of type managed",
			"Parameters json object can only be passed to managed serivce instance",
		)
		return
	}

	// If Service instance of type user-provided then credentials , syslog_drain_url and route_service_url allowed
	if !config.SyslogDrainURL.IsNull() || !config.RouteServiceURL.IsNull() || !config.Credentials.IsNull() {
		if config.Type.ValueString() == managedSerivceInstance {
			resp.Diagnostics.AddAttributeError(
				path.Root("type"),
				"Mistmatch attribute passed to user provided service instance",
				"Allowed attributes for serivce instance of type user provided: credentials, syslog_drain_url, route_service_url",
			)
		}
	}
}

func (r *serviceInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, state serviceInstanceType
	var serviceInstance *cfv3resource.ServiceInstance
	var err error
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 40*time.Minute)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured create timeout", map[string]any{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})
	}

	switch plan.Type.ValueString() {
	case managedSerivceInstance:
		// Resolve plan GUID
		planGUID := plan.ServicePlan.ValueString()
		planName := plan.ServicePlanName.ValueString()
		offeringName := plan.ServiceOfferingName.ValueString()
		if planName != "" && offeringName != "" {
			resolved, resolveErr := r.resolvePlanGUID(ctx, planName, offeringName)
			if resolveErr != nil {
				resp.Diagnostics.AddError("Error resolving service plan", resolveErr.Error())
				return
			}
			planGUID = resolved
		} else {
			planName, offeringName, err = r.lookupPlanNames(ctx, planGUID)
			if err != nil {
				resp.Diagnostics.AddError("Error resolving service plan", fmt.Sprintf("%s", err))
				return
			}

		}

		createServiceInstance := cfv3resource.ServiceInstanceManagedCreate{
			Type: plan.Type.ValueString(),
			Name: plan.Name.ValueString(),
			Relationships: cfv3resource.ServiceInstanceRelationships{
				ServicePlan: &cfv3resource.ToOneRelationship{
					Data: &cfv3resource.Relationship{
						GUID: planGUID,
					},
				},
				Space: &cfv3resource.ToOneRelationship{
					Data: &cfv3resource.Relationship{
						GUID: plan.Space.ValueString(),
					},
				},
			},
			Metadata: cfv3resource.NewMetadata(),
		}
		if !plan.Parameters.IsNull() {
			var params json.RawMessage
			err := json.Unmarshal([]byte(plan.Parameters.ValueString()), &params)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error in unmarshalling parameters",
					"Unable to unmarshal json parameters of service instance"+plan.Name.ValueString()+": "+err.Error(),
				)
				return
			}
			createServiceInstance.Parameters = &params
		}
		if !plan.Tags.IsNull() {
			tags, diags := toTagsList(ctx, plan.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			createServiceInstance.Tags = tags
		}
		labelsDiags := plan.Labels.ElementsAs(ctx, &createServiceInstance.Metadata.Labels, false)
		resp.Diagnostics.Append(labelsDiags...)

		annotationsDiags := plan.Annotations.ElementsAs(ctx, &createServiceInstance.Metadata.Annotations, false)
		resp.Diagnostics.Append(annotationsDiags...)

		jobID, err := r.cfClient.ServiceInstances.CreateManaged(ctx, &createServiceInstance)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error in creating managed service instance",
				"Unable to create service instance "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}
		if err = pollJob(ctx, *r.cfClient, jobID, createTimeout); err != nil {
			resp.Diagnostics.AddError(
				"Unable to verify service instance creation",
				"Service Instance verification failed for+ "+plan.Name.ValueString()+": "+err.Error(),
			)
		}
		serviceInstance, err = r.cfClient.ServiceInstances.Single(ctx, &cfv3client.ServiceInstanceListOptions{
			Names: cfv3client.Filter{
				Values: []string{
					plan.Name.ValueString(),
				},
			},
			SpaceGUIDs: cfv3client.Filter{
				Values: []string{
					plan.Space.ValueString(),
				},
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error get service instance after creation",
				"Unable to fetch created service instance"+plan.Name.ValueString()+": "+err.Error(),
			)
		}

		state, diags = mapResourceServiceInstanceValuesToType(ctx, serviceInstance, plan.Parameters)
		resp.Diagnostics.Append(diags...)
		state.ServicePlanName = types.StringValue(planName)
		state.ServiceOfferingName = types.StringValue(offeringName)

	case userProvidedServiceInstance:

		createServiceInstance := cfv3resource.ServiceInstanceUserProvidedCreate{
			Type: plan.Type.ValueString(),
			Name: plan.Name.ValueString(),
			Relationships: cfv3resource.ServiceInstanceRelationships{
				Space: &cfv3resource.ToOneRelationship{
					Data: &cfv3resource.Relationship{
						GUID: plan.Space.ValueString(),
					},
				},
			},
			Metadata: cfv3resource.NewMetadata(),
		}
		if !plan.Credentials.IsNull() {
			var credentials json.RawMessage
			err := json.Unmarshal([]byte(plan.Credentials.ValueString()), &credentials)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error in unmarshalling credentials",
					"Unable to unmarshal json credentials of service instance"+plan.Name.ValueString()+": "+err.Error(),
				)
				return
			}
			createServiceInstance.Credentials = &credentials
		}
		if !plan.Tags.IsNull() {
			tags, diags := toTagsList(ctx, plan.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			createServiceInstance.Tags = tags
		}

		if !plan.SyslogDrainURL.IsNull() {
			createServiceInstance.SyslogDrainURL = plan.SyslogDrainURL.ValueStringPointer()
		}
		if !plan.RouteServiceURL.IsNull() {
			createServiceInstance.RouteServiceURL = plan.RouteServiceURL.ValueStringPointer()
		}

		labelsDiags := plan.Labels.ElementsAs(ctx, &createServiceInstance.Metadata.Labels, false)
		resp.Diagnostics.Append(labelsDiags...)

		annotationsDiags := plan.Annotations.ElementsAs(ctx, &createServiceInstance.Metadata.Annotations, false)
		resp.Diagnostics.Append(annotationsDiags...)

		_, err = r.cfClient.ServiceInstances.CreateUserProvided(ctx, &createServiceInstance)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error in creating user-provided service instance",
				"Unable to create service instance "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}
		serviceInstance, err = r.cfClient.ServiceInstances.Single(ctx, &cfv3client.ServiceInstanceListOptions{
			Names: cfv3client.Filter{
				Values: []string{
					plan.Name.ValueString(),
				},
			},
			SpaceGUIDs: cfv3client.Filter{
				Values: []string{
					plan.Space.ValueString(),
				},
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error get service instance after creation",
				"Unable to fetch created service instance"+plan.Name.ValueString()+": "+err.Error(),
			)
		}
		state, diags = mapResourceServiceInstanceValuesToType(ctx, serviceInstance, plan.Credentials)
		resp.Diagnostics.Append(diags...)

	}
	state.Timeouts = plan.Timeouts
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	identity := serviceInstanceResourceIdentityModel{
		ServiceInstanceGUID: types.StringValue(state.ID.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)

}

func (r *serviceInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data, newState serviceInstanceType

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	svcInstance, err := r.cfClient.ServiceInstances.Get(ctx, data.ID.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "service_instance", data.ID.ValueString())
		return
	}

	switch svcInstance.Type {
	case managedSerivceInstance:
		newState, diags = mapResourceServiceInstanceValuesToType(ctx, svcInstance, data.Parameters)
		newState.ServicePlanName = data.ServicePlanName
		newState.ServiceOfferingName = data.ServiceOfferingName
		// Update plan and offering name during a refresh
		if data.ServicePlanName.ValueString() == "" && data.ServiceOfferingName.ValueString() == "" {
			planName, offeringName, err := r.lookupPlanNames(ctx, newState.ServicePlan.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("API Error Reading Resource Service Plan", fmt.Sprintf("%s", err))
				return
			}
			newState.ServicePlanName = types.StringValue(planName)
			newState.ServiceOfferingName = types.StringValue(offeringName)
		}
	case userProvidedServiceInstance:
		newState, diags = mapResourceServiceInstanceValuesToType(ctx, svcInstance, data.Credentials)
	}
	newState.Timeouts = data.Timeouts
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)

	var identity serviceInstanceResourceIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = serviceInstanceResourceIdentityModel{
			ServiceInstanceGUID: types.StringValue(data.ID.ValueString()),
		}
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)

}

func (r *serviceInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan, state, previousState serviceInstanceType
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)

	updateTimeout, diags := plan.Timeouts.Update(ctx, 40*time.Minute)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured update timeout", map[string]any{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})
	}
	switch plan.Type.ValueString() {
	case managedSerivceInstance:
		planGUID := previousState.ServicePlan.ValueString()
		planName := previousState.ServicePlanName.ValueString()
		offeringName := previousState.ServiceOfferingName.ValueString()
		var resolveErr error
		if plan.ServicePlanName.ValueString() != "" && plan.ServicePlanName.ValueString() != previousState.ServicePlanName.ValueString() {

			planGUID, resolveErr = r.resolvePlanGUID(ctx, plan.ServicePlanName.ValueString(), plan.ServiceOfferingName.ValueString())
			if resolveErr != nil {
				resp.Diagnostics.AddError("Error resolving service plan", resolveErr.Error())
				return
			}
			planName = plan.ServicePlanName.ValueString()
			offeringName = plan.ServiceOfferingName.ValueString()
		} else if plan.ServicePlan.ValueString() != "" && plan.ServicePlan.ValueString() != previousState.ServicePlan.ValueString() {

			planGUID = plan.ServicePlan.ValueString()

			planName, offeringName, resolveErr = r.lookupPlanNames(ctx, planGUID)
			if resolveErr != nil {
				resp.Diagnostics.AddError("API Error Reading Resource Service Plan (Subaccount)", fmt.Sprintf("%s", resolveErr))
				return
			}

		}

		updateServiceInstance := cfv3resource.ServiceInstanceManagedUpdate{}
		if plan.Name.ValueString() != previousState.Name.ValueString() {
			updateServiceInstance.Name = plan.Name.ValueStringPointer()
		}
		if planGUID != previousState.ServicePlan.ValueString() {
			updateServiceInstance.Relationships = &cfv3resource.ServiceInstanceRelationships{
				ServicePlan: &cfv3resource.ToOneRelationship{
					Data: &cfv3resource.Relationship{
						GUID: planGUID,
					},
				},
			}
		}
		if !plan.Parameters.IsNull() {
			var params json.RawMessage
			err := json.Unmarshal([]byte(plan.Parameters.ValueString()), &params)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error in unmarshalling parameters",
					"Unable to unmarshal json parameters during update of service instance"+plan.Name.ValueString()+": "+err.Error(),
				)
				return
			}
			updateServiceInstance.Parameters = &params
		}

		tags, diags := toTagsList(ctx, plan.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateServiceInstance.Tags = tags

		updateServiceInstance.Metadata, diags = setClientMetadataForUpdate(ctx, previousState.Labels, previousState.Annotations, plan.Labels, plan.Annotations)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		jobID, _, err := r.cfClient.ServiceInstances.UpdateManaged(ctx, previousState.ID.ValueString(), &updateServiceInstance)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error in updating managed service instance",
				"Unable to update service instance "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}
		if jobID != "" {
			if err := pollJob(ctx, *r.cfClient, jobID, updateTimeout); err != nil {
				resp.Diagnostics.AddError(
					"Unable to verify service instance update",
					"Service Instance update verification failed for "+plan.Name.ValueString()+": "+err.Error(),
				)
			}
		}
		serviceInstance, err := r.cfClient.ServiceInstances.Get(ctx, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error get service instance after update",
				"Unable to fetch updated service instance"+plan.Name.ValueString()+": "+err.Error(),
			)
		}
		state, diags = mapResourceServiceInstanceValuesToType(ctx, serviceInstance, plan.Parameters)
		resp.Diagnostics.Append(diags...)
		state.ServicePlanName = types.StringValue(planName)
		state.ServiceOfferingName = types.StringValue(offeringName)
	case userProvidedServiceInstance:

		updateServiceInstance := cfv3resource.ServiceInstanceUserProvidedUpdate{}
		if plan.Name.ValueString() != previousState.Name.ValueString() {
			updateServiceInstance.Name = plan.Name.ValueStringPointer()
		}
		if !plan.Credentials.IsNull() {
			var credentials json.RawMessage
			err := json.Unmarshal([]byte(plan.Credentials.ValueString()), &credentials)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error in unmarshalling credentials",
					"Unable to unmarshal json credentials during update of service instance"+plan.Name.ValueString()+": "+err.Error(),
				)
				return
			}
			updateServiceInstance.Credentials = &credentials
		}

		updateServiceInstance.SyslogDrainURL = plan.SyslogDrainURL.ValueStringPointer()
		updateServiceInstance.RouteServiceURL = plan.RouteServiceURL.ValueStringPointer()

		tags, diags := toTagsList(ctx, plan.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateServiceInstance.Tags = tags

		updateServiceInstance.Metadata, diags = setClientMetadataForUpdate(ctx, previousState.Labels, previousState.Annotations, plan.Labels, plan.Annotations)
		resp.Diagnostics.Append(diags...)
		_, err := r.cfClient.ServiceInstances.UpdateUserProvided(ctx, previousState.ID.ValueString(), &updateServiceInstance)
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error in updating user-provided service instance",
				"Unable to update service instance "+plan.Name.ValueString()+": "+err.Error(),
			)
			return
		}
		serviceInstance, err := r.cfClient.ServiceInstances.Get(ctx, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error get service instance after update",
				"Unable to fetch updated service instance"+plan.Name.ValueString()+": "+err.Error(),
			)
		}
		state, diags = mapResourceServiceInstanceValuesToType(ctx, serviceInstance, plan.Credentials)
		resp.Diagnostics.Append(diags...)
	}
	state.Timeouts = plan.Timeouts
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	// WORKAROUND for OpenTofu compatibility
	// https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/418
	identity := serviceInstanceResourceIdentityModel{
		ServiceInstanceGUID: types.StringValue(state.ID.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
	// END WORKAROUND

}

func (r *serviceInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceInstanceType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := state.Timeouts.Delete(ctx, 40*time.Minute)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured delete timeout", map[string]any{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})
	}

	jobID, err := r.cfClient.ServiceInstances.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error in deleting service instance",
			"Unable to delete service instance "+state.Name.ValueString()+": "+err.Error(),
		)

	}
	if jobID != "" {
		if err := pollJob(ctx, *r.cfClient, jobID, deleteTimeout); err != nil {
			resp.Diagnostics.AddError(
				"Unable to verify service instance deletion",
				"Service Instance deletion verification failed for "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}

}

func (rs *serviceInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("service_instance_guid"), req, resp)
}

// resolvePlanGUID looks up a service plan GUID from its offering name and plan name.
func (r *serviceInstanceResource) resolvePlanGUID(ctx context.Context, servicePlanName, serviceOfferingName string) (string, error) {
	opts := &cfv3client.ServicePlanListOptions{
		Names:                cfv3client.Filter{Values: []string{servicePlanName}},
		ServiceOfferingNames: cfv3client.Filter{Values: []string{serviceOfferingName}},
	}
	plan, err := r.cfClient.ServicePlans.Single(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("could not find unique service plan %q for offering %q: %w", servicePlanName, serviceOfferingName, err)
	}
	return plan.GUID, nil
}

// lookupPlanNames resolves the service plan name and offering name from a plan GUID.
func (r *serviceInstanceResource) lookupPlanNames(ctx context.Context, planGUID string) (planName, offeringName string, err error) {
	plan, err := r.cfClient.ServicePlans.Get(ctx, planGUID)
	if err != nil {
		err = fmt.Errorf("could not look up service plan %q: %w", planGUID, err)
		return
	}
	planName = plan.Name
	offering, offeringErr := r.cfClient.ServiceOfferings.Get(ctx, plan.Relationships.ServiceOffering.Data.GUID)
	if offeringErr != nil {
		err = fmt.Errorf("could not look up service offering for plan %q: %w", planGUID, offeringErr)
		return
	}
	offeringName = offering.Name
	return
}
